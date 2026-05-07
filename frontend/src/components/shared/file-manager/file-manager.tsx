import {
    type MouseEvent as ReactMouseEvent,
    type ReactNode,
    useCallback,
    useEffect,
    useMemo,
    useRef,
    useState,
} from 'react';

import { Checkbox } from '@/components/ui/checkbox';
import {
    ContextMenu,
    ContextMenuContent,
    ContextMenuItem,
    ContextMenuSeparator,
    ContextMenuTrigger,
} from '@/components/ui/context-menu';
import { cn } from '@/lib/utils';

import type { FileManagerRowDisplay, FileManagerRowHandlers } from './file-manager-row';
import type {
    FileManagerAction,
    FileManagerBulkAction,
    FileManagerEmptyAreaAction,
    FileManagerProps,
    FileNode,
} from './file-manager-types';

import { FileManagerBulkActionsBar } from './file-manager-bulk-actions-bar';
import { FileManagerSkeleton } from './file-manager-skeleton';
import { FileManagerTreeNode } from './file-manager-tree-node';
import {
    collectVisibleFlat,
    computeDirSelectionState,
    computeSelectionTotalBytes,
    dedupeOverlappingPaths,
    findNodeByPath,
    getCheckboxState,
} from './file-manager-utils';
import { useFileManagerData } from './use-file-manager-data';
import { useFileManagerDnd } from './use-file-manager-dnd';
import { useFileManagerExpansion } from './use-file-manager-expansion';
import { useFileManagerKeyboardNavigation } from './use-file-manager-keyboard';
import { useFileManagerSelection } from './use-file-manager-selection';

const EMPTY_ACTIONS: readonly FileManagerAction[] = Object.freeze([]);
const EMPTY_BULK_ACTIONS: readonly FileManagerBulkAction[] = Object.freeze([]);
const EMPTY_AREA_ACTIONS: readonly FileManagerEmptyAreaAction[] = Object.freeze([]);

/**
 * Renders the right-click context menu items for the empty area of the tree.
 * Kept as a plain helper so the same JSX can be reused inside the
 * `<ContextMenuContent>` regardless of where it lives in the render tree.
 */
const renderEmptyAreaItems = (items: readonly FileManagerEmptyAreaAction[]): ReactNode[] => {
    const nodes: ReactNode[] = [];

    for (const action of items) {
        if (action.separatorBefore && nodes.length > 0) {
            nodes.push(<ContextMenuSeparator key={`separator-${action.id}`} />);
        }

        const ActionIcon = action.icon;

        nodes.push(
            <ContextMenuItem
                className={cn(
                    action.variant === 'destructive' &&
                        'text-destructive focus:bg-destructive/10 focus:text-destructive',
                )}
                key={action.id}
                onSelect={() => action.onSelect()}
            >
                {ActionIcon ? <ActionIcon className="size-4" /> : null}
                {action.label}
            </ContextMenuItem>,
        );
    }

    return nodes;
};

