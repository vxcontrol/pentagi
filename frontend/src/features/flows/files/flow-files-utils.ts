import type { FileNode } from '@/components/file-manager';
import type { FlowFileFragmentFragment } from '@/graphql/types';

import { baseUrl } from '@/models/api';

import { CONTAINER_PATH_PREFIX, RESOURCES_PATH_PREFIX, UPLOADS_PATH_PREFIX } from './flow-files-constants';

export type FlowFile = FlowFileFragmentFragment;

export interface FlowFilesResponse {
    files: Array<FlowFile>;
    total: number;
}

/** Mirrors `models.ResourceEntry` from the backend (REST shape, not GraphQL). */
export interface ResourceEntry {
    createdAt: string;
    id: number;
    isDir: boolean;
    name: string;
    path: string;
    size: number;
    updatedAt: string;
    userId: number;
}

const ROOT_PREFIXES = [`${UPLOADS_PATH_PREFIX}/`, `${CONTAINER_PATH_PREFIX}/`, `${RESOURCES_PATH_PREFIX}/`];

/**
 * Strips the synthetic root group prefix (`uploads/`, `container/`, `resources/`)
 * from a flow-file path so callers can suggest a sensible default `destination`
 * when promoting the file into the user's global resource library.
 */
export const stripFlowRootPrefix = (path: string): string => {
    for (const prefix of ROOT_PREFIXES) {
        if (path.startsWith(prefix)) {
            return path.slice(prefix.length);
        }
    }

    return path;
};

/**
 * Builds the absolute URL the browser hits to download a single file or directory archive
 * for the given flow. Returns `null` when no flow is selected so callers can disable
 * the download UI without checking `flowId` themselves.
 */
export const buildDownloadHref = (flowId: null | string, file: FileNode): null | string => {
    if (!flowId) {
        return null;
    }

    return `${baseUrl}/flows/${flowId}/files/download?path=${encodeURIComponent(file.path)}`;
};

export const toFileNode = (file: FlowFile): FileNode => ({
    id: file.id,
    isDir: file.isDir,
    modifiedAt: file.modifiedAt,
    name: file.name,
    path: file.path,
    size: file.size,
});

export const pluralizeItems = (count: number): string => (count === 1 ? 'item' : 'items');
