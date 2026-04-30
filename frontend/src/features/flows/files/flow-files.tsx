import { ArrowDownToLine, FolderUp, Info, Loader2, Search, X } from 'lucide-react';
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
import { Button } from '@/components/ui/button';
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';
import { Form, FormControl, FormField, FormItem } from '@/components/ui/form';
import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupInput } from '@/components/ui/input-group';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { StatusType } from '@/graphql/types';
import { copyToClipboard } from '@/lib/report';
import { useFlow } from '@/providers/flow-provider';

import { ROOT_GROUPS } from './flow-files-constants';
import { FlowFilesPullDialog } from './flow-files-pull-dialog';
import { buildDownloadHref } from './flow-files-utils';
import { useFilesDragAndDrop } from './use-files-drag-and-drop';
import { useFlowFilesData } from './use-flow-files-data';
import { useFlowFilesDelete } from './use-flow-files-delete';
import { useFlowFilesRealtime } from './use-flow-files-realtime';
import { useFlowFilesSearch } from './use-flow-files-search';
import { useFlowFilesUpload } from './use-flow-files-upload';

interface FlowFilesContentProps {
    flowId: null | string;
    flowStatus: StatusType | undefined;
}

/**
 * Holds every piece of local state for a single flow. The outer `<FlowFiles />`
 * remounts this component via `key={flowId}` so switching flows is a fresh mount
 * with no leftover form values, drag overlays or pending toasts.
 */
const FlowFilesContent = ({ flowId, flowStatus }: FlowFilesContentProps) => {
    const [isPullDialogOpen, setIsPullDialogOpen] = useState(false);

    const { fileNodes, isInitialLoading, isLoading, refetchFiles } = useFlowFilesData({ flowId });

    useFlowFilesRealtime({ flowId, isPaused: isLoading });

    const search = useFlowFilesSearch();
    const upload = useFlowFilesUpload({ flowId, refetchFiles });
    const deletion = useFlowFilesDelete({ flowId, refetchFiles });

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

    const getDownloadHref = useCallback(
        (file: FileNode): string => buildDownloadHref(flowId, file) ?? '',
        [flowId],
    );

    const fileManagerActions = useMemo<FileManagerAction[]>(
        () => [downloadAction(getDownloadHref), copyPathAction(handleCopyPath), deleteAction(deletion.requestDelete)],
        [getDownloadHref, handleCopyPath, deletion.requestDelete],
    );

    const handleOpenPullDialog = useCallback(() => setIsPullDialogOpen(true), []);
    const handleClosePullDialog = useCallback(() => setIsPullDialogOpen(false), []);
    const handleDeleteDialogOpenChange = useCallback(
        (nextOpen: boolean) => {
            if (!nextOpen) {
                deletion.clearFileToDelete();
            }
        },
        [deletion],
    );

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
                className="hidden"
                key={upload.fileInputKey}
                multiple
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
                                <Button
                                    size="icon-sm"
                                    type="button"
                                    variant="ghost"
                                >
                                    <Info className="text-muted-foreground size-4" />
                                </Button>
                            </TooltipTrigger>
                            <TooltipContent className="max-w-64 text-center text-xs">
                                <p>
                                    <strong>Uploads</strong> are pushed to <code>/work/uploads</code> and are
                                    immediately accessible inside the container.
                                </p>
                                <p className="mt-1">
                                    <strong>Container</strong> files are snapshots pulled via Pull and stored
                                    separately.
                                </p>
                            </TooltipContent>
                        </Tooltip>

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
                                        {upload.isUploading ? <Loader2 className="animate-spin" /> : <FolderUp />}
                                    </Button>
                                </span>
                            </TooltipTrigger>
                            <TooltipContent>Upload files</TooltipContent>
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
                            <TooltipContent>
                                {isContainerRunning
                                    ? 'Pull file or directory from container'
                                    : 'Container is not running'}
                            </TooltipContent>
                        </Tooltip>
                    </div>
                </Form>
            </div>

            <FileManager
                actions={fileManagerActions}
                className="min-h-0 flex-1"
                emptyState={noFilesState}
                files={fileNodes}
                isLoading={isInitialLoading}
                onBulkDelete={deletion.deleteBulk}
                rootGroups={ROOT_GROUPS}
                search={{ emptyState: noMatchesState, query: search.debouncedQuery }}
            />

            <FlowFilesPullDialog
                flowId={flowId}
                isOpen={isPullDialogOpen}
                onClose={handleClosePullDialog}
                onSuccess={refetchFiles}
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
};

const FlowFiles = () => {
    const { flowId, flowStatus } = useFlow();

    return (
        <FlowFilesContent
            flowId={flowId}
            flowStatus={flowStatus}
            key={flowId ?? 'no-flow'}
        />
    );
};

export default FlowFiles;
