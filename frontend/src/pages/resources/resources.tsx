import {
    ColumnsSettings,
    Copy,
    FileSymlink,
    Folder,
    FolderPlus,
    FolderUp,
    Loader2,
    Search,
    Upload,
    X,
} from 'lucide-react';
import { useCallback, useMemo, useState } from 'react';
import { toast } from 'sonner';

import type { OverwriteConflict } from '@/components/shared/overwrite';

import {
    AppHeader,
    AppHeaderAction,
    AppHeaderActions,
    AppHeaderContent,
    AppHeaderTitle,
} from '@/components/layouts/app/app-header';
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
    type FileManagerLabels,
    type FileNode,
    formatModifiedAbsolute,
    formatModifiedRelative,
} from '@/components/shared/file-manager';
import { OverwriteDialog, useOverwrite } from '@/components/shared/overwrite';
import { Button } from '@/components/ui/button';
import {
    DropdownMenu,
    DropdownMenuCheckboxItem,
    DropdownMenuContent,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';
import { FileDropZone } from '@/components/ui/file-drop-zone';
import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupInput } from '@/components/ui/input-group';
import { ResourcesCopyDialog } from '@/features/resources/resources-copy-dialog';
import { ResourcesMkdirDialog } from '@/features/resources/resources-mkdir-dialog';
import { ResourcesMoveDialog } from '@/features/resources/resources-move-dialog';
import { buildResourcesDownloadHref, pluralizeItems, toFileNode } from '@/features/resources/resources-utils';
import { useResourcesDelete } from '@/features/resources/use-resources-delete';
import { useResourcesMove } from '@/features/resources/use-resources-move';
import { useResourcesSearch } from '@/features/resources/use-resources-search';
import { useResourcesUpload } from '@/features/resources/use-resources-upload';
import { useEffectAfterMount } from '@/hooks/use-effect-after-mount';
import { useFilesDragAndDrop } from '@/hooks/use-files-drag-and-drop';
import { usePageStorageKeys } from '@/hooks/use-page-storage-keys';
import { copyToClipboard } from '@/lib/report';
import { migrateLegacyViewOptions, saveViewOptions } from '@/lib/view-options-storage';
import { useResources } from '@/providers/resources-provider';

/**
 * Per-page persisted toggles for FileManager view options:
 *   - `size` / `modified`    — optional column visibility
 *   - `foldersFirst`         — whether directories cluster above files at every
 *                              level when a sort is active
 *   - `isModifiedRelative`   — render Modified as a relative label ("5m ago",
 *                              the FileManager default) when `true`, or as an
 *                              absolute, minute-precision timestamp when `false`
 *
 * All flags persist into the page's `viewOptions` storage bucket; the schema
 * is `Record<string, boolean>` so adding more toggles later does not require
 * a new key.
 */
interface ResourcesViewOptions {
    foldersFirst: boolean;
    isModifiedRelative: boolean;
    modified: boolean;
    size: boolean;
}

const RESOURCES_PATH = '/resources';

/** Defaults match FileManager's out-of-the-box behaviour (relative dates, folders first, both columns visible). */
const defaultViewOptions: ResourcesViewOptions = {
    foldersFirst: true,
    isModifiedRelative: true,
    modified: true,
    size: true,
};

type ResourcesViewOptionKey = keyof ResourcesViewOptions;

const seedViewOptions = (storageKey: string): ResourcesViewOptions => {
    const stored = migrateLegacyViewOptions(RESOURCES_PATH, storageKey);

    return {
        foldersFirst: stored.foldersFirst ?? defaultViewOptions.foldersFirst,
        isModifiedRelative: stored.isModifiedRelative ?? defaultViewOptions.isModifiedRelative,
        modified: stored.modified ?? defaultViewOptions.modified,
        size: stored.size ?? defaultViewOptions.size,
    };
};

