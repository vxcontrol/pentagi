import type { ColumnDef } from '@tanstack/react-table';

import { useMutation, useQuery, useSubscription } from '@apollo/client/react';
import { format } from 'date-fns';
import { enUS, tr as trLocale } from 'date-fns/locale';
import { CalendarIcon, Check, Copy, Ellipsis, ExternalLink, Key, Pencil, Plus, Trash, X } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { useCallback, useId, useMemo, useState } from 'react';
import { type Control, Controller, useFormState } from 'react-hook-form';
import { toast } from 'sonner';
import * as z from 'zod';

import type { ApiTokenFragmentFragment } from '@/graphql/types';

import {
    AppHeader,
    AppHeaderAction,
    AppHeaderActions,
    AppHeaderContent,
    AppHeaderTitle,
} from '@/components/layouts/app/app-header';
import ConfirmationDialog from '@/components/shared/confirmation-dialog';
import { ErrorState } from '@/components/shared/error-state';
import { LoadingState } from '@/components/shared/loading-state';
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
import { Empty, EmptyContent, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';
import { Input } from '@/components/ui/input';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Spinner } from '@/components/ui/spinner';
import {
    ApiTokenCreatedDocument,
    ApiTokenDeletedDocument,
    ApiTokensDocument,
    ApiTokenUpdatedDocument,
    CreateApiTokenDocument,
    DeleteApiTokenDocument,
    TokenStatus as TokenStatusEnum,
    UpdateApiTokenDocument,
} from '@/graphql/types';
import { useAppForm } from '@/hooks/use-app-form';
import { useTableState } from '@/hooks/use-table-state';
import { cn } from '@/lib/utils';
import { formatDate } from '@/lib/utils/format';
import { baseUrl } from '@/models/api';

type APIToken = ApiTokenFragmentFragment;

// 100 mirrors the backend cap (server/models/api_tokens.go + the createAPIToken/updateAPIToken resolvers).
export const tokenNameSchema = z.string().trim().max(100, 'Token name must be 100 characters or less').default('');

