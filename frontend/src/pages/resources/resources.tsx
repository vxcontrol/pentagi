import { Copy, FileSymlink, Folder, FolderPlus, FolderUp, Loader2, Search, X } from 'lucide-react';
import { useCallback, useMemo, useState } from 'react';
import { toast } from 'sonner';

import {
    copyPathAction,
    deleteAction,
    downloadAction,
    FileManager,
    type FileManagerAction,
    type FileNode,
} from '@/components/file-manager';
import ConfirmationDialog from '@/components/shared/confirmation-dialog';
import { Breadcrumb, BreadcrumbItem, BreadcrumbList, BreadcrumbPage } from '@/components/ui/breadcrumb';
import { Button } from '@/components/ui/button';
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';
import { FileDropZone } from '@/components/ui/file-drop-zone';
import { Form, FormControl, FormField, FormItem } from '@/components/ui/form';
import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupInput } from '@/components/ui/input-group';
import { Separator } from '@/components/ui/separator';
import { SidebarTrigger } from '@/components/ui/sidebar';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { ResourcesConflictDialog } from '@/features/resources/resources-conflict-dialog';
import { ResourcesCopyDialog } from '@/features/resources/resources-copy-dialog';
import { ResourcesMkdirDialog } from '@/features/resources/resources-mkdir-dialog';
import { ResourcesMoveDialog } from '@/features/resources/resources-move-dialog';
import { buildResourceDownloadHref, toFileNode } from '@/features/resources/resources-utils';
import { useResourcesDelete } from '@/features/resources/use-resources-delete';
import { useResourcesMove } from '@/features/resources/use-resources-move';
import { useResourcesSearch } from '@/features/resources/use-resources-search';
import { useResourcesUpload } from '@/features/resources/use-resources-upload';
import { useFilesDragAndDrop } from '@/hooks/use-files-drag-and-drop';
import { copyToClipboard } from '@/lib/report';
import { useResources } from '@/providers/resources-provider';

