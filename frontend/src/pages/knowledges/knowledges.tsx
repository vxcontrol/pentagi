import type { ColumnDef } from '@tanstack/react-table';

import { ArrowDown, ArrowUp, Ellipsis, LibraryBig, Loader2, Pencil, Plus, Trash } from 'lucide-react';
import { useState } from 'react';
import { useNavigate } from 'react-router-dom';

import type { BadgeVariant } from '@/components/ui/badge';

import ConfirmationDialog from '@/components/shared/confirmation-dialog';
import { HeaderButton } from '@/components/shared/header-button';
import { Badge } from '@/components/ui/badge';
import { Breadcrumb, BreadcrumbItem, BreadcrumbList, BreadcrumbPage } from '@/components/ui/breadcrumb';
import { Button } from '@/components/ui/button';
import { ContextMenuItem, ContextMenuSeparator } from '@/components/ui/context-menu';
import { DataTable } from '@/components/ui/data-table';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Separator } from '@/components/ui/separator';
import { SidebarTrigger } from '@/components/ui/sidebar';
import { StatusCard } from '@/components/ui/status-card';
import { KnowledgeDocType } from '@/graphql/types';
import { cycleColumnSort } from '@/lib/table-sort';
import { type Knowledge, useKnowledges } from '@/providers/knowledges-provider';

const docTypeBadgeVariant: Record<KnowledgeDocType, BadgeVariant> = {
    [KnowledgeDocType.Answer]: 'blue',
    [KnowledgeDocType.Code]: 'purple',
    [KnowledgeDocType.Guide]: 'green',
};

const docTypeSubtype = (k: Knowledge): null | string => {
    if (k.docType === KnowledgeDocType.Guide) {
        return k.guideType ?? null;
    }

    if (k.docType === KnowledgeDocType.Answer) {
        return k.answerType ?? null;
    }

    if (k.docType === KnowledgeDocType.Code) {
        return k.codeLang ?? null;
    }

    return null;
};

