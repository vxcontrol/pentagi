import type { ColumnDef } from '@tanstack/react-table';

import { format, isToday } from 'date-fns';
import { enUS } from 'date-fns/locale';
import {
    AlertCircle,
    ArrowDown,
    ArrowUp,
    CalendarIcon,
    Check,
    Copy,
    ExternalLink,
    Key,
    Loader2,
    MoreHorizontal,
    Pencil,
    Plus,
    Trash,
    X,
} from 'lucide-react';
import { useCallback, useMemo, useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import { toast } from 'sonner';

import type { ApiTokenFragmentFragment } from '@/graphql/types';

import ConfirmationDialog from '@/components/shared/confirmation-dialog';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Calendar } from '@/components/ui/calendar';
import { DataTable } from '@/components/ui/data-table';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Input } from '@/components/ui/input';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { StatusCard } from '@/components/ui/status-card';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import {
    TokenStatus as TokenStatusEnum,
    useApiTokenCreatedSubscription,
    useApiTokenDeletedSubscription,
    useApiTokensQuery,
    useApiTokenUpdatedSubscription,
    useCreateApiTokenMutation,
    useDeleteApiTokenMutation,
    useUpdateApiTokenMutation,
} from '@/graphql/types';
import { useAdaptiveColumnVisibility } from '@/hooks/use-adaptive-column-visibility';
import { cn } from '@/lib/utils';
import { baseUrl } from '@/models/api';

type APIToken = ApiTokenFragmentFragment;

interface CreateFormData {
    expiresAt: Date | null;
    name: string;
}

interface EditFormData {
    name: string;
    status: TokenStatusEnum;
}

const isTokenExpired = (token: APIToken): boolean => {
    const expiresAt = new Date(token.createdAt);

    expiresAt.setSeconds(expiresAt.getSeconds() + token.ttl);

    return expiresAt < new Date();
};

const getTokenExpirationDate = (token: APIToken): Date => {
    const expiresAt = new Date(token.createdAt);

    expiresAt.setSeconds(expiresAt.getSeconds() + token.ttl);

    return expiresAt;
};

const getStatusDisplay = (
    token: APIToken,
): { label: string; variant: 'default' | 'destructive' | 'outline' | 'secondary' } => {
    const expired = isTokenExpired(token);

    if (expired) {
        return { label: 'expired', variant: 'destructive' };
    }

    if (token.status === 'active') {
        return { label: 'active', variant: 'default' };
    }

    if (token.status === 'revoked') {
        return { label: 'revoked', variant: 'outline' };
    }

    return { label: token.status, variant: 'secondary' };
};

const formatDateTime = (dateString: string) => {
    const date = new Date(dateString);

    if (isToday(date)) {
        return format(date, 'HH:mm:ss', { locale: enUS });
    }

    return format(date, 'd MMM yyyy', { locale: enUS });
};

const formatFullDateTime = (dateString: string) => {
    const date = new Date(dateString);

    return format(date, 'd MMM yyyy, HH:mm:ss', { locale: enUS });
};

const calculateTTL = (expiresAt: Date): number => {
    const now = new Date();
    const diffMs = expiresAt.getTime() - now.getTime();
    const diffSeconds = Math.ceil(diffMs / 1000);

    return Math.max(60, diffSeconds);
};

const copyToClipboard = async (text: string): Promise<boolean> => {
    try {
        await navigator.clipboard.writeText(text);

        return true;
    } catch (error) {
        console.error('Failed to copy to clipboard:', error);

        return false;
    }
};

const SettingsAPITokensHeader = ({ onCreateClick }: { onCreateClick: () => void }) => {
    return (
        <div className="flex items-center justify-between gap-4">
            <div className="flex flex-col gap-2">
                <p className="text-muted-foreground">Manage API tokens for programmatic access</p>
                <div className="flex gap-4 text-sm">
                    <a
                        className="text-primary inline-flex items-center gap-1 underline hover:no-underline"
                        href={`${window.location.origin}${baseUrl}/graphql/playground`}
                        rel="noopener noreferrer"
                        target="_blank"
                    >
                        GraphQL Playground
                        <ExternalLink className="size-3" />
                    </a>
                    <a
                        className="text-primary inline-flex items-center gap-1 underline hover:no-underline"
                        href={`${window.location.origin}${baseUrl}/swagger/index.html`}
                        rel="noopener noreferrer"
                        target="_blank"
                    >
                        Swagger UI
                        <ExternalLink className="size-3" />
                    </a>
                </div>
            </div>

            <Button
                onClick={onCreateClick}
                variant="secondary"
            >
                <Plus className="size-4" />
                Create Token
            </Button>
        </div>
    );
};

