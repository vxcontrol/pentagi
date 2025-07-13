import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from '@/components/ui/accordion';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader } from '@/components/ui/card';
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from '@/components/ui/command';
import { Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { StatusCard } from '@/components/ui/status-card';
import {
    ProviderType,
    ReasoningEffort,
    useCreateProviderMutation,
    useDeleteProviderMutation,
    useSettingsProvidersQuery,
    useUpdateProviderMutation,
    type AgentConfigInput,
    type AgentsConfigInput,
    type ProviderConfigFragmentFragment,
} from '@/graphql/types';
import { cn } from '@/lib/utils';
import { zodResolver } from '@hookform/resolvers/zod';
import { AlertCircle, Check, ChevronsUpDown, Loader2, Save, Trash2 } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { useNavigate, useParams } from 'react-router-dom';
import { z } from 'zod';

type Provider = ProviderConfigFragmentFragment;

// Define agent configuration schema
const agentConfigSchema = z
    .object({
        model: z.string(),
        temperature: z.number().nullable().optional(),
        maxTokens: z.number().nullable().optional(),
        topK: z.number().nullable().optional(),
        topP: z.number().nullable().optional(),
        minLength: z.number().nullable().optional(),
        maxLength: z.number().nullable().optional(),
        repetitionPenalty: z.number().nullable().optional(),
        frequencyPenalty: z.number().nullable().optional(),
        presencePenalty: z.number().nullable().optional(),
        reasoning: z
            .object({
                effort: z.string().nullable().optional(),
                maxTokens: z.number().nullable().optional(),
            })
            .nullable()
            .optional(),
        price: z
            .object({
                input: z.number().nullable().optional(),
                output: z.number().nullable().optional(),
            })
            .nullable()
            .optional(),
    })
    .optional();

// Define form schema
const formSchema = z.object({
    type: z.string().min(1, 'Provider type is required'),
    name: z.string().min(1, 'Provider name is required'),
    agents: z
        .object({
            agent: agentConfigSchema,
            assistant: agentConfigSchema,
            coder: agentConfigSchema,
            pentester: agentConfigSchema,
            searcher: agentConfigSchema,
            generator: agentConfigSchema,
            refiner: agentConfigSchema,
            adviser: agentConfigSchema,
            reflector: agentConfigSchema,
            simple: agentConfigSchema,
            simpleJson: agentConfigSchema,
            enricher: agentConfigSchema,
            installer: agentConfigSchema,
        })
        .optional(),
});

type FormData = z.infer<typeof formSchema>;

// Convert camelCase key to display name (e.g., 'simpleJson' -> 'Simple Json')
const getName = (key: string): string => key.replace(/([A-Z])/g, ' $1').replace(/^./, (item) => item.toUpperCase());

// All agent types (updated to include enricher and installer)
const agentTypes = [
    'agent',
    'assistant',
    'coder',
    'pentester',
    'searcher',
    'generator',
    'refiner',
    'adviser',
    'reflector',
    'simple',
    'simpleJson',
    'enricher',
    'installer',
] as const;

// Helper function to convert string to ReasoningEffort enum
const getReasoningEffort = (effort: string | null | undefined): ReasoningEffort | null => {
    if (!effort) return null;

    switch (effort.toLowerCase()) {
        case 'low':
            return ReasoningEffort.Low;
        case 'medium':
            return ReasoningEffort.Medium;
        case 'high':
            return ReasoningEffort.High;
        default:
            return null;
    }
};

// Helper function to convert form data to GraphQL input
const transformFormToGraphQL = (
    formData: FormData,
): {
    name: string;
    type: ProviderType;
    agents: AgentsConfigInput;
} => {
    const agents: AgentsConfigInput = {} as AgentsConfigInput;

    // Process each agent
    agentTypes.forEach((agentKey) => {
        const agentData = formData.agents?.[agentKey];
        if (agentData) {
            const agentConfig: AgentConfigInput = {
                model: agentData.model || '',
                temperature: agentData.temperature || null,
                maxTokens: agentData.maxTokens || null,
                topK: agentData.topK || null,
                topP: agentData.topP || null,
                minLength: agentData.minLength || null,
                maxLength: agentData.maxLength || null,
                repetitionPenalty: agentData.repetitionPenalty || null,
                frequencyPenalty: agentData.frequencyPenalty || null,
                presencePenalty: agentData.presencePenalty || null,
                reasoning: agentData.reasoning
                    ? {
                          effort: getReasoningEffort(agentData.reasoning.effort),
                          maxTokens: agentData.reasoning.maxTokens || null,
                      }
                    : null,
                price:
                    agentData.price &&
                    typeof agentData.price.input === 'number' &&
                    typeof agentData.price.output === 'number'
                        ? {
                              input: agentData.price.input,
                              output: agentData.price.output,
                          }
                        : null,
            };
            agents[agentKey as keyof AgentsConfigInput] = agentConfig;
        }
    });

    return {
        name: formData.name,
        type: formData.type as ProviderType,
        agents,
    };
};

const SettingsProvider = () => {
    const { providerId } = useParams<{ providerId: string }>();
    const navigate = useNavigate();
    const { data, loading, error } = useSettingsProvidersQuery();
    const [createProvider, { loading: createLoading, error: createError }] = useCreateProviderMutation();
    const [updateProvider, { loading: updateLoading, error: updateError }] = useUpdateProviderMutation();
    const [deleteProvider, { loading: deleteLoading, error: deleteError }] = useDeleteProviderMutation();
    const [submitError, setSubmitError] = useState<string | null>(null);

    const isNew = providerId === 'new';
    const isSubmitting = createLoading || updateLoading || deleteLoading;

    const form = useForm<FormData>({
        resolver: zodResolver(formSchema),
        defaultValues: {
            name: '',
            type: '',
            agents: {},
        },
    });

    // Watch all form values and log to console
    // const formValues = form.watch();
    const selectedType = form.watch('type');

    // useEffect(() => {
    //     console.log('Form values changed:', formValues);
    // }, [formValues]);

    // Fill agents when provider type is selected (only for new providers)
    useEffect(() => {
        if (!isNew || !selectedType || !data?.settingsProviders?.default) {
            return;
        }

        const defaultProvider =
            data.settingsProviders.default[selectedType as keyof typeof data.settingsProviders.default];
        if (defaultProvider?.agents) {
            form.setValue('agents', defaultProvider.agents);
        }
    }, [selectedType, isNew, data, form]);

    // Fill form with data when available
    useEffect(() => {
        if (!data?.settingsProviders) return;

        const providers = data.settingsProviders;

        if (isNew || !providerId) {
            // For new provider, start with empty form
            form.reset({
                name: '',
                type: '',
                agents: {},
            });
            return;
        }

        const provider = providers.userDefined?.find((provider: Provider) => provider.id === +providerId);

        if (!provider) {
            navigate('/settings/providers');
            return;
        }

        const { name, type, agents } = provider;

        form.reset({
            name: name || '',
            type: type || '',
            agents: agents || {},
        });
    }, [data, isNew, providerId, form]);

    const onSubmit = async (formData: FormData) => {
        try {
            setSubmitError(null);
            console.log('Form submitted:', formData);

            const mutationData = transformFormToGraphQL(formData);
            console.log('Mutation data:', mutationData);

            if (isNew) {
                // Create new provider
                await createProvider({
                    variables: mutationData,
                    refetchQueries: ['settingsProviders'],
                });
            } else {
                // Update existing provider
                await updateProvider({
                    variables: {
                        ...mutationData,
                        providerId: providerId!,
                    },
                    refetchQueries: ['settingsProviders'],
                });
            }

            // Navigate back to providers list on success
            navigate('/settings/providers');
        } catch (error) {
            console.error('Submit error:', error);
            setSubmitError(error instanceof Error ? error.message : 'An error occurred while saving');
        }
    };

    const onDelete = async () => {
        if (isNew || !providerId) return;

        try {
            setSubmitError(null);
            console.log('Deleting provider:', providerId);

            await deleteProvider({
                variables: { providerId },
                refetchQueries: ['settingsProviders'],
            });

            // Navigate back to providers list on success
            navigate('/settings/providers');
        } catch (error) {
            console.error('Delete error:', error);
            setSubmitError(error instanceof Error ? error.message : 'An error occurred while deleting');
        }
    };

    if (loading) {
        return (
            <StatusCard
                icon={<Loader2 className="w-16 h-16 animate-spin text-muted-foreground" />}
                title="Loading provider data..."
                description="Please wait while we fetch provider configuration"
            />
        );
    }

    if (error) {
        return (
            <Alert variant="destructive">
                <AlertCircle className="h-4 w-4" />
                <AlertTitle>Error loading provider data</AlertTitle>
                <AlertDescription>{error.message}</AlertDescription>
            </Alert>
        );
    }

    const providers = data?.settingsProviders?.models
        ? Object.keys(data?.settingsProviders.models).filter((key) => key !== '__typename')
        : [];

    const mutationError = createError || updateError || deleteError || submitError;

    return (
        <>
            <Card>
                <CardHeader>
                    <CardDescription>
                        {isNew
                            ? 'Configure a new language model provider'
                            : 'Update provider settings and configuration'}
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <Form {...form}>
                        <form
                            id="provider-form"
                            onSubmit={form.handleSubmit(onSubmit)}
                            className="space-y-6"
                        >
                            {/* Error Alert */}
                            {mutationError && (
                                <Alert variant="destructive">
                                    <AlertCircle className="h-4 w-4" />
                                    <AlertTitle>Error</AlertTitle>
                                    <AlertDescription>
                                        {mutationError instanceof Error ? mutationError.message : mutationError}
                                    </AlertDescription>
                                </Alert>
                            )}

                            {/* Form fields */}
                            <FormField
                                control={form.control}
                                name="type"
                                render={({ field }) => (
                                    <FormItem className="flex flex-col gap-1">
                                        <FormLabel>Type</FormLabel>
                                        <Popover>
                                            <PopoverTrigger asChild>
                                                <FormControl>
                                                    <Button
                                                        variant="outline"
                                                        role="combobox"
                                                        className={cn(
                                                            'w-[200px] justify-between',
                                                            !field.value && 'text-muted-foreground',
                                                        )}
                                                        disabled={isSubmitting}
                                                    >
                                                        {field.value
                                                            ? providers.find((provider) => provider === field.value)
                                                            : 'Select provider'}
                                                        <ChevronsUpDown className="opacity-50" />
                                                    </Button>
                                                </FormControl>
                                            </PopoverTrigger>
                                            <PopoverContent className="w-[200px] p-0">
                                                <Command>
                                                    <CommandInput
                                                        placeholder="Search framework..."
                                                        className="h-9"
                                                    />
                                                    <CommandList>
                                                        <CommandEmpty>No framework found.</CommandEmpty>
                                                        <CommandGroup>
                                                            {providers.map((provider) => (
                                                                <CommandItem
                                                                    value={provider}
                                                                    key={provider}
                                                                    onSelect={() => {
                                                                        form.setValue('type', provider);
                                                                    }}
                                                                >
                                                                    {provider}
                                                                    <Check
                                                                        className={cn(
                                                                            'ml-auto',
                                                                            provider === field.value
                                                                                ? 'opacity-100'
                                                                                : 'opacity-0',
                                                                        )}
                                                                    />
                                                                </CommandItem>
                                                            ))}
                                                        </CommandGroup>
                                                    </CommandList>
                                                </Command>
                                            </PopoverContent>
                                        </Popover>
                                        <FormDescription>The type of language model provider</FormDescription>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={form.control}
                                name="name"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>Name</FormLabel>
                                        <FormControl>
                                            <Input
                                                placeholder="Enter provider name"
                                                disabled={isSubmitting}
                                                {...field}
                                            />
                                        </FormControl>
                                        <FormDescription>A unique name for your provider configuration</FormDescription>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            {/* Agents Configuration Section */}
                            <div className="space-y-4">
                                <div>
                                    <h3 className="text-lg font-medium">Agent Configurations</h3>
                                    <p className="text-sm text-muted-foreground">
                                        Configure settings for each agent type
                                    </p>
                                </div>

                                <Accordion
                                    type="multiple"
                                    className="w-full"
                                >
                                    {agentTypes.map((agentKey, index) => (
                                        <AccordionItem
                                            key={agentKey}
                                            value={agentKey}
                                        >
                                            <AccordionTrigger className="text-left">
                                                {getName(agentKey)}
                                            </AccordionTrigger>
                                            <AccordionContent className="space-y-4 pt-4">
                                                <div className="grid grid-cols-1 md:grid-cols-2 gap-4 p-[1px]">
                                                    {/* Model field */}
                                                    <FormField
                                                        control={form.control}
                                                        name={`agents.${agentKey}.model` as const}
                                                        render={({ field }) => (
                                                            <FormItem>
                                                                <FormLabel>Model</FormLabel>
                                                                <FormControl>
                                                                    <Input
                                                                        placeholder="e.g., gpt-4, claude-3"
                                                                        disabled={isSubmitting}
                                                                        {...field}
                                                                    />
                                                                </FormControl>
                                                                <FormMessage />
                                                            </FormItem>
                                                        )}
                                                    />

                                                    {/* Temperature field */}
                                                    <FormField
                                                        control={form.control}
                                                        name={`agents.${agentKey}.temperature` as const}
                                                        render={({ field }) => (
                                                            <FormItem>
                                                                <FormLabel>Temperature</FormLabel>
                                                                <FormControl>
                                                                    <Input
                                                                        type="number"
                                                                        step="0.1"
                                                                        min="0"
                                                                        max="2"
                                                                        placeholder="0.7"
                                                                        disabled={isSubmitting}
                                                                        {...field}
                                                                        value={field.value ?? ''}
                                                                        onChange={(e) => {
                                                                            const value = e.target.value;
                                                                            field.onChange(
                                                                                value === '' ? null : parseFloat(value),
                                                                            );
                                                                        }}
                                                                    />
                                                                </FormControl>
                                                                <FormMessage />
                                                            </FormItem>
                                                        )}
                                                    />

                                                    {/* Max Tokens field */}
                                                    <FormField
                                                        control={form.control}
                                                        name={`agents.${agentKey}.maxTokens` as const}
                                                        render={({ field }) => (
                                                            <FormItem>
                                                                <FormLabel>Max Tokens</FormLabel>
                                                                <FormControl>
                                                                    <Input
                                                                        type="number"
                                                                        min="1"
                                                                        placeholder="1000"
                                                                        disabled={isSubmitting}
                                                                        {...field}
                                                                        value={field.value ?? ''}
                                                                        onChange={(e) => {
                                                                            const value = e.target.value;
                                                                            field.onChange(
                                                                                value === '' ? null : parseInt(value),
                                                                            );
                                                                        }}
                                                                    />
                                                                </FormControl>
                                                                <FormMessage />
                                                            </FormItem>
                                                        )}
                                                    />

                                                    {/* Top P field */}
                                                    <FormField
                                                        control={form.control}
                                                        name={`agents.${agentKey}.topP` as const}
                                                        render={({ field }) => (
                                                            <FormItem>
                                                                <FormLabel>Top P</FormLabel>
                                                                <FormControl>
                                                                    <Input
                                                                        type="number"
                                                                        step="0.01"
                                                                        min="0"
                                                                        max="1"
                                                                        placeholder="0.9"
                                                                        disabled={isSubmitting}
                                                                        {...field}
                                                                        value={field.value ?? ''}
                                                                        onChange={(e) => {
                                                                            const value = e.target.value;
                                                                            field.onChange(
                                                                                value === '' ? null : parseFloat(value),
                                                                            );
                                                                        }}
                                                                    />
                                                                </FormControl>
                                                                <FormMessage />
                                                            </FormItem>
                                                        )}
                                                    />

                                                    {/* Top K field */}
                                                    <FormField
                                                        control={form.control}
                                                        name={`agents.${agentKey}.topK` as const}
                                                        render={({ field }) => (
                                                            <FormItem>
                                                                <FormLabel>Top K</FormLabel>
                                                                <FormControl>
                                                                    <Input
                                                                        type="number"
                                                                        min="1"
                                                                        placeholder="40"
                                                                        disabled={isSubmitting}
                                                                        {...field}
                                                                        value={field.value ?? ''}
                                                                        onChange={(e) => {
                                                                            const value = e.target.value;
                                                                            field.onChange(
                                                                                value === '' ? null : parseInt(value),
                                                                            );
                                                                        }}
                                                                    />
                                                                </FormControl>
                                                                <FormMessage />
                                                            </FormItem>
                                                        )}
                                                    />

                                                    {/* Min Length field */}
                                                    <FormField
                                                        control={form.control}
                                                        name={`agents.${agentKey}.minLength` as const}
                                                        render={({ field }) => (
                                                            <FormItem>
                                                                <FormLabel>Min Length</FormLabel>
                                                                <FormControl>
                                                                    <Input
                                                                        type="number"
                                                                        min="0"
                                                                        placeholder="0"
                                                                        disabled={isSubmitting}
                                                                        {...field}
                                                                        value={field.value ?? ''}
                                                                        onChange={(e) => {
                                                                            const value = e.target.value;
                                                                            field.onChange(
                                                                                value === '' ? null : parseInt(value),
                                                                            );
                                                                        }}
                                                                    />
                                                                </FormControl>
                                                                <FormMessage />
                                                            </FormItem>
                                                        )}
                                                    />

                                                    {/* Max Length field */}
                                                    <FormField
                                                        control={form.control}
                                                        name={`agents.${agentKey}.maxLength` as const}
                                                        render={({ field }) => (
                                                            <FormItem>
                                                                <FormLabel>Max Length</FormLabel>
                                                                <FormControl>
                                                                    <Input
                                                                        type="number"
                                                                        min="1"
                                                                        placeholder="2000"
                                                                        disabled={isSubmitting}
                                                                        {...field}
                                                                        value={field.value ?? ''}
                                                                        onChange={(e) => {
                                                                            const value = e.target.value;
                                                                            field.onChange(
                                                                                value === '' ? null : parseInt(value),
                                                                            );
                                                                        }}
                                                                    />
                                                                </FormControl>
                                                                <FormMessage />
                                                            </FormItem>
                                                        )}
                                                    />

                                                    {/* Repetition Penalty field */}
                                                    <FormField
                                                        control={form.control}
                                                        name={`agents.${agentKey}.repetitionPenalty` as const}
                                                        render={({ field }) => (
                                                            <FormItem>
                                                                <FormLabel>Repetition Penalty</FormLabel>
                                                                <FormControl>
                                                                    <Input
                                                                        type="number"
                                                                        step="0.01"
                                                                        min="0"
                                                                        max="2"
                                                                        placeholder="1.0"
                                                                        disabled={isSubmitting}
                                                                        {...field}
                                                                        value={field.value ?? ''}
                                                                        onChange={(e) => {
                                                                            const value = e.target.value;
                                                                            field.onChange(
                                                                                value === '' ? null : parseFloat(value),
                                                                            );
                                                                        }}
                                                                    />
                                                                </FormControl>
                                                                <FormMessage />
                                                            </FormItem>
                                                        )}
                                                    />

                                                    {/* Frequency Penalty field */}
                                                    <FormField
                                                        control={form.control}
                                                        name={`agents.${agentKey}.frequencyPenalty` as const}
                                                        render={({ field }) => (
                                                            <FormItem>
                                                                <FormLabel>Frequency Penalty</FormLabel>
                                                                <FormControl>
                                                                    <Input
                                                                        type="number"
                                                                        step="0.01"
                                                                        min="0"
                                                                        max="2"
                                                                        placeholder="0.0"
                                                                        disabled={isSubmitting}
                                                                        {...field}
                                                                        value={field.value ?? ''}
                                                                        onChange={(e) => {
                                                                            const value = e.target.value;
                                                                            field.onChange(
                                                                                value === '' ? null : parseFloat(value),
                                                                            );
                                                                        }}
                                                                    />
                                                                </FormControl>
                                                                <FormMessage />
                                                            </FormItem>
                                                        )}
                                                    />

                                                    {/* Presence Penalty field */}
                                                    <FormField
                                                        control={form.control}
                                                        name={`agents.${agentKey}.presencePenalty` as const}
                                                        render={({ field }) => (
                                                            <FormItem>
                                                                <FormLabel>Presence Penalty</FormLabel>
                                                                <FormControl>
                                                                    <Input
                                                                        type="number"
                                                                        step="0.01"
                                                                        min="0"
                                                                        max="2"
                                                                        placeholder="0.0"
                                                                        disabled={isSubmitting}
                                                                        {...field}
                                                                        value={field.value ?? ''}
                                                                        onChange={(e) => {
                                                                            const value = e.target.value;
                                                                            field.onChange(
                                                                                value === '' ? null : parseFloat(value),
                                                                            );
                                                                        }}
                                                                    />
                                                                </FormControl>
                                                                <FormMessage />
                                                            </FormItem>
                                                        )}
                                                    />
                                                </div>

                                                {/* Reasoning Configuration */}
                                                <div className="mt-6 p-[1px] space-y-4">
                                                    <h4 className="text-sm font-medium">Reasoning Configuration</h4>
                                                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                                        {/* Reasoning Effort field */}
                                                        <FormField
                                                            control={form.control}
                                                            name={`agents.${agentKey}.reasoning.effort` as const}
                                                            render={({ field }) => (
                                                                <FormItem>
                                                                    <FormLabel>Reasoning Effort</FormLabel>
                                                                    <FormControl>
                                                                        <Input
                                                                            placeholder="e.g., low, medium, high"
                                                                            disabled={isSubmitting}
                                                                            {...field}
                                                                            value={field.value ?? ''}
                                                                            onChange={(e) => {
                                                                                const value = e.target.value;
                                                                                field.onChange(
                                                                                    value === '' ? null : value,
                                                                                );
                                                                            }}
                                                                        />
                                                                    </FormControl>
                                                                    <FormMessage />
                                                                </FormItem>
                                                            )}
                                                        />

                                                        {/* Reasoning Max Tokens field */}
                                                        <FormField
                                                            control={form.control}
                                                            name={`agents.${agentKey}.reasoning.maxTokens` as const}
                                                            render={({ field }) => (
                                                                <FormItem>
                                                                    <FormLabel>Reasoning Max Tokens</FormLabel>
                                                                    <FormControl>
                                                                        <Input
                                                                            type="number"
                                                                            min="1"
                                                                            placeholder="1000"
                                                                            disabled={isSubmitting}
                                                                            {...field}
                                                                            value={field.value ?? ''}
                                                                            onChange={(e) => {
                                                                                const value = e.target.value;
                                                                                field.onChange(
                                                                                    value === ''
                                                                                        ? null
                                                                                        : parseInt(value),
                                                                                );
                                                                            }}
                                                                        />
                                                                    </FormControl>
                                                                    <FormMessage />
                                                                </FormItem>
                                                            )}
                                                        />
                                                    </div>
                                                </div>

                                                {/* Price Configuration */}
                                                <div className="mt-6 p-[1px] space-y-4">
                                                    <h4 className="text-sm font-medium">Price Configuration</h4>
                                                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                                        {/* Price Input field */}
                                                        <FormField
                                                            control={form.control}
                                                            name={`agents.${agentKey}.price.input` as const}
                                                            render={({ field }) => (
                                                                <FormItem>
                                                                    <FormLabel>Input Price</FormLabel>
                                                                    <FormControl>
                                                                        <Input
                                                                            type="number"
                                                                            step="0.000001"
                                                                            min="0"
                                                                            placeholder="0.001"
                                                                            disabled={isSubmitting}
                                                                            {...field}
                                                                            value={field.value ?? ''}
                                                                            onChange={(e) => {
                                                                                const value = e.target.value;
                                                                                field.onChange(
                                                                                    value === ''
                                                                                        ? null
                                                                                        : parseFloat(value),
                                                                                );
                                                                            }}
                                                                        />
                                                                    </FormControl>
                                                                    <FormMessage />
                                                                </FormItem>
                                                            )}
                                                        />

                                                        {/* Price Output field */}
                                                        <FormField
                                                            control={form.control}
                                                            name={`agents.${agentKey}.price.output` as const}
                                                            render={({ field }) => (
                                                                <FormItem>
                                                                    <FormLabel>Output Price</FormLabel>
                                                                    <FormControl>
                                                                        <Input
                                                                            type="number"
                                                                            step="0.000001"
                                                                            min="0"
                                                                            placeholder="0.002"
                                                                            disabled={isSubmitting}
                                                                            {...field}
                                                                            value={field.value ?? ''}
                                                                            onChange={(e) => {
                                                                                const value = e.target.value;
                                                                                field.onChange(
                                                                                    value === ''
                                                                                        ? null
                                                                                        : parseFloat(value),
                                                                                );
                                                                            }}
                                                                        />
                                                                    </FormControl>
                                                                    <FormMessage />
                                                                </FormItem>
                                                            )}
                                                        />
                                                    </div>
                                                </div>
                                            </AccordionContent>
                                        </AccordionItem>
                                    ))}
                                </Accordion>
                            </div>
                        </form>
                    </Form>
                </CardContent>
            </Card>

            {/* Sticky buttons at bottom */}
            <div className="flex items-center sticky -bottom-4 bg-background border-t mt-4 -mx-4 -mb-4 p-4 shadow-lg">
                {/* Delete button - only show when editing existing provider */}
                {!isNew && (
                    <Button
                        type="button"
                        variant="destructive"
                        onClick={onDelete}
                        disabled={isSubmitting}
                    >
                        {deleteLoading ? <Loader2 className="h-4 w-4 animate-spin" /> : <Trash2 className="h-4 w-4" />}
                        {deleteLoading ? 'Deleting...' : 'Delete'}
                    </Button>
                )}

                {/* Right side buttons */}
                <div className="flex space-x-2 ml-auto">
                    <Button
                        type="button"
                        variant="outline"
                        onClick={() => navigate('/settings/providers')}
                        disabled={isSubmitting}
                    >
                        Cancel
                    </Button>
                    <Button
                        form="provider-form"
                        variant="secondary"
                        type="submit"
                        disabled={isSubmitting}
                    >
                        {isSubmitting ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
                        {isSubmitting ? 'Saving...' : isNew ? 'Create Provider' : 'Update Provider'}
                    </Button>
                </div>
            </div>
        </>
    );
};

export default SettingsProvider;
