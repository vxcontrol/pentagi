import type { ColumnDef } from '@tanstack/react-table';

import { format, isToday } from 'date-fns';
import { enUS } from 'date-fns/locale';
import {
    AlertCircle,
    ArrowDown,
    ArrowUp,
    Copy,
    Loader2,
    MoreHorizontal,
    Pencil,
    Plus,
    Server,
    Trash,
} from 'lucide-react';
import { useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';

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
import { Switch } from '@/components/ui/switch';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { useAdaptiveColumnVisibility } from '@/hooks/use-adaptive-column-visibility';

interface McpServerConfigSse {
    headers?: Record<string, string>;
    url: string;
}

interface McpServerConfigStdio {
    args?: string[];
    command: string;
    env?: Record<string, string>;
}

interface McpServerItem {
    config: {
        sse?: McpServerConfigSse | null;
        stdio?: McpServerConfigStdio | null;
    };
    createdAt: string; // ISO
    id: number;
    name: string;
    tools: McpTool[];
    transport: McpTransport;
    updatedAt: string; // ISO
}

interface McpTool {
    description?: string;
    enabled?: boolean;
    name: string;
}

type McpTransport = 'sse' | 'stdio';

const SettingsMcpServersHeader = () => {
    const navigate = useNavigate();

    const handleCreate = () => {
        navigate('/settings/mcp-servers/new');
    };

    return (
        <div className="flex items-center justify-between">
            <p className="text-muted-foreground">Manage MCP servers available to the assistant</p>
            <Button
                onClick={handleCreate}
                variant="secondary"
            >
                Create MCP Server
                <Plus className="size-4" />
            </Button>
        </div>
    );
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

const SettingsMcpServers = () => {
    const navigate = useNavigate();

    const { columnVisibility, updateColumnVisibility } = useAdaptiveColumnVisibility({
        columns: [
            { alwaysVisible: true, id: 'name', priority: 0 },
            { id: 'transport', priority: 1 },
            { id: 'tools', priority: 2 },
            { id: 'createdAt', priority: 3 },
            { id: 'updatedAt', priority: 4 },
            { id: 'endpoint', priority: 5 },
        ],
        tableKey: 'mcp-servers',
    });

    // Mocked data stored locally. This can be replaced by a real query later.
    const initialData: McpServerItem[] = useMemo(
        () => [
            {
                config: {
                    sse: null,
                    stdio: {
                        args: ['/opt/mcp/filesystem/index.js', '--root', '/Users/sirozha/Projects'],
                        command: '/usr/local/bin/node',
                        env: { NODE_ENV: 'production' },
                    },
                },
                createdAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 5).toISOString(),
                id: 1,
                name: 'Local Filesystem',
                tools: [
                    { description: 'Read a file from disk', enabled: true, name: 'readFile' },
                    { description: 'Write content to a file', enabled: false, name: 'writeFile' },
                    { description: 'List files in a directory', enabled: true, name: 'listDirectory' },
                ],
                transport: 'stdio',
                updatedAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 1).toISOString(),
            },
            {
                config: {
                    sse: {
                        headers: { Authorization: 'Bearer ***' },
                        url: 'https://mcp.example.com/slack/sse',
                    },
                    stdio: null,
                },
                createdAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 20).toISOString(),
                id: 2,
                name: 'Slack (Prod)',
                tools: [
                    { description: 'Send a message to a channel', enabled: true, name: 'postMessage' },
                    { description: 'Get a list of channels', enabled: true, name: 'listChannels' },
                    { description: 'Fetch Slack user info', enabled: false, name: 'getUserInfo' },
                ],
                transport: 'sse',
                updatedAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 7).toISOString(),
            },
            {
                config: {
                    sse: {
                        headers: { Authorization: 'Bearer ***' },
                        url: 'https://mcp.example.com/github/sse',
                    },
                    stdio: null,
                },
                createdAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 30).toISOString(),
                id: 3,
                name: 'GitHub Issues',
                tools: [
                    { description: 'Create a new issue', enabled: true, name: 'createIssue' },
                    { description: 'Search issues by query', enabled: true, name: 'searchIssues' },
                    { description: 'Add a comment to an issue', enabled: true, name: 'addComment' },
                ],
                transport: 'sse',
                updatedAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 3).toISOString(),
            },
        ],
        [],
    );

    const [servers, setServers] = useState<McpServerItem[]>(initialData);
    const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
    const [deletingServer, setDeletingServer] = useState<McpServerItem | null>(null);
    const [isDeleteLoading, setIsDeleteLoading] = useState(false);
    const [deleteErrorMessage, setDeleteErrorMessage] = useState<null | string>(null);

    // Three-way sorting handler: null -> asc -> desc -> null
    const handleColumnSort = (column: {
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
    };

    const handleEdit = (serverId: number) => {
        navigate(`/settings/mcp-servers/${serverId}`);
    };

    const handleClone = (serverId: number) => {
        setServers((prev) => {
            const source = prev.find((s) => s.id === serverId);

            if (!source) {
                return prev;
            }

            const nextId = (prev.reduce((max, s) => Math.max(max, s.id), 0) || 0) + 1;
            const nowIso = new Date().toISOString();
            const clone: McpServerItem = {
                ...source,
                config: JSON.parse(JSON.stringify(source.config)),
                createdAt: nowIso,
                id: nextId,
                name: `${source.name} (Copy)`,
                tools: JSON.parse(JSON.stringify(source.tools || [])),
                updatedAt: nowIso,
            };

            return [clone, ...prev];
        });
    };

    const handleOpenDeleteDialog = (server: McpServerItem) => {
        setDeletingServer(server);
        setIsDeleteDialogOpen(true);
    };

    const handleDelete = async (serverId?: number) => {
        if (!serverId) {
            return;
        }

        try {
            setIsDeleteLoading(true);
            setDeleteErrorMessage(null);
            // Simulate async delete
            await new Promise((r) => setTimeout(r, 400));
            setServers((prev) => prev.filter((s) => s.id !== serverId));
            setDeletingServer(null);
            setIsDeleteDialogOpen(false);
        } catch {
            setDeleteErrorMessage('Failed to delete MCP server');
        } finally {
            setIsDeleteLoading(false);
        }
    };

    const columns: ColumnDef<McpServerItem>[] = [
        {
            accessorKey: 'name',
            cell: ({ row }) => (
                <div className="flex items-center gap-2 font-medium">{row.getValue('name') as string}</div>
            ),
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
            accessorKey: 'transport',
            cell: ({ row }) => {
                const t = row.getValue('transport') as McpTransport;

                return <Badge variant="outline">{t.toUpperCase()}</Badge>;
            },
            header: 'Transport',
            size: 120,
        },
        {
            cell: ({ row }) => {
                const s = row.original as McpServerItem;
                const total = (s.tools || []).length;

                if (total === 0) {
                    return <span className="text-muted-foreground text-sm">—</span>;
                }

                const enabled = (s.tools || []).filter((t) => t.enabled !== false);
                const first = enabled.slice(0, 3);
                const rest = enabled.length - first.length;
                const disabledCount = total - enabled.length;

                return (
                    <div className="flex w-full flex-wrap items-center gap-1 overflow-hidden">
                        {first.map((t) => (
                            <Badge
                                className="text-[10px]"
                                key={t.name}
                                variant="secondary"
                            >
                                {t.name}
                            </Badge>
                        ))}
                        {rest > 0 && (
                            <Badge
                                className="text-[10px]"
                                variant="outline"
                            >
                                +{rest}
                            </Badge>
                        )}
                        {disabledCount > 0 && (
                            <Badge
                                className="ml-1 text-[10px]"
                                variant="outline"
                            >
                                {disabledCount} disabled
                            </Badge>
                        )}
                    </div>
                );
            },
            header: 'Tools',
            id: 'tools',
            size: 220,
        },
        {
            cell: ({ row }) => {
                const s = row.original as McpServerItem;

                if (s.transport === 'sse' && s.config.sse) {
                    return <span className="text-muted-foreground text-sm break-all">{s.config.sse.url}</span>;
                }

                if (s.transport === 'stdio' && s.config.stdio) {
                    const args = s.config.stdio.args?.join(' ') || '';

                    return (
                        <span className="text-muted-foreground text-sm break-all">
                            {s.config.stdio.command} {args}
                        </span>
                    );
                }

                return <span className="text-muted-foreground text-sm">—</span>;
            },
            header: 'Endpoint',
            id: 'endpoint',
            size: 320,
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
                const server = row.original as McpServerItem;

                return (
                    <div className="flex justify-end opacity-0 transition-opacity group-hover:opacity-100">
                        <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                                <Button
                                    className="size-8 p-0"
                                    onClick={(e) => e.stopPropagation()}
                                    variant="ghost"
                                >
                                    <span className="sr-only">Open menu</span>
                                    <MoreHorizontal className="size-4" />
                                </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent
                                align="end"
                                className="min-w-24"
                                onClick={(e) => e.stopPropagation()}
                            >
                                <DropdownMenuItem onClick={() => handleEdit(server.id)}>
                                    <Pencil className="size-3" />
                                    Edit
                                </DropdownMenuItem>
                                <DropdownMenuItem onClick={() => handleClone(server.id)}>
                                    <Copy className="size-4" />
                                    Clone
                                </DropdownMenuItem>
                                <DropdownMenuSeparator />
                                <DropdownMenuItem
                                    disabled={isDeleteLoading && deletingServer?.id === server.id}
                                    onClick={() => handleOpenDeleteDialog(server)}
                                >
                                    {isDeleteLoading && deletingServer?.id === server.id ? (
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
        const server = row.original as McpServerItem;

        const renderKeyValue = (obj?: Record<string, string>) => {
            if (!obj || Object.keys(obj).length === 0) {
                return <div className="text-muted-foreground text-sm">No data</div>;
            }

            return (
                <div className="flex flex-col gap-1 text-sm">
                    {Object.entries(obj)
                        .filter(([_, v]) => !!v)
                        .map(([k, v]) => (
                            <div key={k}>
                                <span className="text-muted-foreground">{k}:</span> {v}
                            </div>
                        ))}
                </div>
            );
        };

        return (
            <div className="bg-muted/20 flex flex-col gap-4 border-t p-4">
                <h4 className="font-medium">Configuration</h4>
                <hr className="border-muted-foreground/20" />
                {server.transport === 'stdio' && server.config.stdio && (
                    <div className="flex flex-col gap-2">
                        <div className="text-sm font-medium">STDIO</div>
                        <div className="flex flex-col gap-1 text-sm">
                            <div>
                                <span className="text-muted-foreground">Command:</span> {server.config.stdio.command}
                            </div>
                            {!!server.config.stdio.args?.length && (
                                <div>
                                    <span className="text-muted-foreground">Args:</span>{' '}
                                    {server.config.stdio.args.join(' ')}
                                </div>
                            )}
                        </div>
                        <div>
                            <div className="text-sm font-medium">Env</div>
                            {renderKeyValue(server.config.stdio.env)}
                        </div>
                    </div>
                )}
                {server.transport === 'sse' && server.config.sse && (
                    <div className="flex flex-col gap-2">
                        <div className="text-sm font-medium">SSE</div>
                        <div className="flex flex-col gap-1 text-sm">
                            <div>
                                <span className="text-muted-foreground">URL:</span> {server.config.sse.url}
                            </div>
                        </div>
                        <div>
                            <div className="text-sm font-medium">Headers</div>
                            {renderKeyValue(server.config.sse.headers)}
                        </div>
                    </div>
                )}
                <div className="flex flex-col gap-2">
                    <div className="text-sm font-medium">Tools</div>
                    {server.tools?.length ? (
                        <div className="grid grid-cols-1 gap-3 md:grid-cols-2 lg:grid-cols-3">
                            {server.tools.map((t, idx) => (
                                <div
                                    className="flex items-start justify-between gap-4 rounded-md border p-2"
                                    key={`${t.name}-${idx}`}
                                >
                                    <div className="text-sm">
                                        <div className="font-medium">{t.name}</div>
                                        {t.description && <div className="text-muted-foreground">{t.description}</div>}
                                    </div>
                                    <div className="flex items-center gap-2">
                                        <span className="text-muted-foreground text-xs">Enabled</span>
                                        <Switch
                                            aria-label={`Toggle ${t.name}`}
                                            checked={t.enabled !== false}
                                            onCheckedChange={(checked) => {
                                                setServers((prev) =>
                                                    prev.map((s) =>
                                                        s.id === server.id
                                                            ? {
                                                                  ...s,
                                                                  tools: s.tools.map((orig, i) =>
                                                                      i === idx ? { ...orig, enabled: checked } : orig,
                                                                  ),
                                                              }
                                                            : s,
                                                    ),
                                                );
                                            }}
                                        />
                                    </div>
                                </div>
                            ))}
                        </div>
                    ) : (
                        <div className="text-muted-foreground text-sm">No tools available</div>
                    )}
                </div>
            </div>
        );
    };

    if (servers.length === 0) {
        return (
            <div className="flex flex-col gap-4">
                <SettingsMcpServersHeader />
                <StatusCard
                    action={
                        <Button
                            onClick={() => navigate('/settings/mcp-servers/new')}
                            variant="secondary"
                        >
                            <Plus className="size-4" />
                            Add MCP Server
                        </Button>
                    }
                    description="Get started by adding your first MCP server"
                    icon={<Server className="text-muted-foreground size-8" />}
                    title="No MCP servers configured"
                />
            </div>
        );
    }

    return (
        <div className="flex flex-col gap-4">
            <SettingsMcpServersHeader />

            {deleteErrorMessage && (
                <Alert variant="destructive">
                    <AlertCircle className="size-4" />
                    <AlertTitle>Error deleting MCP server</AlertTitle>
                    <AlertDescription>{deleteErrorMessage}</AlertDescription>
                </Alert>
            )}

            <DataTable<McpServerItem>
                columns={columns}
                columnVisibility={columnVisibility}
                data={servers}
                onColumnVisibilityChange={(visibility) => {
                    Object.entries(visibility).forEach(([columnId, isVisible]) => {
                        if (columnVisibility[columnId] !== isVisible) {
                            updateColumnVisibility(columnId, isVisible);
                        }
                    });
                }}
                renderSubComponent={renderSubComponent}
                tableKey="mcp-servers"
            />

            <ConfirmationDialog
                cancelText="Cancel"
                confirmText="Delete"
                handleConfirm={() => handleDelete(deletingServer?.id)}
                handleOpenChange={setIsDeleteDialogOpen}
                isOpen={isDeleteDialogOpen}
                itemName={deletingServer?.name}
                itemType="MCP server"
            />
        </div>
    );
};

export default SettingsMcpServers;
