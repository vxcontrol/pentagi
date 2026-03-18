import { execSync } from 'node:child_process';

export const getGitHash = () => {
    // Check for environment variable first (useful in CI/CD environments)
    if (process.env.GIT_COMMIT_SHA) {
        return process.env.GIT_COMMIT_SHA;
    }
    
    try {
        return execSync('git rev-parse HEAD', { stdio: 'pipe' }).toString().trim();
    } catch {
        // Return a fallback value when not in a git repository
        return 'development';
    }
};
