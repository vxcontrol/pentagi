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

import type { FileManagerAction, FileManagerInternalNode } from './file-manager-types';

import { FileManagerHighlightedName } from './file-manager-highlighted-name';
import { getFileTypeIcon } from './file-manager-icons';
import { formatModified as defaultFormatModified, formatFileSize } from './file-manager-utils';

/**
 * Marker on every interactive child of the row that should NOT bubble into a row click.
 * Detected via `closest()` in the row's click handler — descendants don't need to call
 * `event.stopPropagation()` themselves.
 */
const SKIP_ROW_CLICK_ATTR = 'data-fm-skip-row-click';
const skipRowClickProps = { [SKIP_ROW_CLICK_ATTR]: '' };

interface FileManagerRowProps {
    actions: readonly FileManagerAction[];
    activeRowPath: null | string;
    file: FileManagerInternalNode;
    formatModified?: (modifiedAt: Date | string | undefined) => string;
    gridTemplate: string;
    hasActions: boolean;
    isCheckboxVisible: boolean;
    isExpanded: boolean;
    isModifiedVisible: boolean;
    isSelected: boolean;
    isSizeVisible: boolean;
    onClick: (event: ReactMouseEvent, path: string) => void;
    onFocusRow: (path: string) => void;
    onToggleCheckbox: (path: string) => void;
    onToggleExpand: (path: string, wasExpanded: boolean) => void;
    /** 1-based position of the row inside its parent's child list (for `aria-posinset`). */
    posInSet: number;
    searchQuery?: string;
    /** Total number of siblings the row is part of (for `aria-setsize`). */
    setSize: number;
}

/** Returns `true` when the click originated from an element opted-out of row activation. */
const isClickInsideSkipZone = (target: EventTarget | null): boolean =>
    target instanceof HTMLElement && !!target.closest(`[${SKIP_ROW_CLICK_ATTR}]`);

const buildVisibleActions = (
    actions: readonly FileManagerAction[],
    file: FileManagerInternalNode,
): FileManagerAction[] => actions.filter((action) => action.appliesToDirs || !file.isDir);

const FileManagerRowImpl = ({
    actions,
    activeRowPath,
    file,
    formatModified = defaultFormatModified,
    gridTemplate,
    hasActions,
    isCheckboxVisible,
    isExpanded,
    isModifiedVisible,
    isSelected,
    isSizeVisible,
    onClick,
    onFocusRow,
    onToggleCheckbox,
    onToggleExpand,
    posInSet,
    searchQuery,
    setSize,
}: FileManagerRowProps) => {
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

        if (file.isDir) {
            onToggleExpand(file.path, isExpanded);
        }

        onClick(event, file.path);
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

    const row = (
        <div
            aria-expanded={file.isDir ? isExpanded : undefined}
            aria-level={file.depth + 1}
            aria-posinset={posInSet}
            aria-selected={isSelected}
            aria-setsize={setSize}
            className={cn(
                'group hover:bg-muted/50 grid cursor-pointer items-center gap-3 px-3 py-1.5 transition-colors outline-none',
                'focus-visible:bg-muted/70 focus-visible:ring-ring focus-visible:ring-1',
                isSelected && 'bg-muted',
            )}
            data-path={file.path}
            onClick={handleRowClick}
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
                        checked={isSelected}
                        onCheckedChange={() => onToggleCheckbox(file.path)}
                    />
                </span>
            ) : (
                <span
                    aria-hidden="true"
                    className="size-4"
                />
            )}

            <div className="flex min-w-0 items-center gap-1.5 pl-[calc(var(--fm-depth)*16px)]">
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
                        className="size-4 shrink-0"
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
