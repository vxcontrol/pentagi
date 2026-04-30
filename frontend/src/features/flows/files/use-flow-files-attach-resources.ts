import { useCallback, useState } from 'react';
import { toast } from 'sonner';

import { api, getApiErrorMessage } from '@/lib/axios';

import type { FlowFilesResponse } from './flow-files-utils';

import { FLOW_FILES_ATTACH_RESOURCES_API_PATH, RESOURCES_TARGET_DIRECTORY } from './flow-files-constants';

interface AttachOptions {
    ids: string[];
    shouldOverwrite: boolean;
}

interface AttachResourcesRequestBody {
    force: boolean;
    ids: string[];
}

interface UseFlowFilesAttachResourcesParams {
    flowId: null | string;
    onSuccess?: () => void;
}

interface UseFlowFilesAttachResourcesResult {
    attach: (options: AttachOptions) => Promise<boolean>;
    isAttaching: boolean;
}

const ATTACH_OVERWRITE_HINT = 'Some files already exist — enable "Overwrite" to replace them';

/**
 * Wraps the "attach user resources to flow" REST call (`POST /files/resources`)
 * with toast notifications and a loading flag. Resolves to `true` on success so
 * the caller dialog can close itself on a successful attach.
 */
export const useFlowFilesAttachResources = ({
    flowId,
    onSuccess,
}: UseFlowFilesAttachResourcesParams): UseFlowFilesAttachResourcesResult => {
    const [isAttaching, setIsAttaching] = useState(false);

    const attach = useCallback(
        async ({ ids, shouldOverwrite }: AttachOptions): Promise<boolean> => {
            if (!flowId || ids.length === 0) {
                return false;
            }

            setIsAttaching(true);

            try {
                await api.post<FlowFilesResponse, AttachResourcesRequestBody>(
                    FLOW_FILES_ATTACH_RESOURCES_API_PATH(flowId),
                    {
                        force: shouldOverwrite,
                        ids,
                    },
                    { timeout: 0 },
                );

                toast.success('Resources attached', {
                    description: `Copied ${ids.length} ${ids.length === 1 ? 'item' : 'items'} to ${RESOURCES_TARGET_DIRECTORY}`,
                });
                onSuccess?.();

                return true;
            } catch (error) {
                const description = getApiErrorMessage(error, 'Failed to attach resources', {
                    409: ATTACH_OVERWRITE_HINT,
                });

                toast.error('Attach failed', { description });

                return false;
            } finally {
                setIsAttaching(false);
            }
        },
        [flowId, onSuccess],
    );

    return {
        attach,
        isAttaching,
    };
};
