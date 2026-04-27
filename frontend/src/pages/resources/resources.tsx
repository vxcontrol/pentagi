import type { ColumnDef } from '@tanstack/react-table';

import { format, formatDistanceToNow } from 'date-fns';
import { ArrowDown, ArrowUp, Folder, Loader2, MoreHorizontal, Plus, Trash } from 'lucide-react';
import { useMemo, useRef, useState } from 'react';

import type { ResourceItem } from '@/features/resources/types';

import ConfirmationDialog from '@/components/shared/confirmation-dialog';
import { Breadcrumb, BreadcrumbItem, BreadcrumbList, BreadcrumbPage } from '@/components/ui/breadcrumb';
import { Button } from '@/components/ui/button';
import { ContextMenuItem } from '@/components/ui/context-menu';
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
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { ACCEPTED_FILE_TYPES } from '@/features/resources/constants';
import { ResourcesUploadOptionsDialog } from '@/features/resources/resources-upload-options-dialog';
import { useFileUpload } from '@/features/resources/use-file-upload';
import { useResources } from '@/providers/resources-provider';

const Resources = () => {
    const { deleteResource, resources } = useResources();
    const fileInputRef = useRef<HTMLInputElement>(null);

    const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
    const [deletingResource, setDeletingResource] = useState<null | ResourceItem>(null);
    const [deletingIds, setDeletingIds] = useState<Set<string>>(new Set());

    const { handleFileChange, handleUploadOptionsCancel, handleUploadOptionsConfirm, isUploadOptionsDialogOpen } =
        useFileUpload();

    const handleUploadFile = () => {
        fileInputRef.current?.click();
    };

    const handleFileInputChange = async (event: React.ChangeEvent<HTMLInputElement>) => {
        await handleFileChange(event.target.files, fileInputRef);
    };

    const handleDeleteDialogOpen = (resource: ResourceItem) => {
        setDeletingResource(resource);
        setIsDeleteDialogOpen(true);
    };

    const handleDelete = async () => {
        if (!deletingResource) {
            return;
        }

        setDeletingIds((prev) => new Set(prev).add(deletingResource.id));

        try {
            await deleteResource(deletingResource.id);
            setDeletingResource(null);
        } catch {
            // Error already handled in provider with toast
        } finally {
            setDeletingIds((prev) => {
                const next = new Set(prev);
                next.delete(deletingResource.id);

                return next;
            });
        }
    };

    const columns: ColumnDef<ResourceItem>[] = useMemo(
        () => [
            {
                accessorKey: 'name',
                cell: ({ row }) => {
                    const resource = row.original;

                    return (
                        <div className="flex min-w-0 flex-col gap-0.5">
                            <span className="truncate font-medium">{resource.name}</span>
                            {!resource.isUploaded && (
                                <span className="text-muted-foreground flex items-center gap-1.5 text-xs">
                                    <Loader2 className="size-3 animate-spin" />
                                    Uploading… {resource.progress}%
                                </span>
                            )}
                        </div>
                    );
                },
                enableHiding: false,
                header: ({ column }) => {
                    const sorted = column.getIsSorted();

                    return (
                        <Button
                            className="text-muted-foreground hover:text-primary flex items-center gap-2 p-0 no-underline hover:no-underline"
                            onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
                            variant="link"
                        >
                            Name
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
                accessorKey: 'format',
                cell: ({ row }) => {
                    const value = (row.getValue('format') as string) ?? '';

                    return <span className="text-muted-foreground uppercase">{value || '—'}</span>;
                },
                header: ({ column }) => {
                    const sorted = column.getIsSorted();

                    return (
                        <Button
                            className="text-muted-foreground hover:text-primary flex items-center gap-2 p-0 no-underline hover:no-underline"
                            onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
                            variant="link"
                        >
                            Format
                            {sorted === 'asc' ? (
                                <ArrowDown className="size-4" />
                            ) : sorted === 'desc' ? (
                                <ArrowUp className="size-4" />
                            ) : null}
                        </Button>
                    );
                },
                size: 120,
            },
            {
                accessorKey: 'uploadedAt',
                cell: ({ row }) => {
                    const value = row.getValue('uploadedAt') as string;
                    const date = new Date(value);
                    const relative = formatDistanceToNow(date, { addSuffix: true });
                    const absolute = format(date, 'd MMM yyyy, HH:mm');

                    return (
                        <Tooltip>
                            <TooltipTrigger asChild>
                                <span className="text-muted-foreground cursor-default text-sm">{relative}</span>
                            </TooltipTrigger>
                            <TooltipContent>
                                <span className="text-xs">{absolute}</span>
                            </TooltipContent>
                        </Tooltip>
                    );
                },
                header: ({ column }) => {
                    const sorted = column.getIsSorted();

                    return (
                        <Button
                            className="text-muted-foreground hover:text-primary flex items-center gap-2 p-0 no-underline hover:no-underline"
                            onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
                            variant="link"
                        >
                            Date
                            {sorted === 'asc' ? (
                                <ArrowDown className="size-4" />
                            ) : sorted === 'desc' ? (
                                <ArrowUp className="size-4" />
                            ) : null}
                        </Button>
                    );
                },
                size: 160,
                sortingFn: (rowA, rowB) =>
                    new Date(rowA.original.uploadedAt).getTime() - new Date(rowB.original.uploadedAt).getTime(),
            },
            {
                cell: ({ row }) => {
                    const resource = row.original;
                    const isDeleting = deletingIds.has(resource.id);
                    const disabled = isDeleting || !resource.isUploaded;

                    return (
                        <div className="flex items-center justify-end gap-1 opacity-0 transition-opacity group-hover:opacity-100">
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
                                    <DropdownMenuItem
                                        disabled={disabled}
                                        onClick={() => handleDeleteDialogOpen(resource)}
                                    >
                                        {isDeleting ? (
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
        ],
        [deletingIds],
    );

    const renderRowContextMenu = (resource: ResourceItem) => {
        const isDeleting = deletingIds.has(resource.id);
        const disabled = isDeleting || !resource.isUploaded;

        return (
            <ContextMenuItem
                disabled={disabled}
                onClick={() => handleDeleteDialogOpen(resource)}
            >
                <Trash />
                {isDeleting ? 'Deleting...' : 'Delete'}
            </ContextMenuItem>
        );
    };

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
                            <Folder className="size-4" />
                            <BreadcrumbPage>Resources</BreadcrumbPage>
                        </BreadcrumbItem>
                    </BreadcrumbList>
                </Breadcrumb>
            </div>
            <div className="ml-auto flex items-center gap-2 px-4">
                <Button
                    onClick={handleUploadFile}
                    size="sm"
                    variant="secondary"
                >
                    <Plus />
                    Add file
                </Button>
            </div>
        </header>
    );

    const fileInput = (
        <input
            accept={ACCEPTED_FILE_TYPES}
            className="hidden"
            multiple
            onChange={handleFileInputChange}
            ref={fileInputRef}
            type="file"
        />
    );

    if (!resources.length) {
        return (
            <>
                {pageHeader}
                {fileInput}
                <div className="flex flex-col gap-4 p-4">
                    <StatusCard
                        action={
                            <Button
                                onClick={handleUploadFile}
                                variant="secondary"
                            >
                                <Plus className="size-4" />
                                Upload file
                            </Button>
                        }
                        description="Upload documents so PentAGI agents can reference them during your flows"
                        icon={<Folder className="text-muted-foreground size-8" />}
                        title="No resources yet"
                    />
                </div>
                <ResourcesUploadOptionsDialog
                    isOpen={isUploadOptionsDialogOpen}
                    onCancel={handleUploadOptionsCancel}
                    onConfirm={handleUploadOptionsConfirm}
                />
            </>
        );
    }

    return (
        <>
            {pageHeader}
            {fileInput}
            <div className="flex flex-col gap-4 p-4 pt-0">
                <DataTable
                    columns={columns}
                    data={resources}
                    filterColumn="name"
                    filterPlaceholder="Filter resources..."
                    renderRowContextMenu={renderRowContextMenu}
                />

                <ConfirmationDialog
                    cancelText="Cancel"
                    confirmText="Delete"
                    handleConfirm={handleDelete}
                    handleOpenChange={setIsDeleteDialogOpen}
                    isOpen={isDeleteDialogOpen}
                    itemName={deletingResource?.name}
                    itemType="resource"
                />

                <ResourcesUploadOptionsDialog
                    isOpen={isUploadOptionsDialogOpen}
                    onCancel={handleUploadOptionsCancel}
                    onConfirm={handleUploadOptionsConfirm}
                />
            </div>
        </>
    );
};

export default Resources;
