import { execFileSync } from 'node:child_process';
import { mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { afterAll, beforeAll, describe, expect, it } from 'vitest';

const SCRIPT = join(__dirname, '..', '..', '.github', 'scripts', 'codegen-inputs-changed.sh');

let repo = '';
const sha: Record<string, string> = {};

const git = (...args: string[]) => execFileSync('git', ['-C', repo, ...args], { encoding: 'utf8' }).trim();

const commit = (path: string, body: string, message: string) => {
    mkdirSync(join(repo, path.slice(0, path.lastIndexOf('/'))), { recursive: true });
    writeFileSync(join(repo, path), body);
    git('add', '-A');
    git('-c', 'user.email=e2e@example.com', '-c', 'user.name=e2e', 'commit', '-m', message);

    return git('rev-parse', 'HEAD');
};

const run = (event: string, before: string, baseSha: string, headSha: string) =>
    execFileSync(SCRIPT, [event, before, baseSha, headSha], { cwd: repo, encoding: 'utf8' }).trim();

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

    it('checks a force-push, where the previous head no longer resolves', () => {
        expect(run('pull_request', '0'.repeat(40), sha.base, sha.merge)).toBe('changed=true');
    });

    it('keeps checking a push by its own range', () => {
        expect(run('push', sha.base, '', sha.schema)).toBe('changed=true');
    });

    it('still skips a PR that touches no codegen input', () => {
        expect(run('pull_request', '', sha.schema, sha.unrelated)).toBe('changed=false');
    });
});
