import { execFileSync } from 'node:child_process';
import { existsSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';

const SCRIPT = join(__dirname, 'review-sandbox.sh');

let sandboxRoot = '';

// The script derives its root from TMPDIR, so a per-test TMPDIR keeps every case inside its own
// throwaway tree — a run must never be able to name a real path.
const run = (...args: string[]) => {
    try {
        const stdout = execFileSync(SCRIPT, args, {
            encoding: 'utf8',
            env: { ...process.env, TMPDIR: sandboxRoot },
            stdio: ['ignore', 'pipe', 'pipe'],
        });

        return { code: 0, stdout: stdout.trim() };
    } catch (error) {
        const failure = error as { status: number; stdout: string };

        return { code: failure.status, stdout: failure.stdout.trim() };
    }
};

beforeEach(() => {
    sandboxRoot = mkdtempSync(join(tmpdir(), 'sandbox-test-'));
});

afterEach(() => {
    rmSync(sandboxRoot, { force: true, recursive: true });
    // The worktrees a case created are registered in the real repo, so dropping the directory is
    // only half of it — an unpruned registration would follow the developer around.
    execFileSync('git', ['worktree', 'prune'], { cwd: join(__dirname, '..', '..') });
});

describe('review-sandbox clean — containment', () => {
    it('removes a sandbox it created', () => {
        const { stdout: sandbox } = run('create');

        expect(existsSync(sandbox)).toBe(true);
        expect(run('clean', sandbox).code).toBe(0);
        expect(existsSync(sandbox)).toBe(false);
    });

    it('refuses a path outside the sandbox root and leaves it intact', () => {
        const outsider = mkdtempSync(join(tmpdir(), 'outsider-'));

        writeFileSync(join(outsider, 'keep-me'), 'payload');

        expect(run('clean', outsider).code).not.toBe(0);
        expect(existsSync(join(outsider, 'keep-me'))).toBe(true);

        rmSync(outsider, { force: true, recursive: true });
    });

    it('refuses the sandbox root itself, and a sibling of it', () => {
        const { stdout: sandbox } = run('create');
        const root = join(sandboxRoot, 'pentagi-review-sandboxes');

        expect(run('clean', root).code).not.toBe(0);
        expect(run('clean', `${root}-old`).code).not.toBe(0);
        expect(existsSync(sandbox)).toBe(true);
    });

    it('refuses a traversal that lands outside the root', () => {
        const { stdout: sandbox } = run('create');

        expect(run('clean', `${sandbox}/../../..`).code).not.toBe(0);
        expect(existsSync(sandbox)).toBe(true);
    });
});

describe('review-sandbox create — where the sandbox lands', () => {
    it('honours PENTAGI_SANDBOX_ROOT, the escape hatch for a $TMPDIR on another filesystem', () => {
        const override = mkdtempSync(join(tmpdir(), 'override-root-'));
        const sandbox = execFileSync(SCRIPT, ['create'], {
            encoding: 'utf8',
            env: { ...process.env, PENTAGI_SANDBOX_ROOT: override, TMPDIR: sandboxRoot },
        }).trim();

        expect(sandbox.startsWith(`${override}/pentagi-review-sandboxes/`)).toBe(true);

        execFileSync(SCRIPT, ['clean', sandbox], {
            env: { ...process.env, PENTAGI_SANDBOX_ROOT: override, TMPDIR: sandboxRoot },
        });
        expect(existsSync(sandbox)).toBe(false);

        rmSync(override, { force: true, recursive: true });
    });
});

describe('review-sandbox clean — the sweep', () => {
    it('refuses a bare clean rather than taking the whole root', () => {
        const { stdout: sandbox } = run('create');

        expect(run('clean').code).not.toBe(0);
        expect(existsSync(sandbox)).toBe(true);
    });

    it('sweeps only what has gone stale', () => {
        const { stdout: fresh } = run('create');

        expect(run('clean', '--all').code).toBe(0);
        expect(existsSync(fresh), 'a live sandbox survives the sweep').toBe(true);

        execFileSync('touch', ['-t', '202001010000', fresh]);
        expect(run('clean', '--all').code).toBe(0);
        expect(existsSync(fresh), 'a stale sandbox is swept').toBe(false);
    });
});
