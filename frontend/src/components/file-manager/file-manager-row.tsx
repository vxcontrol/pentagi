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
import { formatFileSize, formatModified } from './file-manager-utils';

/** Marker on every interactive child of the row that should NOT bubble into a row click. */
const SKIP_ROW_CLICK_ATTR = 'data-fm-skip-row-click';

interface FileManagerRowProps {
    actions: FileManagerAction[];
    activeRowPath: null | string;
    depth: number;
    file: FileManagerInternalNode;
    gridTemplate: string;
    hasActions: boolean;
    isExpanded: boolean;
    isSelected: boolean;
    onClick: (event: ReactMouseEvent, path: string) => void;
    onFocusRow: (path: string) => void;
    onToggleCheckbox: (path: string) => void;
    onToggleExpand: (path: string, wasExpanded: boolean) => void;
    searchQuery?: string;
    showCheckbox: boolean;
    showModified: boolean;
    showSize: boolean;
}

/** Returns `true` when the click originated from an element opted-out of row activation. */
const isClickInsideSkipZone = (target: EventTarget | null): boolean =>
    target instanceof HTMLElement && !!target.closest(`[${SKIP_ROW_CLICK_ATTR}]`);

const buildVisibleActions = (actions: FileManagerAction[], file: FileManagerInternalNode): FileManagerAction[] =>
    actions.filter((action) => action.appliesToDirs || !file.isDir);

const FileManagerRowImpl = ({
    actions,
    activeRowPath,
    depth,
    file,
    gridTemplate,
    hasActions,
    isExpanded,
    isSelected,
    onClick,
    onFocusRow,
    onToggleCheckbox,
    onToggleExpand,
    searchQuery,
    showCheckbox,
    showModified,
    showSize,
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
            action.variant === 'destructive' &&
                'text-destructive focus:bg-destructive/10 focus:text-destructive',
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

    const renderActionItems = (kind: 'context' | 'dropdown'): ReactNode[] => {
        const Item = kind === 'context' ? ContextMenuItem : DropdownMenuItem;
        const Sep = kind === 'context' ? ContextMenuSeparator : DropdownMenuSeparator;
        const items: ReactNode[] = [];

        for (const action of visibleActions) {
            if (action.separatorBefore && items.length > 0) {
                items.push(<Sep key={`sep-${action.id}`} />);
            }

            items.push(renderActionItem(Item, action));
        }

        return items;
    };

    const dropdownItems = hasActions ? renderActionItems('dropdown') : [];
    const contextItems = renderActionItems('context');
    const isActiveRow = activeRowPath === file.path;

    const rowStyle = {
        '--fm-depth': depth,
        gridTemplateColumns: gridTemplate,
    } as CSSProperties & Record<'--fm-depth', number>;

    const row = (
        <div
            aria-expanded={file.isDir ? isExpanded : undefined}
            aria-selected={isSelected}
            className={cn(
                'group hover:bg-muted/50 grid cursor-pointer items-center gap-3 px-3 py-1.5 transition-colors outline-none',
                'focus-visible:bg-muted/70 focus-visible:ring-1 focus-visible:ring-ring',
                isSelected && 'bg-muted',
            )}
            data-path={file.path}
            onClick={handleRowClick}
            onFocus={() => onFocusRow(file.path)}
            role="treeitem"
            style={rowStyle}
            tabIndex={isActiveRow ? 0 : -1}
        >
            {showCheckbox ? (
                <span
                    className="flex items-center"
                    {...{ [SKIP_ROW_CLICK_ATTR]: '' }}
                >
                    <Checkbox
                        aria-label={`Select ${file.name}`}
                        checked={isSelected}
                        onCheckedChange={() => onToggleCheckbox(file.path)}
                        onClick={(event) => event.stopPropagation()}
                    />
                </span>
            ) : (
                <span
                    aria-hidden="true"
                    className="size-4"
                />
            )}

            <div
                className="flex min-w-0 items-center gap-1.5 pl-[calc(var(--fm-depth)*16px)]"
            >
                {file.isDir ? (
                    <span
                        aria-hidden="true"
                        className="text-muted-foreground hover:bg-muted -mx-0.5 inline-flex size-4 shrink-0 items-center justify-center rounded transition-colors"
                        onClick={(event) => {
                            event.stopPropagation();
                            onToggleExpand(file.path, isExpanded);
                        }}
                        {...{ [SKIP_ROW_CLICK_ATTR]: '' }}
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

            {showSize && (
                <span className="text-muted-foreground/80 shrink-0 text-xs tabular-nums">
                    {!file.isDir ? formatFileSize(file.size) : ''}
                </span>
            )}

            {showModified && (
                <span className="text-muted-foreground/80 shrink-0 text-xs tabular-nums">
                    {formatModified(file.modifiedAt)}
                </span>
            )}

            {hasActions && (
                <span {...{ [SKIP_ROW_CLICK_ATTR]: '' }}>
                    {dropdownItems.length > 0 ? (
                        <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                                <Button
                                    aria-label="Row actions"
                                    className="opacity-0 group-hover:opacity-100 data-[state=open]:opacity-100"
                                    onClick={(event) => event.stopPropagation()}
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
