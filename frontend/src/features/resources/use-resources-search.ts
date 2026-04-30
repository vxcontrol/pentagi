import { zodResolver } from '@hookform/resolvers/zod';
import { useCallback } from 'react';
import { useForm, type UseFormReturn } from 'react-hook-form';
import { z } from 'zod';

import { useDebouncedValue } from '@/hooks/use-debounced-value';

import { SEARCH_DEBOUNCE_MS } from './resources-constants';

const resourcesSearchFormSchema = z.object({
    search: z.string(),
});

export type ResourcesSearchFormValues = z.infer<typeof resourcesSearchFormSchema>;

interface UseResourcesSearchResult {
    debouncedQuery: string;
    form: UseFormReturn<ResourcesSearchFormValues>;
    rawQuery: string;
    resetSearch: () => void;
}

/** Owns the search form and exposes a debounced version of the typed query. */
export const useResourcesSearch = (): UseResourcesSearchResult => {
    const form = useForm<ResourcesSearchFormValues>({
        defaultValues: { search: '' },
        resolver: zodResolver(resourcesSearchFormSchema),
    });

    const rawQuery = form.watch('search');
    const debouncedQuery = useDebouncedValue(rawQuery, SEARCH_DEBOUNCE_MS);

    const resetSearch = useCallback(() => {
        form.reset({ search: '' });
    }, [form]);

    return {
        debouncedQuery,
        form,
        rawQuery,
        resetSearch,
    };
};
