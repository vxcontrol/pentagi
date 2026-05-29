import { useCallback, useState } from 'react';
import { toast } from 'sonner';
import { z } from 'zod';

import type { OverwriteOutcome } from '@/components/shared/overwrite';
import type { RestResourceList } from '@/features/resources/resources-rest';

import { pluralizeItems } from '@/features/resources/resources-utils';
import { api, getApiErrorMessage, getApiErrorStatusCode } from '@/lib/axios';

import { FLOW_FILES_PROMOTE_API_PATH } from './flow-files-constants';

export const flowFilesPromoteFormSchema = z.object({
    destination: z
        .string()
        .trim()
        .min(1, { message: 'Destination cannot be empty' })
        .refine((value) => !value.startsWith('/'), { message: 'Destination must be a relative path' })
        .refine((value) => !value.split('/').includes('..'), { message: 'Destination must not contain ".."' }),
});

export type FlowFilesPromoteFormValues = z.infer<typeof flowFilesPromoteFormSchema>;

interface PromoteRequestBody {
    destination: string;
    force: boolean;
    sources: readonly string[];
}

interface UseFlowFilesPromoteParams {
    flowId: null | string;
}

interface UseFlowFilesPromoteResult {
    isPromoting: boolean;
    /**
     * Issue a batch promote in a single atomic request and return a discriminated outcome:
     *   - `ok`        — every flow file/dir was promoted (success toast already fired),
     *   - `conflict`  — at least one resource path is occupied (no toast, caller
     *                   resolves via the shared overwrite workflow),
     *   - `error`     — anything else (failure toast already fired).
     *
     * Backend semantics: with one source, `destination` is the exact target
     * path; with multiple sources, `destination` is a base directory and each
     * source lands at `destination/<basename>`. The Apollo cache stays in sync
     * via `resourceAdded` / `resourceUpdated` GraphQL subscriptions.
     */
    promote: (sources: readonly string[], destination: string, force: boolean) => Promise<OverwriteOutcome>;
}

/**
 * Wraps the "promote flow file → user resource" REST call (`POST /files/to-resources`)
 * with toast notifications and a loading flag.
 */
export function useFlowFilesPromote({ flowId }: UseFlowFilesPromoteParams): UseFlowFilesPromoteResult {
    const [isPromoting, setIsPromoting] = useState(false);

    const promote = useCallback(
        async (sources: readonly string[], destination: string, force: boolean): Promise<OverwriteOutcome> => {
            if (!flowId || sources.length === 0) {
                return { kind: 'error' };
            }

            setIsPromoting(true);

            try {
                await api.post<RestResourceList, PromoteRequestBody>(
                    FLOW_FILES_PROMOTE_API_PATH(flowId),
                    {
                        destination: destination.trim(),
                        force,
                        sources,
                    },
                    { timeout: 0 },
                );

                const description =
                    sources.length === 1
                        ? `Stored at ${destination.trim()} in your resource library`
                        : `Stored ${sources.length} ${pluralizeItems(sources.length)} under ${destination.trim()} in your resource library`;

                toast.success('Saved to resources', { description });

                return { kind: 'ok' };
            } catch (error) {
                if (!force && getApiErrorStatusCode(error) === 409) {
                    return { kind: 'conflict' };
                }

                const description = getApiErrorMessage(error, 'Failed to save resource');

                toast.error('Save as resource failed', { description });

                return { kind: 'error' };
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
}
