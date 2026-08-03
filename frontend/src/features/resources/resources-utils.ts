import type { FileNode } from '@/components/shared/file-manager';
import type { UserResourceFragmentFragment } from '@/graphql/types';

import { baseUrl } from '@/models/api';

import { RESOURCES_DOWNLOAD_API_PATH } from './resources-constants';

export const toFileNode = (resource: UserResourceFragmentFragment): FileNode => ({
    id: resource.id,
    isDir: resource.isDir,
    modifiedAt: resource.updatedAt,
    name: resource.name,
    path: resource.path,
    size: resource.size,
});

/**
 * Serialise a list of paths into the `paths[]=…&paths[]=…` query string the
 * backend expects across every multi-path file API. Single-path callers pass
 * a 1-element array — there is no separate `?path=` form, the backend treats
 * `paths[]` as the sole canonical input.
 */
export const buildPathsQuery = (paths: readonly string[]): string => {
    const params = new URLSearchParams();

    for (const p of paths) {
        params.append('paths[]', p);
    }

    return params.toString();
};

/**
 * Build the absolute download URL for one or more resources. The backend decides
 * the response shape on its own:
 *   - 1 regular file  → `application/octet-stream` with `Content-Disposition`
 *     setting the original filename;
 *   - 1 directory     → `application/zip` named `<dirname>.zip`;
 *   - multiple paths  → `application/zip` named `download.zip`, entry names use
 *     the full virtual path so the user keeps directory context.
 *
 * Callers don't need to special-case any of the above — just pass the file list.
 */
export const buildResourcesDownloadHref = (files: readonly FileNode[]): string => {
    const query = buildPathsQuery(files.map((file) => file.path));

    return `${baseUrl}${RESOURCES_DOWNLOAD_API_PATH}?${query}`;
};

export const pluralizeItems = (count: number): string => (count === 1 ? 'item' : 'items');
