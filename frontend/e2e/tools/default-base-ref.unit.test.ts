import { execFileSync, spawnSync } from 'node:child_process';
import { mkdtempSync, rmSync } from 'node:fs';
import { devNull, tmpdir } from 'node:os';
import { join } from 'node:path';
import { afterEach, describe, expect, it } from 'vitest';

import { defaultBaseRef } from './default-base-ref.ts';

const repos: string[] = [];
const gitEnvWithoutUserConfig = { ...process.env, GIT_CONFIG_GLOBAL: devNull, GIT_CONFIG_NOSYSTEM: '1' };

const repoWith = ({ branches, head }: { branches: string[]; head?: string }) => {
    const dir = mkdtempSync(join(tmpdir(), 'default-base-ref-'));
    const git = (...args: string[]) =>
        execFileSync('git', args, { cwd: dir, env: gitEnvWithoutUserConfig, stdio: 'pipe' });

    repos.push(dir);
    git('init', '-q');
    git('-c', 'user.email=e2e@example.com', '-c', 'user.name=e2e', 'commit', '-q', '--allow-empty', '-m', 'root');

    for (const branch of branches) {
        git('update-ref', `refs/remotes/origin/${branch}`, 'HEAD');
    }

    if (head) {
        git('symbolic-ref', 'refs/remotes/origin/HEAD', `refs/remotes/origin/${head}`);
    }

    return dir;
};

afterEach(() => {
    for (const dir of repos.splice(0)) {
        rmSync(dir, { force: true, recursive: true });
    }
});

describe('defaultBaseRef', () => {
    it('names the branch the remote calls default, even when the other name exists too', () => {
        expect(defaultBaseRef(repoWith({ branches: ['main', 'master'], head: 'master' }))).toBe('origin/master');
        expect(defaultBaseRef(repoWith({ branches: ['main', 'master'], head: 'main' }))).toBe('origin/main');
    });

    it('falls back to the default-looking branch that exists when the clone has no origin/HEAD', () => {
        expect(defaultBaseRef(repoWith({ branches: ['main'] }))).toBe('origin/main');
        expect(defaultBaseRef(repoWith({ branches: ['master'] }))).toBe('origin/master');
    });

    it('answers nothing when there is no remote branch to compare with', () => {
        expect(defaultBaseRef(repoWith({ branches: [] }))).toBeNull();
    });
});

describe('the tool that consumes it', () => {
    it('exits 2 and prints no routes when there is no remote branch to diff against', () => {
        const noRemote = repoWith({ branches: [] });
        const { status, stderr, stdout } = spawnSync(
            join(__dirname, '..', '..', 'node_modules', '.bin', 'tsx'),
            [join(__dirname, 'affected.ts')],
            { encoding: 'utf8', env: { ...gitEnvWithoutUserConfig, GIT_DIR: join(noRemote, '.git') } },
        );

        expect({ status, stdout }).toEqual({ status: 2, stdout: '' });
        expect(stderr).not.toContain('fatal:');
    });
});
