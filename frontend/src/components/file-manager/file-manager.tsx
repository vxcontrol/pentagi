import { Trash2 } from 'lucide-react';
import {
    Fragment,
    type KeyboardEvent as ReactKeyboardEvent,
    type ReactNode,
    useCallback,
    useMemo,
    useRef,
    useState,
} from 'react';

import ConfirmationDialog from '@/components/shared/confirmation-dialog';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import { cn } from '@/lib/utils';

import type { FileManagerAction, FileManagerInternalNode, FileManagerProps } from './file-manager-types';

import { FileManagerRow } from './file-manager-row';
import { FileManagerSkeleton } from './file-manager-skeleton';
import {
    buildFileManagerGridTemplate,
    buildFileManagerTree,
    collectAllFilePaths,
    collectVisibleFlat,
    dedupeOverlappingPaths,
    filterFileManagerTree,
    findNodeByPath,
    normalizeRootGroups,
} from './file-manager-utils';
import { useFileManagerExpansion } from './use-file-manager-expansion';
import { useFileManagerSelection } from './use-file-manager-selection';

const EMPTY_ACTIONS: readonly FileManagerAction[] = Object.freeze([]);

const pluralize = (count: number, single: string, plural: string): string =>
    `${count} ${count === 1 ? single : plural}`;

