import type { ColumnDef } from '@tanstack/react-table';

import { NetworkStatus } from '@apollo/client';
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

import type { FlowOverviewFragmentFragment } from '@/graphql/types';

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
import { StatusType, useDeleteFlowMutation, useFinishFlowMutation, useFlowsQuery } from '@/graphql/types';

type Flow = FlowOverviewFragmentFragment;

const statusConfig: Record<
    StatusType,
    { icon?: React.ReactNode; label: string; variant: 'default' | 'destructive' | 'outline' | 'secondary' }
> = {
    [StatusType.Created]: {
        icon: <CircleDashed className="size-3 text-blue-500" />,
        label: 'Created',
        variant: 'outline',
    },
    [StatusType.Failed]: {
        icon: <CircleX className="size-3 text-red-500" />,
        label: 'Failed',
        variant: 'destructive',
    },
    [StatusType.Finished]: {
        icon: <CircleCheck className="size-3 text-green-500" />,
        label: 'Finished',
        variant: 'secondary',
    },
    [StatusType.Running]: {
        icon: <Loader2 className="size-3 animate-spin text-purple-500" />,
        label: 'Running',
        variant: 'default',
    },
    [StatusType.Waiting]: {
        icon: <CircleDashed className="size-3 text-yellow-500" />,
        label: 'Waiting',
        variant: 'outline',
    },
};

const Flows = () => {
    const navigate = useNavigate();
    const { data, error, loading, networkStatus } = useFlowsQuery({
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
                refetchQueries: ['flows'],
                update: (cache) => {
                    // Remove the flow from Apollo cache
                    cache.evict({ id: `Flow:${flowId}` });
                    cache.gc();
                },
                variables: { flowId },
            });

            setDeletingFlow(null);

            toast.success('Flow deleted successfully', {
                description: flowDescription,
                id: loadingToastId,
            });
        } catch (error) {
            const errorMessage = error instanceof Error ? error.message : 'An error occurred while deleting';
            toast.error(errorMessage, {
                description: flowDescription,
                id: loadingToastId,
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
                refetchQueries: ['flows'],
                variables: { flowId },
            });

            toast.success('Flow finished successfully', {
                description: flowDescription,
                id: loadingToastId,
            });
        } catch (error) {
            const errorMessage = error instanceof Error ? error.message : 'An error occurred while finishing';
            toast.error(errorMessage, {
                description: flowDescription,
                id: loadingToastId,
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
            cell: ({ row }) => <div className="font-mono text-sm">{row.getValue('id')}</div>,
            header: ({ column }) => {
                const sorted = column.getIsSorted();

                return (
                    <Button
                        className="flex items-center gap-2 p-0 text-muted-foreground no-underline hover:text-primary hover:no-underline"
                        onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
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
            size: 64,
        },
        {
            accessorKey: 'title',
            cell: ({ row }) => <div className="font-medium">{row.getValue('title')}</div>,
            header: ({ column }) => {
                const sorted = column.getIsSorted();

                return (
                    <Button
                        className="flex items-center gap-2 p-0 text-muted-foreground no-underline hover:text-primary hover:no-underline"
                        onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
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
        },
        {
            accessorKey: 'status',
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
            header: ({ column }) => {
                const sorted = column.getIsSorted();

                return (
                    <Button
                        className="flex items-center gap-2 p-0 text-muted-foreground no-underline hover:text-primary hover:no-underline"
                        onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
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
            size: 160,
        },
        {
            cell: ({ row }) => {
                const flow = row.original;
                const isRunning = ![StatusType.Failed, StatusType.Finished].includes(flow.status);

                return (
                    <div className="flex justify-end">
                        <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                                <Button
                                    className="size-8 p-0"
                                    variant="ghost"
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
                                        disabled={finishingFlowIds.has(flow.id)}
                                        onClick={() => handleFlowFinish(flow.id)}
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
            size: 48,
        },
    ];

    const flows = data?.flows || [];

    const pageHeader = (
        <header className="sticky top-0 z-10 flex h-12 w-full shrink-0 items-center gap-2 border-b bg-background transition-[width,height] ease-linear group-has-[[data-collapsible=icon]]/sidebar-wrapper:h-12">
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
                        icon={<Loader2 className="size-16 animate-spin text-muted-foreground" />}
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
                        icon={<FileText className="size-8 text-muted-foreground" />}
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
                <DataTable
                    columns={columns}
                    data={flows}
                    filterColumn="title"
                    filterPlaceholder="Filter flows..."
                    onRowClick={(flow) => handleFlowOpen(flow.id)}
                />

                <ConfirmationDialog
                    cancelText="Cancel"
                    confirmText="Delete"
                    handleConfirm={() => handleFlowDelete(deletingFlow?.id)}
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
