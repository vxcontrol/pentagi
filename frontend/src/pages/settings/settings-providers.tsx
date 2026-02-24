import type { ColumnDef } from '@tanstack/react-table';

import { format, isToday } from 'date-fns';
import { enUS } from 'date-fns/locale';
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
import { useCallback, useMemo, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';

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
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { ProviderType, useDeleteProviderMutation, useSettingsProvidersQuery } from '@/graphql/types';
import { useAdaptiveColumnVisibility } from '@/hooks/use-adaptive-column-visibility';

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
    const [searchParams, setSearchParams] = useSearchParams();
    const { data, error, loading: isLoading } = useSettingsProvidersQuery();
    const [deleteProvider, { error: deleteError, loading: isDeleteLoading }] = useDeleteProviderMutation();
    const [deleteErrorMessage, setDeleteErrorMessage] = useState<null | string>(null);
    const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
    const [deletingProvider, setDeletingProvider] = useState<null | Provider>(null);
    const navigate = useNavigate();

    const { columnVisibility, updateColumnVisibility } = useAdaptiveColumnVisibility({
        columns: [
            { alwaysVisible: true, id: 'name', priority: 0 },
            { id: 'type', priority: 1 },
            { id: 'createdAt', priority: 2 },
            { id: 'updatedAt', priority: 3 },
        ],
        tableKey: 'providers',
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
                cell: ({ row }) => <div className="font-medium">{row.getValue('name')}</div>,
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
                            className="text-muted-foreground hover:text-primary flex items-center gap-2 p-0 no-underline hover:no-underline"
                            onClick={() => handleColumnSort(column)}
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
                accessorKey: 'updatedAt',
                cell: ({ row }) => {
                    const dateString = row.getValue('updatedAt') as string;

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
        ],
        [
            handleColumnSort,
            handleProviderClone,
            handleProviderDeleteDialogOpen,
            handleProviderEdit,
            isDeleteLoading,
            deletingProvider,
        ],
    );

    const renderSubComponent = ({ row }: { row: any }) => {
        const provider = row.original as Provider;
        const { agents } = provider;

        if (!agents) {
            return <div className="text-muted-foreground p-4 text-sm">No agent configuration available</div>;
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
            <div className="bg-muted/20 border-t p-4">
                <h4 className="font-medium">Agent Configurations</h4>
                <hr className="border-muted-foreground/20 my-4" />
                <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 2xl:grid-cols-5">
                    {agentTypes.map(({ data, key, name }) => {
                        // Get all fields from data, including nested objects
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

    // Check if providers list is empty
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
                columnVisibility={columnVisibility}
                data={providers}
                filterColumn="name"
                filterPlaceholder="Filter provider names..."
                onColumnVisibilityChange={(visibility) => {
                    Object.entries(visibility).forEach(([columnId, isVisible]) => {
                        if (columnVisibility[columnId] !== isVisible) {
                            updateColumnVisibility(columnId, isVisible);
                        }
                    });
                }}
                onPageChange={handlePageChange}
                pageIndex={currentPage}
                renderSubComponent={renderSubComponent}
                tableKey="providers"
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
