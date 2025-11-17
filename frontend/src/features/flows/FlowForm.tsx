import { zodResolver } from '@hookform/resolvers/zod';
import { ArrowUpIcon, Check, ChevronsUpDown, X } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';

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
import { getProviderDisplayName, getProviderIcon } from '@/models/Provider';
import { useProviders } from '@/providers/ProvidersProvider';

const formSchema = z.object({
    message: z.string().trim().min(1, { message: 'Message cannot be empty' }),
    providerName: z.string().trim().min(1, { message: 'Provider must be selected' }),
    useAgents: z.boolean(),
});

export interface FlowFormProps {
    defaultValues?: Partial<FlowFormValues>;
    flowType: 'assistant' | 'automation';
    handleSubmit: (values: FlowFormValues) => Promise<void> | void;
    isLoading?: boolean;
}

export type FlowFormValues = z.infer<typeof formSchema>;

export const FlowForm = ({ defaultValues, flowType, handleSubmit: onSubmit, isLoading = false }: FlowFormProps) => {
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

    const handleKeyDown = ({
        ctrlKey,
        key,
        metaKey,
        preventDefault,
        shiftKey,
    }: React.KeyboardEvent<HTMLTextAreaElement>) => {
        if (isLoading || key !== 'Enter' || shiftKey || ctrlKey || metaKey) {
            return;
        }

        preventDefault();
        handleFormSubmit(onSubmit)();
    };

    return (
        <Form {...form}>
            <form onSubmit={handleFormSubmit(onSubmit)}>
                <FormField
                    control={control}
                    name="message"
                    render={({ field }) => (
                        <FormControl>
                            <InputGroup className="block">
                                <InputGroupTextareaAutosize
                                    {...field}
                                    className="min-h-0"
                                    disabled={isLoading}
                                    maxRows={9}
                                    minRows={1}
                                    onKeyDown={handleKeyDown}
                                    placeholder="Describe your testing scenario..."
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
                                                            disabled={isLoading}
                                                            variant="ghost"
                                                        >
                                                            <span className="flex items-center gap-2">
                                                                {currentProvider &&
                                                                    getProviderIcon(currentProvider, 'size-4 shrink-0')}
                                                                <span className="max-w-40 truncate">
                                                                    {currentProvider
                                                                        ? getProviderDisplayName(currentProvider)
                                                                        : 'Select Provider'}
                                                                </span>
                                                                <ChevronsUpDown className="size-3 shrink-0" />
                                                            </span>
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
                                                                            providerField.onChange(provider.name);
                                                                            setSelectedProvider(provider);
                                                                            setProviderSearch('');
                                                                        }}
                                                                    >
                                                                        <div className="flex w-full min-w-0 items-center gap-2">
                                                                            {getProviderIcon(
                                                                                provider,
                                                                                'size-4 shrink-0',
                                                                            )}

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
                                    {flowType === 'assistant' && (
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
                                                                        disabled={isLoading}
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
                                    <InputGroupButton
                                        className="ml-auto"
                                        disabled={isLoading || !isValid}
                                        size="icon-xs"
                                        type="submit"
                                        variant="default"
                                    >
                                        {isLoading ? <Spinner variant="circle" /> : <ArrowUpIcon />}
                                    </InputGroupButton>
                                </InputGroupAddon>
                            </InputGroup>
                        </FormControl>
                    )}
                />
            </form>
        </Form>
    );
};
