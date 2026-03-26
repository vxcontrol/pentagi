import type { ColumnDef } from '@tanstack/react-table';

import { format, isToday } from 'date-fns';
import { enUS } from 'date-fns/locale';
import {
    ArrowDown,
    ArrowUp,
    Check,
    CheckCircle2,
    Eye,
    FileText,
    GitFork,
    Loader2,
    MoreHorizontal,
    Pause,
    Pencil,
    Plus,
    Star,
    Trash,
    X,
    XCircle,
} from 'lucide-react';
import { useCallback, useMemo, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { toast } from 'sonner';

import { FlowStatusIcon } from '@/components/icons/flow-status-icon';
import { ProviderIcon } from '@/components/icons/provider-icon';
import ConfirmationDialog from '@/components/shared/confirmation-dialog';
import { Badge } from '@/components/ui/badge';
import { Breadcrumb, BreadcrumbItem, BreadcrumbList, BreadcrumbPage } from '@/components/ui/breadcrumb';
import { Button } from '@/components/ui/button';
import { DataTable } from '@/components/ui/data-table';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Input } from '@/components/ui/input';
import { Separator } from '@/components/ui/separator';
import { SidebarTrigger } from '@/components/ui/sidebar';
import { StatusCard } from '@/components/ui/status-card';
import { Toggle } from '@/components/ui/toggle';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { ResultType, StatusType, type TerminalFragmentFragment, useRenameFlowMutation } from '@/graphql/types';
import { useAdaptiveColumnVisibility } from '@/hooks/use-adaptive-column-visibility';
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

const formatDateTime = (dateString: string) => {
    const date = new Date(dateString);

    if (isToday(date)) {
        return format(date, 'HH:mm:ss', { locale: enUS });
    }

    return format(date, 'd MMM yyyy', { locale: enUS });
};

const formatFullDateTime = (dateString: string) => {
    const date = new Date(dateString);

    return format(date, 'd MMM yyyy, HH:mm:ss', { locale: enUS });
};

