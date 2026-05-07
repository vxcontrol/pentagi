import { useCallback, useState } from 'react';
import { toast } from 'sonner';

import type { FileNode } from '@/components/shared/file-manager';

import { buildPathsQuery } from '@/features/resources/resources-utils';
import { api, getApiErrorMessage } from '@/lib/axios';

import { FLOW_FILES_API_PATH } from './flow-files-constants';
import { type FlowFilesResponse, pluralizeItems } from './flow-files-utils';

interface UseFlowFilesDeleteParams {
    flowId: null | string;
    refetchFiles: () => Promise<unknown>;
}

interface UseFlowFilesDeleteResult {
    clearFileToDelete: () => void;
    confirmDelete: () => Promise<void>;
    /**
     * Delete one or more flow files in a single atomic backend call. Both the
     * row-level "Delete" action and the bulk-bar entry funnel through this
     * method — the row flow shows a one-name confirm dialog first via
     * `requestDelete` / `confirmDelete`.
     */
    deleteFiles: (filesToDelete: readonly FileNode[]) => Promise<void>;
    fileToDelete: FileNode | null;
    requestDelete: (file: FileNode) => void;
}

/**
 * Issues one `DELETE /flows/{id}/files?paths[]=…` regardless of how many files
 * are passed in. The backend deduplicates and validates atomically: either every
 * path is removed or nothing is. The response enumerates every entry that was
 * actually deleted (including nested files for directory removals) — we still
 * trigger an explicit refetch because flow files don't have a deletion
 * subscription wired into Apollo yet.
 */
const deleteFlowFilesRequest = (flowId: string, paths: readonly string[]) =>
    api.delete<FlowFilesResponse>(`${FLOW_FILES_API_PATH(flowId)}?${buildPathsQuery(paths)}`);

/**
 * Owns both the single-file and bulk-delete flows. The component drives the
 * confirmation dialog state through the returned `fileToDelete`/`requestDelete`/
 * `clearFileToDelete` triple, while the hook hides every API call, toast and
 * post-delete refetch.
 */
export const useFlowFilesDelete = ({ flowId, refetchFiles }: UseFlowFilesDeleteParams): UseFlowFilesDeleteResult => {
    const [fileToDelete, setFileToDelete] = useState<FileNode | null>(null);

    const requestDelete = useCallback((file: FileNode) => {
        setFileToDelete(file);
    }, []);

    const clearFileToDelete = useCallback(() => {
        setFileToDelete(null);
    }, []);

    const deleteFiles = useCallback(
        async (filesToDelete: readonly FileNode[]) => {
            if (!flowId || filesToDelete.length === 0) {
                return;
            }

            try {
                await deleteFlowFilesRequest(
                    flowId,
                    filesToDelete.map((file) => file.path),
                );

                if (filesToDelete.length === 1) {
                    const single = filesToDelete[0];
                    toast.success(single.isDir ? 'Directory deleted' : 'File deleted');
                } else {
                    toast.success(`${filesToDelete.length} ${pluralizeItems(filesToDelete.length)} deleted`);
                }
            } catch (error) {
                const description = getApiErrorMessage(error, 'Failed to delete file');

                toast.error('Delete failed', { description });
            } finally {
                // Always refetch — succeeded paths now leave the cache, failed
                // paths may have been partially recreated by other clients.
                await refetchFiles();
            }
        },
        [flowId, refetchFiles],
    );

    const confirmDelete = useCallback(async () => {
        if (!fileToDelete) {
            return;
        }

        try {
            await deleteFiles([fileToDelete]);
        } finally {
            setFileToDelete(null);
        }
    }, [deleteFiles, fileToDelete]);

    return {
        clearFileToDelete,
        confirmDelete,
        deleteFiles,
        fileToDelete,
        requestDelete,
    };
};
