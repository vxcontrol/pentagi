import { Check, ChevronsUpDown, GripVertical, Loader2 } from 'lucide-react';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useNavigate, useOutletContext, useParams } from 'react-router-dom';

import type { Provider } from '@/models/Provider';

import {
    Breadcrumb,
    BreadcrumbItem,
    BreadcrumbList,
    BreadcrumbPage,
    BreadcrumbProvider,
    BreadcrumbStatus,
} from '@/components/ui/breadcrumb';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { ResizableHandle, ResizablePanel, ResizablePanelGroup } from '@/components/ui/resizable';
import { Separator } from '@/components/ui/separator';
import { SidebarTrigger } from '@/components/ui/sidebar';
import ChatCentralTabs from '@/features/chat/ChatCentralTabs';
import ChatTabs from '@/features/chat/ChatTabs';
import {
    StatusType,
    useAgentLogAddedSubscription,
    useAssistantCreatedSubscription,
    useAssistantDeletedSubscription,
    useAssistantLogAddedSubscription,
    useAssistantLogsQuery,
    useAssistantLogUpdatedSubscription,
    useAssistantsQuery,
    useAssistantUpdatedSubscription,
    useCallAssistantMutation,
    useCreateAssistantMutation,
    useCreateFlowMutation,
    useDeleteAssistantMutation,
    useFlowQuery,
    useMessageLogAddedSubscription,
    useMessageLogUpdatedSubscription,
    usePutUserInputMutation,
    useScreenshotAddedSubscription,
    useSearchLogAddedSubscription,
    useStopAssistantMutation,
    useStopFlowMutation,
    useTaskCreatedSubscription,
    useTaskUpdatedSubscription,
    useTerminalLogAddedSubscription,
    useVectorStoreLogAddedSubscription,
} from '@/graphql/types';
import { useBreakpoint } from '@/hooks/use-breakpoint';
import { Log } from '@/lib/log';
import { getProviderDisplayName, getProviderIcon, isProviderValid } from '@/models/Provider';

interface MainLayoutContext {
    onChangeSelectedProvider: (provider: Provider) => void;
    selectedFlowId: null | string;
    selectedProvider: null | Provider;
    sortedProviders: Provider[];
}

