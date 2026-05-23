import type { ColumnDef, Row } from '@tanstack/react-table';

import { AlertCircle, ChevronDown, Copy, Ellipsis, Loader2, Pencil, Plus, Settings, Trash } from 'lucide-react';
import { useCallback, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';

import type { ProviderConfigFragmentFragment } from '@/graphql/types';

import Anthropic from '@/components/icons/anthropic';
import Bedrock from '@/components/icons/bedrock';
import Custom from '@/components/icons/custom';
import DeepSeek from '@/components/icons/deepseek';
import Gemini from '@/components/icons/gemini';
import GLM from '@/components/icons/glm';
import Kimi from '@/components/icons/kimi';
import Ollama from '@/components/icons/ollama';
import OpenAi from '@/components/icons/open-ai';
import Qwen from '@/components/icons/qwen';
import ConfirmationDialog from '@/components/shared/confirmation-dialog';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { ContextMenuItem, ContextMenuSeparator } from '@/components/ui/context-menu';
import { DataTable, DataTableColumnHeader } from '@/components/ui/data-table';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { StatusCard } from '@/components/ui/status-card';
import { ProviderType, useDeleteProviderMutation, useSettingsProvidersQuery } from '@/graphql/types';
import { useTableState } from '@/hooks/use-table-state';
import { formatDate } from '@/lib/utils/format';
type Provider = ProviderConfigFragmentFragment;

const providerIcons: Record<ProviderType, React.ComponentType<React.SVGProps<SVGSVGElement>>> = {
    [ProviderType.Anthropic]: Anthropic,
    [ProviderType.Bedrock]: Bedrock,
    [ProviderType.Custom]: Custom,
    [ProviderType.Deepseek]: DeepSeek,
    [ProviderType.Gemini]: Gemini,
    [ProviderType.Glm]: GLM,
    [ProviderType.Kimi]: Kimi,
    [ProviderType.Ollama]: Ollama,
    [ProviderType.Openai]: OpenAi,
    [ProviderType.Qwen]: Qwen,
};

const providerTypes = [
    { label: 'Anthropic', type: ProviderType.Anthropic },
    { label: 'Bedrock', type: ProviderType.Bedrock },
    { label: 'Custom', type: ProviderType.Custom },
    { label: 'DeepSeek', type: ProviderType.Deepseek },
    { label: 'Gemini', type: ProviderType.Gemini },
    { label: 'GLM', type: ProviderType.Glm },
    { label: 'Kimi', type: ProviderType.Kimi },
    { label: 'Ollama', type: ProviderType.Ollama },
    { label: 'OpenAI', type: ProviderType.Openai },
    { label: 'Qwen', type: ProviderType.Qwen },
];

function SettingsProviders() {
    const { data, error, loading: isLoading } = useSettingsProvidersQuery();
    const [deleteProvider, { error: deleteError, loading: isDeleteLoading }] = useDeleteProviderMutation();
    const [deleteErrorMessage, setDeleteErrorMessage] = useState<null | string>(null);
    const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
    const [deletingProvider, setDeletingProvider] = useState<null | Provider>(null);
    const navigate = useNavigate();

    const { filter, pageIndex: currentPage, setFilter, setPage: handlePageChange } = useTableState();

    const handleProviderDelete = useCallback(
        async (providerId: string | undefined) => {
            if (!providerId) {
                return;
            }

            try {
                setDeleteErrorMessage(null);

                await deleteProvider({
                    refetchQueries: ['settingsProviders'],
                    variables: { providerId: providerId.toString() },
                });

                setDeletingProvider(null);
                setDeleteErrorMessage(null);
            } catch (error) {
                setDeleteErrorMessage(error instanceof Error ? error.message : 'An error occurred while deleting');
            }
        },
        [deleteProvider],
    );

    const handleProviderEdit = useCallback(
        (providerId: string) => {
            navigate(`/settings/providers/${providerId}`);
        },
        [navigate],
    );

    const handleProviderClone = useCallback(
        (providerId: string) => {
            navigate(`/settings/providers/new?id=${providerId}`);
        },
        [navigate],
    );

    const handleProviderDeleteDialogOpen = useCallback((provider: Provider) => {
        setDeletingProvider(provider);
        setIsDeleteDialogOpen(true);
    }, []);

    const columns: ColumnDef<Provider>[] = useMemo(
        () => [
            {
                accessorKey: 'name',
                cell: ({ row }) => <div className="truncate font-medium">{row.getValue('name')}</div>,
                enableHiding: false,
                header: ({ column }) => (
                    <DataTableColumnHeader
                        column={column}
                        title="Name"
                    />
                ),
                // Name flexes to fill remaining width — fixed `size` would push
                // the Type column off-screen on narrow viewports (e.g. 375px).
                meta: { searchable: true },
            },
            {
                accessorKey: 'type',
                cell: ({ row }) => {
                    const providerType = row.getValue('type') as ProviderType;
                    const Icon = providerIcons[providerType];
                    const label = providerTypes.find((p) => p.type === providerType)?.label || providerType;

                    return (
                        <Badge
                            className="max-w-full whitespace-nowrap"
                            variant="outline"
                        >
                            {Icon && <Icon className="mr-1 size-3 shrink-0" />}
                            <span className="truncate">{label}</span>
                        </Badge>
                    );
                },
                header: ({ column }) => (
                    <DataTableColumnHeader
                        column={column}
                        title="Type"
                    />
                ),
                meta: { searchable: true },
                minSize: 110,
                size: 160,
            },
            {
                accessorKey: 'createdAt',
                cell: ({ row }) => {
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
                accessorKey: 'updatedAt',
                cell: ({ row }) => {
                    const dateString = row.getValue('updatedAt') as string;

                    return <div className="text-sm">{formatDate(new Date(dateString))}</div>;
                },
                header: ({ column }) => (
                    <DataTableColumnHeader
                        column={column}
                        title="Updated"
                    />
                ),
                size: 120,
                sortingFn: (rowA, rowB) => {
                    const dateA = new Date(rowA.getValue('updatedAt') as string);
                    const dateB = new Date(rowB.getValue('updatedAt') as string);

                    return dateA.getTime() - dateB.getTime();
                },
            },
            {
                cell: ({ row }) => {
                    const provider = row.original;

                    return (
                        <div className="flex justify-end opacity-0 transition-opacity group-hover:opacity-100">
                            <DropdownMenu>
                                <DropdownMenuTrigger asChild>
                                    <Button
                                        aria-label="Open menu"
                                        className="size-8 p-0"
                                        variant="ghost"
                                    >
                                        <Ellipsis />
                                    </Button>
                                </DropdownMenuTrigger>
                                <DropdownMenuContent
                                    align="end"
                                    className="min-w-24"
                                >
                                    <DropdownMenuItem onClick={() => handleProviderEdit(provider.id)}>
                                        <Pencil className="size-3" />
                                        Edit
                                    </DropdownMenuItem>
                                    <DropdownMenuItem onClick={() => handleProviderClone(provider.id)}>
                                        <Copy className="size-4" />
                                        Clone
                                    </DropdownMenuItem>
                                    <DropdownMenuSeparator />
                                    <DropdownMenuItem
                                        disabled={isDeleteLoading && deletingProvider?.id === provider.id}
                                        onClick={() => handleProviderDeleteDialogOpen(provider)}
                                    >
                                        {isDeleteLoading && deletingProvider?.id === provider.id ? (
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
                meta: { preventRowClick: true },
                size: 48,
            },
        ],
        [handleProviderClone, handleProviderDeleteDialogOpen, handleProviderEdit, isDeleteLoading, deletingProvider],
    );

    const renderSubComponent = ({ row }: { row: Row<Provider> }) => {
        const provider = row.original;
        const { agents } = provider;

        if (!agents) {
            return <div className="text-muted-foreground p-4 text-sm">No agent configuration available</div>;
        }

        const getName = (key: string): string =>
            key.replaceAll(/([A-Z])/g, ' $1').replace(/^./, (item) => item.toUpperCase());

        const getFields = (obj: unknown, prefix = ''): { label: string; value: boolean | number | string }[] => {
            if (!obj || typeof obj !== 'object') {
                return [];
            }

            return Object.entries(obj as Record<string, unknown>)
                .filter(([key, value]) => key !== '__typename' && !!value)
                .flatMap(([key, value]) => {
                    const label = `${prefix ? `${prefix} ` : ''}${getName(key)}`;

                    return typeof value === 'object'
                        ? getFields(value, label)
                        : [{ label, value: value as boolean | number | string }];
                });
        };

        const agentTypes = Object.entries(agents)
            .filter(([key]) => key !== '__typename')
            .map(([key, data]) => ({
                data,
                key,
                name: getName(key),
            }))
            .sort((a, b) => a.name.localeCompare(b.name));

        return (
            <div className="bg-muted/20 border-t p-4">
                <h4 className="font-medium">Agent Configurations</h4>
                <hr className="border-muted-foreground/20 my-4" />
                <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 2xl:grid-cols-5">
                    {agentTypes.map(({ data, key, name }) => {
                        const fields = data ? getFields(data) : [];

                        return (
                            <div
                                className="flex flex-col gap-2"
                                key={key}
                            >
                                <div className="text-sm font-medium">{name}</div>
                                {fields.length > 0 ? (
                                    <div className="flex flex-col gap-1 text-sm">
                                        {fields.map(({ label, value }) => (
                                            <div key={label}>
                                                <span className="text-muted-foreground">{label}:</span> {value}
                                            </div>
                                        ))}
                                    </div>
                                ) : (
                                    <div className="text-muted-foreground text-sm">No configuration available</div>
                                )}
                            </div>
                        );
                    })}
                </div>
            </div>
        );
    };

    const renderRowContextMenu = useCallback(
        (provider: Provider) => (
            <>
                <ContextMenuItem onClick={() => handleProviderEdit(provider.id)}>
                    <Pencil />
                    Edit
                </ContextMenuItem>
                <ContextMenuItem onClick={() => handleProviderClone(provider.id)}>
                    <Copy />
                    Clone
                </ContextMenuItem>
                <ContextMenuSeparator />
                <ContextMenuItem
                    disabled={isDeleteLoading && deletingProvider?.id === provider.id}
                    onClick={() => handleProviderDeleteDialogOpen(provider)}
                >
                    <Trash />
                    {isDeleteLoading && deletingProvider?.id === provider.id ? 'Deleting...' : 'Delete'}
                </ContextMenuItem>
            </>
        ),
        [deletingProvider, handleProviderClone, handleProviderDeleteDialogOpen, handleProviderEdit, isDeleteLoading],
    );

    if (isLoading) {
        return (
            <div className="flex flex-col gap-4">
                <SettingsProvidersHeader />
                <StatusCard
                    description="Please wait while we fetch your provider configurations"
                    icon={<Loader2 className="text-muted-foreground size-16 animate-spin" />}
                    title="Loading providers..."
                />
            </div>
        );
    }

    if (error) {
        return (
            <div className="flex flex-col gap-4">
                <SettingsProvidersHeader />
                <Alert variant="destructive">
                    <AlertCircle className="size-4" />
                    <AlertTitle>Error loading providers</AlertTitle>
                    <AlertDescription>{error.message}</AlertDescription>
                </Alert>
            </div>
        );
    }

    const providers = data?.settingsProviders?.userDefined || [];

    if (providers.length === 0) {
        return (
            <div className="flex flex-col gap-4">
                <SettingsProvidersHeader />
                <StatusCard
                    action={
                        <Button
                            onClick={() => navigate('/settings/providers/new')}
                            variant="secondary"
                        >
                            <Plus className="size-4" />
                            Add Provider
                        </Button>
                    }
                    description="Get started by adding your first language model provider"
                    icon={<Settings className="text-muted-foreground size-8" />}
                    title="No providers configured"
                />
            </div>
        );
    }

    return (
        <div className="flex flex-col gap-4">
            <SettingsProvidersHeader />

            {/* Delete Error Alert */}
            {(deleteError || deleteErrorMessage) && (
                <Alert variant="destructive">
                    <AlertCircle className="size-4" />
                    <AlertTitle>Error deleting provider</AlertTitle>
                    <AlertDescription>{deleteError?.message || deleteErrorMessage}</AlertDescription>
                </Alert>
            )}

            <DataTable<Provider>
                columns={columns}
                data={providers}
                empty={{ entityName: 'providers' }}
                filterPlaceholder="Filter providers..."
                filterValue={filter}
                onFilterChange={setFilter}
                onPageChange={handlePageChange}
                pageIndex={currentPage}
                renderRowContextMenu={renderRowContextMenu}
                renderSubComponent={renderSubComponent}
            />

            <ConfirmationDialog
                cancelText="Cancel"
                confirmText="Delete"
                handleConfirm={() => handleProviderDelete(deletingProvider?.id)}
                handleOpenChange={setIsDeleteDialogOpen}
                isOpen={isDeleteDialogOpen}
                itemName={deletingProvider?.name}
                itemType="provider"
            />
        </div>
    );
}

function SettingsProvidersHeader() {
    const navigate = useNavigate();

    const handleProviderCreate = (providerType: string) => {
        navigate(`/settings/providers/new?type=${providerType}`);
    };

    return (
        <div className="flex items-center justify-between gap-4">
            <p className="text-muted-foreground min-w-0 flex-1 truncate">Manage language model providers</p>

            {/*
             * "Create Provider" is a dropdown trigger, not a submit-style action — it
             * opens a menu listing provider types (OpenAI, Anthropic, Custom, …). The
             * `<ChevronDown />` icon plus Radix's `aria-haspopup="menu"` already signal
             * "menu opens" to sighted and AT users; the explicit aria-label adds the
             * intent ("create provider") so screen readers don't just announce
             * "Create Provider, menu" but "Create provider, choose type, menu".
             */}
            <DropdownMenu>
                <DropdownMenuTrigger asChild>
                    <Button
                        aria-label="Create provider — choose type"
                        className="shrink-0"
                        variant="secondary"
                    >
                        Create Provider
                        <ChevronDown className="size-4" />
                    </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent
                    align="end"
                    style={{
                        width: 'var(--radix-dropdown-menu-trigger-width)',
                    }}
                >
                    {providerTypes.map(({ label, type }) => {
                        const Icon = providerIcons[type];

                        return (
                            <DropdownMenuItem
                                key={type}
                                onClick={() => handleProviderCreate(type)}
                            >
                                {Icon && <Icon className="size-4" />}
                                {label}
                            </DropdownMenuItem>
                        );
                    })}
                </DropdownMenuContent>
            </DropdownMenu>
        </div>
    );
}

export default SettingsProviders;
