import { useCallback, useState } from 'react';
import { toast } from 'sonner';
import { z } from 'zod';

import { api, getApiErrorMessage } from '@/lib/axios';

import type { ResourceEntry } from './flow-files-utils';

import { FLOW_FILES_PROMOTE_API_PATH } from './flow-files-constants';

export const flowFilesPromoteFormSchema = z.object({
    destination: z
        .string()
        .trim()
        .min(1, { message: 'Destination cannot be empty' })
        .refine((value) => !value.startsWith('/'), { message: 'Destination must be a relative path' })
        .refine((value) => !value.split('/').includes('..'), { message: 'Destination must not contain ".."' }),
    shouldOverwrite: z.boolean(),
});

export type FlowFilesPromoteFormValues = z.infer<typeof flowFilesPromoteFormSchema>;

interface PromoteRequestBody {
    destination: string;
    force: boolean;
    sourcePath: string;
}

interface UseFlowFilesPromoteParams {
    flowId: null | string;
    onSuccess?: (entry: ResourceEntry) => void;
}

interface UseFlowFilesPromoteResult {
    isPromoting: boolean;
    promote: (sourcePath: string, values: FlowFilesPromoteFormValues) => Promise<boolean>;
}

const PROMOTE_OVERWRITE_HINT = 'Resource already exists — enable "Overwrite" to replace it';

/**
 * Wraps the "promote flow file → user resource" REST call (`POST /files/to-resources`)
 * with toast notifications and a loading flag. Returns a `promote(sourcePath, values)`
 * callback that resolves to `true` on success so the dialog can decide whether to close.
 */
export const useFlowFilesPromote = ({ flowId, onSuccess }: UseFlowFilesPromoteParams): UseFlowFilesPromoteResult => {
    const [isPromoting, setIsPromoting] = useState(false);

    const promote = useCallback(
        async (sourcePath: string, { destination, shouldOverwrite }: FlowFilesPromoteFormValues): Promise<boolean> => {
            if (!flowId) {
                return false;
            }

            setIsPromoting(true);

            try {
                const response = await api.post<ResourceEntry, PromoteRequestBody>(
                    FLOW_FILES_PROMOTE_API_PATH(flowId),
                    {
                        destination: destination.trim(),
                        force: shouldOverwrite,
                        sourcePath,
                    },
                    { timeout: 0 },
                );

                toast.success('Saved to resources', {
                    description: `Stored at ${destination.trim()} in your resource library`,
                });

                if (response.status === 'success' && response.data) {
                    onSuccess?.(response.data);
                }

                return true;
            } catch (error) {
                const description = getApiErrorMessage(error, 'Failed to save resource', {
                    409: PROMOTE_OVERWRITE_HINT,
                });

                toast.error('Save as resource failed', { description });

                return false;
            } finally {
                setIsPromoting(false);
            }
        },
        [flowId, onSuccess],
    );

    return {
        isPromoting,
        promote,
    };
};
