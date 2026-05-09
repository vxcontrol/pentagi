import type { Column } from '@tanstack/react-table';

/**
 * Cycle a TanStack Table column through the three-state sort order
 * `none → asc → desc → none`. Mirrors the header behaviour of `DataTable`
 * and the FileManager sortable headers.
 *
 * Pure with respect to React (does not call any hook), so it can be invoked
 * directly from a header `onClick` without wrapping in `useCallback`/`useMemo`.
 *
 * Generic over the row type so `column` keeps its full TanStack typing at the
 * call site — no need to inline a structural subset like
 * `{ getIsSorted; toggleSorting; clearSorting }` in every page.
 */
export const cycleColumnSort = <TData, TValue = unknown>(column: Column<TData, TValue>): void => {
    const sorted = column.getIsSorted();

    if (sorted === 'asc') {
        column.toggleSorting(true);

        return;
    }

    if (sorted === 'desc') {
        column.clearSorting();

        return;
    }

    column.toggleSorting(false);
};
