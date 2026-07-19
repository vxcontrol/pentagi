export interface RouteSource {
    path: string;
    sources: string[];
}

const owns = (dir: string, file: string): boolean => file === dir || file.startsWith(`${dir}/`);

/**
 * Maps a set of changed repo files to the manifest routes they touch, by
 * prefix-matching each change against a route's owning source dirs. This is the
 * substrate for selective runs and for scoping the exploratory agent to the
 * changed surface. Pure — the CLI and the vitest unit test both drive it.
 *
 * - no frontend change → no routes (a backend-only diff; the caller falls back
 *   to the full suite, and schema-compat covers the backend contract);
 * - a frontend change owned by no route → all routes (shared infra: the e2e
 *   engine, config, src/lib, a provider used everywhere) — conservative on
 *   purpose, since over-running is safe and under-running hides a regression;
 * - otherwise → the routes whose `sources` the changes fall under.
 *
 * Paths are repo-relative (`frontend/...`); `sources` are `frontend/`-relative.
 */
export const affectedRoutes = (changedFiles: string[], manifest: RouteSource[]): RouteSource[] => {
    const frontendChanges = changedFiles
        .filter((file) => file.startsWith('frontend/'))
        .map((file) => file.slice('frontend/'.length));

    if (!frontendChanges.length) {
        return [];
    }

    const isOwned = (file: string) => manifest.some((route) => route.sources.some((dir) => owns(dir, file)));

    if (frontendChanges.some((file) => !isOwned(file))) {
        return manifest;
    }

    return manifest.filter((route) => route.sources.some((dir) => frontendChanges.some((file) => owns(dir, file))));
};
