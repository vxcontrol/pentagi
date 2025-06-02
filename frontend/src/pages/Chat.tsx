import { GripVertical, Loader2 } from 'lucide-react';
import { useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';

import {
    Breadcrumb,
    BreadcrumbItem,
    BreadcrumbList,
    BreadcrumbPage,
    BreadcrumbProvider,
    BreadcrumbStatus,
} from '@/components/ui/breadcrumb';
import { Card, CardContent } from '@/components/ui/card';
import { ResizableHandle, ResizablePanel, ResizablePanelGroup } from '@/components/ui/resizable';
import { Separator } from '@/components/ui/separator';
import { SidebarInset, SidebarProvider, SidebarTrigger } from '@/components/ui/sidebar';
import ChatCentralTabs from '@/features/chat/ChatCentralTabs';
import ChatSidebar from '@/features/chat/ChatSidebar';
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
    useDeleteFlowMutation,
    useFinishFlowMutation,
    useFlowCreatedSubscription,
    useFlowDeletedSubscription,
    useFlowQuery,
    useFlowsQuery,
    useFlowUpdatedSubscription,
    useMessageLogAddedSubscription,
    useMessageLogUpdatedSubscription,
    useProvidersQuery,
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
import type { User } from '@/models/User';

const SELECTED_PROVIDER_KEY = 'selectedProvider';

const Chat = () => {
    const { isDesktop } = useBreakpoint();
    const { flowId } = useParams();
    const navigate = useNavigate();
    const { data: flowsData, refetch: refetchFlows } = useFlowsQuery();
    const [selectedFlowId, setSelectedFlowId] = useState<string | null>(flowId ?? null);
    
    // Add debounced flow ID to prevent rapid switching causing database connection issues
    const [debouncedFlowId, setDebouncedFlowId] = useState<string | null>(flowId ?? null);
    const debounceTimeoutRef = useRef<NodeJS.Timeout | null>(null);
    
    // AbortController to cancel active requests when switching flows
    const abortControllerRef = useRef<AbortController | null>(null);

    // Timeout ref for assistant creation to prevent memory leaks
    const assistantCreationTimeoutRef = useRef<NodeJS.Timeout | null>(null);

    // State for preserving active tabs when switching flows
    const [activeCentralTab, setActiveCentralTab] = useState<string>('automation');
    const [activeTabsTab, setActiveTabsTab] = useState<string>(!isDesktop ? 'automation' : 'terminal');

    const { data: flowData, loading: isFlowLoading } = useFlowQuery({
        variables: { id: debouncedFlowId ?? '' },
        skip: !debouncedFlowId || debouncedFlowId === 'new',
        fetchPolicy: 'cache-and-network',
        nextFetchPolicy: 'cache-first',
        notifyOnNetworkStatusChange: false,
        errorPolicy: 'all',
    });
    const { data: providersData } = useProvidersQuery();

    // Store currently selected assistant ID for each flow
    const [selectedAssistantIds, setSelectedAssistantIds] = useState<Record<string, string | null>>({});

    // Create a ref to track current selectedAssistantIds without triggering re-renders
    const selectedAssistantIdsRef = useRef<Record<string, string | null>>({});

    // Keep the ref updated with the latest state
    useEffect(() => {
        selectedAssistantIdsRef.current = selectedAssistantIds;
    }, [selectedAssistantIds]);

    // Currently selected assistant for this flow
    const selectedAssistantId = useMemo(() => {
        if (!selectedFlowId || selectedFlowId === 'new') return null;
        return selectedAssistantIds[selectedFlowId] || null;
    }, [selectedFlowId, selectedAssistantIds]);

    // Get assistants for the current flow
    const { data: assistantsData } = useAssistantsQuery({
        variables: { flowId: debouncedFlowId ?? '' },
        skip: !debouncedFlowId || debouncedFlowId === 'new',
        fetchPolicy: 'cache-and-network',
        nextFetchPolicy: 'cache-first',
    });

    // Create sorted providers list to ensure consistent order
    const sortedProviders = useMemo(() => {
        const providers = providersData?.providers || [];
        return [...providers].sort();
    }, [providersData?.providers]);

    const [selectedProvider, setSelectedProvider] = useState<string | null>(null);
    const [user, setUser] = useState<User | null>(null);
    const needsProviderUpdateRef = useRef(true);
    const previousFlowIdRef = useRef(flowId);
    const userInitialized = useRef(false);

    // Store for API mutations
    const [createFlow] = useCreateFlowMutation();
    const [deleteFlow] = useDeleteFlowMutation();
    const [finishFlow] = useFinishFlowMutation();
    const [putUserInput] = usePutUserInputMutation();
    const [stopFlow] = useStopFlowMutation();
    const [createAssistant] = useCreateAssistantMutation();
    const [callAssistant] = useCallAssistantMutation();
    const [stopAssistant] = useStopAssistantMutation();
    const [deleteAssistant] = useDeleteAssistantMutation();

    // Fetch assistant logs if an assistant is selected
    const { data: assistantLogsData, refetch: refetchAssistantLogs } = useAssistantLogsQuery({
        variables: { flowId: debouncedFlowId ?? '', assistantId: selectedAssistantId ?? '' },
        skip: !debouncedFlowId || !selectedAssistantId || debouncedFlowId === 'new' || selectedAssistantId === '',
        fetchPolicy: 'cache-and-network',
        nextFetchPolicy: 'cache-first',
    });

    // Check if provider needs initialization or update when:
    // 1. We have no selected provider yet (first login)
    // 2. Selected provider is not in available providers list
    useEffect(() => {
        // On first load or when providers change
        if (sortedProviders.length > 0) {
            if (!selectedProvider) {
                // Case 1: No provider selected yet (first login)
                needsProviderUpdateRef.current = true;
            } else {
                // Case 2: Check if selected provider is still available
                const isProviderStillAvailable = sortedProviders.includes(selectedProvider);
                if (!isProviderStillAvailable) {
                    needsProviderUpdateRef.current = true;
                }
            }
        }
    }, [sortedProviders, selectedProvider]);

    // Update provider when needed - separate from detection for ESLint compliance
    useLayoutEffect(() => {
        if (needsProviderUpdateRef.current && sortedProviders.length > 0) {
            needsProviderUpdateRef.current = false;

            // Check if we have a previously saved provider
            const savedProvider = localStorage.getItem(SELECTED_PROVIDER_KEY);

            // Case 3: If saved provider exists and is valid, use it
            if (savedProvider && sortedProviders.includes(savedProvider)) {
                setSelectedProvider(savedProvider);
            } else if (sortedProviders[0]) {
                const firstProvider = sortedProviders[0];
                setSelectedProvider(firstProvider);
                localStorage.setItem(SELECTED_PROVIDER_KEY, firstProvider);
            }
        }
    }, [sortedProviders]);

    useEffect(() => {
        const auth = localStorage.getItem('auth');

        if (!auth) {
            // Save current path for redirect after login
            const currentPath = window.location.pathname;
            // Only save if it's not the default route
            const returnParam = currentPath !== '/chat/new' ? `?returnUrl=${encodeURIComponent(currentPath)}` : '';
            navigate(`/login${returnParam}`);
            return;
        }

        try {
            const authData = JSON.parse(auth);
            const user = authData?.user;

            if (!user) {
                // Save current path for redirect after login
                const currentPath = window.location.pathname;
                // Only save if it's not the default route
                const returnParam = currentPath !== '/chat/new' ? `?returnUrl=${encodeURIComponent(currentPath)}` : '';
                navigate(`/login${returnParam}`);
                return;
            } else {
                userInitialized.current = true;
                const frameId = window.requestAnimationFrame(() => setUser(user));
                return () => window.cancelAnimationFrame(frameId);
            }
        } catch {
            // If we have a parse error, redirect to login
            navigate('/login');
        }
    }, [navigate]);

    // Handle provider selection changes
    const handleProviderChange = (provider: string) => {
        setSelectedProvider(provider);
        localStorage.setItem(SELECTED_PROVIDER_KEY, provider);
    };

    // Check if selected provider is still valid whenever provider list changes
    useEffect(() => {
        // Skip initial render and only check when we have providers and a previously selected provider
        if (sortedProviders.length > 0 && selectedProvider && !sortedProviders.includes(selectedProvider)) {
            // If selected provider is no longer valid, mark for update
            needsProviderUpdateRef.current = true;
        }
    }, [sortedProviders, selectedProvider]);

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

    // Sync flowId with selectedFlowId when URL changes
    useLayoutEffect(() => {
        if (flowId && flowId !== previousFlowIdRef.current) {
            previousFlowIdRef.current = flowId;
            if (flowId !== selectedFlowId) {
                setSelectedFlowId(flowId);
            }
        }
    }, [flowId, selectedFlowId]);

    const variables = useMemo(() => ({ flowId: debouncedFlowId || '' }), [debouncedFlowId]);
    const skip = useMemo(() => !debouncedFlowId || debouncedFlowId === 'new', [debouncedFlowId]);

    // Global flow subscriptions - always active regardless of selected flow
    useFlowCreatedSubscription();
    useFlowDeletedSubscription();
    useFlowUpdatedSubscription();
    // Flow-specific subscriptions that depend on the selected flow
    useTaskCreatedSubscription({ variables, skip });
    useTaskUpdatedSubscription({ variables, skip });
    useScreenshotAddedSubscription({ variables, skip });
    useTerminalLogAddedSubscription({ variables, skip });
    useMessageLogUpdatedSubscription({ variables, skip });
    useMessageLogAddedSubscription({ variables, skip });
    useAgentLogAddedSubscription({ variables, skip });
    useSearchLogAddedSubscription({ variables, skip });
    useVectorStoreLogAddedSubscription({ variables, skip });

    // Assistant-specific subscriptions
    useAssistantCreatedSubscription({ variables, skip });
    useAssistantUpdatedSubscription({ variables, skip });
    useAssistantDeletedSubscription({ variables, skip });
    useAssistantLogAddedSubscription({ variables, skip });
    useAssistantLogUpdatedSubscription({ variables, skip });

    // Handle selecting an assistant
    const handleSelectAssistant = useCallback((assistantId: string | null) => {
        if (!selectedFlowId || selectedFlowId === 'new') return;

        setSelectedAssistantIds((prev) => ({
            ...prev,
            [selectedFlowId]: assistantId,
        }));
    }, [selectedFlowId]);

    // Function to handle initial assistant selection
    const selectInitialAssistant = useCallback(() => {
        if (!selectedFlowId || selectedFlowId === 'new' || !assistantsData?.assistants?.length) {
            return;
        }

        // Check if we've already made a selection for this flow (including null)
        // This prevents auto-selection after intentional deselection
        if (selectedFlowId in selectedAssistantIdsRef.current) {
            return;
        }

        // Select the first assistant (newest) instead of the last one (oldest)
        const firstAssistantId = assistantsData.assistants[0]?.id;
        if (firstAssistantId) {
            handleSelectAssistant(firstAssistantId);
        }
    }, [selectedFlowId, assistantsData?.assistants, handleSelectAssistant]);

    // Auto-select the last (newest) assistant when flow changes or assistants data updates
    useEffect(() => {
        if (!selectedFlowId || selectedFlowId === 'new' || !assistantsData?.assistants?.length) {
            return;
        }

        // Always select the first assistant (newest) when flow changes or assistants update
        // unless there's already a valid selected assistant for this flow
        const currentSelectedAssistantId = selectedAssistantIds[selectedFlowId];
        const firstAssistantId = assistantsData.assistants[0]?.id;
        
        // Check if current selection is still valid
        const isCurrentSelectionValid = currentSelectedAssistantId && 
            assistantsData.assistants.some(assistant => assistant.id === currentSelectedAssistantId);
        
        // Auto-select first assistant if:
        // 1. No assistant is currently selected for this flow, OR
        // 2. Currently selected assistant is no longer in the list (was deleted)
        if (!isCurrentSelectionValid && firstAssistantId) {
            handleSelectAssistant(firstAssistantId);
        }
    }, [selectedFlowId, assistantsData?.assistants, selectedAssistantIds, handleSelectAssistant]);

    // Trigger initial assistant selection when data changes (legacy support)
    useEffect(() => {
        selectInitialAssistant();
    }, [selectInitialAssistant]);

    // Set selectedAssistantId to null to initiate assistant creation
    const handleInitiateAssistantCreation = () => {
        if (!selectedFlowId || selectedFlowId === 'new') return;

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
            if (selectedProvider && sortedProviders.includes(selectedProvider)) {
                providerToUse = selectedProvider;
            } else if (sortedProviders.length > 0) {
                // If no provider is selected or it's invalid, try to use the first available
                const firstAvailableProvider = sortedProviders[0];
                if (typeof firstAvailableProvider === 'string') {
                    providerToUse = firstAvailableProvider;
                    // Mark for update rather than updating directly
                    needsProviderUpdateRef.current = true;
                }
            }

            // Fail early if we still don't have a provider
            if (!providerToUse) {
                Log.error('No valid provider available for creating a flow');
                return;
            }

            const { data } = await createFlow({
                variables: {
                    modelProvider: providerToUse,
                    input: message,
                },
            });

            if (data?.createFlow) {
                const newFlowId = data.createFlow.id.toString();

                // Navigate to the new flow page but stay on the automation tab
                navigate(`/chat/${newFlowId}`);

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

    const handleChangeSelectedFlowId = (id: string) => {
        setSelectedFlowId(id);
        navigate(`/chat/${id}`);
    };

    const handleDeleteFlow = async (id: string) => {
        try {
            // Store current state before deletion
            const wasCurrentFlow = String(selectedFlowId) === String(id);

            const result = await deleteFlow({
                variables: { flowId: id },
                refetchQueries: ['flows'],
                update: (cache) => {
                    // Remove the flow from Apollo cache
                    cache.evict({ id: `Flow:${id}` });
                    cache.gc();
                },
            });

            if (result.data?.deleteFlow === 'success') {
                // Force refresh the flows list
                await refetchFlows().catch((error) => {
                    Log.error('Failed to refetch flows:', error);
                });

                // If we deleted the currently selected flow, redirect to new flow page
                if (wasCurrentFlow) {
                    setSelectedFlowId('new');
                    navigate('/chat/new');
                }
            }
        } catch (error) {
            Log.error('Error deleting flow:', error);
        }
    };

    const handleFinishFlow = async (id: string) => {
        try {
            await finishFlow({ variables: { flowId: id } });
        } catch {
            // ignore
        }
    };

    // Handle creating a new assistant
    const handleCreateAssistant = useCallback(async (message: string, useAgents: boolean) => {
        if (!message.trim() || !selectedProvider) return;

        try {
            // Backend will create a flow if flowId is 'new' or '0'
            const flowIdToUse = selectedFlowId === 'new' ? '0' : selectedFlowId || '0';

            const { data } = await createAssistant({
                variables: {
                    flowId: flowIdToUse,
                    modelProvider: selectedProvider,
                    input: message.trim(),
                    useAgents,
                },
            });

            if (data?.createAssistant) {
                const { flow, assistant } = data.createAssistant;

                // If we created a new flow, navigate to it
                if (selectedFlowId === 'new' || !selectedFlowId) {
                    // Navigate to the new flow page
                    navigate(`/chat/${flow.id}`);

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
    }, [
        selectedFlowId,
        selectedProvider,
        createAssistant,
        navigate,
        isDesktop,
        setActiveCentralTab,
        setActiveTabsTab,
    ]);

    // Handle calling an existing assistant
    const handleCallAssistant = async (assistantId: string, message: string, useAgents: boolean): Promise<void> => {
        if (!selectedFlowId || selectedFlowId === 'new' || !assistantId || !message.trim()) return;

        try {
            await callAssistant({
                variables: {
                    flowId: selectedFlowId,
                    assistantId,
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
        if (!selectedFlowId || selectedFlowId === 'new' || !assistantId) return;

        try {
            await stopAssistant({
                variables: {
                    flowId: selectedFlowId,
                    assistantId,
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
        if (!selectedFlowId || selectedFlowId === 'new' || !assistantId) return;

        try {
            // Remember if this was the currently selected assistant
            const wasSelected = selectedAssistantId === assistantId;

            await deleteAssistant({
                variables: {
                    flowId: selectedFlowId,
                    assistantId,
                },
                update: (cache) => {
                    // Remove the assistant from Apollo cache
                    cache.evict({ id: `Assistant:${assistantId}` });
                    cache.gc();
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
                    flowData={flowData}
                    selectedFlowId={selectedFlowId}
                    onSubmitAutomationMessage={handleSubmitAutomationMessage}
                    onStopAutomationFlow={handleStopAutomationFlow}
                    assistants={assistantsData?.assistants ?? []}
                    assistantLogs={assistantLogsData?.assistantLogs ?? []}
                    selectedAssistantId={selectedAssistantId}
                    selectedProvider={selectedProvider ?? ''}
                    providers={sortedProviders}
                    onSelectAssistant={handleSelectAssistant}
                    onCreateAssistant={handleInitiateAssistantCreation}
                    onDeleteAssistant={handleDeleteAssistant}
                    onSubmitAssistantMessage={handleCallAssistant}
                    onCreateNewAssistant={handleCreateAssistant}
                    onStopAssistant={handleStopAssistant}
                    activeTab={activeTabsTab}
                    onTabChange={setActiveTabsTab}
                />
            </CardContent>
        </Card>
    );

    return (
        <SidebarProvider>
            <ChatSidebar
                user={user}
                flows={flowsData?.flows ?? []}
                providers={sortedProviders}
                selectedProvider={selectedProvider ?? ''}
                selectedFlowId={selectedFlowId}
                onChangeSelectedFlowId={handleChangeSelectedFlowId}
                onChangeSelectedProvider={handleProviderChange}
                onDeleteFlow={handleDeleteFlow}
                onFinishFlow={handleFinishFlow}
            />
            <SidebarInset>
                <header className="fixed top-0 z-10 flex h-12 w-full shrink-0 items-center gap-2 border-b bg-background transition-[width,height] ease-linear group-has-[[data-collapsible=icon]]/sidebar-wrapper:h-12">
                    <div className="flex items-center gap-2 px-4">
                        <SidebarTrigger className="-ml-1" />
                        <Separator
                            orientation="vertical"
                            className="mr-2 h-4"
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
                                        {flowData?.flow?.title || (selectedFlowId === 'new' ? 'New flow' : 'Select a flow')}
                                    </BreadcrumbPage>
                                </BreadcrumbItem>
                            </BreadcrumbList>
                        </Breadcrumb>
                    </div>
                </header>
                <div className="relative mt-12 flex h-[calc(100dvh-3rem)] w-full max-w-full flex-1">
                    {isFlowLoading && (
                        <div className="absolute inset-0 z-50 flex items-center justify-center bg-background/50">
                            <Loader2 className="size-16 animate-spin" />
                        </div>
                    )}
                    {isDesktop ? (
                        <ResizablePanelGroup
                            direction="horizontal"
                            className="w-full"
                        >
                            <ResizablePanel
                                defaultSize={50}
                                minSize={30}
                            >
                                <Card className="flex h-[calc(100dvh-3rem)] max-w-full flex-col rounded-none border-0">
                                    <CardContent className="flex-1 overflow-auto pr-0 pt-4">
                                        <ChatCentralTabs
                                            selectedFlowId={selectedFlowId}
                                            flowData={flowData}
                                            assistants={assistantsData?.assistants ?? []}
                                            assistantLogs={assistantLogsData?.assistantLogs ?? []}
                                            selectedAssistantId={selectedAssistantId}
                                            selectedProvider={selectedProvider ?? ''}
                                            providers={sortedProviders}
                                            onSelectAssistant={handleSelectAssistant}
                                            onCreateAssistant={handleInitiateAssistantCreation}
                                            onDeleteAssistant={handleDeleteAssistant}
                                            onSubmitAutomationMessage={handleSubmitAutomationMessage}
                                            onSubmitAssistantMessage={handleCallAssistant}
                                            onCreateNewAssistant={handleCreateAssistant}
                                            onStopAutomationFlow={handleStopAutomationFlow}
                                            onStopAssistant={handleStopAssistant}
                                            activeTab={activeCentralTab}
                                            onTabChange={setActiveCentralTab}
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
            </SidebarInset>
        </SidebarProvider>
    );
};

export default Chat;
