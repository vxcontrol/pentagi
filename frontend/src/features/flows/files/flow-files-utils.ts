import type { FileNode } from '@/components/shared/file-manager';
import type { FlowFileFragmentFragment } from '@/graphql/types';

import { buildPathsQuery } from '@/features/resources/resources-utils';
import { baseUrl } from '@/models/api';

import { CONTAINER_PATH_PREFIX, RESOURCES_PATH_PREFIX, UPLOADS_PATH_PREFIX } from './flow-files-constants';

/**
 * Wire shape of `models.ContainerFiles`. `path` echoes back the queried path
 * when exactly one was requested — empty string for multi-path queries.
 */
export interface ContainerFileFailure {
    message: string;
    name: string;
    path: string;
}

export interface ContainerFilesResponse {
    failures?: ContainerFileFailure[];
    files: RestContainerFile[];
    path: string;
    total: number;
    /** True when the directory held more entries than the cap and only the first page was returned. */
    truncated?: boolean;
}

export type FlowFile = FlowFileFragmentFragment;

/**
 * Wire shape of `models.FlowFile` (REST JSON, snake_case). The internal
 * `FlowFile` alias mirrors the GraphQL camelCase fragment for use in the
 * FileManager UI.
 */
export interface FlowFilesResponse {
    files: RestFlowFile[];
    total: number;
}

/** Wire shape of `models.ContainerFile` — matches `RestFlowFile` exactly today. */
export interface RestContainerFile {
    id: string;
    is_dir: boolean;
    modified_at: string;
    name: string;
    path: string;
    size: number;
}

export interface RestFlowFile {
    id: string;
    is_dir: boolean;
    modified_at: string;
    name: string;
    path: string;
    size: number;
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
 * Build the absolute download URL for one or more flow files. Returns `null`
 * when no flow is selected so callers can disable the download UI without
 * checking `flowId` themselves. The backend decides the response shape on its
 * own (single file → attachment, single directory → `<dirname>.zip`, multiple
 * paths → `download.zip`); callers just pass the file list.
 */
export const buildFlowFilesDownloadHref = (flowId: null | string, files: readonly FileNode[]): null | string => {
    if (!flowId) {
        return null;
    }

    const query = buildPathsQuery(files.map((file) => file.path));

    return `${baseUrl}/flows/${flowId}/files/download?${query}`;
};

export const toFileNode = (file: FlowFile): FileNode => ({
    id: file.id,
    isDir: file.isDir,
    modifiedAt: file.modifiedAt,
    name: file.name,
    path: file.path,
    size: file.size,
});

/**
 * Convert a `RestContainerFile` (snake_case wire shape) into a `FileNode` the
 * FileManager can render. Container paths are absolute (`/work/foo.txt`); the
 * Pull dialog sends them straight back to `POST /files/pull?paths=…` so no
 * normalisation is needed here — we pass the path through verbatim.
 */
export const containerFileToFileNode = (file: RestContainerFile): FileNode => ({
    id: file.id,
    isDir: file.is_dir,
    modifiedAt: file.modified_at,
    name: file.name,
    path: file.path,
    size: file.size,
});

export const pluralizeItems = (count: number): string => (count === 1 ? 'item' : 'items');
