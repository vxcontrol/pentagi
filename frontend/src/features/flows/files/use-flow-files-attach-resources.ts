import { useCallback, useState } from 'react';
import { toast } from 'sonner';

import { resourceIdsToWire } from '@/features/resources/resources-rest';
import { api, getApiErrorMessage } from '@/lib/axios';

import type { FlowFilesResponse } from './flow-files-utils';

import { FLOW_FILES_ATTACH_RESOURCES_API_PATH, RESOURCES_TARGET_DIRECTORY } from './flow-files-constants';

interface AttachOptions {
    /** GraphQL `UserResource.id` values (strings). Converted at the REST boundary. */
    ids: string[];
    shouldOverwrite: boolean;
}

/**
 * Wire shape of `models.AddResourcesRequest`. Backend expects `ids` as
 * uint64 (JSON numbers); see {@link resourceIdsToWire} for the bridge from
 * the GraphQL `string` ID convention used internally.
 */
interface AttachResourcesRequestBody {
    force: boolean;
    ids: number[];
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

            let numericIds: number[];

            try {
                numericIds = resourceIdsToWire(ids);
            } catch (error) {
                // Non-numeric IDs indicate an upstream cache contract bug, not a user
                // mistake. Surface a developer-friendly toast and bail out loudly.
                const description = error instanceof Error ? error.message : 'Invalid resource IDs.';

                toast.error('Attach failed', { description });

                return false;
            }

            setIsAttaching(true);

            try {
                await api.post<FlowFilesResponse, AttachResourcesRequestBody>(
                    FLOW_FILES_ATTACH_RESOURCES_API_PATH(flowId),
                    {
                        force: shouldOverwrite,
                        ids: numericIds,
                    },
                    { timeout: 0 },
                );

                toast.success('Resources attached', {
                    description: `Copied ${numericIds.length} ${numericIds.length === 1 ? 'item' : 'items'} to ${RESOURCES_TARGET_DIRECTORY}`,
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
