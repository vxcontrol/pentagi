import { useCallback, useState } from 'react';
import { toast } from 'sonner';

import type { OverwriteOutcome } from '@/components/shared/overwrite';

import { resourceIdsToWire } from '@/features/resources/resources-rest';
import { api, getApiErrorMessage, getApiErrorStatusCode } from '@/lib/axios';

import type { FlowFilesResponse } from './flow-files-utils';

import { FLOW_FILES_ATTACH_RESOURCES_API_PATH, RESOURCES_TARGET_DIRECTORY } from './flow-files-constants';

interface AttachOptions {
    /** GraphQL `UserResource.id` values (strings). Converted at the REST boundary. */
    ids: string[];
    shouldOverwrite: boolean;
}

interface UseFlowFilesAttachResourcesParams {
    flowId: null | string;
    onSuccess?: () => void;
}

interface UseFlowFilesAttachResourcesResult {
    attach: (options: AttachOptions) => Promise<OverwriteOutcome>;
    isAttaching: boolean;
}

/**
 * Wraps the "attach user resources to flow" REST call (`POST /files/resources`)
 * with toast notifications and a loading flag. Returns a discriminated outcome
 * so callers can branch between success, a 409 conflict that warrants a user
 * prompt, and any other failure (already toasted).
 */
export function useFlowFilesAttachResources({
    flowId,
    onSuccess,
}: UseFlowFilesAttachResourcesParams): UseFlowFilesAttachResourcesResult {
    const [isAttaching, setIsAttaching] = useState(false);

    const attach = useCallback(
        async ({ ids, shouldOverwrite }: AttachOptions): Promise<OverwriteOutcome> => {
            if (!flowId || ids.length === 0) {
                return { kind: 'error' };
            }

            let numericIds: number[];

            try {
                numericIds = resourceIdsToWire(ids);
            } catch (error) {
                // Non-numeric IDs indicate an upstream cache contract bug, not a user
                // mistake. Surface a developer-friendly toast and bail out loudly.
                const description = error instanceof Error ? error.message : 'Invalid resource IDs.';

                toast.error('Attach failed', { description });

                return { kind: 'error' };
            }

            setIsAttaching(true);

            try {
                await api.post<FlowFilesResponse, { force: boolean; ids: number[] }>(
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

                return { kind: 'ok' };
            } catch (error) {
                if (!shouldOverwrite && getApiErrorStatusCode(error) === 409) {
                    return { kind: 'conflict' };
                }

                const description = getApiErrorMessage(error, 'Failed to attach resources');

                toast.error('Attach failed', { description });

                return { kind: 'error' };
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
}
