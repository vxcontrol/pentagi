import { useCallback, useMemo, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useDebounce } from 'use-debounce';

import { useLatestRef } from '@/hooks/use-latest-ref';
import { useTableQueryFilterReader } from '@/hooks/use-table-query-filter';
import { mergeHrefWithSearchParams } from '@/lib/url-params';

import { useNavigation } from './use-navigation';

/**
 * Default debounce for the in-sheet local search. Shorter than the URL filter's
 * 200 ms (`useTableQueryFilterReader`) — local search doesn't round-trip
 * through `history` / re-render the route subtree, so we can afford to react
 * faster while still coalescing burst typing.
 */
const DEFAULT_SEARCH_DEBOUNCE_MS = 150;

/**
 * Headless controller for a detail page that walks a filtered list.
 *
 * Owns:
 *   - the filtered/sorted subset (pure `computeNavigation`), narrowed by
 *     the URL filter AND an optional in-controller local search,
 *   - the resolved Prev / Next sibling ids and the pre-formatted position
 *     label (so leaf components don't recompute),
 *   - the sheet open state (controllable via `open` / `onOpenChange`),
 *   - the local search query state (controllable via `searchQuery` /
 *     `onSearchQueryChange`), with its own debounce independent of the URL,
 *   - navigation actions that thread the current `?<filter>=` into every
 *     prev / next / item-select destination.
 *
 * All callbacks are identity-stable across renders that don't change the
 * inputs the action depends on — `<DetailNavigationButtons>`,
 * `<DetailNavigationSheet>`, and any custom chrome rendered against the
 * controller can rely on referential equality for downstream memos.
 */
export interface DetailNavigationController<T extends { id: string }> {
    /** Reset the local search to an empty string. Convenience for "X" buttons. */
    clearSearchQuery: () => void;
    closeSheet: () => void;
    /** Active `currentId` coerced to a string, or `null` when absent. */
    currentId: null | string;
    /** Index of `currentItem` inside `filteredItems`. `-1` when not in subset. */
    currentIndex: number;
    /** Item matching `currentId` inside the filtered subset, or `null`. */
    currentItem: null | T;
    /** Debounced URL filter the controller is filtering against. */
    debouncedFilter: string;
    /**
     * Debounced version of {@link searchQuery} — this is what actually narrows
     * `filteredItems`. Combined with `debouncedFilter` via AND.
     */
    debouncedSearchQuery: string;
    /** Sorted+filtered subset that drives prev / next / sheet listing. */
    filteredItems: readonly T[];
    /** Stable id accessor (defaults to `item.id` when not supplied). */
    getId: (item: T) => string;

    getLabel: (item: T) => string;
    /**
     * Resolved haystack accessor — the caller-supplied `getSearchableText`,
     * or `getLabel` as a fallback. Exposed for advanced consumers; leaf
     * components don't read it because filtering already happened on the
     * way into `filteredItems`.
     */
    getSearchableText: (item: T) => null | string | undefined;
    goToNext: () => void;

    goToPrev: () => void;
    /** Navigate to the given item and close the sheet. */
    handleItemSelect: (item: T) => void;
    /** `true` iff `filteredItems.length > 0`. */
    hasEntries: boolean;
    /** `true` iff `debouncedSearchQuery` is non-empty (UI hints / clear button). */
    isSearchActive: boolean;
    isSheetOpen: boolean;

    /**
     * `true` iff the raw `items` array is empty (pre-filter). The convenience
     * `<DetailNavigationToolbar>` uses this to render `null` on a fresh detail
     * mount when the provider's list hasn't arrived yet.
     */
    itemsEmpty: boolean;
    /** ID of the next filtered sibling, or `null` at the end / off-subset. */
    nextId: null | string;
    openSheet: () => void;

    /** Pre-formatted `"3/10"` or `"–/0"` for the position trigger. */
    positionLabel: string;
    /** ID of the previous filtered sibling, or `null` at the start / off-subset. */
    prevId: null | string;
    /**
     * Current local search value (raw, **not** debounced) — bind directly to
     * an `<input value={...}>` for instant caret feedback. The debounced
     * mirror is what filters the list, see {@link debouncedSearchQuery}.
     */
    searchQuery: string;
    setSearchQuery: (value: string) => void;
    setSheetOpen: (open: boolean) => void;

    /** Same as `filteredItems.length`, named explicitly for clarity. */
    total: number;
}

interface UseDetailNavigationOptions<T extends { id: string }> {
    currentId: null | string | undefined;
    /** Initial value for the uncontrolled case. Defaults to `false`. */
    defaultOpen?: boolean;
    /** Initial local search value for the uncontrolled case. Defaults to `''`. */
    defaultSearchQuery?: string;
    getHref: (item: T) => string;
    getId?: (item: T) => string;
    getLabel: (item: T) => string;
    getSearchableText?: (item: T) => null | string | undefined;
    items: readonly T[];
    onOpenChange?: (open: boolean) => void;