const createTokenFormSchema = z.object({
    // Nullable in form input (the date picker starts empty); the refine gates
    // `formState.isValid` so the Create button stays disabled until a date is set.
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

type CreateTokenFormInput = z.input<typeof createTokenFormSchema>;
type CreateTokenFormValues = z.output<typeof createTokenFormSchema>;
type EditTokenFormInput = z.input<typeof editTokenFormSchema>;
type EditTokenFormValues = z.output<typeof editTokenFormSchema>;

const CREATE_TOKEN_DEFAULTS: CreateTokenFormInput = { expiresAt: null, name: '' };
const EDIT_TOKEN_DEFAULTS: EditTokenFormInput = { name: '', status: TokenStatusEnum.Active };

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
    t: (key: string) => string,
): { label: string; variant: 'default' | 'destructive' | 'outline' | 'secondary' } => {
    const expired = isTokenExpired(token);

    if (expired) {
        return { label: t('apiTokens.expired'), variant: 'destructive' };
    }

    if (token.status === 'active') {
        return { label: t('apiTokens.active'), variant: 'default' };
    }

    if (token.status === 'revoked') {
        return { label: t('apiTokens.revoked'), variant: 'outline' };
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
    control: Control<CreateTokenFormInput>;
    isLoading: boolean;
    onCancel: () => void;
    onSubmit: () => void;
}) {
    const { isValid } = useFormState({ control });
    const { t } = useTranslation();

    return (
        <div className="flex justify-end">
            <Button
                aria-label={isLoading ? t('apiTokens.submitting') : t('apiTokens.submit')}
                className="shrink-0"
                disabled={isLoading || !isValid}
                onClick={onSubmit}
                size="icon-sm"
                variant="ghost"
            >
                {isLoading ? <Spinner variant="circle" /> : <Check />}
            </Button>
            <Button
                aria-label={t('common.cancel')}
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
    control: Control<EditTokenFormInput>;
    isLoading: boolean;
    onCancel: () => void;
    onSubmit: () => void;
}) {
    const { isValid } = useFormState({ control });
    const { t } = useTranslation();

    return (
        <div className="flex justify-end">
            <Button
                aria-label={isLoading ? t('apiTokens.submitting') : t('apiTokens.submit')}
                className="shrink-0"
                disabled={isLoading || !isValid}
                onClick={onSubmit}
                size="icon-sm"
                variant="ghost"
            >
                {isLoading ? <Spinner variant="circle" /> : <Check />}
            </Button>
            <Button
                aria-label={t('common.cancel')}
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
    const { i18n, t } = useTranslation();
    const dateLocale = i18n.language === 'tr' ? trLocale : enUS;
    const { data, error, loading: isLoading, refetch } = useQuery(ApiTokensDocument);
    const [createAPIToken, { loading: isCreateLoading }] = useMutation(CreateApiTokenDocument);
    const [updateAPIToken, { loading: isUpdateLoading }] = useMutation(UpdateApiTokenDocument);
    const [deleteAPIToken, { loading: isDeleteLoading }] = useMutation(DeleteApiTokenDocument);

    const [editingTokenId, setEditingTokenId] = useState<null | string>(null);
    const [creatingToken, setCreatingToken] = useState(false);
    const [tokenSecret, setTokenSecret] = useState<null | string>(null);
    const [showTokenDialog, setShowTokenDialog] = useState(false);
    const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
    const [deletingToken, setDeletingToken] = useState<APIToken | null>(null);

    // Stable per-instance ids — keep label/for and a11y warnings clean even when
    // the row re-mounts due to subscription-driven refetches.
    const createNameFieldId = useId();
    const editNameFieldId = useId();

    // Form state lives in the parent so that subscription-driven DataTable
    // re-renders and row remounts cannot drop user input. <Controller> in each
    // cell re-subscribes to this state on remount — no values are lost.
    const createForm = useAppForm<CreateTokenFormInput, unknown, CreateTokenFormValues>({
        defaultValues: CREATE_TOKEN_DEFAULTS,
        schema: createTokenFormSchema,
    });
    const editForm = useAppForm<EditTokenFormInput, unknown, EditTokenFormValues>({
        defaultValues: EDIT_TOKEN_DEFAULTS,
        schema: editTokenFormSchema,
    });

    const { filter, pageIndex: currentPage, setFilter, setPage: handlePageChange } = useTableState();

    useSubscription(ApiTokenCreatedDocument, {
        onData: ({ client }) => {
            client.refetchQueries({ include: ['apiTokens'] });
        },
    });

    useSubscription(ApiTokenUpdatedDocument, {
        onData: ({ client }) => {
            client.refetchQueries({ include: ['apiTokens'] });
        },
    });

    useSubscription(ApiTokenDeletedDocument, {
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
                            name: values.name?.trim() || null,
                            status: values.status,
                        },
                        tokenId,
                    },
                });

                setEditingTokenId(null);
                editForm.reset(EDIT_TOKEN_DEFAULTS);
            } catch (error) {
                toast.error(t('apiTokens.updateTokenFailed'), {
                    description: error instanceof Error ? error.message : undefined,
                });
            }
        },
        [editForm, t, updateAPIToken],
    );

    const handleCreateNew = useCallback(() => {
        setCreatingToken(true);
        createForm.reset(CREATE_TOKEN_DEFAULTS);
        // The create row is prepended at data index 0, so an active filter or a non-first
        // page would hide it. Clearing the filter also resets pageIndex to 0.
        setFilter('');
    }, [createForm, setFilter]);

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
                        name: values.name?.trim() || null,
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
            toast.error(t('apiTokens.createTokenFailed'), {
                description: error instanceof Error ? error.message : undefined,
            });
        }
    }, [createAPIToken, createForm, t]);

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
                await deleteAPIToken({
                    refetchQueries: ['apiTokens'],
                    variables: { tokenId },
                });

                setDeletingToken(null);
            } catch (error) {
                toast.error(t('apiTokens.deleteTokenFailed'), {
                    description: error instanceof Error ? error.message : undefined,
                });
            }
        },
        [deleteAPIToken, t],
    );

    const handleCopyTokenId = useCallback(
        async (tokenId: string) => {
            const success = await copyToClipboard(tokenId);

            if (success) {
                toast.success(t('apiTokens.tokenIdCopied'));

                return;
            }

            toast.error(t('apiTokens.tokenIdCopyFailed'));
        },
        [t],
    );

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
                                        placeholder={t('apiTokens.tokenNamePlaceholder')}
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
                                        placeholder={t('apiTokens.tokenNamePlaceholder')}
                                    />
                                )}
                            />
                        );
                    }

                    return (
                        <div className="font-medium">
                            {token.name || (
                                <span className="text-muted-foreground font-normal italic">{t('apiTokens.unnamed')}</span>
                            )}
                        </div>
                    );
                },
                enableHiding: false,
                header: ({ column }) => (
                    <DataTableColumnHeader
                        column={column}
                        title={t('apiTokens.tokenName')}
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
                        return <div className="text-muted-foreground text-sm">{t('common.notAvailable')}</div>;
                    }

                    const tokenId = row.getValue('tokenId') as string;

                    return (
                        <div className="flex items-center gap-2">
                            <code className="text-sm">{tokenId}</code>
                            <Button
                                aria-label={t('apiTokens.copyTokenId')}
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
                        title={t('apiTokens.tokenId')}
                    />
                ),
                meta: { columnMenuLabel: t('apiTokens.tokenId'), searchable: true },
                size: 200,
            },
            {
                accessorKey: 'status',
                cell: ({ row }) => {
                    const token = row.original;
                    const isCreating = token.id === 'create-new';

                    if (isCreating) {
                        return <Badge variant="default">{t('apiTokens.active')}</Badge>;
                    }

                    const isEditing = editingTokenId === token.tokenId;
                    const expired = isTokenExpired(token);
                    const statusDisplay = getStatusDisplay(token, t);

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
                                                <SelectItem value={TokenStatusEnum.Active}>{t('apiTokens.active')}</SelectItem>
                                                <SelectItem value={TokenStatusEnum.Revoked}>{t('apiTokens.revoked')}</SelectItem>
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
                        title={t('apiTokens.status')}
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
                                                    format(field.value, 'd MMM yyyy', { locale: dateLocale })
                                                ) : (
                                                    <span>{t('apiTokens.pickDate')}</span>
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
                        title={t('apiTokens.expires')}
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
                        return <div className="text-muted-foreground text-sm">{t('common.notAvailable')}</div>;
                    }

                    const dateString = row.getValue('createdAt') as string;

                    return <div className="text-sm">{formatDate(new Date(dateString))}</div>;
                },
                header: ({ column }) => (
                    <DataTableColumnHeader
                        column={column}
                        title={t('apiTokens.createdAt')}
                    />
                ),
                meta: { columnMenuLabel: t('apiTokens.createdAt') },
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
                                        aria-label={t('apiTokens.openMenu')}
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
                                        {t('common.edit')}
                                    </DropdownMenuItem>
                                    <DropdownMenuItem onClick={() => handleCopyTokenId(token.tokenId)}>
                                        <Copy />
                                        {t('apiTokens.copyTokenId')}
                                    </DropdownMenuItem>
                                    <DropdownMenuSeparator />
                                    <DropdownMenuItem
                                        disabled={isDeleteLoading && deletingToken?.tokenId === token.tokenId}
                                        onClick={() => handleDeleteDialogOpen(token)}
                                    >
                                        {isDeleteLoading && deletingToken?.tokenId === token.tokenId ? (
                                            <>
                                                <Spinner variant="circle" />
                                                {t('apiTokens.deleting')}
                                            </>
                                        ) : (
                                            <>
                                                <Trash />
                                                {t('common.delete')}
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
                        {t('common.edit')}
                    </ContextMenuItem>
                    <ContextMenuItem onClick={() => handleCopyTokenId(token.tokenId)}>
                        <Copy />
                        {t('apiTokens.copyTokenId')}
                    </ContextMenuItem>
                    <ContextMenuSeparator />
                    <ContextMenuItem
                        disabled={isDeleteLoading && deletingToken?.tokenId === token.tokenId}
                        onClick={() => handleDeleteDialogOpen(token)}
                    >
                        <Trash />
                        {isDeleteLoading && deletingToken?.tokenId === token.tokenId
                            ? t('apiTokens.deleting')
                            : t('common.delete')}
                    </ContextMenuItem>
                </>
            );
        },
        [deletingToken, handleCopyTokenId, handleDeleteDialogOpen, handleEdit, isDeleteLoading, t],
    );

    const pageHeader = (
        <AppHeader>
            <AppHeaderContent>
                <AppHeaderTitle icon={<Key className="size-4 shrink-0" />}>{t('apiTokens.title')}</AppHeaderTitle>
            </AppHeaderContent>
            <AppHeaderActions>
                <AppHeaderAction
                    icon={<Plus />}
                    label={t('apiTokens.createToken')}
                    onClick={handleCreateNew}
                    variant="secondary"
                />
                <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                        <Button
                            aria-label={t('apiTokens.developerTools')}
                            className="size-8 p-0"
                            size="sm"
                            variant="ghost"
                        >
                            <Ellipsis />
                        </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent
                        align="end"
                        className="min-w-44"
                    >
                        <DropdownMenuItem asChild>
                            <a
                                href={`${baseUrl}/graphql/playground`}
                                rel="noopener noreferrer"
                                target="_blank"
                            >
                                <ExternalLink />
                                {t('apiTokens.graphqlPlayground')}
                            </a>
                        </DropdownMenuItem>
                        <DropdownMenuItem asChild>
                            <a
                                href={`${baseUrl}/swagger/index.html`}
                                rel="noopener noreferrer"
                                target="_blank"
                            >
                                <ExternalLink />
                                {t('apiTokens.swaggerUi')}
                            </a>
                        </DropdownMenuItem>
                    </DropdownMenuContent>
                </DropdownMenu>
            </AppHeaderActions>
        </AppHeader>
    );

    if (isLoading && !data) {
        return (
            <>
                {pageHeader}
                <div className="flex flex-1 flex-col gap-4 p-4">
                    <LoadingState
                        description={t('apiTokens.loadingTokensDesc')}
                        title={t('apiTokens.loadingTokens')}
                    />
                </div>
            </>
        );
    }

    // Error surface only when there's no data — a failed background refetch must not blank a working list.
    if (error && !data) {
        return (
            <>
                {pageHeader}
                <div className="flex flex-1 flex-col gap-4 p-4">
                    <ErrorState
                        message={error.message}
                        onRetry={refetch}
                        title={t('apiTokens.errorLoadingTokens')}
                    />
                </div>
            </>
        );
    }

    const tokens = data?.apiTokens || [];

    if (tokens.length === 0 && !creatingToken) {
        return (
            <>
                {pageHeader}
                <div className="flex flex-1 flex-col gap-4 p-4">
                    <Empty>
                        <EmptyHeader>
                            <EmptyMedia variant="icon">
                                <Key />
                            </EmptyMedia>
                            <EmptyTitle>{t('apiTokens.noTokensTitle')}</EmptyTitle>
                            <EmptyDescription>{t('apiTokens.noTokensDesc')}</EmptyDescription>
                        </EmptyHeader>
                        <EmptyContent>
                            <Button
                                onClick={handleCreateNew}
                                variant="secondary"
                            >
                                <Plus />
                                {t('apiTokens.createToken')}
                            </Button>
                        </EmptyContent>
                    </Empty>
                </div>
            </>
        );
    }

    return (
        <>
            {pageHeader}
            <div className="flex flex-1 flex-col gap-4 p-4">
                <DataTable<APIToken>
                    columns={columns}
                    data={creatingToken ? [createNewTokenPlaceholder, ...tokens] : tokens}
                    empty={{ entityName: t('apiTokens.title') }}
                    filterPlaceholder={t('apiTokens.filterTokens')}
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
                            <DialogTitle>{t('apiTokens.createTokenTitle')}</DialogTitle>
                            <DialogDescription>{t('apiTokens.createTokenDesc')}</DialogDescription>
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
                                            toast.success(t('apiTokens.tokenCopied'));
                                        } else {
                                            toast.error(t('apiTokens.tokenCopyFailed'));
                                        }
                                    }
                                }}
                                variant="secondary"
                            >
                                <Copy />
                                {t('apiTokens.copyToken')}
                            </Button>
                            <Button
                                className="flex-1"
                                onClick={() => {
                                    setShowTokenDialog(false);
                                    setTokenSecret(null);
                                }}
                                variant="outline"
                            >
                                {t('common.close')}
                            </Button>
                        </div>
                    </DialogContent>
                </Dialog>

                <ConfirmationDialog
                    cancelText={t('common.cancel')}
                    confirmText={t('common.delete')}
                    handleConfirm={() => handleDelete(deletingToken?.tokenId)}
                    handleOpenChange={setIsDeleteDialogOpen}
                    isOpen={isDeleteDialogOpen}
                    itemName={deletingToken?.name || deletingToken?.tokenId}
                    itemType={t('apiTokens.tokenNoun')}
                />
            </div>
        </>
    );
}

export default SettingsAPITokens;
