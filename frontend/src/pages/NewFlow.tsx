import { zodResolver } from '@hookform/resolvers/zod';
import { ArrowUpIcon, Check, ChevronsUpDown } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { useNavigate } from 'react-router-dom';
import { toast } from 'sonner';
import { z } from 'zod';

import { Breadcrumb, BreadcrumbItem, BreadcrumbList, BreadcrumbPage } from '@/components/ui/breadcrumb';
import { Card, CardContent } from '@/components/ui/card';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Form, FormControl, FormField, FormItem, FormLabel } from '@/components/ui/form';
import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupTextarea } from '@/components/ui/input-group';
import { Separator } from '@/components/ui/separator';
import { SidebarTrigger } from '@/components/ui/sidebar';
import { Switch } from '@/components/ui/switch';
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { useCreateAssistantMutation, useCreateFlowMutation } from '@/graphql/types';
import { Log } from '@/lib/log';
import { getProviderDisplayName, getProviderIcon } from '@/models/Provider';
import { useProviders } from '@/providers/ProvidersProvider';

const formSchema = z.object({
    flowType: z.enum(['automation', 'assistant']),
    message: z.string().trim().min(1, { message: 'Message cannot be empty' }),
    providerName: z.string().trim().min(1, { message: 'Provider must be selected' }),
    useAgents: z.boolean(),
});