const Flows = () => {
    const navigate = useNavigate();
    const [searchParams, setSearchParams] = useSearchParams();
    const { deleteFlow, finishFlow, flows, isLoading } = useFlows();
    const { isFavoriteFlow, toggleFavoriteFlow } = useFavorites();
    const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
    const [deletingFlow, setDeletingFlow] = useState<Flow | null>(null);
    const [finishingFlowIds, setFinishingFlowIds] = useState<Set<string>>(new Set());
    const [deletingFlowIds, setDeletingFlowIds] = useState<Set<string>>(new Set());
    const [editingFlowId, setEditingFlowId] = useState<null | string>(null);
    const [editingFlowTitle, setEditingFlowTitle] = useState('');
    const [renameFlowMutation, { loading: isRenameLoading }] = useRenameFlowMutation();

    const { columnVisibility, updateColumnVisibility } = useAdaptiveColumnVisibility({
        columns: [
            { alwaysVisible: true, id: 'id', priority: 0 },
            { alwaysVisible: true, id: 'title', priority: 0 },
            { id: 'status', priority: 1 },
            { id: 'provider', priority: 2 },
            { id: 'createdAt', priority: 3 },
            { id: 'updatedAt', priority: 4 },
            { id: 'terminals', priority: 5 },
        ],
        tableKey: 'flows',
    });

    // Three-way sorting handler: null -> asc -> desc -> null
    const handleColumnSort = useMemo(
        () =>
            (column: {
                clearSorting: () => void;
                getIsSorted: () => 'asc' | 'desc' | false;
                toggleSorting: (desc?: boolean) => void;
            }) => {
                const sorted = column.getIsSorted();

                if (sorted === 'asc') {
                    column.toggleSorting(true);
                } else if (sorted === 'desc') {
                    column.clearSorting();
                } else {
                    column.toggleSorting(false);
                }
            },
        [],
    );

    // Get current page from URL
    const currentPage = useMemo(() => {
        const page = searchParams.get('page');

        return page ? Math.max(0, Number.parseInt(page, 10) - 1) : 0;
    }, [searchParams]);

    // Handle page change
    const handlePageChange = useCallback(
        (pageIndex: number) => {
            const newParams = new URLSearchParams(searchParams);

            if (pageIndex === 0) {
                newParams.delete('page');
            } else {
                newParams.set('page', String(pageIndex + 1));
            }

            setSearchParams(newParams);
        },
        [searchParams, setSearchParams],
    );

    const handleFlowOpen = useCallback(
        (flowId: string) => {
            navigate(`/flows/${flowId}`);
        },
        [navigate],
    );

    const handleFlowDeleteDialogOpen = useCallback((flow: Flow) => {
        setDeletingFlow(flow);
        setIsDeleteDialogOpen(true);
    }, []);

    const handleFlowRenameStart = useCallback((flow: Flow) => {
        setEditingFlowId(flow.id);
        setEditingFlowTitle(flow.title);
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
        if (!editingFlowId || !editingFlowTitle.trim()) {
            return;
        }

        try {
            const { data } = await renameFlowMutation({
                variables: {
                    flowId: editingFlowId,
                    title: editingFlowTitle.trim(),
                },
            });

            if (data?.renameFlow === ResultType.Success) {
                toast.success('Flow renamed successfully');
                setEditingFlowId(null);
                setEditingFlowTitle('');
            }
        } catch (error) {
            const errorMessage = error instanceof Error ? error.message : 'Failed to rename flow';
            toast.error(errorMessage);
        }
    }, [editingFlowId, editingFlowTitle, renameFlowMutation]);

    const handleFlowRenameCancel = useCallback(() => {
        setEditingFlowId(null);
        setEditingFlowTitle('');
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
                header: ({ column }) => {
                    const sorted = column.getIsSorted();

                    return (
                        <Button
                            className="text-muted-foreground hover:text-primary flex items-center gap-2 p-0 no-underline hover:no-underline"
                            onClick={() => handleColumnSort(column)}
                            variant="link"
                        >
                            ID
                            {sorted === 'asc' ? (
                                <ArrowDown className="size-4" />
                            ) : sorted === 'desc' ? (
                                <ArrowUp className="size-4" />
                            ) : null}
                        </Button>
                    );
                },
                maxSize: 80,
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
                            <Input
                                autoFocus
                                className="h-8"
                                onChange={(e) => setEditingFlowTitle(e.target.value)}
                                onClick={(e) => e.stopPropagation()}
                                onKeyDown={(e) => {
                                    if (e.key === 'Enter') {
                                        handleFlowRenameSave();
                                    } else if (e.key === 'Escape') {
                                        handleFlowRenameCancel();
                                    }
                                }}
                                placeholder="Flow title"
                                value={editingFlowTitle}
                            />
                        );
                    }

                    return <div className="truncate font-medium">{title}</div>;
                },
                enableHiding: false,
                header: ({ column }) => {
                    const sorted = column.getIsSorted();

                    return (
                        <Button
                            className="text-muted-foreground hover:text-primary flex items-center gap-2 p-0 no-underline hover:no-underline"
                            onClick={() => handleColumnSort(column)}
                            variant="link"
                        >
                            Title
                            {sorted === 'asc' ? (
                                <ArrowDown className="size-4" />
                            ) : sorted === 'desc' ? (
                                <ArrowUp className="size-4" />
                            ) : null}
                        </Button>
                    );
                },
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
                header: ({ column }) => {
                    const sorted = column.getIsSorted();

                    return (
                        <Button
                            className="text-muted-foreground hover:text-primary flex items-center gap-2 p-0 no-underline hover:no-underline"
                            onClick={() => handleColumnSort(column)}
                            variant="link"
                        >
                            Status
                            {sorted === 'asc' ? (
                                <ArrowDown className="size-4" />
                            ) : sorted === 'desc' ? (
                                <ArrowUp className="size-4" />
                            ) : null}
                        </Button>
                    );
                },
                maxSize: 130,
                minSize: 80,
                size: 100,
            },
            {
                accessorKey: 'provider',
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
                header: ({ column }) => {
                    const sorted = column.getIsSorted();

                    return (
                        <Button
                            className="text-muted-foreground hover:text-primary flex items-center gap-2 p-0 no-underline hover:no-underline"
                            onClick={() => handleColumnSort(column)}
                            variant="link"
                        >
                            Provider
                            {sorted === 'asc' ? (
                                <ArrowDown className="size-4" />
                            ) : sorted === 'desc' ? (
                                <ArrowUp className="size-4" />
                            ) : null}
                        </Button>
                    );
                },
                maxSize: 150,
                minSize: 80,
                size: 100,
                sortingFn: (rowA, rowB) => {
                    const nameA = rowA.original.provider?.name || '';
                    const nameB = rowB.original.provider?.name || '';

                    return nameA.localeCompare(nameB);
                },
            },
            {
                accessorKey: 'terminals',
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
                header: ({ column }) => {
                    const sorted = column.getIsSorted();

                    return (
                        <Button
                            className="text-muted-foreground hover:text-primary flex items-center gap-2 p-0 no-underline hover:no-underline"
                            onClick={() => handleColumnSort(column)}
                            variant="link"
                        >
                            Terminals
                            {sorted === 'asc' ? (
                                <ArrowDown className="size-4" />
                            ) : sorted === 'desc' ? (
                                <ArrowUp className="size-4" />
                            ) : null}
                        </Button>
                    );
                },
                maxSize: 220,
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

                    return (
                        <Tooltip>
                            <TooltipTrigger asChild>
                                <div className="cursor-default text-sm">{formatDateTime(dateString)}</div>
                            </TooltipTrigger>
                            <TooltipContent>
                                <div className="text-xs">{formatFullDateTime(dateString)}</div>
                            </TooltipContent>
                        </Tooltip>
                    );
                },
                header: ({ column }) => {
                    const sorted = column.getIsSorted();

                    return (
                        <Button
                            className="text-muted-foreground hover:text-primary flex items-center gap-2 p-0 no-underline hover:no-underline"
                            onClick={() => handleColumnSort(column)}
                            variant="link"
                        >
                            Created
                            {sorted === 'asc' ? (
                                <ArrowDown className="size-4" />
                            ) : sorted === 'desc' ? (
                                <ArrowUp className="size-4" />
                            ) : null}
                        </Button>
                    );
                },
                maxSize: 140,
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

                    return (
                        <Tooltip>
                            <TooltipTrigger asChild>
                                <div className="cursor-default text-sm">{formatDateTime(dateString)}</div>
                            </TooltipTrigger>
                            <TooltipContent>
                                <div className="text-xs">{formatFullDateTime(dateString)}</div>
                            </TooltipContent>
                        </Tooltip>
                    );
                },
                header: ({ column }) => {
                    const sorted = column.getIsSorted();

                    return (
                        <Button
                            className="text-muted-foreground hover:text-primary flex items-center gap-2 p-0 no-underline hover:no-underline"
                            onClick={() => handleColumnSort(column)}
                            variant="link"
                        >
                            Updated
                            {sorted === 'asc' ? (
                                <ArrowDown className="size-4" />
                            ) : sorted === 'desc' ? (
                                <ArrowUp className="size-4" />
                            ) : null}
                        </Button>
                    );
                },
                maxSize: 140,
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
                    const isEditing = editingFlowId === flow.id;

                    if (isEditing) {
                        return (
                            <div className="flex items-center justify-end gap-1">
                                <Button
                                    className="size-8 p-0"
                                    disabled={isRenameLoading || !editingFlowTitle.trim()}
                                    onClick={(e) => {
                                        e.stopPropagation();
                                        handleFlowRenameSave();
                                    }}
                                    variant="ghost"
                                >
                                    {isRenameLoading ? (
                                        <Loader2 className="size-4 animate-spin" />
                                    ) : (
                                        <Check className="size-4" />
                                    )}
                                </Button>
                                <Button
                                    className="size-8 p-0"
                                    onClick={(e) => {
                                        e.stopPropagation();
                                        handleFlowRenameCancel();
                                    }}
                                    variant="ghost"
                                >
                                    <X className="size-4" />
                                </Button>
                            </div>
                        );
                    }

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
                                        className="size-8 p-0"
                                        onClick={(e) => e.stopPropagation()}
                                        variant="ghost"
                                    >
                                        <MoreHorizontal />
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
                                        <Pencil className="size-3" />
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
                minSize: 90,
                size: 96,
            },
        ],
        [
            deletingFlowIds,
            editingFlowId,
            editingFlowTitle,
            finishingFlowIds,
            handleColumnSort,
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

    // Memoize onRowClick to prevent unnecessary rerenders
    const handleRowClick = useCallback(
        (flow: Flow) => {
            if (editingFlowId !== flow.id) {
                handleFlowOpen(flow.id);
            }
        },
        [editingFlowId, handleFlowOpen],
    );

    const pageHeader = (
        <header className="bg-background sticky top-0 z-10 flex h-12 w-full shrink-0 items-center gap-2 border-b transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-12">
            <div className="flex items-center gap-2 px-4">
                <SidebarTrigger className="-ml-1" />
                <Separator
                    className="h-4"
                    orientation="vertical"
                />
                <Breadcrumb>
                    <BreadcrumbList>
                        <BreadcrumbItem>
                            <GitFork className="size-4" />
                            <BreadcrumbPage>Flows</BreadcrumbPage>
                        </BreadcrumbItem>
                    </BreadcrumbList>
                </Breadcrumb>
            </div>
            <div className="ml-auto flex items-center gap-2 px-4">
                <Button
                    onClick={() => navigate('/flows/new')}
                    size="sm"
                    variant="secondary"
                >
                    <Plus />
                    New Flow
                </Button>
            </div>
        </header>
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

    // Check if flows list is empty
    if (flows.length === 0) {
        return (
            <>
                {pageHeader}
                <div className="flex flex-col gap-4 p-4">
                    <StatusCard
                        action={
                            <Button
                                onClick={() => navigate('/flows/new')}
                                variant="secondary"
                            >
                                <Plus className="size-4" />
                                New Flow
                            </Button>
                        }
                        description="Get started by creating your first conversation flow"
                        icon={<FileText className="text-muted-foreground size-8" />}
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
                    columnVisibility={columnVisibility}
                    data={flows}
                    filterColumn="title"
                    filterPlaceholder="Filter flows..."
                    onColumnVisibilityChange={(visibility) => {
                        Object.entries(visibility).forEach(([columnId, isVisible]) => {
                            if (columnVisibility[columnId] !== isVisible) {
                                updateColumnVisibility(columnId, isVisible);
                            }
                        });
                    }}
                    onPageChange={handlePageChange}
                    onRowClick={handleRowClick}
                    pageIndex={currentPage}
                    tableKey="flows"
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
};

export default Flows;
