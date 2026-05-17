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
 * Backed by `useTableState` so the query lives in the URL (`?q=`) with a
 * localStorage fallback — the FileManager survives reloads with the same
 * filter active and shareable links keep working. The debounce delay is
 * preserved (`SEARCH_DEBOUNCE_MS`) so the existing client-side tree filter
 * still gets the throttled value it expects.
 *
 * `clearPageOnFilterChange: false` because Resources has no `?page=` to
 * reset — leaving the default would also work (deleting a non-existent
 * param is a no-op), but the explicit setting documents intent.
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
