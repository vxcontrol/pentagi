import { useMemo } from 'react';

import type { ListNavigationToolbarProps } from '@/components/shared/list-navigation-toolbar';

import { usePageStorageKeys } from '@/hooks/use-page-storage-keys';
import { useTableQueryFilterReader } from '@/hooks/use-table-query-filter';

/**
 * Data-only slice of `ListNavigationToolbarProps<T>`. Derived from the
 * single source of truth in `list-navigation-toolbar.tsx` so a new required
 * prop (or a rename) can't silently desync the two surfaces.
 *
 * Presentation-only fields (`sheetTitle`, `sheetIcon`, `renderItem`,
 * `sortFn`) are intentionally excluded — those are supplied per-page at the
 * call site of the toolbar, not by the navigation hook.
 */
type ListNavigationToolbarDataProps<T> = Pick<
    ListNavigationToolbarProps<T>,
    'currentId' | 'filter' | 'getHref' | 'getId' | 'getLabel' | 'getSearchableText' | 'items'
>;

/**
 * Restricts the override to strings that start with a `/`. Pure-string types
 * (`string`) would let an empty `""` or a bare `"flows"` slip through and
 * either collide on the shared `filter_4_` key or generate a different slot
 * than the actual list page. Template-literal types catch this at compile
 * time without runtime guards.
 *
 * `${string}` is unconstrained on purpose — we only care that the value
 * begins with a slash, not what follows.
 */
type ParentPath = `/${string}`;

interface UseDetailNavigationOptions<T extends { id: string }> {
    currentId: null | string | undefined;
    getHref: (item: T) => string;
    getId?: (item: T) => string;
    getLabel: (item: T) => string;
    getSearchableText?: (item: T) => null | string | undefined;
    items: readonly T[];
    /**
     * Optional override for the parent list path used to look up the shared
     * filter storage slot. Defaults to the top-level segment of the current
     * pathname, which works for top-level routes (`/flows`, `/templates`,
     * `/knowledges`). Nested routes (`/admin/flows/:id`) must pass an
     * explicit value here because the default would key into `/admin`
     * rather than `/admin/flows`.
     *
     * Typed as a slash-prefixed string so the compiler rejects empty or
     * unprefixed overrides — those used to silently fall back to the
     * top-level path, which produced surprising key collisions. Pass an
     * explicit slash path (e.g. `'/admin/flows'`) when overriding.
     */
    parentPath?: ParentPath;
}

interface UseDetailNavigationResult<T extends { id: string }> {
    debouncedFilter: string;
    toolbarProps: ListNavigationToolbarDataProps<T>;
}

// Module-level so the reference is stable across renders. Typed against
// the wider `{ id: string }` so it is assignable to `(item: T) => string`
// for any `T extends { id: string }` via function-parameter contravariance
// — no cast at the call site.
const defaultGetId = (item: { id: string }): string => item.id;

/**
 * Convenience wrapper for detail pages that need a Prev/Next toolbar tied
 * to the parent list's free-text filter.
 *
 * The hook bundles the three things every detail page repeats:
 *   1. resolving the parent list's storage slot via `usePageStorageKeys`,
 *   2. subscribing to the URL filter through `useTableQueryFilterReader`
 *      (read-only — the detail page never mutates the filter from here),
 *   3. supplying a default `getId` (most domain types use a plain `id`).
 *
 * The returned `toolbarProps` plug straight into `<ListNavigationToolbar>` —
 * callers add presentation-only props (`sheetTitle`, `renderItem`, etc.).
 *
 * Pass each callback through `useCallback`: the toolbar's filter memo relies
 * on identity stability, and an inline arrow defeats it.
 */
export const useDetailNavigation = <T extends { id: string }>({
    currentId,
    getHref,
    getId,
    getLabel,
    getSearchableText,
    items,
    parentPath,
}: UseDetailNavigationOptions<T>): UseDetailNavigationResult<T> => {
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
    // `toolbarProps` memo below would invalidate on every render.
    const resolvedGetId = useMemo<(item: T) => string>(() => getId ?? defaultGetId, [getId]);

    const toolbarProps = useMemo<ListNavigationToolbarDataProps<T>>(
        () => ({
            currentId,
            filter: debouncedFilter,
            getHref,
            getId: resolvedGetId,
            getLabel,
            getSearchableText,
            items,
        }),
        [currentId, debouncedFilter, getHref, getLabel, getSearchableText, items, resolvedGetId],
    );

    return { debouncedFilter, toolbarProps };
};
