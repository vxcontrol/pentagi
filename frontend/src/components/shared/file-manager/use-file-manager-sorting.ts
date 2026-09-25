import { useCallback, useState } from 'react';

import type { FileManagerSortColumn, FileManagerSortState } from './file-manager-types';

interface UseFileManagerSortingResult {
    /** Current active sort. `null` means "no sort, insertion order preserved". */
    sorting: FileManagerSortState;
    /**
     * Cycle the sort state for the given column following the
     * `none → asc → desc → none` order — matches DataTable header behaviour.
     * - First click on an unsorted column → `asc`.
     * - Click on a column already sorted `asc` → flip to `desc`.
     * - Click on a column already sorted `desc` → clear.
     * - Click on a *different* column → switch to that column at `asc`.
     */
    toggleSort: (column: FileManagerSortColumn) => void;
}

export function useFileManagerSorting(): UseFileManagerSortingResult {
    const [sorting, setSorting] = useState<FileManagerSortState>(null);

    const toggleSort = useCallback((column: FileManagerSortColumn) => {
        setSorting((previous) => computeNextSort(previous, column));
    }, []);

    return { sorting, toggleSort };
}

/** Pure reducer for the header three-state cycle. Exported for unit tests. */
export const computeNextSort = (current: FileManagerSortState, column: FileManagerSortColumn): FileManagerSortState => {
    if (current?.column !== column) {
        return { column, direction: 'asc' };
    }

    if (current.direction === 'asc') {
        return { column, direction: 'desc' };
    }

    return null;
};