const Resources = () => {
    const { isInitialLoading, resources } = useResources();
    const search = useResourcesSearch();

    const [isMkdirOpen, setIsMkdirOpen] = useState(false);
    const [fileToRename, setFileToRename] = useState<FileNode | null>(null);
    const [fileToCopy, setFileToCopy] = useState<FileNode | null>(null);

    const upload = useResourcesUpload();
    const deletion = useResourcesDelete();
    const moveAction = useResourcesMove();

    const canAcceptDrop = !upload.isUploading;
    const { dragHandlers, isDragging } = useFilesDragAndDrop({
        canAcceptDrop,
        onDrop: upload.uploadFiles,
    });

    const fileNodes = useMemo<FileNode[]>(() => resources.map(toFileNode), [resources]);

    /**
     * Drag-and-drop entry point: move every dragged item into `destinationDir` by issuing
     * one `PUT /resources/move` per source. We don't pre-check or batch — each request
     * goes out independently so a partial failure still moves the rest.
     */
    const handleMoveItems = useCallback(
        async (sources: FileNode[], destinationDir: string) => {
            await Promise.allSettled(
                sources.map((source) => {
                    const destination = destinationDir ? `${destinationDir}/${source.name}` : source.name;

                    return moveAction.move(source.path, { destination, shouldOverwrite: false });
                }),
            );
        },
        [moveAction],
    );

    const handleCopyPath = useCallback(async (file: FileNode) => {
        const wasCopied = await copyToClipboard(file.path);

        if (wasCopied) {
            toast.success('Path copied to clipboard');

            return;
        }

        toast.error('Failed to copy path');
    }, []);

    /**
     * "Open" gesture — fires on double-click or Enter for a file row.
     * Triggers the same download the dropdown's Download action would, by clicking
     * a transient `<a download>` element. We can't just `window.open()` here because
     * we want the browser's `download` attribute hint (preserves the original
     * filename even when the server sends `Content-Disposition: inline`).
     */
    const handleOpenFile = useCallback((file: FileNode) => {
        const anchor = document.createElement('a');

        anchor.href = buildResourceDownloadHref(file);
        anchor.download = file.name;
        anchor.rel = 'noopener';
        document.body.appendChild(anchor);
        anchor.click();
        anchor.remove();
    }, []);

    const fileManagerActions = useMemo<FileManagerAction[]>(
        () => [
            downloadAction(buildResourceDownloadHref),
            copyPathAction(handleCopyPath),
            {
                appliesToDirs: true,
                icon: FileSymlink,
                id: 'resources-rename',
                label: 'Rename or move',
                onSelect: (file) => setFileToRename(file),
            },
            {
                appliesToDirs: true,
                icon: Copy,
                id: 'resources-copy',
                label: 'Copy to…',
                onSelect: (file) => setFileToCopy(file),
            },
            deleteAction(deletion.requestDelete),
        ],
        [deletion.requestDelete, handleCopyPath],
    );

    const handleDeleteDialogOpenChange = useCallback(
        (nextOpen: boolean) => {
            if (!nextOpen) {
                deletion.clearFileToDelete();
            }
        },
        [deletion],
    );

    const pageHeader = (
        <header className="bg-background sticky top-0 z-10 flex h-12 w-full shrink-0 items-center gap-2 border-b transition-[width,height] ease-linear">
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
        </header>
    );

    const hasResources = resources.length > 0;

    const noResourcesState = (
        <FileDropZone
            actionLabel="Upload files"
            description="Upload documents so PentAGI agents can reference them during your flows. You can also drag & drop files anywhere in this panel."
            hint="Up to 300 MB per file · 2 GB per upload"
            isDragging={isDragging}
            isUploading={upload.isUploading}
            onBrowse={upload.openFilePicker}
            title="No resources yet"
        />
    );

    const noMatchesState = (
        <Empty>
            <EmptyHeader>
                <EmptyMedia variant="icon">
                    <Search />
                </EmptyMedia>
                <EmptyTitle>No matches</EmptyTitle>
                <EmptyDescription>
                    No resources match <code>{search.debouncedQuery.trim()}</code>. Try a different query.
                </EmptyDescription>
            </EmptyHeader>
        </Empty>
    );

    return (
        <>
            {pageHeader}
            <div
                className="relative flex flex-1 flex-col gap-4 p-4"
                {...dragHandlers}
            >
                <input
                    className="hidden"
                    key={upload.fileInputKey}
                    multiple
                    type="file"
                    {...upload.fileInputProps}
                />

                {isDragging && hasResources && (
                    <div className="bg-primary/10 border-primary pointer-events-none absolute inset-2 z-30 flex items-center justify-center rounded-lg border-2 border-dashed">
                        <div className="text-primary flex flex-col items-center gap-2">
                            <FolderUp className="size-8" />
                            <span className="text-sm font-medium">Drop files to upload</span>
                        </div>
                    </div>
                )}

                <Form {...search.form}>
                    <div className="flex gap-2 p-px">
                        <FormField
                            control={search.form.control}
                            name="search"
                            render={({ field }) => (
                                <FormItem className="flex-1">
                                    <FormControl>
                                        <InputGroup>
                                            <InputGroupAddon>
                                                <Search />
                                            </InputGroupAddon>
                                            <InputGroupInput
                                                {...field}
                                                autoComplete="off"
                                                placeholder="Search resources..."
                                                type="text"
                                            />
                                            {field.value && (
                                                <InputGroupAddon align="inline-end">
                                                    <InputGroupButton
                                                        onClick={search.resetSearch}
                                                        type="button"
                                                    >
                                                        <X />
                                                    </InputGroupButton>
                                                </InputGroupAddon>
                                            )}
                                        </InputGroup>
                                    </FormControl>
                                </FormItem>
                            )}
                        />

                        <Tooltip>
                            <TooltipTrigger asChild>
                                <span>
                                    <Button
                                        disabled={upload.isUploading}
                                        onClick={() => setIsMkdirOpen(true)}
                                        size="icon-sm"
                                        type="button"
                                        variant="outline"
                                    >
                                        <FolderPlus />
                                    </Button>
                                </span>
                            </TooltipTrigger>
                            <TooltipContent>Create directory</TooltipContent>
                        </Tooltip>

                        <Tooltip>
                            <TooltipTrigger asChild>
                                <span>
                                    <Button
                                        disabled={upload.isUploading}
                                        onClick={upload.openFilePicker}
                                        size="icon-sm"
                                        type="button"
                                        variant="outline"
                                    >
                                        {upload.isUploading ? <Loader2 className="animate-spin" /> : <FolderUp />}
                                    </Button>
                                </span>
                            </TooltipTrigger>
                            <TooltipContent>Upload files</TooltipContent>
                        </Tooltip>
                    </div>
                </Form>

                <FileManager
                    actions={fileManagerActions}
                    className="min-h-0 flex-1"
                    emptyState={noResourcesState}
                    files={fileNodes}
                    isLoading={isInitialLoading}
                    onBulkDelete={deletion.deleteBulk}
                    onMoveItems={handleMoveItems}
                    onOpen={handleOpenFile}
                    search={{ emptyState: noMatchesState, query: search.debouncedQuery }}
                />

                <ResourcesMkdirDialog
                    isOpen={isMkdirOpen}
                    onClose={() => setIsMkdirOpen(false)}
                />

                <ResourcesMoveDialog
                    file={fileToRename}
                    onClose={() => setFileToRename(null)}
                />

                <ResourcesCopyDialog
                    file={fileToCopy}
                    onClose={() => setFileToCopy(null)}
                />

                <ResourcesConflictDialog
                    conflicts={moveAction.pendingConflicts}
                    onCancel={moveAction.cancelConflicts}
                    onReplaceAll={moveAction.resolveConflicts}
                />

                <ConfirmationDialog
                    confirmText="Delete"
                    handleConfirm={deletion.confirmDelete}
                    handleOpenChange={handleDeleteDialogOpenChange}
                    isOpen={!!deletion.fileToDelete}
                    itemName={deletion.fileToDelete?.name}
                    itemType={deletion.fileToDelete?.isDir ? 'directory' : 'resource'}
                    title={deletion.fileToDelete?.isDir ? 'Delete Directory' : 'Delete Resource'}
                />
            </div>
        </>
    );
};

export default Resources;
