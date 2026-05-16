import { useCallback, useEffect, useMemo, useRef } from 'react';
import { useSearchParams } from 'react-router-dom';

import { useDebouncedValue } from '@/hooks/use-debounced-value';
import { useEffectAfterMount } from '@/hooks/use-effect-after-mount';
import { usePageStorageKeys } from '@/hooks/use-page-storage-keys';
import { loadTableState, updateTableState } from '@/lib/table-state';
import { URL_PARAMS } from '@/lib/url-params';

interface UseTableQueryFilterOptions {
    /**
     * Whether to also reset the `?page=` URL param when the filter changes.
     * Pages that pair URL filter with URL pagination need this so the user
     * doesn't end up "on page 5 of nothing" after narrowing the result set.
     * Defaults to `true`.
     */
    clearPageParamOnChange?: boolean;
    debounceMs?: number;
    /**
     * The query string param that carries the filter. Default `URL_PARAMS.QUERY`
     * (`'q'`) matches the conventional "search query" name and is consistent
     * across pages.
     */
    paramName?: string;
    /**
     * Stable storage key for persisting the most recent value (a fresh tab
     * without `?q=` resumes from here). When omitted, defaults to
     * `usePageStorageKeys().table` — i.e. `table_4_<current pathname>`,
     * which is the right key for any top-level list route. Override only
     * for nested routes or when you need to share a storage slot across
     * paths (the same way a detail page shares its parent list's slot
     * through `useTableQueryFilterReader`).
     */
    storageKey?: string;
}

interface UseTableQueryFilterReaderResult {
    debouncedFilter: string;
    filter: string;
}

interface UseTableQueryFilterResult extends UseTableQueryFilterReaderResult {
    resetFilter: () => void;
    setFilter: (value: string) => void;
}

/**
 * Pure URL reader: surface the live + debounced filter without touching
 * storage or scheduling any side-effects.
 *
 * Re-evaluates whenever `?<paramName>=` changes — so a detail page that
 * subscribes via `useTableQueryFilterReader` (or any caller that wants
 * read-only access) stays in sync with whatever the list page typed.
 */
const useFilterFromUrl = ({
    debounceMs = 200,
    paramName = URL_PARAMS.QUERY,
}: Pick<UseTableQueryFilterOptions, 'debounceMs' | 'paramName'>): UseTableQueryFilterReaderResult => {
    const [searchParams] = useSearchParams();
    const filter = searchParams.get(paramName) ?? '';
    const debouncedFilter = useDebouncedValue(filter, debounceMs);

    return useMemo(() => ({ debouncedFilter, filter }), [debouncedFilter, filter]);
};

/**
 * Read-only subscription to the URL filter for pages that observe the value
 * but never mutate it (typically detail pages, where the toolbar walks the
 * filtered subset but the user types the filter on the list page).
 *
 * Unlike `useTableQueryFilter`, this hook never writes to the URL or storage:
 * a detail page opened by a shared `/flows/:id` link will not silently inject
 * the previous tab's `?q=` into the URL. The `storageKey` option is accepted
 * for API symmetry but ignored here — there is no storage interaction at all.
 */
export const useTableQueryFilterReader = (
    options: UseTableQueryFilterOptions = {},
): UseTableQueryFilterReaderResult => {
    return useFilterFromUrl(options);
};

/**
 * URL ↔ localStorage filter for list pages.
 *
 * - Source of truth is `?<paramName>=` in the URL — that lets users share /
 *   bookmark a filtered view, and any explicit path-level navigation (a
 *   route change, opening a detail page, etc.) is captured by react-router's
 *   own history entry alongside whatever filter was active at that moment.
 * - On mount (and whenever the effective `storageKey` rotates — e.g. the
 *   route changes while the hook stays mounted in a shared layout), if the
 *   URL has no `?<paramName>=` but storage does, we replay the stored value
 *   into the URL (using `replace: true`, so the restore doesn't pollute the
 *   back stack).
 * - Every typed change goes into the URL with `replace: true`. That means
 *   intermediate filter values do **not** create new history entries —
 *   browser back skips past the filter typing session in one step rather
 *   than walking through `'f' → 'fo' → 'foo'`.
 * - Storage is the single sink. The unified `table` slot
 *   (`updateTableState`) carries `filter` alongside the rest of the table
 *   state — sorting, column visibility, page size — under one key per page.
 *   Each commit writes synchronously through `useEffectAfterMount`: a single
 *   source of truth removes the previous double-write pattern (sync clear +
 *   debounced effect) that risked races on refresh.
 */
