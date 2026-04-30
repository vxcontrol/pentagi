import type { FileNode } from '@/components/file-manager';
import type { FlowFileFragmentFragment } from '@/graphql/types';

import { baseUrl } from '@/models/api';

export type FlowFile = FlowFileFragmentFragment;

export interface FlowFilesResponse {
    files: Array<FlowFile>;
    total: number;
}

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
