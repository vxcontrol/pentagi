import { useEffect, useMemo } from 'react';
import { toast } from 'sonner';

import type { FileNode } from '@/components/file-manager';

import { useFlowFilesQuery } from '@/graphql/types';

import { toFileNode } from './flow-files-utils';

interface UseFlowFilesDataParams {
    flowId: null | string;
}

interface UseFlowFilesDataResult {
    fileNodes: FileNode[];
    isInitialLoading: boolean;
    isLoading: boolean;
    refetchFiles: () => Promise<unknown>;
}

const FLOW_FILES_ERROR_TOAST_ID = 'flow-files-error';

/**
 * Loads `flowFiles` for the current flow and converts them into `FileNode`s for the
 * file manager. The `isInitialLoading` flag is derived from Apollo's response shape:
 * it stays `true` only while the very first response is in flight (no cached data
 * yet), so subsequent background `refetch` calls do not flash the skeleton.
 */
export const useFlowFilesData = ({ flowId }: UseFlowFilesDataParams): UseFlowFilesDataResult => {
    const flowFilesVariables = useMemo(() => ({ flowId: flowId ?? '' }), [flowId]);

    const {
        data: flowFilesData,
        error: flowFilesError,
        loading: isLoading,
        refetch,
    } = useFlowFilesQuery({
        skip: !flowId,
        variables: flowFilesVariables,
    });

    useEffect(() => {
        if (flowFilesError) {
            toast.error('Failed to load files', {
                description: flowFilesError.message,
                id: FLOW_FILES_ERROR_TOAST_ID,
            });
        }
    }, [flowFilesError]);

    const fileNodes = useMemo<FileNode[]>(
        () => (flowFilesData?.flowFiles ?? []).map(toFileNode),
        [flowFilesData?.flowFiles],
    );

    const refetchFiles = useMemo(() => () => refetch(), [refetch]);

    const isInitialLoading = isLoading && flowFilesData === undefined;

    return {
        fileNodes,
        isInitialLoading,
        isLoading,
        refetchFiles,
    };
};
