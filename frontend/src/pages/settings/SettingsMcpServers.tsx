import ConfirmationDialog from '@/components/ConfirmationDialog';
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
import { type ColumnDef } from '@tanstack/react-table';
import { AlertCircle, ArrowUpDown, Copy, Loader2, MoreHorizontal, Pencil, Plus, Server, Trash } from 'lucide-react';
import { useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';

type McpTransport = 'stdio' | 'sse';

interface McpTool {
    name: string;
    description?: string;
    enabled?: boolean;
}

interface McpServerConfigStdio {
    command: string;
    args?: string[];
    env?: Record<string, string>;
}

interface McpServerConfigSse {
    url: string;
    headers?: Record<string, string>;
}

interface McpServerItem {
    id: number;
    name: string;
    transport: McpTransport;
    createdAt: string; // ISO
    updatedAt: string; // ISO
    tools: McpTool[];
    config: {
        stdio?: McpServerConfigStdio | null;
        sse?: McpServerConfigSse | null;
    };
}

const SettingsMcpServersHeader = () => {
    const navigate = useNavigate();

    const handleCreate = () => {
        navigate('/settings/mcp-servers/new');
    };

    return (
        <div className="flex items-center justify-between">
            <p className="text-muted-foreground">Manage MCP servers available to the assistant</p>
            <Button
                variant="secondary"
                onClick={handleCreate}
            >
                Create MCP Server
                <Plus className="h-4 w-4" />
            </Button>
        </div>
    );
};

const formatDate = (iso: string) => new Date(iso).toLocaleDateString();

const SettingsMcpServers = () => {
    const navigate = useNavigate();

    // Mocked data stored locally. This can be replaced by a real query later.
    const initialData: McpServerItem[] = useMemo(
        () => [
            {
                id: 1,
                name: 'Local Filesystem',
                transport: 'stdio',
                createdAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 5).toISOString(),
                updatedAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 1).toISOString(),
                tools: [
                    { name: 'readFile', description: 'Read a file from disk', enabled: true },
                    { name: 'writeFile', description: 'Write content to a file', enabled: false },
                    { name: 'listDirectory', description: 'List files in a directory', enabled: true },
                ],
                config: {
                    stdio: {
                        command: '/usr/local/bin/node',
                        args: ['/opt/mcp/filesystem/index.js', '--root', '/Users/sirozha/Projects'],
                        env: { NODE_ENV: 'production' },
                    },
                    sse: null,
                },
            },
            {
                id: 2,
                name: 'Slack (Prod)',
                transport: 'sse',
                createdAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 20).toISOString(),
                updatedAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 7).toISOString(),
                tools: [
                    { name: 'postMessage', description: 'Send a message to a channel', enabled: true },
                    { name: 'listChannels', description: 'Get a list of channels', enabled: true },
                    { name: 'getUserInfo', description: 'Fetch Slack user info', enabled: false },
                ],
                config: {
                    stdio: null,
                    sse: {
                        url: 'https://mcp.example.com/slack/sse',
                        headers: { Authorization: 'Bearer ***' },
                    },
                },
            },
            {
                id: 3,
                name: 'GitHub Issues',
                transport: 'sse',
                createdAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 30).toISOString(),
                updatedAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 3).toISOString(),
                tools: [
                    { name: 'createIssue', description: 'Create a new issue', enabled: true },
                    { name: 'searchIssues', description: 'Search issues by query', enabled: true },
                    { name: 'addComment', description: 'Add a comment to an issue', enabled: true },
                ],
                config: {
                    stdio: null,
                    sse: {
                        url: 'https://mcp.example.com/github/sse',
                        headers: { Authorization: 'Bearer ***' },
                    },
                },
            },
        ],
        [],
    );

    const [servers, setServers] = useState<McpServerItem[]>(initialData);
    const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
    const [deletingServer, setDeletingServer] = useState<McpServerItem | null>(null);
    const [isDeleteLoading, setIsDeleteLoading] = useState(false);
    const [deleteErrorMessage, setDeleteErrorMessage] = useState<string | null>(null);

    const handleEdit = (serverId: number) => {
        navigate(`/settings/mcp-servers/${serverId}`);
    };

    const handleClone = (serverId: number) => {
        setServers((prev) => {
            const source = prev.find((s) => s.id === serverId);
            if (!source) return prev;
            const nextId = (prev.reduce((max, s) => Math.max(max, s.id), 0) || 0) + 1;
            const nowIso = new Date().toISOString();
            const clone: McpServerItem = {
                ...source,
                id: nextId,
                name: `${source.name} (Copy)`,
                createdAt: nowIso,
                updatedAt: nowIso,
                config: JSON.parse(JSON.stringify(source.config)),
                tools: JSON.parse(JSON.stringify(source.tools || [])),
            };
            return [clone, ...prev];
        });
    };

    const handleOpenDeleteDialog = (server: McpServerItem) => {
        setDeletingServer(server);
        setIsDeleteDialogOpen(true);
    };

    const handleDelete = async (serverId?: number) => {
        if (!serverId) return;
        try {
            setIsDeleteLoading(true);
            setDeleteErrorMessage(null);
            // Simulate async delete
            await new Promise((r) => setTimeout(r, 400));
            setServers((prev) => prev.filter((s) => s.id !== serverId));
            setDeletingServer(null);
            setIsDeleteDialogOpen(false);
        } catch (e) {
            setDeleteErrorMessage('Failed to delete MCP server');
        } finally {
            setIsDeleteLoading(false);
        }
    };

    const columns: ColumnDef<McpServerItem>[] = [
        {
            accessorKey: 'name',
            size: 300,
            header: ({ column }) => (
                <Button
                    variant="ghost"
                    className="-mx-4"
                    onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
                >
                    Name
                    <ArrowUpDown className="ml-2 h-4 w-4" />
                </Button>
            ),
            cell: ({ row }) => (
                <div className="flex items-center gap-2 font-medium">{row.getValue('name') as string}</div>
            ),
        },
        {
            accessorKey: 'transport',
            size: 120,
            header: 'Transport',
            cell: ({ row }) => {
                const t = row.getValue('transport') as McpTransport;
                return <Badge variant="outline">{t.toUpperCase()}</Badge>;
            },
        },
        {
            id: 'tools',
            size: 220,
            header: 'Tools',
            cell: ({ row }) => {
                const s = row.original as McpServerItem;
                const total = (s.tools || []).length;
                if (total === 0) return <span className="text-sm text-muted-foreground">—</span>;
                const enabled = (s.tools || []).filter((t) => t.enabled !== false);
                const first = enabled.slice(0, 3);
                const rest = enabled.length - first.length;
                const disabledCount = total - enabled.length;
                return (
                    <div className="flex items-center gap-1 flex-wrap w-full overflow-hidden">
                        {first.map((t) => (
                            <Badge
                                key={t.name}
                                variant="secondary"
                                className="text-[10px]"
                            >
                                {t.name}
                            </Badge>
                        ))}
                        {rest > 0 && (
                            <Badge
                                variant="outline"
                                className="text-[10px]"
                            >
                                +{rest}
                            </Badge>
                        )}
                        {disabledCount > 0 && (
                            <Badge
                                variant="outline"
                                className="text-[10px] ml-1"
                            >
                                {disabledCount} disabled
                            </Badge>
                        )}
                    </div>
                );
            },
        },
        {
            id: 'endpoint',
            size: 320,
            header: 'Endpoint',
            cell: ({ row }) => {
                const s = row.original as McpServerItem;
                if (s.transport === 'sse' && s.config.sse) {
                    return <span className="text-sm text-muted-foreground break-all">{s.config.sse.url}</span>;
                }
                if (s.transport === 'stdio' && s.config.stdio) {
                    const args = s.config.stdio.args?.join(' ') || '';
                    return (
                        <span className="text-sm text-muted-foreground break-all">
                            {s.config.stdio.command} {args}
                        </span>
                    );
                }
                return <span className="text-sm text-muted-foreground">—</span>;
            },
        },
        {
            accessorKey: 'createdAt',
            size: 100,
            header: 'Created',
            cell: ({ row }) => <div className="text-sm">{formatDate(row.getValue('createdAt'))}</div>,
        },
        {
            accessorKey: 'updatedAt',
            size: 100,
            header: 'Updated',
            cell: ({ row }) => <div className="text-sm">{formatDate(row.getValue('updatedAt'))}</div>,
        },
        {
            id: 'actions',
            size: 48,
            enableHiding: false,
            header: () => null,
            cell: ({ row }) => {
                const server = row.original as McpServerItem;
                return (
                    <div className="flex justify-end">
                        <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                                <Button
                                    variant="ghost"
                                    className="h-8 w-8 p-0"
                                    onClick={(e) => e.stopPropagation()}
                                >
                                    <span className="sr-only">Open menu</span>
                                    <MoreHorizontal className="h-4 w-4" />
                                </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent
                                align="end"
                                className="min-w-[6rem]"
                                onClick={(e) => e.stopPropagation()}
                            >
                                <DropdownMenuItem onClick={() => handleEdit(server.id)}>
                                    <Pencil className="h-3 w-3" />
                                    Edit
                                </DropdownMenuItem>
                                <DropdownMenuItem onClick={() => handleClone(server.id)}>
                                    <Copy className="h-4 w-4" />
                                    Clone
                                </DropdownMenuItem>
                                <DropdownMenuSeparator />
                                <DropdownMenuItem
                                    onClick={() => handleOpenDeleteDialog(server)}
                                    disabled={isDeleteLoading && deletingServer?.id === server.id}
                                >
                                    {isDeleteLoading && deletingServer?.id === server.id ? (
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
        const server = row.original as McpServerItem;

        const renderKeyValue = (obj?: Record<string, string>) => {
            if (!obj || !Object.keys(obj).length) return <div className="text-sm text-muted-foreground">No data</div>;
            return (
                <div className="space-y-1 text-sm">
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
            <div className="p-4 bg-muted/20 border-t space-y-4">
                <h4 className="font-medium">Configuration</h4>
                <hr className="border-muted-foreground/20" />
                {server.transport === 'stdio' && server.config.stdio && (
                    <div className="space-y-2">
                        <div className="text-sm font-medium">STDIO</div>
                        <div className="space-y-1 text-sm">
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
                    <div className="space-y-2">
                        <div className="text-sm font-medium">SSE</div>
                        <div className="space-y-1 text-sm">
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
                <div className="space-y-2">
                    <div className="text-sm font-medium">Tools</div>
                    {server.tools?.length ? (
                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
                            {server.tools.map((t, idx) => (
                                <div
                                    key={`${t.name}-${idx}`}
                                    className="flex items-start justify-between gap-4 rounded-md border p-2"
                                >
                                    <div className="text-sm">
                                        <div className="font-medium">{t.name}</div>
                                        {t.description && <div className="text-muted-foreground">{t.description}</div>}
                                    </div>
                                    <div className="flex items-center gap-2">
                                        <span className="text-xs text-muted-foreground">Enabled</span>
                                        <Switch
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
                                            aria-label={`Toggle ${t.name}`}
                                        />
                                    </div>
                                </div>
                            ))}
                        </div>
                    ) : (
                        <div className="text-sm text-muted-foreground">No tools available</div>
                    )}
                </div>
            </div>
        );
    };

    if (!servers.length) {
        return (
            <div className="space-y-4">
                <SettingsMcpServersHeader />
                <StatusCard
                    icon={<Server className="h-8 w-8 text-muted-foreground" />}
                    title="No MCP servers configured"
                    description="Get started by adding your first MCP server"
                    action={
                        <Button
                            onClick={() => navigate('/settings/mcp-servers/new')}
                            variant="secondary"
                        >
                            <Plus className="h-4 w-4" />
                            Add MCP Server
                        </Button>
                    }
                />
            </div>
        );
    }

    return (
        <div className="space-y-4">
            <SettingsMcpServersHeader />

            {deleteErrorMessage && (
                <Alert variant="destructive">
                    <AlertCircle className="h-4 w-4" />
                    <AlertTitle>Error deleting MCP server</AlertTitle>
                    <AlertDescription>{deleteErrorMessage}</AlertDescription>
                </Alert>
            )}

            <DataTable
                columns={columns}
                data={servers}
                renderSubComponent={renderSubComponent}
            />

            <ConfirmationDialog
                isOpen={isDeleteDialogOpen}
                handleOpenChange={setIsDeleteDialogOpen}
                handleConfirm={() => handleDelete(deletingServer?.id)}
                itemName={deletingServer?.name}
                itemType="MCP server"
                confirmText="Delete"
                cancelText="Cancel"
            />
        </div>
    );
};

export default SettingsMcpServers;
