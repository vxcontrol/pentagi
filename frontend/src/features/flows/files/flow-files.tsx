import { ArrowDownToLine, FolderInput, FolderOutput, FolderUp, Loader2, Search, Upload, X } from 'lucide-react';
import { useCallback, useMemo, useState } from 'react';
import { toast } from 'sonner';

import ConfirmationDialog from '@/components/shared/confirmation-dialog';
import {
    bulkCopyPathsAction,
    bulkDeleteAction,
    bulkDownloadAction,
    bulkPromoteAction,
    copyPathAction,
    deleteAction,
    downloadAction,
    FileManager,
    type FileManagerAction,
    type FileManagerBulkAction,
    type FileNode,
} from '@/components/shared/file-manager';
import { Button } from '@/components/ui/button';
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';
import { Form, FormControl, FormField, FormItem } from '@/components/ui/form';
import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupInput } from '@/components/ui/input-group';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { StatusType } from '@/graphql/types';
import { useFilesDragAndDrop } from '@/hooks/use-files-drag-and-drop';
import { copyToClipboard } from '@/lib/report';
import { useFlow } from '@/providers/flow-provider';

import { FlowFilesAttachResourcesDialog } from './flow-files-attach-resources-dialog';
import { ROOT_GROUPS } from './flow-files-constants';
import { FlowFilesPromoteDialog } from './flow-files-promote-dialog';
import { FlowFilesPullDialog } from './flow-files-pull-dialog';
import { buildFlowFilesDownloadHref, pluralizeItems } from './flow-files-utils';
import { useFlowFilesData } from './use-flow-files-data';
import { useFlowFilesDelete } from './use-flow-files-delete';
import { useFlowFilesRealtime } from './use-flow-files-realtime';
import { useFlowFilesSearch } from './use-flow-files-search';
import { useFlowFilesUpload } from './use-flow-files-upload';