const createNewTokenPlaceholder: APIToken = {
    createdAt: new Date().toISOString(),
    id: 'create-new',
    name: null,
    roleId: '0',
    status: TokenStatusEnum.Active,
    tokenId: '',
    ttl: 0,
    updatedAt: new Date().toISOString(),
    userId: '0',
};

const SettingsAPITokens = () => {
    const [searchParams, setSearchParams] = useSearchParams();
    const { data, error, loading: isLoading } = useApiTokensQuery();
    const [createAPIToken, { error: createError, loading: isCreateLoading }] = useCreateApiTokenMutation();
    const [updateAPIToken, { error: updateError, loading: isUpdateLoading }] = useUpdateApiTokenMutation();
    const [deleteAPIToken, { error: deleteError, loading: isDeleteLoading }] = useDeleteApiTokenMutation();

    const [editingTokenId, setEditingTokenId] = useState<null | string>(null);
    const [creatingToken, setCreatingToken] = useState(false);
    const [editFormData, setEditFormData] = useState<EditFormData>({ name: '', status: TokenStatusEnum.Active });
    const [createFormData, setCreateFormData] = useState<CreateFormData>({ expiresAt: null, name: '' });
    const [tokenSecret, setTokenSecret] = useState<null | string>(null);
    const [showTokenDialog, setShowTokenDialog] = useState(false);
    const [deleteErrorMessage, setDeleteErrorMessage] = useState<null | string>(null);
    const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
    const [deletingToken, setDeletingToken] = useState<APIToken | null>(null);

    const { columnVisibility, updateColumnVisibility } = useAdaptiveColumnVisibility({
        columns: [
            { alwaysVisible: true, id: 'name', priority: 0 },
            { alwaysVisible: true, id: 'tokenId', priority: 0 },
            { id: 'status', priority: 1 },
            { id: 'createdAt', priority: 2 },
            { id: 'expires', priority: 3 },
        ],
        tableKey: 'api-tokens',
    });

    // Get current page from URL
    const currentPage = useMemo(() => {
        const page = searchParams.get('page');

        return page ? Math.max(0, Number.parseInt(page, 10) - 1) : 0;
    }, [searchParams]);

    // Handle page change
    const handlePageChange = useCallback(
        (pageIndex: number) => {
            const newParams = new URLSearchParams(searchParams);

            if (pageIndex === 0) {
                newParams.delete('page');
            } else {
                newParams.set('page', String(pageIndex + 1));
            }

            setSearchParams(newParams);
        },
        [searchParams, setSearchParams],
    );

    // Three-way sorting handler: null -> asc -> desc -> null
    const handleColumnSort = useCallback(
        (column: {
            clearSorting: () => void;
            getIsSorted: () => 'asc' | 'desc' | false;
            toggleSorting: (desc?: boolean) => void;
        }) => {
            const sorted = column.getIsSorted();

            if (sorted === 'asc') {
                column.toggleSorting(true);
            } else if (sorted === 'desc') {
                column.clearSorting();
            } else {
                column.toggleSorting(false);
            }
        },
        [],
    );

    useApiTokenCreatedSubscription({
        onData: ({ client }) => {
            client.refetchQueries({ include: ['apiTokens'] });
        },
    });

    useApiTokenUpdatedSubscription({
        onData: ({ client }) => {
            client.refetchQueries({ include: ['apiTokens'] });
        },
    });

    useApiTokenDeletedSubscription({
        onData: ({ client }) => {
            client.refetchQueries({ include: ['apiTokens'] });
        },
    });

    const handleEdit = useCallback((token: APIToken) => {
        setEditingTokenId(token.tokenId);
        setEditFormData({
            name: token.name || '',
            status: token.status,
        });
    }, []);

    const handleCancelEdit = useCallback(() => {
        setEditingTokenId(null);
        setEditFormData({ name: '', status: TokenStatusEnum.Active });
    }, []);

    const handleSave = useCallback(
        async (tokenId: string) => {
            try {
                await updateAPIToken({
                    refetchQueries: ['apiTokens'],
                    variables: {
                        input: {
                            name: editFormData.name || null,
                            status: editFormData.status,
                        },
                        tokenId,
                    },
                });

                setEditingTokenId(null);
                setEditFormData({ name: '', status: TokenStatusEnum.Active });
            } catch (error) {
                console.error('Failed to update token:', error);
            }
        },
        [editFormData, updateAPIToken],
    );

    const handleCreateNew = useCallback(() => {
        setCreatingToken(true);
        setCreateFormData({ expiresAt: null, name: '' });
    }, []);

    const handleCancelCreate = useCallback(() => {
        setCreatingToken(false);
        setCreateFormData({ expiresAt: null, name: '' });
    }, []);

    const handleCreate = useCallback(async () => {
        if (!createFormData.expiresAt) {
            return;
        }

        try {
            const ttl = calculateTTL(createFormData.expiresAt);
            const result = await createAPIToken({
                refetchQueries: ['apiTokens'],
                variables: {
                    input: {
                        name: createFormData.name || null,
                        ttl,
                    },
                },
            });

            if (result.data?.createAPIToken) {
                setTokenSecret(result.data.createAPIToken.token);
                setShowTokenDialog(true);
            }

            setCreatingToken(false);
            setCreateFormData({ expiresAt: null, name: '' });
        } catch (error) {
            console.error('Failed to create token:', error);
        }
    }, [createAPIToken, createFormData]);

    const handleDeleteDialogOpen = useCallback((token: APIToken) => {
        setDeletingToken(token);
        setIsDeleteDialogOpen(true);
    }, []);

    const handleDelete = useCallback(
        async (tokenId: string | undefined) => {
            if (!tokenId) {
                return;
            }

            try {
                setDeleteErrorMessage(null);

                await deleteAPIToken({
                    refetchQueries: ['apiTokens'],
                    variables: { tokenId },
                });

                setDeletingToken(null);
                setDeleteErrorMessage(null);
            } catch (error) {
                setDeleteErrorMessage(error instanceof Error ? error.message : 'An error occurred while deleting');
            }
        },
        [deleteAPIToken],
    );

    const handleCopyTokenId = useCallback(async (tokenId: string) => {
        const success = await copyToClipboard(tokenId);

        if (success) {
            toast.success('Token ID copied to clipboard');

            return;
        }

        toast.error('Failed to copy token ID to clipboard');
    }, []);

    const columns: ColumnDef<APIToken>[] = useMemo(
        () => [
            {
                accessorKey: 'name',
                cell: ({ row }) => {
                    const token = row.original;
                    const isCreating = token.id === 'create-new';
                    const isEditing = editingTokenId === token.tokenId;

                    if (isCreating) {
                        return (
                            <Input
                                autoFocus
                                className="h-8"
                                key="create-name-input"
                                onChange={(e) => setCreateFormData((prev) => ({ ...prev, name: e.target.value }))}
                                placeholder="Token name (optional)"
                                value={createFormData.name}
                            />
                        );
                    }

                    if (isEditing) {
                        return (
                            <Input
                                autoFocus
                                className="h-8"
                                key={`edit-name-input-${token.tokenId}`}
                                onChange={(e) => setEditFormData((prev) => ({ ...prev, name: e.target.value }))}
                                placeholder="Token name (optional)"
                                value={editFormData.name}
                            />
                        );
                    }

                    return (
                        <div className="font-medium">
                            {token.name || <span className="text-muted-foreground">(unnamed)</span>}
                        </div>
                    );
                },
                enableHiding: false,
                header: ({ column }) => {
                    const sorted = column.getIsSorted();

                    return (
                        <Button
                            className="text-muted-foreground hover:text-primary flex items-center gap-2 p-0 no-underline hover:no-underline"
                            onClick={() => handleColumnSort(column)}
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
                size: 300,
            },
            {
                accessorKey: 'tokenId',
                cell: ({ row }) => {
                    const token = row.original;
                    const isCreating = token.id === 'create-new';

                    if (isCreating) {
                        return <div className="text-muted-foreground text-sm">N/A</div>;
                    }

                    const tokenId = row.getValue('tokenId') as string;

                    return (
                        <div className="flex items-center gap-2">
                            <code className="text-sm">{tokenId}</code>
                            <Button
                                className="size-6 p-0"
                                onClick={() => handleCopyTokenId(tokenId)}
                                variant="ghost"
                            >
                                <Copy className="size-3" />
                            </Button>
                        </div>
                    );
                },
                enableHiding: false,
                header: ({ column }) => {
                    const sorted = column.getIsSorted();

                    return (
                        <Button
                            className="text-muted-foreground hover:text-primary flex items-center gap-2 p-0 no-underline hover:no-underline"
                            onClick={() => handleColumnSort(column)}
                            variant="link"
                        >
                            Token ID
                            {sorted === 'asc' ? (
                                <ArrowDown className="size-4" />
                            ) : sorted === 'desc' ? (
                                <ArrowUp className="size-4" />
                            ) : null}
                        </Button>
                    );
                },
                size: 200,
            },
            {
                accessorKey: 'status',
                cell: ({ row }) => {
                    const token = row.original;
                    const isCreating = token.id === 'create-new';

                    if (isCreating) {
                        return <Badge variant="default">active</Badge>;
                    }

                    const isEditing = editingTokenId === token.tokenId;
                    const expired = isTokenExpired(token);
                    const statusDisplay = getStatusDisplay(token);

                    if (isEditing) {
                        if (expired) {
                            return <Badge variant={statusDisplay.variant}>{statusDisplay.label}</Badge>;
                        }

                        return (
                            <Select
                                onValueChange={(value) =>
                                    setEditFormData((prev) => ({ ...prev, status: value as TokenStatusEnum }))
                                }
                                value={editFormData.status}
                            >
                                <SelectTrigger className="h-8 w-32">
                                    <SelectValue />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectGroup>
                                        <SelectItem value={TokenStatusEnum.Active}>active</SelectItem>
                                        <SelectItem value={TokenStatusEnum.Revoked}>revoked</SelectItem>
                                    </SelectGroup>
                                </SelectContent>
                            </Select>
                        );
                    }

                    return <Badge variant={statusDisplay.variant}>{statusDisplay.label}</Badge>;
                },
                header: ({ column }) => {
                    const sorted = column.getIsSorted();

                    return (
                        <Button
                            className="text-muted-foreground hover:text-primary flex items-center gap-2 p-0 no-underline hover:no-underline"
                            onClick={() => handleColumnSort(column)}
                            variant="link"
                        >
                            Status
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
                accessorKey: 'expires',
                cell: ({ row }) => {
                    const token = row.original;
                    const isCreating = token.id === 'create-new';

                    if (isCreating) {
                        const tomorrow = new Date();

                        tomorrow.setDate(tomorrow.getDate() + 1);
                        tomorrow.setHours(0, 0, 0, 0);

                        return (
                            <Popover>
                                <PopoverTrigger asChild>
                                    <Button
                                        className={cn(
                                            'h-8 w-full justify-start text-left font-normal',
                                            !createFormData.expiresAt && 'text-muted-foreground',
                                        )}
                                        variant="outline"
                                    >
                                        <CalendarIcon className="mr-2 size-4" />
                                        {createFormData.expiresAt ? (
                                            createFormData.expiresAt.toLocaleDateString()
                                        ) : (
                                            <span>Pick date</span>
                                        )}
                                    </Button>
                                </PopoverTrigger>
                                <PopoverContent
                                    align="start"
                                    className="w-auto p-0"
                                >
                                    <Calendar
                                        disabled={{ before: tomorrow }}
                                        mode="single"
                                        onSelect={(date) => {
                                            setCreateFormData((prev) => ({ ...prev, expiresAt: date || null }));
                                        }}
                                        selected={createFormData.expiresAt || undefined}
                                    />
                                </PopoverContent>
                            </Popover>
                        );
                    }

                    const expiresAt = getTokenExpirationDate(token);
                    const expiresAtString = expiresAt.toISOString();

                    return (
                        <Tooltip>
                            <TooltipTrigger asChild>
                                <div className="cursor-default text-sm">{formatDateTime(expiresAtString)}</div>
                            </TooltipTrigger>
                            <TooltipContent>
                                <div className="text-xs">{formatFullDateTime(expiresAtString)}</div>
                            </TooltipContent>
                        </Tooltip>
                    );
                },
                header: ({ column }) => {
                    const sorted = column.getIsSorted();

                    return (
                        <Button
                            className="text-muted-foreground hover:text-primary flex items-center gap-2 p-0 no-underline hover:no-underline"
                            onClick={() => handleColumnSort(column)}
                            variant="link"
                        >
                            Expires
                            {sorted === 'asc' ? (
                                <ArrowDown className="size-4" />
                            ) : sorted === 'desc' ? (
                                <ArrowUp className="size-4" />
                            ) : null}
                        </Button>
                    );
                },
                size: 150,
                sortingFn: (rowA, rowB) => {
                    const expiresA = getTokenExpirationDate(rowA.original);
                    const expiresB = getTokenExpirationDate(rowB.original);

                    return expiresA.getTime() - expiresB.getTime();
                },
            },
            {
                accessorKey: 'createdAt',
                cell: ({ row }) => {
                    const token = row.original;
                    const isCreating = token.id === 'create-new';

                    if (isCreating) {
                        return <div className="text-muted-foreground text-sm">N/A</div>;
                    }

                    const dateString = row.getValue('createdAt') as string;

                    return (
                        <Tooltip>
                            <TooltipTrigger asChild>
                                <div className="cursor-default text-sm">{formatDateTime(dateString)}</div>
                            </TooltipTrigger>
                            <TooltipContent>
                                <div className="text-xs">{formatFullDateTime(dateString)}</div>
                            </TooltipContent>
                        </Tooltip>
                    );
                },
                header: ({ column }) => {
                    const sorted = column.getIsSorted();

                    return (
                        <Button
                            className="text-muted-foreground hover:text-primary flex items-center gap-2 p-0 no-underline hover:no-underline"
                            onClick={() => handleColumnSort(column)}
                            variant="link"
                        >
                            Created
                            {sorted === 'asc' ? (
                                <ArrowDown className="size-4" />
                            ) : sorted === 'desc' ? (
                                <ArrowUp className="size-4" />
                            ) : null}
                        </Button>
                    );
                },
                size: 120,
                sortingFn: (rowA, rowB) => {
                    const dateA = new Date(rowA.getValue('createdAt') as string);
                    const dateB = new Date(rowB.getValue('createdAt') as string);

                    return dateA.getTime() - dateB.getTime();
                },
            },
            {
                cell: ({ row }) => {
                    const token = row.original;
                    const isCreating = token.id === 'create-new';
                    const isEditing = editingTokenId === token.tokenId;

                    if (isCreating) {
                        return (
                            <div className="flex justify-end gap-1">
                                <Button
                                    className="size-8 p-0"
                                    disabled={isCreateLoading || !createFormData.expiresAt}
                                    onClick={handleCreate}
                                    variant="ghost"
                                >
                                    {isCreateLoading ? (
                                        <Loader2 className="size-4 animate-spin" />
                                    ) : (
                                        <Check className="size-4" />
                                    )}
                                </Button>
                                <Button
                                    className="size-8 p-0"
                                    onClick={handleCancelCreate}
                                    variant="ghost"
                                >
                                    <X className="size-4" />
                                </Button>
                            </div>
                        );
                    }

                    if (isEditing) {
                        return (
                            <div className="flex justify-end gap-1">
                                <Button
                                    className="size-8 p-0"
                                    disabled={isUpdateLoading}
                                    onClick={() => handleSave(token.tokenId)}
                                    variant="ghost"
                                >
                                    {isUpdateLoading ? (
                                        <Loader2 className="size-4 animate-spin" />
                                    ) : (
                                        <Check className="size-4" />
                                    )}
                                </Button>
                                <Button
                                    className="size-8 p-0"
                                    onClick={handleCancelEdit}
                                    variant="ghost"
                                >
                                    <X className="size-4" />
                                </Button>
                            </div>
                        );
                    }

                    return (
                        <div className="flex justify-end opacity-0 transition-opacity group-hover:opacity-100">
                            <DropdownMenu>
                                <DropdownMenuTrigger asChild>
                                    <Button
                                        className="size-8 p-0"
                                        variant="ghost"
                                    >
                                        <span className="sr-only">Open menu</span>
                                        <MoreHorizontal className="size-4" />
                                    </Button>
                                </DropdownMenuTrigger>
                                <DropdownMenuContent
                                    align="end"
                                    className="min-w-24"
                                >
                                    <DropdownMenuItem onClick={() => handleEdit(token)}>
                                        <Pencil className="size-3" />
                                        Edit
                                    </DropdownMenuItem>
                                    <DropdownMenuSeparator />
                                    <DropdownMenuItem
                                        disabled={isDeleteLoading && deletingToken?.tokenId === token.tokenId}
                                        onClick={() => handleDeleteDialogOpen(token)}
                                    >
                                        {isDeleteLoading && deletingToken?.tokenId === token.tokenId ? (
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
                size: 48,
            },
        ],
        [
            createFormData.expiresAt,
            createFormData.name,
            deletingToken,
            editFormData.name,
            editFormData.status,
            editingTokenId,
            handleCancelCreate,
            handleCancelEdit,
            handleColumnSort,
            handleCopyTokenId,
            handleCreate,
            handleDeleteDialogOpen,
            handleEdit,
            handleSave,
            isCreateLoading,
            isDeleteLoading,
            isUpdateLoading,
        ],
    );

    if (isLoading) {
        return (
            <div className="flex flex-col gap-4">
                <SettingsAPITokensHeader onCreateClick={handleCreateNew} />
                <StatusCard
                    description="Please wait while we fetch your API tokens"
                    icon={<Loader2 className="text-muted-foreground size-16 animate-spin" />}
                    title="Loading tokens..."
                />
            </div>
        );
    }

    if (error) {
        return (
            <div className="flex flex-col gap-4">
                <SettingsAPITokensHeader onCreateClick={handleCreateNew} />
                <Alert variant="destructive">
                    <AlertCircle className="size-4" />
                    <AlertTitle>Error loading tokens</AlertTitle>
                    <AlertDescription>{error.message}</AlertDescription>
                </Alert>
            </div>
        );
    }

    const tokens = data?.apiTokens || [];

    if (tokens.length === 0 && !creatingToken) {
        return (
            <div className="flex flex-col gap-4">
                <SettingsAPITokensHeader onCreateClick={handleCreateNew} />
                <StatusCard
                    action={
                        <Button
                            onClick={handleCreateNew}
                            variant="secondary"
                        >
                            <Plus className="size-4" />
                            Create Token
                        </Button>
                    }
                    description="Create your first API token to access PentAGI programmatically"
                    icon={<Key className="text-muted-foreground size-8" />}
                    title="No API tokens configured"
                />
            </div>
        );
    }

    return (
        <div className="flex flex-col gap-4">
            <SettingsAPITokensHeader onCreateClick={handleCreateNew} />

            {(createError || updateError || deleteError || deleteErrorMessage) && (
                <Alert variant="destructive">
                    <AlertCircle className="size-4" />
                    <AlertTitle>Error</AlertTitle>
                    <AlertDescription>
                        {createError?.message || updateError?.message || deleteError?.message || deleteErrorMessage}
                    </AlertDescription>
                </Alert>
            )}

            <DataTable<APIToken>
                columns={columns}
                columnVisibility={columnVisibility}
                data={creatingToken ? [createNewTokenPlaceholder, ...tokens] : tokens}
                filterColumn="name"
                filterPlaceholder="Filter token names..."
                onColumnVisibilityChange={(visibility) => {
                    Object.entries(visibility).forEach(([columnId, isVisible]) => {
                        if (columnVisibility[columnId] !== isVisible) {
                            updateColumnVisibility(columnId, isVisible);
                        }
                    });
                }}
                onPageChange={handlePageChange}
                pageIndex={currentPage}
                tableKey="api-tokens"
            />

            <Dialog
                onOpenChange={setShowTokenDialog}
                open={showTokenDialog}
            >
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>API Token Created</DialogTitle>
                        <DialogDescription>
                            Copy this token now. You won't be able to see it again for security reasons.
                        </DialogDescription>
                    </DialogHeader>
                    <div className="bg-muted rounded p-4">
                        <code className="text-sm break-all">{tokenSecret}</code>
                    </div>
                    <div className="flex gap-2">
                        <Button
                            className="flex-1"
                            onClick={async () => {
                                if (tokenSecret) {
                                    const success = await copyToClipboard(tokenSecret);

                                    if (success) {
                                        toast.success('Token copied to clipboard');
                                    } else {
                                        toast.error('Failed to copy token to clipboard');
                                    }
                                }
                            }}
                            variant="secondary"
                        >
                            <Copy className="size-4" />
                            Copy Token
                        </Button>
                        <Button
                            className="flex-1"
                            onClick={() => {
                                setShowTokenDialog(false);
                                setTokenSecret(null);
                            }}
                            variant="outline"
                        >
                            Close
                        </Button>
                    </div>
                </DialogContent>
            </Dialog>

            <ConfirmationDialog
                cancelText="Cancel"
                confirmText="Delete"
                handleConfirm={() => handleDelete(deletingToken?.tokenId)}
                handleOpenChange={setIsDeleteDialogOpen}
                isOpen={isDeleteDialogOpen}
                itemName={deletingToken?.name || deletingToken?.tokenId}
                itemType="token"
            />
        </div>
    );
};

export default SettingsAPITokens;
