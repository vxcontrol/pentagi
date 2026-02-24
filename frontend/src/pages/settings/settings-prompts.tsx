import type { ColumnDef } from '@tanstack/react-table';

import {
    AlertCircle,
    ArrowDown,
    ArrowUp,
    Bot,
    Code,
    Loader2,
    MoreHorizontal,
    Pencil,
    RotateCcw,
    Settings,
    Trash2,
    User,
    Wrench,
} from 'lucide-react';
import { Fragment, useState } from 'react';
import { useNavigate } from 'react-router-dom';

import type { AgentPrompt, AgentPrompts, DefaultPrompt, PromptType } from '@/graphql/types';

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
import { useDeletePromptMutation, useSettingsPromptsQuery } from '@/graphql/types';
import { useAdaptiveColumnVisibility } from '@/hooks/use-adaptive-column-visibility';

// Types for table data
type AgentPromptTableData = {
    displayName: string; // Formatted display name
    hasHuman: boolean;
    hasSystem: boolean;
    humanStatus: 'Custom' | 'Default' | 'N/A';
    humanTemplate?: string;
    humanType?: PromptType; // Type for human prompt lookup
    name: string; // Original key (camelCase)
    systemStatus: 'Custom' | 'Default' | 'N/A';
    systemTemplate: string;
    systemType?: PromptType; // Type for system prompt lookup
};

type ToolPromptTableData = {
    displayName: string; // Formatted display name
    name: string; // Original key (camelCase)
    promptType?: PromptType; // Type for prompt lookup
    status: 'Custom' | 'Default' | 'N/A';
    template: string;
};

const SettingsPromptsHeader = () => {
    return (
        <div className="flex items-center justify-between">
            <p className="text-muted-foreground">Manage system and custom prompt templates</p>
        </div>
    );
};

