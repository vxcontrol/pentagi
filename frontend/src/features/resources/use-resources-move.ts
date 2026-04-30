import { useCallback, useState } from 'react';
import { toast } from 'sonner';
import { z } from 'zod';

import { api, type ApiHttpError, getApiErrorMessage } from '@/lib/axios';

import { RESOURCES_MOVE_API_PATH } from './resources-constants';

export const resourcesMoveFormSchema = z.object({
    destination: z
        .string()
        .trim()
        .min(1, { message: 'Destination cannot be empty' })
        .refine((value) => !value.startsWith('/'), { message: 'Destination must be a relative path' })
        .refine((value) => !value.split('/').includes('..'), { message: 'Destination must not contain ".."' }),
    shouldOverwrite: z.boolean(),
});

export interface ResourcesMoveConflict {
    destination: string;
    /** Display name extracted from `destination` for the confirm dialog. */
    destinationName: string;
    sourcePath: string;
}

export type ResourcesMoveFormValues = z.infer<typeof resourcesMoveFormSchema>;

interface MoveRequestBody {
    destination: string;
    force: boolean;
    source: string;
}

interface UseResourcesMoveResult {
    /** Drop the pending conflicts without retrying. */
    cancelConflicts: () => void;
    isMoving: boolean;
    move: (sourcePath: string, values: ResourcesMoveFormValues) => Promise<boolean>;
    /**
     * 409 conflicts collected across one or more parallel `move()` calls. The consumer
     * shows a single "Replace?" dialog summarising the count; clicking Replace retries
     * every entry with `force = true`.
     */
    pendingConflicts: ResourcesMoveConflict[];
    /** Retry every pending conflict with `force = true`. Promise resolves when all settle. */
    resolveConflicts: () => Promise<void>;
}

const getStatusCode = (error: unknown): number | undefined => {
    const httpError = error as ApiHttpError | null | undefined;

    return httpError?.statusCode ?? httpError?.response?.status;
};

const extractName = (path: string): string => path.split('/').pop() ?? path;

/** Wraps `PUT /resources/move` for rename / move operations. */
export const useResourcesMove = (): UseResourcesMoveResult => {
    const [isMoving, setIsMoving] = useState(false);
    const [pendingConflicts, setPendingConflicts] = useState<ResourcesMoveConflict[]>([]);

    const performMove = useCallback(
        async (sourcePath: string, destination: string, force: boolean): Promise<boolean> => {
            setIsMoving(true);

            try {
                await api.put<void, MoveRequestBody>(RESOURCES_MOVE_API_PATH, {
                    destination,
                    force,
                    source: sourcePath,
                });

                toast.success('Resource moved', { description: `Moved to /${destination}` });

                return true;
            } catch (error) {
                // 409 = destination already exists. Push to the conflict array so the
                // consumer can render a single aggregated "Replace?" dialog covering
                // every parallel `move()` call. `force = true` skips this branch (the
                // overwrite was already pre-confirmed), so retries land in the toast
                // path on any unexpected re-conflict.
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

                const description = getApiErrorMessage(error, 'Failed to move resource');

                toast.error('Move failed', { description });

                return false;
            } finally {
                setIsMoving(false);
            }
        },
        [],
    );

    const move = useCallback(
        (sourcePath: string, { destination, shouldOverwrite }: ResourcesMoveFormValues): Promise<boolean> =>
            performMove(sourcePath, destination.trim(), shouldOverwrite),
        [performMove],
    );

    const resolveConflicts = useCallback(async (): Promise<void> => {
        if (pendingConflicts.length === 0) {
            return;
        }

        const conflicts = pendingConflicts;

        setPendingConflicts([]);

        await Promise.allSettled(
            conflicts.map((conflict) => performMove(conflict.sourcePath, conflict.destination, true)),
        );
    }, [pendingConflicts, performMove]);

    const cancelConflicts = useCallback(() => {
        setPendingConflicts([]);
    }, []);

    return {
        cancelConflicts,
        isMoving,
        move,
        pendingConflicts,
        resolveConflicts,
    };
};
