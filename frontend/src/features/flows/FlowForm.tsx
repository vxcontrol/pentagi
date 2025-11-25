import { zodResolver } from '@hookform/resolvers/zod';
import { ArrowUpIcon, Check, ChevronDown, Square, X } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';

import { ProviderIcon } from '@/components/icons/ProviderIcon';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuGroup,
    DropdownMenuItem,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Form, FormControl, FormField, FormLabel } from '@/components/ui/form';
import {
    InputGroup,
    InputGroupAddon,
    InputGroupButton,
    InputGroupInput,
    InputGroupTextareaAutosize,
} from '@/components/ui/input-group';
import { Spinner } from '@/components/ui/spinner';
import { Switch } from '@/components/ui/switch';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { getProviderDisplayName } from '@/models/Provider';
import { useProviders } from '@/providers/ProvidersProvider';

const formSchema = z.object({
    message: z.string().trim().min(1, { message: 'Message cannot be empty' }),
    providerName: z.string().trim().min(1, { message: 'Provider must be selected' }),
    useAgents: z.boolean(),
});

export interface FlowFormProps {
    defaultValues?: Partial<FlowFormValues>;
    isCanceling?: boolean;
    isDisabled?: boolean;
    isLoading?: boolean;
    isProviderDisabled?: boolean;
    isSubmitting?: boolean;
    onCancel?: () => Promise<void> | void;
    onSubmit: (values: FlowFormValues) => Promise<void> | void;
    placeholder?: string;
    type: 'assistant' | 'automation';
}

export type FlowFormValues = z.infer<typeof formSchema>;

