import { zodResolver } from '@hookform/resolvers/zod';
import { ArrowUpIcon, Check, ChevronsUpDown } from 'lucide-react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';

import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Form, FormControl, FormField, FormLabel } from '@/components/ui/form';
import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupTextareaAutosize } from '@/components/ui/input-group';
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
    const { providers } = useProviders();

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
        formState: { isValid },
        handleSubmit: handleFormSubmit,
    } = form;

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
                                                                <span className="max-w-[120px] truncate">
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
                                                        className="[--radius:0.95rem]"
                                                        side="top"
                                                    >
                                                        {providers.map((provider) => (
                                                            <DropdownMenuItem
                                                                className="focus:outline-none focus-visible:outline-none focus-visible:ring-0"
                                                                key={provider.name}
                                                                onSelect={(e) => {
                                                                    e.preventDefault();
                                                                    providerField.onChange(provider.name);
                                                                }}
                                                            >
                                                                <div className="flex w-full min-w-0 items-center gap-2">
                                                                    <div className="shrink-0">
                                                                        {getProviderIcon(provider, 'size-4 shrink-0')}
                                                                    </div>
                                                                    <span className="flex-1 truncate">
                                                                        {getProviderDisplayName(provider)}
                                                                    </span>
                                                                    {providerField.value === provider.name && (
                                                                        <div className="shrink-0">
                                                                            <Check className="size-4 shrink-0" />
                                                                        </div>
                                                                    )}
                                                                </div>
                                                            </DropdownMenuItem>
                                                        ))}
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
