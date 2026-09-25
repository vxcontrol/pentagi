import { Ellipsis, X } from 'lucide-react';
import { type ComponentType, useCallback, useMemo, useState } from 'react';

import ConfirmationDialog from '@/components/shared/confirmation-dialog';
import { Button } from '@/components/ui/button';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { cn } from '@/lib/utils';

import type { FileManagerBulkAction, FileManagerLabels, FileNode } from './file-manager-types';

import { formatFileSize, pluralizeItemsEnglish } from './file-manager-utils';

interface BulkActionButtonProps {
    action: FileManagerBulkAction;
    isDisabled: boolean;
    onClick: (action: FileManagerBulkAction) => void;
}

interface FileManagerBulkActionsBarProps {
    /** Paths the action will carry, already resolved via `resolveActionPaths`. */
    actionPaths: readonly string[];
    actions: readonly FileManagerBulkAction[];
    /** Rows those paths cover, descendants included — what the bar and the confirmation report. */
    affectedCount: number;
    files: FileNode[];
    labels: FileManagerLabels;
    onClearSelection: () => void;
    /** Ticked rows; drives visibility only. */
    selectedCount: number;
    /** Cumulative byte count of the resolved selection. `0` suppresses the size suffix. */
    selectionTotalBytes: number;
}

interface ResolvedAction {
    action: FileManagerBulkAction;
    isDisabled: boolean;
}

/**
 * Footer that appears when at least one row is selected. Drives the cancel
 * control plus an arbitrary list of host-supplied bulk actions:
 *
 *   - Each action is rendered as a button (or pushed into the trailing `…`
 *     overflow menu when `overflow: true`).
 *   - Actions with a `confirm` config trigger the shared ConfirmationDialog
 *     before invoking `onSelect`. The dialog state is owned here so each host
 *     doesn't have to wire its own.
 *   - `actionPaths` arrives already resolved, so a directory never ships
 *     together with one of its descendants — the caller deletes / moves / etc.
 *     only the parent.
 */
