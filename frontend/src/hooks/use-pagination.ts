import { useCallback, useEffect, useMemo, useRef } from 'react';
import { useSearchParams } from 'react-router-dom';

import { URL_PARAMS } from '@/lib/url-params';

interface UsePaginationOptions {
    /**
     * Query string parameter that carries the 1-based page number. Default
     * `URL_PARAMS.PAGE` (`'page'`). Override when a route needs a custom
     * key (e.g. a page that already uses `?page=` for something else).
     */
    paramName?: string;
}

interface UsePaginationResult {
    /**
     * 0-based page index, suitable for handing straight to TanStack Table's
     * `state.pagination.pageIndex`. The URL stores it 1-based for humans;
     * the hook handles the off-by-one in both directions.
     */
    pageIndex: number;
    /** Set the page index (0-based). Pass `0` to drop the URL param entirely. */
    setPage: (pageIndex: number) => void;
}

/**
 * URL-backed pagination state. The 1-based page number lives in `?page=`, so
 * users can bookmark/share specific pages and the back/forward stack reflects
 * page changes.
 *
 * Conventions:
 * - The URL stores 1-based numbers (`?page=2` = second page) because that's
 *   what users see in pagers ("Page 2 of 10"). The hook converts to/from
 *   0-based at the boundary.
 * - `?page=1` is canonicalized away — the first page is the default URL.
 *   That avoids two URLs (`/flows` vs `/flows?page=1`) representing the
 *   same view, which would split the history stack.
 *
 * Pages that also use `useTableQueryFilter` get the `?page=` reset for free
 * when the filter narrows the result set — see that hook's
 * `clearPageParamOnChange` option.
 */
export const usePagination = ({ paramName = URL_PARAMS.PAGE }: UsePaginationOptions = {}): UsePaginationResult => {
    const [searchParams, setSearchParams] = useSearchParams();

    const pageIndex = useMemo(() => {
        const page = searchParams.get(paramName);

        if (!page) {
            return 0;
        }

        const parsed = Number.parseInt(page, 10);

        return Number.isFinite(parsed) ? Math.max(0, parsed - 1) : 0;
    }, [paramName, searchParams]);

    // Canonicalize `?<paramName>=1` to a clean URL whenever it shows up.
    // Two URLs (`/flows` vs `/flows?page=1`) would otherwise denote the same
    // view — splitting the history stack and giving link-sharers an ugly URL
    // for "first page". The effect uses `replace: true` so the rewrite never
    // adds a history entry, and is idempotent: once the param is removed it
    // re-runs and immediately exits the `if`, no loop.
    //
    // `useEffect` (not `useLayoutEffect`) is intentional: the URL bar updates
    // outside React's paint pipeline, so deferring the rewrite past commit
    // costs nothing visible to the user, while letting the canonicalization
    // settle through normal effect scheduling — friendlier to test harnesses
    // and to react-router's own transition handling.
    useEffect(() => {
        if (searchParams.get(paramName) !== '1') {
            return;
        }

        setSearchParams(
            (previous) => {
                const next = new URLSearchParams(previous);
                next.delete(paramName);

                return next;
            },
            { replace: true },
        );
    }, [paramName, searchParams, setSearchParams]);

    // See `use-table-query-filter.ts` for the full rationale — render-time
    // mutation is safe because the ref is only read from event callbacks
    // (`setPage`), never during render. We deliberately skip `useLatestRef`
    // because its passive-`useEffect` sync lags by one commit, which is
    // exactly the gap that lets a batched filter+page race lose state.
    const searchParamsReference = useRef(searchParams);
    // eslint-disable-next-line react-hooks/refs
    searchParamsReference.current = searchParams;

    const setPage = useCallback(
        (newPageIndex: number) => {
            // Avoid the functional `(previous) => ...` form: react-router v6
            // feeds the same pre-batch snapshot to every functional updater
            // queued in a single tick, so a `setFilter` + `setPage` pair
            // (e.g. debounced filter commit landing alongside a paging click)
            // would collapse — the second write erases the first. Prefer
            // `window.location.search` under `BrowserRouter`; fall back to
            // the latest committed `searchParams` (via ref) under
            // `MemoryRouter`, which doesn't sync `window.location`.
            const fromWindow = typeof window !== 'undefined' ? window.location.search : '';
            const next = fromWindow
                ? new URLSearchParams(fromWindow)
                : new URLSearchParams(searchParamsReference.current);

            if (newPageIndex <= 0) {
                next.delete(paramName);
            } else {
                next.set(paramName, String(newPageIndex + 1));
            }

            setSearchParams(next);
        },
        [paramName, setSearchParams],
    );

    return { pageIndex, setPage };
};
