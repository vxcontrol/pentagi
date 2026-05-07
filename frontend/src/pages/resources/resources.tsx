import { Copy, FileSymlink, Folder, FolderPlus, FolderUp, Loader2, Search, X } from 'lucide-react';
import { useCallback, useMemo, useState } from 'react';
import { toast } from 'sonner';

import type { OverwriteConflict } from '@/components/shared/overwrite-confirm-dialog';

import ConfirmationDialog from '@/components/shared/confirmation-dialog';
import {
    bulkCopyAction,
    bulkCopyPathsAction,
    bulkDeleteAction,
    bulkDownloadAction,
    bulkMoveAction,
    copyPathAction,
    deleteAction,
    downloadAction,
    FileManager,
    type FileManagerAction,
    type FileManagerBulkAction,
    type FileManagerEmptyAreaAction,
    type FileNode,
} from '@/components/shared/file-manager';
import { OverwriteConfirmDialog } from '@/components/shared/overwrite-confirm-dialog';
import { useOverwriteAction } from '@/components/shared/use-overwrite-action';
import { Breadcrumb, BreadcrumbItem, BreadcrumbList, BreadcrumbPage } from '@/components/ui/breadcrumb';
import { Button } from '@/components/ui/button';
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';
import { FileDropZone } from '@/components/ui/file-drop-zone';
import { Form, FormControl, FormField, FormItem } from '@/components/ui/form';
import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupInput } from '@/components/ui/input-group';
import { Separator } from '@/components/ui/separator';
import { SidebarTrigger } from '@/components/ui/sidebar';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { ResourcesCopyDialog } from '@/features/resources/resources-copy-dialog';
import { ResourcesMkdirDialog } from '@/features/resources/resources-mkdir-dialog';
import { ResourcesMoveDialog } from '@/features/resources/resources-move-dialog';
import { buildResourcesDownloadHref, pluralizeItems, toFileNode } from '@/features/resources/resources-utils';
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
    // When the user invokes "New folder here" from a directory row's menu, we
    // need the dialog to seed itself with that *specific* directory rather
    // than the focus-derived `currentDir`. Cleared whenever the dialog closes
    // so the toolbar mkdir falls back to the focus default again.
    const [mkdirParentOverride, setMkdirParentOverride] = useState<null | string>(null);
    // Both dialogs accept an array now: a row-action click pushes a single-element
    // array, the bulk bar pushes the full deduped selection. Empty / null array
    // closes the dialog.
    const [filesToMove, setFilesToMove] = useState<FileNode[] | null>(null);
    const [filesToCopy, setFilesToCopy] = useState<FileNode[] | null>(null);
    // The path of the row currently focused in the FileManager (roving tabindex).
    // `null` until the user has actually clicked / Tab-ed into a row, so that we
    // can distinguish "user explicitly chose a directory" from the auto-fallback
    // the FileManager uses for its keyboard handler.
    const [activeRowPath, setActiveRowPath] = useState<null | string>(null);

    // Index resources by virtual path so deriving `currentDir` (and other
    // path-based lookups) stays O(1) regardless of library size.
    const resourcesByPath = useMemo(() => {
        const map = new Map<string, (typeof resources)[number]>();

        for (const resource of resources) {
            map.set(resource.path, resource);
        }

        return map;
    }, [resources]);

    /**
     * Virtual directory the next upload / mkdir / drop should target. Resolved
     * from `activeRowPath`:
     *   - directory row    → the directory's own path (upload lands inside it)
     *   - file row         → the file's parent directory (sibling upload)
     *   - no focused row   → '' (library root)
     *   - stale path       → '' (focused row was deleted by a subscription)
     *
     * Mirrors the same resolution rule that the FileManager's internal
     * drag-and-drop uses (`resolveDropTargetDir` in `use-file-manager-dnd`),
     * so context-aware uploads land where users would expect a file dropped
     * onto the same row to land.
     */
    const currentDir = useMemo(() => {
        if (!activeRowPath) {
            return '';
        }

        const resource = resourcesByPath.get(activeRowPath);

        if (!resource) {
            return '';
        }

        if (resource.isDir) {
            return resource.path;
        }

        const idx = resource.path.lastIndexOf('/');

        return idx === -1 ? '' : resource.path.slice(0, idx);
    }, [activeRowPath, resourcesByPath]);

    const upload = useResourcesUpload({ defaultDir: currentDir });
    const deletion = useResourcesDelete();
    const { move } = useResourcesMove();

    const canAcceptDrop = !upload.isUploading;
    const { dragHandlers, isDragging } = useFilesDragAndDrop({
        canAcceptDrop,
        // `upload.uploadFiles` is reference-stable: it reads the latest
        // `defaultDir` through a ref, so changing the focused row does not
        // invalidate the drag handlers below.
        onDrop: upload.uploadFiles,
    });

    // Per-row external file drop (OS desktop → folder row): forward the
    // dropped files together with the resolved directory so the upload lands
    // exactly where the user released, not in the focus-derived `currentDir`.
    // FileManager already stops propagation on the row, so this never
    // double-fires alongside the page-level `dragHandlers` above.
    const handleExternalFileDrop = useCallback(
        async (droppedFiles: File[], destinationDir: string): Promise<void> => {
            await upload.uploadFiles(droppedFiles, { dir: destinationDir });
        },
        [upload],
    );

    const fileNodes = useMemo<FileNode[]>(() => resources.map(toFileNode), [resources]);

    // Snapshot of every existing path in the library — drives the local
    // preflight for the drag-and-drop move workflow.
    const resourcePaths = useMemo(() => new Set(resources.map((resource) => resource.path)), [resources]);

    /**
     * Drag-and-drop move: ship every dragged item to `destinationDir` in a
     * single atomic batch (`PUT /resources/move` with `sources[]`). The
     * backend handles dedup + transactional writes; the shared overwrite
     * workflow drives the local preflight and the conflict dialog.
     */
    interface DndMovePlan {
        destination: string;
        sources: readonly string[];
        targets: OverwriteConflict[];
    }

    const dndMoveAction = useOverwriteAction<DndMovePlan>({
        execute: (plan, force) => move(plan.sources, plan.destination, force),
        findConflicts: (plan) => {
            const movedPaths = new Set(plan.sources);

            // Targets that match an item being moved are no-ops, not conflicts.
            return plan.targets.filter((t) => !movedPaths.has(t.destination) && resourcePaths.has(t.destination));
        },
        synthesizeFallbackConflicts: (plan) => plan.targets,
    });

    const handleMoveItems = useCallback(
        async (sources: FileNode[], destinationDir: string) => {
            if (sources.length === 0) {
                return;
            }

            const baseDir = destinationDir.replace(/\/+$/, '');
            const targets: OverwriteConflict[] = sources.map((source) => ({
                destination: baseDir ? `${baseDir}/${source.name}` : source.name,
                destinationName: source.name,
            }));

            await dndMoveAction.primaryExecute({
                destination: baseDir,
                sources: sources.map((source) => source.path),
                targets,
            });
        },
        [dndMoveAction],
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
     * Bulk "copy paths" handler: join every selected file's path with `\n` so the
     * user can paste a clean newline-separated list straight into the agent chat,
     * a shell command, or notes. Reports the count for clarity — silent failures
     * confuse users when the clipboard happens to already contain the same text.
     */
    const handleBulkCopyPaths = useCallback(async (paths: string[]) => {
        if (paths.length === 0) {
            return;
        }

        const wasCopied = await copyToClipboard(paths.join('\n'));

        if (wasCopied) {
            toast.success(`${paths.length} ${pluralizeItems(paths.length)} copied to clipboard`);

            return;
        }

        toast.error('Failed to copy paths');
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

        anchor.href = buildResourcesDownloadHref([file]);
        anchor.download = file.name;
        anchor.rel = 'noopener';
        document.body.appendChild(anchor);
        anchor.click();
        anchor.remove();
    }, []);

    // Row-level "Upload here" pre-targets the picker at the chosen directory's
    // path, regardless of `currentDir` (the dropdown trigger doesn't move
    // keyboard focus, so `activeRowPath` may still point at a sibling row).
    const handleUploadHere = useCallback(
        (file: FileNode) => {
            upload.openFilePickerForDir(file.path);
        },
        [upload],
    );

    const handleMkdirHere = useCallback((file: FileNode) => {
        setMkdirParentOverride(file.path);
        setIsMkdirOpen(true);
    }, []);

    const closeMkdirDialog = useCallback(() => {
        setIsMkdirOpen(false);
        setMkdirParentOverride(null);
    }, []);

    const fileManagerActions = useMemo<FileManagerAction[]>(
        () => [
            // Row download is the single-file specialisation of the bulk download:
            // we hand the URL builder a 1-element array so the same backend
            // contract (`?paths[]=`) is used everywhere.
            downloadAction((file) => buildResourcesDownloadHref([file])),
            copyPathAction(handleCopyPath),
            // Directory-only actions — surfaced both in the row dropdown and
            // the right-click context menu (the manager renders both menus
            // from the same `actions` array). `appliesToFiles: false` keeps
            // them off file rows.
            {
                appliesToDirs: true,
                appliesToFiles: false,
                icon: FolderUp,
                id: 'resources-upload-here',
                label: 'Upload files here',
                onSelect: handleUploadHere,
                separatorBefore: true,
            },
            {
                appliesToDirs: true,
                appliesToFiles: false,
                icon: FolderPlus,
                id: 'resources-mkdir-here',
                label: 'New folder here',
                onSelect: handleMkdirHere,
            },
            {
                appliesToDirs: true,
                icon: FileSymlink,
                id: 'resources-rename',
                label: 'Rename or move',
                onSelect: (file) => setFilesToMove([file]),
                separatorBefore: true,
            },
            {
                appliesToDirs: true,
                icon: Copy,
                id: 'resources-copy',
                label: 'Copy to…',
                onSelect: (file) => setFilesToCopy([file]),
            },
            deleteAction(deletion.requestDelete),
        ],
        [deletion.requestDelete, handleCopyPath, handleMkdirHere, handleUploadHere],
    );

    // Bulk-action set, rendered in the bulk-actions bar when at least one row
    // is selected. Order matters — primary CTAs first, then less frequent
    // actions in the overflow `…` menu, then destructive Delete on the right.
    const fileManagerBulkActions = useMemo<FileManagerBulkAction[]>(
        () => [
            bulkDownloadAction(buildResourcesDownloadHref),
            bulkMoveAction((files) => setFilesToMove(files)),
            bulkCopyAction((files) => setFilesToCopy(files), { overflow: true }),
            bulkCopyPathsAction(handleBulkCopyPaths),
            bulkDeleteAction(deletion.deleteFiles),
        ],
        [deletion.deleteFiles, handleBulkCopyPaths],
    );

    // Right-click anywhere outside a row in the tree → mirror the toolbar
    // gestures so users have a closer-to-pointer entry point. Both items
    // resolve through the same focus-derived `currentDir` as the toolbar
    // buttons (via `defaultDir` / `mkdirParentOverride === null`), so the
    // outcome — and the toolbar tooltip telling the user *where* it lands —
    // stays identical no matter which surface the user invokes.
    const fileManagerEmptyAreaActions = useMemo<FileManagerEmptyAreaAction[]>(
        () => [
            {
                icon: FolderUp,
                id: 'resources-empty-upload',
                label: 'Upload files',
                onSelect: upload.openFilePicker,
            },
            {
                icon: FolderPlus,
                id: 'resources-empty-mkdir',
                label: 'New folder',
                onSelect: () => setIsMkdirOpen(true),
            },
        ],
        [upload.openFilePicker],
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

    // Human-readable target for the toolbar tooltips: matches the same wording
    // used in the upload success toast so users see the same "to /reports/2025"
    // / "to your library" phrasing both before and after the action.
    const uploadTargetLabel = currentDir ? `/${currentDir}` : 'library root';

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
                            <TooltipContent>Create directory in {uploadTargetLabel}</TooltipContent>
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
                            <TooltipContent>Upload files to {uploadTargetLabel}</TooltipContent>
                        </Tooltip>
                    </div>
                </Form>

                <FileManager
                    actions={fileManagerActions}
                    bulkActions={fileManagerBulkActions}
                    className="min-h-0 flex-1"
                    emptyAreaActions={fileManagerEmptyAreaActions}
                    emptyState={noResourcesState}
                    files={fileNodes}
                    isLoading={isInitialLoading}
                    onActiveRowChange={setActiveRowPath}
                    onExternalFileDrop={handleExternalFileDrop}
                    onMoveItems={handleMoveItems}
                    onOpen={handleOpenFile}
                    search={{ emptyState: noMatchesState, query: search.debouncedQuery }}
                />

                <ResourcesMkdirDialog
                    defaultParentPath={mkdirParentOverride ?? currentDir}
                    isOpen={isMkdirOpen}
                    onClose={closeMkdirDialog}
                />

                <ResourcesMoveDialog
                    files={filesToMove}
                    onClose={() => setFilesToMove(null)}
                />

                <ResourcesCopyDialog
                    files={filesToCopy}
                    onClose={() => setFilesToCopy(null)}
                />

                <OverwriteConfirmDialog
                    conflicts={dndMoveAction.conflicts}
                    onCancel={dndMoveAction.resetConflicts}
                    onReplaceAll={dndMoveAction.handleReplaceAll}
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
