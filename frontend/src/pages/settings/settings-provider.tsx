import { useMutation, useQuery } from '@apollo/client/react';
import {
    AlertCircle,
    Check,
    CheckCircle,
    ChevronsUpDown,
    Clock,
    Ellipsis,
    Lightbulb,
    Loader2,
    Play,
    Plug,
    Save,
    Trash2,
    XCircle,
} from 'lucide-react';
import { type ComponentProps, useEffect, useMemo, useState } from 'react';
import {
    type Control,
    type FieldPath,
    type FieldValues,
    useController,
    type UseFormSetValue,
    useFormState,
    useWatch,
} from 'react-hook-form';
import { useNavigate, useParams, useSearchParams } from 'react-router-dom';
import { toast } from 'sonner';
import { z } from 'zod';

import type { AgentConfigInput, AgentsConfigInput, ProviderConfigFragmentFragment } from '@/graphql/types';

import {
    AppHeader,
    AppHeaderAction,
    AppHeaderActions,
    AppHeaderContent,
    AppHeaderTitle,
} from '@/components/layouts/app/app-header';
import ConfirmationDialog from '@/components/shared/confirmation-dialog';
import { DetailSplitLayout } from '@/components/shared/detail-split-layout';
import { UnsavedChangesDialog, useUnsavedChangesGuard } from '@/components/shared/unsaved-changes';
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from '@/components/ui/accordion';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from '@/components/ui/command';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { FormSubmitButton } from '@/components/ui/form-submit-button';
import { Input } from '@/components/ui/input';
import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupInput } from '@/components/ui/input-group';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { StatusCard } from '@/components/ui/status-card';
import {
    AgentConfigType,
    CreateProviderDocument,
    DeleteProviderDocument,
    ModelReasoningMode,
    ProviderType,
    ReasoningEffort,
    ReasoningMode,
    SettingsProvidersDocument,
    TestAgentDocument,
    TestProviderDocument,
    UpdateProviderDocument,
} from '@/graphql/types';
import { useAppForm } from '@/hooks/use-app-form';
import { useBreakpoint } from '@/hooks/use-breakpoint';
import { routes } from '@/lib/routes';
import { cn } from '@/lib/utils';

interface ProviderTest {
    error?: null | string;
    latency?: null | number;
    name?: null | string;
    reasoning?: boolean | null;
    result?: boolean | null;
    streaming?: boolean | null;
    type?: null | string;
}

type ProviderTestResults = Record<string, null | undefined | { tests?: null | ProviderTest[] }>;

const formatFieldName = (fieldPath: string): string =>
    fieldPath
        .split('.')
        .map((part) => part.charAt(0).toUpperCase() + part.slice(1).replaceAll(/([A-Z])/g, ' $1'))
        .join(' → ');

const getErrorMessage = (error: unknown): string | undefined =>
    typeof error === 'object' && error !== null && typeof (error as { message?: unknown }).message === 'string'
        ? (error as { message: string }).message
        : undefined;

const formatFormErrors = (errors: Record<string, unknown>, prefix = ''): string =>
    Object.entries(errors)
        .flatMap(([field, error]) => {
            const path = prefix ? `${prefix}.${field}` : field;
            const message = getErrorMessage(error);

            if (message) {
                return [`• ${formatFieldName(path)}: ${message}`];
            }

            if (error && typeof error === 'object') {
                const nested = formatFormErrors(error as Record<string, unknown>, path);

                return nested ? [nested] : [];
            }

            return [];
        })
        .join('\n');

interface BaseFieldProps<T extends FieldValues = FieldValues> extends ControllerProps<T> {
    label: string;
}

interface BaseInputProps {
    placeholder?: string;
}

interface ControllerProps<T extends FieldValues = FieldValues> {
    control: Control<T>;
    disabled?: boolean;
    name: FieldPath<T>;
}

interface FormComboboxItemProps<T extends FieldValues = FieldValues> extends BaseFieldProps<T>, BaseInputProps {
    allowCustom?: boolean;
    contentClass?: string;
    description?: string;
    options: string[];
}

interface FormInputNumberItemProps<T extends FieldValues = FieldValues> extends BaseFieldProps<T>, NumberInputProps {
    description?: string;
    valueType?: 'float' | 'integer';
}

interface FormInputStringItemProps<T extends FieldValues = FieldValues> extends BaseFieldProps<T>, BaseInputProps {
    description?: string;
}

interface FormModelComboboxItemProps<T extends FieldValues = FieldValues> extends BaseFieldProps<T>, BaseInputProps {
    allowCustom?: boolean;
    contentClass?: string;
    description?: string;
    onOptionSelect?: (option: ModelOption) => void;
    options: ModelOption[];
}

interface ModelOption {
    name: string;
    price?: null | { cacheRead: number; cacheWrite: number; input: number; output: number };
    reasoning?: null | {
        cannotDisable?: boolean | null;
        defaultOn?: boolean | null;
        efforts?: null | ReasoningEffort[];
        mode?: ModelReasoningMode | null;
        supported?: boolean | null;
    };
    thinking?: boolean | null;
}

interface NumberInputProps extends BaseInputProps {
    max?: string;
    min?: string;
    step?: string;
}

type Provider = ProviderConfigFragmentFragment;

function FormComboboxItem<T extends FieldValues = FieldValues>({
    allowCustom = true,
    contentClass,
    control,
    description,
    disabled,
    label,
    name,
    options,
    placeholder,
}: FormComboboxItemProps<T>) {
    const { field, fieldState } = useController({
        control,
        defaultValue: undefined,
        disabled,
        name,
    });

    const [isOpen, setIsOpen] = useState(false);
    const [search, setSearch] = useState('');

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
                                        <p className="text-muted-foreground text-sm">No {label.toLowerCase()} found.</p>
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
}

