import { execFileSync, spawnSync } from 'node:child_process';
import { mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { devNull, tmpdir } from 'node:os';
import { join } from 'node:path';
import { afterAll, beforeAll, describe, expect, it } from 'vitest';

const SCRIPT = join(__dirname, '..', '..', '.github', 'scripts', 'codegen-inputs-changed.sh');

let repo = '';
const gitEnvWithoutUserConfig = { ...process.env, GIT_CONFIG_GLOBAL: devNull, GIT_CONFIG_NOSYSTEM: '1' };
const sha = { base: '', merge: '', schema: '', unrelated: '' };

// -c user.* is passed on every invocation (not just `commit`) because `commit-tree` also refuses
// to run without an author identity, and CI runners have none configured globally.
const git = (...args: string[]) =>
    execFileSync('git', ['-C', repo, '-c', 'user.email=e2e@example.com', '-c', 'user.name=e2e', ...args], {
        encoding: 'utf8',
        env: gitEnvWithoutUserConfig,
    }).trim();

const commit = (path: string, body: string, message: string) => {
    mkdirSync(join(repo, path.slice(0, path.lastIndexOf('/'))), { recursive: true });
    writeFileSync(join(repo, path), body);
    git('add', '-A');
    git('commit', '-m', message);

    return git('rev-parse', 'HEAD');
};

const head = () => git('rev-parse', 'HEAD');

const decide = (event: string, before: string, baseSha: string, headSha: string) => {
    const { status, stderr, stdout } = spawnSync(SCRIPT, [event, before, baseSha, headSha], {
        cwd: repo,
        encoding: 'utf8',
        env: gitEnvWithoutUserConfig,
    });

    return {
        answer: stdout.trim(),
        reason: stderr.split('\n').find((line) => line.startsWith('reason=')) ?? '',
        status,
    };
};

const matched = { answer: 'changed=true', reason: '', status: 0 };
const skipped = { answer: 'changed=false', reason: '', status: 0 };

const CODEGEN_INPUTS = [
    'backend/pkg/graph/schema.graphqls',
    'frontend/graphql-schema.graphql',
    'frontend/graphql-codegen.ts',
    'frontend/pnpm-lock.yaml',
    'frontend/src/graphql/types.ts',
];

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
    // A matched input and both fallbacks all print `changed=true`, and CI runs the script under
    // `bash -e`, so every case pins the answer, the reason and the exit status together.
    it('checks a follow-up push whose newest commit did not touch a codegen input', () => {
        expect(decide('pull_request', sha.schema, sha.base, sha.merge)).toEqual(matched);
    });

    it('checks a freshly opened PR, where there is no previous head', () => {
        expect(decide('pull_request', '', sha.base, sha.merge)).toEqual(matched);
    });

    it('checks a force-push, where the previous head no longer resolves', () => {
        expect(decide('push', 'b'.repeat(40), '', sha.schema)).toEqual({
            ...matched,
            reason: 'reason=unresolvable-base',
        });
    });

    it('checks the first push of a branch, where there is no previous head at all', () => {
        expect(decide('push', '0'.repeat(40), '', sha.schema)).toEqual({
            ...matched,
            reason: 'reason=unresolvable-base',
        });
    });

    it('keeps checking a push by its own range', () => {
        expect(decide('push', sha.base, '', sha.schema)).toEqual(matched);
    });

    it('still skips a PR that touches no codegen input', () => {
        expect(decide('pull_request', '', sha.schema, sha.unrelated)).toEqual(skipped);
    });

    it('checks when the range cannot be diffed', () => {
        expect(decide('push', sha.base, '', 'c'.repeat(40))).toEqual({ ...matched, reason: 'reason=diff-failed' });
    });

    it.each(CODEGEN_INPUTS)('checks a push that touches only %s', (input) => {
        const before = head();
        const after = commit(input, `${input} edited\n`, `edit ${input}`);

        expect(decide('push', before, '', after)).toEqual(matched);
    });

    it.each(['vendor/frontend/graphql-schema.graphql', 'frontend/graphql-schema.graphql.orig'])(
        'does not take %s for an input: the whole path has to match',
        (lookalike) => {
            const before = head();
            const after = commit(lookalike, 'copy\n', `add ${lookalike}`);

            expect(decide('push', before, '', after)).toEqual(skipped);
        },
    );

    it('checks a push that deletes an input', () => {
        commit('frontend/graphql-codegen.ts', 'export default {};\n', 'add the codegen config');
        const before = head();

        git('rm', '-q', 'frontend/graphql-codegen.ts');
        git('commit', '-m', 'drop the codegen config');

        expect(decide('push', before, '', head())).toEqual(matched);
    });

    it('checks a push that renames an input away', () => {
        commit('frontend/graphql-schema.graphql', 'query A { a }\n', 'add the operations file');
        const before = head();

        git('mv', 'frontend/graphql-schema.graphql', 'frontend/schema.graphql');
        git('commit', '-m', 'rename the operations file');

        expect(decide('push', before, '', head())).toEqual(matched);
    });

    it('checks a push whose input change sits in a commit below the newest one', () => {
        const before = head();

        commit('README.md', 'base\nmore\nfirst\n', 'first');
        commit('frontend/pnpm-lock.yaml', 'lockfileVersion: 9\n', 'bump the lockfile');
        const after = commit('README.md', 'base\nmore\nfirst\nlast\n', 'last');

        expect(decide('push', before, '', after)).toEqual(matched);
    });

    it('checks a codegen input whose diff is larger than one pipe buffer', () => {
        const before = head();
        const pad = 'x'.repeat(180);
        mkdirSync(join(repo, 'frontend', 'filler'), { recursive: true });

        for (let i = 0; i < 1000; i += 1) {
            writeFileSync(join(repo, 'frontend', 'filler', `${i}-${pad}.txt`), '');
        }

        const bulk = commit('backend/pkg/graph/schema.graphqls', 'type Query { a: Int, b: Int }\n', 'bulk');
        const names = git('diff', '--name-only', before, bulk).split('\n');

        expect(names[0]).toBe('backend/pkg/graph/schema.graphqls');
        expect(names.filter((name) => CODEGEN_INPUTS.includes(name))).toEqual(['backend/pkg/graph/schema.graphqls']);
        expect(names.join('\n').length).toBeGreaterThan(64 * 1024);
        expect(decide('push', before, '', bulk)).toEqual(matched);
    });
});
