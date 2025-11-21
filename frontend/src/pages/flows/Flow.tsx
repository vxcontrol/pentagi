import { GripVertical, Loader2 } from 'lucide-react';
import { useMemo, useState } from 'react';

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
    useAgentLogAddedSubscription,
    useAssistantCreatedSubscription,
    useAssistantDeletedSubscription,
    useAssistantLogAddedSubscription,
    useAssistantLogUpdatedSubscription,
    useAssistantUpdatedSubscription,
    useMessageLogAddedSubscription,
    useMessageLogUpdatedSubscription,
    useScreenshotAddedSubscription,
    useSearchLogAddedSubscription,
    useTaskCreatedSubscription,
    useTaskUpdatedSubscription,
    useTerminalLogAddedSubscription,
    useVectorStoreLogAddedSubscription,
} from '@/graphql/types';
import { useBreakpoint } from '@/hooks/use-breakpoint';
import { useFlow } from '@/providers/FlowProvider';

const Flow = () => {
    const { isDesktop } = useBreakpoint();

    // Get flow data from FlowProvider
    const { flowData, flowId, isLoading: isFlowLoading } = useFlow();

    // State for preserving active tabs when switching flows
    const [activeCentralTab, setActiveCentralTab] = useState<string>('automation');
    const [activeTabsTab, setActiveTabsTab] = useState<string>(!isDesktop ? 'automation' : 'terminal');

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

    const tabsCard = (
        <Card className="flex h-[calc(100dvh-3rem)] max-w-full flex-col rounded-none border-0">
            <CardContent className="flex-1 overflow-auto pr-0 pt-4">
                <ChatTabs
                    activeTab={activeTabsTab}
                    onTabChange={setActiveTabsTab}
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
        </>
    );
};

export default Flow;