function FormInputNumberItem<T extends FieldValues = FieldValues>({
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
}: FormInputNumberItemProps<T>) {
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
}

function FormInputStringItem<T extends FieldValues = FieldValues>({
    control,
    description,
    disabled,
    label,
    name,
    placeholder,
}: FormInputStringItemProps<T>) {
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
}

function FormModelComboboxItem<T extends FieldValues = FieldValues>({
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
}: FormModelComboboxItemProps<T>) {
    const { field, fieldState } = useController({
        control,
        defaultValue: undefined,
        disabled,
        name,
    });

    const [isOpen, setIsOpen] = useState(false);
    const [search, setSearch] = useState('');

    const filteredOptions = options.filter((option) => option.name?.toLowerCase().includes(search?.toLowerCase()));

    const displayValue = field.value ?? '';

    const formatPrice = (
        price?: null | { cacheRead: number; cacheWrite: number; input: number; output: number },
    ): string => {
        if (!price || ((!price.input || price.input === 0) && (!price.output || price.output === 0))) {
            return 'free';
        }

        const formatValue = (value: number): string => {
            return value.toFixed(6).replace(/\.?0+$/, '');
        };

        const basePrice = `$${formatValue(price.input)}/$${formatValue(price.output)}`;

        const hasCachePrices = (price.cacheRead && price.cacheRead > 0) || (price.cacheWrite && price.cacheWrite > 0);

        if (hasCachePrices) {
            const cacheParts: string[] = [];

            if (price.cacheRead && price.cacheRead > 0) {
                cacheParts.push(`R:$${formatValue(price.cacheRead)}`);
            }

            if (price.cacheWrite && price.cacheWrite > 0) {
                cacheParts.push(`W:$${formatValue(price.cacheWrite)}`);
            }

            return `${basePrice} (${cacheParts.join(', ')})`;
        }

        return basePrice;
    };

    return (
        <FormItem>
            <FormLabel>{label}</FormLabel>
            <FormControl>
                <Popover
                    onOpenChange={setIsOpen}
                    open={isOpen}
                >
                    <InputGroup>
                        <InputGroupInput
                            disabled={disabled}
                            onChange={(event) => {
                                const { value } = event.target;
                                field.onChange(value);
                                const matched = options.find((option) => option.name === value);

                                if (matched) {
                                    onOptionSelect?.(matched);
                                }
                            }}
                            placeholder={placeholder}
                            value={displayValue}
                        />
                        <InputGroupAddon align="inline-end">
                            <PopoverTrigger asChild>
                                <InputGroupButton
                                    aria-label={`Open ${label.toLowerCase()} list`}
                                    disabled={disabled}
                                    size="icon-sm"
                                >
                                    <ChevronsUpDown className="opacity-50" />
                                </InputGroupButton>
                            </PopoverTrigger>
                        </InputGroupAddon>
                    </InputGroup>
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
                                        <p className="text-muted-foreground text-sm">No {label.toLowerCase()} found.</p>
                                        {search && allowCustom && (
                                            <Button
                                                className="mt-2"
                                                onClick={() => {
                                                    field.onChange(search);
                                                    onOptionSelect?.({ name: search });
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
                                                        <Lightbulb className="text-muted-foreground size-3" />
                                                    )}
                                                </div>
                                                <span className="text-muted-foreground shrink-0 text-xs whitespace-nowrap">
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
                </Popover>
            </FormControl>
            {description && <FormDescription>{description}</FormDescription>}
            {fieldState.error && <FormMessage>{fieldState.error.message}</FormMessage>}
        </FormItem>
    );
}

const optionalNumber = z.number().nullable().optional();

const requiredString = (message: string) =>
    z
        .string()
        .optional()
        .transform((value) => value ?? '')
        .pipe(z.string().min(1, message));

const agentConfigSchema = z
    .object({
        frequencyPenalty: optionalNumber,
        maxLength: optionalNumber,
        maxTokens: optionalNumber,
        minLength: optionalNumber,
        model: requiredString('Model is required'),
        presencePenalty: optionalNumber,
        price: z
            .object({
                cacheRead: optionalNumber,
                cacheWrite: optionalNumber,
                input: optionalNumber,
                output: optionalNumber,
            })
            .nullable()
            .optional(),
        reasoning: z
            .object({
                effort: z.string().nullable().optional(),
                maxTokens: optionalNumber,
                mode: z.string().nullable().optional(),
            })
            .nullable()
            .optional(),
        repetitionPenalty: optionalNumber,
        temperature: optionalNumber,
        topK: optionalNumber,
        topP: optionalNumber,
    })
    .refine((data) => data.minLength == null || data.maxLength == null || data.minLength <= data.maxLength, {
        message: 'Min length must not exceed max length',
        path: ['minLength'],
    })
    .refine((data) => data.reasoning?.maxTokens == null || data.reasoning.maxTokens <= 32000, {
        message: 'Maximum 32000 tokens',
        path: ['reasoning', 'maxTokens'],
    })
    .optional();

const formSchema = z.object({
    agents: z.record(z.string(), agentConfigSchema).optional(),
    name: requiredString('Provider name is required').pipe(z.string().max(50, 'Maximum 50 characters allowed')),
    type: requiredString('Provider type is required'),
});

type FormAgents = FormInput['agents'];

type FormData = z.output<typeof formSchema>;

type FormInput = z.input<typeof formSchema>;

const getName = (key: string): string => key.replaceAll(/([A-Z])/g, ' $1').replace(/^./, (item) => item.toUpperCase());

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

        case 'max': {
            return ReasoningEffort.Max;
        }

        case 'medium': {
            return ReasoningEffort.Medium;
        }

        case 'xhigh': {
            return ReasoningEffort.Xhigh;
        }

        default: {
            return null;
        }
    }
};