function Resources() {
    const { isInitialLoading, resources } = useResources();
    const search = useResourcesSearch();

    const [isMkdirOpen, setIsMkdirOpen] = useState(false);
    // When the user invokes "New folder here" from a directory row's menu, we
    // need the dialog to seed itself with that *specific* directory. Cleared
    // whenever the dialog closes so the toolbar mkdir falls back to the
    // library root again.
    const [mkdirParentOverride, setMkdirParentOverride] = useState<null | string>(null);
    // Both dialogs accept an array now: a row-action click pushes a single-element
    // array, the bulk bar pushes the full deduped selection. Empty / null array
    // closes the dialog.
    const [filesToMove, setFilesToMove] = useState<FileNode[] | null>(null);
    const [filesToCopy, setFilesToCopy] = useState<FileNode[] | null>(null);

    const { viewOptions: viewOptionsStorageKey } = usePageStorageKeys();
    const [viewOptions, setViewOptions] = useState<ResourcesViewOptions>(() => seedViewOptions(viewOptionsStorageKey));

    useEffectAfterMount(() => {
        // Cast: `ResourcesViewOptions` is structurally a `Record<string, boolean>`
        // but TS doesn't widen object types with declared keys to an index
        // signature implicitly.
        saveViewOptions(viewOptionsStorageKey, viewOptions as unknown as Record<string, boolean>);
    }, [viewOptions, viewOptionsStorageKey]);

    const toggleViewOption = useCallback((option: ResourcesViewOptionKey) => {
        setViewOptions((previous) => ({ ...previous, [option]: !previous[option] }));
    }, []);

    // Keep the `labels` reference stable while the user keeps the same
    // formatting choice — FileManager threads it through the memoized row
    // bundle, and a fresh object identity on every render would invalidate
    // the row memo for the whole tree on unrelated re-renders.
    const fileManagerLabels = useMemo<FileManagerLabels>(
        () => ({
            formatModified: viewOptions.isModifiedRelative ? formatModifiedRelative : formatModifiedAbsolute,
        }),
        [viewOptions.isModifiedRelative],
    );

    // Toolbar / empty-area mkdir + upload always target the library root —
    // row-level "Upload here" / "New folder here" handlers carry their own
    // explicit path, and DnD passes the destination per drop, so these
    // entry points don't need a focus-derived fallback directory.
    const upload = useResourcesUpload();
    const deletion = useResourcesDelete();
    const { move } = useResourcesMove();

    const canAcceptDrop = !upload.isUploading;
    const { dragHandlers, isDragging } = useFilesDragAndDrop({
        canAcceptDrop,
        // Page-level drop falls back to the hook's defaults — i.e. the
        // library root, since `useResourcesUpload` is invoked without a
        // `defaultDir`.
        onDrop: upload.uploadFiles,
    });

    // Per-row external file drop (OS desktop → folder row): forward the
    // dropped files together with the resolved directory so the upload lands
    // exactly where the user released, not in the library root that the
    // page-level `dragHandlers` above default to. FileManager already stops
    // propagation on the row, so this never double-fires.
    const handleExternalFileDrop = useCallback(
        async (droppedFiles: File[], destinationDir: string): Promise<void> => {
            await upload.uploadFiles(droppedFiles, { dir: destinationDir });
        },
        [upload],
    );

    const fileNodes = useMemo<FileNode[]>(() => resources.map(toFileNode), [resources]);

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

    const dndMoveAction = useOverwrite<DndMovePlan>({
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

    // Join the selected paths with `\n` so the result pastes as a clean
    // newline-separated list into the agent chat, a shell command, or notes.
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
                icon: FolderPlus,
                id: 'resources-mkdir-here',
                label: 'New folder',
                onSelect: handleMkdirHere,
                separatorBefore: true,
            },
            {
                appliesToDirs: true,
                appliesToFiles: false,
                icon: Upload,
                id: 'resources-upload-here',
                label: 'Upload files',
                onSelect: handleUploadHere,
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
    // target the library root, identical to the toolbar buttons.
    const fileManagerEmptyAreaActions = useMemo<FileManagerEmptyAreaAction[]>(
        () => [
            {
                icon: FolderPlus,
                id: 'resources-empty-mkdir',
                label: 'New folder',
                onSelect: () => setIsMkdirOpen(true),
            },
            {
                icon: Upload,
                id: 'resources-empty-upload',
                label: 'Upload files',
                onSelect: upload.openFilePicker,
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
        <AppHeader>
            <AppHeaderContent>
                <AppHeaderTitle icon={<Folder className="size-4 shrink-0" />}>Resources</AppHeaderTitle>
            </AppHeaderContent>
            <AppHeaderActions>
                <AppHeaderAction
                    disabled={upload.isUploading}
                    icon={<FolderPlus />}
                    label="New folder"
                    onClick={() => setIsMkdirOpen(true)}
                    variant="outline"
                />
                <AppHeaderAction
                    aria-label={upload.isUploading ? 'Uploading...' : 'Upload files'}
                    disabled={upload.isUploading}
                    icon={upload.isUploading ? <Loader2 className="animate-spin" /> : <Upload />}
                    label={upload.isUploading ? 'Uploading...' : 'Upload files'}
                    onClick={upload.openFilePicker}
                    variant="secondary"
                />
            </AppHeaderActions>
        </AppHeader>
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
                className="relative flex h-[calc(100dvh-3rem)] flex-col gap-4 p-4"
                {...dragHandlers}
            >
                <input
                    aria-hidden="true"
                    className="hidden"
                    key={upload.fileInputKey}
                    multiple
                    name="resource-upload"
                    tabIndex={-1}
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

                <div className="flex items-center gap-2">
                    <InputGroup className="max-w-sm flex-1">
                        <InputGroupInput
                            aria-label="Search resources"
                            autoComplete="off"
                            onChange={(event) => search.setQuery(event.target.value)}
                            placeholder="Search resources..."
                            type="text"
                            value={search.rawQuery}
                        />
                        {search.rawQuery ? (
                            <InputGroupAddon align="inline-end">
                                <InputGroupButton
                                    onClick={search.resetSearch}
                                    type="button"
                                >
                                    <X />
                                </InputGroupButton>
                            </InputGroupAddon>
                        ) : null}
                    </InputGroup>
                    <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                            <Button
                                aria-label="Column settings"
                                className="ml-auto"
                                size="icon"
                                variant="outline"
                            >
                                <ColumnsSettings />
                            </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                            <DropdownMenuCheckboxItem
                                checked={viewOptions.size}
                                onCheckedChange={() => toggleViewOption('size')}
                                onSelect={(event) => event.preventDefault()}
                            >
                                Size
                            </DropdownMenuCheckboxItem>
                            <DropdownMenuCheckboxItem
                                checked={viewOptions.modified}
                                onCheckedChange={() => toggleViewOption('modified')}
                                onSelect={(event) => event.preventDefault()}
                            >
                                Modified
                            </DropdownMenuCheckboxItem>
                            <DropdownMenuSeparator />
                            <DropdownMenuCheckboxItem
                                checked={viewOptions.foldersFirst}
                                onCheckedChange={() => toggleViewOption('foldersFirst')}
                                onSelect={(event) => event.preventDefault()}
                            >
                                Folders first
                            </DropdownMenuCheckboxItem>
                            <DropdownMenuCheckboxItem
                                checked={viewOptions.isModifiedRelative}
                                disabled={!viewOptions.modified}
                                onCheckedChange={() => toggleViewOption('isModifiedRelative')}
                                onSelect={(event) => event.preventDefault()}
                            >
                                Relative dates
                            </DropdownMenuCheckboxItem>
                        </DropdownMenuContent>
                    </DropdownMenu>
                </div>

                <FileManager
                    actions={fileManagerActions}
                    bulkActions={fileManagerBulkActions}
                    className="min-h-0 flex-1"
                    columns={{
                        isModifiedVisible: viewOptions.modified,
                        isSizeVisible: viewOptions.size,
                    }}
                    emptyAreaActions={fileManagerEmptyAreaActions}
                    emptyState={noResourcesState}
                    files={fileNodes}
                    isFoldersFirst={viewOptions.foldersFirst}
                    isLoading={isInitialLoading}
                    labels={fileManagerLabels}
                    onExternalFileDrop={handleExternalFileDrop}
                    onMoveItems={handleMoveItems}
                    onOpen={handleOpenFile}
                    search={{ emptyState: noMatchesState, query: search.debouncedQuery }}
                />

                <ResourcesMkdirDialog
                    defaultParentPath={mkdirParentOverride ?? ''}
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

                <OverwriteDialog
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
                    title={deletion.fileToDelete?.isDir ? 'Delete directory' : 'Delete resource'}
                />
            </div>
        </>
    );
}

export default Resources;
