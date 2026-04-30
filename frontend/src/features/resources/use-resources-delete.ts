import { useCallback, useState } from 'react';
import { toast } from 'sonner';

import type { FileNode } from '@/components/file-manager';

import { api, getApiErrorMessage } from '@/lib/axios';

import { RESOURCES_API_PATH } from './resources-constants';
import { pluralizeItems } from './resources-utils';

interface UseResourcesDeleteParams {
    onAfterDelete?: () => void;
}

interface UseResourcesDeleteResult {
    clearFileToDelete: () => void;
    confirmDelete: () => Promise<void>;
    deleteBulk: (filesToDelete: FileNode[]) => Promise<void>;
    fileToDelete: FileNode | null;
    requestDelete: (file: FileNode) => void;
}

const deleteResourceRequest = (path: string) => api.delete<void>(RESOURCES_API_PATH, { params: { path } });

const reportBulkDeleteOutcome = (succeededCount: number, failedCount: number): void => {
    if (failedCount === 0) {
        toast.success(`${succeededCount} ${pluralizeItems(succeededCount)} deleted`);

        return;
    }

    if (succeededCount === 0) {
        toast.error('Bulk delete failed', {
            description: `Failed to delete ${failedCount} ${pluralizeItems(failedCount)}`,
        });

        return;
    }

    toast.warning(`${succeededCount} succeeded · ${failedCount} failed`);
};

/**
 * Owns both the single-resource and bulk-delete flows. Component drives the
 * confirmation dialog state through the returned `fileToDelete`/`requestDelete`/
 * `clearFileToDelete` triple, while the hook hides every API call and toast.
 *
 * No imperative refetch is performed: the GraphQL `resourceDeleted` subscription
 * is wired into the Apollo cache and removes the deleted entries automatically.
 */
export const useResourcesDelete = ({ onAfterDelete }: UseResourcesDeleteParams = {}): UseResourcesDeleteResult => {
    const [fileToDelete, setFileToDelete] = useState<FileNode | null>(null);

    const requestDelete = useCallback((file: FileNode) => {
        setFileToDelete(file);
    }, []);

    const clearFileToDelete = useCallback(() => {
        setFileToDelete(null);
    }, []);

    const confirmDelete = useCallback(async () => {
        if (!fileToDelete) {
            return;
        }

        try {
            await deleteResourceRequest(fileToDelete.path);
            toast.success(fileToDelete.isDir ? 'Directory deleted' : 'Resource deleted');
            onAfterDelete?.();
        } catch (error) {
            const description = getApiErrorMessage(error, 'Failed to delete resource');

            toast.error('Delete failed', { description });
        } finally {
            setFileToDelete(null);
        }
    }, [fileToDelete, onAfterDelete]);

    const deleteBulk = useCallback(
        async (filesToDelete: FileNode[]) => {
            if (filesToDelete.length === 0) {
                return;
            }

            const results = await Promise.allSettled(filesToDelete.map((file) => deleteResourceRequest(file.path)));
            const succeededCount = results.filter((result) => result.status === 'fulfilled').length;
            const failedCount = results.length - succeededCount;

            reportBulkDeleteOutcome(succeededCount, failedCount);
            onAfterDelete?.();
        },
        [onAfterDelete],
    );

    return {
        clearFileToDelete,
        confirmDelete,
        deleteBulk,
        fileToDelete,
        requestDelete,
    };
};
