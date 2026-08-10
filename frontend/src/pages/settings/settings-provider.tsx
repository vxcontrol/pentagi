import { useMutation, useQuery } from '@apollo/client/react';
import {
    Check,
    CheckCircle,
    ChevronDown,
    Clock,
    Ellipsis,
    Lightbulb,
    Play,
    Plug,
    Save,
    Trash2,
    XCircle,
} from 'lucide-react';
import { type ComponentProps, type FormEvent, useEffect, useMemo, useRef, useState } from 'react';
import {
    type Control,
    type FieldErrors,
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
import { ErrorState } from '@/components/shared/error-state';
import { LoadingState } from '@/components/shared/loading-state';
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
import { Input } from '@/components/ui/input';
import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupInput } from '@/components/ui/input-group';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Spinner } from '@/components/ui/spinner';
import { Textarea } from '@/components/ui/textarea';
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
import { useI18n } from '@/hooks/use-i18n';
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
    const { t } = useI18n();
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
                            className={cn(
                                'h-9 w-full justify-between bg-transparent px-3 font-normal hover:bg-transparent',
                                !displayValue && 'text-muted-foreground',
                            )}
                            disabled={disabled}
                            variant="outline"
                        >
                            {displayValue || placeholder}
                            <ChevronDown className="size-4 opacity-50" />
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
                                placeholder={t('providers.searchLabel', { label: label.toLowerCase() })}
                                value={search}
                            />
                            <CommandList>
                                <CommandEmpty>
                                    <div className="py-2 text-center">
                                        <p className="text-muted-foreground text-sm">
                                            {t('providers.noValueFound', { label: label.toLowerCase() })}
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
                                                {t('providers.useCustomValue', {
                                                    label: label.toLowerCase(),
                                                    search,
                                                })}
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
    const { t } = useI18n();
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
            return t('providers.free');
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
                                    aria-label={t('providers.openLabelList', { label: label.toLowerCase() })}
                                    disabled={disabled}
                                    size="icon-sm"
                                >
                                    <ChevronDown className="size-4 opacity-50" />
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
                                placeholder={t('providers.searchLabel', { label: label.toLowerCase() })}
                                value={search}
                            />
                            <CommandList>
                                <CommandEmpty>
                                    <div className="py-2 text-center">
                                        <p className="text-muted-foreground text-sm">
                                            {t('providers.noValueFound', { label: label.toLowerCase() })}
                                        </p>
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
                                                {t('providers.useCustomValue', {
                                                    label: label.toLowerCase(),
                                                    search,
                                                })}
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

function FormTextareaItem<T extends FieldValues = FieldValues>({
    control,
    description,
    disabled,
    label,
    name,
    placeholder,
}: FormInputStringItemProps<T>) {
    const { field, fieldState } = useController({ control, defaultValue: undefined, disabled, name });

    return (
        <FormItem>
            <FormLabel>{label}</FormLabel>
            <FormControl>
                <Textarea
                    {...field}
                    aria-invalid={fieldState.error ? true : undefined}
                    className="font-mono text-xs"
                    maxHeight={320}
                    minHeight={80}
                    placeholder={placeholder}
                    value={field.value ?? ''}
                />
            </FormControl>
            {description && <FormDescription>{description}</FormDescription>}
            {fieldState.error && <FormMessage>{fieldState.error.message}</FormMessage>}
        </FormItem>
    );
}

const optionalNumber = z.number().nullable().optional();

// Only a JSON object round-trips: extraBody merges into the request body as key/value pairs.
export const optionalJsonObject = z
    .string()
    .optional()
    .refine(
        (value) => {
            if (!value?.trim()) {
                return true;
            }

            try {
                const parsed: unknown = JSON.parse(value);

                return typeof parsed === 'object' && parsed !== null && !Array.isArray(parsed);
            } catch {
                return false;
            }
        },
        { message: 'Must be a valid JSON object' },
    );