const getReasoningMode = (mode: null | string | undefined): null | ReasoningMode => {
    switch (mode) {
        case ReasoningMode.Adaptive: {
            return ReasoningMode.Adaptive;
        }

        case ReasoningMode.Budget: {
            return ReasoningMode.Budget;
        }

        case ReasoningMode.Off: {
            return ReasoningMode.Off;
        }

        default: {
            return null;
        }
    }
};

const reasoningEffortLabel: Record<ReasoningEffort, string> = {
    [ReasoningEffort.High]: 'High',
    [ReasoningEffort.Low]: 'Low',
    [ReasoningEffort.Max]: 'Max',
    [ReasoningEffort.Medium]: 'Medium',
    [ReasoningEffort.Xhigh]: 'Extra High',
};

const defaultReasoningEfforts: ReasoningEffort[] = [ReasoningEffort.Low, ReasoningEffort.Medium, ReasoningEffort.High];

// Gated by the selected model's declared capability (models.yml), not a model-name
// allowlist: adaptive-only models lock to adaptive, and effort options follow the model.
function ReasoningFields({
    agentKey,
    control,
    isLoading,
    models,
    setValue,
}: {
    agentKey: string;
    control: Control<FormInput>;
    isLoading: boolean;
    models: ModelOption[];
    setValue: UseFormSetValue<FormInput>;
}) {
    const selectedModel = useWatch({ control, name: `agents.${agentKey}.model` });
    const reasoningMode = useWatch({ control, name: `agents.${agentKey}.reasoning.mode` });
    const capability = models.find((model) => model.name === selectedModel)?.reasoning ?? null;
    const isAdaptiveOnly = capability?.mode === ModelReasoningMode.AdaptiveOnly;
    const supportsAdaptive = isAdaptiveOnly || capability?.mode === ModelReasoningMode.Adaptive;
    // Off is offered only where the capability confirms a disable actually takes
    // effect (cannotDisable=false). An absent capability, an always-on model, or a
    // model where Off would be a silent no-op keeps the Off option hidden.
    const canDisable = capability != null && capability.cannotDisable !== true;
    const isOff = reasoningMode === ReasoningMode.Off;
    const allowedEfforts =
        capability?.efforts && capability.efforts.length > 0 ? capability.efforts : defaultReasoningEfforts;

    // Reconcile a stale Off when the current model can't disable (e.g. after typing
    // a custom model name): otherwise mode=off is orphaned with no control to clear
    // it and silently persists, disabling thinking with no UI affordance. Guarded on
    // models.length so a still-loading capability list does not wipe a saved Off.
    useEffect(() => {
        if (isOff && !canDisable && models.length > 0) {
            setValue(`agents.${agentKey}.reasoning.mode` as const, null);
        }
    }, [isOff, canDisable, models.length, agentKey, setValue]);

    return (
        <div className="col-span-full p-px">
            <div className="mt-6 flex flex-col gap-4">
                <h4 className="text-sm font-medium">Reasoning Configuration</h4>
                <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                    {(supportsAdaptive || canDisable) && (
                        <FormField
                            control={control}
                            name={`agents.${agentKey}.reasoning.mode`}
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>Reasoning Mode</FormLabel>
                                    <Select
                                        disabled={isLoading || (isAdaptiveOnly && !canDisable)}
                                        onValueChange={(value) => field.onChange(value !== 'none' ? value : null)}
                                        value={field.value ?? (isAdaptiveOnly ? ReasoningMode.Adaptive : 'none')}
                                    >
                                        <FormControl>
                                            <SelectTrigger>
                                                <SelectValue placeholder="Select reasoning mode" />
                                            </SelectTrigger>
                                        </FormControl>
                                        <SelectContent>
                                            {!isAdaptiveOnly && <SelectItem value="none">Not selected</SelectItem>}
                                            {supportsAdaptive && (
                                                <SelectItem value={ReasoningMode.Adaptive}>Adaptive</SelectItem>
                                            )}
                                            {!isAdaptiveOnly && (
                                                <SelectItem value={ReasoningMode.Budget}>Budget</SelectItem>
                                            )}
                                            {canDisable && (
                                                <SelectItem value={ReasoningMode.Off}>Off (no thinking)</SelectItem>
                                            )}
                                        </SelectContent>
                                    </Select>
                                    <FormDescription>
                                        {isAdaptiveOnly
                                            ? canDisable
                                                ? 'This model thinks adaptively; choose Off to disable thinking.'
                                                : 'This model supports only adaptive thinking and cannot be disabled.'
                                            : 'Adaptive lets the model decide how much to think; budget uses a fixed token budget; off disables thinking.'}
                                    </FormDescription>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />
                    )}

                    <FormField
                        control={control}
                        name={`agents.${agentKey}.reasoning.effort`}
                        render={({ field }) => (
                            <FormItem>
                                <FormLabel>Reasoning Effort</FormLabel>
                                <Select
                                    disabled={isLoading || isOff}
                                    onValueChange={(value) => {
                                        const next = value !== 'none' ? value : null;
                                        field.onChange(next);

                                        // max/xhigh are adaptive-thinking effort levels; selecting one
                                        // implies adaptive mode so the backend doesn't drop the reasoning.
                                        if (
                                            supportsAdaptive &&
                                            (next === ReasoningEffort.Xhigh || next === ReasoningEffort.Max)
                                        ) {
                                            setValue(
                                                `agents.${agentKey}.reasoning.mode` as const,
                                                ReasoningMode.Adaptive,
                                            );
                                        }
                                    }}
                                    value={field.value ?? 'none'}
                                >
                                    <FormControl>
                                        <SelectTrigger>
                                            <SelectValue placeholder="Select effort level (optional)" />
                                        </SelectTrigger>
                                    </FormControl>
                                    <SelectContent>
                                        <SelectItem value="none">Not selected</SelectItem>
                                        {allowedEfforts.map((effort) => (
                                            <SelectItem
                                                key={effort}
                                                value={effort}
                                            >
                                                {reasoningEffortLabel[effort]}
                                            </SelectItem>
                                        ))}
                                    </SelectContent>
                                </Select>
                                <FormMessage />
                            </FormItem>
                        )}
                    />

                    <FormInputNumberItem
                        control={control}
                        disabled={isLoading || isOff}
                        label="Reasoning Max Tokens"
                        min="1"
                        name={`agents.${agentKey}.reasoning.maxTokens`}
                        placeholder="1000"
                        valueType="integer"
                    />
                </div>
            </div>
        </div>
    );
}

