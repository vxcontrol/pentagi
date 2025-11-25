import type { ColumnDef } from '@tanstack/react-table';

import {
    AlertCircle,
    ArrowDown,
    ArrowUp,
    ChevronDown,
    Copy,
    Loader2,
    MoreHorizontal,
    Pencil,
    Plus,
    Settings,
    Trash,
} from 'lucide-react';
import { useState } from 'react';
import { useNavigate } from 'react-router-dom';

import type { ProviderConfigFragmentFragment } from '@/graphql/types';

import Anthropic from '@/components/icons/anthropic';
import Bedrock from '@/components/icons/bedrock';
import Custom from '@/components/icons/custom';
import Gemini from '@/components/icons/gemini';
import Ollama from '@/components/icons/ollama';
import OpenAi from '@/components/icons/open-ai';
import ConfirmationDialog from '@/components/shared/confirmation-dialog';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { DataTable } from '@/components/ui/data-table';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { StatusCard } from '@/components/ui/status-card';
import { ProviderType, useDeleteProviderMutation, useSettingsProvidersQuery } from '@/graphql/types';

type Provider = ProviderConfigFragmentFragment;

const providerIcons: Record<ProviderType, React.ComponentType<any>> = {
    [ProviderType.Anthropic]: Anthropic,
    [ProviderType.Bedrock]: Bedrock,
    [ProviderType.Custom]: Custom,
    [ProviderType.Gemini]: Gemini,
    [ProviderType.Ollama]: Ollama,
    [ProviderType.Openai]: OpenAi,
};

const providerTypes = [
    { label: 'Anthropic', type: ProviderType.Anthropic },
    { label: 'Bedrock', type: ProviderType.Bedrock },
    { label: 'Custom', type: ProviderType.Custom },
    { label: 'Gemini', type: ProviderType.Gemini },
    { label: 'Ollama', type: ProviderType.Ollama },
    { label: 'OpenAI', type: ProviderType.Openai },
];

