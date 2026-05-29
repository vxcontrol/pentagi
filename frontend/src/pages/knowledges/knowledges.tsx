import type { ColumnDef } from '@tanstack/react-table';

import { Ellipsis, LibraryBig, Loader2, Pencil, PencilLine, Plus, Trash } from 'lucide-react';
import { useCallback, useRef, useState } from 'react';
import { useLocation, useNavigate, useSearchParams } from 'react-router-dom';
import { toast } from 'sonner';

import type { BadgeVariant } from '@/components/ui/badge';

import ConfirmationDialog from '@/components/shared/confirmation-dialog';
import { HeaderButton } from '@/components/shared/header-button';
import { InlineEditInput } from '@/components/shared/inline-edit';
import { Badge } from '@/components/ui/badge';
import { Breadcrumb, BreadcrumbItem, BreadcrumbList, BreadcrumbPage } from '@/components/ui/breadcrumb';
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
import { InputSearch } from '@/components/ui/input-search';
import { Separator } from '@/components/ui/separator';
import { SidebarTrigger } from '@/components/ui/sidebar';
import { StatusCard } from '@/components/ui/status-card';
import { KnowledgeDocType } from '@/graphql/types';
import { useTableState } from '@/hooks/use-table-state';
import { mergeHrefWithSearchParams, URL_PARAMS } from '@/lib/url-params';
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