const requiredString = (message: string) =>
    z
        .string()
        .optional()
        .transform((value) => value ?? '')
        .pipe(z.string().min(1, message));

// Built as a function of `t` (called from within the component) so validation messages
// follow the active language instead of being frozen in English at module load.
const buildAgentConfigSchema = (t: (key: string) => string) =>
    z
        .object({
            extraBody: optionalJsonObject,
            frequencyPenalty: optionalNumber,
            json: z.boolean().nullable().optional(),
            maxLength: optionalNumber,
            maxTokens: optionalNumber,
            minLength: optionalNumber,
            minP: optionalNumber,
            model: requiredString(t('providers.modelRequired')),
            n: optionalNumber,
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
            responseMimeType: z.string().nullable().optional(),
            temperature: optionalNumber,
            topK: optionalNumber,
            topP: optionalNumber,
        })
        .refine((data) => data.minLength == null || data.maxLength == null || data.minLength <= data.maxLength, {
            message: t('providers.minLengthExceed'),
            path: ['minLength'],
        })
        .refine((data) => data.reasoning?.maxTokens == null || data.reasoning.maxTokens <= 32000, {
            message: t('providers.maxReasoningTokens'),
            path: ['reasoning', 'maxTokens'],
        })
        .optional();

const buildFormSchema = (t: (key: string) => string) =>
    z.object({
        agents: z.record(z.string(), buildAgentConfigSchema(t)).optional(),
        name: requiredString(t('providers.providerNameRequired')).pipe(
            z.string().max(50, t('providers.providerNameMax')),
        ),
        type: requiredString(t('providers.providerTypeRequired')),
    });

const formSchema = buildFormSchema((key: string) => key);

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

