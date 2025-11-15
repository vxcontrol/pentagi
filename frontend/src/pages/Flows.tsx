import { NetworkStatus } from '@apollo/client';
import type { ColumnDef } from '@tanstack/react-table';
import {
    ArrowDown,
    ArrowUp,
    CircleCheck,
    CircleDashed,
    CircleX,
    Eye,
    FileText,
    GitFork,
    Loader2,
    MoreHorizontal,
    Pause,
    Plus,
    Trash,
} from 'lucide-react';
import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { toast } from 'sonner';

import ConfirmationDialog from '@/components/shared/ConfirmationDialog';
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
import { Separator } from '@/components/ui/separator';
import { SidebarTrigger } from '@/components/ui/sidebar';
import { StatusCard } from '@/components/ui/status-card';
import type { FlowOverviewFragmentFragment } from '@/graphql/types';
import {
    StatusType,
    useDeleteFlowMutation,
    useFinishFlowMutation,
    useFlowsQuery,
} from '@/graphql/types';

type Flow = FlowOverviewFragmentFragment;

const statusConfig: Record<
    StatusType,
    { label: string; variant: 'default' | 'secondary' | 'destructive' | 'outline'; icon?: React.ReactNode }
> = {
    [StatusType.Created]: {
        label: 'Created',
        variant: 'outline',
        icon: <CircleDashed className="size-3 text-blue-500" />,
    },
    [StatusType.Running]: {
        label: 'Running',
        variant: 'default',
        icon: <Loader2 className="size-3 animate-spin text-purple-500" />,
    },
    [StatusType.Waiting]: {
        label: 'Waiting',
        variant: 'outline',
        icon: <CircleDashed className="size-3 text-yellow-500" />,
    },
    [StatusType.Finished]: {
        label: 'Finished',
        variant: 'secondary',
        icon: <CircleCheck className="size-3 text-green-500" />,
    },
    [StatusType.Failed]: {
        label: 'Failed',
        variant: 'destructive',
        icon: <CircleX className="size-3 text-red-500" />,
    },
};

