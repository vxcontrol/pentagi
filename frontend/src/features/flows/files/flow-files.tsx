import { zodResolver } from '@hookform/resolvers/zod';
import debounce from 'lodash/debounce';
import {
    ArrowDownToLine,
    Copy,
    Download,
    File,
    Folder,
    FolderUp,
    HardDrive,
    Info,
    Loader2,
    Search,
    Trash2,
    X,
} from 'lucide-react';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useForm } from 'react-hook-form';
import { toast } from 'sonner';
import { z } from 'zod';

import ConfirmationDialog from '@/components/shared/confirmation-dialog';
import { Button, buttonVariants } from '@/components/ui/button';
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
import { axios } from '@/lib/axios';
import { copyToClipboard } from '@/lib/report';
import { cn } from '@/lib/utils';
import { formatDate } from '@/lib/utils/format';
import { baseUrl } from '@/models/api';
import { useFlow } from '@/providers/flow-provider';

// ── types ─────────────────────────────────────────────────────────────────────

interface ApiErrorData {
    code?: string;
    msg?: string;
}

interface ApiResponse<T> {
    data?: T;
    status: string;
}

interface AxiosLikeError {
    message?: string;
    response?: { data?: ApiErrorData; status?: number };
    statusCode?: number;
}

type FileSource = 'container' | 'unknown' | 'uploads';
type FlowFile = FlowFileFragmentFragment;

interface FlowFilesResponse {
    files: Array<FlowFile>;
    total: number;
}

interface FlowFileTreeItem {
    depth: number;
    file: FlowFile;
}

const searchFormSchema = z.object({
    search: z.string(),
});

const unwrapFlowFilesResponse = (response: ApiResponse<FlowFilesResponse>) => {
    if (response.status !== 'success' || !response.data) {
        throw new Error('Unexpected response from server');
    }

    return response.data;
};

const getErrorMessage = (error: unknown, fallback: string): string => {
    const err = error as AxiosLikeError;

    if (err.statusCode === 409 || err.response?.status === 409) {
        return err.response?.data?.msg ?? 'Entry already exists — enable "Overwrite" to replace it';
    }

    if (err.statusCode === 400 || err.response?.status === 400) {
        return err.response?.data?.msg ?? err.message ?? fallback;
    }

    return err.message ?? fallback;
};

const fileSource = (filePath: string): FileSource => {
    if (filePath.startsWith('container/') || filePath === 'container') {
        return 'container';
    }

    if (filePath.startsWith('uploads/') || filePath === 'uploads') {
        return 'uploads';
    }

    return 'unknown';
};

const formatFileSize = (size: number): string => {
    if (size < 1024) {
        return `${size} B`;
    }

    const units = ['KB', 'MB', 'GB', 'TB'];
    let unitIndex = -1;
    let value = size;

    while (value >= 1024 && unitIndex < units.length - 1) {
        value /= 1024;
        unitIndex += 1;
    }

    return `${value.toFixed(value >= 10 ? 0 : 1)} ${units[unitIndex]}`;
};

const buildDownloadHref = (flowId: null | string, file: FlowFile) =>
    `${baseUrl}/flows/${flowId}/files/download?path=${encodeURIComponent(file.path)}`;

const filePathParts = (filePath: string) => filePath.split('/').filter(Boolean);

const flowFileDepth = (filePath: string): number => {
    const parts = filePathParts(filePath);

    return Math.max(parts.length - 2, 0);
};

const compareFlowFileTreePath = (a: FlowFile, b: FlowFile): number => {
    const aParts = filePathParts(a.path).slice(1);
    const bParts = filePathParts(b.path).slice(1);
    const length = Math.min(aParts.length, bParts.length);

    for (let i = 0; i < length; i += 1) {
        const partCompare = (aParts[i] ?? '').localeCompare(bParts[i] ?? '');

        if (partCompare !== 0) {
            return partCompare;
        }
    }

    return aParts.length - bParts.length;
};

const buildFlowFileTree = (files: FlowFile[]): FlowFileTreeItem[] =>
    [...files].sort(compareFlowFileTreePath).map((file) => ({
        depth: flowFileDepth(file.path),
        file,
    }));

// ── section header ─────────────────────────────────────────────────────────────