export const FlowForm = ({
    defaultValues,
    isCanceling,
    isDisabled,
    isLoading,
    isProviderDisabled,
    isSubmitting,
    onCancel,
    onSubmit,
    placeholder = 'Describe what you would like PentAGI to test...',
    type,
}: FlowFormProps) => {
    const { providers, setSelectedProvider } = useProviders();
    const [providerSearch, setProviderSearch] = useState('');

    const filteredProviders = useMemo(() => {
        if (!providerSearch.trim()) {
            return providers;
        }

        const searchLower = providerSearch.toLowerCase();

        return providers.filter((provider) => {
            const displayName = getProviderDisplayName(provider).toLowerCase();

            return displayName.includes(searchLower) || provider.name.toLowerCase().includes(searchLower);
        });
    }, [providers, providerSearch]);

    const form = useForm<FlowFormValues>({
        defaultValues: {
            message: defaultValues?.message ?? '',
            providerName: defaultValues?.providerName ?? '',
            useAgents: defaultValues?.useAgents ?? false,
        },
        mode: 'onChange',
        resolver: zodResolver(formSchema),
    });

    const {
        control,
        formState: { dirtyFields, isValid },
        getValues,
        handleSubmit: handleFormSubmit,
        resetField,
        setValue,
    } = form;

    // Update form values from defaultValues if user hasn't manually changed them
    useEffect(() => {
        if (!defaultValues) {
            return;
        }

        const currentValues = getValues();

        // Update only fields that user hasn't manually changed and that differ from current values
        Object.entries(defaultValues)
            .filter(([fieldName, defaultValue]) => {
                const typedFieldName = fieldName as keyof FlowFormValues;

                return (
                    defaultValue !== undefined &&
                    !dirtyFields[typedFieldName] &&
                    currentValues[typedFieldName] !== defaultValue
                );
            })
            .forEach(([fieldName, defaultValue]) => {
                const typedFieldName = fieldName as keyof FlowFormValues;
                setValue(typedFieldName, defaultValue as never, { shouldDirty: false });
            });
    }, [defaultValues, dirtyFields, setValue, getValues]);

    const isFormDisabled = isDisabled || isLoading || isSubmitting || isCanceling;

    const handleSubmit = async (values: FlowFormValues) => {
        await onSubmit(values);
        resetField('message');
    };

    const handleKeyDown = ({
        ctrlKey,
        key,
        metaKey,
        preventDefault,
        shiftKey,
    }: React.KeyboardEvent<HTMLTextAreaElement>) => {
        if (isFormDisabled || key !== 'Enter' || shiftKey || ctrlKey || metaKey) {
            return;
        }

        preventDefault();
        handleFormSubmit(handleSubmit)();
    };

    return (
        <Form {...form}>
            <form onSubmit={handleFormSubmit(handleSubmit)}>
                <FormField
                    control={control}
                    name="message"
                    render={({ field }) => (
                        <FormControl>
                            <InputGroup className="block">
                                <InputGroupTextareaAutosize
                                    {...field}
                                    autoFocus
                                    className="min-h-0"
                                    disabled={isFormDisabled}
                                    maxRows={9}
                                    minRows={1}
                                    onKeyDown={handleKeyDown}
                                    placeholder={placeholder}
                                />
                                <InputGroupAddon align="block-end">
                                    <FormField
                                        control={control}
                                        name="providerName"
                                        render={({ field: providerField }) => {
                                            const currentProvider = providers.find(
                                                (p) => p.name === providerField.value,
                                            );

                                            return (
                                                <DropdownMenu>
                                                    <DropdownMenuTrigger asChild>
                                                        <InputGroupButton
                                                            disabled={isFormDisabled || isProviderDisabled}
                                                            variant="ghost"
                                                        >
                                                            {currentProvider && (
                                                                <ProviderIcon provider={currentProvider} />
                                                            )}
                                                            <span className="max-w-40 truncate">
                                                                {currentProvider
                                                                    ? getProviderDisplayName(currentProvider)
                                                                    : 'Select Provider'}
                                                            </span>
                                                            <ChevronDown />
                                                        </InputGroupButton>
                                                    </DropdownMenuTrigger>
                                                    <DropdownMenuContent
                                                        align="start"
                                                        side="top"
                                                    >
                                                        <DropdownMenuGroup className="-m-1 rounded-none p-0">
                                                            <InputGroup className="-mb-1 rounded-none border-0 shadow-none [&:has([data-slot=input-group-control]:focus-visible)]:border-0 [&:has([data-slot=input-group-control]:focus-visible)]:ring-0">
                                                                <InputGroupInput
                                                                    onChange={(event) =>
                                                                        setProviderSearch(event.target.value)
                                                                    }
                                                                    onClick={(event) => event.stopPropagation()}
                                                                    onKeyDown={(event) => event.stopPropagation()}
                                                                    placeholder="Search..."
                                                                    value={providerSearch}
                                                                />
                                                                {providerSearch && (
                                                                    <InputGroupAddon align="inline-end">
                                                                        <InputGroupButton
                                                                            onClick={(event) => {
                                                                                event.stopPropagation();
                                                                                setProviderSearch('');
                                                                            }}
                                                                        >
                                                                            <X />
                                                                        </InputGroupButton>
                                                                    </InputGroupAddon>
                                                                )}
                                                            </InputGroup>
                                                            <DropdownMenuSeparator />
                                                        </DropdownMenuGroup>
                                                        <DropdownMenuGroup className="max-h-64 overflow-y-auto">
                                                            {!filteredProviders.length ? (
                                                                <DropdownMenuItem
                                                                    className="min-h-16 justify-center"
                                                                    disabled
                                                                >
                                                                    {providerSearch
                                                                        ? 'No results found'
                                                                        : 'No available providers'}
                                                                </DropdownMenuItem>
                                                            ) : (
                                                                filteredProviders.map((provider) => (
                                                                    <DropdownMenuItem
                                                                        key={provider.name}
                                                                        onSelect={() => {
                                                                            if (isFormDisabled || isProviderDisabled) {
                                                                                return;
                                                                            }

                                                                            providerField.onChange(provider.name);
                                                                            setSelectedProvider(provider);
                                                                            setProviderSearch('');
                                                                        }}
                                                                    >
                                                                        <div className="flex w-full min-w-0 items-center gap-2">
                                                                            <ProviderIcon
                                                                                className="size-4 shrink-0"
                                                                                provider={provider}
                                                                            />

                                                                            <span className="flex-1 truncate">
                                                                                {getProviderDisplayName(provider)}
                                                                            </span>
                                                                            {providerField.value === provider.name && (
                                                                                <Check className="ml-auto size-4 shrink-0" />
                                                                            )}
                                                                        </div>
                                                                    </DropdownMenuItem>
                                                                ))
                                                            )}
                                                        </DropdownMenuGroup>
                                                    </DropdownMenuContent>
                                                </DropdownMenu>
                                            );
                                        }}
                                    />
                                    {type === 'assistant' && (
                                        <FormField
                                            control={control}
                                            name="useAgents"
                                            render={({ field: useAgentsField }) => (
                                                <TooltipProvider>
                                                    <Tooltip>
                                                        <TooltipTrigger asChild>
                                                            <div className="flex items-center">
                                                                <FormControl>
                                                                    <Switch
                                                                        checked={useAgentsField.value}
                                                                        disabled={isFormDisabled}
                                                                        onCheckedChange={useAgentsField.onChange}
                                                                    />
                                                                </FormControl>
                                                                <FormLabel
                                                                    className="flex cursor-pointer pl-2 text-xs font-normal"
                                                                    onClick={() =>
                                                                        useAgentsField.onChange(!useAgentsField.value)
                                                                    }
                                                                >
                                                                    Use Agents
                                                                </FormLabel>
                                                            </div>
                                                        </TooltipTrigger>
                                                        <TooltipContent>
                                                            <p className="max-w-48">
                                                                Enable multi-agent collaboration for complex tasks
                                                            </p>
                                                        </TooltipContent>
                                                    </Tooltip>
                                                </TooltipProvider>
                                            )}
                                        />
                                    )}
                                    {!isLoading || isSubmitting ? (
                                        <InputGroupButton
                                            className="ml-auto"
                                            disabled={isSubmitting || !isValid}
                                            size="icon-xs"
                                            type="submit"
                                            variant="default"
                                        >
                                            {isSubmitting ? <Spinner variant="circle" /> : <ArrowUpIcon />}
                                        </InputGroupButton>
                                    ) : (
                                        <InputGroupButton
                                            className="ml-auto"
                                            disabled={isCanceling || !onCancel}
                                            onClick={() => onCancel?.()}
                                            size="icon-xs"
                                            type="button"
                                            variant="destructive"
                                        >
                                            {isCanceling ? <Spinner variant="circle" /> : <Square />}
                                        </InputGroupButton>
                                    )}
                                </InputGroupAddon>
                            </InputGroup>
                        </FormControl>
                    )}
                />
            </form>
        </Form>
    );
};
