import type { ColumnDef } from '@tanstack/react-table';

/**
 * Resolve a TanStack `ColumnDef`'s id the way TanStack does internally:
 * explicit `id` wins, then `accessorKey` when it's a plain string. Display
 * columns and `accessorFn` columns without an explicit id resolve to
 * `undefined` — callers must filter those out before passing the id to
 * APIs that require one (e.g. `getColumn`, predicate matching).
 */
export const getColumnId = <TData, TValue>(column: ColumnDef<TData, TValue>): string | undefined => {
    const withId = column as { id?: string };

    if (withId.id) {
        return withId.id;
    }

    const withAccessor = column as { accessorKey?: string };

    return typeof withAccessor.accessorKey === 'string' ? withAccessor.accessorKey : undefined;
};
