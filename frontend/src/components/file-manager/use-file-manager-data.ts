import { useMemo } from 'react';

import type {
    FileManagerColumnsConfig,
    FileManagerInternalNode,
    FileManagerRootGroup,
    FileNode,
} from './file-manager-types';

import {
    buildFileManagerGridTemplate,
    buildFileManagerTree,
    collectAllNodePaths,
    collectSubtreePaths,
    filterFileManagerTree,
    normalizeRootGroups,
} from './file-manager-utils';

interface UseFileManagerDataArgs {
    columns: FileManagerColumnsConfig | undefined;
    files: FileNode[];
    /**
     * Whether the manager will render an actions column. Pulled from props by
     * the caller so the column-template stays in lock-step with the row layout
     * — `FileManager` and the row component must agree on the same `gridTemplate`.
     */
    hasActions: boolean;
    rootGroups: FileManagerRootGroup[] | undefined;
    /** Raw `search.query` from props; trimming and emptiness checks live inside the hook. */
    searchQuery: string | undefined;
}

interface UseFileManagerDataResult {
    /**
     * Universe of selectable paths (files + real directories) in the *visible*
     * tree, in DFS order. Used as the source of truth for "select all" and for
     * pruning stale selection entries when the file list changes.
     */
    allSelectablePaths: string[];
    /**
     * Pre-computed list of every selectable path in each directory's subtree
     * (the directory itself plus all descendants). Keyed by directory path.
     * Powers both the tri-state checkbox value and the "toggle whole subtree"
     * gesture without re-walking the tree on every render.
     */
    dirSubtreePaths: Map<string, readonly string[]>;
    /**
     * Full tree built from the input files (groups expanded, parent placeholders
     * filled). Required by drag-and-drop's `findNode` which needs every node,
     * including those filtered out of the visible tree by the search query.
     */
    fullTree: FileManagerInternalNode[];
    /** CSS `grid-template-columns` value shared by the header and every row. */
    gridTemplate: string;
    /** True when a non-empty search query is active. Drives auto-expand-on-search. */
    isFiltering: boolean;
    isModifiedVisible: boolean;
    isSizeVisible: boolean;
    /**
     * Normalized version of `rootGroups` (trailing slashes stripped). Returned
     * even when the result is reference-equal to the input, so consumers can
     * treat the value as the canonical group list.
     */
    normalizedRootGroups: FileManagerRootGroup[] | undefined;
    /** Trimmed search query (`''` when missing or whitespace-only). */
    trimmedSearch: string;
    /**
     * The tree as the user actually sees it — same shape as `fullTree` when no
     * filter is active, otherwise narrowed to nodes matching the search query
     * (with their ancestors and descendants kept).
     */
    visibleTree: FileManagerInternalNode[];
}

/**
 * Pure, memoized data layer for `FileManager`. Owns the full chain
 * `files → tree → visibleTree → derived collections (paths, subtree map, grid template)`
 * so the host component is free to focus on wiring (selection / expansion /
 * keyboard / DnD hooks) and rendering.
 *
 * `flatVisible` is intentionally NOT returned here — it depends on
 * `expandedPaths`, which depends on `visibleTree`, which depends on this hook.
 * Computing it inside would force a circular dependency between the data hook
 * and the expansion hook. The host derives it with a single `useMemo`.
 */
export const useFileManagerData = ({
    columns,
    files,
    hasActions,
    rootGroups,
    searchQuery,
}: UseFileManagerDataArgs): UseFileManagerDataResult => {
    const isSizeVisible = columns?.isSizeVisible ?? true;
    const isModifiedVisible = columns?.isModifiedVisible ?? true;

    const normalizedRootGroups = useMemo(() => normalizeRootGroups(rootGroups), [rootGroups]);

    const fullTree = useMemo(() => buildFileManagerTree(files, normalizedRootGroups), [files, normalizedRootGroups]);

    const trimmedSearch = searchQuery?.trim() ?? '';
    const isFiltering = trimmedSearch.length > 0;

    const visibleTree = useMemo(
        () => (isFiltering ? filterFileManagerTree(fullTree, trimmedSearch) : fullTree),
        [fullTree, isFiltering, trimmedSearch],
    );

    const allSelectablePaths = useMemo(() => collectAllNodePaths(visibleTree), [visibleTree]);

    // Pre-compute the subtree path list for every directory (and group root) once
    // per tree shape: this powers both the tri-state checkbox value and the
    // "toggle whole subtree" gesture without re-walking the tree on every render.
    const dirSubtreePaths = useMemo(() => {
        const map = new Map<string, readonly string[]>();

        const visit = (nodes: FileManagerInternalNode[]): void => {
            for (const node of nodes) {
                if (node.isDir || node.isGroupRoot) {
                    map.set(node.path, collectSubtreePaths(node));
                }

                if (node.children.length > 0) {
                    visit(node.children);
                }
            }
        };

        visit(visibleTree);

        return map;
    }, [visibleTree]);

    const gridTemplate = useMemo(
        () => buildFileManagerGridTemplate(isSizeVisible, isModifiedVisible, hasActions),
        [hasActions, isModifiedVisible, isSizeVisible],
    );

    return {
        allSelectablePaths,
        dirSubtreePaths,
        fullTree,
        gridTemplate,
        isFiltering,
        isModifiedVisible,
        isSizeVisible,
        normalizedRootGroups,
        trimmedSearch,
        visibleTree,
    };
};
