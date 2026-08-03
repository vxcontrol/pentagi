import { useTableState } from '@/hooks/use-table-state';

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
 * Backed by `useTableState` so the query lives in the URL (`?q=`) — reloads
 * keep the filter active and shareable links work, but a fresh navigation
 * back to `/resources` starts clean. The debounce delay is preserved
 * (`SEARCH_DEBOUNCE_MS`) so the existing client-side tree filter still gets
 * the throttled value it expects.
 */
export function useResourcesSearch(): UseResourcesSearchResult {
    const { debouncedFilter, filter, resetFilter, setFilter } = useTableState({
        clearPageOnFilterChange: false,
        debounceMs: SEARCH_DEBOUNCE_MS,
    });

    return {
        debouncedQuery: debouncedFilter,
        rawQuery: filter,
        resetSearch: resetFilter,
        setQuery: setFilter,
    };
}