export const useTableQueryFilter = (options: UseTableQueryFilterOptions = {}): UseTableQueryFilterResult => {
    const { clearPageParamOnChange = true, paramName = URL_PARAMS.QUERY, storageKey: explicitStorageKey } = options;
    const [searchParams, setSearchParams] = useSearchParams();
    const { table: defaultStorageKey } = usePageStorageKeys();
    const storageKey = explicitStorageKey ?? defaultStorageKey;

    const { debouncedFilter, filter } = useFilterFromUrl(options);

    // Track which storageKey we've already restored from. When the key
    // rotates (route change inside a persistent layout, or an explicit
    // override prop change) we reset the guard so the new slot gets one
    // restore attempt — this is the fix for the "lazy useState captured the
    // mount-time key forever" pitfall the previous design had.
    const restoredForKeyReference = useRef<null | string>(null);

    // `filter` participates in the dep array directly, not through a ref.
    // Two reasons:
    //   1. The `restoredForKeyReference` guard makes the effect a no-op on
    //      keystrokes — the early return fires before any storage write,
    //      so the per-keystroke re-runs cost is one Map lookup.
    //   2. `useLatestRef`-style refs have a one-commit lag (the sync
    //      `useEffect` runs *after* the consumer). Reading the URL filter
    //      that way at the moment `storageKey` rotates would see the
    //      *previous* route's value. Plain dep gives the freshest value.
    useEffect(() => {
        if (restoredForKeyReference.current === storageKey) {
            return;
        }

        restoredForKeyReference.current = storageKey;

        if (filter.length > 0) {
            // The URL already has a value — that beats storage. Sync it into
            // storage so a fresh tab without `?<paramName>=` resumes from
            // this intent (e.g. user landed via a shared filtered link).
            updateTableState(storageKey, { filter });

            return;
        }

        const stored = loadTableState(storageKey).filter;

        if (!stored || stored.length === 0) {
            return;
        }

        setSearchParams(
            (previous) => {
                const next = new URLSearchParams(previous);
                next.set(paramName, stored);

                // Intentionally do NOT delete `?page=` here. The user's
                // explicit `?page=N` (from a shared link or refresh) is their
                // own request; the replay is restoring prior filter state,
                // not a fresh filter change. If the replayed filter happens
                // to make the requested page out-of-range, the page-index
                // clamp in `DataTable` reconciles it on the next render —
                // see the `safePageIndex` derivation. `clearPageParamOnChange`
                // still applies to `setFilter` below, where the user is
                // actively changing the filter and "page 5 of nothing" is
                // the failure mode we want to avoid.
                return next;
            },
            { replace: true },
        );
    }, [filter, paramName, setSearchParams, storageKey]);

    // Persist the URL filter into storage on every commit. Skipping the
    // first render is intentional: a fresh-mount empty `filter` would
    // wipe a freshly-restored storage entry before the restore effect
    // above has had a chance to replay it into the URL.
    useEffectAfterMount(() => {
        updateTableState(storageKey, { filter: filter.length > 0 ? filter : undefined });
    }, [filter, storageKey]);

    // Sync a ref to the latest committed `searchParams` on every render.
    // Used as a fallback below when `window.location` isn't a faithful
    // mirror of the router state — that happens under `MemoryRouter`,
    // which is what our hook tests use. Updating the ref in render is safe
    // because we never read it during render; we only read it inside event
    // callbacks (`setFilter` below). `useEffect`-synced refs (see
    // `useLatestRef`) lag one commit, which is fatal here: two updates
    // fired in the same tick must both see the same authoritative latest
    // value, not the prior-render snapshot.
    const searchParamsReference = useRef(searchParams);
    // eslint-disable-next-line react-hooks/refs
    searchParamsReference.current = searchParams;

    const setFilter = useCallback(
        (value: string) => {
            // Avoid the functional `(previous) => ...` form of
            // `setSearchParams`. When two updates fire in the same React tick
            // (e.g. a debounced filter commit landing alongside a paging
            // button click), react-router v6 feeds **both** updaters the same
            // pre-batch snapshot — the second write erases the first, so
            // `?q=foo` plus `setPage(5)` collapses to `?page=6` and `q` is
            // lost. `window.location.search` is the freshest source under
            // `BrowserRouter` (production) and immune to that batching race;
            // under `MemoryRouter` (tests) the in-memory history is *not*
            // synced to `window.location`, so we fall back to the ref-stashed
            // latest react-router snapshot.
            const fromWindow = typeof window !== 'undefined' ? window.location.search : '';
            const next = fromWindow
                ? new URLSearchParams(fromWindow)
                : new URLSearchParams(searchParamsReference.current);

            if (value.length === 0) {
                next.delete(paramName);
            } else {
                next.set(paramName, value);
            }

            if (clearPageParamOnChange) {
                next.delete(URL_PARAMS.PAGE);
            }

            setSearchParams(next, { replace: true });
        },
        [clearPageParamOnChange, paramName, setSearchParams],
    );

    const resetFilter = useCallback(() => {
        setFilter('');
    }, [setFilter]);

    return useMemo(
        () => ({ debouncedFilter, filter, resetFilter, setFilter }),
        [debouncedFilter, filter, resetFilter, setFilter],
    );
};
