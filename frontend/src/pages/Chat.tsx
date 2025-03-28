import { GripVertical, Loader2 } from 'lucide-react';
import { useEffect, useLayoutEffect, useMemo, useRef, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';

import { Breadcrumb, BreadcrumbItem, BreadcrumbList, BreadcrumbPage } from '@/components/ui/breadcrumb';
import { Card, CardContent, CardFooter } from '@/components/ui/card';
import { ResizableHandle, ResizablePanel, ResizablePanelGroup } from '@/components/ui/resizable';
import { Separator } from '@/components/ui/separator';
import { SidebarInset, SidebarProvider, SidebarTrigger } from '@/components/ui/sidebar';
import ChatForm from '@/features/chat/ChatForm';
import ChatMessages from '@/features/chat/ChatMessages';
import ChatSidebar from '@/features/chat/ChatSidebar';
import ChatTabs from '@/features/chat/ChatTabs';
import {
    StatusType,
    useAgentLogAddedSubscription,
    useCreateFlowMutation,
    useDeleteFlowMutation,
    useFinishFlowMutation,
    useFlowQuery,
    useFlowsQuery,
    useFlowUpdatedSubscription,
    useMessageLogAddedSubscription,
    useMessageLogUpdatedSubscription,
    useProvidersQuery,
    usePutUserInputMutation,
    useScreenshotAddedSubscription,
    useSearchLogAddedSubscription,
    useTaskCreatedSubscription,
    useTaskUpdatedSubscription,
    useTerminalLogAddedSubscription,
    useVectorStoreLogAddedSubscription,
} from '@/graphql/types';
import { useBreakpoint } from '@/hooks/use-breakpoint';
import { client } from '@/lib/apollo';
import { Log } from '@/lib/log';
import type { User } from '@/models/User';

const SELECTED_PROVIDER_KEY = 'selectedProvider';

const Chat = () => {
    const { flowId } = useParams();
    const navigate = useNavigate();
    const { data: flowsData, refetch: refetchFlows } = useFlowsQuery();
    const [selectedFlowId, setSelectedFlowId] = useState<string | null>(flowId ?? null);
    const { data: flowData, loading: isFlowLoading, refetch: refetchFlow } = useFlowQuery({
        variables: { id: selectedFlowId ?? '' },
        skip: !selectedFlowId || selectedFlowId === 'new',
        fetchPolicy: 'cache-and-network',
        nextFetchPolicy: 'cache-first',
        notifyOnNetworkStatusChange: false,
    });
    const { data: providersData } = useProvidersQuery();

    // Create sorted providers list to ensure consistent order
    const sortedProviders = useMemo(() => {
        const providers = providersData?.providers || [];
        return [...providers].sort();
    }, [providersData?.providers]);

    const [selectedProvider, setSelectedProvider] = useState<string | null>(() => {
        const savedProvider = localStorage.getItem(SELECTED_PROVIDER_KEY);
        return savedProvider || null;
    });
    const [user, setUser] = useState<User | null>(null);
    const { isDesktop } = useBreakpoint();
    const needsProviderUpdateRef = useRef(false);
    const needsUserUpdateRef = useRef(false);
    const userDataRef = useRef<User | null>(null);
    const previousFlowIdRef = useRef(flowId);

    useEffect(() => {
        const auth = localStorage.getItem('auth');

        if (!auth) {
            navigate('/login');
            return;
        }

        const user = JSON.parse(auth)?.user;

        if (!user) {
            navigate('/login');
            return;
        }

        needsUserUpdateRef.current = true;
        userDataRef.current = user;
    }, [navigate]);

    useLayoutEffect(() => {
        if (needsUserUpdateRef.current && userDataRef.current) {
            needsUserUpdateRef.current = false;
            setUser(userDataRef.current);
        }
    }, []);

    // Handle provider selection changes
    const handleProviderChange = (provider: string) => {
        setSelectedProvider(provider);
        localStorage.setItem(SELECTED_PROVIDER_KEY, provider);
    };

    // Initialize provider only if none is selected and providers are available
    useEffect(() => {
        if (sortedProviders.length > 0 && !selectedProvider) {
            needsProviderUpdateRef.current = true;
        }
    }, [sortedProviders, selectedProvider]);

    // Update provider in layout effect
    useLayoutEffect(() => {
        if (needsProviderUpdateRef.current && sortedProviders.length > 0) {
            needsProviderUpdateRef.current = false;
            const firstProvider = sortedProviders[0];
            if (firstProvider) {
                setSelectedProvider(firstProvider);
                localStorage.setItem(SELECTED_PROVIDER_KEY, firstProvider);
            }
        }
    }, [sortedProviders]);

    const [createFlow] = useCreateFlowMutation();
    const [deleteFlow] = useDeleteFlowMutation();
    const [finishFlow] = useFinishFlowMutation();
    const [putUserInput] = usePutUserInputMutation();

    // Sync flowId with selectedFlowId when URL changes
    useLayoutEffect(() => {
        if (flowId && flowId !== previousFlowIdRef.current) {
            previousFlowIdRef.current = flowId;
            if (flowId !== selectedFlowId) {
                setSelectedFlowId(flowId);
            }
        }
    }, [flowId, selectedFlowId]);

    // Keep this useEffect separate for data fetching
    useEffect(() => {
        // Only refetch if we have a valid flowId and it's not a new flow
        if (selectedFlowId && selectedFlowId !== 'new' && !isFlowLoading) {
            refetchFlow().catch((error) => {
                Log.error('Failed to refetch flow data:', error);
            });
        }
    }, [selectedFlowId, isFlowLoading, refetchFlow]);

    // Set up GraphQL subscriptions with proper dependency tracking
    useEffect(() => {
        // Don't set up subscriptions for 'new' flows or when we don't have a flow ID
        if (!selectedFlowId || selectedFlowId === 'new') {
            return;
        }
        // Log subscription setup for debugging
        Log.debug(`Setting up subscriptions for flow ID: ${selectedFlowId}`);
    }, [selectedFlowId]);

    const variables = useMemo(() => ({ flowId: selectedFlowId || '' }), [selectedFlowId]);
    const skip = useMemo(() => !selectedFlowId || selectedFlowId === 'new', [selectedFlowId]);

    // Use the memoized variables and skip value for all subscriptions
    useFlowUpdatedSubscription({ variables, skip });
    useTaskCreatedSubscription({ variables, skip });
    useTaskUpdatedSubscription({ variables, skip });
    useScreenshotAddedSubscription({ variables, skip });
    useTerminalLogAddedSubscription({ variables, skip });
    useMessageLogUpdatedSubscription({ variables, skip });
    useMessageLogAddedSubscription({ variables, skip });
    useAgentLogAddedSubscription({ variables, skip });
    useSearchLogAddedSubscription({ variables, skip });
    useVectorStoreLogAddedSubscription({ variables, skip });

    const handleSubmit = async (message: string) => {
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

            const { data } = await createFlow({
                variables: {
                    modelProvider: selectedProvider ?? '',
                    input: message,
                },
            });

            if (data?.createFlow) {
                // Update flows data after creating a new flow
                const newFlowId = data.createFlow.id.toString();
                Log.debug(`Created new flow with ID: ${newFlowId}`);
                // Force refresh cache for the new flow
                client.cache.evict({ fieldName: 'tasks' });
                // Navigate to the new flow page
                navigate(`/chat/${newFlowId}`);
            }
        } catch (error) {
            Log.error('Error submitting message:', error);
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

    const tabsCard = (
        <Card className="flex h-[calc(100dvh-3rem)] max-w-full flex-col rounded-none border-0">
            <CardContent className="flex-1 overflow-auto pr-0 pt-4">
                <ChatTabs
                    flowData={flowData}
                    selectedFlowId={selectedFlowId}
                    onMessageSubmit={handleSubmit}
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
                                <Card className="flex h-[calc(100dvh-3rem)] flex-col rounded-none border-0">
                                    <CardContent className="flex-1 overflow-y-auto pb-0">
                                        <ChatMessages logs={flowData?.messageLogs ?? []} />
                                    </CardContent>
                                    <CardFooter className="sticky bottom-0 border-t bg-background pt-4">
                                        <ChatForm
                                            selectedFlowId={selectedFlowId}
                                            flowStatus={flowData?.flow?.status}
                                            onSubmit={handleSubmit}
                                        />
                                    </CardFooter>
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
