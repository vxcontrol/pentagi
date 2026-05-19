import { useCallback, useEffect, useMemo, useRef } from 'react';
import { useSearchParams } from 'react-router-dom';

import { useDebouncedValue } from '@/hooks/use-debounced-value';
import { useEffectAfterMount } from '@/hooks/use-effect-after-mount';
import { usePageStorageKeys } from '@/hooks/use-page-storage-keys';
import { loadTableState, updateTableState } from '@/lib/table-state';
import { URL_PARAMS } from '@/lib/url-params';

/**
 * Atomic partial update for the table's URL state. All fields are optional;
 * `null` clears the corresponding param. `replace` controls whether the
 * navigation creates a new history entry. The whole patch is applied in
 * a single `setSearchParams` call so multi-field updates never race.
 */
interface TableStateUpdate {
    /** New filter value, `null`/`''` to clear. */
    filter?: null | string;
    /** New 0-based page index. Values `<= 0` clear the param. */
    pageIndex?: number;
    /** Pass `true` to navigate without adding a history entry. */
    replace?: boolean;
}

interface UseTableStateOptions {
    /**
     * Whether `setFilter` should also drop `?page=` from the URL. Pages that
     * pair the URL filter with URL pagination need this so the user doesn't
     * end up "on page 5 of nothing" after narrowing the result set. Pages
     * that have no `?page=` to begin with can leave the default — deleting a
     * non-existent param is a no-op.
     */
    clearPageOnFilterChange?: boolean;
    debounceMs?: number;
    /** Query string param for the filter value. Default `URL_PARAMS.QUERY` (`'q'`). */
    filterParamName?: string;
    /** Query string param for the 1-based page number. Default `URL_PARAMS.PAGE` (`'page'`). */
    pageParamName?: string;
    /**
     * Suppresses the one-time storage → URL restore on mount. Use on pages
     * that route the `?q=` filter alongside a sibling URL parameter with
     * its own filter semantics — e.g. `/knowledges` combines `?q=` (client
     * text filter) with `?qs=` (server semantic search). Restoring a
     * previously-typed `?q=` on top of a freshly-opened `?qs=` link would
     * silently narrow the search result the user navigated to. The caller
     * decides per-mount whether restore should fire; the value is captured
     * via a ref at mount time so a later toggle doesn't replay a restore
     * the user has since worked past.
     */
    skipRestore?: boolean;
    /**
     * Stable storage key for persisting the filter value (a fresh tab without
     * `?<filterParamName>=` resumes from here). Defaults to
     * `usePageStorageKeys().table` — i.e. `table_4_<current pathname>`.
     */
    storageKey?: string;
}

interface UseTableStateResult {
    debouncedFilter: string;
    filter: string;
    pageIndex: number;
    resetFilter: () => void;
    setFilter: (value: string) => void;
    setPage: (pageIndex: number) => void;
    /**
     * Atomic multi-field update. Prefer this when changing both `filter` and
     * `pageIndex` from the same event (or when adding more fields in the
     * future): all changes land in a single `setSearchParams` call, so the
     * race between two consecutive top-level updaters can never happen.
     */
    update: (patch: TableStateUpdate) => void;
}

/**
 * Unified URL + storage state for tables.
 *
 * Replaces the split `useTableQueryFilter` / `usePagination` pair. The split
 * design suffered from a batching race: react-router v6 feeds every
 * functional `setSearchParams(updater)` queued in a single tick the same
 * pre-batch snapshot, so a `setFilter` + `setPage` pair (e.g. a debounced
 * filter commit landing alongside a paging click) would collapse — the
 * second write erased the first and `q` was lost. Funnelling every URL
 * write through a single `update` here removes the race by construction:
 * there is never more than one in-flight `setSearchParams` per logical
 * intent. The few cases where two intents still fire in the same tick
 * (e.g. an external effect mutating the URL while we batch our own write)
 * read the freshest URL via `window.location.search`, with a ref-stashed
 * react-router snapshot as the fallback for `MemoryRouter`-based tests.
 *
 * Read-only siblings (detail pages reading the list's filter without
 * mutating it) should keep using `useTableQueryFilterReader` — no shared
 * URL writes, no race.
 */