const Knowledges = () => {
    const navigate = useNavigate();
    const { deleteKnowledge, isLoading, knowledges } = useKnowledges();
    const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
    const [deletingKnowledge, setDeletingKnowledge] = useState<Knowledge | null>(null);
    const [deletingIds, setDeletingIds] = useState<Set<string>>(new Set());

    const handleOpen = (id: string) => {
        navigate(`/knowledges/${id}`);
    };

    const handleDeleteDialogOpen = (knowledge: Knowledge) => {
        setDeletingKnowledge(knowledge);
        setIsDeleteDialogOpen(true);
    };

    const handleDelete = async () => {
        if (!deletingKnowledge) {
            return;
        }

        setDeletingIds((prev) => new Set(prev).add(deletingKnowledge.id));

        try {
            await deleteKnowledge(deletingKnowledge.id);
            setDeletingKnowledge(null);
        } catch {
            // Error already handled in provider with toast
        } finally {
            setDeletingIds((prev) => {
                const next = new Set(prev);
                next.delete(deletingKnowledge.id);

                return next;
            });
        }
    };

    const columns: ColumnDef<Knowledge>[] = [
        {
            accessorKey: 'docType',
            cell: ({ row }) => {
                const docType = row.getValue('docType') as KnowledgeDocType;
                const subtype = docTypeSubtype(row.original);

                return (
                    <div className="flex min-w-0 items-center gap-1.5 overflow-hidden">
                        <Badge
                            className="shrink-0 whitespace-nowrap"
                            variant={docTypeBadgeVariant[docType]}
                        >
                            {docType}
                        </Badge>
                        {subtype ? (
                            <span
                                className="text-muted-foreground truncate text-xs"
                                title={subtype}
                            >
                                {subtype}
                            </span>
                        ) : null}
                    </div>
                );
            },
            header: ({ column }) => {
                const sorted = column.getIsSorted();

                return (
                    <Button
                        className="text-muted-foreground hover:text-primary flex items-center gap-2 p-0 no-underline hover:no-underline"
                        onClick={() => cycleColumnSort(column)}
                        variant="link"
                    >
                        Type
                        {sorted === 'asc' ? (
                            <ArrowDown className="size-4" />
                        ) : sorted === 'desc' ? (
                            <ArrowUp className="size-4" />
                        ) : null}
                    </Button>
                );
            },
            maxSize: 180,
            minSize: 110,
            size: 130,
        },
        {
            accessorKey: 'question',
            cell: ({ row }) => {
                const question = row.getValue('question') as string;

                return (
                    <div
                        className="truncate font-medium"
                        title={question}
                    >
                        {question}
                    </div>
                );
            },
            header: ({ column }) => {
                const sorted = column.getIsSorted();

                return (
                    <Button
                        className="text-muted-foreground hover:text-primary flex items-center gap-2 p-0 no-underline hover:no-underline"
                        onClick={() => cycleColumnSort(column)}
                        variant="link"
                    >
                        Question
                        {sorted === 'asc' ? (
                            <ArrowDown className="size-4" />
                        ) : sorted === 'desc' ? (
                            <ArrowUp className="size-4" />
                        ) : null}
                    </Button>
                );
            },
            minSize: 180,
            size: 280,
        },
        {
            accessorKey: 'content',
            cell: ({ row }) => {
                const content = (row.getValue('content') as string) ?? '';

                return (
                    <div
                        className="text-muted-foreground truncate text-sm"
                        title={content}
                    >
                        {content}
                    </div>
                );
            },
            enableSorting: false,
            header: () => (
                <span className="text-muted-foreground inline-flex items-center text-sm font-medium">Preview</span>
            ),
            maxSize: 800,
            minSize: 160,
            size: 380,
        },
        {
            cell: ({ row }) => {
                const k = row.original;

                return (
                    <div className="flex items-center justify-end gap-1 overflow-hidden">
                        {k.flowId ? (
                            <Badge
                                className="shrink-0 whitespace-nowrap"
                                variant="outline"
                            >
                                flow #{k.flowId}
                            </Badge>
                        ) : null}
                        <Badge
                            className="shrink-0 whitespace-nowrap"
                            variant={k.manual ? 'secondary' : 'outline'}
                        >
                            {k.manual ? 'manual' : 'agent'}
                        </Badge>
                    </div>
                );
            },
            enableSorting: false,
            header: () => null,
            id: 'flags',
            maxSize: 200,
            minSize: 110,
            size: 150,
        },
        {
            cell: ({ row }) => {
                const k = row.original;

                return (
                    <div className="flex items-center justify-end gap-1 opacity-0 transition-opacity group-hover:opacity-100">
                        <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                                <Button
                                    className="size-8 p-0"
                                    onClick={(event) => event.stopPropagation()}
                                    variant="ghost"
                                >
                                    <Ellipsis />
                                </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent
                                align="end"
                                className="min-w-24"
                                onClick={(event) => event.stopPropagation()}
                            >
                                <DropdownMenuItem onClick={() => handleOpen(k.id)}>
                                    <Pencil />
                                    Edit
                                </DropdownMenuItem>
                                <DropdownMenuItem
                                    disabled={deletingIds.has(k.id)}
                                    onClick={() => handleDeleteDialogOpen(k)}
                                >
                                    {deletingIds.has(k.id) ? (
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
            maxSize: 70,
            meta: { preventRowClick: true },
            minSize: 50,
            size: 60,
        },
    ];

    const renderRowContextMenu = (k: Knowledge) => (
        <>
            <ContextMenuItem onClick={() => handleOpen(k.id)}>
                <Pencil />
                Edit
            </ContextMenuItem>
            <ContextMenuSeparator />
            <ContextMenuItem
                disabled={deletingIds.has(k.id)}
                onClick={() => handleDeleteDialogOpen(k)}
            >
                <Trash />
                {deletingIds.has(k.id) ? 'Deleting...' : 'Delete'}
            </ContextMenuItem>
        </>
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
                            <LibraryBig className="size-4" />
                            <BreadcrumbPage>Knowledges</BreadcrumbPage>
                        </BreadcrumbItem>
                    </BreadcrumbList>
                </Breadcrumb>
            </div>
            <div className="ml-auto flex items-center gap-2 px-4">
                <HeaderButton
                    icon={<Plus />}
                    label="New Knowledge"
                    onClick={() => navigate('/knowledges/new')}
                    variant="secondary"
                />
            </div>
        </header>
    );

    if (isLoading && !knowledges.length) {
        return (
            <>
                {pageHeader}
                <div className="flex flex-col gap-4 p-4">
                    <StatusCard
                        description="Please wait while we fetch your knowledge documents"
                        icon={<Loader2 className="text-muted-foreground size-16 animate-spin" />}
                        title="Loading knowledges..."
                    />
                </div>
            </>
        );
    }

    if (!knowledges.length) {
        return (
            <>
                {pageHeader}
                <div className="flex flex-col gap-4 p-4">
                    <StatusCard
                        action={
                            <Button
                                onClick={() => navigate('/knowledges/new')}
                                variant="secondary"
                            >
                                <Plus className="size-4" />
                                New Knowledge
                            </Button>
                        }
                        description="Create your first knowledge document to enrich the vector store"
                        icon={<LibraryBig className="text-muted-foreground size-8" />}
                        title="No knowledge documents yet"
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
                    data={knowledges}
                    filterColumn="question"
                    filterPlaceholder="Filter knowledge documents..."
                    onRowClick={(k) => handleOpen(k.id)}
                    renderRowContextMenu={renderRowContextMenu}
                />

                <ConfirmationDialog
                    cancelText="Cancel"
                    confirmText="Delete"
                    handleConfirm={handleDelete}
                    handleOpenChange={setIsDeleteDialogOpen}
                    isOpen={isDeleteDialogOpen}
                    itemName={deletingKnowledge?.question}
                    itemType="knowledge document"
                />
            </div>
        </>
    );
};

export default Knowledges;