const NewFlow = () => {
    const navigate = useNavigate();

    const { providers, selectedProvider } = useProviders();

    const [createFlow] = useCreateFlowMutation();
    const [createAssistant] = useCreateAssistantMutation();
    const [isFlowCreating, setIsCreatingFlow] = useState(false);

    const form = useForm<z.infer<typeof formSchema>>({
        defaultValues: {
            flowType: 'automation',
            message: '',
            providerName: selectedProvider?.name ?? '',
            useAgents: false,
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
        watch,
    } = form;

    const flowType = watch('flowType');

    // Update provider when selectedProvider changes, if user hasn't manually selected one
    useEffect(() => {
        const currentProviderName = getValues('providerName');
        const isProviderFieldDirty = dirtyFields.providerName;

        // Only update if user hasn't manually changed the field and either:
        // 1. Current value is empty
        // 2. selectedProvider is available and different from current
        if (!isProviderFieldDirty && selectedProvider?.name && selectedProvider.name !== currentProviderName) {
            setValue('providerName', selectedProvider.name);
        }
    }, [selectedProvider, dirtyFields.providerName, getValues, setValue]);

    const handleSubmit = async (values: z.infer<typeof formSchema>) => {
        const { flowType, message, providerName, useAgents } = values;

        const input = message.trim();
        const modelProvider = providerName.trim();

        if (!input || !modelProvider || isFlowCreating) {
            return;
        }

        try {
            setIsCreatingFlow(true);

            if (flowType === 'automation') {
                const { data } = await createFlow({
                    variables: {
                        input,
                        modelProvider,
                    },
                });

                if (data?.createFlow) {
                    // Navigate to the new flow page
                    navigate(`/flows/${data.createFlow.id}`);
                }
            } else {
                const { data } = await createAssistant({
                    variables: {
                        flowId: '0',
                        input,
                        modelProvider,
                        useAgents,
                    },
                });

                if (data?.createAssistant?.flow) {
                    // Navigate to the new flow page
                    navigate(`/flows/${data.createAssistant.flow.id}`);
                }
            }
        } catch (error) {
            const description = error instanceof Error ? error.message : 'An error occurred while creating the flow';
            const errorTitle = flowType === 'automation' ? 'Failed to create flow' : 'Failed to create assistant';
            toast.error(errorTitle, {
                description,
            });
            Log.error(`Error creating ${flowType}:`, error);
        } finally {
            setIsCreatingFlow(false);
        }
    };

    const handleKeyDown = (event: React.KeyboardEvent<HTMLTextAreaElement>) => {
        if (isFlowCreating) {
            return;
        }

        const isEnterPress = event.key === 'Enter';
        const isCtrlEnter = isEnterPress && (event.ctrlKey || event.metaKey);

        if (!isCtrlEnter) {
            return;
        }

        event.preventDefault();
        handleFormSubmit(handleSubmit)();
    };

    return (
        <>
            <header className="sticky top-0 z-10 flex h-12 w-full shrink-0 items-center gap-2 border-b bg-background transition-[width,height] ease-linear group-has-[[data-collapsible=icon]]/sidebar-wrapper:h-12">
                <div className="flex w-full items-center justify-between gap-2 px-4">
                    <div className="flex items-center gap-2">
                        <SidebarTrigger className="-ml-1" />
                        <Separator
                            className="mr-2 h-4"
                            orientation="vertical"
                        />
                        <Breadcrumb>
                            <BreadcrumbList>
                                <BreadcrumbItem>
                                    <BreadcrumbPage>New flow</BreadcrumbPage>
                                </BreadcrumbItem>
                            </BreadcrumbList>
                        </Breadcrumb>
                    </div>
                </div>
            </header>
            <div className="flex h-[calc(100dvh-3rem)] w-full max-w-full flex-1 items-center justify-center">
                <Card className="mx-4 w-full max-w-2xl">
                    <CardContent className="pt-6">
                        <div className="space-y-4">
                            <div className="text-center">
                                <h1 className="text-2xl font-semibold">Create a new flow</h1>
                                <p className="mt-2 text-muted-foreground">
                                    Describe what you would like PentAGI to test
                                </p>
                            </div>
                            <Form {...form}>
                                <form
                                    className="space-y-4"
                                    onSubmit={handleFormSubmit(handleSubmit)}
                                >
                                    <FormField
                                        control={control}
                                        name="flowType"
                                        render={({ field }) => (
                                            <FormItem>
                                                <FormControl>
                                                    <Tabs
                                                        onValueChange={field.onChange}
                                                        value={field.value}
                                                    >
                                                        <TabsList className="grid w-full grid-cols-2">
                                                            <TabsTrigger
                                                                disabled={isFlowCreating}
                                                                value="automation"
                                                            >
                                                                Automation
                                                            </TabsTrigger>
                                                            <TabsTrigger
                                                                disabled={isFlowCreating}
                                                                value="assistant"
                                                            >
                                                                Assistant
                                                            </TabsTrigger>
                                                        </TabsList>
                                                    </Tabs>
                                                </FormControl>
                                            </FormItem>
                                        )}
                                    />
                                    <FormField
                                        control={control}
                                        name="message"
                                        render={({ field }) => (
                                            <FormControl>
                                                <InputGroup>
                                                    <InputGroupTextarea
                                                        {...field}
                                                        disabled={isFlowCreating}
                                                        onKeyDown={handleKeyDown}
                                                        placeholder="Describe your testing scenario..."
                                                        rows={4}
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
                                                                                disabled={isFlowCreating}
                                                                                variant="ghost"
                                                                            >
                                                                                <span className="flex items-center gap-2">
                                                                                    {currentProvider &&
                                                                                        getProviderIcon(
                                                                                            currentProvider,
                                                                                            'h-4 w-4 shrink-0',
                                                                                        )}
                                                                                    <span className="max-w-[120px] truncate">
                                                                                        {currentProvider
                                                                                            ? getProviderDisplayName(
                                                                                                  currentProvider,
                                                                                              )
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
                                                                                        providerField.onChange(
                                                                                            provider.name,
                                                                                        );
                                                                                    }}
                                                                                >
                                                                                    <div className="flex w-full min-w-0 items-center gap-2">
                                                                                        <div className="shrink-0">
                                                                                            {getProviderIcon(
                                                                                                provider,
                                                                                                'h-4 w-4 shrink-0',
                                                                                            )}
                                                                                        </div>
                                                                                        <span className="flex-1 truncate">
                                                                                            {getProviderDisplayName(
                                                                                                provider,
                                                                                            )}
                                                                                        </span>
                                                                                        {providerField.value ===
                                                                                            provider.name && (
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
                                                                                            checked={
                                                                                                useAgentsField.value
                                                                                            }
                                                                                            disabled={isFlowCreating}
                                                                                            onCheckedChange={
                                                                                                useAgentsField.onChange
                                                                                            }
                                                                                        />
                                                                                    </FormControl>
                                                                                    <FormLabel
                                                                                        className="flex cursor-pointer pl-2 text-xs font-normal"
                                                                                        onClick={() =>
                                                                                            useAgentsField.onChange(
                                                                                                !useAgentsField.value,
                                                                                            )
                                                                                        }
                                                                                    >
                                                                                        Use Agents
                                                                                    </FormLabel>
                                                                                </div>
                                                                            </TooltipTrigger>
                                                                            <TooltipContent>
                                                                                <p className="max-w-48">
                                                                                    Enable multi-agent collaboration for
                                                                                    complex tasks
                                                                                </p>
                                                                            </TooltipContent>
                                                                        </Tooltip>
                                                                    </TooltipProvider>
                                                                )}
                                                            />
                                                        )}
                                                        <InputGroupButton
                                                            className="ml-auto"
                                                            disabled={isFlowCreating || !isValid}
                                                            size="icon-xs"
                                                            type="submit"
                                                            variant="default"
                                                        >
                                                            <ArrowUpIcon />
                                                        </InputGroupButton>
                                                    </InputGroupAddon>
                                                </InputGroup>
                                            </FormControl>
                                        )}
                                    />
                                </form>
                            </Form>
                        </div>
                    </CardContent>
                </Card>
            </div>
        </>
    );
};

export default NewFlow;
