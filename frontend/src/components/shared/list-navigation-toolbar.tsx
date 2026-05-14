import { type ReactNode, useCallback, useMemo, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';

import { ListNavigationButtons } from '@/components/shared/list-navigation-buttons';
import { ListNavigationSheet } from '@/components/shared/list-navigation-sheet';
import { useFilteredListNavigation } from '@/hooks/use-filtered-list-navigation';
import { mergeHrefWithSearchParams } from '@/lib/url-params';

export interface ListNavigationToolbarProps<T> {
    currentId: null | string | undefined;
    /**
     * Current free-text filter the list page applies. The toolbar narrows the
     * navigable subset by running the same `matchesTextFilter` semantics the
     * list uses — so Prev/Next never drifts out of sync with what the user
     * sees in the table.
     */
    filter: string;
    /**
     * Build a navigation target for the given item. The current value of
     * `?<filterParamName>=` is forwarded so detail-page navigation preserves
     * the filter the user picked on the list.
     */
    getHref: (item: T) => string;
    /**
     * Stable accessor for the row's ID. The toolbar uses it for keying both
     * the navigation lookup and the rendered Sheet list.
     */
    getId: (item: T) => string;
    /**
     * Default label for items in the Sheet list. Used when `renderItem` is
     * not provided. Also acts as the default haystack for the filter when
     * `getSearchableText` is omitted (matches the common case where the list
     * column filter targets the same field as the label).
     */
    getLabel: (item: T) => string;
    /**
     * Optional override for the haystack the filter runs against. Pass when
     * the list column filter is keyed on a different field than the visible
     * label (e.g. label is `title || #id`, filter is plain `title`).
     */
    getSearchableText?: (item: T) => null | string | undefined;
    items: readonly T[];
    /**
     * Optional override for the row in the Sheet. Receives the item and a
     * boolean indicating whether it's the current detail page.
     */
    renderItem?: (item: T, isCurrent: boolean) => ReactNode;
    sheetIcon?: ReactNode;
    sheetTitle: string;
    /** Optional comparator. When omitted, items appear in input order. */
    sortFn?: (a: T, b: T) => number;
}

/**
 * Prev / Position / Next toolbar shown on detail pages whose list page
 * supports a free-text filter.
 *
 * - "Prev" / "Next" walk the same filtered subset the list page renders.
 * - The middle button opens `<ListNavigationSheet>`, a listbox of every
 *   entry in the filtered subset with the current one highlighted — the
 *   user can jump anywhere in one click.
 * - When the current item is missing from the filtered subset (e.g. the
 *   server-loaded record doesn't match the active filter, or the record
 *   was just deleted) prev/next are disabled and the position label falls
 *   back to `"–/total"` so the user can still open the Sheet to pick a
 *   neighbour.
 * - Sibling navigation uses `replace: true` so paging through a list of 50
 *   items doesn't push 50 history entries the user has to back through.
 *
 * Returns `null` when there are no items at all — there's nothing useful to
 * render and a "–/0" disabled toolbar would just be visual noise.
 */
export function ListNavigationToolbar<T>({
    currentId,
    filter,
    getHref,
    getId,
    getLabel,
    getSearchableText,
    items,
    renderItem,
    sheetIcon,
    sheetTitle,
    sortFn,
}: ListNavigationToolbarProps<T>) {
    const navigate = useNavigate();
    const [searchParams] = useSearchParams();
    const [isSheetOpen, setIsSheetOpen] = useState(false);

    // The toolbar owns the haystack-resolution rule so list and detail always
    // agree on what "matches the filter" means — callers only describe the
    // data shape (`getSearchableText` / `getLabel`), never the comparison.
    const haystack = getSearchableText ?? getLabel;

    const { currentIndex, filteredItems, nextId, prevId, total } = useFilteredListNavigation({
        currentId,
        getId,
        getSearchableText: haystack,
        items,
        query: filter,
        sortFn,
    });

    // Forward the current filter (and any other URL state) when navigating
    // to a sibling — without it the user would lose `?q=` the moment they
    // hit Prev/Next, breaking the "stay inside the filtered subset" promise.
    // `mergeHrefWithSearchParams` round-trips through `URL` so the hash
    // fragment survives and getHref's own query params win on collision.
    const buildHref = useCallback(
        (item: T) => mergeHrefWithSearchParams(getHref(item), searchParams),
        [getHref, searchParams],
    );

    const goTo = useCallback(
        (id: null | string) => {
            if (!id) {
                return;
            }

            const target = filteredItems.find((item) => String(getId(item)) === id);

            if (!target) {
                return;
            }

            navigate(buildHref(target), { replace: true });
        },
        [buildHref, filteredItems, getId, navigate],
    );

    const handleItemSelect = useCallback(
        (item: T) => {
            setIsSheetOpen(false);
            navigate(buildHref(item), { replace: true });
        },
        [buildHref, navigate],
    );

    const goToPrev = useCallback(() => goTo(prevId), [goTo, prevId]);
    const goToNext = useCallback(() => goTo(nextId), [goTo, nextId]);
    const openSheet = useCallback(() => setIsSheetOpen(true), []);

    const positionLabel = useMemo(
        () => (total === 0 || currentIndex === -1 ? `–/${total}` : `${currentIndex + 1}/${total}`),
        [currentIndex, total],
    );

    // Nothing meaningful to render before items show up — saves the user from
    // a momentary "–/0" flash while the parent provider's data is in flight.
    if (items.length === 0) {
        return null;
    }

    return (
        <>
            <ListNavigationButtons
                hasEntries={total > 0}
                nextDisabled={!nextId}
                onNext={goToNext}
                onOpen={openSheet}
                onPrev={goToPrev}
                positionLabel={positionLabel}
                prevDisabled={!prevId}
                sheetTitle={sheetTitle}
            />
            <ListNavigationSheet<T>
                currentId={currentId}
                currentIndex={currentIndex}
                getId={getId}
                getLabel={getLabel}
                items={filteredItems}
                onItemSelect={handleItemSelect}
                onOpenChange={setIsSheetOpen}
                open={isSheetOpen}
                renderItem={renderItem}
                sheetIcon={sheetIcon}
                sheetTitle={sheetTitle}
                total={total}
            />
        </>
    );
}
