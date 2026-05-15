import type { FileNode } from '@/components/shared/file-manager';
import type { OverwriteConflict } from '@/components/shared/overwrite';

import { CONTAINER_PATH_PREFIX, RESOURCES_PATH_PREFIX } from './flow-files-constants';

const splitName = (path: string): string => path.split('/').pop() ?? path;

/**
 * Convert an absolute container path (`/work/foo.txt`) into the cache-relative
 * path under which the Pull endpoint stores it (`container/work/foo.txt`).
 *
 * Mirrors the server-side `SanitizeContainerCachePath` normalisation:
 * leading slashes are stripped, the rest is appended verbatim under the
 * synthetic `container/` group prefix used in flow file paths. An empty
 * result is impossible: pulling the literal `/` is rejected by the backend.
 */
const containerPathToCachePath = (containerPath: string): string => {
    const rel = containerPath.replace(/^\/+/, '');

    return `${CONTAINER_PATH_PREFIX}/${rel}`;
};

/**
 * Local preflight for the Pull dialog: cross-check the absolute container
 * paths the user is about to pull against the flow's existing cache mirror.
 *
 * Only **exact** root-path collisions are reported — covers the common case
 * (re-pulling a file the user already has) without an extra REST round-trip.
 * Nested conflicts (the user pulls `/etc/` while only `/etc/nginx.conf` is
 * cached) still surface server-side as a 409 and are auto-redialed by the
 * caller through the same `OverwriteDialog` flow.
 */
export const findPullConflicts = (
    pullTargets: readonly string[],
    cachedFiles: readonly FileNode[],
): OverwriteConflict[] => {
    if (pullTargets.length === 0 || cachedFiles.length === 0) {
        return [];
    }

    const cachedPaths = new Set(cachedFiles.map((file) => file.path));
    const conflicts: OverwriteConflict[] = [];

    for (const containerPath of pullTargets) {
        const cachePath = containerPathToCachePath(containerPath);

        if (cachedPaths.has(cachePath)) {
            conflicts.push({
                destination: cachePath,
                destinationName: splitName(cachePath),
            });
        }
    }

    return conflicts;
};

/**
 * Local preflight for the Attach Resources dialog. Each picked resource is
 * copied into `resources/<resource_path>` inside the flow cache, so we
 * cross-check the destination paths against the existing flow files.
 *
 * The caller passes the resource virtual paths (`foo/bar.txt`) — exactly the
 * shape stored in the resource library — and the deduped list of flow files.
 * Directory resources never collide on their own (the backend always
 * `MkdirAll`s parents), so we keep the check focused on file paths.
 */
export const findAttachConflicts = (
    selectedResourcePaths: readonly string[],
    cachedFiles: readonly FileNode[],
): OverwriteConflict[] => {
    if (selectedResourcePaths.length === 0 || cachedFiles.length === 0) {
        return [];
    }

    const flowResourcePaths = new Set<string>();

    for (const file of cachedFiles) {
        if (!file.isDir && file.path.startsWith(`${RESOURCES_PATH_PREFIX}/`)) {
            flowResourcePaths.add(file.path);
        }
    }

    if (flowResourcePaths.size === 0) {
        return [];
    }

    const conflicts: OverwriteConflict[] = [];

    for (const resourcePath of selectedResourcePaths) {
        const destination = `${RESOURCES_PATH_PREFIX}/${resourcePath}`;

        if (flowResourcePaths.has(destination)) {
            conflicts.push({
                destination,
                destinationName: splitName(resourcePath),
            });
        }
    }

    return conflicts;
};
