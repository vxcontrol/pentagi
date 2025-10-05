import ConfirmationDialog from '@/components/ConfirmationDialog';
import Anthropic from '@/components/icons/Anthropic';
import Bedrock from '@/components/icons/Bedrock';
import Custom from '@/components/icons/Custom';
import Gemini from '@/components/icons/Gemini';
import Ollama from '@/components/icons/Ollama';
import OpenAi from '@/components/icons/OpenAi';
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
import {
    ProviderType,
    useDeleteProviderMutation,
    useSettingsProvidersQuery,
    type ProviderConfigFragmentFragment,
} from '@/graphql/types';
import { type ColumnDef } from '@tanstack/react-table';
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
    { type: ProviderType.Anthropic, label: 'Anthropic' },
    { type: ProviderType.Bedrock, label: 'Bedrock' },
    { type: ProviderType.Custom, label: 'Custom' },
    { type: ProviderType.Gemini, label: 'Gemini' },
    { type: ProviderType.Ollama, label: 'Ollama' },
    { type: ProviderType.Openai, label: 'OpenAI' },
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
                        <ChevronDown className="h-4 w-4" />
                    </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent
                    align="end"
                    style={{
                        width: 'var(--radix-dropdown-menu-trigger-width)',
                    }}
                >
                    {providerTypes.map(({ type, label }) => {
                        const Icon = providerIcons[type];
                        return (
                            <DropdownMenuItem
                                key={type}
                                onClick={() => handleProviderCreate(type)}
                            >
                                {Icon && <Icon className="h-4 w-4" />}
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
    const { data, loading: isLoading, error } = useSettingsProvidersQuery();
    const [deleteProvider, { loading: isDeleteLoading, error: deleteError }] = useDeleteProviderMutation();
    const [deleteErrorMessage, setDeleteErrorMessage] = useState<string | null>(null);
    const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
    const [deletingProvider, setDeletingProvider] = useState<Provider | null>(null);
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
                variables: { providerId: providerId.toString() },
                refetchQueries: ['settingsProviders'],
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
            size: 400,
            header: ({ column }) => {
                const sorted = column.getIsSorted();

                return (
                    <Button
                        variant="link"
                        onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
                        className="flex items-center gap-2 p-0 no-underline hover:no-underline text-muted-foreground hover:text-primary"
                    >
                        Name
                        {sorted === 'asc' ? (
                            <ArrowDown className="h-4 w-4" />
                        ) : sorted === 'desc' ? (
                            <ArrowUp className="h-4 w-4" />
                        ) : null}
                    </Button>
                );
            },
            cell: ({ row }) => <div className="font-medium">{row.getValue('name')}</div>,
        },
        {
            accessorKey: 'type',
            size: 160,
            header: ({ column }) => {
                const sorted = column.getIsSorted();

                return (
                    <Button
                        variant="link"
                        onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
                        className="flex items-center gap-2 p-0 no-underline hover:no-underline text-muted-foreground hover:text-primary"
                    >
                        Type
                        {sorted === 'asc' ? (
                            <ArrowDown className="h-4 w-4" />
                        ) : sorted === 'desc' ? (
                            <ArrowUp className="h-4 w-4" />
                        ) : null}
                    </Button>
                );
            },
            cell: ({ row }) => {
                const providerType = row.getValue('type') as ProviderType;
                const Icon = providerIcons[providerType];
                return (
                    <Badge variant="outline">
                        {Icon && <Icon className="h-3 w-3 mr-1" />}
                        {providerTypes.find((p) => p.type === providerType)?.label || providerType}
                    </Badge>
                );
            },
        },
        {
            accessorKey: 'createdAt',
            size: 120,
            header: ({ column }) => {
                const sorted = column.getIsSorted();

                return (
                    <Button
                        variant="link"
                        onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
                        className="flex items-center gap-2 p-0 no-underline hover:no-underline text-muted-foreground hover:text-primary"
                    >
                        Created
                        {sorted === 'asc' ? (
                            <ArrowDown className="h-4 w-4" />
                        ) : sorted === 'desc' ? (
                            <ArrowUp className="h-4 w-4" />
                        ) : null}
                    </Button>
                );
            },
            cell: ({ row }) => {
                const date = new Date(row.getValue('createdAt'));
                return <div className="text-sm">{date.toLocaleDateString()}</div>;
            },
        },
        {
            accessorKey: 'updatedAt',
            size: 120,
            header: ({ column }) => {
                const sorted = column.getIsSorted();

                return (
                    <Button
                        variant="link"
                        onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
                        className="flex items-center gap-2 p-0 no-underline hover:no-underline text-muted-foreground hover:text-primary"
                    >
                        Updated
                        {sorted === 'asc' ? (
                            <ArrowDown className="h-4 w-4" />
                        ) : sorted === 'desc' ? (
                            <ArrowUp className="h-4 w-4" />
                        ) : null}
                    </Button>
                );
            },
            cell: ({ row }) => {
                const date = new Date(row.getValue('updatedAt'));
                return <div className="text-sm">{date.toLocaleDateString()}</div>;
            },
        },
        {
            id: 'actions',
            size: 48,
            enableHiding: false,
            header: () => null,
            cell: ({ row }) => {
                const provider = row.original;

                return (
                    <div className="flex justify-end">
                        <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                                <Button
                                    variant="ghost"
                                    className="h-8 w-8 p-0"
                                >
                                    <span className="sr-only">Open menu</span>
                                    <MoreHorizontal className="h-4 w-4" />
                                </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent
                                align="end"
                                className="min-w-[6rem]"
                            >
                                <DropdownMenuItem onClick={() => handleProviderEdit(provider.id)}>
                                    <Pencil className="h-3 w-3" />
                                    Edit
                                </DropdownMenuItem>
                                <DropdownMenuItem onClick={() => handleProviderClone(provider.id)}>
                                    <Copy className="h-4 w-4" />
                                    Clone
                                </DropdownMenuItem>
                                <DropdownMenuSeparator />
                                <DropdownMenuItem
                                    onClick={() => handleProviderDeleteDialogOpen(provider)}
                                    disabled={isDeleteLoading && deletingProvider?.id === provider.id}
                                >
                                    {isDeleteLoading && deletingProvider?.id === provider.id ? (
                                        <>
                                            <Loader2 className="h-4 w-4 animate-spin" />
                                            Deleting...
                                        </>
                                    ) : (
                                        <>
                                            <Trash className="h-4 w-4" />
                                            Delete
                                        </>
                                    )}
                                </DropdownMenuItem>
                            </DropdownMenuContent>
                        </DropdownMenu>
                    </div>
                );
            },
        },
    ];

    const renderSubComponent = ({ row }: { row: any }) => {
        const provider = row.original as Provider;
        const agents = provider.agents;

        if (!agents) {
            return <div className="p-4 text-sm text-muted-foreground">No agent configuration available</div>;
        }

        // Convert camelCase key to display name (e.g., 'simpleJson' -> 'Simple Json')
        const getName = (key: string): string =>
            key.replace(/([A-Z])/g, ' $1').replace(/^./, (item) => item.toUpperCase());

        // Recursively extract all fields from an object, flattening nested objects
        const getFields = (obj: any, prefix = ''): { label: string; value: number | string | boolean }[] => {
            if (!obj || typeof obj !== 'object') {
                return [];
            }

            return Object.entries(obj)
                .filter(([key, value]) => key !== '__typename' && !!value)
                .flatMap(([key, value]) => {
                    const label = `${prefix ? `${prefix} ` : ''}${getName(key)}`;

                    return typeof value === 'object'
                        ? getFields(value, label)
                        : [{ label, value: value as string | number | boolean }];
                });
        };

        // Dynamically create agent types from object keys
        const agentTypes = Object.entries(agents)
            .filter(([key]) => key !== '__typename')
            .map(([key, data]) => ({
                name: getName(key),
                key,
                data,
            }))
            .sort((a, b) => a.name.localeCompare(b.name));

        return (
            <div className="p-4 bg-muted/20 border-t">
                <h4 className="font-medium">Agent Configurations</h4>
                <hr className="my-4 border-muted-foreground/20" />
                <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 2xl:grid-cols-5 gap-4">
                    {agentTypes.map(({ name, key, data }) => {
                        // Get all fields from data, including nested objects
                        const fields = data ? getFields(data) : [];

                        return (
                            <div
                                key={key}
                                className="space-y-2"
                            >
                                <div className="font-medium text-sm">{name}</div>
                                {!!fields.length ? (
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
                    icon={<Loader2 className="w-16 h-16 animate-spin text-muted-foreground" />}
                    title="Loading providers..."
                    description="Please wait while we fetch your provider configurations"
                />
            </div>
        );
    }

    if (error) {
        return (
            <div className="space-y-4">
                <SettingsProvidersHeader />
                <Alert variant="destructive">
                    <AlertCircle className="h-4 w-4" />
                    <AlertTitle>Error loading providers</AlertTitle>
                    <AlertDescription>{error.message}</AlertDescription>
                </Alert>
            </div>
        );
    }

    const providers = data?.settingsProviders?.userDefined || [];

    // Check if providers list is empty
    if (!providers.length) {
        return (
            <div className="space-y-4">
                <SettingsProvidersHeader />
                <StatusCard
                    icon={<Settings className="h-8 w-8 text-muted-foreground" />}
                    title="No providers configured"
                    description="Get started by adding your first language model provider"
                    action={
                        <Button
                            onClick={() => navigate('/settings/providers/new')}
                            variant="secondary"
                        >
                            <Plus className="h-4 w-4" />
                            Add Provider
                        </Button>
                    }
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
                    <AlertCircle className="h-4 w-4" />
                    <AlertTitle>Error deleting provider</AlertTitle>
                    <AlertDescription>{deleteError?.message || deleteErrorMessage}</AlertDescription>
                </Alert>
            )}

            <DataTable
                columns={columns}
                data={providers}
                renderSubComponent={renderSubComponent}
                filterColumn="name"
                filterPlaceholder="Filter provider names..."
            />

            <ConfirmationDialog
                isOpen={isDeleteDialogOpen}
                handleOpenChange={setIsDeleteDialogOpen}
                handleConfirm={() => handleProviderDelete(deletingProvider?.id)}
                itemName={deletingProvider?.name}
                itemType="provider"
                confirmText="Delete"
                cancelText="Cancel"
            />
        </div>
    );
};

export default SettingsProviders;
