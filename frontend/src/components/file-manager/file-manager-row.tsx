import { ChevronRight, MoreVertical } from 'lucide-react';
import { type CSSProperties, memo, type MouseEvent as ReactMouseEvent, type ReactNode, useMemo } from 'react';

import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import {
    ContextMenu,
    ContextMenuContent,
    ContextMenuItem,
    ContextMenuSeparator,
    ContextMenuTrigger,
} from '@/components/ui/context-menu';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { cn } from '@/lib/utils';

import type { FileManagerAction, FileManagerInternalNode, FileNode } from './file-manager-types';
import type { FileManagerNodeDndHandlers } from './use-file-manager-dnd';

import { FileManagerHighlightedName } from './file-manager-highlighted-name';
import { getFileTypeIcon } from './file-manager-icons';
import { formatModified as defaultFormatModified, formatFileSize } from './file-manager-utils';

/**
 * Marker on every interactive child of the row that should NOT bubble into a row
 * click or double-click. Detected via `closest()` in the row's handlers — descendants
 * don't need to call `event.stopPropagation()` themselves.
 */
const SKIP_ROW_CLICK_ATTR = 'data-fm-skip-row-click';
const skipRowClickProps = { [SKIP_ROW_CLICK_ATTR]: '' };

/**
 * Layout/visibility/i18n props that are identical for every row in the tree.
 * `FileManager` builds this object once with `useMemo` so memoized rows do not
 * have to compare seven separate primitives on every parent re-render — a single
 * reference check is enough.
 */
export interface FileManagerRowDisplay {
    formatModified?: (modifiedAt: Date | string | undefined) => string;
    gridTemplate: string;
    hasActions: boolean;
    isCheckboxVisible: boolean;
    isModifiedVisible: boolean;
    isSizeVisible: boolean;
    searchQuery?: string;
}

/**
 * Stable callback bundle shared by every row. All handlers are produced by
 * hooks that go through the latest-ref pattern, so this object is built once
 * and never invalidates the row memo.
 */
export interface FileManagerRowHandlers {
    onClick: (event: ReactMouseEvent, path: string, subtreePaths?: readonly string[]) => void;
    onFocusRow: (path: string) => void;
    onOpen?: (file: FileNode) => void;
    onToggleExpand: (path: string, wasExpanded: boolean) => void;
    /**
     * Polymorphic selection toggle: file rows pass just `path`, directory rows
     * pass the precomputed subtree so the whole branch flips in one gesture.
     */
    onToggleSelection: (path: string, subtreePaths?: readonly string[]) => void;
}

interface FileManagerRowProps {
    actions: readonly FileManagerAction[];
    activeRowPath: null | string;
    /**
     * Tri-state checkbox value for directory rows (`true`, `false`, `'indeterminate'`).
     * `undefined` for file rows — files fall back to `isSelected`.
     */
    dirCheckboxState?: 'indeterminate' | boolean;
    /**
     * Pre-computed list of every selectable path in the directory's subtree
     * (the directory itself plus all descendants). `undefined` for files.
     * Captured by the directory-checkbox click handler so a single gesture
     * flips the entire branch.
     */
    dirSubtreePaths?: readonly string[];
    /** Per-tree shared layout / i18n bundle (one stable reference). */
    display: FileManagerRowDisplay;
    /** Drag/drop handlers for this row. `null` when intra-tree DnD is disabled. */
    dnd: FileManagerNodeDndHandlers | null;
    file: FileManagerInternalNode;
    /** Per-tree shared callback bundle (one stable reference). */
    handlers: FileManagerRowHandlers;
    isExpanded: boolean;
    isSelected: boolean;
    /** 1-based position of the row inside its parent's child list (for `aria-posinset`). */
    posInSet: number;
    /** Total number of siblings the row is part of (for `aria-setsize`). */
    setSize: number;
}

/**
 * Returns `true` when the click originated from an element opted-out of row activation.
 *
 * `Element` (not `HTMLElement`) is the correct guard: `<svg>` and its children
 * (`<path>` etc.) are `SVGElement`s, which do NOT extend `HTMLElement` even
 * though they share the `Element.closest()` API. Using `HTMLElement` here would
 * make a click on the actual painted pixels of an icon (chevron, action button
 * icon, …) bypass the skip-marker check and re-trigger row selection.
 */
const isClickInsideSkipZone = (target: EventTarget | null): boolean =>
    target instanceof Element && !!target.closest(`[${SKIP_ROW_CLICK_ATTR}]`);

const buildVisibleActions = (
    actions: readonly FileManagerAction[],
    file: FileManagerInternalNode,
): FileManagerAction[] => actions.filter((action) => action.appliesToDirs || !file.isDir);

