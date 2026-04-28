import { type MouseEvent as ReactMouseEvent, useCallback, useMemo, useRef, useState } from 'react';

import { resolveSelectionModifier } from './file-manager-utils';

interface UseFileManagerSelection {
    clearSelection: () => void;
    /** True when every selectable file is currently selected. */
    readonly isAllSelected: boolean;
    /** True when at least one — but not all — files are selected. */
    readonly isSomeSelected: boolean;
    /**
     * Click handler honoring single / toggle / range semantics.
     * Range computed against the supplied `flatVisible` order.
     */
    onRowClick: (event: ReactMouseEvent, path: string) => void;
    /** Toggle a single path — used by checkboxes. */
    onToggleCheckbox: (path: string) => void;
    selectedPaths: Set<string>;
    /** Replace entire selection. */
    setSelection: (paths: Set<string>) => void;
    /** Toggle "select all": clears if everything was selected, otherwise picks every file. */
    toggleSelectAll: () => void;
}

interface UseFileManagerSelectionArgs {
    /** Universe of selectable file paths in the *visible* tree. */
    allFilePaths: string[];
    /** Visible nodes in DFS order — used for shift-range selection. */
    flatVisible: string[];
}

/**
 * Selection state without `useEffect`: `rawSelectedPaths` may contain stale entries
 * after files change; we derive a pruned `selectedPaths` during render. The pruned
 * Set keeps reference identity when nothing was stripped, so memoized consumers
 * downstream don't re-render needlessly.
 */
export const useFileManagerSelection = ({
    allFilePaths,
    flatVisible,
}: UseFileManagerSelectionArgs): UseFileManagerSelection => {
    const [rawSelectedPaths, setRawSelectedPaths] = useState<Set<string>>(() => new Set());
    const lastClickedRef = useRef<null | string>(null);

    const allFilePathsSet = useMemo(() => new Set(allFilePaths), [allFilePaths]);

    const selectedPaths = useMemo(() => {
        if (rawSelectedPaths.size === 0) {
            return rawSelectedPaths;
        }

        let allValid = true;
        const valid = new Set<string>();

        for (const path of rawSelectedPaths) {
            if (allFilePathsSet.has(path)) {
                valid.add(path);
            } else {
                allValid = false;
            }
        }

        return allValid ? rawSelectedPaths : valid;
    }, [allFilePathsSet, rawSelectedPaths]);

    const onRowClick = useCallback(
        (event: ReactMouseEvent, path: string) => {
            const modifier = resolveSelectionModifier(event);

            setRawSelectedPaths((prev) => {
                if (modifier === 'single') {
                    lastClickedRef.current = path;

                    return new Set([path]);
                }

                if (modifier === 'toggle') {
                    const next = new Set(prev);

                    if (next.has(path)) {
                        next.delete(path);
                    } else {
                        next.add(path);
                    }

                    lastClickedRef.current = path;

                    return next;
                }

                if (!lastClickedRef.current) {
                    lastClickedRef.current = path;

                    return new Set([path]);
                }

                const fromIdx = flatVisible.indexOf(lastClickedRef.current);
                const toIdx = flatVisible.indexOf(path);

                if (fromIdx < 0 || toIdx < 0) {
                    return new Set([path]);
                }

                const [a, b] = fromIdx <= toIdx ? [fromIdx, toIdx] : [toIdx, fromIdx];

                return new Set(flatVisible.slice(a, b + 1));
            });
        },
        [flatVisible],
    );

    const onToggleCheckbox = useCallback((path: string) => {
        setRawSelectedPaths((prev) => {
            const next = new Set(prev);

            if (next.has(path)) {
                next.delete(path);
            } else {
                next.add(path);
            }

            lastClickedRef.current = path;

            return next;
        });
    }, []);

    const isAllSelected = allFilePaths.length > 0 && selectedPaths.size === allFilePaths.length;
    const isSomeSelected = selectedPaths.size > 0 && !isAllSelected;

    const toggleSelectAll = useCallback(() => {
        setRawSelectedPaths(() => (isAllSelected ? new Set() : new Set(allFilePaths)));
    }, [allFilePaths, isAllSelected]);

    const clearSelection = useCallback(() => {
        setRawSelectedPaths((prev) => (prev.size === 0 ? prev : new Set()));
    }, []);

    const setSelection = useCallback((paths: Set<string>) => {
        setRawSelectedPaths(paths);
    }, []);

    return {
        clearSelection,
        isAllSelected,
        isSomeSelected,
        onRowClick,
        onToggleCheckbox,
        selectedPaths,
        setSelection,
        toggleSelectAll,
    };
};
