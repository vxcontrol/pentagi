import { zodResolver } from '@hookform/resolvers/zod';
import {
    AlertCircle,
    Check,
    CheckCircle,
    ChevronsUpDown,
    Clock,
    Lightbulb,
    Loader2,
    Play,
    Save,
    Trash2,
    XCircle,
} from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import { useController, useForm, useFormState, useWatch } from 'react-hook-form';
import { useNavigate, useParams, useSearchParams } from 'react-router-dom';
import { z } from 'zod';

import type {
    AgentConfigInput,
    AgentsConfigInput,
    ProviderConfigFragmentFragment,
    ProviderType,
} from '@/graphql/types';

import ConfirmationDialog from '@/components/shared/confirmation-dialog';
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from '@/components/ui/accordion';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader } from '@/components/ui/card';
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from '@/components/ui/command';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { StatusCard } from '@/components/ui/status-card';
import {
    AgentConfigType,
    ReasoningEffort,
    useCreateProviderMutation,
    useDeleteProviderMutation,
    useSettingsProvidersQuery,
    useTestAgentMutation,
    useTestProviderMutation,
    useUpdateProviderMutation,
} from '@/graphql/types';
import { cn } from '@/lib/utils';

interface BaseFieldProps extends ControllerProps {
    label: string;
}

interface BaseInputProps {
    placeholder?: string;
}

// Universal field components using useController
interface ControllerProps {
    control: any;
    disabled?: boolean;
    name: string;
}

interface FormInputNumberItemProps extends BaseFieldProps, NumberInputProps {
    description?: string;
    valueType?: 'float' | 'integer';
}

interface FormInputStringItemProps extends BaseFieldProps, BaseInputProps {
    description?: string;
}

interface NumberInputProps extends BaseInputProps {
    max?: string;
    min?: string;
    step?: string;
}

type Provider = ProviderConfigFragmentFragment;

