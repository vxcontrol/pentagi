export interface RouteSource {
    path: string;
    sources: string[];
}

const owns = (dir: string, file: string): boolean => file === dir || file.startsWith(`${dir}/`);

/**
 * Maps a set of changed repo files to the manifest routes they touch, by
 * prefix-matching each change against a route's owning source dirs.
 *
 * - no frontend change → no routes (the caller falls back to the full suite);
 * - a frontend change owned by no route → all routes (shared infra: e2e engine,
 *   config, src/lib, an everywhere-provider) — over-run rather than under-run;
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