export function FileManagerBulkActionsBar({
    actionPaths,
    actions,
    affectedCount,
    files,
    labels,
    onClearSelection,
    selectedCount,
    selectionTotalBytes,
}: FileManagerBulkActionsBarProps) {
    const [pendingAction, setPendingAction] = useState<FileManagerBulkAction | null>(null);

    const dedupedFiles = useMemo(() => {
        if (actionPaths.length === 0) {
            return [];
        }

        const actionPathSet = new Set(actionPaths);

        return files.filter((file) => actionPathSet.has(file.path));
    }, [actionPaths, files]);

    const visibleActions = useMemo<ResolvedAction[]>(
        () =>
            actions
                .filter((action) => !action.isHidden?.(dedupedFiles))
                .map((action) => ({
                    action,
                    isDisabled: dedupedFiles.length === 0 || (action.isDisabled?.(dedupedFiles) ?? false),
                })),
        [actions, dedupedFiles],
    );

    const inlineActions = useMemo(() => visibleActions.filter((entry) => !entry.action.overflow), [visibleActions]);
    const overflowActions = useMemo(() => visibleActions.filter((entry) => entry.action.overflow), [visibleActions]);

    const runAction = useCallback(
        async (action: FileManagerBulkAction) => {
            await action.onSelect(dedupedFiles);
        },
        [dedupedFiles],
    );

    const handleActionClick = useCallback(
        (action: FileManagerBulkAction) => {
            if (action.confirm) {
                setPendingAction(action);

                return;
            }

            void runAction(action);
        },
        [runAction],
    );

    const handleConfirm = useCallback(async () => {
        if (!pendingAction) {
            return;
        }

        await runAction(pendingAction);
        // ConfirmationDialog auto-closes after a fulfilled `handleConfirm`,
        // but we own `pendingAction` separately so the dialog tree fully
        // unmounts on the next tick.
        setPendingAction(null);
    }, [pendingAction, runAction]);

    const handleConfirmOpenChange = useCallback((isOpen: boolean) => {
        if (!isOpen) {
            setPendingAction(null);
        }
    }, []);

    if (selectedCount === 0 || actions.length === 0) {
        return null;
    }

    const pluralize = labels.pluralizeItems ?? pluralizeItemsEnglish;
    const countLabel = pluralize(affectedCount);
    const baseSelectedText = labels.selectedLabel?.(affectedCount) ?? `${affectedCount} selected`;
    const sizeSuffix = (labels.formatSelectionSize ?? formatFileSize)(selectionTotalBytes);
    const selectedText = sizeSuffix ? `${baseSelectedText} · ${sizeSuffix}` : baseSelectedText;
    const cancelText = labels.bulkCancel ?? 'Cancel';
    const moreActionsText = labels.bulkMoreActions ?? 'More actions';

    return (
        <>
            <div className="bg-background flex flex-wrap items-center gap-2 border-t px-3 py-2">
                <span className="text-muted-foreground text-sm">{selectedText}</span>
                <div className="ml-auto flex items-center gap-2">
                    <Button
                        aria-label={cancelText}
                        className="max-sm:size-8 max-sm:px-0"
                        onClick={onClearSelection}
                        size="sm"
                        variant="ghost"
                    >
                        <X className="sm:hidden" />
                        <span className="hidden sm:inline">{cancelText}</span>
                    </Button>

                    {inlineActions.map(({ action, isDisabled }) => (
                        <BulkActionButton
                            action={action}
                            isDisabled={isDisabled}
                            key={action.id}
                            onClick={handleActionClick}
                        />
                    ))}

                    {overflowActions.length > 0 && (
                        <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                                <Button
                                    aria-label={moreActionsText}
                                    size="icon-sm"
                                    variant="outline"
                                >
                                    <Ellipsis />
                                </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="end">
                                {overflowActions.map(({ action, isDisabled }) => (
                                    <DropdownMenuItem
                                        className={cn(
                                            action.variant === 'destructive' &&
                                                'text-destructive focus:text-destructive focus:bg-destructive/10',
                                        )}
                                        disabled={isDisabled}
                                        key={action.id}
                                        onSelect={(event) => {
                                            event.preventDefault();
                                            handleActionClick(action);
                                        }}
                                    >
                                        {action.icon ? <action.icon /> : null}
                                        {action.label}
                                    </DropdownMenuItem>
                                ))}
                            </DropdownMenuContent>
                        </DropdownMenu>
                    )}
                </div>
            </div>

            {pendingAction?.confirm && (
                <ConfirmationDialog
                    confirmText={pendingAction.confirm.confirmText ?? pendingAction.label}
                    confirmVariant={pendingAction.variant === 'destructive' ? 'destructive' : 'default'}
                    description={pendingAction.confirm.description?.(countLabel)}
                    handleConfirm={handleConfirm}
                    handleOpenChange={handleConfirmOpenChange}
                    isOpen={!!pendingAction}
                    title={pendingAction.confirm.title(countLabel)}
                />
            )}
        </>
    );
}

function BulkActionButton({ action, isDisabled, onClick }: BulkActionButtonProps) {
    const Icon = action.icon as ComponentType<{ className?: string }> | undefined;
    const button = (
        <Button
            className={cn(action.icon && 'max-sm:size-8 max-sm:px-0')}
            disabled={isDisabled}
            onClick={() => onClick(action)}
            size="sm"
            variant={action.variant === 'destructive' ? 'destructive' : 'outline'}
        >
            {Icon ? <Icon /> : null}
            <span className={cn(action.icon ? 'hidden sm:inline' : undefined)}>{action.label}</span>
        </Button>
    );

    if (!action.icon) {
        return button;
    }

    return (
        <Tooltip>
            <TooltipTrigger asChild>{button}</TooltipTrigger>
            <TooltipContent>{action.label}</TooltipContent>
        </Tooltip>
    );
}
