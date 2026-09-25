import { execFileSync } from 'node:child_process';

export function defaultBaseRef(repoRoot: string): null | string {
    const git = (args: string[]) =>
        execFileSync('git', args, { cwd: repoRoot, stdio: ['ignore', 'pipe', 'ignore'] })
            .toString()
            .trim();

    try {
        return git(['symbolic-ref', '--short', 'refs/remotes/origin/HEAD']);
    } catch {
        for (const candidate of ['origin/master', 'origin/main']) {
            try {
                git(['rev-parse', '--verify', '--quiet', candidate]);

                return candidate;
            } catch {
                continue;
            }
        }

        return null;
    }
}