export const FileManager = ({
    actions,
    className,
    emptyState,
    files,
    isLoading,
    onBulkDelete,
    rootGroups,
    searchEmptyState,
    searchQuery,
    showModifiedColumn = true,
    showSizeColumn = true,
}: FileManagerProps) => {
    const normalizedRootGroups = useMemo(() => normalizeRootGroups(rootGroups), [rootGroups]);

    const fullTree = useMemo(
        () => buildFileManagerTree(files, normalizedRootGroups),
        [files, normalizedRootGroups],
    );

    const trimmedSearch = searchQuery?.trim() ?? '';
    const isFiltering = trimmedSearch.length > 0;

    const visibleTree = useMemo(
        () => (isFiltering ? filterFileManagerTree(fullTree, trimmedSearch) : fullTree),
        [fullTree, isFiltering, trimmedSearch],
    );

    const allFilePaths = useMemo(() => collectAllFilePaths(visibleTree), [visibleTree]);

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
    } = useFileManagerSelection({ allFilePaths, flatVisible });

    const showCheckbox = !!onBulkDelete;
    const hasActions = !!actions?.length;
    const gridTemplate = useMemo(
        () => buildFileManagerGridTemplate(showSizeColumn, showModifiedColumn, hasActions),
        [hasActions, showModifiedColumn, showSizeColumn],
    );

    const [bulkConfirmOpen, setBulkConfirmOpen] = useState(false);

    // ── roving tabindex / keyboard navigation ────────────────────────────────
    const [activeRowPath, setActiveRowPath] = useState<null | string>(null);
    const containerRef = useRef<HTMLDivElement | null>(null);

    const effectiveActiveRow = useMemo(() => {
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

    const handleKeyDown = useCallback(
        (event: ReactKeyboardEvent<HTMLDivElement>) => {
            if (event.key === 'Escape') {
                clearSelection();

                return;
            }

            if (!effectiveActiveRow || flatVisible.length === 0) {
                return;
            }

            const idx = flatVisible.indexOf(effectiveActiveRow);

            const moveTo = (nextIdx: number) => {
                const nextPath = flatVisible[Math.max(0, Math.min(nextIdx, flatVisible.length - 1))] ?? null;

                setActiveRowPath(nextPath);
                focusRow(nextPath);
            };

            switch (event.key) {
                case ' ':
                case 'Spacebar':
                    event.preventDefault();

                    if (showCheckbox) {
                        onToggleCheckbox(effectiveActiveRow);
                    }

                    return;
                case 'A':
                case 'a':
                    if (event.ctrlKey || event.metaKey) {
                        event.preventDefault();
                        toggleSelectAll();
                    }

                    return;
                case 'ArrowDown':
                    event.preventDefault();
                    moveTo(idx + 1);

                    return;
                case 'ArrowLeft':
                case 'ArrowRight': {
                    const node = findNodeByPath(visibleTree, effectiveActiveRow);

                    if (!node?.isDir) {
                        return;
                    }

                    const isExpanded = expandedPaths.has(node.path);
                    const wantsExpand = event.key === 'ArrowRight';

                    if (wantsExpand !== isExpanded) {
                        event.preventDefault();
                        toggleExpand(node.path, isExpanded);
                    }

                    return;
                }

                case 'ArrowUp':
                    event.preventDefault();
                    moveTo(idx - 1);

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
            clearSelection,
            effectiveActiveRow,
            expandedPaths,
            flatVisible,
            focusRow,
            onToggleCheckbox,
            showCheckbox,
            toggleExpand,
            toggleSelectAll,
            visibleTree,
        ],
    );

    const handleBulkDeleteConfirm = useCallback(async () => {
        if (!onBulkDelete) {
            return;
        }

        const dedupedPaths = new Set(dedupeOverlappingPaths(selectedPaths));
        const filesToDelete = files.filter((file) => dedupedPaths.has(file.path));

        await onBulkDelete(filesToDelete);
        setSelection(new Set());
    }, [files, onBulkDelete, selectedPaths, setSelection]);

    if (isLoading) {
        return (
            <div className={className}>
                <FileManagerSkeleton />
            </div>
        );
    }

    if (files.length === 0) {
        return <div className={className}>{emptyState}</div>;
    }

    if (isFiltering && visibleTree.length === 0) {
        return <div className={className}>{searchEmptyState ?? emptyState}</div>;
    }

    const selectedCountLabel = pluralize(selectedPaths.size, 'item', 'items');

    const renderNode = (node: FileManagerInternalNode): ReactNode => {
        const isExpanded = expandedPaths.has(node.path);
        const isSelected = selectedPaths.has(node.path);

        return (
            <Fragment key={node.id}>
                <FileManagerRow
                    actions={actions ?? (EMPTY_ACTIONS as FileManagerAction[])}
                    activeRowPath={effectiveActiveRow}
                    depth={node.depth}
                    file={node}
                    gridTemplate={gridTemplate}
                    hasActions={hasActions}
                    isExpanded={isExpanded}
                    isSelected={isSelected}
                    onClick={onRowClick}
                    onFocusRow={handleRowFocus}
                    onToggleCheckbox={onToggleCheckbox}
                    onToggleExpand={toggleExpand}
                    searchQuery={trimmedSearch || undefined}
                    showCheckbox={showCheckbox && !node.isGroupRoot}
                    showModified={showModifiedColumn}
                    showSize={showSizeColumn}
                />
                {node.isDir && isExpanded && node.children.length > 0 && node.children.map(renderNode)}
            </Fragment>
        );
    };

    return (
        <div
            className={cn('bg-card flex flex-col overflow-hidden rounded-lg border', className)}
            onClick={(event) => {
                if (event.target === event.currentTarget) {
                    clearSelection();
                }
            }}
            onKeyDown={handleKeyDown}
            ref={containerRef}
        >
            <div
                aria-hidden="true"
                className="bg-muted/30 text-muted-foreground grid items-center gap-3 border-b px-3 py-2 text-xs font-medium"
                style={{ gridTemplateColumns: gridTemplate }}
            >
                {showCheckbox ? (
                    <Checkbox
                        aria-label="Select all"
                        checked={isAllSelected ? true : isSomeSelected ? 'indeterminate' : false}
                        onCheckedChange={toggleSelectAll}
                    />
                ) : (
                    <span className="size-4" />
                )}
                <span>Name</span>
                {showSizeColumn && <span>Size</span>}
                {showModifiedColumn && <span>Modified</span>}
                {hasActions && <span className="size-7" />}
            </div>

            <div
                aria-label="File tree"
                className="flex flex-1 flex-col overflow-y-auto py-1"
                role="tree"
            >
                {visibleTree.map(renderNode)}
            </div>

            {selectedPaths.size > 0 && (
                <div className="bg-background flex items-center justify-between gap-2 border-t px-3 py-2">
                    <span className="text-muted-foreground text-sm">{selectedPaths.size} selected</span>
                    <div className="flex items-center gap-2">
                        <Button
                            onClick={clearSelection}
                            size="sm"
                            variant="ghost"
                        >
                            Cancel
                        </Button>
                        {onBulkDelete && (
                            <Button
                                onClick={() => setBulkConfirmOpen(true)}
                                size="sm"
                                variant="destructive"
                            >
                                <Trash2 className="size-4" />
                                Delete
                            </Button>
                        )}
                    </div>
                </div>
            )}

            <ConfirmationDialog
                confirmText="Delete"
                description={`This will delete ${selectedCountLabel}. This action cannot be undone.`}
                handleConfirm={handleBulkDeleteConfirm}
                handleOpenChange={(isOpen) => {
                    if (!isOpen) {
                        setBulkConfirmOpen(false);
                    }
                }}
                isOpen={bulkConfirmOpen}
                title={`Delete ${selectedCountLabel}`}
            />
        </div>
    );
};
