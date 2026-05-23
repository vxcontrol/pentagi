import type { ColumnDef } from '@tanstack/react-table';

import { zodResolver } from '@hookform/resolvers/zod';
import { format } from 'date-fns';
import { enUS } from 'date-fns/locale';
import {
    AlertCircle,
    CalendarIcon,
    Check,
    Copy,
    Ellipsis,
    ExternalLink,
    Key,
    Loader2,
    Pencil,
    Plus,
    Trash,
    X,
} from 'lucide-react';
import { useCallback, useId, useMemo, useState } from 'react';
import { type Control, Controller, useForm, useFormState } from 'react-hook-form';
import { toast } from 'sonner';
import * as z from 'zod';

import type { ApiTokenFragmentFragment } from '@/graphql/types';

import ConfirmationDialog from '@/components/shared/confirmation-dialog';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Calendar } from '@/components/ui/calendar';
import { ContextMenuItem, ContextMenuSeparator } from '@/components/ui/context-menu';
import { DataTable, DataTableColumnHeader } from '@/components/ui/data-table';
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
import { useTableState } from '@/hooks/use-table-state';
import { cn } from '@/lib/utils';
import { formatDate } from '@/lib/utils/format';
import { baseUrl } from '@/models/api';

type APIToken = ApiTokenFragmentFragment;

const tokenNameSchema = z.string().trim().max(255, 'Token name must be 255 characters or less').default('');

const createTokenFormSchema = z.object({
    // Nullable in the form state (the date picker starts empty) but required
    // for submission — the refine drives `formState.isValid`, which gates the
    // Create button without manual checks. Kept as `Date | null` in the
    // inferred type so `defaultValues` can be `null` without a cast.
    expiresAt: z
        .date()
        .nullable()
        .refine((value) => value !== null, { message: 'Expiration date is required' }),
    name: tokenNameSchema,
});

const editTokenFormSchema = z.object({
    name: tokenNameSchema,
    status: z.nativeEnum(TokenStatusEnum),
});

type CreateTokenFormValues = z.infer<typeof createTokenFormSchema>;
type EditTokenFormValues = z.infer<typeof editTokenFormSchema>;

const CREATE_TOKEN_DEFAULTS: CreateTokenFormValues = { expiresAt: null, name: '' };
const EDIT_TOKEN_DEFAULTS: EditTokenFormValues = { name: '', status: TokenStatusEnum.Active };

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

function SettingsAPITokensHeader({ onCreateClick }: { onCreateClick: () => void }) {
    return (
        <div className="flex items-center justify-between gap-4">
            <div className="flex min-w-0 flex-1 flex-col gap-2">
                <p className="text-muted-foreground truncate">Manage API tokens for programmatic access</p>
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
                className="shrink-0"
                onClick={onCreateClick}
                variant="secondary"
            >
                <Plus className="size-4" />
                Create Token
            </Button>
        </div>
    );
}

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

// Inline-row action buttons live in their own components so the validity
// subscription via `useFormState` re-renders only this small subtree on form
// changes, not the entire `SettingsAPITokens` parent / DataTable.
function CreateRowActions({
    control,
    isLoading,
    onCancel,
    onSubmit,
}: {
    control: Control<CreateTokenFormValues>;
    isLoading: boolean;
    onCancel: () => void;
    onSubmit: () => void;
}) {
    const { isValid } = useFormState({ control });

    return (
        <div className="flex justify-end">
            <Button
                aria-label={isLoading ? 'Submitting…' : 'Submit'}
                className="shrink-0"
                disabled={isLoading || !isValid}
                onClick={onSubmit}
                size="icon-sm"
                variant="ghost"
            >
                {isLoading ? <Loader2 className="animate-spin" /> : <Check />}
            </Button>
            <Button
                aria-label="Cancel"
                className="shrink-0"
                onClick={onCancel}
                size="icon-sm"
                variant="ghost"
            >
                <X />
            </Button>
        </div>
    );
}