const SectionHeader = ({ children, icon }: { children: React.ReactNode; icon: React.ReactNode }) => (
    <div className="flex items-center gap-2 py-1">
        <div className="text-muted-foreground flex items-center gap-1.5 text-xs font-medium">
            {icon}
            {children}
        </div>
        <div className="bg-border h-px flex-1" />
    </div>
);

// ── file item ──────────────────────────────────────────────────────────────────

interface FlowFileItemProps {
    depth?: number;
    file: FlowFile;
    flowId: null | string;
    onCopyPath: (path: string) => void;
    onDelete: (file: FlowFile) => void;
}

const FlowFileItem = ({ depth = 0, file, flowId, onCopyPath, onDelete }: FlowFileItemProps) => (
    <div
        className="flex flex-col items-start"
        style={{ paddingLeft: depth * 20 }}
    >
        <div className="bg-card text-card-foreground w-full rounded-xl border p-3 shadow-sm">
            <div className="flex items-center justify-between gap-2">
                <div className="flex min-w-0 flex-1 items-center gap-2">
                    {file.isDir ? (
                        <Folder className="text-muted-foreground size-4 shrink-0" />
                    ) : (
                        <File className="text-muted-foreground size-4 shrink-0" />
                    )}
                    <span className="truncate text-sm font-semibold">{file.name}</span>
                </div>

                <Tooltip>
                    <TooltipTrigger asChild>
                        <a
                            className={cn(buttonVariants({ size: 'icon-xs', variant: 'ghost' }))}
                            download={file.isDir ? `${file.name}.zip` : file.name}
                            href={buildDownloadHref(flowId, file)}
                        >
                            <Download className="size-3.5" />
                        </a>
                    </TooltipTrigger>
                    <TooltipContent>{file.isDir ? 'Download as ZIP' : 'Download'}</TooltipContent>
                </Tooltip>

                <Tooltip>
                    <TooltipTrigger asChild>
                        <Button
                            onClick={() => onDelete(file)}
                            size="icon-xs"
                            type="button"
                            variant="ghost"
                        >
                            <Trash2 className="size-3.5" />
                        </Button>
                    </TooltipTrigger>
                    <TooltipContent>{file.isDir ? 'Delete directory' : 'Delete file'}</TooltipContent>
                </Tooltip>
            </div>
        </div>

        <div className="text-muted-foreground mt-1 flex items-center gap-1 px-1 text-xs">
            {!file.isDir && (
                <>
                    <span className="text-muted-foreground/50">{formatFileSize(file.size)}</span>
                    <span className="text-muted-foreground/30">·</span>
                </>
            )}
            <Tooltip>
                <TooltipTrigger asChild>
                    <Copy
                        className="hover:text-foreground size-3 shrink-0 cursor-pointer transition-colors"
                        onClick={() => onCopyPath(file.path)}
                    />
                </TooltipTrigger>
                <TooltipContent>Copy cache path</TooltipContent>
            </Tooltip>
            <span className="text-muted-foreground/30">·</span>
            <span className="text-muted-foreground/50">{formatDate(new Date(file.modifiedAt))}</span>
        </div>
    </div>
);

// ── pull dialog ────────────────────────────────────────────────────────────────

interface PullDialogProps {
    flowId: null | string;
    isPulling: boolean;
    onClose: () => void;
    onSuccess: () => void;
    open: boolean;
}

