import { useCallback, useState } from 'react';
import { toast } from 'sonner';
import { z } from 'zod';

import type { OverwriteOutcome } from '@/components/shared/overwrite';

import { api, getApiErrorMessage, getApiErrorStatusCode } from '@/lib/axios';

import { RESOURCES_COPY_API_PATH } from './resources-constants';
import { pluralizeItems } from './resources-utils';

export const resourcesCopyFormSchema = z.object({
    destination: z
        .string()
        .trim()
        .min(1, { message: 'Destination cannot be empty' })
        .refine((value) => !value.startsWith('/'), { message: 'Destination must be a relative path' })
        .refine((value) => !value.split('/').includes('..'), { message: 'Destination must not contain ".."' }),
});

export type ResourcesCopyFormValues = z.infer<typeof resourcesCopyFormSchema>;

interface CopyRequestBody {
    destination: string;
    force: boolean;
    sources: readonly string[];
}

interface UseResourcesCopyResult {
    /**
     * Issue a batch copy in a single atomic request and return a discriminated outcome:
     *   - `ok`        — every source was copied (success toast already fired),
     *   - `conflict`  — at least one destination is occupied (no toast, caller
     *                   resolves via the shared overwrite workflow),
     *   - `error`     — anything else (failure toast already fired).
     *
     * Backend semantics: with one source, `destination` is the exact target
     * path; with multiple sources, `destination` is a base directory and each
     * source lands at `destination/<basename>`. Either way, the whole batch
     * executes inside one DB transaction.
     */
    copy: (sources: readonly string[], destination: string, force: boolean) => Promise<OverwriteOutcome>;
    isCopying: boolean;
}

/** Wraps `POST /resources/copy` for single and batch copy operations. */
export function useResourcesCopy(): UseResourcesCopyResult {
    const [isCopying, setIsCopying] = useState(false);

    const copy = useCallback(
        async (sources: readonly string[], destination: string, force: boolean): Promise<OverwriteOutcome> => {
            if (sources.length === 0) {
                return { kind: 'error' };
            }

            setIsCopying(true);

            try {
                await api.post<void, CopyRequestBody>(RESOURCES_COPY_API_PATH, {
                    destination,
                    force,
                    sources,
                });

                const description =
                    sources.length === 1
                        ? `Copied to /${destination}`
                        : `Copied ${sources.length} ${pluralizeItems(sources.length)} into /${destination}`;

                toast.success('Resource copied', { description });

                return { kind: 'ok' };
            } catch (error) {
                if (!force && getApiErrorStatusCode(error) === 409) {
                    return { kind: 'conflict' };
                }

                const description = getApiErrorMessage(error, 'Failed to copy resource');

                toast.error('Copy failed', { description });

                return { kind: 'error' };
            } finally {
                setIsCopying(false);
            }
        },
        [],
    );

    return {
        copy,
        isCopying,
    };
}
