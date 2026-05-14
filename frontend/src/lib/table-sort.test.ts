import type { Column } from '@tanstack/react-table';

import { describe, expect, it, vi } from 'vitest';

import { cycleColumnSort } from './table-sort';

interface FakeColumn {
    clearSorting: ReturnType<typeof vi.fn>;
    getIsSorted: ReturnType<typeof vi.fn>;
    toggleSorting: ReturnType<typeof vi.fn>;
}

// Build a stub that satisfies the subset of `Column` we actually call. Using
// `unknown as Column<...>` keeps the test free of @tanstack/react-table
// internals while still letting us pass the stub to the typed function.
const makeColumn = (sorted: 'asc' | 'desc' | false): FakeColumn => ({
    clearSorting: vi.fn(),
    getIsSorted: vi.fn().mockReturnValue(sorted),
    toggleSorting: vi.fn(),
});

describe('cycleColumnSort', () => {
    it('starts the cycle on a column with no sort by setting ascending', () => {
        const column = makeColumn(false);
        cycleColumnSort(column as unknown as Column<unknown>);

        expect(column.toggleSorting).toHaveBeenCalledWith(false);
        expect(column.clearSorting).not.toHaveBeenCalled();
    });

    it('advances ascending to descending', () => {
        const column = makeColumn('asc');
        cycleColumnSort(column as unknown as Column<unknown>);

        expect(column.toggleSorting).toHaveBeenCalledWith(true);
        expect(column.clearSorting).not.toHaveBeenCalled();
    });

    it('clears the sort after descending — the third tap removes the sort entirely', () => {
        const column = makeColumn('desc');
        cycleColumnSort(column as unknown as Column<unknown>);

        expect(column.clearSorting).toHaveBeenCalledTimes(1);
        expect(column.toggleSorting).not.toHaveBeenCalled();
    });

    it('reads `getIsSorted` exactly once per call — no double-evaluation between branches', () => {
        const column = makeColumn('asc');
        cycleColumnSort(column as unknown as Column<unknown>);

        expect(column.getIsSorted).toHaveBeenCalledTimes(1);
    });
});