const SettingsPrompts = () => {
    const { data, error, loading: isLoading } = useSettingsPromptsQuery();
    const [deletePrompt, { loading: isDeleteLoading }] = useDeletePromptMutation();
    const navigate = useNavigate();

    // Reset dialog states
    const [resetDialogOpen, setResetDialogOpen] = useState(false);
    const [resetOperation, setResetOperation] = useState<null | {
        displayName: string;
        promptName: string;
        type: 'all' | 'human' | 'system' | 'tool';
    }>(null);

    const { columnVisibility: agentColumnVisibility, updateColumnVisibility: updateAgentColumnVisibility } =
        useAdaptiveColumnVisibility({
            columns: [
                { alwaysVisible: true, id: 'displayName', priority: 0 },
                { id: 'systemStatus', priority: 1 },
                { id: 'humanStatus', priority: 2 },
            ],
            tableKey: 'prompts-agents',
        });

    const { columnVisibility: toolColumnVisibility, updateColumnVisibility: updateToolColumnVisibility } =
        useAdaptiveColumnVisibility({
            columns: [
                { alwaysVisible: true, id: 'displayName', priority: 0 },
                { id: 'status', priority: 1 },
            ],
            tableKey: 'prompts-tools',
        });

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

    // Handler for editing any prompt (agent or tool)
    const handlePromptEdit = (promptName: string) => {
        navigate(`/settings/prompts/${promptName}`);
    };

    // Reset dialog handlers
    const handleResetDialogOpen = (
        type: 'all' | 'human' | 'system' | 'tool',
        promptName: string,
        displayName: string,
    ) => {
        setResetOperation({ displayName, promptName, type });
        setResetDialogOpen(true);
    };

    const handleResetPrompt = async () => {
        if (!resetOperation || !data?.settingsPrompts?.default) {
            return;
        }

        try {
            const { promptName, type } = resetOperation;
            const { agents } = data.settingsPrompts.default;
            const { tools } = data.settingsPrompts.default;
            const userDefined = data.settingsPrompts.userDefined || [];

            if (type === 'tool') {
                // Handle tool prompt reset
                const toolPrompt = tools?.[promptName as keyof typeof tools];

                if (toolPrompt?.type) {
                    // Find the user-defined prompt with matching type
                    const userPrompt = userDefined.find((p) => p.type === toolPrompt.type);

                    if (userPrompt) {
                        await deletePrompt({
                            refetchQueries: ['settingsPrompts'],
                            variables: { promptId: userPrompt.id },
                        });
                    }
                }
            } else {
                // Handle agent prompt reset
                const agentPrompts = agents?.[promptName as keyof typeof agents] as AgentPrompts;

                if (agentPrompts) {
                    const systemType = agentPrompts.system?.type;
                    const humanType = agentPrompts.human?.type;

                    if (type === 'system' && systemType) {
                        const userPrompt = userDefined.find((p) => p.type === systemType);

                        if (userPrompt) {
                            await deletePrompt({
                                refetchQueries: ['settingsPrompts'],
                                variables: { promptId: userPrompt.id },
                            });
                        }
                    } else if (type === 'human' && humanType) {
                        const userPrompt = userDefined.find((p) => p.type === humanType);

                        if (userPrompt) {
                            await deletePrompt({
                                refetchQueries: ['settingsPrompts'],
                                variables: { promptId: userPrompt.id },
                            });
                        }
                    } else if (type === 'all') {
                        if (systemType) {
                            const userSystemPrompt = userDefined.find((p) => p.type === systemType);

                            if (userSystemPrompt) {
                                await deletePrompt({
                                    refetchQueries: ['settingsPrompts'],
                                    variables: { promptId: userSystemPrompt.id },
                                });
                            }
                        }

                        if (humanType) {
                            const userHumanPrompt = userDefined.find((p) => p.type === humanType);

                            if (userHumanPrompt) {
                                await deletePrompt({
                                    refetchQueries: ['settingsPrompts'],
                                    variables: { promptId: userHumanPrompt.id },
                                });
                            }
                        }
                    }
                }
            }

            setResetOperation(null);
        } catch (error) {
            console.error('Failed to reset prompt:', error);
        }
    };

    // Helper function to check if reset is available for specific prompt type
    const canResetPrompt = (promptName: string, resetType: 'all' | 'human' | 'system' | 'tool'): boolean => {
        if (!data?.settingsPrompts?.default || !data?.settingsPrompts?.userDefined) {
            return false;
        }

        const { userDefined } = data.settingsPrompts;
        const { agents } = data.settingsPrompts.default;
        const { tools } = data.settingsPrompts.default;

        if (resetType === 'tool') {
            const toolPrompt = tools?.[promptName as keyof typeof tools];

            return toolPrompt?.type ? userDefined.some((p) => p.type === toolPrompt.type) : false;
        } else {
            const agentPrompts = agents?.[promptName as keyof typeof agents] as AgentPrompts;

            if (!agentPrompts) {
                return false;
            }

            const systemType = agentPrompts.system?.type;
            const humanType = agentPrompts.human?.type;

            switch (resetType) {
                case 'all': {
                    const hasCustomSystem = systemType ? userDefined.some((p) => p.type === systemType) : false;
                    const hasCustomHuman = humanType ? userDefined.some((p) => p.type === humanType) : false;

                    return hasCustomSystem || hasCustomHuman;
                }

                case 'human': {
                    return humanType ? userDefined.some((p) => p.type === humanType) : false;
                }

                case 'system': {
                    return systemType ? userDefined.some((p) => p.type === systemType) : false;
                }
                // No default
            }
        }

        return false;
    };

    // Transform agents data for table
    const getAgentPromptsData = (): AgentPromptTableData[] => {
        if (!data?.settingsPrompts?.default?.agents) {
            return [];
        }

        const { agents } = data.settingsPrompts.default;
        const userDefined = data.settingsPrompts.userDefined || [];
        const agentEntries: AgentPromptTableData[] = [];

        // Helper function to format agent name
        const formatName = (key: string): string => {
            return key.replaceAll(/([A-Z])/g, ' $1').replace(/^./, (str) => str.toUpperCase());
        };

        // Process each agent
        Object.entries(agents).forEach(([key, prompts]) => {
            if (key === '__typename') {
                return;
            }

            const systemType = (prompts as AgentPrompt | AgentPrompts)?.system?.type;
            const humanType = (prompts as AgentPrompts)?.human?.type;

            // Check if user has custom prompts
            const hasCustomSystem = userDefined.some((p) => p.type === systemType);
            const hasCustomHuman = humanType ? userDefined.some((p) => p.type === humanType) : false;

            const agentData: AgentPromptTableData = {
                displayName: formatName(key),
                hasHuman: !!(prompts as AgentPrompts)?.human,
                hasSystem: !!(prompts as AgentPrompt | AgentPrompts)?.system,
                humanStatus: (prompts as AgentPrompts)?.human ? (hasCustomHuman ? 'Custom' : 'Default') : 'N/A',
                humanTemplate: (prompts as AgentPrompts)?.human?.template,
                humanType,
                name: key,
                systemStatus: (prompts as AgentPrompt | AgentPrompts)?.system
                    ? hasCustomSystem
                        ? 'Custom'
                        : 'Default'
                    : 'N/A',
                systemTemplate: (prompts as AgentPrompt | AgentPrompts)?.system?.template || '',
                systemType,
            };

            agentEntries.push(agentData);
        });

        return agentEntries.sort((a, b) => a.name.localeCompare(b.name));
    };

    // Transform tools data for table
    const getToolPromptsData = (): ToolPromptTableData[] => {
        if (!data?.settingsPrompts?.default?.tools) {
            return [];
        }

        const { tools } = data.settingsPrompts.default;
        const userDefined = data.settingsPrompts.userDefined || [];
        const toolEntries: ToolPromptTableData[] = [];

        // Helper function to format tool name
        const formatName = (key: string): string => {
            return key.replaceAll(/([A-Z])/g, ' $1').replace(/^./, (str) => str.toUpperCase());
        };

        // Process each tool
        Object.entries(tools).forEach(([key, prompt]) => {
            if (key === '__typename') {
                return;
            }

            const toolType = (prompt as DefaultPrompt)?.type;
            const hasCustomTool = userDefined.some((p) => p.type === toolType);

            const toolData: ToolPromptTableData = {
                displayName: formatName(key),
                name: key,
                promptType: toolType,
                status: (prompt as DefaultPrompt)?.template ? (hasCustomTool ? 'Custom' : 'Default') : 'N/A',
                template: (prompt as DefaultPrompt)?.template || '',
            };

            toolEntries.push(toolData);
        });

        return toolEntries.sort((a, b) => a.name.localeCompare(b.name));
    };

    // Agent prompts table columns
    const agentColumns: ColumnDef<AgentPromptTableData>[] = [
        {
            accessorKey: 'displayName',
            cell: ({ row }) => (
                <div className="flex items-center gap-2">
                    <span className="font-medium">{row.original.displayName}</span>
                </div>
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
                        Agent Name
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
            accessorKey: 'systemStatus',
            cell: ({ row }) => {
                const status = row.getValue('systemStatus') as string;

                return (
                    <Badge variant={status === 'Custom' ? 'default' : status === 'Default' ? 'secondary' : 'outline'}>
                        {status}
                    </Badge>
                );
            },
            header: 'System Prompt',
            size: 100,
        },
        {
            accessorKey: 'humanStatus',
            cell: ({ row }) => {
                const status = row.getValue('humanStatus') as string;

                return (
                    <Badge variant={status === 'Custom' ? 'default' : status === 'Default' ? 'secondary' : 'outline'}>
                        {status}
                    </Badge>
                );
            },
            header: 'Human Prompt',
            size: 100,
        },
        {
            cell: ({ row }) => {
                const agent = row.original;

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
                                <DropdownMenuItem onClick={() => handlePromptEdit(agent.name)}>
                                    <Pencil className="size-3" />
                                    Edit
                                </DropdownMenuItem>
                                {(canResetPrompt(agent.name, 'system') ||
                                    canResetPrompt(agent.name, 'human') ||
                                    canResetPrompt(agent.name, 'all')) && <DropdownMenuSeparator />}
                                {canResetPrompt(agent.name, 'system') && (
                                    <DropdownMenuItem
                                        disabled={
                                            isDeleteLoading &&
                                            resetOperation?.promptName === agent.name &&
                                            resetOperation?.type === 'system'
                                        }
                                        onClick={() => handleResetDialogOpen('system', agent.name, agent.displayName)}
                                    >
                                        {isDeleteLoading &&
                                        resetOperation?.promptName === agent.name &&
                                        resetOperation?.type === 'system' ? (
                                            <>
                                                <Loader2 className="size-3 animate-spin" />
                                                Resetting...
                                            </>
                                        ) : (
                                            <>
                                                <RotateCcw className="size-3" />
                                                Reset System
                                            </>
                                        )}
                                    </DropdownMenuItem>
                                )}
                                {agent.hasHuman && canResetPrompt(agent.name, 'human') && (
                                    <DropdownMenuItem
                                        disabled={
                                            isDeleteLoading &&
                                            resetOperation?.promptName === agent.name &&
                                            resetOperation?.type === 'human'
                                        }
                                        onClick={() => handleResetDialogOpen('human', agent.name, agent.displayName)}
                                    >
                                        {isDeleteLoading &&
                                        resetOperation?.promptName === agent.name &&
                                        resetOperation?.type === 'human' ? (
                                            <>
                                                <Loader2 className="size-3 animate-spin" />
                                                Resetting...
                                            </>
                                        ) : (
                                            <>
                                                <RotateCcw className="size-3" />
                                                Reset Human
                                            </>
                                        )}
                                    </DropdownMenuItem>
                                )}
                                {canResetPrompt(agent.name, 'all') && (
                                    <DropdownMenuItem
                                        disabled={
                                            isDeleteLoading &&
                                            resetOperation?.promptName === agent.name &&
                                            resetOperation?.type === 'all'
                                        }
                                        onClick={() => handleResetDialogOpen('all', agent.name, agent.displayName)}
                                    >
                                        {isDeleteLoading &&
                                        resetOperation?.promptName === agent.name &&
                                        resetOperation?.type === 'all' ? (
                                            <>
                                                <Loader2 className="size-3 animate-spin" />
                                                Resetting...
                                            </>
                                        ) : (
                                            <>
                                                <Trash2 className="size-3" />
                                                Reset All
                                            </>
                                        )}
                                    </DropdownMenuItem>
                                )}
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

    // Tool prompts table columns
    const toolColumns: ColumnDef<ToolPromptTableData>[] = [
        {
            accessorKey: 'displayName',
            cell: ({ row }) => (
                <div className="flex items-center gap-2">
                    <span className="font-medium">{row.original.displayName}</span>
                </div>
            ),
            enableHiding: false,
            header: ({ column }) => {
                const sorted = column.getIsSorted();

                return (
                    <Button
                        className="text-muted-foreground hover:text-primary flex items-center gap-2 p-0 hover:no-underline"
                        onClick={() => handleColumnSort(column)}
                        variant="link"
                    >
                        Tool Name
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
            accessorKey: 'status',
            cell: ({ row }) => {
                const status = row.getValue('status') as string;

                return (
                    <Badge variant={status === 'Custom' ? 'default' : status === 'Default' ? 'secondary' : 'outline'}>
                        {status}
                    </Badge>
                );
            },
            header: 'Prompt',
            size: 100,
        },
        {
            cell: ({ row }) => {
                const tool = row.original;

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
                                <DropdownMenuItem onClick={() => handlePromptEdit(tool.name)}>
                                    <Pencil className="size-3" />
                                    Edit
                                </DropdownMenuItem>
                                {canResetPrompt(tool.name, 'tool') && (
                                    <>
                                        <DropdownMenuSeparator />
                                        <DropdownMenuItem
                                            disabled={
                                                isDeleteLoading &&
                                                resetOperation?.promptName === tool.name &&
                                                resetOperation?.type === 'tool'
                                            }
                                            onClick={() => handleResetDialogOpen('tool', tool.name, tool.displayName)}
                                        >
                                            {isDeleteLoading &&
                                            resetOperation?.promptName === tool.name &&
                                            resetOperation?.type === 'tool' ? (
                                                <>
                                                    <Loader2 className="size-3 animate-spin" />
                                                    Resetting...
                                                </>
                                            ) : (
                                                <>
                                                    <RotateCcw className="size-3" />
                                                    Reset
                                                </>
                                            )}
                                        </DropdownMenuItem>
                                    </>
                                )}
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

    // Render sub-component for agent prompts
    const renderAgentSubComponent = ({ row }: { row: any }) => {
        const agent = row.original as AgentPromptTableData;

        // Find userDefined prompts for this agent type
        const userSystemPrompt = data?.settingsPrompts?.userDefined?.find((p) => p.type === agent.systemType);
        const userHumanPrompt = data?.settingsPrompts?.userDefined?.find((p) => p.type === agent.humanType);

        // Use userDefined templates if available, otherwise use default
        const systemTemplate = userSystemPrompt?.template || agent.systemTemplate;
        const humanTemplate = userHumanPrompt?.template || agent.humanTemplate;

        return (
            <div className="bg-muted/20 flex flex-col gap-4 border-t p-4">
                <h4 className="font-medium">Prompt Templates</h4>
                <hr className="border-muted-foreground/20" />

                <div className="flex flex-col gap-4">
                    {agent.hasSystem && (
                        <div>
                            <h5 className="mb-2 flex items-center gap-2 text-sm font-medium">
                                <Code className="size-3" />
                                System Prompt
                                {userSystemPrompt && (
                                    <Badge
                                        className="text-xs"
                                        variant="secondary"
                                    >
                                        Custom
                                    </Badge>
                                )}
                            </h5>
                            <pre className="bg-muted max-h-64 overflow-auto rounded-md p-3 text-xs whitespace-pre-wrap">
                                {systemTemplate}
                            </pre>
                        </div>
                    )}

                    {agent.hasHuman && humanTemplate && (
                        <div>
                            <h5 className="mb-2 flex items-center gap-2 text-sm font-medium">
                                <User className="size-3" />
                                Human Prompt
                                {userHumanPrompt && (
                                    <Badge
                                        className="text-xs"
                                        variant="secondary"
                                    >
                                        Custom
                                    </Badge>
                                )}
                            </h5>
                            <pre className="bg-muted max-h-64 overflow-auto rounded-md p-3 text-xs whitespace-pre-wrap">
                                {humanTemplate}
                            </pre>
                        </div>
                    )}
                </div>
            </div>
        );
    };

    // Render sub-component for tool prompts
    const renderToolSubComponent = ({ row }: { row: any }) => {
        const tool = row.original as ToolPromptTableData;

        // Find userDefined prompt for this tool type
        const userToolPrompt = data?.settingsPrompts?.userDefined?.find((p) => p.type === tool.promptType);

        // Use userDefined template if available, otherwise use default
        const template = userToolPrompt?.template || tool.template;

        return (
            <div className="bg-muted/20 border-t p-4">
                <div className="mb-2 flex items-center gap-2">
                    <h5 className="text-sm font-medium">Template</h5>
                    {userToolPrompt && (
                        <Badge
                            className="text-xs"
                            variant="secondary"
                        >
                            Custom
                        </Badge>
                    )}
                </div>
                <pre className="bg-muted max-h-64 overflow-auto rounded-md p-3 text-xs whitespace-pre-wrap">
                    {template}
                </pre>
            </div>
        );
    };

    if (isLoading) {
        return (
            <div className="flex flex-col gap-4">
                <SettingsPromptsHeader />
                <StatusCard
                    description="Please wait while we fetch your prompt templates"
                    icon={<Loader2 className="text-muted-foreground size-16 animate-spin" />}
                    title="Loading prompts..."
                />
            </div>
        );
    }

    if (error) {
        return (
            <div className="flex flex-col gap-4">
                <SettingsPromptsHeader />
                <Alert variant="destructive">
                    <AlertCircle className="size-4" />
                    <AlertTitle>Error loading prompts</AlertTitle>
                    <AlertDescription>{error.message}</AlertDescription>
                </Alert>
            </div>
        );
    }

    const agentPrompts = getAgentPromptsData();
    const toolPrompts = getToolPromptsData();

    if (agentPrompts.length === 0 && toolPrompts.length === 0) {
        return (
            <div className="flex flex-col gap-4">
                <SettingsPromptsHeader />
                <StatusCard
                    description="Prompt templates could not be loaded"
                    icon={<Settings className="text-muted-foreground size-8" />}
                    title="No prompts available"
                />
            </div>
        );
    }

    return (
        <Fragment>
            <div className="flex flex-col gap-6">
                <SettingsPromptsHeader />

                {/* Agent Prompts Section */}
                {agentPrompts.length > 0 && (
                    <div className="flex flex-col gap-2">
                        <div className="flex items-center gap-2">
                            <Bot className="text-muted-foreground size-5" />
                            <h2 className="text-lg font-semibold">Agent Prompts</h2>
                            <Badge variant="secondary">{agentPrompts.length}</Badge>
                        </div>
                        <p className="text-muted-foreground text-sm">System and human prompts for AI agents</p>
                        <DataTable<AgentPromptTableData>
                            columns={agentColumns}
                            columnVisibility={agentColumnVisibility}
                            data={agentPrompts}
                            filterColumn="displayName"
                            filterPlaceholder="Filter agent names..."
                            initialPageSize={1000}
                            onColumnVisibilityChange={(visibility) => {
                                Object.entries(visibility).forEach(([columnId, isVisible]) => {
                                    if (agentColumnVisibility[columnId] !== isVisible) {
                                        updateAgentColumnVisibility(columnId, isVisible);
                                    }
                                });
                            }}
                            renderSubComponent={renderAgentSubComponent}
                            tableKey="prompts-agents"
                        />
                    </div>
                )}

                {/* Tool Prompts Section */}
                {toolPrompts.length > 0 && (
                    <div className="flex flex-col gap-2">
                        <div className="flex items-center gap-2">
                            <Wrench className="text-muted-foreground size-5" />
                            <h2 className="text-lg font-semibold">Tool Prompts</h2>
                            <Badge variant="secondary">{toolPrompts.length}</Badge>
                        </div>
                        <p className="text-muted-foreground text-sm">Prompt templates for system tools and utilities</p>
                        <DataTable<ToolPromptTableData>
                            columns={toolColumns}
                            columnVisibility={toolColumnVisibility}
                            data={toolPrompts}
                            filterColumn="displayName"
                            filterPlaceholder="Filter tool names..."
                            initialPageSize={1000}
                            onColumnVisibilityChange={(visibility) => {
                                Object.entries(visibility).forEach(([columnId, isVisible]) => {
                                    if (toolColumnVisibility[columnId] !== isVisible) {
                                        updateToolColumnVisibility(columnId, isVisible);
                                    }
                                });
                            }}
                            renderSubComponent={renderToolSubComponent}
                            tableKey="prompts-tools"
                        />
                    </div>
                )}
            </div>

            <ConfirmationDialog
                cancelText="Cancel"
                cancelVariant="outline"
                confirmIcon={<RotateCcw />}
                confirmText="Reset"
                confirmVariant="destructive"
                description={
                    resetOperation?.type === 'system'
                        ? `Are you sure you want to reset the system prompt for "${resetOperation.displayName}"? This will revert it to the default template and cannot be undone.`
                        : resetOperation?.type === 'human'
                          ? `Are you sure you want to reset the human prompt for "${resetOperation.displayName}"? This will revert it to the default template and cannot be undone.`
                          : resetOperation?.type === 'all'
                            ? `Are you sure you want to reset all prompts for "${resetOperation.displayName}"? This will revert both system and human prompts to their default templates and cannot be undone.`
                            : `Are you sure you want to reset the prompt for "${resetOperation?.displayName}"? This will revert it to the default template and cannot be undone.`
                }
                handleConfirm={handleResetPrompt}
                handleOpenChange={setResetDialogOpen}
                isOpen={resetDialogOpen}
                title={`Reset ${resetOperation?.displayName || 'Prompt'}`}
            />
        </Fragment>
    );
};

export default SettingsPrompts;
