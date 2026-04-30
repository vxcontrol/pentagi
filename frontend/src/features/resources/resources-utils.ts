import type { FileNode } from '@/components/file-manager';
import type { UserResourceFragmentFragment } from '@/graphql/types';

import { baseUrl } from '@/models/api';

import { RESOURCES_DOWNLOAD_API_PATH } from './resources-constants';

/** Convert a GraphQL `UserResource` into a `FileNode` consumed by the FileManager. */
export const toFileNode = (resource: UserResourceFragmentFragment): FileNode => ({
    id: resource.id,
    isDir: resource.isDir,
    modifiedAt: resource.updatedAt,
    name: resource.name,
    path: resource.path,
    size: resource.size,
});

/** Absolute URL used by the FileManager download action. */
export const buildResourceDownloadHref = (file: FileNode): string =>
    `${baseUrl}${RESOURCES_DOWNLOAD_API_PATH}?path=${encodeURIComponent(file.path)}`;

export const pluralizeItems = (count: number): string => (count === 1 ? 'item' : 'items');
