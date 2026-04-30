import { useCallback, useState } from 'react';
import { toast } from 'sonner';
import { z } from 'zod';

import { api, getApiErrorMessage } from '@/lib/axios';

import { RESOURCES_MKDIR_API_PATH } from './resources-constants';

export const resourcesMkdirFormSchema = z.object({
    path: z
        .string()
        .trim()
        .min(1, { message: 'Path cannot be empty' })
        .refine((value) => !value.startsWith('/'), { message: 'Path must be relative (no leading "/")' })
        .refine((value) => !value.split('/').includes('..'), { message: 'Path must not contain ".."' }),
});

export type ResourcesMkdirFormValues = z.infer<typeof resourcesMkdirFormSchema>;

interface MkdirRequestBody {
    path: string;
}

interface UseResourcesMkdirResult {
    isCreating: boolean;
    mkdir: (values: ResourcesMkdirFormValues) => Promise<boolean>;
}

/** Wraps `POST /resources/mkdir` (idempotent — returns existing dir on hit). */
export const useResourcesMkdir = (): UseResourcesMkdirResult => {
    const [isCreating, setIsCreating] = useState(false);

    const mkdir = useCallback(async ({ path }: ResourcesMkdirFormValues): Promise<boolean> => {
        setIsCreating(true);

        try {
            await api.post<void, MkdirRequestBody>(RESOURCES_MKDIR_API_PATH, { path: path.trim() });

            toast.success('Directory created', { description: `Created at /${path.trim()}` });

            return true;
        } catch (error) {
            const description = getApiErrorMessage(error, 'Failed to create directory');

            toast.error('Create directory failed', { description });

            return false;
        } finally {
            setIsCreating(false);
        }
    }, []);

    return {
        isCreating,
        mkdir,
    };
};
