import ConfirmationDialog from '@/components/shared/ConfirmationDialog';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
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
import {
    StatusType,
    useDeleteFlowMutation,
    useFinishFlowMutation,
    useFlowsQuery,
    type FlowOverviewFragmentFragment,
} from '@/graphql/types';
import { type ColumnDef } from '@tanstack/react-table';
import {
    AlertCircle,
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
    Plus,
    Trash,
} from 'lucide-react';
import { useState } from 'react';
import { useNavigate } from 'react-router-dom';

type Flow = FlowOverviewFragmentFragment;

const statusConfig: Record<
    StatusType,
    { label: string; variant: 'default' | 'secondary' | 'destructive' | 'outline'; icon?: React.ReactNode }
> = {
    [StatusType.Created]: {
        label: 'Created',
        variant: 'outline',
        icon: <CircleDashed className="h-3 w-3 text-blue-500" />,
    },
    [StatusType.Running]: {
        label: 'Running',
        variant: 'default',
        icon: <Loader2 className="h-3 w-3 text-purple-500 animate-spin" />,
    },
    [StatusType.Waiting]: {
        label: 'Waiting',
        variant: 'outline',
        icon: <CircleDashed className="h-3 w-3 text-yellow-500" />,
    },
    [StatusType.Finished]: {
        label: 'Finished',
        variant: 'secondary',
        icon: <CircleCheck className="h-3 w-3 text-green-500" />,
    },
    [StatusType.Failed]: {
        label: 'Failed',
        variant: 'destructive',
        icon: <CircleX className="h-3 w-3 text-red-500" />,
    },
};

const Flows = () => {
    const navigate = useNavigate();
    const { data, loading: isLoading, error, refetch } = useFlowsQuery();
    const [deleteFlow, { loading: isDeleteLoading, error: deleteError }] = useDeleteFlowMutation();
    const [finishFlow, { loading: isFinishLoading }] = useFinishFlowMutation();
    const [deleteErrorMessage, setDeleteErrorMessage] = useState<string | null>(null);
    const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
    const [deletingFlow, setDeletingFlow] = useState<Flow | null>(null);

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

        try {
            setDeleteErrorMessage(null);

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
            setDeleteErrorMessage(null);
            await refetch();
        } catch (error) {
            setDeleteErrorMessage(error instanceof Error ? error.message : 'An error occurred while deleting');
        }
    };

    const handleFlowFinish = async (flowId: string) => {
        try {
            await finishFlow({
                variables: { flowId },
                refetchQueries: ['flows'],
            });
            await refetch();
        } catch (error) {
            // ignore
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
                        className="flex items-center gap-2 p-0 no-underline hover:no-underline text-muted-foreground hover:text-primary"
                    >
                        ID
                        {sorted === 'asc' ? (
                            <ArrowDown className="h-4 w-4" />
                        ) : sorted === 'desc' ? (
                            <ArrowUp className="h-4 w-4" />
                        ) : null}
                    </Button>
                );
            },
            cell: ({ row }) => <div className="text-sm font-mono">{row.getValue('id')}</div>,
        },
        {
            accessorKey: 'title',
            header: ({ column }) => {
                const sorted = column.getIsSorted();

                return (
                    <Button
                        variant="link"
                        onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
                        className="flex items-center gap-2 p-0 no-underline hover:no-underline text-muted-foreground hover:text-primary"
                    >
                        Title
                        {sorted === 'asc' ? (
                            <ArrowDown className="h-4 w-4" />
                        ) : sorted === 'desc' ? (
                            <ArrowUp className="h-4 w-4" />
                        ) : null}
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
                        className="flex items-center gap-2 p-0 no-underline hover:no-underline text-muted-foreground hover:text-primary"
                    >
                        Status
                        {sorted === 'asc' ? (
                            <ArrowDown className="h-4 w-4" />
                        ) : sorted === 'desc' ? (
                            <ArrowUp className="h-4 w-4" />
                        ) : null}
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
                const isRunning = flow.status === StatusType.Running;

                return (
                    <div className="flex justify-end">
                        <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                                <Button
                                    variant="ghost"
                                    className="h-8 w-8 p-0"
                                >
                                    <MoreHorizontal />
                                </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent
                                align="end"
                                className="min-w-[6rem]"
                            >
                                <DropdownMenuItem onClick={() => handleFlowOpen(flow.id)}>
                                    <Eye />
                                    View
                                </DropdownMenuItem>
                                {isRunning && (
                                    <>
                                        <DropdownMenuSeparator />
                                        <DropdownMenuItem
                                            onClick={() => handleFlowFinish(flow.id)}
                                            disabled={isFinishLoading}
                                        >
                                            {isFinishLoading ? (
                                                <>
                                                    <Loader2 className="animate-spin" />
                                                    Finishing...
                                                </>
                                            ) : (
                                                <>
                                                    <CircleDashed />
                                                    Finish
                                                </>
                                            )}
                                        </DropdownMenuItem>
                                    </>
                                )}
                                <DropdownMenuSeparator />
                                <DropdownMenuItem
                                    onClick={() => handleFlowDeleteDialogOpen(flow)}
                                    disabled={isDeleteLoading && deletingFlow?.id === flow.id}
                                >
                                    {isDeleteLoading && deletingFlow?.id === flow.id ? (
                                        <>
                                            <Loader2 className="h-4 w-4 animate-spin" />
                                            Deleting...
                                        </>
                                    ) : (
                                        <>
                                            <Trash className="h-4 w-4" />
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

    if (error) {
        return (
            <>
                {pageHeader}
                <div className="flex flex-col gap-4 p-4">
                    <Alert variant="destructive">
                        <AlertCircle className="size-4" />
                        <AlertTitle>Error loading flows</AlertTitle>
                        <AlertDescription>{error?.message}</AlertDescription>
                    </Alert>
                </div>
            </>
        );
    }

    // Check if flows list is empty
    if (!flows.length) {
        return (
            <>
                {pageHeader}
                <div className="flex flex-col gap-4 p-4">
                    <StatusCard
                        icon={<FileText className="size-8 text-muted-foreground" />}
                        title="No flows found"
                        description="Get started by creating your first conversation flow"
                        action={
                            <Button
                                onClick={() => navigate('/flows/new')}
                                variant="secondary"
                            >
                                <Plus className="h-4 w-4" />
                                New Flow
                            </Button>
                        }
                    />
                </div>
            </>
        );
    }

    return (
        <>
            {pageHeader}
            <div className="flex flex-col gap-4 p-4 pt-0">
                {/* Delete Error Alert */}
                {(deleteError || deleteErrorMessage) && (
                    <Alert variant="destructive">
                        <AlertCircle className="h-4 w-4" />
                        <AlertTitle>Error deleting flow</AlertTitle>
                        <AlertDescription>{deleteError?.message || deleteErrorMessage}</AlertDescription>
                    </Alert>
                )}

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