const FormInputStringItem: React.FC<FormInputStringItemProps> = ({
    control,
    description,
    disabled,
    label,
    name,
    placeholder,
}) => {
    const { field, fieldState } = useController({
        control,
        defaultValue: undefined,
        disabled,
        name,
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
    control,
    description,
    disabled,
    label,
    max,
    min,
    name,
    placeholder,
    step,
    valueType = 'float',
}) => {
    const { field, fieldState } = useController({
        control,
        defaultValue: undefined,
        disabled,
        name,
    });

    const parseValue = (value: string) => {
        if (value === '') {
            return null;
        }

        return valueType === 'float' ? Number.parseFloat(value) : Number.parseInt(value);
    };

    const inputProps = {
        max,
        min,
        placeholder,
        step,
        type: 'number' as const,
    };

    return (
        <FormItem>
            <FormLabel>{label}</FormLabel>
            <FormControl>
                <Input
                    {...field}
                    {...inputProps}
                    onChange={(event) => {
                        const { value } = event.target;
                        field.onChange(parseValue(value));
                    }}
                    value={field.value ?? ''}
                />
            </FormControl>
            {description && <FormDescription>{description}</FormDescription>}
            {fieldState.error && <FormMessage>{fieldState.error.message}</FormMessage>}
        </FormItem>
    );
};

interface FormComboboxItemProps extends BaseFieldProps, BaseInputProps {
    allowCustom?: boolean;
    contentClass?: string;
    description?: string;
    options: string[];
}

const FormComboboxItem: React.FC<FormComboboxItemProps> = ({
    allowCustom = true,
    contentClass,
    control,
    description,
    disabled,
    label,
    name,
    options,
    placeholder,
}) => {
    const { field, fieldState } = useController({
        control,
        defaultValue: undefined,
        disabled,
        name,
    });

    const [isOpen, setIsOpen] = useState(false);
    const [search, setSearch] = useState('');

    // Filter options based on search
    const filteredOptions = options.filter((option) => option?.toLowerCase().includes(search?.toLowerCase()));

    const displayValue = field.value ?? '';

    return (
        <FormItem>
            <FormLabel>{label}</FormLabel>
            <FormControl>
                <Popover
                    onOpenChange={setIsOpen}
                    open={isOpen}
                >
                    <PopoverTrigger asChild>
                        <Button
                            className={cn('w-full justify-between', !displayValue && 'text-muted-foreground')}
                            disabled={disabled}
                            variant="outline"
                        >
                            {displayValue || placeholder}
                            <ChevronsUpDown className="opacity-50" />
                        </Button>
                    </PopoverTrigger>
                    <PopoverContent
                        align="start"
                        className={cn(contentClass, 'p-0')}
                        style={{
                            maxHeight: 'var(--radix-popover-content-available-height)',
                            width: 'var(--radix-popover-trigger-width)',
                        }}
                    >
                        <Command>
                            <CommandInput
                                className="h-9"
                                onValueChange={setSearch}
                                placeholder={`Search ${label.toLowerCase()}...`}
                                value={search}
                            />
                            <CommandList>
                                <CommandEmpty>
                                    <div className="py-2 text-center">
                                        <p className="text-sm text-muted-foreground">No {label.toLowerCase()} found.</p>
                                        {search && allowCustom && (
                                            <Button
                                                className="mt-2"
                                                onClick={() => {
                                                    field.onChange(search);
                                                    setIsOpen(false);
                                                    setSearch('');
                                                }}
                                                size="sm"
                                                variant="ghost"
                                            >
                                                Use "{search}" as custom {label.toLowerCase()}
                                            </Button>
                                        )}
                                    </div>
                                </CommandEmpty>
                                <CommandGroup>
                                    {filteredOptions.map((option) => (
                                        <CommandItem
                                            key={option}
                                            onSelect={() => {
                                                field.onChange(option);
                                                setIsOpen(false);
                                                setSearch('');
                                            }}
                                            value={option}
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

interface FormModelComboboxItemProps extends BaseFieldProps, BaseInputProps {
    allowCustom?: boolean;
    contentClass?: string;
    description?: string;
    onOptionSelect?: (option: ModelOption) => void;
    options: ModelOption[];
}

interface ModelOption {
    name: string;
    price?: null | { input: number; output: number };
    thinking?: boolean;
}

const FormModelComboboxItem: React.FC<FormModelComboboxItemProps> = ({
    allowCustom = true,
    contentClass,
    control,
    description,
    disabled,
    label,
    name,
    onOptionSelect,
    options,
    placeholder,
}) => {
    const { field, fieldState } = useController({
        control,
        defaultValue: undefined,
        disabled,
        name,
    });

    const [isOpen, setIsOpen] = useState(false);
    const [search, setSearch] = useState('');

    // Filter options based on search
    const filteredOptions = options.filter((option) => option.name?.toLowerCase().includes(search?.toLowerCase()));

    const displayValue = field.value ?? '';

    // Format price for display
    const formatPrice = (price?: null | { input: number; output: number }): string => {
        if (!price || ((!price.input || price.input === 0) && (!price.output || price.output === 0))) {
            return 'free';
        }

        const formatValue = (value: number): string => {
            return value.toFixed(6).replace(/\.?0+$/, '');
        };

        return `$${formatValue(price.input)}/$${formatValue(price.output)}`;
    };

    return (
        <FormItem>
            <FormLabel>{label}</FormLabel>
            <FormControl>
                <Popover
                    onOpenChange={setIsOpen}
                    open={isOpen}
                >
                    <div className="flex w-full">
                        {/* Input field - main control */}
                        <Input
                            className="rounded-r-none border-r-0 focus-visible:z-10"
                            disabled={disabled}
                            onChange={(event) => field.onChange(event.target.value)}
                            placeholder={placeholder}
                            value={displayValue}
                        />
                        {/* Dropdown trigger button */}
                        <PopoverTrigger asChild>
                            <Button
                                className="rounded-l-none border-l-0 px-3 hover:z-10"
                                disabled={disabled}
                                type="button"
                                variant="outline"
                            >
                                <ChevronsUpDown className="size-4 opacity-50" />
                            </Button>
                        </PopoverTrigger>
                        <PopoverContent
                            align="end"
                            className={cn(contentClass, 'w-80 p-0 sm:w-[480px] md:w-[640px]')}
                        >
                            <Command>
                                <CommandInput
                                    className="h-9"
                                    onValueChange={setSearch}
                                    placeholder={`Search ${label.toLowerCase()}...`}
                                    value={search}
                                />
                                <CommandList>
                                    <CommandEmpty>
                                        <div className="py-2 text-center">
                                            <p className="text-sm text-muted-foreground">
                                                No {label.toLowerCase()} found.
                                            </p>
                                            {search && allowCustom && (
                                                <Button
                                                    className="mt-2"
                                                    onClick={() => {
                                                        field.onChange(search);
                                                        setIsOpen(false);
                                                        setSearch('');
                                                    }}
                                                    size="sm"
                                                    variant="ghost"
                                                >
                                                    Use "{search}" as custom {label.toLowerCase()}
                                                </Button>
                                            )}
                                        </div>
                                    </CommandEmpty>
                                    <CommandGroup>
                                        {filteredOptions.map((option) => (
                                            <CommandItem
                                                key={option.name}
                                                onSelect={() => {
                                                    field.onChange(option.name);
                                                    onOptionSelect?.(option);
                                                    setIsOpen(false);
                                                    setSearch('');
                                                }}
                                                value={option.name}
                                            >
                                                <div className="flex w-full min-w-0 items-center justify-between gap-2">
                                                    <div className="flex min-w-0 items-center gap-2">
                                                        <span className="truncate">{option.name}</span>
                                                        {option.thinking && (
                                                            <Lightbulb className="size-3 text-muted-foreground" />
                                                        )}
                                                    </div>
                                                    <span className="flex-shrink-0 whitespace-nowrap text-xs text-muted-foreground">
                                                        {formatPrice(option.price)}
                                                    </span>
                                                </div>
                                                <Check
                                                    className={cn(
                                                        'ml-auto',
                                                        displayValue === option.name ? 'opacity-100' : 'opacity-0',
                                                    )}
                                                />
                                            </CommandItem>
                                        ))}
                                    </CommandGroup>
                                </CommandList>
                            </Command>
                        </PopoverContent>
                    </div>
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
        frequencyPenalty: z.preprocess(
            (value) => (value === '' || value === undefined ? null : value),
            z.number().nullable().optional(),
        ),
        maxLength: z.preprocess(
            (value) => (value === '' || value === undefined ? null : value),
            z.number().nullable().optional(),
        ),
        maxTokens: z.preprocess(
            (value) => (value === '' || value === undefined ? null : value),
            z.number().nullable().optional(),
        ),
        minLength: z.preprocess(
            (value) => (value === '' || value === undefined ? null : value),
            z.number().nullable().optional(),
        ),
        model: z.preprocess((value) => value || '', z.string().min(1, 'Model is required')),
        presencePenalty: z.preprocess(
            (value) => (value === '' || value === undefined ? null : value),
            z.number().nullable().optional(),
        ),
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
        repetitionPenalty: z.preprocess(
            (value) => (value === '' || value === undefined ? null : value),
            z.number().nullable().optional(),
        ),
        temperature: z.preprocess(
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
    })
    .optional();

// Define form schema
const formSchema = z.object({
    agents: z.record(z.string(), agentConfigSchema).optional(),
    name: z.preprocess(
        (value) => value || '',
        z.string().min(1, 'Provider name is required').max(50, 'Maximum 50 characters allowed'),
    ),
    type: z.preprocess((value) => value || '', z.string().min(1, 'Provider type is required')),
});

// Type for agents field in form
type FormAgents = FormData['agents'];

type FormData = z.infer<typeof formSchema>;

// Convert camelCase key to display name (e.g., 'simpleJson' -> 'Simple Json')
const getName = (key: string): string => key.replaceAll(/([A-Z])/g, ' $1').replace(/^./, (item) => item.toUpperCase());

// Helper function to convert string to ReasoningEffort enum
const getReasoningEffort = (effort: null | string | undefined): null | ReasoningEffort => {
    if (!effort) {
        return null;
    }

    switch (effort.toLowerCase()) {
        case 'high': {
            return ReasoningEffort.High;
        }

        case 'low': {
            return ReasoningEffort.Low;
        }

        case 'medium': {
            return ReasoningEffort.Medium;
        }

        default: {
            return null;
        }
    }
};

// Helper function to convert form data to GraphQL input
const transformFormToGraphQL = (
    formData: FormData,
): {
    agents: AgentsConfigInput;
    name: string;
    type: ProviderType;
} => {
    const agents = Object.entries(formData.agents || {})
        .filter(([key, data]) => key !== '__typename' && data?.model)
        .reduce((configs, [key, data]) => {
            const config: AgentConfigInput = {
                frequencyPenalty: data?.frequencyPenalty ?? null,
                maxLength: data?.maxLength ?? null,
                maxTokens: data?.maxTokens ?? null,
                minLength: data?.minLength ?? null,
                model: data!.model, // After filter, data and model are guaranteed to exist
                presencePenalty: data?.presencePenalty ?? null,
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
                reasoning: data?.reasoning
                    ? {
                          effort: getReasoningEffort(data?.reasoning.effort),
                          maxTokens: data?.reasoning.maxTokens ?? null,
                      }
                    : null,
                repetitionPenalty: data?.repetitionPenalty ?? null,
                temperature: data?.temperature ?? null,
                topK: data?.topK ?? null,
                topP: data?.topP ?? null,
            };

            return { ...configs, [key]: config };
        }, {} as AgentsConfigInput);

    return {
        agents,
        name: formData.name,
        type: formData.type as ProviderType,
    };
};

// Helper function to recursively remove __typename from objects
const normalizeGraphQLData = (obj: unknown): unknown => {
    if (obj === null || obj === undefined) {
        return obj;
    }

    if (Array.isArray(obj)) {
        return obj.map(normalizeGraphQLData);
    }

    if (typeof obj === 'object') {
        return Object.fromEntries(
            Object.entries(obj)
                .filter(([key]) => key !== '__typename')
                .map(([key, value]) => [key, normalizeGraphQLData(value)]),
        );
    }

    return obj;
};

interface TestResultsDialogProps {
    handleOpenChange: (isOpen: boolean) => void;
    isOpen: boolean;
    results: any;
}

// Component to render test results dialog
const TestResultsDialog = ({ handleOpenChange, isOpen, results }: TestResultsDialogProps) => {
    if (!results) {
        return null;
    }

    // Transform results object to array, removing __typename
    const agentResults = Object.entries(results)
        .filter(([key]) => key !== '__typename')
        .map(([agentType, agentData]: [string, any]) => ({
            agentType,
            tests: agentData?.tests || [],
        }));

    const getStatusIcon = (result: boolean) => {
        if (result === true) {
            return <CheckCircle className="size-4 text-green-500" />;
        } else if (result === false) {
            return <XCircle className="size-4 text-red-500" />;
        } else {
            return <Clock className="size-4 text-yellow-500" />;
        }
    };

    const getStatusColor = (result: boolean) => {
        if (result === true) {
            return 'text-green-600';
        } else if (result === false) {
            return 'text-red-600';
        } else {
            return 'text-yellow-600';
        }
    };

    return (
        <Dialog
            onOpenChange={handleOpenChange}
            open={isOpen}
        >
            <DialogContent className="flex max-h-[80vh] max-w-4xl flex-col">
                <DialogHeader className="flex-shrink-0">
                    <DialogTitle>Provider Test Results</DialogTitle>
                </DialogHeader>
                <div className="flex-1 space-y-6 overflow-y-auto">
                    <Accordion
                        className="w-full"
                        type="multiple"
                    >
                        {agentResults.map(({ agentType, tests }) => {
                            const testsCount = tests.length;
                            const successTestsCount = tests.filter((test: any) => test.result === true).length;

                            return (
                                <AccordionItem
                                    key={agentType}
                                    value={agentType}
                                >
                                    <AccordionTrigger className="text-left">
                                        <div className="mr-4 flex w-full items-center justify-between">
                                            <span className="text-lg font-semibold capitalize">{agentType}</span>
                                            <span className="text-sm text-muted-foreground">
                                                {successTestsCount}/{testsCount} tests passed
                                            </span>
                                        </div>
                                    </AccordionTrigger>
                                    <AccordionContent>
                                        <div className="space-y-3 pt-2">
                                            {tests.map((test: any, index: number) => (
                                                <div
                                                    className="rounded-lg border p-3"
                                                    key={index}
                                                >
                                                    <div className="mb-2 flex items-start justify-between">
                                                        <div className="flex items-center gap-2">
                                                            {getStatusIcon(test.result)}
                                                            <span className="font-medium">{test.name}</span>
                                                            {test.type && (
                                                                <span className="text-sm text-muted-foreground">
                                                                    ({test.type})
                                                                </span>
                                                            )}
                                                        </div>
                                                        <div className="flex items-center gap-3 text-sm text-muted-foreground">
                                                            {test.reasoning !== undefined && (
                                                                <span>Reasoning: {test.reasoning ? 'Yes' : 'No'}</span>
                                                            )}
                                                            {test.streaming !== undefined && (
                                                                <span>Streaming: {test.streaming ? 'Yes' : 'No'}</span>
                                                            )}
                                                            {test.latency && <span>Latency: {test.latency}ms</span>}
                                                        </div>
                                                    </div>
                                                    <div
                                                        className={`text-sm font-medium ${getStatusColor(test.result)}`}
                                                    >
                                                        Result:{' '}
                                                        {test.result === true
                                                            ? 'Success'
                                                            : test.result === false
                                                              ? 'Failed'
                                                              : 'Unknown'}
                                                    </div>
                                                    {test.error && (
                                                        <div className="mt-2 rounded border border-red-200 bg-red-50 p-2 text-sm text-red-700">
                                                            <strong>Error:</strong> {test.error}
                                                        </div>
                                                    )}
                                                </div>
                                            ))}
                                            {tests.length === 0 && (
                                                <div className="py-4 text-center text-muted-foreground">
                                                    No tests available for this agent
                                                </div>
                                            )}
                                        </div>
                                    </AccordionContent>
                                </AccordionItem>
                            );
                        })}
                    </Accordion>
                </div>
            </DialogContent>
        </Dialog>
    );
};

// Static mapping of agent keys to GraphQL enum types
const agentTypesMap: Record<string, AgentConfigType> = {
    adviser: AgentConfigType.Adviser,
    assistant: AgentConfigType.Assistant,
    coder: AgentConfigType.Coder,
    enricher: AgentConfigType.Enricher,
    generator: AgentConfigType.Generator,
    installer: AgentConfigType.Installer,
    pentester: AgentConfigType.Pentester,
    primaryAgent: AgentConfigType.PrimaryAgent,
    refiner: AgentConfigType.Refiner,
    reflector: AgentConfigType.Reflector,
    searcher: AgentConfigType.Searcher,
    simple: AgentConfigType.Simple,
    simpleJson: AgentConfigType.SimpleJson,
};

// Helper function to extract agent types from agents object
const extractAgentTypes = (agents: unknown): null | string[] => {
    if (!agents || typeof agents !== 'object') {
        return null;
    }

    const types = Object.entries(agents)
        .filter(([key, data]) => key !== '__typename' && data)
        .map(([key]) => key)
        .sort();

    return types.length > 0 ? types : null;
};

const SettingsProvider = () => {
    const { providerId } = useParams<{ providerId: string }>();
    const navigate = useNavigate();
    const [searchParams, setSearchParams] = useSearchParams();
    const { data, error, loading } = useSettingsProvidersQuery();
    const [createProvider, { error: createError, loading: isCreateLoading }] = useCreateProviderMutation();
    const [updateProvider, { error: updateError, loading: isUpdateLoading }] = useUpdateProviderMutation();
    const [deleteProvider, { error: deleteError, loading: isDeleteLoading }] = useDeleteProviderMutation();
    const [testProvider, { error: testError, loading: isTestLoading }] = useTestProviderMutation();
    const [testAgent, { error: agentTestError, loading: isAgentTestLoading }] = useTestAgentMutation();
    const [currentAgentKey, setCurrentAgentKey] = useState<null | string>(null);
    const [submitError, setSubmitError] = useState<null | string>(null);
    const [isTestDialogOpen, setIsTestDialogOpen] = useState(false);
    const [testResults, setTestResults] = useState<any>(null);
    const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
    const [isLeaveDialogOpen, setIsLeaveDialogOpen] = useState(false);
    const [pendingBrowserBack, setPendingBrowserBack] = useState(false);
    const allowBrowserLeaveRef = useRef(false);
    const hasPushedBlockerStateRef = useRef(false);

    const isNew = providerId === 'new';
    const isLoading = isCreateLoading || isUpdateLoading || isDeleteLoading;

    const form = useForm<FormData>({
        defaultValues: {
            agents: {},
            name: undefined,
            type: undefined,
        },
        resolver: zodResolver(formSchema),
    });

    const { control, formState, handleSubmit: handleFormSubmit, reset, setValue, trigger, watch } = form;

    const { isDirty } = useFormState({ control });

    // Maintain a blocker state at the top of history when form is dirty
    useEffect(() => {
        if (isDirty && !hasPushedBlockerStateRef.current) {
            window.history.pushState({ __pentagiBlock__: true }, '');
            hasPushedBlockerStateRef.current = true;
        }
    }, [isDirty]);

    // Intercept browser back using popstate when form is dirty
    useEffect(() => {
        const handlePopState = () => {
            if (!isDirty) {
                return;
            }

            if (allowBrowserLeaveRef.current) {
                // Allow single leave without blocking
                allowBrowserLeaveRef.current = false;

                return;
            }

            // User navigated back off the blocker entry to the previous one; go forward to stay
            setPendingBrowserBack(true);
            setIsLeaveDialogOpen(true);
            // Return to the blocker entry
            window.history.forward();
        };

        window.addEventListener('popstate', handlePopState, { capture: true });

        return () => {
            window.removeEventListener('popstate', handlePopState, { capture: true });
        };
    }, [isDirty]);

    // Watch selected type
    const selectedType = useWatch({ control, name: 'type' });

    // Watch provider name for delete confirmation dialog
    const providerName = useWatch({ control, name: 'name' });

    // Read query parameters for form initialization (stable)
    const formQueryParams = useMemo(
        () => ({
            id: searchParams.get('id'),
            type: searchParams.get('type'),
        }),
        [searchParams],
    );

    // Get dynamic agent types from data
    const getAgentTypes = () => {
        // Try to get agents from specific sources in priority order
        const agentsSource =
            // For new providers, use default provider for selected type
            (isNew &&
                selectedType &&
                data?.settingsProviders?.default?.[selectedType as keyof typeof data.settingsProviders.default]
                    ?.agents) ||
            // For existing providers, use current provider's agents
            (!isNew &&
                providerId &&
                data?.settingsProviders?.userDefined?.find((p: Provider) => p.id == providerId)?.agents) ||
            // Fallback to any available default provider
            (data?.settingsProviders?.default &&
                Object.values(data.settingsProviders.default).find((provider) => provider?.agents)?.agents) ||
            null;

        // Extract and return agent types, or fallback to hardcoded list
        return extractAgentTypes(agentsSource) ?? Object.keys(agentTypesMap);
    };

    const agentTypes = getAgentTypes();

    // Get available models filtered by selected provider type
    const availableModels = useMemo(() => {
        if (!data?.settingsProviders?.models || !selectedType) {
            return [];
        }

        // Filter models by selected provider type
        const { models } = data.settingsProviders;
        const providerModels = models[selectedType as keyof typeof models];

        if (!providerModels?.length) {
            return [];
        }

        return providerModels
            .map((model: any) => ({
                name: model.name,
                price: model.price,
                thinking: model.thinking,
            }))
            .filter((model) => model.name) // Remove any models without names
            .sort((a, b) => a.name.localeCompare(b.name));
    }, [data, selectedType]);

    // Fill agents when provider type is selected (only for new providers)
    useEffect(() => {
        if (!isNew || !selectedType || !data?.settingsProviders?.default || availableModels.length === 0) {
            return;
        }

        const defaultProvider =
            data.settingsProviders.default[selectedType as keyof typeof data.settingsProviders.default];

        if (defaultProvider?.agents) {
            const agents = Object.fromEntries(
                Object.entries(defaultProvider.agents)
                    .filter(([key]) => key !== '__typename')
                    .map(([key, data]) => {
                        // const agent = Object.fromEntries(
                        //     Object.entries(data).filter(([key]) => key !== '__typename'),
                        // ) as AgentConfigInput;
                        const agent = { ...data };

                        // Check if the model from defaultProvider exists in availableModels
                        if (agent.model && !availableModels.find((m) => m.name === agent.model)) {
                            // Use first available model if default model not found
                            agent.model = availableModels[0]?.name || agent.model;
                        }

                        return [key, agent];
                    }),
            );

            setValue('agents', normalizeGraphQLData(agents) as FormAgents);
        }
    }, [availableModels, data, isNew, selectedType, setValue]);

    // Update query parameter when type changes (only for new providers)
    useEffect(() => {
        if (!isNew) {
            // Clear query parameters for existing providers
            if (searchParams.size > 0) {
                setSearchParams({});
            }

            return;
        }

        // Don't update query params if we're copying from existing provider
        const queryId = searchParams.get('id');

        if (queryId) {
            return;
        }

        // Don't update query params on initial load if we're reading from query params
        const queryType = searchParams.get('type');

        if (!selectedType && queryType) {
            return;
        }

        // Update query parameter based on selected type
        setSearchParams((prev) => {
            const params = new URLSearchParams(prev);

            if (selectedType) {
                params.set('type', selectedType);
            } else {
                params.delete('type');
            }

            return params;
        });
    }, [selectedType, setSearchParams, isNew, searchParams]); // Include searchParams since we read from it

    // Fill form with data when available
    useEffect(() => {
        if (!data?.settingsProviders) {
            return;
        }

        const providers = data.settingsProviders;

        if (isNew || !providerId) {
            // For new provider, start with empty form but check for type query parameter
            const queryType = formQueryParams.type ?? undefined;
            const queryId = formQueryParams.id;

            // If we have an id in query params, copy from existing provider
            if (queryId && data?.settingsProviders?.userDefined) {
                const sourceProvider = data.settingsProviders.userDefined.find((p: Provider) => p.id == queryId);

                if (sourceProvider) {
                    const { agents, name, type: sourceType } = sourceProvider;

                    reset({
                        agents: agents ? (normalizeGraphQLData(agents) as FormAgents) : {},
                        name: `${name} (Copy)`,
                        type: sourceType ?? undefined,
                    });

                    return;
                }
            } else if (queryType && data?.settingsProviders?.default) {
                const defaultProvider =
                    data.settingsProviders.default[queryType as keyof typeof data.settingsProviders.default];

                reset({
                    agents: defaultProvider?.agents ? (normalizeGraphQLData(defaultProvider.agents) as FormAgents) : {},
                    name: undefined,
                    type: queryType,
                });
            }

            // Default new provider form - but only if selectedType is not set
            // to avoid conflicts with agent filling useEffect
            if (!selectedType) {
                reset({
                    agents: {},
                    name: undefined,
                    type: queryType,
                });
            }

            return;
        }

        const provider = providers.userDefined?.find((provider: Provider) => provider.id == providerId);

        if (!provider) {
            navigate('/settings/providers');

            return;
        }

        const { agents, name, type } = provider;

        reset({
            agents: agents ? (normalizeGraphQLData(agents) as FormAgents) : {},
            name: name || undefined,
            type: type || undefined,
        });
    }, [data, formQueryParams, isNew, navigate, providerId, reset, selectedType]);

    const handleSubmit = async () => {
        // Get all form data including disabled fields
        // Note: getValues() excludes disabled fields, watch() includes them
        const formData = watch();

        try {
            setSubmitError(null);

            const mutationData = transformFormToGraphQL(formData);

            if (isNew) {
                // Create new provider
                await createProvider({
                    refetchQueries: ['settingsProviders'],
                    variables: mutationData,
                });
            } else {
                // Update existing provider
                await updateProvider({
                    refetchQueries: ['settingsProviders'],
                    variables: {
                        ...mutationData,
                        providerId: providerId!,
                    },
                });
            }

            // Navigate back to providers list on success
            navigate('/settings/providers');
        } catch (error) {
            console.error('Submit error:', error);
            setSubmitError(error instanceof Error ? error.message : 'An error occurred while saving');
        }
    };

    const handleDelete = () => {
        if (isNew || !providerId) {
            return;
        }

        setIsDeleteDialogOpen(true);
    };

    const handleConfirmDelete = async () => {
        if (isNew || !providerId) {
            return;
        }

        try {
            setSubmitError(null);

            await deleteProvider({
                refetchQueries: ['settingsProviders'],
                variables: { providerId },
            });

            // Navigate back to providers list on success
            navigate('/settings/providers');
        } catch (error) {
            console.error('Delete error:', error);
            setSubmitError(error instanceof Error ? error.message : 'An error occurred while deleting');
        }
    };

    // Test entire provider (all agents)
    const handleTest = async () => {
        // Trigger form validation
        const isValid = await trigger();

        if (!isValid) {
            const { errors } = formState;

            // Helper function to format field names for display
            const formatFieldName = (fieldPath: string): string => {
                return fieldPath
                    .split('.')
                    .map((part) => {
                        // Capitalize first letter and add spaces before uppercase letters
                        return part.charAt(0).toUpperCase() + part.slice(1).replaceAll(/([A-Z])/g, ' $1');
                    })
                    .join(' → ');
            };

            // Show validation errors to user
            const errorMessages = Object.entries(errors)
                .map(([field, error]: [string, any]) => {
                    if (error?.message) {
                        return `• ${formatFieldName(field)}: ${error.message}`;
                    }

                    if (error && typeof error === 'object') {
                        // Handle nested errors (like agents.simple.model)
                        return Object.entries(error)
                            .map(([subField, subError]: [string, any]) => {
                                if (subError?.message) {
                                    return `• ${formatFieldName(`${field}.${subField}`)}: ${subError.message}`;
                                }

                                if (subError && typeof subError === 'object') {
                                    return Object.entries(subError)
                                        .map(([nestedField, nestedError]: [string, any]) => {
                                            if (nestedError?.message) {
                                                return `• ${formatFieldName(`${field}.${subField}.${nestedField}`)}: ${nestedError.message}`;
                                            }

                                            return null;
                                        })
                                        .filter(Boolean)
                                        .join('\n');
                                }

                                return null;
                            })
                            .filter(Boolean)
                            .join('\n');
                    }

                    return null;
                })
                .filter(Boolean)
                .join('\n');

            setSubmitError(`Please fix the following validation errors:\n\n${errorMessages}`);

            return;
        }

        try {
            setSubmitError(null);

            // Get form data and transform it - including disabled fields
            const formData = watch();
            const { agents, type } = transformFormToGraphQL(formData);
            const result = await testProvider({
                variables: {
                    agents,
                    type,
                },
            });

            setTestResults(result.data?.testProvider);
            setIsTestDialogOpen(true);
        } catch (error) {
            console.error('Test error:', error);
            setSubmitError(error instanceof Error ? error.message : 'An error occurred while testing');
        }
    };

    // Test a single agent (uses testAgent where supported, otherwise falls back to filtered provider test)
    const handleTestAgent = async (agentKey: string) => {
        // Validate only fields for this agent and general required fields
        const isValid = await trigger();

        if (!isValid) {
            const { errors } = formState;
            const formatFieldName = (fieldPath: string): string =>
                fieldPath
                    .split('.')
                    .map((part) => part.charAt(0).toUpperCase() + part.slice(1).replaceAll(/([A-Z])/g, ' $1'))
                    .join(' → ');

            const errorMessages = Object.entries(errors)
                .map(([field, error]: [string, any]) => {
                    if (error?.message) {
                        return `• ${formatFieldName(field)}: ${error.message}`;
                    }

                    if (error && typeof error === 'object') {
                        return Object.entries(error)
                            .map(([subField, subError]: [string, any]) => {
                                if (subError?.message) {
                                    return `• ${formatFieldName(`${field}.${subField}`)}: ${subError.message}`;
                                }

                                if (subError && typeof subError === 'object') {
                                    return Object.entries(subError)
                                        .map(([nestedField, nestedError]: [string, any]) => {
                                            if (nestedError?.message) {
                                                return `• ${formatFieldName(`${field}.${subField}.${nestedField}`)}: ${nestedError.message}`;
                                            }

                                            return null;
                                        })
                                        .filter(Boolean)
                                        .join('\n');
                                }

                                return null;
                            })
                            .filter(Boolean)
                            .join('\n');
                    }

                    return null;
                })
                .filter(Boolean)
                .join('\n');

            setSubmitError(`Please fix the following validation errors:\n\n${errorMessages}`);

            return;
        }

        try {
            setSubmitError(null);
            setCurrentAgentKey(agentKey);
            // Note: getValues() excludes disabled fields, watch() includes them
            const formData = watch();
            const { agents, type } = transformFormToGraphQL(formData);

            const agent = agents[agentKey as keyof AgentsConfigInput] as AgentConfigInput;

            const singleResult = await testAgent({
                variables: { agent, agentType: agentTypesMap[agentKey] ?? AgentConfigType.Simple, type },
            });
            setTestResults({ [agentKey]: singleResult.data?.testAgent });
            setIsTestDialogOpen(true);
            setCurrentAgentKey(null);

            return;
        } catch (error) {
            console.error('Test error:', error);
            setSubmitError(error instanceof Error ? error.message : 'An error occurred while testing');
            setCurrentAgentKey(null);
        }
    };

    const handleBack = () => {
        if (isDirty) {
            setIsLeaveDialogOpen(true);

            return;
        }

        navigate('/settings/providers');
    };

    const handleConfirmLeave = () => {
        if (pendingBrowserBack) {
            allowBrowserLeaveRef.current = true;
            setPendingBrowserBack(false);
            // Skip the blocker entry and go to the real previous page
            window.history.go(-2);

            return;
        }

        navigate('/settings/providers');
    };

    const handleLeaveDialogOpenChange = (open: boolean) => {
        if (!open && pendingBrowserBack) {
            setPendingBrowserBack(false);
        }

        setIsLeaveDialogOpen(open);
    };

    if (loading) {
        return (
            <StatusCard
                description="Please wait while we fetch provider configuration"
                icon={<Loader2 className="size-16 animate-spin text-muted-foreground" />}
                title="Loading provider data..."
            />
        );
    }

    if (error) {
        return (
            <Alert variant="destructive">
                <AlertCircle className="size-4" />
                <AlertTitle>Error loading provider data</AlertTitle>
                <AlertDescription>{error.message}</AlertDescription>
            </Alert>
        );
    }

    const providers = data?.settingsProviders?.models
        ? Object.keys(data?.settingsProviders.models).filter((key) => key !== '__typename')
        : [];

    const mutationError = createError || updateError || deleteError || testError || agentTestError || submitError;

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
                            className="space-y-6"
                            id="provider-form"
                            onSubmit={handleFormSubmit(handleSubmit)}
                        >
                            {/* Error Alert */}
                            {mutationError && (
                                <Alert variant="destructive">
                                    <AlertCircle className="size-4" />
                                    <AlertTitle>Error</AlertTitle>
                                    <AlertDescription>
                                        {mutationError instanceof Error ? (
                                            mutationError.message
                                        ) : (
                                            <div className="whitespace-pre-line">{mutationError}</div>
                                        )}
                                    </AlertDescription>
                                </Alert>
                            )}

                            {/* Form fields */}
                            <FormComboboxItem
                                allowCustom={false}
                                control={control}
                                description="The type of language model provider"
                                disabled={isLoading || !!selectedType}
                                label="Type"
                                name="type"
                                options={providers}
                                placeholder="Select provider"
                            />

                            <FormInputStringItem
                                control={control}
                                description="A unique name for your provider configuration"
                                disabled={isLoading}
                                label="Name"
                                name="name"
                                placeholder="Enter provider name"
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
                                    className="w-full"
                                    type="multiple"
                                >
                                    {agentTypes.map((agentKey) => (
                                        <AccordionItem
                                            key={agentKey}
                                            value={agentKey}
                                        >
                                            <AccordionTrigger className="group text-left hover:no-underline">
                                                <div className="flex w-full items-center justify-between gap-2">
                                                    <span className="group-hover:underline">{getName(agentKey)}</span>
                                                    <span
                                                        className={cn(
                                                            'mr-2 flex items-center gap-1 rounded border px-2 py-1 text-xs hover:bg-accent hover:text-accent-foreground',
                                                            (isTestLoading || isAgentTestLoading) &&
                                                                'pointer-events-none cursor-not-allowed opacity-50',
                                                        )}
                                                        onClick={(event) => {
                                                            if (isTestLoading || isAgentTestLoading) {
                                                                return;
                                                            }

                                                            event.stopPropagation();
                                                            handleTestAgent(agentKey);
                                                        }}
                                                    >
                                                        {isAgentTestLoading && currentAgentKey === agentKey ? (
                                                            <Loader2 className="size-4 animate-spin" />
                                                        ) : (
                                                            <Play className="size-4" />
                                                        )}
                                                        <span className="!no-underline hover:!no-underline">
                                                            {isAgentTestLoading && currentAgentKey === agentKey
                                                                ? 'Testing...'
                                                                : 'Test'}
                                                        </span>
                                                    </span>
                                                </div>
                                            </AccordionTrigger>
                                            <AccordionContent className="space-y-4 pt-4">
                                                <div className="grid grid-cols-1 gap-4 p-px md:grid-cols-2">
                                                    {/* Model field */}
                                                    <FormModelComboboxItem
                                                        control={control}
                                                        disabled={isLoading}
                                                        label="Model"
                                                        name={`agents.${agentKey}.model`}
                                                        onOptionSelect={(option) => {
                                                            {
                                                                /* Update price fields */
                                                            }

                                                            const price = option?.price;

                                                            setValue(
                                                                `agents.${agentKey}.price.input` as const,
                                                                price?.input ?? null,
                                                            );
                                                            setValue(
                                                                `agents.${agentKey}.price.output` as const,
                                                                price?.output ?? null,
                                                            );
                                                        }}
                                                        options={availableModels}
                                                        placeholder="Select or enter model name"
                                                    />

                                                    {/* Temperature field */}
                                                    <FormInputNumberItem
                                                        control={control}
                                                        disabled={isLoading}
                                                        label="Temperature"
                                                        max="2"
                                                        min="0"
                                                        name={`agents.${agentKey}.temperature`}
                                                        placeholder="0.7"
                                                        step="0.1"
                                                    />

                                                    {/* Max Tokens field */}
                                                    <FormInputNumberItem
                                                        control={control}
                                                        disabled={isLoading}
                                                        label="Max Tokens"
                                                        min="1"
                                                        name={`agents.${agentKey}.maxTokens`}
                                                        placeholder="1000"
                                                        valueType="integer"
                                                    />

                                                    {/* Top P field */}
                                                    <FormInputNumberItem
                                                        control={control}
                                                        disabled={isLoading}
                                                        label="Top P"
                                                        max="1"
                                                        min="0"
                                                        name={`agents.${agentKey}.topP`}
                                                        placeholder="0.9"
                                                        step="0.01"
                                                    />

                                                    {/* Top K field */}
                                                    <FormInputNumberItem
                                                        control={control}
                                                        disabled={isLoading}
                                                        label="Top K"
                                                        min="1"
                                                        name={`agents.${agentKey}.topK`}
                                                        placeholder="40"
                                                        valueType="integer"
                                                    />

                                                    {/* Min Length field */}
                                                    <FormInputNumberItem
                                                        control={control}
                                                        disabled={isLoading}
                                                        label="Min Length"
                                                        min="0"
                                                        name={`agents.${agentKey}.minLength`}
                                                        placeholder="0"
                                                        valueType="integer"
                                                    />

                                                    {/* Max Length field */}
                                                    <FormInputNumberItem
                                                        control={control}
                                                        disabled={isLoading}
                                                        label="Max Length"
                                                        min="1"
                                                        name={`agents.${agentKey}.maxLength`}
                                                        placeholder="2000"
                                                        valueType="integer"
                                                    />

                                                    {/* Repetition Penalty field */}
                                                    <FormInputNumberItem
                                                        control={control}
                                                        disabled={isLoading}
                                                        label="Repetition Penalty"
                                                        max="2"
                                                        min="0"
                                                        name={`agents.${agentKey}.repetitionPenalty`}
                                                        placeholder="1.0"
                                                        step="0.01"
                                                    />

                                                    {/* Frequency Penalty field */}
                                                    <FormInputNumberItem
                                                        control={control}
                                                        disabled={isLoading}
                                                        label="Frequency Penalty"
                                                        max="2"
                                                        min="0"
                                                        name={`agents.${agentKey}.frequencyPenalty`}
                                                        placeholder="0.0"
                                                        step="0.01"
                                                    />

                                                    {/* Presence Penalty field */}
                                                    <FormInputNumberItem
                                                        control={control}
                                                        disabled={isLoading}
                                                        label="Presence Penalty"
                                                        max="2"
                                                        min="0"
                                                        name={`agents.${agentKey}.presencePenalty`}
                                                        placeholder="0.0"
                                                        step="0.01"
                                                    />
                                                </div>

                                                {/* Reasoning Configuration */}
                                                <div className="col-span-full p-px">
                                                    <div className="mt-6 space-y-4">
                                                        <h4 className="text-sm font-medium">Reasoning Configuration</h4>
                                                        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                                                            {/* Reasoning Effort field */}
                                                            <FormField
                                                                control={control}
                                                                name={`agents.${agentKey}.reasoning.effort`}
                                                                render={({ field }) => (
                                                                    <FormItem>
                                                                        <FormLabel>Reasoning Effort</FormLabel>
                                                                        <Select
                                                                            defaultValue={field.value ?? 'none'}
                                                                            disabled={isLoading}
                                                                            onValueChange={(value) =>
                                                                                field.onChange(
                                                                                    value !== 'none' ? value : null,
                                                                                )
                                                                            }
                                                                        >
                                                                            <FormControl>
                                                                                <SelectTrigger>
                                                                                    <SelectValue placeholder="Select effort level (optional)" />
                                                                                </SelectTrigger>
                                                                            </FormControl>
                                                                            <SelectContent>
                                                                                <SelectItem value="none">
                                                                                    Not selected
                                                                                </SelectItem>
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
                                                                control={control}
                                                                disabled={isLoading}
                                                                label="Reasoning Max Tokens"
                                                                min="1"
                                                                name={`agents.${agentKey}.reasoning.maxTokens`}
                                                                placeholder="1000"
                                                                valueType="integer"
                                                            />
                                                        </div>
                                                    </div>
                                                </div>

                                                {/* Price Configuration */}
                                                <div className="col-span-full p-px">
                                                    <div className="mt-6 space-y-4">
                                                        <h4 className="text-sm font-medium">Price Configuration</h4>
                                                        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                                                            {/* Price Input field */}
                                                            <FormInputNumberItem
                                                                control={control}
                                                                disabled={isLoading}
                                                                label="Input Price"
                                                                min="0"
                                                                name={`agents.${agentKey}.price.input`}
                                                                placeholder="0.001"
                                                                step="0.000001"
                                                            />

                                                            {/* Price Output field */}
                                                            <FormInputNumberItem
                                                                control={control}
                                                                disabled={isLoading}
                                                                label="Output Price"
                                                                min="0"
                                                                name={`agents.${agentKey}.price.output`}
                                                                placeholder="0.002"
                                                                step="0.000001"
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
            <div className="sticky -bottom-4 -mx-4 -mb-4 mt-4 flex items-center border-t bg-background p-4 shadow-lg">
                <div className="flex space-x-2">
                    {/* Delete button - only show when editing existing provider */}
                    {!isNew && (
                        <Button
                            disabled={isLoading}
                            onClick={handleDelete}
                            type="button"
                            variant="destructive"
                        >
                            {isDeleteLoading ? (
                                <Loader2 className="size-4 animate-spin" />
                            ) : (
                                <Trash2 className="size-4" />
                            )}
                            {isDeleteLoading ? 'Deleting...' : 'Delete'}
                        </Button>
                    )}
                    <Button
                        disabled={isLoading || isTestLoading || isAgentTestLoading}
                        onClick={() => handleTest()}
                        type="button"
                        variant="outline"
                    >
                        {isTestLoading ? <Loader2 className="size-4 animate-spin" /> : <Play className="size-4" />}
                        {isTestLoading ? 'Testing...' : 'Test'}
                    </Button>
                </div>

                <div className="ml-auto flex space-x-2">
                    <Button
                        disabled={isLoading}
                        onClick={handleBack}
                        type="button"
                        variant="outline"
                    >
                        Cancel
                    </Button>
                    <Button
                        disabled={isLoading}
                        form="provider-form"
                        type="submit"
                        variant="secondary"
                    >
                        {isLoading ? <Loader2 className="size-4 animate-spin" /> : <Save className="size-4" />}
                        {isLoading ? 'Saving...' : isNew ? 'Create Provider' : 'Update Provider'}
                    </Button>
                </div>
            </div>

            <TestResultsDialog
                handleOpenChange={setIsTestDialogOpen}
                isOpen={isTestDialogOpen}
                results={testResults}
            />

            <ConfirmationDialog
                cancelText="Cancel"
                confirmText="Delete"
                handleConfirm={handleConfirmDelete}
                handleOpenChange={setIsDeleteDialogOpen}
                isOpen={isDeleteDialogOpen}
                itemName={providerName}
                itemType="provider"
            />

            <ConfirmationDialog
                cancelText="Stay"
                confirmIcon={undefined}
                confirmText="Leave"
                confirmVariant="destructive"
                description="You have unsaved changes. Are you sure you want to leave without saving?"
                handleConfirm={handleConfirmLeave}
                handleOpenChange={handleLeaveDialogOpenChange}
                isOpen={isLeaveDialogOpen}
                title="Discard changes?"
            />
        </>
    );
};

export default SettingsProvider;
