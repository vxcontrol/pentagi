import { useTableQueryFilter } from '@/hooks/use-table-query-filter';

import { SEARCH_DEBOUNCE_MS } from './resources-constants';

interface UseResourcesSearchResult {
    debouncedQuery: string;
    rawQuery: string;
    resetSearch: () => void;
    setQuery: (value: string) => void;
}

/**
 * Search state for the Resources file manager.
 *
 * Backed by `useTableQueryFilter` so the search query lives in the URL
 * (`?q=`) with a localStorage fallback — the FileManager survives reloads
 * with the same filter active and shareable links keep working. The debounce
 * delay is preserved (`SEARCH_DEBOUNCE_MS`) so the existing client-side
 * tree filter still gets the throttled value it expects. Storage key is
 * defaulted to the current pathname inside `useTableQueryFilter`.
 *
 * `clearPageParamOnChange: false` because the Resources page has no `?page=`
 * — there's no pagination to reset on every keystroke.
 */
export const useResourcesSearch = (): UseResourcesSearchResult => {
    const { debouncedFilter, filter, resetFilter, setFilter } = useTableQueryFilter({
        clearPageParamOnChange: false,
        debounceMs: SEARCH_DEBOUNCE_MS,
    });

    return {
        debouncedQuery: debouncedFilter,
        rawQuery: filter,
        resetSearch: resetFilter,
        setQuery: setFilter,
    };
};
