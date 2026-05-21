import { useCallback, useEffect, useRef, useState } from 'react';

import { useEffectAfterMount } from '@/hooks/use-effect-after-mount';

import type { FileManagerSortColumn, FileManagerSortState } from './file-manager-types';

import { loadFileManagerSorting, saveFileManagerSorting } from './file-manager-storage';

interface UseFileManagerSortingArgs {
    /**
     * Controlled sort value. When provided, the hook does not own state and
     * delegates updates to `onSortingChange`. Persistence (`sortStorageKey`)
     * is also bypassed — the parent owns the lifecycle.
     */
    controlledSorting?: FileManagerSortState;
    /**
     * Initial value used when no `sortStorageKey` payload is found. Ignored in
     * controlled mode (when `controlledSorting` is not undefined).
     */
    initialSorting?: FileManagerSortState;
    /** Fires whenever the active sort changes (uncontrolled mode only). */
    onSortingChange?: (sorting: FileManagerSortState) => void;
    /**
     * `localStorage` key for persisting the active sort between reloads.
     * Ignored in controlled mode. Pass `undefined` to disable persistence.
     */
    sortStorageKey?: string;
}

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

/**
 * Owns the active sort for a `FileManager`. Supports three modes:
 *   1. **Uncontrolled, ephemeral** — `sortStorageKey` and `controlledSorting`
 *      both omitted. State lives only for the component lifetime.
 *   2. **Uncontrolled, persisted** — `sortStorageKey` is set. State is loaded
 *      from `localStorage` on mount and rewritten whenever the user changes it.
 *   3. **Controlled** — `controlledSorting` is set. The hook reflects that
 *      value verbatim and fires `onSortingChange` on user interaction.
 */
export function useFileManagerSorting({
    controlledSorting,
    initialSorting = null,
    onSortingChange,
    sortStorageKey,
}: UseFileManagerSortingArgs): UseFileManagerSortingResult {
    const isControlled = controlledSorting !== undefined;

    // `localStorage` access is guarded inside `loadFileManagerSorting`, so the
    // lazy initializer is safe in SSR-style environments too.
    const [internalSorting, setInternalSorting] = useState<FileManagerSortState>(() => {
        if (isControlled) {
            return controlledSorting ?? null;
        }

        if (sortStorageKey) {
            const stored = loadFileManagerSorting(sortStorageKey);

            if (stored !== null) {
                return stored;
            }
        }

        return initialSorting;
    });

    const sorting = isControlled ? (controlledSorting ?? null) : internalSorting;

    useEffectAfterMount(() => {
        if (!isControlled && sortStorageKey) {
            saveFileManagerSorting(sortStorageKey, internalSorting);
        }
    }, [internalSorting, sortStorageKey, isControlled]);

    // Latest-ref trick keeps `toggleSort` stable across re-renders even when
    // the parent passes a fresh `onSortingChange` callback or the resolved
    // sort value changes between renders.
    const onSortingChangeRef = useRef(onSortingChange);
    const sortingRef = useRef(sorting);

    useEffect(() => {
        onSortingChangeRef.current = onSortingChange;
    }, [onSortingChange]);

    useEffect(() => {
        sortingRef.current = sorting;
    }, [sorting]);

    const isControlledRef = useRef(isControlled);

    useEffect(() => {
        isControlledRef.current = isControlled;
    }, [isControlled]);

    const toggleSort = useCallback((column: FileManagerSortColumn) => {
        const next = computeNextSort(sortingRef.current, column);

        if (!isControlledRef.current) {
            setInternalSorting(next);
        }

        onSortingChangeRef.current?.(next);
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
