import type { ColumnDef } from '@tanstack/react-table';

import { ArrowDown, ArrowUp, Eye, FileText, GitFork, Loader2, MoreHorizontal, Pause, Plus, Trash } from 'lucide-react';
import { useState } from 'react';
import { useNavigate } from 'react-router-dom';

import { FlowStatusIcon } from '@/components/icons/flow-status-icon';
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
import { Separator } from '@/components/ui/separator';
import { SidebarTrigger } from '@/components/ui/sidebar';
import { StatusCard } from '@/components/ui/status-card';
import { StatusType } from '@/graphql/types';
import { type Flow, useFlows } from '@/providers/flows-provider';

const statusConfig: Record<
    StatusType,
    { label: string; variant: 'default' | 'destructive' | 'outline-solid' | 'secondary' }
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

const Flows = () => {
    const navigate = useNavigate();
    const { deleteFlow, finishFlow, flows, isLoading } = useFlows();
    const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
    const [deletingFlow, setDeletingFlow] = useState<Flow | null>(null);
    const [finishingFlowIds, setFinishingFlowIds] = useState<Set<string>>(new Set());
    const [deletingFlowIds, setDeletingFlowIds] = useState<Set<string>>(new Set());

    const handleFlowOpen = (flowId: string) => {
        navigate(`/flows/${flowId}`);
    };

    const handleFlowDeleteDialogOpen = (flow: Flow) => {
        setDeletingFlow(flow);
        setIsDeleteDialogOpen(true);
    };

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

    const handleFlowFinish = async (flow: Flow) => {
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
            size: 48,
        },
    ];

    const pageHeader = (
        <header className="sticky top-0 z-10 flex h-12 w-full shrink-0 items-center gap-2 border-b bg-background transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-12">
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
