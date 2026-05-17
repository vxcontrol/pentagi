import { useCallback, useState } from 'react';
import { toast } from 'sonner';

import type { FileNode } from '@/components/shared/file-manager';

import { api, getApiErrorMessage } from '@/lib/axios';

import { RESOURCES_API_PATH } from './resources-constants';
import { buildPathsQuery, pluralizeItems } from './resources-utils';

interface UseResourcesDeleteParams {
    onAfterDelete?: () => void;
}

interface UseResourcesDeleteResult {
    clearFileToDelete: () => void;
    confirmDelete: () => Promise<void>;
    /**
     * Delete one or more resources in a single atomic backend call. Both the
     * row-level "Delete" action and the bulk-bar entry funnel through this
     * method — the only difference is that the row flow shows a one-name
     * confirm dialog first via `requestDelete` / `confirmDelete`.
     */
    deleteFiles: (filesToDelete: readonly FileNode[]) => Promise<void>;
    fileToDelete: FileNode | null;
    requestDelete: (file: FileNode) => void;
}

/**
 * Issues one `DELETE /resources/?paths[]=…` regardless of how many files are
 * passed in. The backend deduplicates and validates atomically: either every
 * path is removed or nothing is, so callers don't have to reconcile partial
 * failures. The response payload enumerates every entry that was actually
 * removed — but we ignore it here because the GraphQL `resourceDeleted`
 * subscription already drives the Apollo cache update.
 */
const deleteResourcesRequest = (paths: readonly string[]) =>
    api.delete<void>(`${RESOURCES_API_PATH}?${buildPathsQuery(paths)}`);

/**
 * Owns both the single-resource and bulk-delete flows. The component drives the
 * confirmation dialog state through the returned `fileToDelete`/`requestDelete`/
 * `clearFileToDelete` triple, while the hook hides every API call and toast.
 *
 * No imperative refetch is performed: the GraphQL `resourceDeleted` subscription
 * is wired into the Apollo cache and removes the deleted entries automatically.
 */
export function useResourcesDelete({ onAfterDelete }: UseResourcesDeleteParams = {}): UseResourcesDeleteResult {
    const [fileToDelete, setFileToDelete] = useState<FileNode | null>(null);

    const requestDelete = useCallback((file: FileNode) => {
        setFileToDelete(file);
    }, []);

    const clearFileToDelete = useCallback(() => {
        setFileToDelete(null);
    }, []);

    const deleteFiles = useCallback(
        async (filesToDelete: readonly FileNode[]) => {
            if (filesToDelete.length === 0) {
                return;
            }

            try {
                await deleteResourcesRequest(filesToDelete.map((file) => file.path));

                if (filesToDelete.length === 1) {
                    const single = filesToDelete[0];
                    toast.success(single.isDir ? 'Directory deleted' : 'Resource deleted');
                } else {
                    toast.success(`${filesToDelete.length} ${pluralizeItems(filesToDelete.length)} deleted`);
                }

                onAfterDelete?.();
            } catch (error) {
                const description = getApiErrorMessage(error, 'Failed to delete resource');

                toast.error('Delete failed', { description });
            }
        },
        [onAfterDelete],
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
}