function EditRowActions({
    control,
    isLoading,
    onCancel,
    onSubmit,
}: {
    control: Control<EditTokenFormValues>;
    isLoading: boolean;
    onCancel: () => void;
    onSubmit: () => void;
}) {
    const { isValid } = useFormState({ control });

    return (
        <div className="flex justify-end">
            <Button
                aria-label={isLoading ? 'Submitting…' : 'Submit'}
                className="shrink-0"
                disabled={isLoading || !isValid}
                onClick={onSubmit}
                size="icon-sm"
                variant="ghost"
            >
                {isLoading ? <Loader2 className="animate-spin" /> : <Check />}
            </Button>
            <Button
                aria-label="Cancel"
                className="shrink-0"
                onClick={onCancel}
                size="icon-sm"
                variant="ghost"
            >
                <X />
            </Button>
        </div>
    );
}

function SettingsAPITokens() {
    const { data, error, loading: isLoading } = useApiTokensQuery();
    const [createAPIToken, { error: createError, loading: isCreateLoading }] = useCreateApiTokenMutation();
    const [updateAPIToken, { error: updateError, loading: isUpdateLoading }] = useUpdateApiTokenMutation();
    const [deleteAPIToken, { error: deleteError, loading: isDeleteLoading }] = useDeleteApiTokenMutation();

    const [editingTokenId, setEditingTokenId] = useState<null | string>(null);
    const [creatingToken, setCreatingToken] = useState(false);
    const [tokenSecret, setTokenSecret] = useState<null | string>(null);
    const [showTokenDialog, setShowTokenDialog] = useState(false);
    const [deleteErrorMessage, setDeleteErrorMessage] = useState<null | string>(null);
    const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
    const [deletingToken, setDeletingToken] = useState<APIToken | null>(null);

    // Stable per-instance ids — keep label/for and a11y warnings clean even when
    // the row re-mounts due to subscription-driven refetches.
    const createNameFieldId = useId();
    const editNameFieldId = useId();

    // Form state lives in the parent so that subscription-driven DataTable
    // re-renders and row remounts cannot drop user input. <Controller> in each
    // cell re-subscribes to this state on remount — no values are lost.
    const createForm = useForm<CreateTokenFormValues>({
        defaultValues: CREATE_TOKEN_DEFAULTS,
        mode: 'onChange',
        resolver: zodResolver(createTokenFormSchema),
    });
    const editForm = useForm<EditTokenFormValues>({
        defaultValues: EDIT_TOKEN_DEFAULTS,
        mode: 'onChange',
        resolver: zodResolver(editTokenFormSchema),
    });

    const { filter, pageIndex: currentPage, setFilter, setPage: handlePageChange } = useTableState();

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

    const handleEdit = useCallback(
        (token: APIToken) => {
            setEditingTokenId(token.tokenId);
            editForm.reset({ name: token.name ?? '', status: token.status });
        },
        [editForm],
    );

    const handleCancelEdit = useCallback(() => {
        setEditingTokenId(null);
        editForm.reset(EDIT_TOKEN_DEFAULTS);
    }, [editForm]);

    const handleSave = useCallback(
        async (tokenId: string) => {
            const valid = await editForm.trigger();

            if (!valid) {
                return;
            }

            const values = editForm.getValues();

            try {
                await updateAPIToken({
                    refetchQueries: ['apiTokens'],
                    variables: {
                        input: {
                            name: values.name.trim() || null,
                            status: values.status,
                        },
                        tokenId,
                    },
                });

                setEditingTokenId(null);
                editForm.reset(EDIT_TOKEN_DEFAULTS);
            } catch (error) {
                console.error('Failed to update token:', error);
            }
        },
        [editForm, updateAPIToken],
    );

    const handleCreateNew = useCallback(() => {
        setCreatingToken(true);
        createForm.reset(CREATE_TOKEN_DEFAULTS);
    }, [createForm]);

    const handleCancelCreate = useCallback(() => {
        setCreatingToken(false);
        createForm.reset(CREATE_TOKEN_DEFAULTS);
    }, [createForm]);

    const handleCreate = useCallback(async () => {
        const valid = await createForm.trigger();

        if (!valid) {
            return;
        }

        const values = createForm.getValues();

        if (!values.expiresAt) {
            return;
        }

        try {
            const ttl = calculateTTL(values.expiresAt);
            const result = await createAPIToken({
                refetchQueries: ['apiTokens'],
                variables: {
                    input: {
                        name: values.name.trim() || null,
                        ttl,
                    },
                },
            });

            if (result.data?.createAPIToken) {
                setTokenSecret(result.data.createAPIToken.token);
                setShowTokenDialog(true);
            }

            setCreatingToken(false);
            createForm.reset(CREATE_TOKEN_DEFAULTS);
        } catch (error) {
            console.error('Failed to create token:', error);
        }
    }, [createAPIToken, createForm]);

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
                            <Controller
                                control={createForm.control}
                                name="name"
                                render={({ field }) => (
                                    <Input
                                        {...field}
                                        autoComplete="off"
                                        autoFocus
                                        className="h-8"
                                        id={createNameFieldId}
                                        placeholder="Token name (optional)"
                                    />
                                )}
                            />
                        );
                    }

                    if (isEditing) {
                        return (
                            <Controller
                                control={editForm.control}
                                name="name"
                                render={({ field }) => (
                                    <Input
                                        {...field}
                                        autoComplete="off"
                                        autoFocus
                                        className="h-8"
                                        id={editNameFieldId}
                                        placeholder="Token name (optional)"
                                    />
                                )}
                            />
                        );
                    }

                    return (
                        <div className="font-medium">
                            {token.name || <span className="text-muted-foreground font-normal italic">(unnamed)</span>}
                        </div>
                    );
                },
                enableHiding: false,
                header: ({ column }) => (
                    <DataTableColumnHeader
                        column={column}
                        title="Name"
                    />
                ),
                meta: { searchable: true },
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
                                aria-label="Copy token ID"
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
                header: ({ column }) => (
                    <DataTableColumnHeader
                        column={column}
                        title="Token ID"
                    />
                ),
                meta: { columnMenuLabel: 'Token ID', searchable: true },
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
                            <Controller
                                control={editForm.control}
                                name="status"
                                render={({ field }) => (
                                    <Select
                                        onValueChange={field.onChange}
                                        value={field.value}
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
                                )}
                            />
                        );
                    }

                    return <Badge variant={statusDisplay.variant}>{statusDisplay.label}</Badge>;
                },
                header: ({ column }) => (
                    <DataTableColumnHeader
                        column={column}
                        title="Status"
                    />
                ),
                meta: { searchable: true },
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
                            <Controller
                                control={createForm.control}
                                name="expiresAt"
                                render={({ field }) => (
                                    <Popover>
                                        <PopoverTrigger asChild>
                                            <Button
                                                className={cn(
                                                    'h-8 w-full justify-start text-left font-normal',
                                                    !field.value && 'text-muted-foreground',
                                                )}
                                                variant="outline"
                                            >
                                                <CalendarIcon className="mr-2 size-4" />
                                                {field.value ? (
                                                    format(field.value, 'd MMM yyyy', { locale: enUS })
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
                                                onSelect={(date) => field.onChange(date ?? null)}
                                                selected={field.value ?? undefined}
                                            />
                                        </PopoverContent>
                                    </Popover>
                                )}
                            />
                        );
                    }

                    const expiresAt = getTokenExpirationDate(token);
                    const expiresAtString = expiresAt.toISOString();

                    return <div className="text-sm">{formatDate(new Date(expiresAtString))}</div>;
                },
                header: ({ column }) => (
                    <DataTableColumnHeader
                        column={column}
                        title="Expires"
                    />
                ),
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

                    return <div className="text-sm">{formatDate(new Date(dateString))}</div>;
                },
                header: ({ column }) => (
                    <DataTableColumnHeader
                        column={column}
                        title="Created"
                    />
                ),
                meta: { columnMenuLabel: 'Created' },
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
                            <CreateRowActions
                                control={createForm.control}
                                isLoading={isCreateLoading}
                                onCancel={handleCancelCreate}
                                onSubmit={handleCreate}
                            />
                        );
                    }

                    if (isEditing) {
                        return (
                            <EditRowActions
                                control={editForm.control}
                                isLoading={isUpdateLoading}
                                onCancel={handleCancelEdit}
                                onSubmit={() => handleSave(token.tokenId)}
                            />
                        );
                    }

                    return (
                        <div className="flex justify-end opacity-0 transition-opacity group-hover:opacity-100">
                            <DropdownMenu>
                                <DropdownMenuTrigger asChild>
                                    <Button
                                        aria-label="Open menu"
                                        className="shrink-0"
                                        size="icon-sm"
                                        variant="ghost"
                                    >
                                        <Ellipsis />
                                    </Button>
                                </DropdownMenuTrigger>
                                <DropdownMenuContent
                                    align="end"
                                    className="min-w-24"
                                >
                                    <DropdownMenuItem onClick={() => handleEdit(token)}>
                                        <Pencil />
                                        Edit
                                    </DropdownMenuItem>
                                    <DropdownMenuItem onClick={() => handleCopyTokenId(token.tokenId)}>
                                        <Copy />
                                        Copy Token ID
                                    </DropdownMenuItem>
                                    <DropdownMenuSeparator />
                                    <DropdownMenuItem
                                        disabled={isDeleteLoading && deletingToken?.tokenId === token.tokenId}
                                        onClick={() => handleDeleteDialogOpen(token)}
                                    >
                                        {isDeleteLoading && deletingToken?.tokenId === token.tokenId ? (
                                            <>
                                                <Loader2 className="animate-spin" />
                                                Deleting...
                                            </>
                                        ) : (
                                            <>
                                                <Trash />
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
        [
            createForm.control,
            createNameFieldId,
            deletingToken,
            editForm.control,
            editNameFieldId,
            editingTokenId,
            handleCancelCreate,
            handleCancelEdit,
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

    const renderRowContextMenu = useCallback(
        (token: APIToken) => {
            if (token.id === 'create-new') {
                return null;
            }

            return (
                <>
                    <ContextMenuItem onClick={() => handleEdit(token)}>
                        <Pencil />
                        Edit
                    </ContextMenuItem>
                    <ContextMenuItem onClick={() => handleCopyTokenId(token.tokenId)}>
                        <Copy />
                        Copy Token ID
                    </ContextMenuItem>
                    <ContextMenuSeparator />
                    <ContextMenuItem
                        disabled={isDeleteLoading && deletingToken?.tokenId === token.tokenId}
                        onClick={() => handleDeleteDialogOpen(token)}
                    >
                        <Trash />
                        {isDeleteLoading && deletingToken?.tokenId === token.tokenId ? 'Deleting...' : 'Delete'}
                    </ContextMenuItem>
                </>
            );
        },
        [deletingToken, handleCopyTokenId, handleDeleteDialogOpen, handleEdit, isDeleteLoading],
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
                data={creatingToken ? [createNewTokenPlaceholder, ...tokens] : tokens}
                empty={{ entityName: 'API tokens' }}
                filterPlaceholder="Filter tokens..."
                filterValue={filter}
                onFilterChange={setFilter}
                onPageChange={handlePageChange}
                pageIndex={currentPage}
                renderRowContextMenu={renderRowContextMenu}
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
}

// Helper subcomponents so we can use useFormState/useWatch without subscribing
// the whole table to every keystroke. Each watches its own form's validity.

export default SettingsAPITokens;