const SettingsProvidersHeader = () => {
    const navigate = useNavigate();

    const handleProviderCreate = (providerType: string) => {
        navigate(`/settings/providers/new?type=${providerType}`);
    };

    return (
        <div className="flex items-center justify-between gap-4">
            <p className="text-muted-foreground">Manage language model providers</p>

            <DropdownMenu>
                <DropdownMenuTrigger asChild>
                    <Button variant="secondary">
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
};

const SettingsProviders = () => {
    const { data, error, loading: isLoading } = useSettingsProvidersQuery();
    const [deleteProvider, { error: deleteError, loading: isDeleteLoading }] = useDeleteProviderMutation();
    const [deleteErrorMessage, setDeleteErrorMessage] = useState<null | string>(null);
    const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
    const [deletingProvider, setDeletingProvider] = useState<null | Provider>(null);
    const navigate = useNavigate();

    const handleProviderEdit = (providerId: string) => {
        navigate(`/settings/providers/${providerId}`);
    };

    const handleProviderClone = (providerId: string) => {
        navigate(`/settings/providers/new?id=${providerId}`);
    };

    const handleProviderDeleteDialogOpen = (provider: Provider) => {
        setDeletingProvider(provider);
        setIsDeleteDialogOpen(true);
    };

    const handleProviderDelete = async (providerId: string | undefined) => {
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
    };

    const columns: ColumnDef<Provider>[] = [
        {
            accessorKey: 'name',
            cell: ({ row }) => <div className="font-medium">{row.getValue('name')}</div>,
            header: ({ column }) => {
                const sorted = column.getIsSorted();

                return (
                    <Button
                        className="flex items-center gap-2 p-0 text-muted-foreground no-underline hover:text-primary hover:no-underline"
                        onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
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
            size: 400,
        },
        {
            accessorKey: 'type',
            cell: ({ row }) => {
                const providerType = row.getValue('type') as ProviderType;
                const Icon = providerIcons[providerType];

                return (
                    <Badge variant="outline">
                        {Icon && <Icon className="mr-1 size-3" />}
                        {providerTypes.find((p) => p.type === providerType)?.label || providerType}
                    </Badge>
                );
            },
            header: ({ column }) => {
                const sorted = column.getIsSorted();

                return (
                    <Button
                        className="flex items-center gap-2 p-0 text-muted-foreground no-underline hover:text-primary hover:no-underline"
                        onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
                        variant="link"
                    >
                        Type
                        {sorted === 'asc' ? (
                            <ArrowDown className="size-4" />
                        ) : sorted === 'desc' ? (
                            <ArrowUp className="size-4" />
                        ) : null}
                    </Button>
                );
            },
            size: 160,
        },
        {
            accessorKey: 'createdAt',
            cell: ({ row }) => {
                const date = new Date(row.getValue('createdAt'));

                return <div className="text-sm">{date.toLocaleDateString()}</div>;
            },
            header: ({ column }) => {
                const sorted = column.getIsSorted();

                return (
                    <Button
                        className="flex items-center gap-2 p-0 text-muted-foreground no-underline hover:text-primary hover:no-underline"
                        onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
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
        },
        {
            accessorKey: 'updatedAt',
            cell: ({ row }) => {
                const date = new Date(row.getValue('updatedAt'));

                return <div className="text-sm">{date.toLocaleDateString()}</div>;
            },
            header: ({ column }) => {
                const sorted = column.getIsSorted();

                return (
                    <Button
                        className="flex items-center gap-2 p-0 text-muted-foreground no-underline hover:text-primary hover:no-underline"
                        onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
                        variant="link"
                    >
                        Updated
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
            cell: ({ row }) => {
                const provider = row.original;

                return (
                    <div className="flex justify-end">
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
            size: 48,
        },
    ];

    const renderSubComponent = ({ row }: { row: any }) => {
        const provider = row.original as Provider;
        const { agents } = provider;

        if (!agents) {
            return <div className="p-4 text-sm text-muted-foreground">No agent configuration available</div>;
        }

        // Convert camelCase key to display name (e.g., 'simpleJson' -> 'Simple Json')
        const getName = (key: string): string =>
            key.replaceAll(/([A-Z])/g, ' $1').replace(/^./, (item) => item.toUpperCase());

        // Recursively extract all fields from an object, flattening nested objects
        const getFields = (obj: any, prefix = ''): { label: string; value: boolean | number | string }[] => {
            if (!obj || typeof obj !== 'object') {
                return [];
            }

            return Object.entries(obj)
                .filter(([key, value]) => key !== '__typename' && !!value)
                .flatMap(([key, value]) => {
                    const label = `${prefix ? `${prefix} ` : ''}${getName(key)}`;

                    return typeof value === 'object'
                        ? getFields(value, label)
                        : [{ label, value: value as boolean | number | string }];
                });
        };

        // Dynamically create agent types from object keys
        const agentTypes = Object.entries(agents)
            .filter(([key]) => key !== '__typename')
            .map(([key, data]) => ({
                data,
                key,
                name: getName(key),
            }))
            .sort((a, b) => a.name.localeCompare(b.name));

        return (
            <div className="border-t bg-muted/20 p-4">
                <h4 className="font-medium">Agent Configurations</h4>
                <hr className="my-4 border-muted-foreground/20" />
                <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 2xl:grid-cols-5">
                    {agentTypes.map(({ data, key, name }) => {
                        // Get all fields from data, including nested objects
                        const fields = data ? getFields(data) : [];

                        return (
                            <div
                                className="space-y-2"
                                key={key}
                            >
                                <div className="text-sm font-medium">{name}</div>
                                {fields.length > 0 ? (
                                    <div className="space-y-1 text-sm">
                                        {fields.map(({ label, value }) => (
                                            <div key={label}>
                                                <span className="text-muted-foreground">{label}:</span> {value}
                                            </div>
                                        ))}
                                    </div>
                                ) : (
                                    <div className="text-sm text-muted-foreground">No configuration available</div>
                                )}
                            </div>
                        );
                    })}
                </div>
            </div>
        );
    };

    if (isLoading) {
        return (
            <div className="space-y-4">
                <SettingsProvidersHeader />
                <StatusCard
                    description="Please wait while we fetch your provider configurations"
                    icon={<Loader2 className="size-16 animate-spin text-muted-foreground" />}
                    title="Loading providers..."
                />
            </div>
        );
    }

    if (error) {
        return (
            <div className="space-y-4">
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

    // Check if providers list is empty
    if (providers.length === 0) {
        return (
            <div className="space-y-4">
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
                    icon={<Settings className="size-8 text-muted-foreground" />}
                    title="No providers configured"
                />
            </div>
        );
    }

    return (
        <div className="space-y-4">
            <SettingsProvidersHeader />

            {/* Delete Error Alert */}
            {(deleteError || deleteErrorMessage) && (
                <Alert variant="destructive">
                    <AlertCircle className="size-4" />
                    <AlertTitle>Error deleting provider</AlertTitle>
                    <AlertDescription>{deleteError?.message || deleteErrorMessage}</AlertDescription>
                </Alert>
            )}

            <DataTable
                columns={columns}
                data={providers}
                filterColumn="name"
                filterPlaceholder="Filter provider names..."
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
};

export default SettingsProviders;