const Flows = () => {
    const navigate = useNavigate();
    const { data, loading, error, networkStatus } = useFlowsQuery({
        notifyOnNetworkStatusChange: true,
    });
    const isLoading = loading && networkStatus === NetworkStatus.loading;
    const [deleteFlow] = useDeleteFlowMutation();
    const [finishFlow] = useFinishFlowMutation();
    const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
    const [deletingFlow, setDeletingFlow] = useState<Flow | null>(null);
    const [finishingFlowIds, setFinishingFlowIds] = useState<Set<string>>(new Set());
    const [deletingFlowIds, setDeletingFlowIds] = useState<Set<string>>(new Set());

    useEffect(() => {
        if (error) {
            toast.error('Error loading flows', {
                description: error.message,
            });
        }
    }, [error]);

    const handleFlowOpen = (flowId: string) => {
        navigate(`/flows/${flowId}`);
    };

    const handleFlowDeleteDialogOpen = (flow: Flow) => {
        setDeletingFlow(flow);
        setIsDeleteDialogOpen(true);
    };

    const handleFlowDelete = async (flowId: string | undefined) => {
        if (!flowId) {
            return;
        }

        const flowTitle = deletingFlow?.title || 'Unknown';
        const flowDescription = `${flowTitle} (ID: ${flowId})`;

        const loadingToastId = toast.loading('Deleting flow...', {
            description: flowDescription,
        });

        try {
            setDeletingFlowIds((previousIds) => new Set(previousIds).add(flowId));

            await deleteFlow({
                variables: { flowId },
                refetchQueries: ['flows'],
                update: (cache) => {
                    // Remove the flow from Apollo cache
                    cache.evict({ id: `Flow:${flowId}` });
                    cache.gc();
                },
            });

            setDeletingFlow(null);

            toast.success('Flow deleted successfully', {
                id: loadingToastId,
                description: flowDescription,
            });
        } catch (error) {
            const errorMessage = error instanceof Error ? error.message : 'An error occurred while deleting';
            toast.error(errorMessage, {
                id: loadingToastId,
                description: flowDescription,
            });
        } finally {
            setDeletingFlowIds((previousIds) => {
                const newIds = new Set(previousIds);
                newIds.delete(flowId);
                return newIds;
            });
        }
    };

    const handleFlowFinish = async (flowId: string) => {
        const flow = data?.flows?.find((f) => f.id === flowId);
        const flowTitle = flow?.title || 'Unknown';
        const flowDescription = `${flowTitle} (ID: ${flowId})`;

        const loadingToastId = toast.loading('Finishing flow...', {
            description: flowDescription,
        });

        try {
            setFinishingFlowIds((previousIds) => new Set(previousIds).add(flowId));

            await finishFlow({
                variables: { flowId },
                refetchQueries: ['flows'],
            });

            toast.success('Flow finished successfully', {
                id: loadingToastId,
                description: flowDescription,
            });
        } catch (error) {
            const errorMessage = error instanceof Error ? error.message : 'An error occurred while finishing';
            toast.error(errorMessage, {
                id: loadingToastId,
                description: flowDescription,
            });
        } finally {
            setFinishingFlowIds((previousIds) => {
                const newIds = new Set(previousIds);
                newIds.delete(flowId);
                return newIds;
            });
        }
    };

    const columns: ColumnDef<Flow>[] = [
        {
            accessorKey: 'id',
            size: 64,
            header: ({ column }) => {
                const sorted = column.getIsSorted();

                return (
                    <Button
                        variant="link"
                        onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
                        className="flex items-center gap-2 p-0 text-muted-foreground no-underline hover:text-primary hover:no-underline"
                    >
                        ID
                        {sorted === 'asc' ? (
                            <ArrowDown className="size-4" />
                        ) : sorted === 'desc'
                            ? (
                                <ArrowUp className="size-4" />
                            )
                            : null}
                    </Button>
                );
            },
            cell: ({ row }) => <div className="font-mono text-sm">{row.getValue('id')}</div>,
        },
        {
            accessorKey: 'title',
            header: ({ column }) => {
                const sorted = column.getIsSorted();

                return (
                    <Button
                        variant="link"
                        onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
                        className="flex items-center gap-2 p-0 text-muted-foreground no-underline hover:text-primary hover:no-underline"
                    >
                        Title
                        {sorted === 'asc' ? (
                            <ArrowDown className="size-4" />
                        ) : sorted === 'desc'
                            ? (
                                <ArrowUp className="size-4" />
                            )
                            : null}
                    </Button>
                );
            },
            cell: ({ row }) => <div className="font-medium">{row.getValue('title')}</div>,
        },
        {
            accessorKey: 'status',
            size: 160,
            header: ({ column }) => {
                const sorted = column.getIsSorted();

                return (
                    <Button
                        variant="link"
                        onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
                        className="flex items-center gap-2 p-0 text-muted-foreground no-underline hover:text-primary hover:no-underline"
                    >
                        Status
                        {sorted === 'asc' ? (
                            <ArrowDown className="size-4" />
                        ) : sorted === 'desc'
                            ? (
                                <ArrowUp className="size-4" />
                            )
                            : null}
                    </Button>
                );
            },
            cell: ({ row }) => {
                const status = row.getValue('status') as StatusType;
                const config = statusConfig[status];
                return (
                    <Badge variant={config.variant}>
                        {config.icon}
                        {config.label}
                    </Badge>
                );
            },
        },
        {
            id: 'actions',
            size: 48,
            enableHiding: false,
            header: () => null,
            cell: ({ row }) => {
                const flow = row.original;
                const isRunning = ![StatusType.Finished, StatusType.Failed].includes(flow.status);

                return (
                    <div className="flex justify-end">
                        <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                                <Button
                                    variant="ghost"
                                    className="size-8 p-0"
                                >
                                    <MoreHorizontal />
                                </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent
                                align="end"
                                className="min-w-24"
                            >
                                <DropdownMenuItem onClick={() => handleFlowOpen(flow.id)}>
                                    <Eye />
                                    View
                                </DropdownMenuItem>
                                {isRunning && (
                                    <DropdownMenuItem
                                        onClick={() => handleFlowFinish(flow.id)}
                                        disabled={finishingFlowIds.has(flow.id)}
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
                                    onClick={() => handleFlowDeleteDialogOpen(flow)}
                                    disabled={deletingFlowIds.has(flow.id)}
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
        },
    ];

    const flows = data?.flows || [];

    const pageHeader = (
        <header className="sticky top-0 z-10 flex h-12 w-full shrink-0 items-center gap-2 border-b bg-background transition-[width,height] ease-linear group-has-[[data-collapsible=icon]]/sidebar-wrapper:h-12">
            <div className="flex items-center gap-2 px-4">
                <SidebarTrigger className="-ml-1" />
                <Separator
                    orientation="vertical"
                    className="h-4"
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
                    variant="secondary"
                    size="sm"
                    onClick={() => navigate('/flows/new')}
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
                        icon={<Loader2 className="size-16 animate-spin text-muted-foreground" />}
                        title="Loading flows..."
                        description="Please wait while we fetch your conversation flows"
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
                        icon={<FileText className="size-8 text-muted-foreground" />}
                        title="No flows found"
                        description="Get started by creating your first conversation flow"
                        action={(
                            <Button
                                onClick={() => navigate('/flows/new')}
                                variant="secondary"
                            >
                                <Plus className="size-4" />
                                New Flow
                            </Button>
                        )}
                    />
                </div>
            </>
        );
    }

    return (
        <>
            {pageHeader}
            <div className="flex flex-col gap-4 p-4 pt-0">
                <DataTable
                    columns={columns}
                    data={flows}
                    filterColumn="title"
                    filterPlaceholder="Filter flows..."
                    onRowClick={(flow) => handleFlowOpen(flow.id)}
                />

                <ConfirmationDialog
                    isOpen={isDeleteDialogOpen}
                    handleOpenChange={setIsDeleteDialogOpen}
                    handleConfirm={() => handleFlowDelete(deletingFlow?.id)}
                    itemName={deletingFlow?.title}
                    itemType="flow"
                    confirmText="Delete"
                    cancelText="Cancel"
                />
            </div>
        </>
    );
};

export default Flows;
