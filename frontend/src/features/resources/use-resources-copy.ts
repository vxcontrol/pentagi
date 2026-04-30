import { useCallback, useState } from 'react';
import { toast } from 'sonner';
import { z } from 'zod';

import { api, type ApiHttpError, getApiErrorMessage } from '@/lib/axios';

import { RESOURCES_COPY_API_PATH } from './resources-constants';

export const resourcesCopyFormSchema = z.object({
    destination: z
        .string()
        .trim()
        .min(1, { message: 'Destination cannot be empty' })
        .refine((value) => !value.startsWith('/'), { message: 'Destination must be a relative path' })
        .refine((value) => !value.split('/').includes('..'), { message: 'Destination must not contain ".."' }),
    shouldOverwrite: z.boolean(),
});

export interface ResourcesCopyConflict {
    destination: string;
    /** Display name extracted from `destination` for the confirm dialog. */
    destinationName: string;
    sourcePath: string;
}

export type ResourcesCopyFormValues = z.infer<typeof resourcesCopyFormSchema>;

interface CopyRequestBody {
    destination: string;
    force: boolean;
    source: string;
}

interface UseResourcesCopyResult {
    /** Drop the pending conflicts without retrying. */
    cancelConflicts: () => void;
    copy: (sourcePath: string, values: ResourcesCopyFormValues) => Promise<boolean>;
    isCopying: boolean;
    /**
     * 409 conflicts collected across one or more parallel `copy()` calls. The consumer
     * shows a single "Replace?" dialog summarising the count; clicking Replace retries
     * every entry with `force = true`.
     */
    pendingConflicts: ResourcesCopyConflict[];
    /** Retry every pending conflict with `force = true`. Promise resolves when all settle. */
    resolveConflicts: () => Promise<void>;
}

const getStatusCode = (error: unknown): number | undefined => {
    const httpError = error as ApiHttpError | null | undefined;

    return httpError?.statusCode ?? httpError?.response?.status;
};

const extractName = (path: string): string => path.split('/').pop() ?? path;

/** Wraps `POST /resources/copy`. */
export const useResourcesCopy = (): UseResourcesCopyResult => {
    const [isCopying, setIsCopying] = useState(false);
    const [pendingConflicts, setPendingConflicts] = useState<ResourcesCopyConflict[]>([]);

    const performCopy = useCallback(
        async (sourcePath: string, destination: string, force: boolean): Promise<boolean> => {
            setIsCopying(true);

            try {
                await api.post<void, CopyRequestBody>(RESOURCES_COPY_API_PATH, {
                    destination,
                    force,
                    source: sourcePath,
                });

                toast.success('Resource copied', { description: `Copied to /${destination}` });

                return true;
            } catch (error) {
                if (getStatusCode(error) === 409 && !force) {
                    setPendingConflicts((prev) => [
                        ...prev,
                        {
                            destination,
                            destinationName: extractName(destination),
                            sourcePath,
                        },
                    ]);

                    return false;
                }

                const description = getApiErrorMessage(error, 'Failed to copy resource');

                toast.error('Copy failed', { description });

                return false;
            } finally {
                setIsCopying(false);
            }
        },
        [],
    );

    const copy = useCallback(
        (sourcePath: string, { destination, shouldOverwrite }: ResourcesCopyFormValues): Promise<boolean> =>
            performCopy(sourcePath, destination.trim(), shouldOverwrite),
        [performCopy],
    );

    const resolveConflicts = useCallback(async (): Promise<void> => {
        if (pendingConflicts.length === 0) {
            return;
        }

        const conflicts = pendingConflicts;

        setPendingConflicts([]);

        await Promise.allSettled(
            conflicts.map((conflict) => performCopy(conflict.sourcePath, conflict.destination, true)),
        );
    }, [pendingConflicts, performCopy]);

    const cancelConflicts = useCallback(() => {
        setPendingConflicts([]);
    }, []);

    return {
        cancelConflicts,
        copy,
        isCopying,
        pendingConflicts,
        resolveConflicts,
    };
};