    /**
     * Fires on every `setSearchQuery` call — useful when the consumer wants
     * to mirror the value into URL params, localStorage, or analytics. Like
     * `onOpenChange`, it fires in both controlled and uncontrolled mode.
     */
    onSearchQueryChange?: (query: string) => void;

    /**
     * Controlled-mode opt-in for the sheet. When `open` is `undefined` the
     * controller owns the state internally; when a value is provided the
     * caller owns it. `onOpenChange` always fires so a fully-controlled
     * consumer can observe every set.
     *
     * Mirrors the `useControllable` pattern from
     * `@/components/ui/autocomplete.tsx`.
     */
    open?: boolean;

    /**
     * Debounce (ms) applied between the raw `searchQuery` typed by the user
     * and the `debouncedSearchQuery` that actually filters `filteredItems`.
     * Defaults to {@link DEFAULT_SEARCH_DEBOUNCE_MS}. Pass `0` for instant
     * filtering on small lists.
     */
    searchDebounceMs?: number;

    /**
     * Controlled-mode opt-in for the local search query — same `undefined`
     * → uncontrolled / value → controlled convention as `open`. Use this
     * when the consumer wants to surface the search outside the sheet (e.g.
     * a page-level search box that also drives sibling navigation).
     */
    searchQuery?: string;
    sortFn?: (a: T, b: T) => number;
}

// Module-level so the reference is stable across renders. Typed against
// the wider `{ id: string }` so it is assignable to `(item: T) => string`
// for any `T extends { id: string }` via function-parameter contravariance
// — no cast at the call site.
const defaultGetId = (item: { id: string }): string => item.id;

/**
 * Build the headless `DetailNavigationController<T>` for a detail page.
 *
 * Bundles the three moving parts every detail page repeats:
 *   1. subscribing to the URL filter through `useTableQueryFilterReader`
 *      (read-only — the detail page never mutates the filter from here),
 *   2. running the pure `useNavigation` core against the filtered subset,
 *   3. wiring identity-stable `goToPrev` / `goToNext` / `handleItemSelect`
 *      with `?<filter>=` forwarded through `mergeHrefWithSearchParams`.
 *
 * The returned controller drives `<DetailNavigationToolbar>`,
 * `<DetailNavigationButtons>`, and `<DetailNavigationSheet>` (or any custom
 * chrome a consumer wants to render). All function fields are wrapped in
 * `useCallback` with `useLatestRef`-stabilized closures over `useSearchParams`
 * and the caller-supplied accessors — React Router v6 returns a fresh
 * `URLSearchParams` each render, and threading it through callback deps would
 * defeat the entire memoization goal.
 *
 * Pass per-feature callbacks through `useCallback` (or, as the existing
 * `use-flow-detail-navigation` / `use-template-detail-navigation` /
 * `use-knowledge-detail-navigation` do, hoist them to module scope). An
 * inline arrow would still work — `useLatestRef` reads the most recent
 * version at fire time — but it forfeits `useNavigation`'s internal
 * memoization on `getSearchableText`.
 */