export function useTableState(options: UseTableStateOptions = {}): UseTableStateResult {
    const {
        clearPageOnFilterChange = true,
        debounceMs = 200,
        filterParamName = URL_PARAMS.QUERY,
        pageParamName = URL_PARAMS.PAGE,
        skipRestore = false,
        storageKey: explicitStorageKey,
    } = options;

    const [searchParams, setSearchParams] = useSearchParams();
    const { table: defaultStorageKey } = usePageStorageKeys();
    const storageKey = explicitStorageKey ?? defaultStorageKey;

    const filter = searchParams.get(filterParamName) ?? '';
    const debouncedFilter = useDebouncedValue(filter, debounceMs);
    const pageIndex = useMemo(() => {
        const raw = searchParams.get(pageParamName);

        if (!raw) {
            return 0;
        }

        const parsed = Number.parseInt(raw, 10);

        return Number.isFinite(parsed) ? Math.max(0, parsed - 1) : 0;
    }, [pageParamName, searchParams]);

    // Sync a ref to the latest committed `searchParams` on every render.
    // Used as the `MemoryRouter` fallback below — that environment doesn't
    // sync its in-memory history to `window.location`. Mutating the ref in
    // render is safe because we only read it from event callbacks
    // (`update`), never during render; `useEffect`-based sync (a-la
    // `useLatestRef`) lags one commit, which is exactly the gap that
    // re-introduces the race we're trying to eliminate.
    const searchParamsReference = useRef(searchParams);
    // eslint-disable-next-line react-hooks/refs
    searchParamsReference.current = searchParams;

    /**
     * Read the freshest possible `URLSearchParams` for the next write.
     *
     * Under `BrowserRouter` (production), `window.location.search` reflects
     * the URL bar one frame ahead of react-router's internal snapshot when
     * batched updates are in flight — that's the seam we exploit to merge
     * multi-source URL mutations safely. Under `MemoryRouter` (tests)
     * `window.location` doesn't track the in-memory history, so we fall
     * back to the rendered react-router snapshot.
     */
    const readLatestParams = useCallback((): URLSearchParams => {
        const fromWindow = typeof window !== 'undefined' ? window.location.search : '';

        return fromWindow ? new URLSearchParams(fromWindow) : new URLSearchParams(searchParamsReference.current);
    }, []);

    // Canonicalize `?<pageParamName>=1` away. The first page is the default
    // URL, so two URLs (`/flows` vs `/flows?page=1`) would otherwise denote
    // the same view and split the history stack. Idempotent: once the param
    // is removed the effect re-runs and immediately exits the early return.
    useEffect(() => {
        if (searchParams.get(pageParamName) !== '1') {
            return;
        }

        const next = readLatestParams();

        next.delete(pageParamName);
        setSearchParams(next, { replace: true });
    }, [pageParamName, readLatestParams, searchParams, setSearchParams]);

    // Replay the persisted filter into the URL when (a) the URL doesn't
    // already carry one, and (b) the storage has a non-empty value. Run
    // exactly once per storageKey rotation — `restoredForKeyReference`
    // guards against repeating the replay on every render.
    const restoredForKeyReference = useRef<null | string>(null);
    // Snapshot the `skipRestore` decision at mount via a ref so a later
    // toggle (e.g. the user clearing `?qs=`) doesn't fire the restore
    // mid-session — by then they've already engaged with the page and a
    // surprise URL mutation would be jarring.
    const skipRestoreRef = useRef(skipRestore);

    useEffect(() => {
        if (restoredForKeyReference.current === storageKey) {
            return;
        }

        restoredForKeyReference.current = storageKey;

        if (skipRestoreRef.current) {
            return;
        }

        if (filter.length > 0) {
            // URL already has a value — that beats storage. Mirror it back
            // into storage so a fresh tab without `?q=` resumes from this
            // intent (shared filtered links).
            updateTableState(storageKey, { filter });

            return;
        }

        const stored = loadTableState(storageKey).filter;

        if (!stored || stored.length === 0) {
            return;
        }

        const next = readLatestParams();

        next.set(filterParamName, stored);
        // Replace, not push: restoring prior state shouldn't add a history
        // entry the user didn't ask for.
        setSearchParams(next, { replace: true });
    }, [filter, filterParamName, readLatestParams, setSearchParams, storageKey]);

    // Persist the URL filter into storage on every commit. Skipping the
    // first render is intentional: a fresh-mount empty `filter` would wipe
    // a freshly-restored storage entry before the restore effect above has
    // had a chance to replay it into the URL.
    useEffectAfterMount(() => {
        updateTableState(storageKey, { filter: filter.length > 0 ? filter : undefined });
    }, [filter, storageKey]);

    // Coalesce every `update(...)` call fired in the same microtask into one
    // `setSearchParams`. The first call schedules the flush; subsequent calls
    // merge their patch into the pending buffer instead of issuing their own
    // navigation. This is what makes race conditions impossible *by
    // construction*: it doesn't matter whether two updates come from the
    // same event handler, from two effects, or from a debounced commit
    // landing alongside a synchronous click — only one navigation happens,
    // with both fields applied.
    const pendingPatchReference = useRef<null | {
        filter: null | string | undefined;
        filterPresent: boolean;
        pageIndex: number | undefined;
        pageIndexPresent: boolean;
        // Replace resolution: any push-intent (`replace: false`) wins, so
        // intentional history entries (paging clicks) survive coalescence
        // with replace-only updates (filter typing).
        replace: boolean;
    }>(null);

    const update = useCallback(
        (patch: TableStateUpdate) => {
            const filterPresent = 'filter' in patch;
            const pageIndexPresent = 'pageIndex' in patch;
            const requestedReplace = patch.replace ?? false;

            if (pendingPatchReference.current === null) {
                pendingPatchReference.current = {
                    filter: patch.filter,
                    filterPresent,
                    pageIndex: patch.pageIndex,
                    pageIndexPresent,
                    replace: requestedReplace,
                };

                queueMicrotask(() => {
                    const merged = pendingPatchReference.current;

                    pendingPatchReference.current = null;

                    if (merged === null) {
                        return;
                    }

                    const next = readLatestParams();

                    if (merged.filterPresent) {
                        if (!merged.filter) {
                            next.delete(filterParamName);
                        } else {
                            next.set(filterParamName, merged.filter);
                        }
                    }

                    if (merged.pageIndexPresent) {
                        const newIndex = merged.pageIndex ?? 0;

                        if (newIndex <= 0) {
                            next.delete(pageParamName);
                        } else {
                            next.set(pageParamName, String(newIndex + 1));
                        }
                    }

                    setSearchParams(next, { replace: merged.replace });
                });

                return;
            }

            // Merge into the in-flight patch — the queued microtask will see
            // the fused result.
            if (filterPresent) {
                pendingPatchReference.current.filter = patch.filter;
                pendingPatchReference.current.filterPresent = true;
            }

            if (pageIndexPresent) {
                pendingPatchReference.current.pageIndex = patch.pageIndex;
                pendingPatchReference.current.pageIndexPresent = true;
            }

            if (!requestedReplace) {
                pendingPatchReference.current.replace = false;
            }
        },
        [filterParamName, pageParamName, readLatestParams, setSearchParams],
    );

    const setFilter = useCallback(
        (value: string) => {
            // Typing keystrokes commit through here as well — we don't want
            // intermediate filter values cluttering the history stack, so
            // every filter change is a `replace`. The implicit page reset
            // is bundled into the same atomic update so the resulting URL
            // is consistent in one transition rather than two.
            update({
                filter: value.length === 0 ? null : value,
                pageIndex: clearPageOnFilterChange ? 0 : undefined,
                replace: true,
            });
        },
        [clearPageOnFilterChange, update],
    );

    const setPage = useCallback(
        (newPageIndex: number) => {
            // Paging is an intentional user action — push, not replace, so
            // back-button steps through the visited pages.
            update({ pageIndex: newPageIndex });
        },
        [update],
    );

    const resetFilter = useCallback(() => setFilter(''), [setFilter]);

    return useMemo(
        () => ({
            debouncedFilter,
            filter,
            pageIndex,
            resetFilter,
            setFilter,
            setPage,
            update,
        }),
        [debouncedFilter, filter, pageIndex, resetFilter, setFilter, setPage, update],
    );
}
