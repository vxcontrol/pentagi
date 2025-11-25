import { GripVertical, Loader2 } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';

import { FlowStatusIcon } from '@/components/icons/flow-status-icon';
import { ProviderIcon } from '@/components/icons/provider-icon';
import { Breadcrumb, BreadcrumbItem, BreadcrumbList, BreadcrumbPage } from '@/components/ui/breadcrumb';
import { Card, CardContent } from '@/components/ui/card';
import { ResizableHandle, ResizablePanel, ResizablePanelGroup } from '@/components/ui/resizable';
import { Separator } from '@/components/ui/separator';
import { SidebarTrigger } from '@/components/ui/sidebar';
import FlowCentralTabs from '@/features/flows/flow-central-tabs';
import FlowTabs from '@/features/flows/flow-tabs';
import { useBreakpoint } from '@/hooks/use-breakpoint';
import { formatName } from '@/lib/utils/format';
import { useFlow } from '@/providers/flow-provider';

const Flow = () => {
    const { isDesktop } = useBreakpoint();
    const navigate = useNavigate();

    // Get flow data from FlowProvider
    const { flowData, flowError, isLoading: isFlowLoading } = useFlow();

    // Redirect to flows list if there's an error loading flow data or flow not found
    useEffect(() => {
        if (flowError || (!isFlowLoading && !flowData?.flow)) {
            navigate('/flows', { replace: true });
        }
    }, [flowError, flowData, isFlowLoading, navigate]);

    // State for preserving active tabs when switching flows
    const [activeTabsTab, setActiveTabsTab] = useState<string>(!isDesktop ? 'automation' : 'terminal');

    const tabsCard = (
        <Card className="flex h-[calc(100dvh-3rem)] max-w-full flex-col rounded-none border-0">
            <CardContent className="flex-1 overflow-auto pr-0 pt-4">
                <FlowTabs
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
                                <BreadcrumbItem className="gap-2">
                                    {flowData?.flow && (
                                        <>
                                            <FlowStatusIcon
                                                status={flowData.flow.status}
                                                tooltip={formatName(flowData.flow.status)}
                                            />

                                            <ProviderIcon
                                                provider={flowData.flow.provider}
                                                tooltip={formatName(flowData.flow.provider.name)}
                                            />
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
                                    <FlowCentralTabs />
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
