import { zodResolver } from '@hookform/resolvers/zod';
import debounce from 'lodash/debounce';
import { ArrowDownToLine, FolderUp, HardDrive, Info, Loader2, Search, X } from 'lucide-react';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useForm } from 'react-hook-form';
import { toast } from 'sonner';
import { z } from 'zod';

import {
    copyPathAction,
    deleteAction,
    downloadAction,
    FileManager,
    type FileManagerAction,
    type FileManagerRootGroup,
    type FileNode,
} from '@/components/file-manager';
import ConfirmationDialog from '@/components/shared/confirmation-dialog';
import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';
import { Form, FormControl, FormField } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupInput } from '@/components/ui/input-group';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import {
    type FlowFileFragmentFragment,
    StatusType,
    useFlowFileAddedSubscription,
    useFlowFileDeletedSubscription,
    useFlowFilesQuery,
    useFlowFileUpdatedSubscription,
} from '@/graphql/types';
import { api, getApiErrorMessage, unwrapApiResponse } from '@/lib/axios';
import { copyToClipboard } from '@/lib/report';
import { baseUrl } from '@/models/api';
import { useFlow } from '@/providers/flow-provider';

type FlowFile = FlowFileFragmentFragment;

interface FlowFilesResponse {
    files: Array<FlowFile>;
    total: number;
}

const searchFormSchema = z.object({
    search: z.string(),
});

const buildDownloadHref = (flowId: null | string, file: FileNode) =>
    `${baseUrl}/flows/${flowId}/files/download?path=${encodeURIComponent(file.path)}`;

const toFileNode = (file: FlowFile): FileNode => ({
    id: file.id,
    isDir: file.isDir,
    modifiedAt: file.modifiedAt,
    name: file.name,
    path: file.path,
    size: file.size,
});

const ROOT_GROUPS: FileManagerRootGroup[] = [
    { defaultOpen: true, icon: FolderUp, id: 'uploads', label: 'Uploads', pathPrefix: 'uploads' },
    { defaultOpen: true, icon: HardDrive, id: 'container', label: 'Container', pathPrefix: 'container' },
];

// ── pull dialog ────────────────────────────────────────────────────────────────

interface PullDialogProps {
    flowId: null | string;
    onClose: () => void;
    onSuccess: () => void;
    open: boolean;
}

const PullDialog = ({ flowId, onClose, onSuccess, open }: PullDialogProps) => {
    const [containerPath, setContainerPath] = useState('');
    const [shouldOverwrite, setShouldOverwrite] = useState(false);
    const [isPulling, setIsPulling] = useState(false);

    useEffect(() => {
        if (open) {
            setContainerPath('');
            setShouldOverwrite(false);
        }
    }, [open]);

    const handlePull = useCallback(async () => {
        if (!flowId || !containerPath.trim()) {
            return;
        }

        setIsPulling(true);

        try {
            await api.post<FlowFilesResponse>(
                `/flows/${flowId}/files/pull`,
                {
                    force: shouldOverwrite,
                    path: containerPath.trim(),
                },
                // Copying a directory out of the container can take longer than the default 30s
                // (large logs, deep trees) — disable the per-call timeout entirely.
                { timeout: 0 },
            );
            toast.success('Pulled from container', {
                description: `Saved to local cache under container/`,
            });
            onSuccess();
            onClose();
        } catch (error) {
            const description = getApiErrorMessage(error, 'Failed to pull from container', {
                409: 'Entry already exists — enable "Overwrite" to replace it',
            });

            toast.error('Pull failed', { description });
        } finally {
            setIsPulling(false);
        }
    }, [flowId, containerPath, shouldOverwrite, onSuccess, onClose]);

    return (
        <Dialog
            onOpenChange={(isOpen) => {
                if (!isOpen && !isPulling) {
                    onClose();
                }
            }}
            open={open}
        >
            <DialogContent>
                <DialogHeader>
                    <DialogTitle className="flex items-center gap-2">
                        <ArrowDownToLine className="size-4" />
                        Pull from container
                    </DialogTitle>
                    <DialogDescription>
                        Enter a path inside the running container. The file or directory will be synced to the local
                        cache under <code>container/</code>.
                    </DialogDescription>
                </DialogHeader>

                <div className="flex flex-col gap-4">
                    <div className="flex flex-col gap-1.5">
                        <Label htmlFor="container-path">Container path</Label>
                        <Input
                            autoComplete="off"
                            autoFocus
                            disabled={isPulling}
                            id="container-path"
                            onChange={(event) => setContainerPath(event.target.value)}
                            onKeyDown={(event) => {
                                if (event.key === 'Enter' && containerPath.trim() && !isPulling) {
                                    void handlePull();
                                }
                            }}
                            placeholder="/etc/nginx/conf"
                            value={containerPath}
                        />
                    </div>

                    <div className="flex items-center gap-2">
                        <Switch
                            checked={shouldOverwrite}
                            disabled={isPulling}
                            id="force-pull"
                            onCheckedChange={setShouldOverwrite}
                        />
                        <Label
                            className="cursor-pointer font-normal"
                            htmlFor="force-pull"
                        >
                            Overwrite if already cached
                        </Label>
                    </div>
                </div>

                <div className="flex justify-end gap-2">
                    <Button
                        disabled={isPulling}
                        onClick={onClose}
                        variant="outline"
                    >
                        Cancel
                    </Button>
                    <Button
                        disabled={!containerPath.trim() || isPulling}
                        onClick={() => void handlePull()}
                    >
                        {isPulling ? <Loader2 className="animate-spin" /> : <ArrowDownToLine />}
                        Pull
                    </Button>
                </div>
            </DialogContent>
        </Dialog>
    );
};

