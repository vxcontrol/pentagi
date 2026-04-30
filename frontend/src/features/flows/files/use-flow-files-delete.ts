import { useCallback, useState } from 'react';
import { toast } from 'sonner';

import type { FileNode } from '@/components/file-manager';

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
    deleteBulk: (filesToDelete: FileNode[]) => Promise<void>;
    fileToDelete: FileNode | null;
    requestDelete: (file: FileNode) => void;
}

const deleteFileRequest = (flowId: string, path: string) =>
    api.delete<FlowFilesResponse>(FLOW_FILES_API_PATH(flowId), { params: { path } });

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

    const confirmDelete = useCallback(async () => {
        if (!flowId || !fileToDelete) {
            return;
        }

        try {
            await deleteFileRequest(flowId, fileToDelete.path);
            toast.success(fileToDelete.isDir ? 'Directory deleted' : 'File deleted');
            await refetchFiles();
        } catch (error) {
            const description = getApiErrorMessage(error, 'Failed to delete file');

            toast.error('Delete failed', { description });
        } finally {
            setFileToDelete(null);
        }
    }, [flowId, fileToDelete, refetchFiles]);

    const deleteBulk = useCallback(
        async (filesToDelete: FileNode[]) => {
            if (!flowId || filesToDelete.length === 0) {
                return;
            }

            const results = await Promise.allSettled(
                filesToDelete.map((file) => deleteFileRequest(flowId, file.path)),
            );
            const succeededCount = results.filter((result) => result.status === 'fulfilled').length;
            const failedCount = results.length - succeededCount;

            reportBulkDeleteOutcome(succeededCount, failedCount);

            await refetchFiles();
        },
        [flowId, refetchFiles],
    );

    return {
        clearFileToDelete,
        confirmDelete,
        deleteBulk,
        fileToDelete,
        requestDelete,
    };
};
