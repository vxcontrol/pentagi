import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from '@/components/ui/accordion';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader } from '@/components/ui/card';
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from '@/components/ui/command';
import { Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
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
import { useEffect, useMemo, useState } from 'react';
import { useController, useForm } from 'react-hook-form';
import { useNavigate, useParams } from 'react-router-dom';
import { z } from 'zod';

type Provider = ProviderConfigFragmentFragment;

// Universal field components using useController
interface ControllerProps {
    name: string;
    control: any;
    disabled?: boolean;
}

interface BaseInputProps {
    placeholder?: string;
}

interface NumberInputProps extends BaseInputProps {
    step?: string;
    min?: string;
    max?: string;
}

interface BaseFieldProps extends ControllerProps {
    label: string;
}

interface FormInputStringItemProps extends BaseFieldProps, BaseInputProps {
    description?: string;
}

interface FormInputNumberItemProps extends BaseFieldProps, NumberInputProps {
    valueType?: 'float' | 'integer';
    description?: string;
}

const FormInputStringItem: React.FC<FormInputStringItemProps> = ({
    name,
    control,
    disabled,
    label,
    placeholder,
    description,
}) => {
    const { field, fieldState } = useController({
        name,
        control,
        defaultValue: undefined,
        disabled,
    });

    const inputProps = { placeholder };

    return (
        <FormItem>
            <FormLabel>{label}</FormLabel>
            <FormControl>
                <Input
                    {...field}
                    {...inputProps}
                    value={field.value ?? ''}
                />
            </FormControl>
            {description && <FormDescription>{description}</FormDescription>}
            {fieldState.error && <FormMessage>{fieldState.error.message}</FormMessage>}
        </FormItem>
    );
};

const FormInputNumberItem: React.FC<FormInputNumberItemProps> = ({
    name,
    control,
    disabled,
    label,
    placeholder,
    step,
    min,
    max,
    valueType = 'float',
    description,
}) => {
    const { field, fieldState } = useController({
        name,
        control,
        defaultValue: undefined,
        disabled,
    });

    const parseValue = (value: string) => {
        if (value === '') {
            return undefined;
        }

        return valueType === 'float' ? parseFloat(value) : parseInt(value);
    };

    const inputProps = {
        type: 'number' as const,
        step,
        min,
        max,
        placeholder,
    };

    return (
        <FormItem>
            <FormLabel>{label}</FormLabel>
            <FormControl>
                <Input
                    {...field}
                    {...inputProps}
                    value={field.value ?? ''}
                    onChange={(event) => {
                        const value = event.target.value;
                        field.onChange(parseValue(value));
                    }}
                />
            </FormControl>
            {description && <FormDescription>{description}</FormDescription>}
            {fieldState.error && <FormMessage>{fieldState.error.message}</FormMessage>}
        </FormItem>
    );
};

interface FormComboboxItemProps extends BaseFieldProps, BaseInputProps {
    options: string[] | { provider: string; models: string[] }[];
    allowCustom?: boolean;
    width?: string;
    description?: string;
}

const FormComboboxItem: React.FC<FormComboboxItemProps> = ({
    name,
    control,
    disabled,
    label,
    placeholder,
    options,
    allowCustom = true,
    width = 'w-[400px]',
    description,
}) => {
    const { field, fieldState } = useController({
        name,
        control,
        defaultValue: undefined,
        disabled,
    });

    const [open, setOpen] = useState(false);
    const [search, setSearch] = useState('');

    // Handle both flat array and grouped data
    const isGroupedData = options.length > 0 && typeof options[0] === 'object';

    const filteredOptions = isGroupedData
        ? (options as { provider: string; models: string[] }[])
              .map((group) => ({
                  provider: group.provider,
                  models: group.models.filter((model) => model.toLowerCase().includes(search.toLowerCase())),
              }))
              .filter((group) => group.models.length > 0)
        : (options as string[]).filter((option) => option.toLowerCase().includes(search.toLowerCase()));

    const displayValue = field.value ?? '';

    return (
        <FormItem>
            <FormLabel>{label}</FormLabel>
            <FormControl>
                <Popover
                    open={open}
                    onOpenChange={setOpen}
                >
                    <PopoverTrigger asChild>
                        <Button
                            variant="outline"
                            role="combobox"
                            aria-expanded={open}
                            className={cn('w-full justify-between', !displayValue && 'text-muted-foreground')}
                            disabled={disabled}
                        >
                            {displayValue || placeholder}
                            <ChevronsUpDown className="opacity-50" />
                        </Button>
                    </PopoverTrigger>
                    <PopoverContent
                        className={cn(width, 'p-0')}
                        align="start"
                    >
                        <Command>
                            <CommandInput
                                placeholder={`Search ${label.toLowerCase()}...`}
                                className="h-9"
                                value={search}
                                onValueChange={setSearch}
                            />
                            <CommandList>
                                <CommandEmpty>
                                    <div className="text-center py-2">
                                        <p className="text-sm text-muted-foreground">No {label.toLowerCase()} found.</p>
                                        {search && allowCustom && (
                                            <Button
                                                variant="ghost"
                                                size="sm"
                                                className="mt-2"
                                                onClick={() => {
                                                    field.onChange(search);
                                                    setOpen(false);
                                                    setSearch('');
                                                }}
                                            >
                                                Use "{search}" as custom {label.toLowerCase()}
                                            </Button>
                                        )}
                                    </div>
                                </CommandEmpty>
                                {isGroupedData ? (
                                    // Render grouped data
                                    (filteredOptions as { provider: string; models: string[] }[]).map((group) => (
                                        <CommandGroup
                                            key={group.provider}
                                            heading={group.provider}
                                        >
                                            {group.models.map((model) => (
                                                <CommandItem
                                                    key={model}
                                                    value={model}
                                                    onSelect={() => {
                                                        field.onChange(model);
                                                        setOpen(false);
                                                        setSearch('');
                                                    }}
                                                >
                                                    {model}
                                                    <Check
                                                        className={cn(
                                                            'ml-auto',
                                                            displayValue === model ? 'opacity-100' : 'opacity-0',
                                                        )}
                                                    />
                                                </CommandItem>
                                            ))}
                                        </CommandGroup>
                                    ))
                                ) : (
                                    // Render flat data
                                    <CommandGroup>
                                        {(filteredOptions as string[]).map((option) => (
                                            <CommandItem
                                                key={option}
                                                value={option}
                                                onSelect={() => {
                                                    field.onChange(option);
                                                    setOpen(false);
                                                    setSearch('');
                                                }}
                                            >
                                                {option}
                                                <Check
                                                    className={cn(
                                                        'ml-auto',
                                                        displayValue === option ? 'opacity-100' : 'opacity-0',
                                                    )}
                                                />
                                            </CommandItem>
                                        ))}
                                    </CommandGroup>
                                )}
                            </CommandList>
                        </Command>
                    </PopoverContent>
                </Popover>
            </FormControl>
            {description && <FormDescription>{description}</FormDescription>}
            {fieldState.error && <FormMessage>{fieldState.error.message}</FormMessage>}
        </FormItem>
    );
};

// Define agent configuration schema
const agentConfigSchema = z
    .object({
        model: z.preprocess((value) => value || '', z.string().min(1, 'Model is required')),
        temperature: z.preprocess(
            (value) => (value === '' || value === undefined ? null : value),
            z.number().nullable().optional(),
        ),
        maxTokens: z.preprocess(
            (value) => (value === '' || value === undefined ? null : value),
            z.number().nullable().optional(),
        ),
        topK: z.preprocess(
            (value) => (value === '' || value === undefined ? null : value),
            z.number().nullable().optional(),
        ),
        topP: z.preprocess(
            (value) => (value === '' || value === undefined ? null : value),
            z.number().nullable().optional(),
        ),
        minLength: z.preprocess(
            (value) => (value === '' || value === undefined ? null : value),
            z.number().nullable().optional(),
        ),
        maxLength: z.preprocess(
            (value) => (value === '' || value === undefined ? null : value),
            z.number().nullable().optional(),
        ),
        repetitionPenalty: z.preprocess(
            (value) => (value === '' || value === undefined ? null : value),
            z.number().nullable().optional(),
        ),
        frequencyPenalty: z.preprocess(
            (value) => (value === '' || value === undefined ? null : value),
            z.number().nullable().optional(),
        ),
        presencePenalty: z.preprocess(
            (value) => (value === '' || value === undefined ? null : value),
            z.number().nullable().optional(),
        ),
        reasoning: z
            .object({
                effort: z.preprocess(
                    (value) => (value === '' || value === undefined ? null : value),
                    z.string().nullable().optional(),
                ),
                maxTokens: z.preprocess(
                    (value) => (value === '' || value === undefined ? null : value),
                    z.number().nullable().optional(),
                ),
            })
            .nullable()
            .optional(),
        price: z
            .object({
                input: z.preprocess(
                    (value) => (value === '' || value === undefined ? null : value),
                    z.number().nullable().optional(),
                ),
                output: z.preprocess(
                    (value) => (value === '' || value === undefined ? null : value),
                    z.number().nullable().optional(),
                ),
            })
            .nullable()
            .optional(),
    })
    .optional();

// Define form schema
const formSchema = z.object({
    type: z.preprocess((value) => value || '', z.string().min(1, 'Provider type is required')),
    name: z.preprocess((value) => value || '', z.string().min(1, 'Provider name is required')),
    agents: z.record(z.string(), agentConfigSchema).optional(),
});

type FormData = z.infer<typeof formSchema>;

// Convert camelCase key to display name (e.g., 'simpleJson' -> 'Simple Json')
const getName = (key: string): string => key.replace(/([A-Z])/g, ' $1').replace(/^./, (item) => item.toUpperCase());

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
    const agents = Object.entries(formData.agents || {})
        .filter(([, data]) => data?.model)
        .reduce((configs, [key, data]) => {
            const config: AgentConfigInput = {
                model: data!.model, // After filter, data and model are guaranteed to exist
                temperature: data?.temperature ?? null,
                maxTokens: data?.maxTokens ?? null,
                topK: data?.topK ?? null,
                topP: data?.topP ?? null,
                minLength: data?.minLength ?? null,
                maxLength: data?.maxLength ?? null,
                repetitionPenalty: data?.repetitionPenalty ?? null,
                frequencyPenalty: data?.frequencyPenalty ?? null,
                presencePenalty: data?.presencePenalty ?? null,
                reasoning: data?.reasoning
                    ? {
                          effort: getReasoningEffort(data?.reasoning.effort),
                          maxTokens: data?.reasoning.maxTokens ?? null,
                      }
                    : null,
                price:
                    data?.price &&
                    data?.price.input !== null &&
                    data?.price.output !== null &&
                    typeof data?.price.input === 'number' &&
                    typeof data?.price.output === 'number'
                        ? {
                              input: data?.price.input,
                              output: data?.price.output,
                          }
                        : null,
            };

            return { ...configs, [key]: config };
        }, {} as AgentsConfigInput);

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
            name: undefined,
            type: undefined,
            agents: {},
        },
    });

    // Watch selected type
    const selectedType = form.watch('type');

    // Get dynamic agent types from data
    const agentTypes = useMemo(() => {
        let agentsSource = null;

        if (isNew && selectedType && data?.settingsProviders?.default) {
            // For new providers, use default provider for selected type
            agentsSource =
                data.settingsProviders.default?.[selectedType as keyof typeof data.settingsProviders.default]?.agents;
        } else if (!isNew && providerId && data?.settingsProviders?.userDefined) {
            // For existing providers, use current provider's agents
            const provider = data.settingsProviders.userDefined.find((p: Provider) => p.id === +providerId);
            agentsSource = provider?.agents;
        }

        // If no specific source, try to get from any available default provider
        if (!agentsSource && data?.settingsProviders?.default) {
            const defaultProviders = data.settingsProviders.default;
            // Try to find first available default provider with agents
            const firstDefaultProvider = Object.values(defaultProviders).find((provider) => provider?.agents);
            agentsSource = firstDefaultProvider?.agents;
        }

        // Extract agent types from the source
        if (agentsSource) {
            return Object.entries(agentsSource)
                .filter(([key, data]) => key !== '__typename' && data)
                .map(([key]) => key)
                .sort();
        }

        // Fallback to hardcoded list if no data available
        return [
            'adviser',
            'agent',
            'assistant',
            'coder',
            'enricher',
            'generator',
            'installer',
            'pentester',
            'reflector',
            'refiner',
            'searcher',
            'simple',
            'simpleJson',
        ];
    }, [isNew, selectedType, providerId, data]);

    // Get all available models grouped by provider
    const availableModels = useMemo(() => {
        if (!data?.settingsProviders?.models) {
            return [];
        }

        return Object.entries(data.settingsProviders.models)
            .filter(([provider, models]) => provider !== '__typename' && models?.length)
            .map(([provider, models]) => ({
                provider: getName(provider),
                models: models?.map((model) => model.name)?.sort() ?? [],
            }))
            .sort((a, b) => a.provider.localeCompare(b.provider));
    }, [data]);

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
                name: undefined,
                type: undefined,
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
            name: name || undefined,
            type: type || undefined,
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
                            <FormComboboxItem
                                name="type"
                                label="Type"
                                placeholder="Select provider"
                                options={providers}
                                control={form.control}
                                disabled={isSubmitting}
                                allowCustom={false}
                                width="w-[200px]"
                                description="The type of language model provider"
                            />

                            <FormInputStringItem
                                name="name"
                                label="Name"
                                placeholder="Enter provider name"
                                control={form.control}
                                disabled={isSubmitting}
                                description="A unique name for your provider configuration"
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
                                                    <FormComboboxItem
                                                        name={`agents.${agentKey}.model`}
                                                        label="Model"
                                                        placeholder="Select or enter model name"
                                                        options={availableModels}
                                                        control={form.control}
                                                        disabled={isSubmitting}
                                                    />

                                                    {/* Temperature field */}
                                                    <FormInputNumberItem
                                                        name={`agents.${agentKey}.temperature`}
                                                        label="Temperature"
                                                        placeholder="0.7"
                                                        control={form.control}
                                                        disabled={isSubmitting}
                                                        step="0.1"
                                                        min="0"
                                                        max="2"
                                                    />

                                                    {/* Max Tokens field */}
                                                    <FormInputNumberItem
                                                        name={`agents.${agentKey}.maxTokens`}
                                                        label="Max Tokens"
                                                        placeholder="1000"
                                                        control={form.control}
                                                        disabled={isSubmitting}
                                                        min="1"
                                                        valueType="integer"
                                                    />

                                                    {/* Top P field */}
                                                    <FormInputNumberItem
                                                        name={`agents.${agentKey}.topP`}
                                                        label="Top P"
                                                        placeholder="0.9"
                                                        control={form.control}
                                                        disabled={isSubmitting}
                                                        step="0.01"
                                                        min="0"
                                                        max="1"
                                                    />

                                                    {/* Top K field */}
                                                    <FormInputNumberItem
                                                        name={`agents.${agentKey}.topK`}
                                                        label="Top K"
                                                        placeholder="40"
                                                        control={form.control}
                                                        disabled={isSubmitting}
                                                        min="1"
                                                        valueType="integer"
                                                    />

                                                    {/* Min Length field */}
                                                    <FormInputNumberItem
                                                        name={`agents.${agentKey}.minLength`}
                                                        label="Min Length"
                                                        placeholder="0"
                                                        control={form.control}
                                                        disabled={isSubmitting}
                                                        min="0"
                                                        valueType="integer"
                                                    />

                                                    {/* Max Length field */}
                                                    <FormInputNumberItem
                                                        name={`agents.${agentKey}.maxLength`}
                                                        label="Max Length"
                                                        placeholder="2000"
                                                        control={form.control}
                                                        disabled={isSubmitting}
                                                        min="1"
                                                        valueType="integer"
                                                    />

                                                    {/* Repetition Penalty field */}
                                                    <FormInputNumberItem
                                                        name={`agents.${agentKey}.repetitionPenalty`}
                                                        label="Repetition Penalty"
                                                        placeholder="1.0"
                                                        control={form.control}
                                                        disabled={isSubmitting}
                                                        step="0.01"
                                                        min="0"
                                                        max="2"
                                                    />

                                                    {/* Frequency Penalty field */}
                                                    <FormInputNumberItem
                                                        name={`agents.${agentKey}.frequencyPenalty`}
                                                        label="Frequency Penalty"
                                                        placeholder="0.0"
                                                        control={form.control}
                                                        disabled={isSubmitting}
                                                        step="0.01"
                                                        min="0"
                                                        max="2"
                                                    />

                                                    {/* Presence Penalty field */}
                                                    <FormInputNumberItem
                                                        name={`agents.${agentKey}.presencePenalty`}
                                                        label="Presence Penalty"
                                                        placeholder="0.0"
                                                        control={form.control}
                                                        disabled={isSubmitting}
                                                        step="0.01"
                                                        min="0"
                                                        max="2"
                                                    />
                                                </div>

                                                {/* Reasoning Configuration */}
                                                <div className="col-span-full">
                                                    <div className="mt-6 space-y-4">
                                                        <h4 className="text-sm font-medium">Reasoning Configuration</h4>
                                                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                                            {/* Reasoning Effort field */}
                                                            <FormField
                                                                control={form.control}
                                                                name={`agents.${agentKey}.reasoning.effort`}
                                                                render={({ field }) => (
                                                                    <FormItem>
                                                                        <FormLabel>Reasoning Effort</FormLabel>
                                                                        <Select
                                                                            onValueChange={field.onChange}
                                                                            defaultValue={field.value || undefined}
                                                                            disabled={isSubmitting}
                                                                        >
                                                                            <FormControl>
                                                                                <SelectTrigger>
                                                                                    <SelectValue placeholder="Select effort level (optional)" />
                                                                                </SelectTrigger>
                                                                            </FormControl>
                                                                            <SelectContent>
                                                                                <SelectItem value={ReasoningEffort.Low}>
                                                                                    Low
                                                                                </SelectItem>
                                                                                <SelectItem
                                                                                    value={ReasoningEffort.Medium}
                                                                                >
                                                                                    Medium
                                                                                </SelectItem>
                                                                                <SelectItem
                                                                                    value={ReasoningEffort.High}
                                                                                >
                                                                                    High
                                                                                </SelectItem>
                                                                            </SelectContent>
                                                                        </Select>
                                                                        <FormMessage />
                                                                    </FormItem>
                                                                )}
                                                            />

                                                            {/* Reasoning Max Tokens field */}
                                                            <FormInputNumberItem
                                                                name={`agents.${agentKey}.reasoning.maxTokens`}
                                                                label="Reasoning Max Tokens"
                                                                placeholder="1000"
                                                                control={form.control}
                                                                disabled={isSubmitting}
                                                                min="1"
                                                                valueType="integer"
                                                            />
                                                        </div>
                                                    </div>
                                                </div>

                                                {/* Price Configuration */}
                                                <div className="col-span-full">
                                                    <div className="mt-6 space-y-4">
                                                        <h4 className="text-sm font-medium">Price Configuration</h4>
                                                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                                            {/* Price Input field */}
                                                            <FormInputNumberItem
                                                                name={`agents.${agentKey}.price.input`}
                                                                label="Input Price"
                                                                placeholder="0.001"
                                                                control={form.control}
                                                                disabled={isSubmitting}
                                                                step="0.000001"
                                                                min="0"
                                                            />

                                                            {/* Price Output field */}
                                                            <FormInputNumberItem
                                                                name={`agents.${agentKey}.price.output`}
                                                                label="Output Price"
                                                                placeholder="0.002"
                                                                control={form.control}
                                                                disabled={isSubmitting}
                                                                step="0.000001"
                                                                min="0"
                                                            />
                                                        </div>
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
