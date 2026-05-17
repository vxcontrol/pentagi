import { useMemo } from 'react';
import { useSearchParams } from 'react-router-dom';

import { useDebouncedValue } from '@/hooks/use-debounced-value';
import { URL_PARAMS } from '@/lib/url-params';

interface UseTableQueryFilterReaderOptions {
    debounceMs?: number;
    /**
     * The query string param to read. Default `URL_PARAMS.QUERY` (`'q'`).
     */
    paramName?: string;
    /**
     * Storage key — accepted for API symmetry with `useTableState` and the
     * historical mutate variant, but ignored here. The reader is pure-URL
     * and never touches storage; the option exists only so detail pages can
     * pass the parent list's storage key without conditional logic.
     */
    storageKey?: string;
}

interface UseTableQueryFilterReaderResult {
    debouncedFilter: string;
    filter: string;
}

/**
 * Read-only subscription to the URL filter for pages that observe the value
 * but never mutate it (typically detail pages, where the toolbar walks the
 * filtered subset but the user types the filter on the list page).
 *
 * The hook never writes to the URL or storage: a detail page opened via a
 * shared `/flows/:id` link will not silently inject the previous tab's
 * `?q=` into the URL. Pages that *do* need to mutate the filter live on
 * `useTableState` instead — it owns the URL ↔ storage roundtrip plus
 * atomic multi-field updates that the previous split design couldn't do
 * race-free.
 */
export function useTableQueryFilterReader({
    debounceMs = 200,
    paramName = URL_PARAMS.QUERY,
}: UseTableQueryFilterReaderOptions = {}): UseTableQueryFilterReaderResult {
    const [searchParams] = useSearchParams();
    const filter = searchParams.get(paramName) ?? '';
    const debouncedFilter = useDebouncedValue(filter, debounceMs);

    return useMemo(() => ({ debouncedFilter, filter }), [debouncedFilter, filter]);
}