// ── main component ─────────────────────────────────────────────────────────────

const FlowFiles = () => {
    const { flowId, flowStatus } = useFlow();
    const inputRef = useRef<HTMLInputElement | null>(null);
    const dragCounterRef = useRef(0);
    const [isUploading, setIsUploading] = useState(false);
    const [isDragging, setIsDragging] = useState(false);
    const [fileToDelete, setFileToDelete] = useState<FileNode | null>(null);
    const [showPullDialog, setShowPullDialog] = useState(false);
    const [debouncedSearch, setDebouncedSearch] = useState('');
    const flowFilesVariables = useMemo(() => ({ flowId: flowId ?? '' }), [flowId]);
    const {
        data: flowFilesData,
        error: flowFilesError,
        loading: isLoading,
        refetch: refetchFiles,
    } = useFlowFilesQuery({
        skip: !flowId,
        variables: flowFilesVariables,
    });

    const form = useForm<z.infer<typeof searchFormSchema>>({
        defaultValues: { search: '' },
        resolver: zodResolver(searchFormSchema),
    });

    const searchValue = form.watch('search');

    const debouncedUpdateSearch = useMemo(
        () =>
            debounce((value: string) => {
                setDebouncedSearch(value);
            }, 300),
        [],
    );

    useEffect(() => {
        debouncedUpdateSearch(searchValue);

        return () => {
            debouncedUpdateSearch.cancel();
        };
    }, [searchValue, debouncedUpdateSearch]);

    useEffect(() => {
        form.reset({ search: '' });
        setDebouncedSearch('');
        debouncedUpdateSearch.cancel();
    }, [flowId, form, debouncedUpdateSearch]);

    const isContainerRunning = flowStatus === StatusType.Running || flowStatus === StatusType.Waiting;

    // Pause subscriptions until the initial query has loaded so that the
    // `flowFiles` cache field exists before subscription-driven updates arrive.
    const isSubscriptionPaused = !flowId || isLoading;

    useFlowFileAddedSubscription({
        skip: isSubscriptionPaused,
        variables: flowFilesVariables,
    });
    useFlowFileUpdatedSubscription({
        skip: isSubscriptionPaused,
        variables: flowFilesVariables,
    });
    useFlowFileDeletedSubscription({
        skip: isSubscriptionPaused,
        variables: flowFilesVariables,
    });

    useEffect(() => {
        if (flowFilesError) {
            toast.error('Failed to load files', {
                description: flowFilesError.message,
                id: 'flow-files-error',
            });
        }
    }, [flowFilesError]);

    // ── upload ─────────────────────────────────────────────────────────────────

    const handleCopyPath = useCallback((file: FileNode) => {
        void copyToClipboard(file.path).then((wasCopied) => {
            if (wasCopied) {
                toast.success('Path copied to clipboard');
            } else {
                toast.error('Failed to copy path');
            }
        });
    }, []);

    const getDownloadHrefForFile = useCallback(
        (file: FileNode) => buildDownloadHref(flowId, file),
        [flowId],
    );

    const fileManagerActions = useMemo<FileManagerAction[]>(
        () => [
            downloadAction(getDownloadHrefForFile),
            copyPathAction(handleCopyPath),
            deleteAction(setFileToDelete),
        ],
        [getDownloadHrefForFile, handleCopyPath],
    );

    const uploadFiles = useCallback(
        async (selectedFiles: File[]) => {
            if (!flowId || !selectedFiles.length) {
                return;
            }

            const formData = new FormData();

            selectedFiles.forEach((file) => formData.append('files', file));

            setIsUploading(true);

            try {
                const response = await api.post<FlowFilesResponse, FormData>(
                    `/flows/${flowId}/files/`,
                    formData,
                    {
                        // Browser sets the multipart boundary automatically when Content-Type is unset.
                        headers: { 'Content-Type': undefined },
                        // Uploads can take longer than the default 30s — disable timeout for this call.
                        timeout: 0,
                    },
                );
                const data = unwrapApiResponse(response);
                const uploadedCount = data.files?.length ?? selectedFiles.length;

                toast.success(uploadedCount === 1 ? 'File uploaded' : `${uploadedCount} files uploaded`, {
                    description:
                        uploadedCount === 1
                            ? `Available at /work/uploads/${data.files?.[0]?.name ?? ''}`
                            : `${uploadedCount} files are now available under /work/uploads`,
                });

                await refetchFiles();
            } catch (error) {
                const description = getApiErrorMessage(error, 'Failed to upload files', {
                    409: 'Entry already exists — enable "Overwrite" to replace it',
                });

                toast.error('Upload failed', { description });
            } finally {
                setIsUploading(false);
            }
        },
        [flowId, refetchFiles],
    );

    const handleFileSelection = useCallback(
        async (event: React.ChangeEvent<HTMLInputElement>) => {
            const selectedFiles = Array.from(event.target.files ?? []);

            try {
                await uploadFiles(selectedFiles);
            } finally {
                event.target.value = '';
            }
        },
        [uploadFiles],
    );

    // ── drag & drop ────────────────────────────────────────────────────────────

    const canAcceptDrop = !!flowId && !isUploading;

    const handleDragEnter = useCallback(
        (event: React.DragEvent<HTMLDivElement>) => {
            if (!canAcceptDrop || !event.dataTransfer.types?.includes('Files')) {
                return;
            }

            event.preventDefault();
            event.stopPropagation();
            dragCounterRef.current += 1;
            setIsDragging(true);
        },
        [canAcceptDrop],
    );

    const handleDragOver = useCallback(
        (event: React.DragEvent<HTMLDivElement>) => {
            if (!canAcceptDrop || !event.dataTransfer.types?.includes('Files')) {
                return;
            }

            event.preventDefault();
            event.stopPropagation();
            event.dataTransfer.dropEffect = 'copy';
        },
        [canAcceptDrop],
    );

    const handleDragLeave = useCallback((event: React.DragEvent<HTMLDivElement>) => {
        event.preventDefault();
        event.stopPropagation();
        dragCounterRef.current = Math.max(dragCounterRef.current - 1, 0);

        if (dragCounterRef.current === 0) {
            setIsDragging(false);
        }
    }, []);

    const handleDrop = useCallback(
        (event: React.DragEvent<HTMLDivElement>) => {
            event.preventDefault();
            event.stopPropagation();
            dragCounterRef.current = 0;
            setIsDragging(false);

            if (!canAcceptDrop) {
                return;
            }

            const droppedFiles = Array.from(event.dataTransfer.files ?? []);

            if (droppedFiles.length === 0) {
                return;
            }

            void uploadFiles(droppedFiles);
        },
        [canAcceptDrop, uploadFiles],
    );

    // Reset overlay state when flow changes mid-drag.
    useEffect(() => {
        dragCounterRef.current = 0;
        setIsDragging(false);
    }, [flowId]);

    const handleDeleteFile = useCallback(async () => {
        if (!flowId || !fileToDelete) {
            return;
        }

        try {
            await api.delete<FlowFilesResponse>(`/flows/${flowId}/files/`, {
                params: { path: fileToDelete.path },
            });
            toast.success(fileToDelete.isDir ? 'Directory deleted' : 'File deleted');
            await refetchFiles();
        } catch (error) {
            const description = getApiErrorMessage(error, 'Failed to delete file');

            toast.error('Delete failed', { description });
        } finally {
            setFileToDelete(null);
        }
    }, [flowId, fileToDelete, refetchFiles]);

    const handleBulkDelete = useCallback(
        async (filesToDelete: FileNode[]) => {
            if (!flowId || filesToDelete.length === 0) {
                return;
            }

            const results = await Promise.allSettled(
                filesToDelete.map((file) =>
                    api.delete<FlowFilesResponse>(`/flows/${flowId}/files/`, {
                        params: { path: file.path },
                    }),
                ),
            );
            const succeeded = results.filter((result) => result.status === 'fulfilled').length;
            const failed = results.length - succeeded;

            if (failed === 0) {
                toast.success(`${succeeded} ${succeeded === 1 ? 'item' : 'items'} deleted`);
            } else if (succeeded === 0) {
                toast.error('Bulk delete failed', {
                    description: `Failed to delete ${failed} ${failed === 1 ? 'item' : 'items'}`,
                });
            } else {
                toast.warning(`${succeeded} succeeded · ${failed} failed`);
            }

            await refetchFiles();
        },
        [flowId, refetchFiles],
    );

    const files = useMemo(() => flowFilesData?.flowFiles ?? [], [flowFilesData?.flowFiles]);
    const fileNodes = useMemo<FileNode[]>(() => files.map(toFileNode), [files]);
    const isInitialLoading = isLoading && fileNodes.length === 0;

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
                    No files match <code>{debouncedSearch.trim()}</code>. Try a different query.
                </EmptyDescription>
            </EmptyHeader>
        </Empty>
    );

    // ── render ─────────────────────────────────────────────────────────────────

    return (
        <div
            className="relative flex h-full flex-col"
            onDragEnter={handleDragEnter}
            onDragLeave={handleDragLeave}
            onDragOver={handleDragOver}
            onDrop={handleDrop}
        >
            <input
                className="hidden"
                multiple
                onChange={handleFileSelection}
                ref={inputRef}
                type="file"
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
                <Form {...form}>
                    <div className="flex gap-2 p-px">
                        <FormField
                            control={form.control}
                            name="search"
                            render={({ field }) => (
                                <FormControl>
                                    <InputGroup className="flex-1">
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
                                                    onClick={() => {
                                                        form.reset({ search: '' });
                                                        setDebouncedSearch('');
                                                        debouncedUpdateSearch.cancel();
                                                    }}
                                                    type="button"
                                                >
                                                    <X />
                                                </InputGroupButton>
                                            </InputGroupAddon>
                                        )}
                                    </InputGroup>
                                </FormControl>
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
                                        disabled={isUploading || isLoading}
                                        onClick={() => inputRef.current?.click()}
                                        size="icon-sm"
                                        type="button"
                                        variant="outline"
                                    >
                                        {isUploading ? <Loader2 className="animate-spin" /> : <FolderUp />}
                                    </Button>
                                </span>
                            </TooltipTrigger>
                            <TooltipContent>Upload files</TooltipContent>
                        </Tooltip>

                        <Tooltip>
                            <TooltipTrigger asChild>
                                <span>
                                    <Button
                                        disabled={!isContainerRunning || isLoading || isUploading}
                                        onClick={() => setShowPullDialog(true)}
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
                onBulkDelete={handleBulkDelete}
                rootGroups={ROOT_GROUPS}
                searchEmptyState={noMatchesState}
                searchQuery={debouncedSearch}
            />

            <PullDialog
                flowId={flowId}
                onClose={() => setShowPullDialog(false)}
                onSuccess={() => void refetchFiles()}
                open={showPullDialog}
            />

            <ConfirmationDialog
                confirmText="Delete"
                handleConfirm={handleDeleteFile}
                handleOpenChange={(isOpen) => {
                    if (!isOpen) {
                        setFileToDelete(null);
                    }
                }}
                isOpen={!!fileToDelete}
                itemName={fileToDelete?.name}
                itemType={fileToDelete?.isDir ? 'directory' : 'file'}
                title={fileToDelete?.isDir ? 'Delete Directory' : 'Delete File'}
            />
        </div>
    );
};

export default FlowFiles;
