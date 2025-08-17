import ConfirmationDialog from '@/components/ConfirmationDialog';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader } from '@/components/ui/card';
import { Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { StatusCard } from '@/components/ui/status-card';
import { Switch } from '@/components/ui/switch';
import { zodResolver } from '@hookform/resolvers/zod';
import { Loader2, Play, Plus, Save, Server, Trash2 } from 'lucide-react';
import { Fragment, useMemo, useState } from 'react';
import { Controller, useFieldArray, useForm } from 'react-hook-form';
import { useNavigate, useParams } from 'react-router-dom';
import { z } from 'zod';

type McpTransport = 'stdio' | 'sse';

interface KeyValuePair {
    key: string;
    value: string;
}

const keyValueSchema = z.object({
    key: z.string().min(1, 'Key is required'),
    value: z.string().min(1, 'Value is required'),
});

const formSchema = z.object({
    name: z
        .string({ required_error: 'Name is required' })
        .min(1, 'Name is required')
        .max(50, 'Maximum 50 characters allowed'),
    transport: z.enum(['stdio', 'sse'], { required_error: 'Transport is required' }),
    stdio: z
        .object({
            command: z.string().min(1, 'Command is required'),
            args: z.string().optional().nullable(),
            env: z.array(keyValueSchema).optional().default([]),
        })
        .optional(),
    sse: z
        .object({
            url: z.string().min(1, 'URL is required'),
            headers: z.array(keyValueSchema).optional().default([]),
        })
        .optional(),
    tools: z
        .array(
            z.object({
                name: z.string().min(1, 'Tool name is required'),
                description: z.string().optional(),
                enabled: z.boolean().optional().default(true),
            }),
        )
        .default([]),
});

type FormData = z.infer<typeof formSchema>;

// Mock helpers
const getMockServerById = (id: number) => {
    const samples = [
        {
            id: 1,
            name: 'Local Filesystem',
            transport: 'stdio' as McpTransport,
            stdio: {
                command: '/usr/local/bin/node',
                args: '/opt/mcp/filesystem/index.js --root /Users/sirozha/Projects',
                env: [{ key: 'NODE_ENV', value: 'production' }],
            },
            sse: undefined,
            tools: [
                { name: 'readFile', description: 'Read a file from disk', enabled: true },
                { name: 'writeFile', description: 'Write content to a file', enabled: false },
            ],
        },
        {
            id: 2,
            name: 'Slack (Prod)',
            transport: 'sse' as McpTransport,
            sse: {
                url: 'https://mcp.example.com/slack/sse',
                headers: [{ key: 'Authorization', value: 'Bearer ***' }],
            },
            stdio: undefined,
            tools: [
                { name: 'postMessage', description: 'Send a message to a channel', enabled: true },
                { name: 'getUserInfo', description: 'Fetch Slack user info', enabled: false },
            ],
        },
    ];
    return samples.find((s) => s.id === id);
};

const SettingsMcpServer = () => {
    const navigate = useNavigate();
    const params = useParams();
    const isNew = params.mcpServerId === undefined;
    const [submitError, setSubmitError] = useState<string | null>(null);
    const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
    const [isTestLoading, setIsTestLoading] = useState(false);
    const [testMessage, setTestMessage] = useState<string | null>(null);
    const [testError, setTestError] = useState<string | null>(null);
    const [toolTestLoadingIndex, setToolTestLoadingIndex] = useState<number | null>(null);
    const [toolTestIndex, setToolTestIndex] = useState<number | null>(null);
    const [toolTestMessage, setToolTestMessage] = useState<string | null>(null);
    const [toolTestError, setToolTestError] = useState<string | null>(null);

    const defaults: FormData = useMemo(() => {
        if (!isNew) {
            const id = Number(params.mcpServerId);
            const found = !Number.isNaN(id) ? getMockServerById(id) : undefined;
            if (found) {
                return {
                    name: found.name,
                    transport: found.transport,
                    stdio: found.stdio,
                    sse: found.sse,
                    tools: found.tools,
                } as FormData;
            }
        }
        return {
            name: '',
            transport: 'stdio',
            stdio: { command: '', args: '', env: [] },
            tools: [],
        } as FormData;
    }, [isNew, params.mcpServerId]);

    const form = useForm<FormData>({
        resolver: zodResolver(formSchema),
        defaultValues: defaults,
        mode: 'onChange',
    });

    const transport = form.watch('transport');

    // Field arrays
    const stdioEnvArray = useFieldArray({ control: form.control, name: 'stdio.env' as const });
    const sseHeadersArray = useFieldArray({ control: form.control, name: 'sse.headers' as const });
    const toolsArray = useFieldArray({ control: form.control, name: 'tools' as const });

    const handleAddKeyValue = (target: 'env' | 'headers') => {
        if (target === 'env') {
            stdioEnvArray.append({ key: '', value: '' });
        } else {
            sseHeadersArray.append({ key: '', value: '' });
        }
    };

    const handleSubmit = async (data: FormData) => {
        try {
            setSubmitError(null);
            // Simulate request
            await new Promise((r) => setTimeout(r, 400));
            navigate('/settings/mcp-servers');
        } catch (e) {
            setSubmitError('Failed to save MCP server');
        }
    };

    const handleDelete = () => {
        if (isNew) return;
        setIsDeleteDialogOpen(true);
    };

    const handleConfirmDelete = async () => {
        try {
            // Simulate delete
            await new Promise((r) => setTimeout(r, 300));
            navigate('/settings/mcp-servers');
        } catch (_) {
            setSubmitError('Failed to delete MCP server');
        }
    };

    const handleTest = async () => {
        setTestMessage(null);
        setTestError(null);
        // Validate minimal required fields based on transport
        const valid = await form.trigger();
        if (!valid) {
            setTestError('Please fix validation errors before testing');
            return;
        }
        try {
            setIsTestLoading(true);
            // Simulate connectivity test
            await new Promise((r) => setTimeout(r, 600));
            setTestMessage('Connection successful');
        } catch (_) {
            setTestError('Connection failed');
        } finally {
            setIsTestLoading(false);
        }
    };

    const handleTestTool = async (index: number) => {
        setToolTestIndex(index);
        setToolTestLoadingIndex(index);
        setToolTestMessage(null);
        setToolTestError(null);
        try {
            // Basic validation: tool must have a name
            const toolName = form.getValues(`tools.${index}.name` as const) as string | undefined;
            if (!toolName) {
                setToolTestError('Tool name is required');
                return;
            }
            // Simulate tool invocation
            await new Promise((r) => setTimeout(r, 600));
            setToolTestMessage('Tool test passed');
        } catch (_) {
            setToolTestError('Tool test failed');
        } finally {
            setToolTestLoadingIndex(null);
        }
    };

    if (!isNew && !getMockServerById(Number(params.mcpServerId))) {
        return (
            <StatusCard
                icon={<Server className="h-8 w-8 text-muted-foreground" />}
                title="MCP Server not found"
                description="The requested MCP server could not be located in mock data"
                action={
                    <Button
                        variant="secondary"
                        onClick={() => navigate('/settings/mcp-servers')}
                    >
                        Back to list
                    </Button>
                }
            />
        );
    }

    return (
        <Fragment>
            <Card>
                <CardHeader>
                    <CardDescription>
                        {isNew ? 'Configure a new MCP server' : 'Update MCP server settings'}
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <Form {...form}>
                        <form
                            id="mcp-server-form"
                            onSubmit={form.handleSubmit(handleSubmit)}
                            className="space-y-6"
                        >
                            {(submitError || testMessage || testError) && (
                                <Alert variant="destructive">
                                    {submitError && (
                                        <>
                                            <AlertTitle>Error</AlertTitle>
                                            <AlertDescription>{submitError}</AlertDescription>
                                        </>
                                    )}
                                    {testError && (
                                        <>
                                            <AlertTitle>Test Failed</AlertTitle>
                                            <AlertDescription>{testError}</AlertDescription>
                                        </>
                                    )}
                                    {testMessage && !submitError && !testError && (
                                        <>
                                            <AlertTitle>Test Passed</AlertTitle>
                                            <AlertDescription>{testMessage}</AlertDescription>
                                        </>
                                    )}
                                </Alert>
                            )}

                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <FormField
                                    control={form.control}
                                    name="name"
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel>Name</FormLabel>
                                            <FormControl>
                                                <Input
                                                    {...field}
                                                    placeholder="Enter server name"
                                                />
                                            </FormControl>
                                            <FormDescription>A unique name for this MCP server</FormDescription>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />

                                <FormField
                                    control={form.control}
                                    name="transport"
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel>Transport</FormLabel>
                                            <Select
                                                onValueChange={(v: McpTransport) => {
                                                    field.onChange(v);
                                                    // Normalize opposite config to avoid stale values
                                                    if (v === 'stdio') {
                                                        form.setValue('sse', undefined);
                                                        if (!form.getValues('stdio')) {
                                                            form.setValue('stdio', { command: '', args: '', env: [] });
                                                        }
                                                    } else {
                                                        form.setValue('stdio', undefined);
                                                        if (!form.getValues('sse')) {
                                                            form.setValue('sse', { url: '', headers: [] });
                                                        }
                                                    }
                                                }}
                                                defaultValue={field.value}
                                            >
                                                <FormControl>
                                                    <SelectTrigger>
                                                        <SelectValue placeholder="Select transport" />
                                                    </SelectTrigger>
                                                </FormControl>
                                                <SelectContent>
                                                    <SelectItem value="stdio">STDIO</SelectItem>
                                                    <SelectItem value="sse">SSE</SelectItem>
                                                </SelectContent>
                                            </Select>
                                            <FormDescription>
                                                STDIO for local process; SSE for remote URL
                                            </FormDescription>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />
                            </div>

                            {/* STDIO configuration */}
                            {transport === 'stdio' && (
                                <div className="space-y-4">
                                    <h3 className="text-lg font-medium">STDIO Configuration</h3>
                                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                        <FormField
                                            control={form.control}
                                            name="stdio.command"
                                            render={({ field }) => (
                                                <FormItem>
                                                    <FormLabel>Command</FormLabel>
                                                    <FormControl>
                                                        <Input
                                                            {...field}
                                                            placeholder="/usr/local/bin/node"
                                                        />
                                                    </FormControl>
                                                    <FormMessage />
                                                </FormItem>
                                            )}
                                        />
                                        <FormField
                                            control={form.control}
                                            name="stdio.args"
                                            render={({ field }) => (
                                                <FormItem>
                                                    <FormLabel>Args</FormLabel>
                                                    <FormControl>
                                                        <Input
                                                            {...field}
                                                            value={field.value ?? ''}
                                                            placeholder="/path/to/script.js --flag value"
                                                        />
                                                    </FormControl>
                                                    <FormDescription>Space-separated arguments</FormDescription>
                                                    <FormMessage />
                                                </FormItem>
                                            )}
                                        />
                                    </div>

                                    <div>
                                        <div className="flex items-center justify-between mb-2">
                                            <h4 className="text-sm font-medium">Environment Variables</h4>
                                            <Button
                                                type="button"
                                                variant="outline"
                                                size="sm"
                                                onClick={() => handleAddKeyValue('env')}
                                            >
                                                <Plus className="h-3 w-3" /> Add
                                            </Button>
                                        </div>
                                        <div className="space-y-2">
                                            {stdioEnvArray.fields.length === 0 && (
                                                <div className="text-sm text-muted-foreground">No variables</div>
                                            )}
                                            {stdioEnvArray.fields.map((field, index) => (
                                                <div
                                                    key={field.id}
                                                    className="grid grid-cols-1 md:grid-cols-5 gap-2"
                                                >
                                                    <Controller
                                                        control={form.control}
                                                        name={`stdio.env.${index}.key` as const}
                                                        render={({ field }) => (
                                                            <Input
                                                                {...field}
                                                                placeholder="KEY"
                                                                className="md:col-span-2"
                                                            />
                                                        )}
                                                    />
                                                    <Controller
                                                        control={form.control}
                                                        name={`stdio.env.${index}.value` as const}
                                                        render={({ field }) => (
                                                            <Input
                                                                {...field}
                                                                placeholder="VALUE"
                                                                className="md:col-span-2"
                                                            />
                                                        )}
                                                    />
                                                    <Button
                                                        type="button"
                                                        variant="ghost"
                                                        className="justify-self-start"
                                                        onClick={() => stdioEnvArray.remove(index)}
                                                    >
                                                        <Trash2 className="h-4 w-4" />
                                                    </Button>
                                                </div>
                                            ))}
                                        </div>
                                    </div>
                                </div>
                            )}

                            {/* SSE configuration */}
                            {transport === 'sse' && (
                                <div className="space-y-4">
                                    <h3 className="text-lg font-medium">SSE Configuration</h3>
                                    <FormField
                                        control={form.control}
                                        name="sse.url"
                                        render={({ field }) => (
                                            <FormItem>
                                                <FormLabel>URL</FormLabel>
                                                <FormControl>
                                                    <Input
                                                        {...field}
                                                        placeholder="https://mcp.example.com/sse"
                                                    />
                                                </FormControl>
                                                <FormMessage />
                                            </FormItem>
                                        )}
                                    />

                                    <div>
                                        <div className="flex items-center justify-between mb-2">
                                            <h4 className="text-sm font-medium">Headers</h4>
                                            <Button
                                                type="button"
                                                variant="outline"
                                                size="sm"
                                                onClick={() => handleAddKeyValue('headers')}
                                            >
                                                <Plus className="h-3 w-3" /> Add
                                            </Button>
                                        </div>
                                        <div className="space-y-2">
                                            {sseHeadersArray.fields.length === 0 && (
                                                <div className="text-sm text-muted-foreground">No headers</div>
                                            )}
                                            {sseHeadersArray.fields.map((field, index) => (
                                                <div
                                                    key={field.id}
                                                    className="grid grid-cols-1 md:grid-cols-5 gap-2"
                                                >
                                                    <Controller
                                                        control={form.control}
                                                        name={`sse.headers.${index}.key` as const}
                                                        render={({ field }) => (
                                                            <Input
                                                                {...field}
                                                                placeholder="Header"
                                                                className="md:col-span-2"
                                                            />
                                                        )}
                                                    />
                                                    <Controller
                                                        control={form.control}
                                                        name={`sse.headers.${index}.value` as const}
                                                        render={({ field }) => (
                                                            <Input
                                                                {...field}
                                                                placeholder="Value"
                                                                className="md:col-span-2"
                                                            />
                                                        )}
                                                    />
                                                    <Button
                                                        type="button"
                                                        variant="ghost"
                                                        className="justify-self-start"
                                                        onClick={() => sseHeadersArray.remove(index)}
                                                    >
                                                        <Trash2 className="h-4 w-4" />
                                                    </Button>
                                                </div>
                                            ))}
                                        </div>
                                    </div>
                                </div>
                            )}

                            {/* Tools configuration - only for existing servers; toggles only */}
                            {!isNew && (
                                <div className="space-y-4">
                                    <div>
                                        <h3 className="text-lg font-medium">Tools</h3>
                                        <p className="text-sm text-muted-foreground">
                                            Enable or disable available tools
                                        </p>
                                    </div>
                                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
                                        {toolsArray.fields.length === 0 && (
                                            <div className="text-sm text-muted-foreground">No tools</div>
                                        )}
                                        {toolsArray.fields.map((tool, index) => (
                                            <div
                                                key={tool.id}
                                                className="flex flex-col gap-2 rounded-md border p-2"
                                            >
                                                <div className="flex items-start justify-between gap-4">
                                                    <div className="flex-1 text-sm">
                                                        <div className="font-medium truncate">
                                                            {form.watch(`tools.${index}.name`) || 'tool'}
                                                        </div>
                                                        {form.watch(`tools.${index}.description`) && (
                                                            <div className="text-muted-foreground">
                                                                {form.watch(`tools.${index}.description`) as string}
                                                            </div>
                                                        )}
                                                    </div>
                                                    <div className="flex items-center gap-2">
                                                        <span className="text-xs text-muted-foreground">Enabled</span>
                                                        <Controller
                                                            control={form.control}
                                                            name={`tools.${index}.enabled` as const}
                                                            render={({ field }) => (
                                                                <Switch
                                                                    checked={!!field.value}
                                                                    onCheckedChange={field.onChange}
                                                                    aria-label={`Toggle ${form.getValues(`tools.${index}.name`) || 'tool'}`}
                                                                />
                                                            )}
                                                        />
                                                        <Button
                                                            type="button"
                                                            size="sm"
                                                            variant="outline"
                                                            onClick={() => handleTestTool(index)}
                                                            disabled={toolTestLoadingIndex === index}
                                                        >
                                                            {toolTestLoadingIndex === index ? (
                                                                <Loader2 className="h-3 w-3 animate-spin" />
                                                            ) : (
                                                                <Play className="h-3 w-3" />
                                                            )}
                                                            {toolTestLoadingIndex === index ? 'Testing...' : 'Test'}
                                                        </Button>
                                                    </div>
                                                </div>
                                                {toolTestIndex === index && (toolTestMessage || toolTestError) && (
                                                    <div className="text-xs mt-1">
                                                        {toolTestMessage && (
                                                            <span className="text-green-600">{toolTestMessage}</span>
                                                        )}
                                                        {toolTestError && (
                                                            <span className="text-red-600">{toolTestError}</span>
                                                        )}
                                                    </div>
                                                )}
                                            </div>
                                        ))}
                                    </div>
                                </div>
                            )}
                        </form>
                    </Form>
                </CardContent>
            </Card>

            {/* Sticky buttons */}
            <div className="flex items-center sticky -bottom-4 bg-background border-t mt-4 -mx-4 -mb-4 p-4 shadow-lg">
                <div className="flex space-x-2">
                    {!isNew && (
                        <Button
                            type="button"
                            variant="destructive"
                            onClick={handleDelete}
                        >
                            <Trash2 className="h-4 w-4" />
                            Delete
                        </Button>
                    )}
                    <Button
                        type="button"
                        variant="outline"
                        onClick={handleTest}
                        disabled={isTestLoading}
                    >
                        {isTestLoading ? <Loader2 className="h-4 w-4 animate-spin" /> : <Play className="h-4 w-4" />}
                        {isTestLoading ? 'Testing...' : 'Test'}
                    </Button>
                </div>
                <div className="flex space-x-2 ml-auto">
                    <Button
                        type="button"
                        variant="outline"
                        onClick={() => navigate('/settings/mcp-servers')}
                    >
                        Cancel
                    </Button>
                    <Button
                        form="mcp-server-form"
                        variant="secondary"
                        type="submit"
                        disabled={form.formState.isSubmitting}
                    >
                        {form.formState.isSubmitting ? (
                            <Loader2 className="h-4 w-4 animate-spin" />
                        ) : (
                            <Save className="h-4 w-4" />
                        )}
                        {form.formState.isSubmitting ? 'Saving...' : isNew ? 'Create MCP Server' : 'Update MCP Server'}
                    </Button>
                </div>
            </div>

            <ConfirmationDialog
                isOpen={isDeleteDialogOpen}
                handleOpenChange={setIsDeleteDialogOpen}
                handleConfirm={handleConfirmDelete}
                itemName={form.watch('name')}
                itemType="MCP server"
                confirmText="Delete"
                cancelText="Cancel"
            />
        </Fragment>
    );
};

export default SettingsMcpServer;
