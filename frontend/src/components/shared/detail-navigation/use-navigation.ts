import { useMemo } from 'react';

import { createTextMatcher } from './text-filter';

interface NavigationInput<T> {
    currentId: null | string | undefined;
    getId: (item: T) => string;
    /**
     * Optional accessor for the text the filter runs against. Only consulted
     * when `query` is non-empty — pages with no filter don't need it.
     */
    getSearchableText?: (item: T) => null | string | undefined;
    items: readonly T[];
    /**
     * Free-text filter to narrow the navigable subset. Empty / `undefined` is
     * treated as "no filter" and every item passes. Matching reuses
     * `createTextMatcher`, so behaviour matches the list page's column
     * filter (substring, case + diacritic insensitive).
     */
    query?: string;
    sortFn?: (a: T, b: T) => number;
}

interface NavigationResult<T> {
    currentIndex: number;
    currentItem: null | T;
    filteredItems: readonly T[];
    nextId: null | string;
    prevId: null | string;
    total: number;
}

/**
 * Pure core of {@link useNavigation}: filter, sort, and resolve Prev/Next
 * around `currentId`. Exposed without React so the algorithm can be
 * unit-tested directly — the hook is a thin `useMemo` wrapper around this.
 *
 * The Map of `id → index` is built once per invocation so the `currentId`
 * lookup is O(1) instead of an `Array.find` per render. The map is
 * intentionally not exposed — callers should walk `filteredItems` or use the
 * returned `prevId` / `nextId`.
 */
export const computeNavigation = <T>({
    currentId,
    getId,
    getSearchableText,
    items,
    query,
    sortFn,
}: NavigationInput<T>): NavigationResult<T> => {
    const hasQuery = query !== undefined && query.length > 0;
    const filtered =
        hasQuery && getSearchableText
            ? items.filter((item) => createTextMatcher(query)(getSearchableText(item)))
            : items;
    const ordered = sortFn ? [...filtered].sort(sortFn) : filtered;

    const indexById = new Map<string, number>();

    ordered.forEach((item, index) => {
        // Apollo commonly hydrates GraphQL `ID` fields as numbers even though
        // route params are strings — normalize so lookups stay stable.
        indexById.set(String(getId(item)), index);
    });

    const currentIndex = currentId != null ? (indexById.get(String(currentId)) ?? -1) : -1;

    if (currentIndex === -1) {
        return {
            currentIndex: -1,
            currentItem: null,
            filteredItems: ordered,
            nextId: null,
            prevId: null,
            total: ordered.length,
        };
    }

    const prevItem = currentIndex > 0 ? ordered[currentIndex - 1] : null;
    const nextItem = currentIndex < ordered.length - 1 ? ordered[currentIndex + 1] : null;

    return {
        currentIndex,
        currentItem: ordered[currentIndex] ?? null,
        filteredItems: ordered,
        nextId: nextItem ? String(getId(nextItem)) : null,
        prevId: prevItem ? String(getId(prevItem)) : null,
        total: ordered.length,
    };
};

interface UseNavigationOptions<T> {
    currentId: null | string | undefined;
    getId: (item: T) => string;
    /**
     * Accessor for the text the filter runs against. Only required when
     * `query` is non-empty.
     */
    getSearchableText?: (item: T) => null | string | undefined;
    items: readonly T[];
    /** Free-text filter. Empty / `undefined` → unfiltered. */
    query?: string;
    /**
     * Optional comparator. When omitted, the input array order is preserved —
     * the navigation walks `items` exactly as the caller arranged them, which
     * matches what list pages already render.
     */
    sortFn?: (a: T, b: T) => number;
}

type UseNavigationResult<T> = NavigationResult<T>;

/**
 * Resolve Prev/Next siblings for a detail page that's tied to a filtered list.
 *
 * The list page holds the same `items` and free-text `query`, so this hook
 * produces the same ordering as what the user sees in the table — Prev/Next
 * stays in lockstep with the filter even after the user lands on a detail
 * URL via a shared link.
 *
 * If the current item is missing from the filtered subset (deleted, or no
 * longer matches the filter), `currentIndex` is `-1` and both `prevId` /
 * `nextId` are `null` — the caller renders disabled Prev/Next buttons in that
 * case rather than navigating into an unrelated neighbour.
 *
 * Callers pass *data* (`query`, `getSearchableText`) rather than a prebuilt
 * predicate. That removes the identity-stability footgun the predicate-based
 * shape had: a fresh-arrow `(item) => …` callback every render would have
 * defeated the internal `useMemo`. With this shape, only `query` flips per
 * keystroke, and `getSearchableText` is naturally module-scoped at the
 * feature level (see `use-flow-detail-navigation` etc.).
 */
export const useNavigation = <T>({
    currentId,
    getId,
    getSearchableText,
    items,
    query,
    sortFn,
}: UseNavigationOptions<T>): UseNavigationResult<T> => {
    return useMemo(
        () => computeNavigation({ currentId, getId, getSearchableText, items, query, sortFn }),
        [currentId, getId, getSearchableText, items, query, sortFn],
    );
};