export const FileManager = ({
    actions,
    bulkActions,
    className,
    columns,
    emptyAreaActions,
    emptyState,
    enableSelection,
    files,
    isLoading,
    labels,
    onActiveRowChange,
    onExternalFileDrop,
    onMoveItems,
    onOpen,
    onOpenDirectory,
    onSelectionChange,
    rootGroups,
    search,
}: FileManagerProps) => {
    const effectiveBulkActions = bulkActions ?? EMPTY_BULK_ACTIONS;
    const hasBulkActions = effectiveBulkActions.length > 0;
    const isCheckboxVisible = enableSelection ?? hasBulkActions;
    const hasActions = !!actions?.length;

    const {
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
    } = useFileManagerData({
        columns,
        files,
        hasActions,
        rootGroups,
        searchQuery: search?.query,
    });

    const { expandedPaths, toggleExpand } = useFileManagerExpansion({
        isFiltering,
        normalizedRootGroups,
        visibleTree,
    });

    // Owned by the host (not the data hook) because it depends on `expandedPaths`,
    // which in turn depends on `visibleTree` from the data hook — moving it inside
    // would form a circular hook dependency.
    const flatVisible = useMemo(() => collectVisibleFlat(visibleTree, expandedPaths), [visibleTree, expandedPaths]);

    const {
        clearSelection,
        isAllSelected,
        isSomeSelected,
        onRowClick,
        onToggleSelection,
        selectedPaths,
        toggleSelectAll,
    } = useFileManagerSelection({ allSelectablePaths, dirSubtreePaths, flatVisible });

    // Cumulative byte total of the deduped selection — fed into the bulk bar's
    // size suffix ("3 selected · 14.2 MB"). Recomputed on selection / tree
    // changes; cheap because the dedup keeps the visit list short and each
    // subtree walk uses the same `findNodeByPath` traversal as the bar's other
    // helpers. Skipped entirely when the bar is hidden so a no-checkbox tree
    // never pays the cost.
    const selectionTotalBytes = useMemo(() => {
        if (!hasBulkActions || selectedPaths.size === 0) {
            return 0;
        }

        return computeSelectionTotalBytes(fullTree, dedupeOverlappingPaths(selectedPaths));
    }, [fullTree, hasBulkActions, selectedPaths]);

    // Tri-state checkbox values per directory: derived from `selectedPaths` so a
    // single state change updates every parent checkbox in lock-step. The map is
    // re-built whenever the selection or the tree shape changes; rows pull only
    // their own value out of it (see `FileManagerTreeNode`) which keeps the
    // memoized `FileManagerRow` from re-rendering for unrelated paths.
    //
    // Counting logic lives in `computeDirSelectionState` — see its JSDoc for
    // why the directory's own path is excluded from the count.
    const dirSelectionStates = useMemo(() => {
        const map = new Map<string, 'indeterminate' | boolean>();

        for (const [path, paths] of dirSubtreePaths) {
            map.set(path, computeDirSelectionState({ path, paths, selectedPaths }));
        }

        return map;
    }, [dirSubtreePaths, selectedPaths]);

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

    // Same latest-ref pattern for `onOpen` / `onOpenDirectory`: both bleed into
    // the keyboard handler's deps and through `TreeNode` → `Row`, so a
    // non-memoized parent callback would otherwise invalidate the memo on every
    // row whenever the parent re-renders.
    const onOpenRef = useRef(onOpen);
    const onOpenDirectoryRef = useRef(onOpenDirectory);

    useEffect(() => {
        onOpenRef.current = onOpen;
    }, [onOpen]);

    useEffect(() => {
        onOpenDirectoryRef.current = onOpenDirectory;
    }, [onOpenDirectory]);

    const handleOpen = useCallback((file: FileNode) => {
        onOpenRef.current?.(file);
    }, []);

    // We need the row + keyboard handler to know whether the consumer registered
    // a custom directory-open callback so the default expand/collapse is bypassed
    // only when an override actually exists. Wrapping in a stable callback keeps
    // the row memo intact across parent re-renders, but we conditionally pass it
    // (vs. an always-defined wrapper) by gating on the prop reference itself.
    const handleOpenDirectory = useCallback((dir: FileNode) => {
        onOpenDirectoryRef.current?.(dir);
    }, []);

    const stableHandleOpenDirectory = onOpenDirectory ? handleOpenDirectory : undefined;

    // ── roving tabindex / keyboard navigation ────────────────────────────────
    const [activeRowPath, setActiveRowPath] = useState<null | string>(null);
    const containerRef = useRef<HTMLDivElement | null>(null);

    const resolvedActiveRow = useMemo(() => {
        if (activeRowPath && flatVisible.includes(activeRowPath)) {
            return activeRowPath;
        }

        return flatVisible[0] ?? null;
    }, [activeRowPath, flatVisible]);

    // Mirror the `onSelectionChange` plumbing for the focused row: stash the
    // callback in a ref so the effect only re-fires on actual `activeRowPath`
    // changes, not on every parent re-render passing a fresh function. We
    // emit the raw `activeRowPath` (not `resolvedActiveRow`) so consumers can
    // distinguish "user picked something" from the auto-fallback to the first
    // visible row that the roving tabindex uses internally.
    const onActiveRowChangeRef = useRef(onActiveRowChange);

    useEffect(() => {
        onActiveRowChangeRef.current = onActiveRowChange;
    }, [onActiveRowChange]);

    useEffect(() => {
        onActiveRowChangeRef.current?.(activeRowPath);
    }, [activeRowPath]);

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
        onOpen: handleOpen,
        onOpenDirectory: stableHandleOpenDirectory,
        onSelectAll: toggleSelectAll,
        onSetActiveRow: setActiveRowPath,
        onToggleExpand: toggleExpand,
        onToggleSelection,
        resolvedActiveRow,
        visibleTree,
    });

    const dnd = useFileManagerDnd({
        findNode: useCallback((path: string) => findNodeByPath(fullTree, path), [fullTree]),
        onClearSelection: clearSelection,
        onExternalFileDrop,
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

    // Pull labels resolution and search-query trim up to render-top so they can
    // feed into `display` below. Both must be computed unconditionally — the
    // memoization that follows is guarded by `Object.is` on each dep, so a
    // missing `labels` prop ({} on every render) only invalidates `formatModified`
    // when the inner reference actually changes.
    const effectiveLabels = labels ?? {};
    const formatModified = effectiveLabels.formatModified;
    const effectiveActions = actions ?? EMPTY_ACTIONS;
    const searchQuery = trimmedSearch || undefined;

    // Bundle all per-tree-shared layout / i18n props into a single object so the
    // memoized `FileManagerRow` only does one reference check (instead of seven
    // primitive comparisons) on every parent re-render.
    const display = useMemo<FileManagerRowDisplay>(
        () => ({
            formatModified,
            gridTemplate,
            hasActions,
            isCheckboxVisible,
            isModifiedVisible,
            isSizeVisible,
            searchQuery,
        }),
        [formatModified, gridTemplate, hasActions, isCheckboxVisible, isModifiedVisible, isSizeVisible, searchQuery],
    );

    // Same trick for callbacks. Every dep here is already stabilized through
    // a ref or empty deps inside its source hook, so `handlers` is constructed
    // exactly once per `FileManager` instance and never invalidates row memo.
    const handlers = useMemo<FileManagerRowHandlers>(
        () => ({
            onClick: onRowClick,
            onFocusRow: handleRowFocus,
            onOpen: handleOpen,
            onOpenDirectory: stableHandleOpenDirectory,
            onToggleExpand: toggleExpand,
            onToggleSelection,
        }),
        [handleOpen, handleRowFocus, onRowClick, onToggleSelection, stableHandleOpenDirectory, toggleExpand],
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

    const effectiveEmptyAreaActions = emptyAreaActions ?? EMPTY_AREA_ACTIONS;
    const hasEmptyAreaActions = effectiveEmptyAreaActions.length > 0;

    // The scrollable tree element. Wrapped in a Radix `<ContextMenu>` below
    // when the host registered empty-area items — right-clicks on rows still
    // open the row-level menu because rows stop the contextmenu event there
    // (see `file-manager-row.tsx`), so the outer trigger only fires for
    // clicks outside any row, which is the entire point.
    const treeBody = (
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
                    'bg-primary/10 ring-primary [border-radius:inherit] ring-1 ring-inset',
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
                    dirSelectionStates={dirSelectionStates}
                    dirSubtreePaths={dirSubtreePaths}
                    display={display}
                    expandedPaths={expandedPaths}
                    handlers={handlers}
                    key={node.id}
                    node={node}
                    posInSet={index + 1}
                    selectedPaths={selectedPaths}
                    setSize={visibleTree.length}
                />
            ))}
        </div>
    );

    const tree = hasEmptyAreaActions ? (
        <ContextMenu>
            <ContextMenuTrigger asChild>{treeBody}</ContextMenuTrigger>
            <ContextMenuContent>{renderEmptyAreaItems(effectiveEmptyAreaActions)}</ContextMenuContent>
        </ContextMenu>
    ) : (
        treeBody
    );

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

            {tree}

            {hasBulkActions && (
                <FileManagerBulkActionsBar
                    actions={effectiveBulkActions}
                    files={files}
                    labels={effectiveLabels}
                    onClearSelection={clearSelection}
                    selectedPaths={selectedPaths}
                    selectionTotalBytes={selectionTotalBytes}
                />
            )}
        </div>
    );
};
