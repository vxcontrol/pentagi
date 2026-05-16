import { useCallback, useMemo, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';

import { useLatestRef } from '@/hooks/use-latest-ref';
import { usePageStorageKeys } from '@/hooks/use-page-storage-keys';
import { useTableQueryFilterReader } from '@/hooks/use-table-query-filter';
import { mergeHrefWithSearchParams } from '@/lib/url-params';

import { useNavigation } from './use-navigation';

/**
 * Headless controller for a detail page that walks a filtered list.
 *
 * Owns:
 *   - the filtered/sorted subset (pure `computeNavigation`),
 *   - the resolved Prev / Next sibling ids and the pre-formatted position
 *     label (so leaf components don't recompute),
 *   - the sheet open state (controllable via `open` / `onOpenChange`),
 *   - navigation actions that thread the current `?<filter>=` into every
 *     prev / next / item-select destination.
 *
 * All callbacks are identity-stable across renders that don't change the
 * inputs the action depends on — `<DetailNavigationButtons>`,
 * `<DetailNavigationSheet>`, and any custom chrome rendered against the
 * controller can rely on referential equality for downstream memos.
 */
export interface DetailNavigationController<T extends { id: string }> {
    closeSheet: () => void;
    /** Active `currentId` coerced to a string, or `null` when absent. */
    currentId: null | string;
    /** Index of `currentItem` inside `filteredItems`. `-1` when not in subset. */
    currentIndex: number;
    /** Item matching `currentId` inside the filtered subset, or `null`. */
    currentItem: null | T;
    /** Debounced URL filter the controller is filtering against. */
    debouncedFilter: string;
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
    setSheetOpen: (open: boolean) => void;

    /** Same as `filteredItems.length`, named explicitly for clarity. */
    total: number;
}

/**
 * Restricts the override to strings that start with a `/`. Pure-string types
 * (`string`) would let an empty `""` or a bare `"flows"` slip through and
 * either collide on the shared `filter_4_` key or generate a different slot
 * than the actual list page. Template-literal types catch this at compile
 * time without runtime guards.
 */
type ParentPath = `/${string}`;

interface UseDetailNavigationOptions<T extends { id: string }> {
    currentId: null | string | undefined;
    /** Initial value for the uncontrolled case. Defaults to `false`. */
    defaultOpen?: boolean;
    getHref: (item: T) => string;
    getId?: (item: T) => string;
    getLabel: (item: T) => string;
    getSearchableText?: (item: T) => null | string | undefined;
    items: readonly T[];
    onOpenChange?: (open: boolean) => void;

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
     * Optional override for the parent list path used to look up the shared
     * filter storage slot. Defaults to the top-level segment of the current
     * pathname, which works for top-level routes (`/flows`, `/templates`,
     * `/knowledges`). Nested routes (`/admin/flows/:id`) must pass an
     * explicit value here because the default would key into `/admin`
     * rather than `/admin/flows`.
     */
    parentPath?: ParentPath;
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
 * Bundles the four moving parts every detail page repeats:
 *   1. resolving the parent list's storage slot via `usePageStorageKeys`,
 *   2. subscribing to the URL filter through `useTableQueryFilterReader`
 *      (read-only — the detail page never mutates the filter from here),
 *   3. running the pure `useNavigation` core against the filtered subset,
 *   4. wiring identity-stable `goToPrev` / `goToNext` / `handleItemSelect`
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
export const useDetailNavigation = <T extends { id: string }>({
    currentId,
    defaultOpen,
    getHref,
    getId,
    getLabel,
    getSearchableText,
    items,
    onOpenChange,
    open,
    parentPath,
    sortFn,
}: UseDetailNavigationOptions<T>): DetailNavigationController<T> => {
    const navigate = useNavigate();
    const [searchParams] = useSearchParams();

    // `parentPath` is typed as `/${string}` so the empty / unprefixed case
    // is rejected at compile time — we only need to branch on presence here.
    const hasExplicitParentPath = parentPath !== undefined;
    const { table: filterStorageKey } = usePageStorageKeys({
        pathname: hasExplicitParentPath ? parentPath : undefined,
        useTopLevel: !hasExplicitParentPath,
    });
    const { debouncedFilter } = useTableQueryFilterReader({ storageKey: filterStorageKey });

    // Memoize so passing `getId={undefined}` keeps the reference stable when
    // the caller re-renders for unrelated reasons. Without the memo the
    // downstream `useNavigation` memo would invalidate on every render.
    const resolvedGetId = useMemo<(item: T) => string>(() => getId ?? defaultGetId, [getId]);
    const resolvedGetSearchableText = useMemo<(item: T) => null | string | undefined>(
        () => getSearchableText ?? getLabel,
        [getSearchableText, getLabel],
    );

    const { currentIndex, currentItem, filteredItems, nextId, prevId, total } = useNavigation<T>({
        currentId,
        getId: resolvedGetId,
        getSearchableText: resolvedGetSearchableText,
        items,
        query: debouncedFilter,
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
            // Close the sheet *before* navigating — preserves the pre-refactor
            // ordering so a route change can't unmount the sheet while its
            // close callback is still in flight.
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

    return useMemo<DetailNavigationController<T>>(
        () => ({
            closeSheet,
            currentId: normalizedCurrentId,
            currentIndex,
            currentItem,
            debouncedFilter,
            filteredItems,
            getId: resolvedGetId,
            getLabel,
            getSearchableText: resolvedGetSearchableText,
            goToNext,
            goToPrev,
            handleItemSelect,
            hasEntries,
            isSheetOpen,
            itemsEmpty,
            nextId,
            openSheet,
            positionLabel,
            prevId,
            setSheetOpen,
            total,
        }),
        [
            closeSheet,
            normalizedCurrentId,
            currentIndex,
            currentItem,
            debouncedFilter,
            filteredItems,
            resolvedGetId,
            getLabel,
            resolvedGetSearchableText,
            goToNext,
            goToPrev,
            handleItemSelect,
            hasEntries,
            isSheetOpen,
            itemsEmpty,
            nextId,
            openSheet,
            positionLabel,
            prevId,
            setSheetOpen,
            total,
        ],
    );
};
