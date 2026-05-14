import type { Column } from '@tanstack/react-table';
import type { ReactNode } from 'react';

import { ArrowDown, ArrowUp } from 'lucide-react';

import { Button } from '@/components/ui/button';
import { cycleColumnSort } from '@/lib/table-sort';

interface SortableColumnHeaderProps<TData, TValue> {
    /** TanStack column. Sorting state and the cycle action both target it. */
    column: Column<TData, TValue>;
    /** Visible label rendered as the button text. */
    label: ReactNode;
}

/**
 * Reusable header for sortable `DataTable` columns. Wraps the conventional
 * "label + asc/desc arrow + onClick → cycleColumnSort" pattern so every list
 * page (Flows, Knowledges, Templates, Settings/*) reuses one component
 * instead of repeating ~20 lines per column header.
 *
 * The header reads sort direction directly from `column.getIsSorted()` so the
 * caller does not have to plumb it. The arrow icon mirrors the TanStack
 * convention: `asc` shows ↓ ("ascending continues to grow downward") and
 * `desc` shows ↑. No arrow means "not sorted by this column".
 */
export function SortableColumnHeader<TData, TValue = unknown>({
    column,
    label,
}: SortableColumnHeaderProps<TData, TValue>) {
    const sorted = column.getIsSorted();

    return (
        <Button
            className="text-muted-foreground hover:text-primary flex items-center gap-2 p-0 no-underline hover:no-underline"
            onClick={() => cycleColumnSort(column)}
            variant="link"
        >
            {label}
            {sorted === 'asc' ? <ArrowDown className="size-4" /> : null}
            {sorted === 'desc' ? <ArrowUp className="size-4" /> : null}
        </Button>
    );
}