const PullDialog = ({ flowId, isPulling: externalIsPulling, onClose, onSuccess, open }: PullDialogProps) => {
    const [containerPath, setContainerPath] = useState('');
    const [force, setForce] = useState(false);
    const [isPulling, setIsPulling] = useState(false);

    useEffect(() => {
        if (!open) {
            setContainerPath('');
            setForce(false);
        }
    }, [open]);

    const handlePull = useCallback(async () => {
        if (!flowId || !containerPath.trim()) {
            return;
        }

        setIsPulling(true);

        try {
            await axios.post<unknown, ApiResponse<FlowFilesResponse>>(`/flows/${flowId}/files/pull`, {
                force,
                path: containerPath.trim(),
            });
            toast.success('Pulled from container', {
                description: `Saved to local cache under container/`,
            });
            onSuccess();
            onClose();
        } catch (error) {
            const description = getErrorMessage(error, 'Failed to pull from container');
            toast.error('Pull failed', { description });
        } finally {
            setIsPulling(false);
        }
    }, [flowId, containerPath, force, onSuccess, onClose]);

    const isSubmitting = isPulling || externalIsPulling;

    return (
        <Dialog
            onOpenChange={(v) => {
                if (!v && !isSubmitting) {
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
                            disabled={isSubmitting}
                            id="container-path"
                            onChange={(e) => setContainerPath(e.target.value)}
                            onKeyDown={(e) => {
                                if (e.key === 'Enter' && containerPath.trim() && !isSubmitting) {
                                    void handlePull();
                                }
                            }}
                            placeholder="/etc/nginx/conf"
                            value={containerPath}
                        />
                    </div>

                    <div className="flex items-center gap-2">
                        <Switch
                            checked={force}
                            disabled={isSubmitting}
                            id="force-pull"
                            onCheckedChange={setForce}
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
                        disabled={isSubmitting}
                        onClick={onClose}
                        variant="outline"
                    >
                        Cancel
                    </Button>
                    <Button
                        disabled={!containerPath.trim() || isSubmitting}
                        onClick={() => void handlePull()}
                    >
                        {isSubmitting ? <Loader2 className="animate-spin" /> : <ArrowDownToLine />}
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
    const [isUploading, setIsUploading] = useState(false);
    const [fileToDelete, setFileToDelete] = useState<FlowFile | null>(null);
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
        return () => {
            debouncedUpdateSearch.cancel();
        };
    }, [debouncedUpdateSearch]);

    useEffect(() => {
        form.reset({ search: '' });
        setDebouncedSearch('');
        debouncedUpdateSearch.cancel();
    }, [flowId, form, debouncedUpdateSearch]);

    const isContainerRunning = flowStatus === StatusType.Running || flowStatus === StatusType.Waiting;

    useFlowFileAddedSubscription({
        skip: !flowId,
        variables: flowFilesVariables,
    });
    useFlowFileUpdatedSubscription({
        skip: !flowId,
        variables: flowFilesVariables,
    });
    useFlowFileDeletedSubscription({
        skip: !flowId,
        variables: flowFilesVariables,
    });

    useEffect(() => {
        if (flowFilesError) {
            toast.error('Failed to load files', { description: flowFilesError.message });
        }
    }, [flowFilesError]);

    // ── upload ─────────────────────────────────────────────────────────────────

    const handleCopyPath = useCallback(async (filePath: string) => {
        const success = await copyToClipboard(filePath);

        if (success) {
            toast.success('Path copied to clipboard');
        } else {
            toast.error('Failed to copy path');
        }
    }, []);

    const handleFileSelection = useCallback(
        async (event: React.ChangeEvent<HTMLInputElement>) => {
            if (!flowId) {
                return;
            }

            const selectedFiles = Array.from(event.target.files ?? []);

            if (selectedFiles.length === 0) {
                return;
            }

            const formData = new FormData();

            for (const file of selectedFiles) {
                formData.append('files', file);
            }

            setIsUploading(true);

            try {
                const response = await axios.post<unknown, ApiResponse<FlowFilesResponse>>(
                    `/flows/${flowId}/files/`,
                    formData,
                    {
                        headers: {
                            'Content-Type': undefined,
                        },
                    },
                );
                const data = unwrapFlowFilesResponse(response);
                const uploadedCount = data.files?.length ?? selectedFiles.length;

                toast.success(uploadedCount === 1 ? 'File uploaded' : `${uploadedCount} files uploaded`, {
                    description:
                        uploadedCount === 1
                            ? `Available at /work/uploads/${data.files?.[0]?.name ?? ''}`
                            : `${uploadedCount} files are now available under /work/uploads`,
                });

                await refetchFiles();
            } catch (error) {
                const description = getErrorMessage(error, 'Failed to upload files');
                toast.error('Upload failed', { description });
            } finally {
                setIsUploading(false);
                event.target.value = '';
            }
        },
        [flowId, refetchFiles],
    );

    const handleDeleteFile = useCallback(async () => {
        if (!flowId || !fileToDelete) {
            return;
        }

        try {
            await axios.delete<unknown, ApiResponse<FlowFilesResponse>>(`/flows/${flowId}/files/`, {
                params: { path: fileToDelete.path },
            });
            toast.success(fileToDelete.isDir ? 'Directory deleted' : 'File deleted');
            await refetchFiles();
        } catch (error) {
            const description = getErrorMessage(error, 'Failed to delete file');
            toast.error('Delete failed', { description });
        } finally {
            setFileToDelete(null);
        }
    }, [flowId, fileToDelete, refetchFiles]);

    // ── filtering + grouping ───────────────────────────────────────────────────

    const files = useMemo(() => flowFilesData?.flowFiles ?? [], [flowFilesData?.flowFiles]);

    const filteredFiles = useMemo(() => {
        const search = debouncedSearch.toLowerCase().trim();

        if (!search) {
            return files;
        }

        return files.filter((f) => f.name.toLowerCase().includes(search) || f.path.toLowerCase().includes(search));
    }, [files, debouncedSearch]);

    const groups = useMemo(() => {
        const uploads: FlowFile[] = [];
        const container: FlowFile[] = [];

        for (const f of filteredFiles) {
            const src = fileSource(f.path);

            if (src === 'uploads') {
                uploads.push(f);
            } else if (src === 'container') {
                container.push(f);
            }
        }

        return {
            container: buildFlowFileTree(container),
            uploads: buildFlowFileTree(uploads),
        };
    }, [filteredFiles]);

    const hasFiles = filteredFiles.length > 0;

    // ── render ─────────────────────────────────────────────────────────────────

    return (
        <div className="flex h-full flex-col">
            <input
                className="hidden"
                multiple
                onChange={handleFileSelection}
                ref={inputRef}
                type="file"
            />

            {/* Toolbar — same sticky pattern as screenshots / vector-stores */}
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

                        {/* Info hint */}
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

                        {/* Upload button */}
                        <Tooltip>
                            <TooltipTrigger asChild>
                                <Button
                                    disabled={isUploading || isLoading}
                                    onClick={() => inputRef.current?.click()}
                                    size="icon-sm"
                                    type="button"
                                    variant="outline"
                                >
                                    {isUploading ? <Loader2 className="animate-spin" /> : <FolderUp />}
                                </Button>
                            </TooltipTrigger>
                            <TooltipContent>Upload files</TooltipContent>
                        </Tooltip>

                        {/* Pull button — wraps in span so tooltip works when disabled */}
                        <Tooltip>
                            <TooltipTrigger asChild>
                                <span>
                                    <Button
                                        disabled={!isContainerRunning || isLoading}
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

            {/* File list grouped by source */}
            {hasFiles ? (
                <div className="flex flex-col gap-1 overflow-y-auto">
                    {groups.uploads.length > 0 && (
                        <>
                            <SectionHeader icon={<FolderUp className="size-3" />}>Uploads</SectionHeader>
                            <div className="flex flex-col gap-3 pb-3">
                                {groups.uploads.map(({ depth, file }) => (
                                    <FlowFileItem
                                        depth={depth}
                                        file={file}
                                        flowId={flowId}
                                        key={`${file.path}-${file.modifiedAt}`}
                                        onCopyPath={(p) => void handleCopyPath(p)}
                                        onDelete={setFileToDelete}
                                    />
                                ))}
                            </div>
                        </>
                    )}

                    {groups.container.length > 0 && (
                        <>
                            <SectionHeader icon={<HardDrive className="size-3" />}>Container</SectionHeader>
                            <div className="flex flex-col gap-3 pb-3">
                                {groups.container.map(({ depth, file }) => (
                                    <FlowFileItem
                                        depth={depth}
                                        file={file}
                                        flowId={flowId}
                                        key={`${file.path}-${file.modifiedAt}`}
                                        onCopyPath={(p) => void handleCopyPath(p)}
                                        onDelete={setFileToDelete}
                                    />
                                ))}
                            </div>
                        </>
                    )}
                </div>
            ) : (
                <Empty>
                    <EmptyHeader>
                        <EmptyMedia variant="icon">
                            <FolderUp />
                        </EmptyMedia>
                        <EmptyTitle>No files in cache</EmptyTitle>
                        <EmptyDescription>
                            Upload files to make them available at <code>/work/uploads</code>, or use Pull to sync files
                            from the running container.
                        </EmptyDescription>
                    </EmptyHeader>
                </Empty>
            )}

            <PullDialog
                flowId={flowId}
                isPulling={false}
                onClose={() => setShowPullDialog(false)}
                onSuccess={() => void refetchFiles()}
                open={showPullDialog}
            />

            <ConfirmationDialog
                confirmText="Delete"
                handleConfirm={() => void handleDeleteFile()}
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
