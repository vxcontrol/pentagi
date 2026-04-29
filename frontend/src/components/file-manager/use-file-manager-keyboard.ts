import { type KeyboardEvent as ReactKeyboardEvent, useCallback } from 'react';

import type { FileManagerInternalNode } from './file-manager-types';

import { clamp, findNodeByPath } from './file-manager-utils';

interface UseFileManagerKeyboardNavigationArgs {
    expandedPaths: Set<string>;
    /** Visible nodes in DFS order — universe of focusable rows. */
    flatVisible: string[];
    focusRow: (path: null | string) => void;
    isCheckboxVisible: boolean;
    onClearSelection: () => void;
    onSelectAll: () => void;
    onSetActiveRow: (path: null | string) => void;
    onToggleCheckbox: (path: string) => void;
    onToggleExpand: (path: string, wasExpanded: boolean) => void;
    /** Currently focused row, or `null` if nothing is focused. */
    resolvedActiveRow: null | string;
    visibleTree: FileManagerInternalNode[];
}

/**
 * Keyboard handler implementing the WAI-ARIA Tree pattern:
 * Arrow Up/Down → move focus, Home/End → jump to ends, Arrow Left/Right → collapse/expand
 * directories, Space → toggle the row's checkbox (only when checkboxes are shown),
 * Ctrl/Cmd+A → select all, Escape → clear selection.
 */
export const useFileManagerKeyboardNavigation = ({
    expandedPaths,
    flatVisible,
    focusRow,
    isCheckboxVisible,
    onClearSelection,
    onSelectAll,
    onSetActiveRow,
    onToggleCheckbox,
    onToggleExpand,
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
                case 'Spacebar':
                    event.preventDefault();

                    if (isCheckboxVisible) {
                        onToggleCheckbox(resolvedActiveRow);
                    }

                    return;
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
            onSelectAll,
            onSetActiveRow,
            onToggleCheckbox,
            onToggleExpand,
            resolvedActiveRow,
            visibleTree,
        ],
    );
