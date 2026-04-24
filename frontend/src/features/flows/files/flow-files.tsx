import { AlertCircle, Copy, Download, FileUp, FolderUp, Loader2, RefreshCw } from 'lucide-react';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { toast } from 'sonner';

import { Button, buttonVariants } from '@/components/ui/button';
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';
import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupInput } from '@/components/ui/input-group';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { axios } from '@/lib/axios';
import { copyToClipboard } from '@/lib/report';
import { cn } from '@/lib/utils';
import { formatDate } from '@/lib/utils/format';
import { baseUrl } from '@/models/api';
import { useFlow } from '@/providers/flow-provider';

interface ApiResponse<T> {
    data?: T;
    status: string;
}

interface FlowFile {
    modifiedAt: string;
    name: string;
    path: string;
    size: number;
}

interface FlowFilesResponse {
    files: Array<FlowFile>;
    total: number;
}

const unwrapFlowFilesResponse = (response: ApiResponse<FlowFilesResponse>) => {
    if (response.status !== 'success' || !response.data) {
        throw new Error('Unexpected flow files response');
    }

    return response.data;
};

const formatFileSize = (size: number) => {
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

const FlowFiles = () => {
    const { flowId } = useFlow();
    const inputRef = useRef<HTMLInputElement | null>(null);
    const [files, setFiles] = useState<Array<FlowFile>>([]);
    const [isLoading, setIsLoading] = useState(false);
    const [isUploading, setIsUploading] = useState(false);
    const [searchValue, setSearchValue] = useState('');

    const loadFiles = useCallback(async () => {
        if (!flowId) {
            setFiles([]);

            return;
        }

        setIsLoading(true);

        try {
            const response = await axios.get<unknown, ApiResponse<FlowFilesResponse>>(`/flows/${flowId}/files`);
            const data = unwrapFlowFilesResponse(response);
            setFiles(data.files ?? []);
        } catch (error) {
            const description = error instanceof Error ? error.message : 'An error occurred while loading flow files';
            toast.error('Failed to load files', {
                description,
            });
        } finally {
            setIsLoading(false);
        }
    }, [flowId]);

    useEffect(() => {
        void loadFiles();
    }, [loadFiles]);

    const handleUploadButtonClick = useCallback(() => {
        inputRef.current?.click();
    }, []);

    const handleCopyPath = useCallback(async (filePath: string) => {
        const success = await copyToClipboard(filePath);

        if (success) {
            toast.success('Path copied to clipboard');

            return;
        }

        toast.error('Failed to copy path');
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
                    `/flows/${flowId}/files`,
                    formData,
                    {
                        headers: {
                            'Content-Type': undefined,
                        },
                    },
                );
                const data = unwrapFlowFilesResponse(response);
                const uploadedCount = data.files?.length ?? selectedFiles.length;

                toast.success(uploadedCount === 1 ? 'File uploaded successfully' : 'Files uploaded successfully', {
                    description:
                        uploadedCount === 1
                            ? `${data.files?.[0]?.path ?? '/work/uploads'}`
                            : `${uploadedCount} files are now available under /work/uploads`,
                });

                await loadFiles();
            } catch (error) {
                const description = error instanceof Error ? error.message : 'An error occurred while uploading files';
                toast.error('Failed to upload files', {
                    description,
                });
            } finally {
                setIsUploading(false);
                event.target.value = '';
            }
        },
        [flowId, loadFiles],
    );

    const filteredFiles = useMemo(() => {
        const normalizedSearch = searchValue.toLowerCase().trim();

        if (!normalizedSearch) {
            return files;
        }

        return files.filter((file) => {
            return (
                file.name.toLowerCase().includes(normalizedSearch) || file.path.toLowerCase().includes(normalizedSearch)
            );
        });
    }, [files, searchValue]);

    return (
        <div className="flex h-full flex-col gap-4">
            <input
                className="hidden"
                multiple
                onChange={handleFileSelection}
                ref={inputRef}
                type="file"
            />

            <div className="bg-background sticky top-0 z-10 flex flex-col gap-3 pb-4">
                <div className="rounded-lg border px-4 py-3 text-sm">
                    <div className="flex items-start gap-3">
                        <AlertCircle className="text-muted-foreground mt-0.5 size-4 shrink-0" />
                        <div className="space-y-1">
                            <p className="font-medium">Uploaded files are shared with the whole flow.</p>
                            <p className="text-muted-foreground">
                                PentAGI stores user uploads under <code>/work/uploads</code>, so automation, assistants,
                                and container commands can access the same files immediately.
                            </p>
                        </div>
                    </div>
                </div>

                <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
                    <InputGroup className="flex-1">
                        <InputGroupAddon>
                            <FileUp />
                        </InputGroupAddon>
                        <InputGroupInput
                            autoComplete="off"
                            onChange={(event) => setSearchValue(event.target.value)}
                            placeholder="Filter uploaded files..."
                            type="text"
                            value={searchValue}
                        />
                        {searchValue && (
                            <InputGroupAddon align="inline-end">
                                <InputGroupButton
                                    onClick={() => setSearchValue('')}
                                    type="button"
                                >
                                    Clear
                                </InputGroupButton>
                            </InputGroupAddon>
                        )}
                    </InputGroup>

                    <div className="flex items-center gap-2">
                        <Button
                            disabled={isLoading || isUploading}
                            onClick={() => void loadFiles()}
                            type="button"
                            variant="outline"
                        >
                            {isLoading ? <Loader2 className="animate-spin" /> : <RefreshCw />}
                            Refresh
                        </Button>
                        <Button
                            disabled={isUploading}
                            onClick={handleUploadButtonClick}
                            type="button"
                        >
                            {isUploading ? <Loader2 className="animate-spin" /> : <FolderUp />}
                            Upload
                        </Button>
                    </div>
                </div>
            </div>

            {filteredFiles.length > 0 ? (
                <div className="flex flex-1 flex-col gap-3 overflow-y-auto">
                    {filteredFiles.map((file) => (
                        <div
                            className="bg-card text-card-foreground rounded-xl border p-4 shadow-sm"
                            key={`${file.name}-${file.modifiedAt}`}
                        >
                            <div className="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
                                <div className="min-w-0 space-y-2">
                                    <div className="flex items-center gap-2">
                                        <FileUp className="text-muted-foreground size-4 shrink-0" />
                                        <p className="truncate font-semibold">{file.name}</p>
                                    </div>
                                    <div className="text-muted-foreground space-y-1 text-sm">
                                        <p>{formatFileSize(file.size)}</p>
                                        <div className="flex items-center gap-2">
                                            <code className="bg-muted rounded px-2 py-1 text-xs">{file.path}</code>
                                            <Tooltip>
                                                <TooltipTrigger asChild>
                                                    <button
                                                        className={cn(
                                                            buttonVariants({ size: 'icon-xs', variant: 'ghost' }),
                                                        )}
                                                        onClick={() => void handleCopyPath(file.path)}
                                                        type="button"
                                                    >
                                                        <Copy />
                                                    </button>
                                                </TooltipTrigger>
                                                <TooltipContent>Copy container path</TooltipContent>
                                            </Tooltip>
                                        </div>
                                        <p>{formatDate(new Date(file.modifiedAt))}</p>
                                    </div>
                                </div>

                                <a
                                    className={cn(buttonVariants({ size: 'sm', variant: 'outline' }), 'shrink-0')}
                                    href={`${baseUrl}/flows/${flowId}/files/${encodeURIComponent(file.name)}`}
                                >
                                    <Download />
                                    Download
                                </a>
                            </div>
                        </div>
                    ))}
                </div>
            ) : (
                <Empty>
                    <EmptyHeader>
                        <EmptyMedia variant="icon">
                            <FolderUp />
                        </EmptyMedia>
                        <EmptyTitle>No uploaded files</EmptyTitle>
                        <EmptyDescription>
                            Upload files into <code>/work/uploads</code> when you want this flow, its assistants, or
                            container commands to use them.
                        </EmptyDescription>
                    </EmptyHeader>
                </Empty>
            )}
        </div>
    );
};

export default FlowFiles;