const Chat = () => {
    const { isDesktop } = useBreakpoint();
    const { flowId } = useParams();
    const navigate = useNavigate();

    // Get data from MainLayout context
    const { onChangeSelectedProvider, selectedFlowId, selectedProvider, sortedProviders } =
        useOutletContext<MainLayoutContext>();

    // Add debounced flow ID to prevent rapid switching causing database connection issues
    const [debouncedFlowId, setDebouncedFlowId] = useState<null | string>(flowId ?? null);
    const debounceTimeoutRef = useRef<NodeJS.Timeout | null>(null);

    // AbortController to cancel active requests when switching flows
    const abortControllerRef = useRef<AbortController | null>(null);

    // Timeout ref for assistant creation to prevent memory leaks
    const assistantCreationTimeoutRef = useRef<NodeJS.Timeout | null>(null);

    // State for preserving active tabs when switching flows
    const [activeCentralTab, setActiveCentralTab] = useState<string>('automation');
    const [activeTabsTab, setActiveTabsTab] = useState<string>(!isDesktop ? 'automation' : 'terminal');

    const { data: flowData, loading: isFlowLoading } = useFlowQuery({
        errorPolicy: 'all',
        fetchPolicy: 'cache-and-network',
        nextFetchPolicy: 'cache-first',
        notifyOnNetworkStatusChange: false,
        skip: !debouncedFlowId || debouncedFlowId === 'new',
        variables: { id: debouncedFlowId ?? '' },
    });

    // Store currently selected assistant ID for each flow
    const [selectedAssistantIds, setSelectedAssistantIds] = useState<Record<string, null | string>>({});

    // Get assistants for the current flow
    const { data: assistantsData } = useAssistantsQuery({
        fetchPolicy: 'cache-and-network',
        nextFetchPolicy: 'cache-first',
        skip: !debouncedFlowId || debouncedFlowId === 'new',
        variables: { flowId: debouncedFlowId ?? '' },
    });

    // Currently selected assistant for this flow (with auto-selection logic)
    const selectedAssistantId = useMemo(() => {
        if (!selectedFlowId || selectedFlowId === 'new') {
            return null;
        }

        // If we have an explicit selection for this flow, use it
        if (selectedFlowId in selectedAssistantIds) {
            const explicitSelection = selectedAssistantIds[selectedFlowId];

            // If explicit selection is null, user wants to create new assistant
            if (explicitSelection === null) {
                return null;
            }

            // If explicit selection still exists in assistants list, use it
            if (assistantsData?.assistants?.some((assistant) => assistant.id === explicitSelection)) {
                return explicitSelection;
            }

            // Explicit selection was deleted, fall through to auto-select
        }

        // Auto-select first (newest) assistant if available
        const firstAssistantId = assistantsData?.assistants?.[0]?.id;

        return firstAssistantId || null;
    }, [selectedFlowId, selectedAssistantIds, assistantsData?.assistants]);

    // Store for API mutations
    const [createFlow] = useCreateFlowMutation();
    const [putUserInput] = usePutUserInputMutation();
    const [stopFlow] = useStopFlowMutation();
    const [createAssistant] = useCreateAssistantMutation();
    const [callAssistant] = useCallAssistantMutation();
    const [stopAssistant] = useStopAssistantMutation();
    const [deleteAssistant] = useDeleteAssistantMutation();

    // Fetch assistant logs if an assistant is selected
    const { data: assistantLogsData, refetch: refetchAssistantLogs } = useAssistantLogsQuery({
        fetchPolicy: 'cache-and-network',
        nextFetchPolicy: 'cache-first',
        skip: !debouncedFlowId || !selectedAssistantId || debouncedFlowId === 'new' || selectedAssistantId === '',
        variables: { assistantId: selectedAssistantId ?? '', flowId: debouncedFlowId ?? '' },
    });

    // Debounce selectedFlowId changes to prevent excessive database connections
    // When users rapidly switch between flows, each switch triggers multiple GraphQL queries/subscriptions
    // This can quickly exhaust the database connection pool, causing "too many clients already" errors
    useEffect(() => {
        // Cancel any pending requests from previous flow
        if (abortControllerRef.current) {
            abortControllerRef.current.abort();
        }

        // Create new AbortController for this flow
        abortControllerRef.current = new AbortController();

        // Clear any existing timeout
        if (debounceTimeoutRef.current) {
            clearTimeout(debounceTimeoutRef.current);
        }

        // Set new timeout
        debounceTimeoutRef.current = setTimeout(() => {
            setDebouncedFlowId(selectedFlowId);
            debounceTimeoutRef.current = null;
        }, 300);

        // Cleanup on flow change or unmount
        return () => {
            if (debounceTimeoutRef.current) {
                clearTimeout(debounceTimeoutRef.current);
                debounceTimeoutRef.current = null;
            }

            if (abortControllerRef.current) {
                abortControllerRef.current.abort();
                abortControllerRef.current = null;
            }
        };
    }, [selectedFlowId]);

    // Additional cleanup on component unmount to prevent memory leaks
    useEffect(() => {
        return () => {
            if (assistantCreationTimeoutRef.current) {
                clearTimeout(assistantCreationTimeoutRef.current);
                assistantCreationTimeoutRef.current = null;
            }
        };
    }, []);

    const variables = useMemo(() => ({ flowId: debouncedFlowId || '' }), [debouncedFlowId]);
    const skip = useMemo(() => !debouncedFlowId || debouncedFlowId === 'new', [debouncedFlowId]);

    // Flow-specific subscriptions that depend on the selected flow
    useTaskCreatedSubscription({ skip, variables });
    useTaskUpdatedSubscription({ skip, variables });
    useScreenshotAddedSubscription({ skip, variables });
    useTerminalLogAddedSubscription({ skip, variables });
    useMessageLogUpdatedSubscription({ skip, variables });
    useMessageLogAddedSubscription({ skip, variables });
    useAgentLogAddedSubscription({ skip, variables });
    useSearchLogAddedSubscription({ skip, variables });
    useVectorStoreLogAddedSubscription({ skip, variables });

    // Assistant-specific subscriptions
    useAssistantCreatedSubscription({ skip, variables });
    useAssistantUpdatedSubscription({ skip, variables });
    useAssistantDeletedSubscription({ skip, variables });
    useAssistantLogAddedSubscription({ skip, variables });
    useAssistantLogUpdatedSubscription({ skip, variables });

    // Handle selecting an assistant
    const handleSelectAssistant = useCallback(
        (assistantId: null | string) => {
            if (!selectedFlowId || selectedFlowId === 'new') {
                return;
            }

            setSelectedAssistantIds((prev) => ({
                ...prev,
                [selectedFlowId]: assistantId,
            }));
        },
        [selectedFlowId],
    );

    // Set selectedAssistantId to null to initiate assistant creation
    const handleInitiateAssistantCreation = () => {
        if (!selectedFlowId || selectedFlowId === 'new') {
            return;
        }

        handleSelectAssistant(null);
    };

    const handleSubmitAutomationMessage = async (message: string) => {
        if (!selectedFlowId || flowData?.flow?.status === StatusType.Finished) {
            return;
        }

        try {
            if (selectedFlowId !== 'new') {
                await putUserInput({
                    variables: {
                        flowId: selectedFlowId ?? '',
                        input: message,
                    },
                });

                return;
            }

            // Double check that we have a valid provider before creating a flow
            let providerToUse = '';

            // First try to use the selected provider if it's valid
            if (selectedProvider && isProviderValid(selectedProvider, sortedProviders)) {
                providerToUse = selectedProvider.name;
            } else if (sortedProviders.length > 0) {
                // If no provider is selected or it's invalid, try to use the first available
                const firstAvailableProvider = sortedProviders[0];

                if (firstAvailableProvider) {
                    providerToUse = firstAvailableProvider.name;
                }
            }

            // Fail early if we still don't have a provider
            if (!providerToUse) {
                Log.error('No valid provider available for creating a flow');

                return;
            }

            const { data } = await createFlow({
                variables: {
                    input: message,
                    modelProvider: providerToUse,
                },
            });

            if (data?.createFlow) {
                const newFlowId = data.createFlow.id.toString();

                // Navigate to the new flow page but stay on the automation tab
                navigate(`/flows/${newFlowId}`);

                // Keep user on the automation tab
                setActiveCentralTab('automation');

                if (!isDesktop) {
                    setActiveTabsTab('automation');
                }
            }
        } catch (error) {
            Log.error('Error submitting message:', error);
            throw error;
        }
    };

    const handleStopAutomationFlow = async (flowId: string) => {
        try {
            await stopFlow({
                variables: {
                    flowId,
                },
            });
        } catch (error) {
            Log.error('Error stopping flow:', error);
        }
    };

    // Handle creating a new assistant
    const handleCreateAssistant = useCallback(
        async (message: string, useAgents: boolean) => {
            if (!message.trim() || !selectedProvider) {
                return;
            }

            try {
                // Backend will create a flow if flowId is 'new' or '0'
                const flowIdToUse = selectedFlowId === 'new' ? '0' : selectedFlowId || '0';

                const { data } = await createAssistant({
                    variables: {
                        flowId: flowIdToUse,
                        input: message.trim(),
                        modelProvider: selectedProvider.name,
                        useAgents,
                    },
                });

                if (data?.createAssistant) {
                    const { assistant, flow } = data.createAssistant;

                    // If we created a new flow, navigate to it
                    if (selectedFlowId === 'new' || !selectedFlowId) {
                        // Navigate to the new flow page
                        navigate(`/flows/${flow.id}`);

                        // Set Assistant tab as active
                        setActiveCentralTab('assistant');

                        if (!isDesktop) {
                            setActiveTabsTab('assistant');
                        }
                    }

                    // Select the newly created assistant
                    if (assistant?.id) {
                        // Clear any existing timeout to prevent memory leaks
                        if (assistantCreationTimeoutRef.current) {
                            clearTimeout(assistantCreationTimeoutRef.current);
                        }

                        // Use setTimeout to prevent race conditions with queries
                        assistantCreationTimeoutRef.current = setTimeout(() => {
                            setSelectedAssistantIds((prev) => ({
                                ...prev,
                                [flow.id]: assistant.id,
                            }));
                            assistantCreationTimeoutRef.current = null;
                        }, 300);
                    }
                }
            } catch (error) {
                Log.error('Error creating assistant:', error);
                throw error;
            }
        },
        [selectedFlowId, selectedProvider, createAssistant, navigate, isDesktop, setActiveCentralTab, setActiveTabsTab],
    );

    // Handle calling an existing assistant
    const handleCallAssistant = async (assistantId: string, message: string, useAgents: boolean): Promise<void> => {
        if (!selectedFlowId || selectedFlowId === 'new' || !assistantId || !message.trim()) {
            return;
        }

        try {
            await callAssistant({
                variables: {
                    assistantId,
                    flowId: selectedFlowId,
                    input: message.trim(),
                    useAgents,
                },
            });

            // Refresh assistant logs after calling - only if we have a selected assistant
            if (selectedAssistantId) {
                refetchAssistantLogs();
            }
        } catch (error) {
            Log.error('Error calling assistant:', error);
            throw error;
        }
    };

    // Handle stopping an assistant
    const handleStopAssistant = async (assistantId: string) => {
        if (!selectedFlowId || selectedFlowId === 'new' || !assistantId) {
            return;
        }

        try {
            await stopAssistant({
                variables: {
                    assistantId,
                    flowId: selectedFlowId,
                },
            });

            // Force refresh assistant data after stopping - only if we have a selected assistant
            if (selectedAssistantId) {
                refetchAssistantLogs();
            }
        } catch (error) {
            Log.error('Error stopping assistant:', error);
            throw error;
        }
    };

    // Handle deleting an assistant
    const handleDeleteAssistant = async (assistantId: string) => {
        if (!selectedFlowId || selectedFlowId === 'new' || !assistantId) {
            return;
        }

        try {
            // Remember if this was the currently selected assistant
            const wasSelected = selectedAssistantId === assistantId;

            await deleteAssistant({
                update: (cache) => {
                    // Remove the assistant from Apollo cache
                    cache.evict({ id: `Assistant:${assistantId}` });
                    cache.gc();
                },
                variables: {
                    assistantId,
                    flowId: selectedFlowId,
                },
            });

            // If we deleted the currently selected assistant, reset selection
            if (wasSelected) {
                setSelectedAssistantIds((prev) => {
                    const updatedState = { ...prev };
                    updatedState[selectedFlowId] = null;

                    return updatedState;
                });
            }
        } catch (error) {
            Log.error('Error deleting assistant:', error);
            throw error;
        }
    };

    const tabsCard = (
        <Card className="flex h-[calc(100dvh-3rem)] max-w-full flex-col rounded-none border-0">
            <CardContent className="flex-1 overflow-auto pr-0 pt-4">
                <ChatTabs
                    activeTab={activeTabsTab}
                    assistantLogs={assistantLogsData?.assistantLogs ?? []}
                    assistants={assistantsData?.assistants ?? []}
                    flowData={flowData}
                    onCreateAssistant={handleInitiateAssistantCreation}
                    onCreateNewAssistant={handleCreateAssistant}
                    onDeleteAssistant={handleDeleteAssistant}
                    onSelectAssistant={handleSelectAssistant}
                    onStopAssistant={handleStopAssistant}
                    onStopAutomationFlow={handleStopAutomationFlow}
                    onSubmitAssistantMessage={handleCallAssistant}
                    onSubmitAutomationMessage={handleSubmitAutomationMessage}
                    onTabChange={setActiveTabsTab}
                    providers={sortedProviders}
                    selectedAssistantId={selectedAssistantId}
                    selectedFlowId={selectedFlowId}
                    selectedProvider={selectedProvider}
                />
            </CardContent>
        </Card>
    );

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
                                    {flowData?.flow && (
                                        <>
                                            <BreadcrumbStatus status={flowData.flow.status} />
                                            <BreadcrumbProvider provider={flowData.flow.provider} />
                                        </>
                                    )}
                                    <BreadcrumbPage>
                                        {flowData?.flow?.title ||
                                            (selectedFlowId === 'new' ? 'New flow' : 'Select a flow')}
                                    </BreadcrumbPage>
                                </BreadcrumbItem>
                            </BreadcrumbList>
                        </Breadcrumb>
                    </div>
                    <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                            <Button
                                className="h-8 gap-1 focus:outline-none focus-visible:outline-none focus-visible:ring-0"
                                size="sm"
                                variant="ghost"
                            >
                                <span className="max-w-[120px] truncate">
                                    {selectedProvider ? getProviderDisplayName(selectedProvider) : 'Select Provider'}
                                </span>
                                <ChevronsUpDown className="size-4 shrink-0" />
                            </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent
                            align="end"
                            className="w-fit min-w-[150px] max-w-[280px]"
                            onCloseAutoFocus={(e) => {
                                e.preventDefault();
                            }}
                        >
                            {sortedProviders.map((provider) => (
                                <DropdownMenuItem
                                    className="focus:outline-none focus-visible:outline-none focus-visible:ring-0"
                                    key={provider.name}
                                    onSelect={(e) => {
                                        e.preventDefault();
                                        onChangeSelectedProvider(provider);
                                    }}
                                >
                                    <div className="flex w-full min-w-0 items-center gap-2">
                                        <div className="shrink-0">{getProviderIcon(provider, 'h-4 w-4 shrink-0')}</div>
                                        <span className="max-w-[180px] flex-1 truncate">
                                            {getProviderDisplayName(provider)}
                                        </span>
                                        {selectedProvider?.name === provider.name && (
                                            <div className="shrink-0">
                                                <Check className="size-4 shrink-0" />
                                            </div>
                                        )}
                                    </div>
                                </DropdownMenuItem>
                            ))}
                        </DropdownMenuContent>
                    </DropdownMenu>
                </div>
            </header>
            <div className="relative flex h-[calc(100dvh-3rem)] w-full max-w-full flex-1">
                {isFlowLoading && (
                    <div className="absolute inset-0 z-50 flex items-center justify-center bg-background/50">
                        <Loader2 className="size-16 animate-spin" />
                    </div>
                )}
                {isDesktop ? (
                    <ResizablePanelGroup
                        className="w-full"
                        direction="horizontal"
                    >
                        <ResizablePanel
                            defaultSize={50}
                            minSize={30}
                        >
                            <Card className="flex h-[calc(100dvh-3rem)] max-w-full flex-col rounded-none border-0">
                                <CardContent className="flex-1 overflow-auto pr-0 pt-4">
                                    <ChatCentralTabs
                                        activeTab={activeCentralTab}
                                        assistantLogs={assistantLogsData?.assistantLogs ?? []}
                                        assistants={assistantsData?.assistants ?? []}
                                        flowData={flowData}
                                        onCreateAssistant={handleInitiateAssistantCreation}
                                        onCreateNewAssistant={handleCreateAssistant}
                                        onDeleteAssistant={handleDeleteAssistant}
                                        onSelectAssistant={handleSelectAssistant}
                                        onStopAssistant={handleStopAssistant}
                                        onStopAutomationFlow={handleStopAutomationFlow}
                                        onSubmitAssistantMessage={handleCallAssistant}
                                        onSubmitAutomationMessage={handleSubmitAutomationMessage}
                                        onTabChange={setActiveCentralTab}
                                        providers={sortedProviders}
                                        selectedAssistantId={selectedAssistantId}
                                        selectedFlowId={selectedFlowId}
                                        selectedProvider={selectedProvider}
                                    />
                                </CardContent>
                            </Card>
                        </ResizablePanel>
                        <ResizableHandle withHandle>
                            <GripVertical className="size-4" />
                        </ResizableHandle>
                        <ResizablePanel
                            defaultSize={50}
                            minSize={30}
                        >
                            {tabsCard}
                        </ResizablePanel>
                    </ResizablePanelGroup>
                ) : (
                    tabsCard
                )}
            </div>
        </>
    );
};

export default Chat;
