import type { ColumnDef } from '@tanstack/react-table';

import { Ellipsis, FileText, Loader2, Pencil, PencilLine, Plus, Trash } from 'lucide-react';
import { useCallback, useRef, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { toast } from 'sonner';

import ConfirmationDialog from '@/components/shared/confirmation-dialog';
import { HeaderButton } from '@/components/shared/header-button';
import { InlineEditInput } from '@/components/shared/inline-edit';
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
import { Separator } from '@/components/ui/separator';
import { SidebarTrigger } from '@/components/ui/sidebar';
import { StatusCard } from '@/components/ui/status-card';
import { useTableState } from '@/hooks/use-table-state';
import { mergeHrefWithSearchParams } from '@/lib/url-params';
import { type Template, useTemplates } from '@/providers/templates-provider';

function Templates() {
    const navigate = useNavigate();
    const location = useLocation();
    const { deleteTemplate, templates, updateTemplate } = useTemplates();
    const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
    const [deletingTemplate, setDeletingTemplate] = useState<null | Template>(null);
    const [deletingIds, setDeletingIds] = useState<Set<string>>(new Set());
    const [editingTemplateId, setEditingTemplateId] = useState<null | string>(null);
    const [isRenameLoading, setIsRenameLoading] = useState(false);
    const editingInputRef = useRef<HTMLInputElement>(null);

    const { filter, setFilter } = useTableState();

    const handleTemplateOpen = useCallback(
        (templateId: string) => {
            navigate(mergeHrefWithSearchParams(`/templates/${templateId}`, new URLSearchParams(location.search)));
        },
        [navigate, location.search],
    );

    const handleDeleteDialogOpen = useCallback((template: Template) => {
        setDeletingTemplate(template);
        setIsDeleteDialogOpen(true);
    }, []);

    const handleTemplateRenameStart = useCallback((template: Template) => {
        setEditingTemplateId(template.id);
    }, []);

    const handleTemplateRenameCancel = useCallback(() => {
        setEditingTemplateId(null);
    }, []);

    const handleTemplateRenameSave = useCallback(async () => {
        const newTitle = editingInputRef.current?.value.trim();

        if (!editingTemplateId || !newTitle) {
            return;
        }

        const template = templates.find((t) => t.id === editingTemplateId);

        if (!template) {
            return;
        }

        if (newTitle === template.title) {
            setEditingTemplateId(null);

            return;
        }

        setIsRenameLoading(true);

        try {
            await updateTemplate(editingTemplateId, { text: template.text, title: newTitle });
            toast.success('Template renamed successfully');
            setEditingTemplateId(null);
        } catch {
            // Error already handled in provider with toast
        } finally {
            setIsRenameLoading(false);
        }
    }, [editingTemplateId, templates, updateTemplate]);

    const handleDelete = async () => {
        if (!deletingTemplate) {
            return;
        }

        setDeletingIds((prev) => new Set(prev).add(deletingTemplate.id));

        try {
            await deleteTemplate(deletingTemplate.id);
            setDeletingTemplate(null);
        } catch {
            // Error already handled in provider with toast
        } finally {
            setDeletingIds((prev) => {
                const next = new Set(prev);
                next.delete(deletingTemplate.id);

                return next;
            });
        }
    };

    const columns: ColumnDef<Template>[] = [
        {
            accessorKey: 'title',
            cell: ({ row }) => {
                const template = row.original;
                const isEditing = editingTemplateId === template.id;
                const title = row.getValue('title') as string;

                if (isEditing) {
                    return (
                        <div onClick={(e) => e.stopPropagation()}>
                            <InlineEditInput
                                autoFocus
                                busy={isRenameLoading}
                                defaultValue={title}
                                inputRef={editingInputRef}
                                onCancel={handleTemplateRenameCancel}
                                onSave={handleTemplateRenameSave}
                                placeholder="Template title"
                            />
                        </div>
                    );
                }

                return <div className="max-w-[380px] truncate font-medium">{title}</div>;
            },
            header: ({ column }) => (
                <DataTableColumnHeader
                    column={column}
                    title="Title"
                />
            ),
            meta: { searchable: true },
        },
        {
            accessorKey: 'text',
            cell: ({ row }) => {
                const text = (row.getValue('text') as string) ?? '';

                return <div className="text-muted-foreground max-w-[380px] truncate text-sm">{text}</div>;
            },
            header: ({ column }) => (
                <DataTableColumnHeader
                    column={column}
                    title="Text"
                />
            ),
            meta: { searchable: true },
        },
        {
            cell: ({ row }) => {
                const template = row.original;

                return (
                    <div className="flex items-center justify-end gap-1 opacity-0 transition-opacity group-hover:opacity-100">
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
                                <DropdownMenuItem onClick={() => handleTemplateOpen(template.id)}>
                                    <Pencil />
                                    Edit
                                </DropdownMenuItem>
                                <DropdownMenuItem onClick={() => handleTemplateRenameStart(template)}>
                                    <Pencil className="size-3" />
                                    Rename
                                </DropdownMenuItem>
                                <DropdownMenuSeparator />
                                <DropdownMenuItem
                                    disabled={deletingIds.has(template.id)}
                                    onClick={() => handleDeleteDialogOpen(template)}
                                >
                                    {deletingIds.has(template.id) ? (
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
            meta: { preventRowClick: true },
            size: 48,
        },
    ];

    const renderRowContextMenu = (template: Template) => (
        <>
            <ContextMenuItem onClick={() => handleTemplateOpen(template.id)}>
                <Pencil />
                Edit
            </ContextMenuItem>
            <ContextMenuItem onClick={() => handleTemplateRenameStart(template)}>
                <PencilLine />
                Rename
            </ContextMenuItem>
            <ContextMenuSeparator />
            <ContextMenuItem
                disabled={deletingIds.has(template.id)}
                onClick={() => handleDeleteDialogOpen(template)}
            >
                <Trash />
                {deletingIds.has(template.id) ? 'Deleting...' : 'Delete'}
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
                            <FileText className="size-4 shrink-0" />
                            <BreadcrumbPage className="min-w-0 truncate">Templates</BreadcrumbPage>
                        </BreadcrumbItem>
                    </BreadcrumbList>
                </Breadcrumb>
            </div>
            <div className="flex shrink-0 items-center gap-2 px-4">
                <HeaderButton
                    icon={<Plus />}
                    label="New Template"
                    onClick={() => navigate('/templates/new')}
                    variant="secondary"
                />
            </div>
        </header>
    );

    if (!templates.length) {
        return (
            <>
                {pageHeader}
                <div className="flex flex-col gap-4 p-4">
                    <StatusCard
                        action={
                            <Button
                                onClick={() => navigate('/templates/new')}
                                variant="secondary"
                            >
                                <Plus className="size-4" />
                                New Template
                            </Button>
                        }
                        description="Create your first template to get started"
                        icon={<FileText className="text-muted-foreground size-8" />}
                        title="No templates yet"
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
                    data={templates}
                    empty={{ entityName: 'templates' }}
                    filterPlaceholder="Filter templates..."
                    filterValue={filter}
                    onFilterChange={setFilter}
                    onRowClick={(template) => {
                        if (editingTemplateId !== template.id) {
                            handleTemplateOpen(template.id);
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
                    itemName={deletingTemplate?.title}
                    itemType="template"
                />
            </div>
        </>
    );
}

export default Templates;