function Knowledges() {
    const navigate = useNavigate();
    const location = useLocation();
    const { deleteKnowledge, isLoading, knowledges, updateKnowledge } = useKnowledges();
    const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
    const [deletingKnowledge, setDeletingKnowledge] = useState<Knowledge | null>(null);
    const [deletingIds, setDeletingIds] = useState<Set<string>>(new Set());
    const [editingKnowledgeId, setEditingKnowledgeId] = useState<null | string>(null);
    const [isRenameLoading, setIsRenameLoading] = useState(false);
    const editingInputRef = useRef<HTMLInputElement>(null);

    const [searchParams, setSearchParams] = useSearchParams();
    const { filter, setFilter } = useTableState();

    // Source-of-truth for the semantic-search input is the URL. The
    // `KnowledgesProvider` reads the same `?qs=` and debounces it before
    // hitting `searchKnowledge`, so we keep the input's `value` un-debounced
    // here — the user gets instant feedback in the box, the network only
    // fires after 400 ms of inactivity.
    const semanticQuery = searchParams.get(URL_PARAMS.SEARCH) ?? '';
    const handleSemanticQueryChange = useCallback(
        (value: string) => {
            setSearchParams(
                (prev) => {
                    const next = new URLSearchParams(prev);

                    if (value.trim().length === 0) {
                        // Drop the param entirely so the URL stays canonical
                        // (`/knowledges`, not `/knowledges?qs=`) — list-mode
                        // and the cache key both prefer the absent form.
                        next.delete(URL_PARAMS.SEARCH);
                    } else {
                        next.set(URL_PARAMS.SEARCH, value);
                    }

                    return next;
                },
                // Replace so typing keystrokes don't pile up in the history
                // stack — each char would otherwise be its own back-button
                // stop. Same convention as `useTableState` uses for `?q=`.
                { replace: true },
            );
        },
        [setSearchParams],
    );

    const handleOpen = useCallback(
        (id: string) => {
            navigate(mergeHrefWithSearchParams(`/knowledges/${id}`, new URLSearchParams(location.search)));
        },
        [navigate, location.search],
    );

    const handleDeleteDialogOpen = useCallback((knowledge: Knowledge) => {
        setDeletingKnowledge(knowledge);
        setIsDeleteDialogOpen(true);
    }, []);

    const handleKnowledgeRenameStart = useCallback((knowledge: Knowledge) => {
        setEditingKnowledgeId(knowledge.id);
    }, []);

    const handleKnowledgeRenameCancel = useCallback(() => {
        setEditingKnowledgeId(null);
    }, []);

    const handleKnowledgeRenameSave = useCallback(async () => {
        const newQuestion = editingInputRef.current?.value.trim();

        if (!editingKnowledgeId || !newQuestion) {
            return;
        }

        const knowledge = knowledges.find((k) => k.id === editingKnowledgeId);

        if (!knowledge) {
            return;
        }

        if (newQuestion === knowledge.question) {
            setEditingKnowledgeId(null);

            return;
        }

        setIsRenameLoading(true);

        try {
            // Backend requires `content` on update (it always re-embeds), so we
            // pass it through unchanged from the cached document.
            await updateKnowledge(editingKnowledgeId, {
                content: knowledge.content,
                question: newQuestion,
            });
            toast.success('Knowledge renamed successfully');
            setEditingKnowledgeId(null);
        } catch {
            // Error already handled in provider with toast
        } finally {
            setIsRenameLoading(false);
        }
    }, [editingKnowledgeId, knowledges, updateKnowledge]);

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
            header: ({ column }) => (
                <DataTableColumnHeader
                    column={column}
                    title="Type"
                />
            ),
            maxSize: 180,
            meta: { columnMenuLabel: 'Type', searchable: true },
            minSize: 110,
            size: 130,
        },
        {
            accessorKey: 'question',
            cell: ({ row }) => {
                const knowledge = row.original;
                const isEditing = editingKnowledgeId === knowledge.id;
                const question = row.getValue('question') as string;

                if (isEditing) {
                    return (
                        <div onClick={(e) => e.stopPropagation()}>
                            <InlineEditInput
                                autoFocus
                                busy={isRenameLoading}
                                defaultValue={question}
                                inputRef={editingInputRef}
                                onCancel={handleKnowledgeRenameCancel}
                                onSave={handleKnowledgeRenameSave}
                                placeholder="Knowledge question"
                            />
                        </div>
                    );
                }

                return (
                    <div
                        className="truncate font-medium"
                        title={question}
                    >
                        {question}
                    </div>
                );
            },
            header: ({ column }) => (
                <DataTableColumnHeader
                    column={column}
                    title="Question"
                />
            ),
            meta: { columnMenuLabel: 'Question', searchable: true },
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
            meta: { columnMenuLabel: 'Preview', searchable: true },
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
            header: () => (
                <span className="text-muted-foreground inline-flex w-full items-center justify-end text-sm font-medium">
                    Flags
                </span>
            ),
            id: 'flags',
            maxSize: 200,
            meta: { columnMenuLabel: 'Flags' },
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
                                    aria-label="Open menu"
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
                                <DropdownMenuItem onClick={() => handleKnowledgeRenameStart(k)}>
                                    <PencilLine />
                                    Rename
                                </DropdownMenuItem>
                                <DropdownMenuSeparator />
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
            <ContextMenuItem onClick={() => handleKnowledgeRenameStart(k)}>
                <PencilLine />
                Rename
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
            <div className="flex min-w-0 flex-1 items-center gap-2 px-4">
                <SidebarTrigger className="-ml-1 shrink-0" />
                <Separator
                    className="h-4 shrink-0"
                    orientation="vertical"
                />
                <Breadcrumb className="min-w-0 flex-1">
                    <BreadcrumbList className="min-w-0 flex-nowrap">
                        <BreadcrumbItem className="min-w-0">
                            <LibraryBig className="size-4 shrink-0" />
                            <BreadcrumbPage className="min-w-0 truncate">Knowledges</BreadcrumbPage>
                        </BreadcrumbItem>
                    </BreadcrumbList>
                </Breadcrumb>
            </div>
            <div className="flex shrink-0 items-center gap-2 px-4">
                <InputSearch
                    ariaLabel="Search knowledge documents"
                    // Use Mod+K — Mod+F is reserved as the page-wide default
                    // because we don't want to conflict with the browser's
                    // own find-in-page on every screen, but this list is one
                    // of the few that benefits from a dedicated shortcut.
                    hotkey="k"
                    maxWidth={220}
                    onSearchChange={handleSemanticQueryChange}
                    placeholder="Semantic search..."
                    searchQuery={semanticQuery}
                />
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
                                <Plus />
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
                    empty={{ entityName: 'knowledge documents' }}
                    filterPlaceholder="Filter knowledge documents..."
                    filterValue={filter}
                    onFilterChange={setFilter}
                    onRowClick={(k) => {
                        if (editingKnowledgeId !== k.id) {
                            handleOpen(k.id);
                        }
                    }}
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
}

export default Knowledges;