export function useDetailNavigation<T extends { id: string }>({
    currentId,
    defaultOpen,
    defaultSearchQuery,
    getHref,
    getId,
    getLabel,
    getSearchableText,
    items,
    onOpenChange,
    onSearchQueryChange,
    open,
    searchDebounceMs,
    searchQuery,
    sortFn,
}: UseDetailNavigationOptions<T>): DetailNavigationController<T> {
    const navigate = useNavigate();
    const [searchParams] = useSearchParams();

    const { debouncedFilter } = useTableQueryFilterReader();

    // Memoize so passing `getId={undefined}` keeps the reference stable when
    // the caller re-renders for unrelated reasons. Without the memo the
    // downstream `useNavigation` memo would invalidate on every render.
    const resolvedGetId = useMemo<(item: T) => string>(() => getId ?? defaultGetId, [getId]);
    const resolvedGetSearchableText = useMemo<(item: T) => null | string | undefined>(
        () => getSearchableText ?? getLabel,
        [getSearchableText, getLabel],
    );

    // Controllable local search (same pattern as the sheet's `open` below).
    // Uncontrolled is the common case — most consumers want the controller to
    // own the value so the in-sheet input "just works". Controlled callers
    // (e.g. a page that mirrors the value to its own URL param) pass
    // `searchQuery` and observe via `onSearchQueryChange`.
    const onSearchQueryChangeRef = useLatestRef(onSearchQueryChange);
    const [internalSearchQuery, setInternalSearchQuery] = useState(defaultSearchQuery ?? '');
    const isSearchControlled = searchQuery !== undefined;
    const activeSearchQuery = isSearchControlled ? searchQuery : internalSearchQuery;
    const [debouncedSearchQuery] = useDebounce(activeSearchQuery, searchDebounceMs ?? DEFAULT_SEARCH_DEBOUNCE_MS);

    const setSearchQuery = useCallback(
        (next: string) => {
            if (!isSearchControlled) {
                setInternalSearchQuery(next);
            }

            onSearchQueryChangeRef.current?.(next);
        },
        [isSearchControlled, onSearchQueryChangeRef],
    );
    const clearSearchQuery = useCallback(() => setSearchQuery(''), [setSearchQuery]);

    // Stabilise the query tuple so `useNavigation`'s `useMemo` only refires
    // when either source actually moved. Inline `[a, b]` per render would
    // defeat the memo even when both strings are unchanged.
    const queryTerms = useMemo(
        () => [debouncedFilter, debouncedSearchQuery] as const,
        [debouncedFilter, debouncedSearchQuery],
    );

    const { currentIndex, currentItem, filteredItems, nextId, prevId, total } = useNavigation<T>({
        currentId,
        getId: resolvedGetId,
        getSearchableText: resolvedGetSearchableText,
        items,
        query: queryTerms,
        sortFn,
    });

    // Controlled-mode sheet state (mirrors `useControllable` from
    // `components/ui/autocomplete.tsx:27-46`). When `open` is provided the
    // caller owns the state; otherwise the controller owns it.
    // `onOpenChange` fires on every set so fully-controlled consumers can
    // observe transitions.
    const onOpenChangeRef = useLatestRef(onOpenChange);
    const [internalOpen, setInternalOpen] = useState(defaultOpen ?? false);
    const isOpenControlled = open !== undefined;
    const isSheetOpen = isOpenControlled ? open : internalOpen;

    const setSheetOpen = useCallback(
        (next: boolean) => {
            if (!isOpenControlled) {
                setInternalOpen(next);
            }

            onOpenChangeRef.current?.(next);
        },
        [isOpenControlled, onOpenChangeRef],
    );
    const openSheet = useCallback(() => setSheetOpen(true), [setSheetOpen]);
    const closeSheet = useCallback(() => setSheetOpen(false), [setSheetOpen]);

    // `useSearchParams` from React Router v6 returns a fresh `URLSearchParams`
    // every render. Threading it (or any caller-supplied accessor that might
    // not be stable) through `useCallback` deps would invalidate every
    // navigation callback on every render. Stash through `useLatestRef` and
    // read at fire-time instead — handlers fire from user clicks / keyboard
    // events, so the one-commit lag documented on `useLatestRef` never bites.
    const searchParamsRef = useLatestRef(searchParams);
    const getHrefRef = useLatestRef(getHref);
    const getIdRef = useLatestRef(resolvedGetId);
    const filteredItemsRef = useLatestRef(filteredItems);

    const buildHref = useCallback(
        (item: T) => mergeHrefWithSearchParams(getHrefRef.current(item), searchParamsRef.current),
        [getHrefRef, searchParamsRef],
    );

    const handleItemSelect = useCallback(
        (item: T) => {
            // Close the sheet *before* navigating so a route change can't unmount
            // the sheet while its close callback is still in flight.
            setSheetOpen(false);
            navigate(buildHref(item), { replace: true });
        },
        [buildHref, navigate, setSheetOpen],
    );

    const goTo = useCallback(
        (id: null | string) => {
            if (!id) {
                return;
            }

            const target = filteredItemsRef.current.find((item) => String(getIdRef.current(item)) === id);

            if (!target) {
                return;
            }

            navigate(buildHref(target), { replace: true });
        },
        [buildHref, filteredItemsRef, getIdRef, navigate],
    );

    const goToPrev = useCallback(() => goTo(prevId), [goTo, prevId]);
    const goToNext = useCallback(() => goTo(nextId), [goTo, nextId]);

    const positionLabel = useMemo(
        () => (total === 0 || currentIndex === -1 ? `–/${total}` : `${currentIndex + 1}/${total}`),
        [currentIndex, total],
    );

    const normalizedCurrentId = currentId != null ? String(currentId) : null;
    const hasEntries = filteredItems.length > 0;
    const itemsEmpty = items.length === 0;
    const isSearchActive = debouncedSearchQuery.length > 0;

    return useMemo<DetailNavigationController<T>>(
        () => ({
            clearSearchQuery,
            closeSheet,
            currentId: normalizedCurrentId,
            currentIndex,
            currentItem,
            debouncedFilter,
            debouncedSearchQuery,
            filteredItems,
            getId: resolvedGetId,
            getLabel,
            getSearchableText: resolvedGetSearchableText,
            goToNext,
            goToPrev,
            handleItemSelect,
            hasEntries,
            isSearchActive,
            isSheetOpen,
            itemsEmpty,
            nextId,
            openSheet,
            positionLabel,
            prevId,
            searchQuery: activeSearchQuery,
            setSearchQuery,
            setSheetOpen,
            total,
        }),
        [
            activeSearchQuery,
            clearSearchQuery,
            closeSheet,
            normalizedCurrentId,
            currentIndex,
            currentItem,
            debouncedFilter,
            debouncedSearchQuery,
            filteredItems,
            resolvedGetId,
            getLabel,
            resolvedGetSearchableText,
            goToNext,
            goToPrev,
            handleItemSelect,
            hasEntries,
            isSearchActive,
            isSheetOpen,
            itemsEmpty,
            nextId,
            openSheet,
            positionLabel,
            prevId,
            setSearchQuery,
            setSheetOpen,
            total,
        ],
    );
}
