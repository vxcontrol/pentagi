import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { DataTable } from '@/components/ui/data-table';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { StatusCard } from '@/components/ui/status-card';
import { useSettingsProvidersQuery, type ProviderConfigFragmentFragment } from '@/graphql/types';
import { type ColumnDef } from '@tanstack/react-table';
import {
    AlertCircle,
    ArrowUpDown,
    ChevronDown,
    ChevronRight,
    Loader2,
    MoreHorizontal,
    Plus,
    Settings,
} from 'lucide-react';
import { useNavigate } from 'react-router-dom';

type Provider = ProviderConfigFragmentFragment;

const SettingsProvidersHeader = () => {
    const navigate = useNavigate();

    return (
        <div className="flex items-center justify-between">
            <p className="text-muted-foreground">Manage language model providers</p>

            <Button
                onClick={() => navigate('/settings/providers/new')}
                variant="secondary"
            >
                <Plus className="h-4 w-4" />
                Add Provider
            </Button>
        </div>
    );
};

const SettingsProviders = () => {
    const { data, loading, error } = useSettingsProvidersQuery();
    const navigate = useNavigate();

    const handleProviderEdit = (providerId: string) => {
        navigate(`/settings/providers/${providerId}`);
    };

    const handleProviderDelete = (providerId: string) => {
        console.log('delete', providerId);
    };

    const columns: ColumnDef<Provider>[] = [
        {
            id: 'expander',
            size: 16,
            enableHiding: false,
            meta: {
                headerClassName: 'pr-0',
                cellClassName: 'pr-0',
            },
            header: () => null,
            cell: ({ row }) => {
                return (
                    <>
                        {row.getIsExpanded() ? (
                            <ChevronDown className="h-4 w-4" />
                        ) : (
                            <ChevronRight className="h-4 w-4" />
                        )}
                    </>
                );
            },
        },
        {
            accessorKey: 'name',
            size: 400,
            header: ({ column }) => {
                return (
                    <Button
                        variant="ghost"
                        className="-mx-4"
                        onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
                    >
                        Name
                        <ArrowUpDown className="ml-2 h-4 w-4" />
                    </Button>
                );
            },
            cell: ({ row }) => <div className="font-medium">{row.getValue('name')}</div>,
        },
        {
            accessorKey: 'type',
            size: 100,
            header: ({ column }) => {
                return (
                    <Button
                        variant="ghost"
                        className="-mx-4"
                        onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
                    >
                        Type
                        <ArrowUpDown className="ml-2 h-4 w-4" />
                    </Button>
                );
            },
            cell: ({ row }) => <Badge variant="outline">{row.getValue('type')}</Badge>,
        },
        {
            accessorKey: 'createdAt',
            size: 100,
            header: ({ column }) => {
                return (
                    <Button
                        variant="ghost"
                        className="-mx-4"
                        onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
                    >
                        Created
                        <ArrowUpDown className="ml-2 h-4 w-4" />
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
            size: 100,
            header: ({ column }) => {
                return (
                    <Button
                        variant="ghost"
                        className="-mx-4"
                        onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
                    >
                        Updated
                        <ArrowUpDown className="ml-2 h-4 w-4" />
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
                            <DropdownMenuContent align="end">
                                <DropdownMenuItem onClick={() => handleProviderEdit(provider.id)}>
                                    Edit provider
                                </DropdownMenuItem>
                                <DropdownMenuItem onClick={() => handleProviderDelete(provider.id)}>
                                    Delete provider
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
            }));

        return (
            <div className="p-4 bg-muted/20 border-t">
                <h4 className="font-medium">Agent Configurations</h4>
                <hr className="my-4 border-muted-foreground/20" />
                <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
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

    if (loading) {
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
            <DataTable
                columns={columns}
                data={providers}
                renderSubComponent={renderSubComponent}
            />
        </div>
    );
};

export default SettingsProviders;
