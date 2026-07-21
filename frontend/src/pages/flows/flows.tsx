import type { ColumnDef } from '@tanstack/react-table';

import { useMutation } from '@apollo/client/react';
import { Ellipsis, Eye, GitFork, Loader2, Pause, Pencil, PencilLine, Plus, Star, Trash } from 'lucide-react';
import { CheckCircle2, XCircle } from 'lucide-react';
import { useCallback, useMemo, useRef, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { toast } from 'sonner';

import { FlowStatusIcon } from '@/components/icons/flow-status-icon';
import { ProviderIcon } from '@/components/icons/provider-icon';
import {
    AppHeader,
    AppHeaderAction,
    AppHeaderActions,
    AppHeaderContent,
    AppHeaderTitle,
} from '@/components/layouts/app/app-header';
import ConfirmationDialog from '@/components/shared/confirmation-dialog';
import { InlineEditInput } from '@/components/shared/inline-edit';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { ContextMenuItem, ContextMenuSeparator } from '@/components/ui/context-menu';
import { DataTable, DataTableColumnHeader } from '@/components/ui/data-table';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { StatusCard } from '@/components/ui/status-card';
import { Toggle } from '@/components/ui/toggle';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { RenameFlowDocument, ResultType, StatusType, type TerminalFragmentFragment } from '@/graphql/types';
import { useTableState } from '@/hooks/use-table-state';
import { routes } from '@/lib/routes';
import { mergeHrefWithSearchParams } from '@/lib/url-params';
import { formatDate } from '@/lib/utils/format';
import { useFavorites } from '@/providers/favorites-provider';
import { type Flow, useFlows } from '@/providers/flows-provider';

const statusConfig: Record<
    StatusType,
    { label: string; variant: 'default' | 'destructive' | 'outline' | 'secondary' }
> = {
    [StatusType.Created]: {
        label: 'Created',
        variant: 'outline',
    },
    [StatusType.Failed]: {
        label: 'Failed',
        variant: 'destructive',
    },
    [StatusType.Finished]: {
        label: 'Finished',
        variant: 'secondary',
    },
    [StatusType.Running]: {
        label: 'Running',
        variant: 'default',
    },
    [StatusType.Waiting]: {
        label: 'Waiting',
        variant: 'outline',
    },
};

function Flows() {
    const navigate = useNavigate();
    const location = useLocation();
    const { deleteFlow, finishFlow, flows, isLoading } = useFlows();
    const { isFavoriteFlow, toggleFavoriteFlow } = useFavorites();
    const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
    const [deletingFlow, setDeletingFlow] = useState<Flow | null>(null);
    const [finishingFlowIds, setFinishingFlowIds] = useState<Set<string>>(new Set());
    const [deletingFlowIds, setDeletingFlowIds] = useState<Set<string>>(new Set());
    const [editingFlowId, setEditingFlowId] = useState<null | string>(null);
    const editingInputRef = useRef<HTMLInputElement>(null);
    const [renameFlowMutation, { loading: isRenameLoading }] = useMutation(RenameFlowDocument);

    const { filter, pageIndex: currentPage, setFilter, setPage: handlePageChange } = useTableState();

    const handleFlowOpen = useCallback(
        (flowId: string) => {
            navigate(mergeHrefWithSearchParams(routes.flow(flowId), new URLSearchParams(location.search)));
        },
        [navigate, location.search],
    );

    const handleFlowDeleteDialogOpen = useCallback((flow: Flow) => {
        setDeletingFlow(flow);
        setIsDeleteDialogOpen(true);
    }, []);

    const handleFlowRenameStart = useCallback((flow: Flow) => {
        setEditingFlowId(flow.id);
    }, []);

    const handleFlowDelete = async () => {
        if (!deletingFlow) {
            return;
        }

        setDeletingFlowIds((previousIds) => new Set(previousIds).add(deletingFlow.id));

        try {
            const success = await deleteFlow(deletingFlow);

            if (success) {
                setDeletingFlow(null);
            }
        } finally {
            setDeletingFlowIds((previousIds) => {
                const newIds = new Set(previousIds);
                newIds.delete(deletingFlow.id);

                return newIds;
            });
        }
    };

    const handleFlowRenameSave = useCallback(async () => {
        const newTitle = editingInputRef.current?.value.trim();

        if (!editingFlowId || !newTitle) {
            return;
        }

        try {
            const { data } = await renameFlowMutation({
                variables: {
                    flowId: editingFlowId,
                    title: newTitle,
                },
            });

            if (data?.renameFlow === ResultType.Success) {
                toast.success('Flow renamed successfully');
                setEditingFlowId(null);
            }
        } catch (error) {
            const errorMessage = error instanceof Error ? error.message : 'Failed to rename flow';
            toast.error(errorMessage);
        }
    }, [editingFlowId, renameFlowMutation]);

    const handleFlowRenameCancel = useCallback(() => {
        setEditingFlowId(null);
    }, []);

    const handleFlowFinish = useCallback(
        async (flow: Flow) => {
            setFinishingFlowIds((previousIds) => new Set(previousIds).add(flow.id));

            try {
                await finishFlow(flow);
            } finally {
                setFinishingFlowIds((previousIds) => {
                    const newIds = new Set(previousIds);
                    newIds.delete(flow.id);

                    return newIds;
                });
            }
        },
        [finishFlow],
    );

    const columns: ColumnDef<Flow>[] = useMemo(
        () => [
            {
                accessorKey: 'id',
                cell: ({ row }) => <div className="font-mono text-sm">{row.getValue('id')}</div>,
                enableHiding: false,
                header: ({ column }) => (
                    <DataTableColumnHeader
                        column={column}
                        title="ID"
                    />
                ),
                maxSize: 80,
                meta: { searchable: true },
                minSize: 60,
                size: 70,
            },
            {
                accessorKey: 'title',
                cell: ({ row }) => {
                    const flow = row.original;
                    const isEditing = editingFlowId === flow.id;
                    const title = row.getValue('title') as string;

                    if (isEditing) {
                        return (
                            <div onClick={(e) => e.stopPropagation()}>
                                <InlineEditInput
                                    autoFocus
                                    busy={isRenameLoading}
                                    defaultValue={title}
                                    inputRef={editingInputRef}
                                    onCancel={handleFlowRenameCancel}
                                    onSave={handleFlowRenameSave}
                                    placeholder="Flow title"
                                />
                            </div>
                        );
                    }

                    return <div className="truncate font-medium">{title}</div>;
                },
                enableHiding: false,
                header: ({ column }) => (
                    <DataTableColumnHeader
                        column={column}
                        title="Title"
                    />
                ),
                meta: { searchable: true },
                minSize: 200,
                size: 300,
            },
            {
                accessorKey: 'status',
                cell: ({ row }) => {
                    const status = row.getValue('status') as StatusType;
                    const config = statusConfig[status];

                    return (
                        <Badge variant={config.variant}>
                            <FlowStatusIcon
                                className="size-3"
                                status={status}
                            />
                            {config.label}
                        </Badge>
                    );
                },
                header: ({ column }) => (
                    <DataTableColumnHeader
                        column={column}
                        title="Status"
                    />
                ),
                maxSize: 130,
                meta: { searchable: true },
                minSize: 80,
                size: 100,
            },
            {
                // accessorFn returns the provider name as a plain string so it
                // participates in the DataTable global filter (search input).
                // The cell renderer still reads the original provider object
                // directly through `row.original`, so the icon + label stay
                // intact.
                accessorFn: (row) => row.provider?.name ?? '',
                cell: ({ row }) => {
                    const flow = row.original;

                    return (
                        <div className="flex items-center gap-2">
                            <ProviderIcon
                                className="size-4"
                                provider={flow.provider}
                            />
                            <span className="text-sm">{flow.provider?.name || 'N/A'}</span>
                        </div>
                    );
                },
                header: ({ column }) => (
                    <DataTableColumnHeader
                        column={column}
                        title="Provider"
                    />
                ),
                id: 'provider',
                maxSize: 150,
                meta: { searchable: true },
                minSize: 80,
                size: 100,
                sortingFn: (rowA, rowB) => {
                    const nameA = rowA.original.provider?.name || '';
                    const nameB = rowB.original.provider?.name || '';

                    return nameA.localeCompare(nameB);
                },
            },
            {
                // accessorFn joins all terminal images into one string for the
                // global search; the cell still derives its presentation from
                // the original array on `row.original`, and sortingFn keeps
                // ordering by count (more intuitive than alphabetical).
                accessorFn: (row) => (row.terminals ?? []).map((t) => t.image).join(' '),
                cell: ({ row }) => {
                    const flow = row.original;
                    const terminals = flow.terminals || [];

                    if (terminals.length === 0) {
                        return <span className="text-muted-foreground text-sm">No terminals</span>;
                    }

                    const isAnyConnected = terminals.some((t: TerminalFragmentFragment) => t.connected);
                    const images = [...new Set(terminals.map((t: TerminalFragmentFragment) => t.image))];

                    return (
                        <Tooltip>
                            <TooltipTrigger asChild>
                                <div className="flex items-center gap-2 overflow-hidden">
                                    {isAnyConnected ? (
                                        <CheckCircle2 className="size-4 shrink-0 text-green-500" />
                                    ) : (
                                        <XCircle className="text-muted-foreground size-4 shrink-0" />
                                    )}
                                    <span className="truncate text-sm">{images.join(', ')}</span>
                                </div>
                            </TooltipTrigger>
                            <TooltipContent>
                                <div className="flex flex-col gap-1">
                                    {terminals.map((terminal: TerminalFragmentFragment) => (
                                        <div
                                            className="flex items-center gap-2"
                                            key={terminal.id}
                                        >
                                            <span className="text-xs">{terminal.image}</span>
                                            <span className="text-muted-foreground text-xs">
                                                ({terminal.connected ? 'connected' : 'disconnected'})
                                            </span>
                                        </div>
                                    ))}
                                </div>
                            </TooltipContent>
                        </Tooltip>
                    );
                },
                header: ({ column }) => (
                    <DataTableColumnHeader
                        column={column}
                        title="Terminals"
                    />
                ),
                id: 'terminals',
                maxSize: 220,
                meta: { searchable: true },
                minSize: 160,
                size: 180,
                sortingFn: (rowA, rowB) => {
                    const terminalsA = rowA.original.terminals || [];
                    const terminalsB = rowB.original.terminals || [];

                    return terminalsA.length - terminalsB.length;
                },
            },
            {
                accessorKey: 'createdAt',
                cell: ({ row }) => {
                    const dateString = row.getValue('createdAt') as string;

                    return <div className="text-sm">{formatDate(new Date(dateString))}</div>;
                },
                header: ({ column }) => (
                    <DataTableColumnHeader
                        column={column}
                        title="Created"
                    />
                ),
                maxSize: 140,
                meta: { columnMenuLabel: 'Created' },
                minSize: 100,
                size: 120,
                sortingFn: (rowA, rowB) => {
                    const dateA = new Date(rowA.getValue('createdAt') as string);
                    const dateB = new Date(rowB.getValue('createdAt') as string);

                    return dateA.getTime() - dateB.getTime();
                },
            },
            {
                accessorKey: 'updatedAt',
                cell: ({ row }) => {
                    const dateString = row.getValue('updatedAt') as string;

                    return <div className="text-sm">{formatDate(new Date(dateString))}</div>;
                },
                header: ({ column }) => (
                    <DataTableColumnHeader
                        column={column}
                        title="Updated"
                    />
                ),
                maxSize: 140,
                meta: { columnMenuLabel: 'Updated' },
                minSize: 100,
                size: 120,
                sortingFn: (rowA, rowB) => {
                    const dateA = new Date(rowA.getValue('updatedAt') as string);
                    const dateB = new Date(rowB.getValue('updatedAt') as string);

                    return dateA.getTime() - dateB.getTime();
                },
            },
            {
                cell: ({ row }) => {
                    const flow = row.original;
                    const isRunning = ![StatusType.Failed, StatusType.Finished].includes(flow.status);

                    return (
                        <div className="flex items-center justify-end gap-1 opacity-0 transition-opacity group-hover:opacity-100">
                            <Toggle
                                aria-label="Toggle favorite"
                                className="border-none data-[state=on]:bg-transparent data-[state=on]:*:[svg]:fill-yellow-500 data-[state=on]:*:[svg]:stroke-yellow-500"
                                onClick={async (event) => {
                                    event.stopPropagation();
                                    await toggleFavoriteFlow(flow.id);
                                }}
                                pressed={isFavoriteFlow(flow.id)}
                                size="sm"
                                variant="outline"
                            >
                                <Star className="size-4" />
                            </Toggle>
                            <DropdownMenu>
                                <DropdownMenuTrigger asChild>
                                    <Button
                                        aria-label="Open menu"
                                        className="size-8 p-0"
                                        onClick={(e) => e.stopPropagation()}
                                        variant="ghost"
                                    >
                                        <Ellipsis />
                                    </Button>
                                </DropdownMenuTrigger>
                                <DropdownMenuContent
                                    align="end"
                                    className="min-w-24"
                                    onClick={(e) => e.stopPropagation()}
                                >
                                    <DropdownMenuItem onClick={() => handleFlowOpen(flow.id)}>
                                        <Eye />
                                        View
                                    </DropdownMenuItem>
                                    <DropdownMenuItem onClick={() => handleFlowRenameStart(flow)}>
                                        <PencilLine className="size-3" />
                                        Rename
                                    </DropdownMenuItem>
                                    {isRunning && (
                                        <DropdownMenuItem
                                            disabled={finishingFlowIds.has(flow.id)}
                                            onClick={() => handleFlowFinish(flow)}
                                        >
                                            {finishingFlowIds.has(flow.id) ? (
                                                <>
                                                    <Loader2 className="animate-spin" />
                                                    Finishing...
                                                </>
                                            ) : (
                                                <>
                                                    <Pause />
                                                    Finish
                                                </>
                                            )}
                                        </DropdownMenuItem>
                                    )}
                                    <DropdownMenuSeparator />
                                    <DropdownMenuItem
                                        disabled={deletingFlowIds.has(flow.id)}
                                        onClick={() => handleFlowDeleteDialogOpen(flow)}
                                    >
                                        {deletingFlowIds.has(flow.id) ? (
                                            <>
                                                <Loader2 className="size-4 animate-spin" />
                                                Deleting...
                                            </>
                                        ) : (
                                            <>
                                                <Trash className="size-4" />
                                                Delete
                                            </>
                                        )}
                                    </DropdownMenuItem>
                                </DropdownMenuContent>
                            </DropdownMenu>
                        </div>
                    );
                },
                enableHiding: false,
                header: () => null,
                id: 'actions',
                maxSize: 100,
                meta: { preventRowClick: true },
                minSize: 90,
                size: 96,
            },
        ],
        [
            deletingFlowIds,
            editingFlowId,
            finishingFlowIds,
            handleFlowDeleteDialogOpen,
            handleFlowFinish,
            handleFlowOpen,
            handleFlowRenameCancel,
            handleFlowRenameSave,
            handleFlowRenameStart,
            isFavoriteFlow,
            isRenameLoading,
            toggleFavoriteFlow,
        ],
    );

    const renderRowContextMenu = useCallback(
        (flow: Flow) => {
            const isRunning = ![StatusType.Failed, StatusType.Finished].includes(flow.status);

            return (
                <>
                    <ContextMenuItem onClick={async () => toggleFavoriteFlow(flow.id)}>
                        <Star />
                        {isFavoriteFlow(flow.id) ? 'Remove from favorites' : 'Add to favorites'}
                    </ContextMenuItem>
                    <ContextMenuSeparator />
                    <ContextMenuItem onClick={() => handleFlowOpen(flow.id)}>
                        <Eye />
                        View
                    </ContextMenuItem>
                    <ContextMenuItem onClick={() => handleFlowRenameStart(flow)}>
                        <Pencil />
                        Rename
                    </ContextMenuItem>

                    {isRunning && (
                        <ContextMenuItem
                            disabled={finishingFlowIds.has(flow.id)}
                            onClick={() => handleFlowFinish(flow)}
                        >
                            <Pause />
                            {finishingFlowIds.has(flow.id) ? 'Finishing...' : 'Finish'}
                        </ContextMenuItem>
                    )}
                    <ContextMenuSeparator />
                    <ContextMenuItem
                        disabled={deletingFlowIds.has(flow.id)}
                        onClick={() => handleFlowDeleteDialogOpen(flow)}
                    >
                        <Trash />
                        {deletingFlowIds.has(flow.id) ? 'Deleting...' : 'Delete'}
                    </ContextMenuItem>
                </>
            );
        },
        [
            deletingFlowIds,
            finishingFlowIds,
            handleFlowDeleteDialogOpen,
            handleFlowFinish,
            handleFlowOpen,
            handleFlowRenameStart,
            isFavoriteFlow,
            toggleFavoriteFlow,
        ],
    );

    const handleRowClick = useCallback(
        (flow: Flow) => {
            if (editingFlowId !== flow.id) {
                handleFlowOpen(flow.id);
            }
        },
        [editingFlowId, handleFlowOpen],
    );

    const pageHeader = (
        <AppHeader>
            <AppHeaderContent>
                <AppHeaderTitle icon={<GitFork className="size-4 shrink-0" />}>Flows</AppHeaderTitle>
            </AppHeaderContent>
            <AppHeaderActions>
                <AppHeaderAction
                    icon={<Plus />}
                    label="New Flow"
                    onClick={() => navigate(routes.newFlow)}
                    variant="secondary"
                />
            </AppHeaderActions>
        </AppHeader>
    );

    if (isLoading) {
        return (
            <>
                {pageHeader}
                <div className="flex flex-col gap-4 p-4">
                    <StatusCard
                        description="Please wait while we fetch your conversation flows"
                        icon={<Loader2 className="text-muted-foreground size-16 animate-spin" />}
                        title="Loading flows..."
                    />
                </div>
            </>
        );
    }

    if (flows.length === 0) {
        return (
            <>
                {pageHeader}
                <div className="flex flex-col gap-4 p-4">
                    <StatusCard
                        action={
                            <Button
                                onClick={() => navigate(routes.newFlow)}
                                variant="secondary"
                            >
                                <Plus />
                                New Flow
                            </Button>
                        }
                        description="Get started by creating your first conversation flow"
                        icon={<GitFork className="text-muted-foreground size-8" />}
                        title="No flows found"
                    />
                </div>
            </>
        );
    }

    return (
        <>
            {pageHeader}
            <div className="flex flex-col gap-4 p-4 pt-0">
                <DataTable<Flow>
                    columns={columns}
                    data={flows}
                    empty={{ entityName: 'flows' }}
                    filterPlaceholder="Filter flows..."
                    filterValue={filter}
                    isVirtualized
                    onFilterChange={setFilter}
                    onPageChange={handlePageChange}
                    onRowClick={handleRowClick}
                    pageIndex={currentPage}
                    renderRowContextMenu={renderRowContextMenu}
                />

                <ConfirmationDialog
                    cancelText="Cancel"
                    confirmText="Delete"
                    handleConfirm={handleFlowDelete}
                    handleOpenChange={setIsDeleteDialogOpen}
                    isOpen={isDeleteDialogOpen}
                    itemName={deletingFlow?.title}
                    itemType="flow"
                />
            </div>
        </>
    );
}

export default Flows;