const transformFormToGraphQL = (
    formData: FormInput,
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
                model: data?.model ?? '',
                presencePenalty: data?.presencePenalty ?? null,
                price:
                    data?.price &&
                    typeof data?.price.input === 'number' &&
                    typeof data?.price.output === 'number' &&
                    typeof data?.price.cacheRead === 'number' &&
                    typeof data?.price.cacheWrite === 'number'
                        ? {
                              cacheRead: data.price.cacheRead,
                              cacheWrite: data.price.cacheWrite,
                              input: data.price.input,
                              output: data.price.output,
                          }
                        : null,
                reasoning: data?.reasoning
                    ? {
                          effort: getReasoningEffort(data?.reasoning.effort),
                          maxTokens: data?.reasoning.maxTokens ?? null,
                          mode: getReasoningMode(data?.reasoning.mode),
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
        name: formData.name ?? '',
        type: z.nativeEnum(ProviderType).parse(formData.type),
    };
};

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
    results: null | ProviderTestResults;
}

function TestResultsDialog({ handleOpenChange, isOpen, results }: TestResultsDialogProps) {
    if (!results) {
        return null;
    }

    const agentResults = Object.entries(results)
        .filter(([key]) => key !== '__typename')
        .map(([agentType, agentData]) => ({
            agentType,
            tests: agentData?.tests || [],
        }));

    const getStatusIcon = (result: boolean | null | undefined) => {
        if (result === true) {
            return <CheckCircle className="size-4 shrink-0 text-green-500" />;
        }

        if (result === false) {
            return <XCircle className="size-4 shrink-0 text-red-500" />;
        }

        return <Clock className="size-4 shrink-0 text-yellow-500" />;
    };

    const getResultBadge = (result: boolean | null | undefined) => {
        if (result === true) {
            return (
                <Badge
                    className="shrink-0 border-green-500/40 bg-green-500/10 text-green-600"
                    variant="outline"
                >
                    Success
                </Badge>
            );
        }

        if (result === false) {
            return (
                <Badge
                    className="shrink-0"
                    variant="destructive"
                >
                    Failed
                </Badge>
            );
        }

        return (
            <Badge
                className="shrink-0"
                variant="secondary"
            >
                Unknown
            </Badge>
        );
    };

    return (
        <Dialog
            onOpenChange={handleOpenChange}
            open={isOpen}
        >
            <DialogContent className="flex max-h-[80vh] flex-col sm:max-w-3xl">
                <DialogHeader className="shrink-0">
                    <DialogTitle>Provider Test Results</DialogTitle>
                </DialogHeader>
                <div className="flex flex-1 flex-col overflow-y-auto">
                    <Accordion
                        className="w-full"
                        type="multiple"
                    >
                        {agentResults.map(({ agentType, tests }) => {
                            const testsCount = tests.length;
                            const successTestsCount = tests.filter((test) => test.result === true).length;
                            const isAllPassed = testsCount > 0 && successTestsCount === testsCount;
                            const isNonePassed = testsCount > 0 && successTestsCount === 0;

                            return (
                                <AccordionItem
                                    key={agentType}
                                    value={agentType}
                                >
                                    <AccordionTrigger className="group text-left hover:no-underline">
                                        <div className="mr-3 flex w-full items-center justify-between gap-3">
                                            <span className="font-semibold group-hover:underline">
                                                {getName(agentType)}
                                            </span>
                                            <Badge
                                                className={
                                                    isAllPassed
                                                        ? 'shrink-0 border-green-500/40 bg-green-500/10 text-green-600'
                                                        : 'shrink-0'
                                                }
                                                variant={
                                                    isNonePassed ? 'destructive' : isAllPassed ? 'outline' : 'secondary'
                                                }
                                            >
                                                {successTestsCount}/{testsCount} passed
                                            </Badge>
                                        </div>
                                    </AccordionTrigger>
                                    <AccordionContent>
                                        <div className="flex flex-col gap-2 pt-1">
                                            {tests.map((test, index) => (
                                                <div
                                                    className="rounded-lg border p-3"
                                                    key={index}
                                                >
                                                    <div className="flex items-start justify-between gap-2">
                                                        <div className="flex min-w-0 items-start gap-2">
                                                            <span className="mt-0.5">{getStatusIcon(test.result)}</span>
                                                            <div className="min-w-0">
                                                                <div className="font-medium break-words">
                                                                    {test.name}
                                                                </div>
                                                                {test.type && (
                                                                    <div className="text-muted-foreground text-xs break-words">
                                                                        {test.type}
                                                                    </div>
                                                                )}
                                                            </div>
                                                        </div>
                                                        {getResultBadge(test.result)}
                                                    </div>
                                                    {(test.reasoning !== undefined ||
                                                        test.streaming !== undefined ||
                                                        Boolean(test.latency)) && (
                                                        <div className="mt-2 flex flex-wrap gap-1.5">
                                                            {test.reasoning !== undefined && (
                                                                <Badge variant="outline">
                                                                    Reasoning: {test.reasoning ? 'Yes' : 'No'}
                                                                </Badge>
                                                            )}
                                                            {test.streaming !== undefined && (
                                                                <Badge variant="outline">
                                                                    Streaming: {test.streaming ? 'Yes' : 'No'}
                                                                </Badge>
                                                            )}
                                                            {Boolean(test.latency) && (
                                                                <Badge variant="outline">{test.latency} ms</Badge>
                                                            )}
                                                        </div>
                                                    )}
                                                    {test.error && (
                                                        <div className="border-destructive/30 bg-destructive/5 mt-2 max-h-40 overflow-auto rounded-md border p-2">
                                                            <pre className="text-destructive font-mono text-xs break-words whitespace-pre-wrap">
                                                                {test.error}
                                                            </pre>
                                                        </div>
                                                    )}
                                                </div>
                                            ))}
                                            {tests.length === 0 && (
                                                <div className="text-muted-foreground py-4 text-center text-sm">
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
}

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

interface DeleteProviderDialogProps extends Pick<
    ComponentProps<typeof ConfirmationDialog>,
    'handleConfirm' | 'handleOpenChange' | 'isOpen'
> {
    control: Control<FormInput>;
}

// Don't hoist this useWatch to the parent — a name keystroke would re-render the whole form.
function DeleteProviderDialog({ control, handleConfirm, handleOpenChange, isOpen }: DeleteProviderDialogProps) {
    const providerName = useWatch({ control, name: 'name' });

    return (
        <ConfirmationDialog
            cancelText="Cancel"
            confirmText="Delete"
            handleConfirm={handleConfirm}
            handleOpenChange={handleOpenChange}
            isOpen={isOpen}
            itemName={providerName}
            itemType="provider"
        />
    );
}

function SettingsProvider() {
    const { providerId } = useParams<{ providerId: string }>();
    const navigate = useNavigate();
    const [searchParams, setSearchParams] = useSearchParams();
    const { data, error, loading } = useQuery(SettingsProvidersDocument);
    const [createProvider, { loading: isCreateLoading }] = useMutation(CreateProviderDocument);
    const [updateProvider, { loading: isUpdateLoading }] = useMutation(UpdateProviderDocument);
    const [deleteProvider, { loading: isDeleteLoading }] = useMutation(DeleteProviderDocument);
    const [testProvider, { loading: isTestLoading }] = useMutation(TestProviderDocument);
    const [testAgent, { loading: isAgentTestLoading }] = useMutation(TestAgentDocument);
    const [currentAgentKey, setCurrentAgentKey] = useState<null | string>(null);
    const [submitError, setSubmitError] = useState<null | string>(null);
    const [isTestDialogOpen, setIsTestDialogOpen] = useState(false);
    const [testResults, setTestResults] = useState<null | ProviderTestResults>(null);
    const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);

    const isNew = providerId === 'new';
    const isLoading = isCreateLoading || isUpdateLoading || isDeleteLoading;
    const { isDesktop } = useBreakpoint();

    const form = useAppForm<FormInput, unknown, FormData>({
        defaultValues: {
            agents: {},
            name: undefined,
            type: undefined,
        },
        schema: formSchema,
    });

    const { control, formState, handleSubmit: handleFormSubmit, reset, setValue, trigger, watch } = form;

    const { isDirty } = useFormState({ control });

    useEffect(() => {
        if (submitError) {
            toast.error(submitError);
        }
    }, [submitError]);

    const selectedType = useWatch({ control, name: 'type' });

    const formQueryParams = useMemo(
        () => ({
            id: searchParams.get('id'),
            type: searchParams.get('type'),
        }),
        [searchParams],
    );

    const getAgentTypes = () => {
        const agentsSource =
            (isNew &&
                selectedType &&
                data?.settingsProviders?.default?.[selectedType as keyof typeof data.settingsProviders.default]
                    ?.agents) ||
            (!isNew &&
                providerId &&
                data?.settingsProviders?.userDefined?.find((p: Provider) => p.id == providerId)?.agents) ||
            (data?.settingsProviders?.default &&
                Object.values(data.settingsProviders.default).find((provider) => provider?.agents)?.agents) ||
            null;

        return extractAgentTypes(agentsSource) ?? Object.keys(agentTypesMap);
    };

    const agentTypes = getAgentTypes();

    const availableModels = useMemo(() => {
        if (!data?.settingsProviders?.models || !selectedType) {
            return [];
        }

        const { models } = data.settingsProviders;
        const providerModels = models[selectedType as keyof typeof models];

        if (!providerModels?.length) {
            return [];
        }

        return providerModels
            .map((model) => ({
                name: model.name,
                price: model.price
                    ? {
                          cacheRead: model.price.cacheRead ?? 0,
                          cacheWrite: model.price.cacheWrite ?? 0,
                          input: model.price.input ?? 0,
                          output: model.price.output ?? 0,
                      }
                    : null,
                reasoning: model.reasoning ?? null,
                thinking: model.thinking,
            }))
            .filter((model) => model.name)
            .sort((a, b) => a.name.localeCompare(b.name));
    }, [data, selectedType]);

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
                        const agent = { ...data };

                        if (agent.model && !availableModels.find((m) => m.name === agent.model)) {
                            agent.model = availableModels[0]?.name || agent.model;
                        }

                        return [key, agent];
                    }),
            );

            setValue('agents', normalizeGraphQLData(agents) as FormAgents);
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [availableModels, data, isNew, selectedType]);

    useEffect(() => {
        if (!isNew) {
            if (searchParams.size > 0) {
                setSearchParams({});
            }

            return;
        }

        const queryId = searchParams.get('id');

        if (queryId) {
            return;
        }

        const queryType = searchParams.get('type');

        if (!selectedType && queryType) {
            return;
        }

        setSearchParams((prev) => {
            const params = new URLSearchParams(prev);

            if (selectedType) {
                params.set('type', selectedType);
            } else {
                params.delete('type');
            }

            return params;
        });
    }, [selectedType, setSearchParams, isNew, searchParams]);

    useEffect(() => {
        if (!data?.settingsProviders) {
            return;
        }

        const providers = data.settingsProviders;

        if (isNew || !providerId) {
            const queryType = formQueryParams.type ?? undefined;
            const queryId = formQueryParams.id;

            // A hand-typed ?type= URL bypasses the create menu's enabled-only filter; an
            // unknown or disabled type would otherwise create a dead provider or dump a raw
            // zod error on submit. Bounce it to the list. (Clone-by-id is gated separately below.)
            if (!queryId && queryType && !providers.enabled[queryType as keyof typeof providers.enabled]) {
                toast.error(`Provider type "${queryType}" is not available`);
                navigate(routes.settings.providers, { replace: true });

                return;
            }

            if (queryId && data?.settingsProviders?.userDefined) {
                const sourceProvider = data.settingsProviders.userDefined.find((p: Provider) => p.id == queryId);

                if (sourceProvider) {
                    const { agents, name, type: sourceType } = sourceProvider;

                    // Cloning a provider whose type is now disabled would only make
                    // another dead one — gate it the same as the ?type= path.
                    if (sourceType && !providers.enabled[sourceType as keyof typeof providers.enabled]) {
                        toast.error(`Provider type "${sourceType}" is not available`);
                        navigate(routes.settings.providers, { replace: true });

                        return;
                    }

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

            // Bail out of the empty-form reset when `selectedType` is set — the agent-filling
            // effect above is the source of truth in that case and would fight us.
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
            navigate(routes.settings.providers);

            return;
        }

        const { agents, name, type } = provider;

        reset({
            agents: agents ? (normalizeGraphQLData(agents) as FormAgents) : {},
            name: name || undefined,
            type: type || undefined,
        });
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [data, formQueryParams, isNew, providerId, selectedType]);

    const performSave = async (): Promise<boolean> => {
        // watch() — not getValues() — because disabled fields must be included in the payload.
        const formData = watch();

        try {
            setSubmitError(null);

            const mutationData = transformFormToGraphQL(formData);

            if (isNew) {
                await createProvider({
                    refetchQueries: ['settingsProviders'],
                    variables: mutationData,
                });
            } else {
                await updateProvider({
                    refetchQueries: ['settingsProviders'],
                    variables: {
                        ...mutationData,
                        providerId: providerId!,
                    },
                });
            }

            return true;
        } catch (error) {
            console.error('Submit error:', error);
            setSubmitError(error instanceof Error ? error.message : 'An error occurred while saving');

            return false;
        }
    };

    const onSaveAndLeave = async (): Promise<boolean> => {
        setSubmitError(null);
        const valid = await trigger();

        if (!valid) {
            setSubmitError(
                `Please fix the following validation errors:\n\n${formatFormErrors(formState.errors as Record<string, unknown>)}`,
            );

            return false;
        }

        return performSave();
    };

    // Validity is only needed to gate the unsaved-changes dialog; subscribing to formState.isValid
    // would make RHF re-run the whole zod schema on every keystroke. Validate lazily when the dialog opens.
    const [isFormValid, setIsFormValid] = useState(true);

    const unsavedGuard = useUnsavedChangesGuard({
        isDirty,
        isFormValid,
        onSave: onSaveAndLeave,
    });

    useEffect(() => {
        if (!unsavedGuard.isOpen) {
            return;
        }

        let cancelled = false;

        void (async () => {
            const valid = await trigger();

            if (!cancelled) {
                setIsFormValid(valid);
            }
        })();

        return () => {
            cancelled = true;
        };
    }, [unsavedGuard.isOpen, trigger]);

    const handleSubmit = async () => {
        const saved = await performSave();

        if (saved) {
            unsavedGuard.skipNextBlock();
            navigate(routes.settings.providers);
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

            navigate(routes.settings.providers);
        } catch (error) {
            console.error('Delete error:', error);
            setSubmitError(error instanceof Error ? error.message : 'An error occurred while deleting');
        }
    };

    const handleTest = async () => {
        setSubmitError(null);
        const isValid = await trigger();

        if (!isValid) {
            setSubmitError(
                `Please fix the following validation errors:\n\n${formatFormErrors(formState.errors as Record<string, unknown>)}`,
            );

            return;
        }

        try {
            setSubmitError(null);

            const formData = watch();
            const { agents, type } = transformFormToGraphQL(formData);
            const result = await testProvider({
                variables: {
                    agents,
                    type,
                },
            });

            setTestResults((result.data?.testProvider ?? null) as null | ProviderTestResults);
            setIsTestDialogOpen(true);
        } catch (error) {
            console.error('Test error:', error);
            setSubmitError(error instanceof Error ? error.message : 'An error occurred while testing');
        }
    };

    const handleTestAgent = async (agentKey: string) => {
        setSubmitError(null);
        const isValid = await trigger();

        if (!isValid) {
            setSubmitError(
                `Please fix the following validation errors:\n\n${formatFormErrors(formState.errors as Record<string, unknown>)}`,
            );

            return;
        }

        try {
            setSubmitError(null);
            setCurrentAgentKey(agentKey);
            // watch() — not getValues() — because disabled fields must be included in the payload.
            const formData = watch();
            const { agents, type } = transformFormToGraphQL(formData);

            const agent = agents[agentKey as keyof AgentsConfigInput] as AgentConfigInput;

            const singleResult = await testAgent({
                variables: { agent, agentType: agentTypesMap[agentKey] ?? AgentConfigType.Simple, type },
            });
            setTestResults({ [agentKey]: singleResult.data?.testAgent } as ProviderTestResults);
            setIsTestDialogOpen(true);
            setCurrentAgentKey(null);

            return;
        } catch (error) {
            console.error('Test error:', error);
            setSubmitError(error instanceof Error ? error.message : 'An error occurred while testing');
            setCurrentAgentKey(null);
        }
    };

    if (loading) {
        return (
            <>
                <AppHeader>
                    <AppHeaderContent>
                        <AppHeaderTitle icon={<Plug className="size-4 shrink-0" />}>
                            {isNew ? 'Create Provider' : 'Edit Provider'}
                        </AppHeaderTitle>
                    </AppHeaderContent>
                </AppHeader>
                <div className="flex flex-1 items-center justify-center p-4">
                    <StatusCard
                        description="Please wait while we fetch provider configuration"
                        icon={<Loader2 className="text-muted-foreground size-16 animate-spin" />}
                        title="Loading provider data..."
                    />
                </div>
            </>
        );
    }

    if (error) {
        return (
            <>
                <AppHeader>
                    <AppHeaderContent>
                        <AppHeaderTitle icon={<Plug className="size-4 shrink-0" />}>
                            {isNew ? 'Create Provider' : 'Edit Provider'}
                        </AppHeaderTitle>
                    </AppHeaderContent>
                </AppHeader>
                <div className="flex flex-1 items-center justify-center p-4">
                    <StatusCard
                        description={error.message}
                        icon={<AlertCircle className="text-destructive size-16" />}
                        title="Error loading provider data"
                    />
                </div>
            </>
        );
    }

    const providers = data?.settingsProviders?.models
        ? Object.keys(data?.settingsProviders.models).filter((key) => key !== '__typename')
        : [];

    const metaFields = (
        <div className="flex flex-col gap-6">
            <div className="flex flex-col gap-2 text-center">
                <h2 className="text-2xl font-semibold">{isNew ? 'Create a new provider' : 'Edit provider'}</h2>
                <p className="text-muted-foreground">
                    {isNew ? 'Configure a new language model provider' : 'Update provider settings and configuration'}
                </p>
            </div>

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
        </div>
    );

    const agentConfigs = (
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
                                    'hover:bg-accent hover:text-accent-foreground mr-2 flex items-center gap-1 rounded border px-2 py-1 text-xs',
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
                                <span className="no-underline! hover:no-underline!">
                                    {isAgentTestLoading && currentAgentKey === agentKey ? 'Testing...' : 'Test'}
                                </span>
                            </span>
                        </div>
                    </AccordionTrigger>
                    <AccordionContent className="flex flex-col gap-4 pt-4">
                        <div className="grid grid-cols-1 gap-4 p-px md:grid-cols-2">
                            <FormModelComboboxItem
                                control={control}
                                disabled={isLoading}
                                label="Model"
                                name={`agents.${agentKey}.model`}
                                onOptionSelect={(option) => {
                                    const price = option?.price;

                                    setValue(`agents.${agentKey}.price.input` as const, price?.input ?? null);
                                    setValue(`agents.${agentKey}.price.output` as const, price?.output ?? null);
                                    setValue(`agents.${agentKey}.price.cacheRead` as const, price?.cacheRead ?? null);
                                    setValue(`agents.${agentKey}.price.cacheWrite` as const, price?.cacheWrite ?? null);

                                    // Reset reasoning on model change: adaptive-only models lock
                                    // to adaptive, others clear the now-stale mode/effort/budget.
                                    setValue(
                                        `agents.${agentKey}.reasoning.mode` as const,
                                        option?.reasoning?.mode === ModelReasoningMode.AdaptiveOnly
                                            ? ReasoningMode.Adaptive
                                            : null,
                                    );
                                    setValue(`agents.${agentKey}.reasoning.effort` as const, null);
                                    setValue(`agents.${agentKey}.reasoning.maxTokens` as const, null);
                                }}
                                options={availableModels}
                                placeholder="Select or enter model name"
                            />

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

                            <FormInputNumberItem
                                control={control}
                                disabled={isLoading}
                                label="Max Tokens"
                                min="1"
                                name={`agents.${agentKey}.maxTokens`}
                                placeholder="1000"
                                valueType="integer"
                            />

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

                            <FormInputNumberItem
                                control={control}
                                disabled={isLoading}
                                label="Top K"
                                min="1"
                                name={`agents.${agentKey}.topK`}
                                placeholder="40"
                                valueType="integer"
                            />

                            <FormInputNumberItem
                                control={control}
                                disabled={isLoading}
                                label="Min Length"
                                min="0"
                                name={`agents.${agentKey}.minLength`}
                                placeholder="0"
                                valueType="integer"
                            />

                            <FormInputNumberItem
                                control={control}
                                disabled={isLoading}
                                label="Max Length"
                                min="1"
                                name={`agents.${agentKey}.maxLength`}
                                placeholder="2000"
                                valueType="integer"
                            />

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

                        <ReasoningFields
                            agentKey={agentKey}
                            control={control}
                            isLoading={isLoading}
                            models={availableModels}
                            setValue={setValue}
                        />

                        <div className="col-span-full p-px">
                            <div className="mt-6 flex flex-col gap-4">
                                <h4 className="text-sm font-medium">Price Configuration</h4>
                                <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                                    <FormInputNumberItem
                                        control={control}
                                        description="Price per 1M input tokens"
                                        disabled={isLoading}
                                        label="Input Price"
                                        min="0"
                                        name={`agents.${agentKey}.price.input`}
                                        placeholder="0.001"
                                        step="0.000001"
                                    />

                                    <FormInputNumberItem
                                        control={control}
                                        description="Price per 1M output tokens"
                                        disabled={isLoading}
                                        label="Output Price"
                                        min="0"
                                        name={`agents.${agentKey}.price.output`}
                                        placeholder="0.002"
                                        step="0.000001"
                                    />

                                    <FormInputNumberItem
                                        control={control}
                                        description="Price per 1M cached read tokens"
                                        disabled={isLoading}
                                        label="Cache Read Price"
                                        min="0"
                                        name={`agents.${agentKey}.price.cacheRead`}
                                        placeholder="0.0001"
                                        step="0.000001"
                                    />

                                    <FormInputNumberItem
                                        control={control}
                                        description="Price per 1M cache write tokens"
                                        disabled={isLoading}
                                        label="Cache Write Price"
                                        min="0"
                                        name={`agents.${agentKey}.price.cacheWrite`}
                                        placeholder="0.00015"
                                        step="0.000001"
                                    />
                                </div>
                            </div>
                        </div>
                    </AccordionContent>
                </AccordionItem>
            ))}
        </Accordion>
    );

    return (
        <div className={isDesktop ? 'flex h-[100dvh] min-h-0 flex-col' : 'flex min-h-[100dvh] flex-col'}>
            <AppHeader>
                <AppHeaderContent>
                    <AppHeaderTitle icon={<Plug className="size-4 shrink-0" />}>
                        {isNew ? 'Create Provider' : 'Edit Provider'}
                    </AppHeaderTitle>
                </AppHeaderContent>
                <AppHeaderActions>
                    <AppHeaderAction
                        disabled={isLoading || isTestLoading || isAgentTestLoading}
                        icon={isTestLoading ? <Loader2 className="size-4 animate-spin" /> : <Play className="size-4" />}
                        label={isTestLoading ? 'Testing...' : 'Test'}
                        onClick={() => handleTest()}
                        type="button"
                        variant="outline"
                    />
                    <FormSubmitButton
                        form="provider-form"
                        icon={<Save className="size-4" />}
                        loading={isLoading}
                        size="sm"
                        variant="secondary"
                    >
                        {isNew ? 'Create' : 'Save'}
                    </FormSubmitButton>
                    {!isNew && (
                        <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                                <Button
                                    aria-label="Provider actions"
                                    className="size-8 p-0"
                                    type="button"
                                    variant="ghost"
                                >
                                    <Ellipsis />
                                </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent
                                align="end"
                                className="min-w-24"
                            >
                                <DropdownMenuItem
                                    disabled={isDeleteLoading}
                                    onClick={handleDelete}
                                >
                                    {isDeleteLoading ? (
                                        <Loader2 className="size-4 animate-spin" />
                                    ) : (
                                        <Trash2 className="size-4" />
                                    )}
                                    {isDeleteLoading ? 'Deleting...' : 'Delete'}
                                </DropdownMenuItem>
                            </DropdownMenuContent>
                        </DropdownMenu>
                    )}
                </AppHeaderActions>
            </AppHeader>
            <Form {...form}>
                <form
                    className="flex min-h-0 flex-1 flex-col"
                    id="provider-form"
                    noValidate
                    onSubmit={handleFormSubmit(handleSubmit)}
                >
                    {isDesktop ? (
                        <DetailSplitLayout
                            content={agentConfigs}
                            contentClassName="h-full min-h-0 overflow-y-auto p-4"
                            panel={metaFields}
                        />
                    ) : (
                        <div className="flex min-h-0 flex-1 flex-col gap-4 p-4">
                            {metaFields}
                            {agentConfigs}
                        </div>
                    )}
                </form>
            </Form>

            <TestResultsDialog
                handleOpenChange={setIsTestDialogOpen}
                isOpen={isTestDialogOpen}
                results={testResults}
            />

            <DeleteProviderDialog
                control={control}
                handleConfirm={handleConfirmDelete}
                handleOpenChange={setIsDeleteDialogOpen}
                isOpen={isDeleteDialogOpen}
            />

            <UnsavedChangesDialog
                canSave={isFormValid}
                handleCancel={unsavedGuard.handleCancel}
                handleDiscard={unsavedGuard.handleDiscard}
                handleOpenChange={unsavedGuard.handleOpenChange}
                handleSaveAndLeave={unsavedGuard.handleSaveAndLeave}
                isOpen={unsavedGuard.isOpen}
                isSavingFromDialog={unsavedGuard.isSavingFromDialog}
            />
        </div>
    );
}

export default SettingsProvider;
