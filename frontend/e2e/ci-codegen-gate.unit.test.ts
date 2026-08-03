import { execFileSync, spawnSync } from 'node:child_process';
import { mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { afterAll, beforeAll, describe, expect, it } from 'vitest';

const SCRIPT = join(__dirname, '..', '..', '.github', 'scripts', 'codegen-inputs-changed.sh');

let repo = '';
const sha = { base: '', merge: '', schema: '', unrelated: '' };

// -c user.* is passed on every invocation (not just `commit`) because `commit-tree` also refuses
// to run without an author identity, and CI runners have none configured globally.
const git = (...args: string[]) =>
    execFileSync('git', ['-C', repo, '-c', 'user.email=e2e@example.com', '-c', 'user.name=e2e', ...args], {
        encoding: 'utf8',
    }).trim();

const commit = (path: string, body: string, message: string) => {
    mkdirSync(join(repo, path.slice(0, path.lastIndexOf('/'))), { recursive: true });
    writeFileSync(join(repo, path), body);
    git('add', '-A');
    git('commit', '-m', message);

    return git('rev-parse', 'HEAD');
};

const run = (event: string, before: string, baseSha: string, headSha: string) =>
    execFileSync(SCRIPT, [event, before, baseSha, headSha], { cwd: repo, encoding: 'utf8' }).trim();

const runReason = (event: string, before: string, baseSha: string, headSha: string) =>
    spawnSync(SCRIPT, [event, before, baseSha, headSha], { cwd: repo, encoding: 'utf8' }).stderr.trim();

// The PR arrives as the merge commit GitHub builds, not as the branch head — the range the gate
// picks has to span the whole PR, not the newest push.
beforeAll(() => {
    repo = mkdtempSync(join(tmpdir(), 'codegen-gate-'));
    git('init', '-q', '-b', 'main');
    sha.base = commit('README.md', 'base\n', 'base');
    sha.schema = commit('backend/pkg/graph/schema.graphqls', 'type Query { a: Int }\n', 'edit the schema');
    sha.unrelated = commit('README.md', 'base\nmore\n', 'unrelated follow-up');
    sha.merge = git('commit-tree', `${sha.unrelated}^{tree}`, '-p', sha.base, '-p', sha.unrelated, '-m', 'merge');
});

afterAll(() => rmSync(repo, { force: true, recursive: true }));

describe('codegen freshness gate — range selection', () => {
    it('checks a follow-up push whose newest commit did not touch a codegen input', () => {
        expect(run('pull_request', sha.schema, sha.base, sha.merge)).toBe('changed=true');
    });

    it('checks a freshly opened PR, where there is no previous head', () => {
        expect(run('pull_request', '', sha.base, sha.merge)).toBe('changed=true');
    });

    // The answer alone cannot tell this branch from the diff-failed fallback below, which prints
    // the same `changed=true`; only the reason distinguishes them.
    it('checks a force-push, where the previous head no longer resolves', () => {
        expect(run('push', 'b'.repeat(40), '', sha.schema)).toBe('changed=true');
        expect(runReason('push', 'b'.repeat(40), '', sha.schema)).toContain('reason=unresolvable-base');
    });

    it('checks the first push of a branch, where there is no previous head at all', () => {
        expect(run('push', '0'.repeat(40), '', sha.schema)).toBe('changed=true');
        expect(runReason('push', '0'.repeat(40), '', sha.schema)).toContain('reason=unresolvable-base');
    });

    it('keeps checking a push by its own range', () => {
        expect(run('push', sha.base, '', sha.schema)).toBe('changed=true');
    });

    it('still skips a PR that touches no codegen input', () => {
        expect(run('pull_request', '', sha.schema, sha.unrelated)).toBe('changed=false');
    });

    it('checks a codegen input whose diff is larger than one pipe buffer', () => {
        const pad = 'x'.repeat(180);
        mkdirSync(join(repo, 'frontend', 'filler'), { recursive: true });

        for (let i = 0; i < 1000; i += 1) {
            writeFileSync(join(repo, 'frontend', 'filler', `${i}-${pad}.txt`), '');
        }

        const bulk = commit('backend/pkg/graph/schema.graphqls', 'type Query { a: Int, b: Int }\n', 'bulk');
        const names = git('diff', '--name-only', sha.unrelated, bulk);

        expect(names.split('\n')[0]).toBe('backend/pkg/graph/schema.graphqls');
        expect(names.length).toBeGreaterThan(64 * 1024);
        expect(run('push', sha.unrelated, '', bulk)).toBe('changed=true');
    });
});
