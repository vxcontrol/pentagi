import { type MouseEvent as ReactMouseEvent, useCallback, useEffect, useMemo, useRef, useState } from 'react';

import { Checkbox } from '@/components/ui/checkbox';
import { cn } from '@/lib/utils';

import type { FileManagerAction, FileManagerProps } from './file-manager-types';

import { FileManagerBulkActionsBar } from './file-manager-bulk-actions-bar';
import { FileManagerSkeleton } from './file-manager-skeleton';
import { FileManagerTreeNode } from './file-manager-tree-node';
import {
    buildFileManagerGridTemplate,
    buildFileManagerTree,
    collectAllNodePaths,
    collectVisibleFlat,
    filterFileManagerTree,
    findNodeByPath,
    getCheckboxState,
    normalizeRootGroups,
} from './file-manager-utils';
import { useFileManagerDnd } from './use-file-manager-dnd';
import { useFileManagerExpansion } from './use-file-manager-expansion';
import { useFileManagerKeyboardNavigation } from './use-file-manager-keyboard';
import { useFileManagerSelection } from './use-file-manager-selection';

const EMPTY_ACTIONS: readonly FileManagerAction[] = Object.freeze([]);

export const FileManager = ({
    actions,
    className,
    columns,
    emptyState,
    enableSelection,
    files,
    isLoading,
    labels,
    onBulkDelete,
    onMoveItems,
    onSelectionChange,
    rootGroups,
    search,
}: FileManagerProps) => {
    const isSizeVisible = columns?.isSizeVisible ?? true;
    const isModifiedVisible = columns?.isModifiedVisible ?? true;

    const normalizedRootGroups = useMemo(() => normalizeRootGroups(rootGroups), [rootGroups]);

    const fullTree = useMemo(() => buildFileManagerTree(files, normalizedRootGroups), [files, normalizedRootGroups]);

    const trimmedSearch = search?.query?.trim() ?? '';
    const isFiltering = trimmedSearch.length > 0;

    const visibleTree = useMemo(
        () => (isFiltering ? filterFileManagerTree(fullTree, trimmedSearch) : fullTree),
        [fullTree, isFiltering, trimmedSearch],
    );

    const allSelectablePaths = useMemo(() => collectAllNodePaths(visibleTree), [visibleTree]);

    const { expandedPaths, toggleExpand } = useFileManagerExpansion({
        isFiltering,
        normalizedRootGroups,
        visibleTree,
    });

    const flatVisible = useMemo(() => collectVisibleFlat(visibleTree, expandedPaths), [visibleTree, expandedPaths]);

    const {
        clearSelection,
        isAllSelected,
        isSomeSelected,
        onRowClick,
        onToggleCheckbox,
        selectedPaths,
        setSelection,
        toggleSelectAll,
    } = useFileManagerSelection({ allSelectablePaths, flatVisible });

    const isCheckboxVisible = enableSelection ?? !!onBulkDelete;
    const hasActions = !!actions?.length;

    // Report selection changes upstream without forcing parents to memoize the
    // callback — stash it in a ref so the effect only re-fires when the actual
    // selection changes (not when a fresh function instance is passed in).
    const onSelectionChangeRef = useRef(onSelectionChange);

    useEffect(() => {
        onSelectionChangeRef.current = onSelectionChange;
    }, [onSelectionChange]);

    useEffect(() => {
        onSelectionChangeRef.current?.(selectedPaths);
    }, [selectedPaths]);
    const gridTemplate = useMemo(
        () => buildFileManagerGridTemplate(isSizeVisible, isModifiedVisible, hasActions),
        [hasActions, isModifiedVisible, isSizeVisible],
    );

    // ── roving tabindex / keyboard navigation ────────────────────────────────
    const [activeRowPath, setActiveRowPath] = useState<null | string>(null);
    const containerRef = useRef<HTMLDivElement | null>(null);

    const resolvedActiveRow = useMemo(() => {
        if (activeRowPath && flatVisible.includes(activeRowPath)) {
            return activeRowPath;
        }

        return flatVisible[0] ?? null;
    }, [activeRowPath, flatVisible]);

    const focusRow = useCallback((path: null | string) => {
        if (!path) {
            return;
        }

        const next = containerRef.current?.querySelector<HTMLElement>(
            `[role="treeitem"][data-path="${CSS.escape(path)}"]`,
        );

        next?.focus();
    }, []);

    const handleRowFocus = useCallback((path: string) => {
        setActiveRowPath(path);
    }, []);

    const handleKeyDown = useFileManagerKeyboardNavigation({
        expandedPaths,
        flatVisible,
        focusRow,
        isCheckboxVisible,
        onClearSelection: clearSelection,
        onSelectAll: toggleSelectAll,
        onSetActiveRow: setActiveRowPath,
        onToggleCheckbox,
        onToggleExpand: toggleExpand,
        resolvedActiveRow,
        visibleTree,
    });

    const dnd = useFileManagerDnd({
        findNode: useCallback((path: string) => findNodeByPath(fullTree, path), [fullTree]),
        onClearSelection: clearSelection,
        onMoveItems,
        selectedPaths,
    });

    const handleContainerClick = useCallback(
        (event: ReactMouseEvent<HTMLDivElement>) => {
            if (event.target === event.currentTarget) {
                clearSelection();
            }
        },
        [clearSelection],
    );

    if (isLoading) {
        return (
            <div className={className}>
                <FileManagerSkeleton
                    columns={columns}
                    hasActions={hasActions}
                    isCheckboxVisible={isCheckboxVisible}
                />
            </div>
        );
    }

    if (files.length === 0) {
        return <div className={className}>{emptyState}</div>;
    }

    if (isFiltering && visibleTree.length === 0) {
        return <div className={className}>{search?.emptyState ?? emptyState}</div>;
    }

    const effectiveActions = actions ?? EMPTY_ACTIONS;
    const effectiveLabels = labels ?? {};
    const formatModified = effectiveLabels.formatModified;

    return (
        <div
            className={cn('bg-card flex flex-col overflow-hidden rounded-lg border', className)}
            onClick={handleContainerClick}
            onKeyDown={handleKeyDown}
            ref={containerRef}
        >
            <div
                className="bg-muted/30 text-muted-foreground grid items-center gap-3 border-b px-3 py-2 text-xs font-medium"
                style={{ gridTemplateColumns: gridTemplate }}
            >
                {isCheckboxVisible ? (
                    <Checkbox
                        aria-label={effectiveLabels.selectAllAriaLabel ?? 'Select all'}
                        checked={getCheckboxState(isAllSelected, isSomeSelected)}
                        onCheckedChange={toggleSelectAll}
                    />
                ) : (
                    <span
                        aria-hidden="true"
                        className="size-4"
                    />
                )}
                <span aria-hidden="true">{effectiveLabels.columnName ?? 'Name'}</span>
                {isSizeVisible && <span aria-hidden="true">{effectiveLabels.columnSize ?? 'Size'}</span>}
                {isModifiedVisible && <span aria-hidden="true">{effectiveLabels.columnModified ?? 'Modified'}</span>}
                {hasActions && (
                    <span
                        aria-hidden="true"
                        className="size-7"
                    />
                )}
            </div>

            <div
                aria-label="File tree"
                aria-multiselectable={isCheckboxVisible || undefined}
                className={cn(
                    'flex flex-1 flex-col overflow-y-auto py-1 transition-colors',
                    // Highlight the whole tree only when the cursor is actually hovering
                    // the empty area outside any row — that's the only place a "drop to
                    // root" will be accepted. `border-radius: inherit` makes the inset
                    // ring follow the outer container's rounded corners (top corners are
                    // hidden behind the header, so only the bottom is visually affected).
                    dnd.container.isRootDropTarget &&
                        'bg-primary/10 ring-primary ring-1 ring-inset [border-radius:inherit]',
                )}
                onDragEnter={dnd.isEnabled ? dnd.container.onDragEnter : undefined}
                onDragLeave={dnd.isEnabled ? dnd.container.onDragLeave : undefined}
                onDragOver={dnd.isEnabled ? dnd.container.onDragOver : undefined}
                onDrop={dnd.isEnabled ? dnd.container.onDrop : undefined}
                role="tree"
            >
                {visibleTree.map((node, index) => (
                    <FileManagerTreeNode
                        actions={effectiveActions}
                        activeRowPath={resolvedActiveRow}
                        bindNodeDnd={dnd.bindNodeDnd}
                        expandedPaths={expandedPaths}
                        formatModified={formatModified}
                        gridTemplate={gridTemplate}
                        hasActions={hasActions}
                        isCheckboxVisible={isCheckboxVisible}
                        isModifiedVisible={isModifiedVisible}
                        isSizeVisible={isSizeVisible}
                        key={node.id}
                        node={node}
                        onClick={onRowClick}
                        onFocusRow={handleRowFocus}
                        onToggleCheckbox={onToggleCheckbox}
                        onToggleExpand={toggleExpand}
                        posInSet={index + 1}
                        searchQuery={trimmedSearch || undefined}
                        selectedPaths={selectedPaths}
                        setSize={visibleTree.length}
                    />
                ))}
            </div>

            {onBulkDelete && (
                <FileManagerBulkActionsBar
                    files={files}
                    labels={effectiveLabels}
                    onBulkDelete={onBulkDelete}
                    onClearSelection={clearSelection}
                    onSelectionChange={setSelection}
                    selectedPaths={selectedPaths}
                />
            )}
        </div>
    );
};
