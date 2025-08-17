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
import {
    useDeletePromptMutation,
    useSettingsPromptsQuery,
    type AgentPrompt,
    type AgentPrompts,
    type DefaultPrompt,
    type PromptType,
} from '@/graphql/types';
import { type ColumnDef } from '@tanstack/react-table';
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

// Types for table data
type AgentPromptTableData = {
    name: string; // Original key (camelCase)
    displayName: string; // Formatted display name
    hasSystem: boolean;
    hasHuman: boolean;
    systemTemplate: string;
    humanTemplate?: string;
    systemStatus: 'Default' | 'Custom' | 'N/A';
    humanStatus: 'Default' | 'Custom' | 'N/A';
    systemType?: PromptType; // Type for system prompt lookup
    humanType?: PromptType; // Type for human prompt lookup
};

type ToolPromptTableData = {
    name: string; // Original key (camelCase)
    displayName: string; // Formatted display name
    template: string;
    status: 'Default' | 'Custom' | 'N/A';
    promptType?: PromptType; // Type for prompt lookup
};

const SettingsPromptsHeader = () => {
    return (
        <div className="flex items-center justify-between">
            <p className="text-muted-foreground">Manage system and custom prompt templates</p>
        </div>
    );
};

const SettingsPrompts = () => {
    const { data, loading: isLoading, error } = useSettingsPromptsQuery();
    const [deletePrompt, { loading: isDeleteLoading }] = useDeletePromptMutation();
    const navigate = useNavigate();

    // Reset dialog states
    const [resetDialogOpen, setResetDialogOpen] = useState(false);
    const [resetOperation, setResetOperation] = useState<{
        type: 'system' | 'human' | 'all' | 'tool';
        promptName: string;
        displayName: string;
    } | null>(null);

    // Handler for editing any prompt (agent or tool)
    const handlePromptEdit = (promptName: string) => {
        navigate(`/settings/prompts/${promptName}`);
    };

    // Reset dialog handlers
    const handleResetDialogOpen = (
        type: 'system' | 'human' | 'all' | 'tool',
        promptName: string,
        displayName: string,
    ) => {
        setResetOperation({ type, promptName, displayName });
        setResetDialogOpen(true);
    };

    const handleResetPrompt = async () => {
        if (!resetOperation || !data?.settingsPrompts?.default) return;

        try {
            const { promptName, type } = resetOperation;
            const agents = data.settingsPrompts.default.agents;
            const tools = data.settingsPrompts.default.tools;
            const userDefined = data.settingsPrompts.userDefined || [];

            if (type === 'tool') {
                // Handle tool prompt reset
                const toolPrompt = tools?.[promptName as keyof typeof tools];
                if (toolPrompt?.type) {
                    // Find the user-defined prompt with matching type
                    const userPrompt = userDefined.find((p) => p.type === toolPrompt.type);
                    if (userPrompt) {
                        await deletePrompt({
                            variables: { promptId: userPrompt.id },
                            refetchQueries: ['settingsPrompts'],
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
                                variables: { promptId: userPrompt.id },
                                refetchQueries: ['settingsPrompts'],
                            });
                        }
                    } else if (type === 'human' && humanType) {
                        const userPrompt = userDefined.find((p) => p.type === humanType);
                        if (userPrompt) {
                            await deletePrompt({
                                variables: { promptId: userPrompt.id },
                                refetchQueries: ['settingsPrompts'],
                            });
                        }
                    } else if (type === 'all') {
                        if (systemType) {
                            const userSystemPrompt = userDefined.find((p) => p.type === systemType);
                            if (userSystemPrompt) {
                                await deletePrompt({
                                    variables: { promptId: userSystemPrompt.id },
                                    refetchQueries: ['settingsPrompts'],
                                });
                            }
                        }
                        if (humanType) {
                            const userHumanPrompt = userDefined.find((p) => p.type === humanType);
                            if (userHumanPrompt) {
                                await deletePrompt({
                                    variables: { promptId: userHumanPrompt.id },
                                    refetchQueries: ['settingsPrompts'],
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
    const canResetPrompt = (promptName: string, resetType: 'system' | 'human' | 'all' | 'tool'): boolean => {
        if (!data?.settingsPrompts?.default || !data?.settingsPrompts?.userDefined) return false;

        const userDefined = data.settingsPrompts.userDefined;
        const agents = data.settingsPrompts.default.agents;
        const tools = data.settingsPrompts.default.tools;

        if (resetType === 'tool') {
            const toolPrompt = tools?.[promptName as keyof typeof tools];
            return toolPrompt?.type ? userDefined.some((p) => p.type === toolPrompt.type) : false;
        } else {
            const agentPrompts = agents?.[promptName as keyof typeof agents] as AgentPrompts;
            if (!agentPrompts) return false;

            const systemType = agentPrompts.system?.type;
            const humanType = agentPrompts.human?.type;

            if (resetType === 'system') {
                return systemType ? userDefined.some((p) => p.type === systemType) : false;
            } else if (resetType === 'human') {
                return humanType ? userDefined.some((p) => p.type === humanType) : false;
            } else if (resetType === 'all') {
                const hasCustomSystem = systemType ? userDefined.some((p) => p.type === systemType) : false;
                const hasCustomHuman = humanType ? userDefined.some((p) => p.type === humanType) : false;
                return hasCustomSystem || hasCustomHuman;
            }
        }
        return false;
    };

    // Transform agents data for table
    const getAgentPromptsData = (): AgentPromptTableData[] => {
        if (!data?.settingsPrompts?.default?.agents) return [];

        const agents = data.settingsPrompts.default.agents;
        const userDefined = data.settingsPrompts.userDefined || [];
        const agentEntries: AgentPromptTableData[] = [];

        // Helper function to format agent name
        const formatName = (key: string): string => {
            return key.replace(/([A-Z])/g, ' $1').replace(/^./, (str) => str.toUpperCase());
        };

        // Process each agent
        Object.entries(agents).forEach(([key, prompts]) => {
            if (key === '__typename') return;

            const systemType = (prompts as AgentPrompts | AgentPrompt)?.system?.type;
            const humanType = (prompts as AgentPrompts)?.human?.type;

            // Check if user has custom prompts
            const hasCustomSystem = userDefined.some((p) => p.type === systemType);
            const hasCustomHuman = humanType ? userDefined.some((p) => p.type === humanType) : false;

            const agentData: AgentPromptTableData = {
                name: key,
                displayName: formatName(key),
                hasSystem: !!(prompts as AgentPrompts | AgentPrompt)?.system,
                hasHuman: !!(prompts as AgentPrompts)?.human,
                systemTemplate: (prompts as AgentPrompts | AgentPrompt)?.system?.template || '',
                humanTemplate: (prompts as AgentPrompts)?.human?.template,
                systemStatus: !!(prompts as AgentPrompts | AgentPrompt)?.system
                    ? hasCustomSystem
                        ? 'Custom'
                        : 'Default'
                    : 'N/A',
                humanStatus: !!(prompts as AgentPrompts)?.human ? (hasCustomHuman ? 'Custom' : 'Default') : 'N/A',
                systemType,
                humanType,
            };

            agentEntries.push(agentData);
        });

        return agentEntries.sort((a, b) => a.name.localeCompare(b.name));
    };

    // Transform tools data for table
    const getToolPromptsData = (): ToolPromptTableData[] => {
        if (!data?.settingsPrompts?.default?.tools) return [];

        const tools = data.settingsPrompts.default.tools;
        const userDefined = data.settingsPrompts.userDefined || [];
        const toolEntries: ToolPromptTableData[] = [];

        // Helper function to format tool name
        const formatName = (key: string): string => {
            return key.replace(/([A-Z])/g, ' $1').replace(/^./, (str) => str.toUpperCase());
        };

        // Process each tool
        Object.entries(tools).forEach(([key, prompt]) => {
            if (key === '__typename') return;

            const toolType = (prompt as DefaultPrompt)?.type;
            const hasCustomTool = userDefined.some((p) => p.type === toolType);

            const toolData: ToolPromptTableData = {
                name: key,
                displayName: formatName(key),
                template: (prompt as DefaultPrompt)?.template || '',
                status: (prompt as DefaultPrompt)?.template ? (hasCustomTool ? 'Custom' : 'Default') : 'N/A',
                promptType: toolType,
            };

            toolEntries.push(toolData);
        });

        return toolEntries.sort((a, b) => a.name.localeCompare(b.name));
    };

    // Agent prompts table columns
    const agentColumns: ColumnDef<AgentPromptTableData>[] = [
        {
            accessorKey: 'displayName',
            size: 200,
            header: ({ column }) => {
                const sorted = column.getIsSorted();

                return (
                    <Button
                        variant="link"
                        onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
                        className="flex items-center gap-2 p-0 no-underline hover:no-underline text-muted-foreground hover:text-primary"
                    >
                        Agent Name
                        {sorted === 'asc' ? (
                            <ArrowDown className="h-4 w-4" />
                        ) : sorted === 'desc' ? (
                            <ArrowUp className="h-4 w-4" />
                        ) : null}
                    </Button>
                );
            },
            cell: ({ row }) => (
                <div className="flex items-center gap-2">
                    <span className="font-medium">{row.original.displayName}</span>
                </div>
            ),
        },
        {
            accessorKey: 'systemStatus',
            size: 100,
            header: 'System Prompt',
            cell: ({ row }) => {
                const status = row.getValue('systemStatus') as string;
                return (
                    <Badge variant={status === 'Custom' ? 'default' : status === 'Default' ? 'secondary' : 'outline'}>
                        {status}
                    </Badge>
                );
            },
        },
        {
            accessorKey: 'humanStatus',
            size: 100,
            header: 'Human Prompt',
            cell: ({ row }) => {
                const status = row.getValue('humanStatus') as string;
                return (
                    <Badge variant={status === 'Custom' ? 'default' : status === 'Default' ? 'secondary' : 'outline'}>
                        {status}
                    </Badge>
                );
            },
        },
        {
            id: 'actions',
            size: 48,
            enableHiding: false,
            header: () => null,
            cell: ({ row }) => {
                const agent = row.original;

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
                                <DropdownMenuItem onClick={() => handlePromptEdit(agent.name)}>
                                    <Pencil className="h-3 w-3" />
                                    Edit
                                </DropdownMenuItem>
                                {(canResetPrompt(agent.name, 'system') ||
                                    canResetPrompt(agent.name, 'human') ||
                                    canResetPrompt(agent.name, 'all')) && <DropdownMenuSeparator />}
                                {canResetPrompt(agent.name, 'system') && (
                                    <DropdownMenuItem
                                        onClick={() => handleResetDialogOpen('system', agent.name, agent.displayName)}
                                        disabled={
                                            isDeleteLoading &&
                                            resetOperation?.promptName === agent.name &&
                                            resetOperation?.type === 'system'
                                        }
                                    >
                                        {isDeleteLoading &&
                                        resetOperation?.promptName === agent.name &&
                                        resetOperation?.type === 'system' ? (
                                            <>
                                                <Loader2 className="h-3 w-3 animate-spin" />
                                                Resetting...
                                            </>
                                        ) : (
                                            <>
                                                <RotateCcw className="h-3 w-3" />
                                                Reset System
                                            </>
                                        )}
                                    </DropdownMenuItem>
                                )}
                                {agent.hasHuman && canResetPrompt(agent.name, 'human') && (
                                    <DropdownMenuItem
                                        onClick={() => handleResetDialogOpen('human', agent.name, agent.displayName)}
                                        disabled={
                                            isDeleteLoading &&
                                            resetOperation?.promptName === agent.name &&
                                            resetOperation?.type === 'human'
                                        }
                                    >
                                        {isDeleteLoading &&
                                        resetOperation?.promptName === agent.name &&
                                        resetOperation?.type === 'human' ? (
                                            <>
                                                <Loader2 className="h-3 w-3 animate-spin" />
                                                Resetting...
                                            </>
                                        ) : (
                                            <>
                                                <RotateCcw className="h-3 w-3" />
                                                Reset Human
                                            </>
                                        )}
                                    </DropdownMenuItem>
                                )}
                                {canResetPrompt(agent.name, 'all') && (
                                    <DropdownMenuItem
                                        onClick={() => handleResetDialogOpen('all', agent.name, agent.displayName)}
                                        disabled={
                                            isDeleteLoading &&
                                            resetOperation?.promptName === agent.name &&
                                            resetOperation?.type === 'all'
                                        }
                                    >
                                        {isDeleteLoading &&
                                        resetOperation?.promptName === agent.name &&
                                        resetOperation?.type === 'all' ? (
                                            <>
                                                <Loader2 className="h-3 w-3 animate-spin" />
                                                Resetting...
                                            </>
                                        ) : (
                                            <>
                                                <Trash2 className="h-3 w-3" />
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
        },
    ];

    // Tool prompts table columns
    const toolColumns: ColumnDef<ToolPromptTableData>[] = [
        {
            accessorKey: 'displayName',
            size: 300,
            header: ({ column }) => {
                const sorted = column.getIsSorted();

                return (
                    <Button
                        variant="link"
                        className="flex items-center gap-2 p-0 hover:no-underline text-muted-foreground hover:text-primary"
                        onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
                    >
                        Tool Name
                        {sorted === 'asc' ? (
                            <ArrowDown className="h-4 w-4" />
                        ) : sorted === 'desc' ? (
                            <ArrowUp className="h-4 w-4" />
                        ) : null}
                    </Button>
                );
            },
            cell: ({ row }) => (
                <div className="flex items-center gap-2">
                    <span className="font-medium">{row.original.displayName}</span>
                </div>
            ),
        },
        {
            accessorKey: 'status',
            size: 100,
            header: 'Prompt',
            cell: ({ row }) => {
                const status = row.getValue('status') as string;
                return (
                    <Badge variant={status === 'Custom' ? 'default' : status === 'Default' ? 'secondary' : 'outline'}>
                        {status}
                    </Badge>
                );
            },
        },
        {
            id: 'actions',
            size: 48,
            enableHiding: false,
            header: () => null,
            cell: ({ row }) => {
                const tool = row.original;

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
                                <DropdownMenuItem onClick={() => handlePromptEdit(tool.name)}>
                                    <Pencil className="h-3 w-3" />
                                    Edit
                                </DropdownMenuItem>
                                {canResetPrompt(tool.name, 'tool') && (
                                    <>
                                        <DropdownMenuSeparator />
                                        <DropdownMenuItem
                                            onClick={() => handleResetDialogOpen('tool', tool.name, tool.displayName)}
                                            disabled={
                                                isDeleteLoading &&
                                                resetOperation?.promptName === tool.name &&
                                                resetOperation?.type === 'tool'
                                            }
                                        >
                                            {isDeleteLoading &&
                                            resetOperation?.promptName === tool.name &&
                                            resetOperation?.type === 'tool' ? (
                                                <>
                                                    <Loader2 className="h-3 w-3 animate-spin" />
                                                    Resetting...
                                                </>
                                            ) : (
                                                <>
                                                    <RotateCcw className="h-3 w-3" />
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
            <div className="p-4 bg-muted/20 border-t space-y-4">
                <h4 className="font-medium">Prompt Templates</h4>
                <hr className="border-muted-foreground/20" />

                <div className="space-y-4">
                    {agent.hasSystem && (
                        <div>
                            <h5 className="font-medium text-sm mb-2 flex items-center gap-2">
                                <Code className="h-3 w-3" />
                                System Prompt
                                {userSystemPrompt && (
                                    <Badge
                                        variant="secondary"
                                        className="text-xs"
                                    >
                                        Custom
                                    </Badge>
                                )}
                            </h5>
                            <pre className="text-xs bg-muted p-3 rounded-md overflow-auto max-h-64 whitespace-pre-wrap">
                                {systemTemplate}
                            </pre>
                        </div>
                    )}

                    {agent.hasHuman && humanTemplate && (
                        <div>
                            <h5 className="font-medium text-sm mb-2 flex items-center gap-2">
                                <User className="h-3 w-3" />
                                Human Prompt
                                {userHumanPrompt && (
                                    <Badge
                                        variant="secondary"
                                        className="text-xs"
                                    >
                                        Custom
                                    </Badge>
                                )}
                            </h5>
                            <pre className="text-xs bg-muted p-3 rounded-md overflow-auto max-h-64 whitespace-pre-wrap">
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
            <div className="p-4 bg-muted/20 border-t">
                <div className="mb-2 flex items-center gap-2">
                    <h5 className="font-medium text-sm">Template</h5>
                    {userToolPrompt && (
                        <Badge
                            variant="secondary"
                            className="text-xs"
                        >
                            Custom
                        </Badge>
                    )}
                </div>
                <pre className="text-xs bg-muted p-3 rounded-md overflow-auto max-h-64 whitespace-pre-wrap">
                    {template}
                </pre>
            </div>
        );
    };

    if (isLoading) {
        return (
            <div className="space-y-4">
                <SettingsPromptsHeader />
                <StatusCard
                    icon={<Loader2 className="w-16 h-16 animate-spin text-muted-foreground" />}
                    title="Loading prompts..."
                    description="Please wait while we fetch your prompt templates"
                />
            </div>
        );
    }

    if (error) {
        return (
            <div className="space-y-4">
                <SettingsPromptsHeader />
                <Alert variant="destructive">
                    <AlertCircle className="h-4 w-4" />
                    <AlertTitle>Error loading prompts</AlertTitle>
                    <AlertDescription>{error.message}</AlertDescription>
                </Alert>
            </div>
        );
    }

    const agentPrompts = getAgentPromptsData();
    const toolPrompts = getToolPromptsData();

    if (!agentPrompts.length && !toolPrompts.length) {
        return (
            <div className="space-y-4">
                <SettingsPromptsHeader />
                <StatusCard
                    icon={<Settings className="h-8 w-8 text-muted-foreground" />}
                    title="No prompts available"
                    description="Prompt templates could not be loaded"
                />
            </div>
        );
    }

    return (
        <Fragment>
            <div className="space-y-6">
                <SettingsPromptsHeader />

                {/* Agent Prompts Section */}
                {agentPrompts.length > 0 && (
                    <div className="space-y-2">
                        <div className="flex items-center gap-2">
                            <Bot className="h-5 w-5 text-muted-foreground" />
                            <h2 className="text-lg font-semibold">Agent Prompts</h2>
                            <Badge variant="secondary">{agentPrompts.length}</Badge>
                        </div>
                        <p className="text-sm text-muted-foreground">System and human prompts for AI agents</p>
                        <DataTable
                            columns={agentColumns}
                            data={agentPrompts}
                            renderSubComponent={renderAgentSubComponent}
                            initialPageSize={1000}
                            filterColumn="displayName"
                            filterPlaceholder="Filter agent names..."
                        />
                    </div>
                )}

                {/* Tool Prompts Section */}
                {toolPrompts.length > 0 && (
                    <div className="space-y-2">
                        <div className="flex items-center gap-2">
                            <Wrench className="h-5 w-5 text-muted-foreground" />
                            <h2 className="text-lg font-semibold">Tool Prompts</h2>
                            <Badge variant="secondary">{toolPrompts.length}</Badge>
                        </div>
                        <p className="text-sm text-muted-foreground">Prompt templates for system tools and utilities</p>
                        <DataTable
                            columns={toolColumns}
                            data={toolPrompts}
                            renderSubComponent={renderToolSubComponent}
                            initialPageSize={1000}
                            filterColumn="displayName"
                            filterPlaceholder="Filter tool names..."
                        />
                    </div>
                )}
            </div>

            <ConfirmationDialog
                isOpen={resetDialogOpen}
                handleOpenChange={setResetDialogOpen}
                handleConfirm={handleResetPrompt}
                title={`Reset ${resetOperation?.displayName || 'Prompt'}`}
                description={
                    resetOperation?.type === 'system'
                        ? `Are you sure you want to reset the system prompt for "${resetOperation.displayName}"? This will revert it to the default template and cannot be undone.`
                        : resetOperation?.type === 'human'
                          ? `Are you sure you want to reset the human prompt for "${resetOperation.displayName}"? This will revert it to the default template and cannot be undone.`
                          : resetOperation?.type === 'all'
                            ? `Are you sure you want to reset all prompts for "${resetOperation.displayName}"? This will revert both system and human prompts to their default templates and cannot be undone.`
                            : `Are you sure you want to reset the prompt for "${resetOperation?.displayName}"? This will revert it to the default template and cannot be undone.`
                }
                confirmText="Reset"
                cancelText="Cancel"
                confirmVariant="destructive"
                cancelVariant="outline"
                confirmIcon={<RotateCcw />}
            />
        </Fragment>
    );
};

export default SettingsPrompts;
