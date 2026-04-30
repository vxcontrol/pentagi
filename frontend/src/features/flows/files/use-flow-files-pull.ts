import { useCallback, useState } from 'react';
import { toast } from 'sonner';
import { z } from 'zod';

import { api, getApiErrorMessage } from '@/lib/axios';

import { CONTAINER_TARGET_DIRECTORY, FLOW_FILES_PULL_API_PATH } from './flow-files-constants';
import { type FlowFilesResponse } from './flow-files-utils';

export const flowFilesPullFormSchema = z.object({
    containerPath: z.string().trim().min(1, { message: 'Container path cannot be empty' }),
    shouldOverwrite: z.boolean(),
});

export type FlowFilesPullFormValues = z.infer<typeof flowFilesPullFormSchema>;

interface UseFlowFilesPullParams {
    flowId: null | string;
    onSuccess: () => void;
}

interface UseFlowFilesPullResult {
    isPulling: boolean;
    pull: (values: FlowFilesPullFormValues) => Promise<boolean>;
}

const PULL_OVERWRITE_HINT = 'Entry already exists — enable "Overwrite" to replace it';

/**
 * Wraps the "pull from container" REST call with toast notifications and a loading flag.
 * Returns a `pull(values)` callback that resolves to `true` on success, so the dialog
 * can decide whether to close itself.
 */
export const useFlowFilesPull = ({ flowId, onSuccess }: UseFlowFilesPullParams): UseFlowFilesPullResult => {
    const [isPulling, setIsPulling] = useState(false);

    const pull = useCallback(
        async ({ containerPath, shouldOverwrite }: FlowFilesPullFormValues): Promise<boolean> => {
            if (!flowId) {
                return false;
            }

            setIsPulling(true);

            try {
                await api.post<FlowFilesResponse>(
                    FLOW_FILES_PULL_API_PATH(flowId),
                    {
                        force: shouldOverwrite,
                        path: containerPath,
                    },
                    // Copying a directory out of the container can take longer than the default 30s
                    // (large logs, deep trees) — disable the per-call timeout entirely.
                    { timeout: 0 },
                );

                toast.success('Pulled from container', {
                    description: `Saved to local cache under ${CONTAINER_TARGET_DIRECTORY}`,
                });
                onSuccess();

                return true;
            } catch (error) {
                const description = getApiErrorMessage(error, 'Failed to pull from container', {
                    409: PULL_OVERWRITE_HINT,
                });

                toast.error('Pull failed', { description });

                return false;
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
};