const getReasoningEffortLabel = (t: (key: string) => string, effort: ReasoningEffort): string => {
    const labels: Record<ReasoningEffort, string> = {
        [ReasoningEffort.High]: t('providers.reasoningHigh'),
        [ReasoningEffort.Low]: t('providers.reasoningLow'),
        [ReasoningEffort.Max]: t('providers.reasoningMax'),
        [ReasoningEffort.Medium]: t('providers.reasoningMedium'),
        [ReasoningEffort.Xhigh]: t('providers.reasoningExtraHigh'),
    };

    return labels[effort];
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
    const { t } = useI18n();
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
                <h4 className="text-sm font-medium">{t('providers.reasoningConfig')}</h4>
                <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                    {(supportsAdaptive || canDisable) && (
                        <FormField
                            control={control}
                            name={`agents.${agentKey}.reasoning.mode`}
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>{t('providers.reasoningMode')}</FormLabel>
                                    <Select
                                        disabled={isLoading || (isAdaptiveOnly && !canDisable)}
                                        onValueChange={(value) => field.onChange(value !== 'none' ? value : null)}
                                        value={field.value ?? (isAdaptiveOnly ? ReasoningMode.Adaptive : 'none')}
                                    >
                                        <FormControl>
                                            <SelectTrigger>
                                                <SelectValue placeholder={t('providers.selectReasoningMode')} />
                                            </SelectTrigger>
                                        </FormControl>
                                        <SelectContent>
                                            {!isAdaptiveOnly && (
                                                <SelectItem value="none">{t('providers.notSelected')}</SelectItem>
                                            )}
                                            {supportsAdaptive && (
                                                <SelectItem value={ReasoningMode.Adaptive}>
                                                    {t('providers.adaptive')}
                                                </SelectItem>
                                            )}
                                            {!isAdaptiveOnly && (
                                                <SelectItem value={ReasoningMode.Budget}>
                                                    {t('providers.budget')}
                                                </SelectItem>
                                            )}
                                            {canDisable && (
                                                <SelectItem value={ReasoningMode.Off}>{t('providers.off')}</SelectItem>
                                            )}
                                        </SelectContent>
                                    </Select>
                                    <FormDescription>
                                        {isAdaptiveOnly
                                            ? canDisable
                                                ? t('providers.adaptiveThinkingDesc')
                                                : t('providers.adaptiveOnlyDesc')
                                            : t('providers.reasoningModeDesc')}
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
                                <FormLabel>{t('providers.reasoningEffort')}</FormLabel>
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
                                            <SelectValue placeholder={t('providers.selectEffortLevel')} />
                                        </SelectTrigger>
                                    </FormControl>
                                    <SelectContent>
                                        <SelectItem value="none">{t('providers.notSelected')}</SelectItem>
                                        {allowedEfforts.map((effort) => (
                                            <SelectItem
                                                key={effort}
                                                value={effort}
                                            >
                                                {getReasoningEffortLabel(t, effort)}
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
                        label={t('providers.reasoningMaxTokens')}
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

export const transformFormToGraphQL = (
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
                extraBody: data?.extraBody?.trim() ? (JSON.parse(data.extraBody) as Record<string, unknown>) : null,
                frequencyPenalty: data?.frequencyPenalty ?? null,
                // Not user-editable: carried through so saving a provider does not strip what the
                // shipped defaults set (json drives WithJSONMode on the simple_json agent).
                json: data?.json ?? null,
                maxLength: data?.maxLength ?? null,
                maxTokens: data?.maxTokens ?? null,
                minLength: data?.minLength ?? null,
                minP: data?.minP ?? null,
                model: data?.model ?? '',
                n: data?.n ?? null,
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
                responseMimeType: data?.responseMimeType ?? null,
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

export const normalizeGraphQLData = (obj: unknown): unknown => {
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
                .map(([key, value]) => {
                    if (key === 'extraBody') {
                        return [key, value && typeof value === 'object' ? JSON.stringify(value, null, 4) : ''];
                    }

                    return [key, normalizeGraphQLData(value)];
                }),
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
    const { t } = useI18n();

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
                    className="shrink-0"
                    variant="green"
                >
                    {t('providers.testResultSuccess')}
                </Badge>
            );
        }

        if (result === false) {
            return (
                <Badge
                    className="shrink-0"
                    variant="destructive"
                >
                    {t('providers.testResultFailed')}
                </Badge>
            );
        }

        return (
            <Badge
                className="shrink-0"
                variant="secondary"
            >
                {t('providers.testResultUnknown')}
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
                    <DialogTitle>{t('providers.testResults')}</DialogTitle>
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
                                                className="shrink-0"
                                                variant={
                                                    isNonePassed ? 'destructive' : isAllPassed ? 'green' : 'secondary'
                                                }
                                            >
                                                {successTestsCount}/{testsCount} {t('providers.passed')}
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
                                                                    {t('providers.reasoning')}:{' '}
                                                                    {test.reasoning ? t('common.yes') : t('common.no')}
                                                                </Badge>
                                                            )}
                                                            {test.streaming !== undefined && (
                                                                <Badge variant="outline">
                                                                    {t('providers.streaming')}:{' '}
                                                                    {test.streaming ? t('common.yes') : t('common.no')}
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
                                                    {t('providers.noTestsAvailable')}
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
    const { t } = useI18n();
    const providerName = useWatch({ control, name: 'name' });

    return (
        <ConfirmationDialog
            cancelText={t('common.cancel')}
            confirmText={t('common.delete')}
            handleConfirm={handleConfirm}
            handleOpenChange={handleOpenChange}
            isOpen={isOpen}
            itemName={providerName}
            itemType="provider"
        />
    );
}

function SettingsProvider() {
    const { t } = useI18n();
    const { providerId } = useParams<{ providerId: string }>();
    const navigate = useNavigate();
    const [searchParams, setSearchParams] = useSearchParams();
    const { data, error, loading, refetch } = useQuery(SettingsProvidersDocument);
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

    const localizedFormSchema = useMemo(() => buildFormSchema(t), [t]);

    const form = useAppForm<FormInput, unknown, FormData>({
        defaultValues: {
            agents: {},
            name: undefined,
            type: undefined,
        },
        // The seeding effect re-runs on every settingsProviders refetch (cache-and-network +
        // replaceWithIncoming gives a fresh `data`); without this a background refetch landing
        // mid-edit silently wipes the unsaved form.
        resetOptions: { keepDirtyValues: true },
        schema: localizedFormSchema,
    });

    const { control, formState, handleSubmit: handleFormSubmit, reset, setValue, trigger, watch } = form;

    const { isDirty } = useFormState({ control });
    const seededTypeRef = useRef<null | string>(null);

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

        // setValue is outside the form's keepDirtyValues, so a background refetch re-runs this effect
        // with a fresh `data` identity and overwrites agent edits the user has not saved. Re-seed only
        // when the type actually changed — that is the case where the previous type's agents are wrong.
        const isSameType = seededTypeRef.current === selectedType;

        seededTypeRef.current = selectedType;

        if (isSameType && form.getFieldState('agents').isDirty) {
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
                toast.error(t('providers.providerTypeNotAvailable', { type: queryType }));
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
                        toast.error(t('providers.providerTypeNotAvailable', { type: sourceType }));
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
            setSubmitError(error instanceof Error ? error.message : t('providers.saveFailed'));

            return false;
        }
    };

    const onSaveAndLeave = async (): Promise<boolean> => {
        setSubmitError(null);
        const valid = await trigger();

        if (!valid) {
            setSubmitError(
                t('providers.validationErrors', {
                    errors: formatFormErrors(formState.errors as Record<string, unknown>),
                }),
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

    const handleInvalidSubmit = (errors: FieldErrors<FormInput>) => {
        setSubmitError(
            t('providers.validationErrors', { errors: formatFormErrors(errors as Record<string, unknown>) }),
        );
    };

    const handleFormEvent = async (event: FormEvent<HTMLFormElement>) => {
        setSubmitError(null);

        await handleFormSubmit(handleSubmit, handleInvalidSubmit)(event);
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
            setSubmitError(error instanceof Error ? error.message : t('providers.deleteFailed2'));
        }
    };

    const handleTest = async () => {
        setSubmitError(null);
        const isValid = await trigger();

        if (!isValid) {
            setSubmitError(
                t('providers.validationErrors', {
                    errors: formatFormErrors(formState.errors as Record<string, unknown>),
                }),
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
            setSubmitError(error instanceof Error ? error.message : t('providers.testFailed'));
        }
    };

    const handleTestAgent = async (agentKey: string) => {
        setSubmitError(null);
        const isValid = await trigger();

        if (!isValid) {
            setSubmitError(
                t('providers.validationErrors', {
                    errors: formatFormErrors(formState.errors as Record<string, unknown>),
                }),
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
            setSubmitError(error instanceof Error ? error.message : t('providers.testFailed'));
            setCurrentAgentKey(null);
        }
    };

    if (loading && !data) {
        return (
            <>
                <AppHeader>
                    <AppHeaderContent>
                        <AppHeaderTitle icon={<Plug className="size-4 shrink-0" />}>
                            {isNew ? t('providers.createProvider') : t('providers.editProvider')}
                        </AppHeaderTitle>
                    </AppHeaderContent>
                </AppHeader>
                <div className="flex flex-1 items-center justify-center p-4">
                    <LoadingState
                        description={t('providers.loadingProviderDataDesc')}
                        title={t('providers.loadingProviderData')}
                    />
                </div>
            </>
        );
    }

    if (error && !data) {
        return (
            <>
                <AppHeader>
                    <AppHeaderContent>
                        <AppHeaderTitle icon={<Plug className="size-4 shrink-0" />}>
                            {isNew ? t('providers.createProvider') : t('providers.editProvider')}
                        </AppHeaderTitle>
                    </AppHeaderContent>
                </AppHeader>
                <div className="flex flex-1 items-center justify-center p-4">
                    <ErrorState
                        message={error.message}
                        onRetry={refetch}
                        title={t('providers.errorLoadingProviderData')}
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
                <h2 className="text-2xl font-semibold">
                    {isNew ? t('providers.createNewProvider') : t('providers.editProviderTitle')}
                </h2>
                <p className="text-muted-foreground">
                    {isNew ? t('providers.configureNewProvider') : t('providers.updateProviderSettings')}
                </p>
            </div>

            <FormComboboxItem
                allowCustom={false}
                control={control}
                description={t('providers.providerTypeDesc')}
                disabled={isLoading || !!selectedType}
                label={t('providers.providerType')}
                name="type"
                options={providers}
                placeholder={t('providers.selectProvider')}
            />

            <FormInputStringItem
                control={control}
                description={t('providers.providerNameDesc')}
                disabled={isLoading}
                label={t('providers.providerName')}
                name="name"
                placeholder={t('providers.enterProviderName')}
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
                            <Button
                                asChild
                                className={cn(
                                    'mr-2',
                                    (isTestLoading || isAgentTestLoading) && 'pointer-events-none opacity-50',
                                )}
                                size="xs"
                                variant="outline"
                            >
                                <span
                                    onClick={(event) => {
                                        if (isTestLoading || isAgentTestLoading) {
                                            return;
                                        }

                                        event.stopPropagation();
                                        handleTestAgent(agentKey);
                                    }}
                                    onKeyDown={(event) => {
                                        if (event.key !== 'Enter' && event.key !== ' ') {
                                            return;
                                        }

                                        event.preventDefault();
                                        event.stopPropagation();

                                        if (isTestLoading || isAgentTestLoading) {
                                            return;
                                        }

                                        handleTestAgent(agentKey);
                                    }}
                                    role="button"
                                    tabIndex={isTestLoading || isAgentTestLoading ? -1 : 0}
                                >
                                    {isAgentTestLoading && currentAgentKey === agentKey ? (
                                        <Spinner variant="circle" />
                                    ) : (
                                        <Play />
                                    )}
                                    <span className="no-underline! hover:no-underline!">
                                        {isAgentTestLoading && currentAgentKey === agentKey ? t('providers.testing') : t('common.test')}
                                    </span>
                                </span>
                            </Button>
                        </div>
                    </AccordionTrigger>
                    <AccordionContent className="flex flex-col gap-4 pt-4">
                        <div className="grid grid-cols-1 gap-4 p-px md:grid-cols-2">
                            <FormModelComboboxItem
                                control={control}
                                disabled={isLoading}
                                label={t('providers.model')}
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
                                placeholder={t('providers.selectOrEnterModel')}
                            />

                            <FormInputNumberItem
                                control={control}
                                disabled={isLoading}
                                label={t('providers.temperature')}
                                max="2"
                                min="0"
                                name={`agents.${agentKey}.temperature`}
                                placeholder="0.7"
                                step="0.1"
                            />

                            <FormInputNumberItem
                                control={control}
                                disabled={isLoading}
                                label={t('providers.maxTokens')}
                                min="1"
                                name={`agents.${agentKey}.maxTokens`}
                                placeholder="1000"
                                valueType="integer"
                            />

                            <FormInputNumberItem
                                control={control}
                                disabled={isLoading}
                                label={t('providers.topP')}
                                max="1"
                                min="0"
                                name={`agents.${agentKey}.topP`}
                                placeholder="0.9"
                                step="0.01"
                            />

                            <FormInputNumberItem
                                control={control}
                                disabled={isLoading}
                                label={t('providers.topK')}
                                min="1"
                                name={`agents.${agentKey}.topK`}
                                placeholder="40"
                                valueType="integer"
                            />

                            <FormInputNumberItem
                                control={control}
                                disabled={isLoading}
                                label={t('providers.minLength')}
                                min="0"
                                name={`agents.${agentKey}.minLength`}
                                placeholder="0"
                                valueType="integer"
                            />

                            <FormInputNumberItem
                                control={control}
                                disabled={isLoading}
                                label={t('providers.maxLength')}
                                min="1"
                                name={`agents.${agentKey}.maxLength`}
                                placeholder="2000"
                                valueType="integer"
                            />

                            <FormInputNumberItem
                                control={control}
                                disabled={isLoading}
                                label={t('providers.repetitionPenalty')}
                                max="2"
                                min="0"
                                name={`agents.${agentKey}.repetitionPenalty`}
                                placeholder="1.0"
                                step="0.01"
                            />

                            <FormInputNumberItem
                                control={control}
                                disabled={isLoading}
                                label={t('providers.frequencyPenalty')}
                                max="2"
                                min="0"
                                name={`agents.${agentKey}.frequencyPenalty`}
                                placeholder="0.0"
                                step="0.01"
                            />

                            <FormInputNumberItem
                                control={control}
                                disabled={isLoading}
                                label={t('providers.presencePenalty')}
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
                                <h4 className="text-sm font-medium">{t('providers.priceConfig')}</h4>
                                <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                                    <FormInputNumberItem
                                        control={control}
                                        description={t('providers.inputPriceDesc')}
                                        disabled={isLoading}
                                        label={t('providers.inputPrice')}
                                        min="0"
                                        name={`agents.${agentKey}.price.input`}
                                        placeholder="0.001"
                                        step="0.000001"
                                    />

                                    <FormInputNumberItem
                                        control={control}
                                        description={t('providers.outputPriceDesc')}
                                        disabled={isLoading}
                                        label={t('providers.outputPrice')}
                                        min="0"
                                        name={`agents.${agentKey}.price.output`}
                                        placeholder="0.002"
                                        step="0.000001"
                                    />

                                    <FormInputNumberItem
                                        control={control}
                                        description={t('providers.cacheReadPriceDesc')}
                                        disabled={isLoading}
                                        label={t('providers.cacheReadPrice')}
                                        min="0"
                                        name={`agents.${agentKey}.price.cacheRead`}
                                        placeholder="0.0001"
                                        step="0.000001"
                                    />

                                    <FormInputNumberItem
                                        control={control}
                                        description={t('providers.cacheWritePriceDesc')}
                                        disabled={isLoading}
                                        label={t('providers.cacheWritePrice')}
                                        min="0"
                                        name={`agents.${agentKey}.price.cacheWrite`}
                                        placeholder="0.00015"
                                        step="0.000001"
                                    />
                                </div>
                            </div>
                        </div>

                        <div className="col-span-full p-px">
                            <div className="mt-6 flex flex-col gap-4">
                                <h4 className="text-sm font-medium">{t('providers.extraBodyTitle')}</h4>
                                <FormTextareaItem
                                    control={control}
                                    description={t('providers.extraBodyDesc')}
                                    disabled={isLoading}
                                    label={t('providers.extraBodyLabel')}
                                    name={`agents.${agentKey}.extraBody`}
                                    placeholder={'{\n    "chat_template_kwargs": { "enable_thinking": false }\n}'}
                                />
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
                        {isNew ? t('providers.createProvider') : t('providers.editProvider')}
                    </AppHeaderTitle>
                </AppHeaderContent>
                <AppHeaderActions>
                    <AppHeaderAction
                        disabled={isLoading || isTestLoading || isAgentTestLoading}
                        icon={isTestLoading ? <Spinner variant="circle" /> : <Play />}
                        label={isTestLoading ? t('providers.testing') : t('common.test')}
                        onClick={() => handleTest()}
                        type="button"
                        variant="outline"
                    />
                    <AppHeaderAction
                        form="provider-form"
                        icon={<Save />}
                        label={isNew ? t('common.create') : t('common.save')}
                        loading={isLoading}
                        type="submit"
                    />
                    {!isNew && (
                        <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                                <Button
                                    aria-label={t('providers.providerActions')}
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
                                    {isDeleteLoading ? <Spinner variant="circle" /> : <Trash2 />}
                                    {isDeleteLoading ? t('providers.deletingProvider') : t('common.delete')}
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
                    onSubmit={handleFormEvent}
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