const FileManagerRowImpl = ({
    actions,
    activeRowPath,
    dirCheckboxState,
    dirSubtreePaths,
    display,
    dnd,
    file,
    handlers,
    isExpanded,
    isSelected,
    posInSet,
    setSize,
}: FileManagerRowProps) => {
    const {
        formatModified = defaultFormatModified,
        gridTemplate,
        hasActions,
        isCheckboxVisible,
        isModifiedVisible,
        isSizeVisible,
        searchQuery,
    } = display;
    const { onClick, onFocusRow, onOpen, onToggleExpand, onToggleSelection } = handlers;

    const { icon: Icon, tone } = useMemo(
        () =>
            file.groupIcon
                ? { icon: file.groupIcon, tone: 'text-blue-400' }
                : getFileTypeIcon({ isDir: file.isDir, isOpen: isExpanded, name: file.name }),
        [file.groupIcon, file.isDir, file.name, isExpanded],
    );

    const visibleActions = useMemo(() => buildVisibleActions(actions, file), [actions, file]);

    const handleRowClick = (event: ReactMouseEvent) => {
        if (isClickInsideSkipZone(event.target)) {
            return;
        }

        // Hand the precomputed subtree paths to the selection hook for directory
        // rows: a plain or `Cmd`/`Ctrl`+click on a folder then operates on the
        // entire branch — including descendants of a collapsed folder — instead
        // of just the folder's own path.
        onClick(event, file.path, file.isDir ? dirSubtreePaths : undefined);
    };

    // Double-click is the row's "open" gesture. For directories it expands or
    // collapses (decoupling expansion from the single click keeps `Shift`/`Cmd`+click
    // pure selection gestures and matches Finder/Explorer). For files it forwards
    // to `onOpen` — typically wired to download / preview / open-in-tab. The
    // chevron icon and arrow keys remain alternative ways to expand without selecting.
    const handleRowDoubleClick = (event: ReactMouseEvent) => {
        if (isClickInsideSkipZone(event.target)) {
            return;
        }

        if (file.isDir) {
            event.preventDefault();
            onToggleExpand(file.path, isExpanded);

            return;
        }

        if (onOpen) {
            event.preventDefault();
            onOpen(file);
        }
    };

    const renderActionItem = (
        Component: typeof ContextMenuItem | typeof DropdownMenuItem,
        action: FileManagerAction,
    ) => {
        const ActionIcon = action.icon;
        const itemClassName = cn(
            action.variant === 'destructive' && 'text-destructive focus:bg-destructive/10 focus:text-destructive',
        );

        if (action.getHref) {
            return (
                <Component
                    asChild
                    className={itemClassName}
                    key={action.id}
                >
                    <a
                        download={action.getHrefDownloadAttr?.(file) ?? true}
                        href={action.getHref(file)}
                    >
                        {ActionIcon ? <ActionIcon className="size-4" /> : null}
                        {action.label}
                    </a>
                </Component>
            );
        }

        return (
            <Component
                className={itemClassName}
                key={action.id}
                onSelect={() => action.onSelect(file)}
            >
                {ActionIcon ? <ActionIcon className="size-4" /> : null}
                {action.label}
            </Component>
        );
    };

    const renderActionItems = (menuKind: 'context' | 'dropdown'): ReactNode[] => {
        const MenuItem = menuKind === 'context' ? ContextMenuItem : DropdownMenuItem;
        const MenuSeparator = menuKind === 'context' ? ContextMenuSeparator : DropdownMenuSeparator;
        const items: ReactNode[] = [];

        for (const action of visibleActions) {
            if (action.separatorBefore && items.length > 0) {
                items.push(<MenuSeparator key={`separator-${action.id}`} />);
            }

            items.push(renderActionItem(MenuItem, action));
        }

        return items;
    };

    const dropdownItems = hasActions ? renderActionItems('dropdown') : [];
    const contextItems = renderActionItems('context');
    const isActiveRow = activeRowPath === file.path;

    const rowStyle = {
        '--fm-depth': file.depth,
        gridTemplateColumns: gridTemplate,
    } as CSSProperties & Record<'--fm-depth', number>;

    const isDraggable = !!dnd && !file.isGroupRoot;
    const isDropTarget = dnd?.isDropTarget ?? false;
    const isBeingDragged = dnd?.isBeingDragged ?? false;

    const row = (
        <div
            aria-expanded={file.isDir ? isExpanded : undefined}
            aria-level={file.depth + 1}
            aria-posinset={posInSet}
            aria-selected={isSelected}
            aria-setsize={setSize}
            className={cn(
                'group hover:bg-accent grid cursor-pointer items-center gap-3 px-3 py-1.5 transition-colors outline-none',
                'focus-visible:bg-muted/70 focus-visible:ring-ring focus-visible:ring-1',
                // `select-none` keeps double-click reserved for expand/collapse
                // — without it the browser would highlight the row's text on dblclick.
                'select-none',
                isSelected && 'bg-muted',
                isDropTarget && 'bg-primary/10 ring-primary/40 ring-1 ring-inset',
                // Ghost every row that's part of the in-flight drag (the grabbed row
                // plus any other selected rows being moved together) so the user sees
                // the entire batch on the move, not just the row whose drag image the
                // browser is rendering. Combine reduced opacity + dashed outline so
                // the effect is unmistakable even when the row is already selected
                // (`bg-muted` would otherwise mute the opacity contrast).
                isBeingDragged && 'border-muted-foreground/40 border border-dashed opacity-40',
                // Counter-act the 1px border above to keep the row from shifting layout
                // when the dashed border kicks in.
                !isBeingDragged && 'border border-transparent',
            )}
            data-path={file.path}
            draggable={isDraggable}
            onClick={handleRowClick}
            onDoubleClick={handleRowDoubleClick}
            onDragEnd={dnd?.onDragEnd}
            onDragEnter={dnd?.onDragEnter}
            onDragLeave={dnd?.onDragLeave}
            onDragOver={dnd?.onDragOver}
            onDragStart={dnd?.onDragStart}
            onDrop={dnd?.onDrop}
            onFocus={() => onFocusRow(file.path)}
            role="treeitem"
            style={rowStyle}
            tabIndex={isActiveRow ? 0 : -1}
        >
            {isCheckboxVisible ? (
                <span
                    className="flex items-center"
                    {...skipRowClickProps}
                >
                    <Checkbox
                        aria-label={`Select ${file.name}`}
                        // Directories surface a tri-state value derived from their
                        // descendants; files (and edge cases without a precomputed
                        // value) fall back to the row's own selection flag.
                        checked={file.isDir ? (dirCheckboxState ?? isSelected) : isSelected}
                        // For folders we hand the precomputed subtree to the
                        // selection hook so one gesture flips the entire branch
                        // (the directory itself + every descendant); files just
                        // toggle their own path.
                        onCheckedChange={() => onToggleSelection(file.path, file.isDir ? dirSubtreePaths : undefined)}
                    />
                </span>
            ) : (
                <span
                    aria-hidden="true"
                    className="size-4"
                />
            )}

            <div className="relative flex min-w-0 items-center gap-1.5 self-stretch pl-[calc(var(--fm-depth)*16px)]">
                {Array.from({ length: file.depth }, (_, i) => (
                    <span
                        aria-hidden="true"
                        className="bg-border pointer-events-none absolute -inset-y-1.75 w-px"
                        key={i}
                        style={{ left: `${i * 16 + 6}px` }}
                    />
                ))}
                {file.isDir ? (
                    <span
                        aria-hidden="true"
                        className="text-muted-foreground hover:bg-muted -mx-0.5 inline-flex size-4 shrink-0 items-center justify-center rounded transition-colors"
                        onClick={() => onToggleExpand(file.path, isExpanded)}
                        {...skipRowClickProps}
                    >
                        <ChevronRight className={cn('size-3.5 transition-transform', isExpanded && 'rotate-90')} />
                    </span>
                ) : (
                    <span
                        aria-hidden="true"
                        className="-mx-0.5 size-4 shrink-0"
                    />
                )}
                <Icon className={cn('size-4 shrink-0', tone)} />
                <FileManagerHighlightedName
                    className={cn('text-sm', file.isGroupRoot && 'font-semibold')}
                    name={file.name}
                    query={searchQuery}
                />
            </div>

            {isSizeVisible && (
                <span className="text-muted-foreground/80 shrink-0 text-xs tabular-nums">
                    {!file.isDir ? formatFileSize(file.size) : ''}
                </span>
            )}

            {isModifiedVisible && (
                <span className="text-muted-foreground/80 shrink-0 text-xs tabular-nums">
                    {formatModified(file.modifiedAt)}
                </span>
            )}

            {hasActions && (
                <span {...skipRowClickProps}>
                    {dropdownItems.length > 0 ? (
                        <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                                <Button
                                    aria-label="Row actions"
                                    className="opacity-0 group-hover:opacity-100 data-[state=open]:opacity-100"
                                    size="icon-xs"
                                    variant="ghost"
                                >
                                    <MoreVertical className="size-3.5" />
                                </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="end">{dropdownItems}</DropdownMenuContent>
                        </DropdownMenu>
                    ) : null}
                </span>
            )}
        </div>
    );

    if (contextItems.length === 0) {
        return row;
    }

    return (
        <ContextMenu>
            <ContextMenuTrigger asChild>{row}</ContextMenuTrigger>
            <ContextMenuContent>{contextItems}</ContextMenuContent>
        </ContextMenu>
    );
};

FileManagerRowImpl.displayName = 'FileManagerRow';

export const FileManagerRow = memo(FileManagerRowImpl);
