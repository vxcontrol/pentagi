import { useCallback, useState } from 'react';
import { toast } from 'sonner';
import { z } from 'zod';

import type { RestResourceList } from '@/features/resources/resources-rest';

import { api, getApiErrorMessage } from '@/lib/axios';

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
    source: string;
}

interface UseFlowFilesPromoteParams {
    flowId: null | string;
}

interface UseFlowFilesPromoteResult {
    isPromoting: boolean;
    promote: (source: string, values: FlowFilesPromoteFormValues) => Promise<boolean>;
}

const PROMOTE_OVERWRITE_HINT = 'Resource already exists — enable "Overwrite" to replace it';

/**
 * Wraps the "promote flow file → user resource" REST call (`POST /files/to-resources`)
 * with toast notifications and a loading flag. Returns a `promote(source, values)`
 * callback that resolves to `true` on success so the dialog can decide whether to close.
 *
 * The endpoint accepts both single files and directories — the backend always returns
 * a `ResourceList` covering every entry it created or updated. The response payload
 * itself is discarded: the resource library Apollo cache is kept in sync via the
 * `resourceAdded` / `resourceUpdated` GraphQL subscriptions.
 */
export const useFlowFilesPromote = ({ flowId }: UseFlowFilesPromoteParams): UseFlowFilesPromoteResult => {
    const [isPromoting, setIsPromoting] = useState(false);

    const promote = useCallback(
        async (source: string, { destination, shouldOverwrite }: FlowFilesPromoteFormValues): Promise<boolean> => {
            if (!flowId) {
                return false;
            }

            setIsPromoting(true);

            try {
                await api.post<RestResourceList, PromoteRequestBody>(
                    FLOW_FILES_PROMOTE_API_PATH(flowId),
                    {
                        destination: destination.trim(),
                        force: shouldOverwrite,
                        source,
                    },
                    { timeout: 0 },
                );

                toast.success('Saved to resources', {
                    description: `Stored at ${destination.trim()} in your resource library`,
                });

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
        [flowId],
    );

    return {
        isPromoting,
        promote,
    };
};
