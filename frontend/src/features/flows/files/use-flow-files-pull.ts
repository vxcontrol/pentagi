import { useCallback, useState } from 'react';
import { toast } from 'sonner';

import type { OverwriteOutcome } from '@/components/shared/overwrite';

import { api, getApiErrorMessage, getApiErrorStatusCode } from '@/lib/axios';

import { CONTAINER_TARGET_DIRECTORY, FLOW_FILES_PULL_API_PATH } from './flow-files-constants';
import { type FlowFilesResponse, pluralizeItems } from './flow-files-utils';

interface UseFlowFilesPullParams {
    flowId: null | string;
    onSuccess: () => void;
}

interface UseFlowFilesPullResult {
    isPulling: boolean;
    /**
     * Pulls one or more absolute container paths into the local cache. Returns
     * a discriminated outcome so callers can branch between success, a 409
     * conflict that warrants a user prompt, and any other failure (toasted).
     * The backend deduplicates and validates atomically (fail-fast on cache
     * conflicts unless `force=true`).
     */
    pull: (paths: readonly string[], force: boolean) => Promise<OverwriteOutcome>;
}

/**
 * Wraps the "pull from container" REST call. The hook is path-agnostic — it
 * always sends the array as `paths` in the body (single-path callers just pass
 * a 1-element array), matching the wire shape the rest of our file APIs use.
 */
export function useFlowFilesPull({ flowId, onSuccess }: UseFlowFilesPullParams): UseFlowFilesPullResult {
    const [isPulling, setIsPulling] = useState(false);

    const pull = useCallback(
        async (paths: readonly string[], force: boolean): Promise<OverwriteOutcome> => {
            if (!flowId || paths.length === 0) {
                return { kind: 'error' };
            }

            setIsPulling(true);

            try {
                await api.post<FlowFilesResponse>(
                    FLOW_FILES_PULL_API_PATH(flowId),
                    {
                        force,
                        paths,
                    },
                    // Copying a directory out of the container can take longer than the default 30s
                    // (large logs, deep trees) — disable the per-call timeout entirely.
                    { timeout: 0 },
                );

                const description =
                    paths.length === 1
                        ? `Saved to local cache under ${CONTAINER_TARGET_DIRECTORY}`
                        : `Saved ${paths.length} ${pluralizeItems(paths.length)} to local cache under ${CONTAINER_TARGET_DIRECTORY}`;

                toast.success('Pulled from container', { description });
                onSuccess();

                return { kind: 'ok' };
            } catch (error) {
                if (!force && getApiErrorStatusCode(error) === 409) {
                    // The caller resolves conflicts through a richer dialog, so we
                    // intentionally skip the toast here to avoid double-surfacing.
                    return { kind: 'conflict' };
                }

                const description = getApiErrorMessage(error, 'Failed to pull from container');

                toast.error('Pull failed', { description });

                return { kind: 'error' };
            } finally {
                setIsPulling(false);
            }
        },
        [flowId, onSuccess],
    );

    return {
        isPulling,
        pull,
    };
}
