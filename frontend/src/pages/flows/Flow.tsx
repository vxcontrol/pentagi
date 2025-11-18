import { GripVertical, Loader2 } from 'lucide-react';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useParams } from 'react-router-dom';
import { toast } from 'sonner';

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
import { useProviders } from '@/providers/ProvidersProvider';

const Flow = () => {
    const { isDesktop } = useBreakpoint();

    // Get flowId directly from URL params
    const { flowId } = useParams();

    // Get providers data from ProvidersProvider
    const { providers, selectedProvider } = useProviders();

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
        skip: !flowId,
        variables: { id: flowId ?? '' },
    });

    // Store currently selected assistant ID for each flow
    const [selectedAssistantIds, setSelectedAssistantIds] = useState<Record<string, null | string>>({});

    // Get assistants for the current flow
    const { data: assistantsData } = useAssistantsQuery({
        fetchPolicy: 'cache-and-network',
        nextFetchPolicy: 'cache-first',
        skip: !flowId,
        variables: { flowId: flowId ?? '' },
    });

    // Currently selected assistant for this flow (with auto-selection logic)
    const selectedAssistantId = useMemo(() => {
        if (!flowId) {
            return null;
        }

        // If we have an explicit selection for this flow, use it
        if (flowId in selectedAssistantIds) {
            const explicitSelection = selectedAssistantIds[flowId];

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
    }, [flowId, selectedAssistantIds, assistantsData?.assistants]);

    // Store for API mutations
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
        skip: !flowId || !selectedAssistantId || selectedAssistantId === '',
        variables: { assistantId: selectedAssistantId ?? '', flowId: flowId ?? '' },
    });

    // Additional cleanup on component unmount to prevent memory leaks
    useEffect(() => {
        return () => {
            if (assistantCreationTimeoutRef.current) {
                clearTimeout(assistantCreationTimeoutRef.current);
                assistantCreationTimeoutRef.current = null;
            }
        };
    }, []);

    const variables = useMemo(() => ({ flowId: flowId || '' }), [flowId]);
    const skip = useMemo(() => !flowId, [flowId]);

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
            if (!flowId) {
                return;
            }

            setSelectedAssistantIds((prev) => ({
                ...prev,
                [flowId]: assistantId,
            }));
        },
        [flowId],
    );

    // Set selectedAssistantId to null to initiate assistant creation
    const handleInitiateAssistantCreation = () => {
        if (!flowId) {
            return;
        }

        handleSelectAssistant(null);
    };

    const handleSubmitAutomationMessage = async (message: string) => {
        if (!flowId || flowData?.flow?.status === StatusType.Finished) {
            return;
        }

        try {
            await putUserInput({
                variables: {
                    flowId: flowId ?? '',
                    input: message,
                },
            });
        } catch (error) {
            const description = error instanceof Error ? error.message : 'An error occurred while submitting message';
            toast.error('Failed to submit message', {
                description,
            });
            Log.error('Error submitting message:', error);
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
            const description = error instanceof Error ? error.message : 'An error occurred while stopping flow';
            toast.error('Failed to stop flow', {
                description,
            });
            Log.error('Error stopping flow:', error);
        }
    };

    // Handle creating a new assistant
    const handleCreateAssistant = useCallback(
        async (message: string, useAgents: boolean) => {
            if (!message.trim() || !selectedProvider || !flowId) {
                return;
            }

            try {
                const { data } = await createAssistant({
                    variables: {
                        flowId,
                        input: message.trim(),
                        modelProvider: selectedProvider.name,
                        useAgents,
                    },
                });

                if (data?.createAssistant) {
                    const { assistant, flow } = data.createAssistant;

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
                const description =
                    error instanceof Error ? error.message : 'An error occurred while creating assistant';
                toast.error('Failed to create assistant', {
                    description,
                });
                Log.error('Error creating assistant:', error);
            }
        },
        [flowId, selectedProvider, createAssistant],
    );

    // Handle calling an existing assistant
    const handleCallAssistant = async (assistantId: string, message: string, useAgents: boolean): Promise<void> => {
        if (!flowId || !assistantId || !message.trim()) {
            return;
        }

        try {
            await callAssistant({
                variables: {
                    assistantId,
                    flowId,
                    input: message.trim(),
                    useAgents,
                },
            });

            // Refresh assistant logs after calling - only if we have a selected assistant
            if (selectedAssistantId) {
                refetchAssistantLogs();
            }
        } catch (error) {
            const description = error instanceof Error ? error.message : 'An error occurred while calling assistant';
            toast.error('Failed to call assistant', {
                description,
            });
            Log.error('Error calling assistant:', error);
        }
    };

    // Handle stopping an assistant
    const handleStopAssistant = async (assistantId: string) => {
        if (!flowId || !assistantId) {
            return;
        }

        try {
            await stopAssistant({
                variables: {
                    assistantId,
                    flowId,
                },
            });

            // Force refresh assistant data after stopping - only if we have a selected assistant
            if (selectedAssistantId) {
                refetchAssistantLogs();
            }
        } catch (error) {
            const description = error instanceof Error ? error.message : 'An error occurred while stopping assistant';
            toast.error('Failed to stop assistant', {
                description,
            });
            Log.error('Error stopping assistant:', error);
        }
    };

    // Handle deleting an assistant
    const handleDeleteAssistant = async (assistantId: string) => {
        if (!flowId || !assistantId) {
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
                    flowId,
                },
            });

            // If we deleted the currently selected assistant, reset selection
            if (wasSelected) {
                setSelectedAssistantIds((prev) => {
                    const updatedState = { ...prev };
                    updatedState[flowId] = null;

                    return updatedState;
                });
            }
        } catch (error) {
            const description = error instanceof Error ? error.message : 'An error occurred while deleting assistant';
            toast.error('Failed to delete assistant', {
                description,
            });
            Log.error('Error deleting assistant:', error);
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
                    providers={providers}
                    selectedAssistantId={selectedAssistantId}
                    selectedFlowId={flowId ?? null}
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
                                    <BreadcrumbPage>{flowData?.flow?.title || 'Select a flow'}</BreadcrumbPage>
                                </BreadcrumbItem>
                            </BreadcrumbList>
                        </Breadcrumb>
                    </div>
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
                                        providers={providers}
                                        selectedAssistantId={selectedAssistantId}
                                        selectedFlowId={flowId ?? null}
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

export default Flow;
