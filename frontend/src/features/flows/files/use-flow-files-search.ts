import { zodResolver } from '@hookform/resolvers/zod';
import { useCallback } from 'react';
import { useForm, type UseFormReturn } from 'react-hook-form';
import { z } from 'zod';

import { useDebouncedValue } from '@/hooks/use-debounced-value';

import { SEARCH_DEBOUNCE_MS } from './flow-files-constants';

const flowFilesSearchFormSchema = z.object({
    search: z.string(),
});

export type FlowFilesSearchFormValues = z.infer<typeof flowFilesSearchFormSchema>;

interface UseFlowFilesSearchResult {
    debouncedQuery: string;
    form: UseFormReturn<FlowFilesSearchFormValues>;
    rawQuery: string;
    resetSearch: () => void;
}

/**
 * Owns the search form and exposes a debounced version of the typed query.
 *
 * The hook intentionally has no `resetKey` parameter: the consumer is expected to
 * remount the subtree (via `key={flowId}` on the surrounding component) when the
 * flow changes, which clears form state without an imperative effect.
 */
export const useFlowFilesSearch = (): UseFlowFilesSearchResult => {
    const form = useForm<FlowFilesSearchFormValues>({
        defaultValues: { search: '' },
        resolver: zodResolver(flowFilesSearchFormSchema),
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
