import { type KeyboardEvent as ReactKeyboardEvent, useCallback } from 'react';

import type { FileManagerInternalNode, FileNode } from './file-manager-types';

import { clamp, collectSubtreePaths, findNodeByPath } from './file-manager-utils';

interface UseFileManagerKeyboardNavigationArgs {
    expandedPaths: Set<string>;
    /** Visible nodes in DFS order — universe of focusable rows. */
    flatVisible: string[];
    focusRow: (path: null | string) => void;
    isCheckboxVisible: boolean;
    onClearSelection: () => void;
    /** Open handler for file rows (Enter). Directories always toggle expand. */
    onOpen?: (file: FileNode) => void;
    onSelectAll: () => void;
    onSetActiveRow: (path: null | string) => void;
    onToggleExpand: (path: string, wasExpanded: boolean) => void;
    /** Polymorphic selection toggle — same shape as the row checkbox handler. */
    onToggleSelection: (path: string, subtreePaths?: readonly string[]) => void;
    /** Currently focused row, or `null` if nothing is focused. */
    resolvedActiveRow: null | string;
    visibleTree: FileManagerInternalNode[];
}

/**
 * Keyboard handler implementing the WAI-ARIA Tree pattern:
 * Arrow Up/Down → move focus, Home/End → jump to ends, Arrow Left/Right → collapse/expand
 * directories, Enter → expand/collapse a directory or fire `onOpen` for a file,
 * Space → toggle the row's checkbox (only when checkboxes are shown),
 * Ctrl/Cmd+A → select all, Escape → clear selection.
 */
export const useFileManagerKeyboardNavigation = ({
    expandedPaths,
    flatVisible,
    focusRow,
    isCheckboxVisible,
    onClearSelection,
    onOpen,
    onSelectAll,
    onSetActiveRow,
    onToggleExpand,
    onToggleSelection,
    resolvedActiveRow,
    visibleTree,
}: UseFileManagerKeyboardNavigationArgs) =>
    useCallback(
        (event: ReactKeyboardEvent<HTMLDivElement>) => {
            if (event.key === 'Escape') {
                onClearSelection();

                return;
            }

            if (!resolvedActiveRow || flatVisible.length === 0) {
                return;
            }

            const index = flatVisible.indexOf(resolvedActiveRow);

            const moveTo = (nextIndex: number) => {
                const nextPath = flatVisible[clamp(0, nextIndex, flatVisible.length - 1)] ?? null;

                onSetActiveRow(nextPath);
                focusRow(nextPath);
            };

            switch (event.key) {
                case ' ':
                case 'Spacebar': {
                    event.preventDefault();

                    if (!isCheckboxVisible) {
                        return;
                    }

                    // Mirror the click semantics of the row checkbox: pressing
                    // Space on a directory flips its whole subtree (the dir +
                    // every descendant), on a file it toggles just that path.
                    const node = findNodeByPath(visibleTree, resolvedActiveRow);
                    const subtreePaths =
                        node && (node.isDir || node.isGroupRoot) ? collectSubtreePaths(node) : undefined;

                    onToggleSelection(resolvedActiveRow, subtreePaths);

                    return;
                }

                case 'A':
                case 'a':
                    if (event.ctrlKey || event.metaKey) {
                        event.preventDefault();
                        onSelectAll();
                    }

                    return;
                case 'ArrowDown':
                    event.preventDefault();
                    moveTo(index + 1);

                    return;
                case 'ArrowLeft':
                case 'ArrowRight': {
                    const node = findNodeByPath(visibleTree, resolvedActiveRow);

                    if (!node?.isDir) {
                        return;
                    }

                    const isExpanded = expandedPaths.has(node.path);
                    const wantsExpand = event.key === 'ArrowRight';

                    if (wantsExpand !== isExpanded) {
                        event.preventDefault();
                        onToggleExpand(node.path, isExpanded);
                    }

                    return;
                }

                case 'ArrowUp':
                    event.preventDefault();
                    moveTo(index - 1);

                    return;
                case 'End':
                    event.preventDefault();
                    moveTo(flatVisible.length - 1);

                    return;
                case 'Enter': {
                    const node = findNodeByPath(visibleTree, resolvedActiveRow);

                    if (!node) {
                        return;
                    }

                    // Always swallow Enter on a focused row so an enclosing form
                    // doesn't accidentally submit when the user just wanted to
                    // open a file or toggle a folder.
                    event.preventDefault();

                    if (node.isDir) {
                        onToggleExpand(node.path, expandedPaths.has(node.path));

                        return;
                    }

                    onOpen?.(node);

                    return;
                }

                case 'Home':
                    event.preventDefault();
                    moveTo(0);

                    return;
                default:
                    return;
            }
        },
        [
            expandedPaths,
            flatVisible,
            focusRow,
            isCheckboxVisible,
            onClearSelection,
            onOpen,
            onSelectAll,
            onSetActiveRow,
            onToggleExpand,
            onToggleSelection,
            resolvedActiveRow,
            visibleTree,
        ],
    );
