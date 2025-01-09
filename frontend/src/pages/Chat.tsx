import { GripVertical, Loader2 } from 'lucide-react';
import { useEffect, useState } from 'react';
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
    useTaskUpdatedSubscription,
    useTerminalLogAddedSubscription,
    useVectorStoreLogAddedSubscription,
} from '@/graphql/types';
import { useBreakpoint } from '@/hooks/use-breakpoint';
import type { User } from '@/models/User';

const Chat = () => {
    const { flowId } = useParams();
    const navigate = useNavigate();
    const { data: flowsData } = useFlowsQuery();
    const [selectedFlowId, setSelectedFlowId] = useState<string | null>(flowId ?? null);
    const { data: flowData, loading: isFlowLoading } = useFlowQuery({
        variables: { id: selectedFlowId ?? '' },
        skip: !selectedFlowId || selectedFlowId === 'new',
    });
    const { data: providersData } = useProvidersQuery();
    const [selectedProvider, setSelectedProvider] = useState<string | null>(null);
    const [user, setUser] = useState<User | null>(null);
    const { isDesktop } = useBreakpoint();

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

        setUser(user);
    }, [navigate]);

    useEffect(() => {
        if (providersData?.providers?.[0]) {
            setSelectedProvider(providersData.providers[0]);
        }
    }, [providersData]);

    const [createFlow] = useCreateFlowMutation();
    const [deleteFlow] = useDeleteFlowMutation();
    const [finishFlow] = useFinishFlowMutation();
    const [putUserInput] = usePutUserInputMutation();

    const variables = { flowId: selectedFlowId || '' };
    const skip = !selectedFlowId || selectedFlowId === 'new';

    useFlowUpdatedSubscription({
        variables,
        skip,
    });

    useTaskUpdatedSubscription({
        variables,
        skip,
    });

    useScreenshotAddedSubscription({
        variables,
        skip,
    });

    useTerminalLogAddedSubscription({
        variables,
        skip,
    });

    useMessageLogUpdatedSubscription({
        variables,
        skip,
    });

    useMessageLogAddedSubscription({
        variables,
        skip,
    });

    useAgentLogAddedSubscription({
        variables,
        skip,
    });

    useSearchLogAddedSubscription({
        variables,
        skip,
    });

    useVectorStoreLogAddedSubscription({
        variables,
        skip,
    });

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
                navigate(`/chat/${data.createFlow.id}`);
            }
        } catch {
            // ignore
        }
    };

    const handleChangeSelectedFlowId = (id: string) => {
        setSelectedFlowId(id);
        navigate(`/chat/${id}`);
    };

    useEffect(() => {
        if (flowId && flowId !== selectedFlowId) {
            setSelectedFlowId(flowId);
        }
    }, [flowId]);

    const handleDeleteFlow = async (id: string) => {
        try {
            await deleteFlow({ variables: { flowId: id } });
            if (selectedFlowId === id) {
                setSelectedFlowId(null);
                navigate('/chat');
            }
        } catch {
            // ignore
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
                providers={providersData?.providers ?? []}
                selectedProvider={selectedProvider ?? ''}
                selectedFlowId={selectedFlowId}
                onChangeSelectedFlowId={handleChangeSelectedFlowId}
                onChangeSelectedProvider={setSelectedProvider}
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
                                        {flowData?.flow?.title ||
                                            (selectedFlowId === 'new' ? 'New flow' : 'Select a flow')}
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