function FlowFiles() {
    const { flowId, flowStatus } = useFlow();
    const [isPullDialogOpen, setIsPullDialogOpen] = useState(false);
    const [isAttachResourcesDialogOpen, setIsAttachResourcesDialogOpen] = useState(false);
    const [filesToPromote, setFilesToPromote] = useState<FileNode[] | null>(null);

    const { fileNodes, isInitialLoading, isLoading } = useFlowFilesData({ flowId });

    useFlowFilesRealtime({ flowId, isPaused: isLoading });

    const search = useFlowFilesSearch();
    const upload = useFlowFilesUpload({ flowId });
    const deletion = useFlowFilesDelete({ flowId });

    const canAcceptDrop = !!flowId && !upload.isUploading;
    const { dragHandlers, isDragging } = useFilesDragAndDrop({
        canAcceptDrop,
        onDrop: upload.uploadFiles,
    });

    const isContainerRunning = flowStatus === StatusType.Running || flowStatus === StatusType.Waiting;
    const isPullDisabled = !isContainerRunning || isLoading || upload.isUploading;

    const handleCopyPath = useCallback(async (file: FileNode) => {
        const wasCopied = await copyToClipboard(file.path);

        if (wasCopied) {
            toast.success('Path copied to clipboard');

            return;
        }

        toast.error('Failed to copy path');
    }, []);

    /**
     * Join the selected paths with `\n` so the result pastes as a clean
     * newline-separated list into the agent chat, a shell command, or a tool argument.
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

    // `flowId` may be missing (no flow selected yet) — return '' so FileManager
    // renders a noop link instead of crashing on `null`.
    const getRowDownloadHref = useCallback(
        (file: FileNode): string => buildFlowFilesDownloadHref(flowId, [file]) ?? '',
        [flowId],
    );

    const getBulkDownloadHref = useCallback(
        (files: FileNode[]): string => buildFlowFilesDownloadHref(flowId, files) ?? '',
        [flowId],
    );

    const handleRequestPromote = useCallback((file: FileNode) => {
        setFilesToPromote([file]);
    }, []);

    const handleClosePromoteDialog = useCallback(() => setFilesToPromote(null), []);

    const promoteAction = useMemo<FileManagerAction>(
        () => ({
            appliesToDirs: true,
            icon: FolderOutput,
            id: 'flow-files-save-as-resource',
            label: 'Save as resource',
            onSelect: handleRequestPromote,
        }),
        [handleRequestPromote],
    );

    const fileManagerActions = useMemo<FileManagerAction[]>(
        () => [
            downloadAction(getRowDownloadHref),
            copyPathAction(handleCopyPath),
            promoteAction,
            deleteAction(deletion.requestDelete),
        ],
        [getRowDownloadHref, handleCopyPath, promoteAction, deletion.requestDelete],
    );

    const fileManagerBulkActions = useMemo<FileManagerBulkAction[]>(
        () => [
            bulkDownloadAction(getBulkDownloadHref),
            bulkPromoteAction((files) => setFilesToPromote(files)),
            bulkCopyPathsAction(handleBulkCopyPaths),
            bulkDeleteAction(deletion.deleteFiles),
        ],
        [deletion.deleteFiles, getBulkDownloadHref, handleBulkCopyPaths],
    );

    const handleOpenPullDialog = useCallback(() => setIsPullDialogOpen(true), []);
    const handleClosePullDialog = useCallback(() => setIsPullDialogOpen(false), []);
    const handleOpenAttachResourcesDialog = useCallback(() => setIsAttachResourcesDialogOpen(true), []);
    const handleCloseAttachResourcesDialog = useCallback(() => setIsAttachResourcesDialogOpen(false), []);
    const handleDeleteDialogOpenChange = useCallback(
        (nextOpen: boolean) => {
            if (!nextOpen) {
                deletion.clearFileToDelete();
            }
        },
        [deletion],
    );

    const isAttachResourcesDisabled = !flowId || isLoading || upload.isUploading;

    const noFilesState = (
        <Empty>
            <EmptyHeader>
                <EmptyMedia variant="icon">
                    <FolderUp />
                </EmptyMedia>
                <EmptyTitle>No files in cache</EmptyTitle>
                <EmptyDescription>
                    Upload files to make them available at <code>/work/uploads</code>, or use Pull to sync files from
                    the running container. You can also drag &amp; drop files here.
                </EmptyDescription>
            </EmptyHeader>
        </Empty>
    );

    const noMatchesState = (
        <Empty>
            <EmptyHeader>
                <EmptyMedia variant="icon">
                    <Search />
                </EmptyMedia>
                <EmptyTitle>No matches</EmptyTitle>
                <EmptyDescription>
                    No files match <code>{search.debouncedQuery.trim()}</code>. Try a different query.
                </EmptyDescription>
            </EmptyHeader>
        </Empty>
    );

    return (
        <div
            className="relative flex h-full flex-col"
            {...dragHandlers}
        >
            <input
                aria-hidden="true"
                className="hidden"
                key={upload.fileInputKey}
                multiple
                name="flow-file-upload"
                tabIndex={-1}
                type="file"
                {...upload.fileInputProps}
            />

            {isDragging && (
                <div className="bg-primary/10 border-primary pointer-events-none absolute inset-0 z-30 flex items-center justify-center rounded-lg border-2 border-dashed">
                    <div className="text-primary flex flex-col items-center gap-2">
                        <FolderUp className="size-8" />
                        <span className="text-sm font-medium">Drop files to upload</span>
                    </div>
                </div>
            )}

            <div className="bg-background sticky top-0 z-10 pb-4">
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
                                                placeholder="Search files..."
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
                                        disabled={upload.isUploading || isLoading}
                                        onClick={upload.openFilePicker}
                                        size="icon-sm"
                                        type="button"
                                        variant="outline"
                                    >
                                        {upload.isUploading ? <Loader2 className="animate-spin" /> : <Upload />}
                                    </Button>
                                </span>
                            </TooltipTrigger>
                            <TooltipContent className="max-w-64 text-center text-xs">
                                <p className="font-medium">Upload files</p>
                                <p className="mt-1">
                                    Pushed to <code>/work/uploads</code> — immediately accessible inside the container.
                                </p>
                            </TooltipContent>
                        </Tooltip>

                        <Tooltip>
                            <TooltipTrigger asChild>
                                <span>
                                    <Button
                                        disabled={isAttachResourcesDisabled}
                                        onClick={handleOpenAttachResourcesDialog}
                                        size="icon-sm"
                                        type="button"
                                        variant="outline"
                                    >
                                        <FolderInput />
                                    </Button>
                                </span>
                            </TooltipTrigger>
                            <TooltipContent className="max-w-64 text-center text-xs">
                                <p className="font-medium">Attach resources</p>
                                <p className="mt-1">
                                    Copied from the library to <code>/work/resources</code> — immediately accessible
                                    inside the container.
                                </p>
                            </TooltipContent>
                        </Tooltip>

                        <Tooltip>
                            <TooltipTrigger asChild>
                                <span>
                                    <Button
                                        disabled={isPullDisabled}
                                        onClick={handleOpenPullDialog}
                                        size="icon-sm"
                                        type="button"
                                        variant="outline"
                                    >
                                        <ArrowDownToLine />
                                    </Button>
                                </span>
                            </TooltipTrigger>
                            <TooltipContent className="max-w-64 text-center text-xs">
                                {isContainerRunning ? (
                                    <>
                                        <p className="font-medium">Pull file or directory from container</p>
                                        <p className="mt-1">
                                            Snapshots are stored separately under <strong>Container</strong>.
                                        </p>
                                    </>
                                ) : (
                                    <p className="font-medium">Container is not running</p>
                                )}
                            </TooltipContent>
                        </Tooltip>
                    </div>
                </Form>
            </div>

            <FileManager
                actions={fileManagerActions}
                bulkActions={fileManagerBulkActions}
                className="min-h-0 flex-1"
                emptyState={noFilesState}
                files={fileNodes}
                isLoading={isInitialLoading}
                rootGroups={ROOT_GROUPS}
                search={{ emptyState: noMatchesState, query: search.debouncedQuery }}
            />

            <FlowFilesPullDialog
                cachedFiles={fileNodes}
                flowId={flowId}
                isOpen={isPullDialogOpen}
                onClose={handleClosePullDialog}
            />

            <FlowFilesAttachResourcesDialog
                cachedFiles={fileNodes}
                flowId={flowId}
                isOpen={isAttachResourcesDialogOpen}
                onClose={handleCloseAttachResourcesDialog}
            />

            <FlowFilesPromoteDialog
                files={filesToPromote}
                flowId={flowId}
                onClose={handleClosePromoteDialog}
            />

            <ConfirmationDialog
                confirmText="Delete"
                handleConfirm={deletion.confirmDelete}
                handleOpenChange={handleDeleteDialogOpenChange}
                isOpen={!!deletion.fileToDelete}
                itemName={deletion.fileToDelete?.name}
                itemType={deletion.fileToDelete?.isDir ? 'directory' : 'file'}
                title={deletion.fileToDelete?.isDir ? 'Delete Directory' : 'Delete File'}
            />
        </div>
    );
}

export default FlowFiles;
