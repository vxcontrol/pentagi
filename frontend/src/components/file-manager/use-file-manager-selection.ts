import { type MouseEvent as ReactMouseEvent, useCallback, useEffect, useMemo, useRef, useState } from 'react';

import {
    computeRowClickSelection,
    computeToggleSelectAll,
    computeToggleSelection,
    resolveSelectionModifier,
} from './file-manager-utils';

interface UseFileManagerSelection {
    clearSelection: () => void;
    /** True when every selectable file is currently selected. */
    readonly isAllSelected: boolean;
    /** True when at least one — but not all — files are selected. */
    readonly isSomeSelected: boolean;
    /**
     * Click handler honoring single / toggle / range semantics.
     * Range computed against the supplied `flatVisible` order.
     *
     * When `subtreePaths` is provided (always for directory rows — the directory's
     * own path plus every descendant), the single / toggle modifiers operate on
     * the whole branch instead of just the row's path. This means a plain click
     * on a folder row replaces the selection with the entire subtree (matching
     * what the folder's tri-state checkbox would do), even when the folder is
     * collapsed and its descendants aren't currently rendered. Shift+click still
     * defines a visible range — the anchor is reset to the folder itself.
     */
    onRowClick: (event: ReactMouseEvent, path: string, subtreePaths?: readonly string[]) => void;
    /**
     * Polymorphic toggle used by row checkboxes and the Space keyboard shortcut.
     *
     *   - File rows / no `subtreePaths` → flip just `path`.
     *   - Directory rows pass the precomputed subtree (the directory itself plus
     *     every descendant) → all-or-nothing toggle of the whole branch: if every
     *     path is already selected the branch is cleared, otherwise the missing
     *     ones are added.
     *
     * `path` is also treated as the anchor for subsequent shift-range clicks.
     */
    onToggleSelection: (path: string, subtreePaths?: readonly string[]) => void;
    selectedPaths: Set<string>;
    /** Replace entire selection. */
    setSelection: (paths: Set<string>) => void;
    /** Toggle "select all": clears if everything was selected, otherwise picks every file. */
    toggleSelectAll: () => void;
}

interface UseFileManagerSelectionArgs {
    /** Universe of selectable paths (files + real directories) in the *visible* tree. */
    allSelectablePaths: string[];
    /** Visible nodes in DFS order — used for shift-range selection. */
    flatVisible: string[];
}

/**
 * Selection state without `useEffect`: `rawSelectedPaths` may contain stale entries
 * after files change; we derive a pruned `selectedPaths` during render. The pruned
 * Set keeps reference identity when nothing was stripped, so memoized consumers
 * downstream don't re-render needlessly.
 *
 * All real selection logic lives in pure reducers inside `file-manager-utils`
 * (`computeRowClickSelection`, `computeToggleSelection`, `computeToggleSelectAll`).
 * The hook is a thin wrapper that owns the state Set, the anchor ref, and a few
 * latest-ref pointers needed to keep the callbacks reference-stable across
 * expand/collapse and tree-shape changes.
 */
export const useFileManagerSelection = ({
    allSelectablePaths,
    flatVisible,
}: UseFileManagerSelectionArgs): UseFileManagerSelection => {
    const [rawSelectedPaths, setRawSelectedPaths] = useState<Set<string>>(() => new Set());
    const lastClickedRef = useRef<null | string>(null);

    // Stash the visible-order list and the universe of selectable paths in refs
    // so the click and select-all callbacks can stay reference-stable across
    // expand/collapse and tree-shape changes. Without this, `onRowClick` would
    // be re-created on every expansion (its deps include `flatVisible`) and
    // invalidate the `onClick` prop on every memoized row — kicking the entire
    // tree into a re-render on the most common user gesture. Same pattern is
    // already used in `use-file-manager-dnd` for `selectionRef`.
    const flatVisibleRef = useRef(flatVisible);
    const allSelectablePathsRef = useRef(allSelectablePaths);

    useEffect(() => {
        flatVisibleRef.current = flatVisible;
    }, [flatVisible]);

    useEffect(() => {
        allSelectablePathsRef.current = allSelectablePaths;
    }, [allSelectablePaths]);

    const allSelectablePathsSet = useMemo(() => new Set(allSelectablePaths), [allSelectablePaths]);

    const selectedPaths = useMemo(() => {
        if (rawSelectedPaths.size === 0) {
            return rawSelectedPaths;
        }

        let allValid = true;
        const valid = new Set<string>();

        for (const path of rawSelectedPaths) {
            if (allSelectablePathsSet.has(path)) {
                valid.add(path);
            } else {
                allValid = false;
            }
        }

        return allValid ? rawSelectedPaths : valid;
    }, [allSelectablePathsSet, rawSelectedPaths]);

    const onRowClick = useCallback((event: ReactMouseEvent, path: string, subtreePaths?: readonly string[]) => {
        const modifier = resolveSelectionModifier(event);

        setRawSelectedPaths((prev) => {
            const { next, nextAnchor } = computeRowClickSelection({
                anchor: lastClickedRef.current,
                flatVisible: flatVisibleRef.current,
                modifier,
                path,
                prev,
                subtreePaths,
            });

            // Storing into the ref from inside the updater is safe here: the
            // updater is called with a fresh `prev` on each invocation but the
            // anchor it produces depends only on `modifier`, `path`, the current
            // ref values and the supplied `subtreePaths` — so even if React
            // re-runs us in strict mode, every run computes the same `nextAnchor`.
            lastClickedRef.current = nextAnchor;

            return next;
        });
    }, []);

    const onToggleSelection = useCallback((path: string, subtreePaths?: readonly string[]) => {
        // Anchor follow-up shift-range clicks at the directory itself (or the
        // file's own path), not at the deepest leaf — matches what users expect
        // from Finder/Explorer.
        lastClickedRef.current = path;

        setRawSelectedPaths((prev) => computeToggleSelection({ path, prev, subtreePaths }));
    }, []);

    const isAllSelected = allSelectablePaths.length > 0 && selectedPaths.size === allSelectablePaths.length;
    const isSomeSelected = selectedPaths.size > 0 && !isAllSelected;

    // Functional setState reads the current selectable universe through the ref
    // declared above, so the callback identity stays stable regardless of how
    // many times the file list or selection changes.
    const toggleSelectAll = useCallback(() => {
        setRawSelectedPaths((prev) =>
            computeToggleSelectAll({ allSelectablePaths: allSelectablePathsRef.current, prev }),
        );
    }, []);

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
        onToggleSelection,
        selectedPaths,
        setSelection,
        toggleSelectAll,
    };
};
