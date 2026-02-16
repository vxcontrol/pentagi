import { useMemo, useState } from 'react';
import { useSearchParams } from 'react-router-dom';

import { ScrollArea, ScrollBar } from '@/components/ui/scroll-area';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import FlowDashboard from '@/features/flows/dashboard/flow-dashboard';
import FlowAssistantMessages from '@/features/flows/messages/flow-assistant-messages';
import FlowAutomationMessages from '@/features/flows/messages/flow-automation-messages';
import { useFlow } from '@/providers/flow-provider';

const FlowCentralTabs = () => {
    const { flowData, isLoading } = useFlow();
    const [searchParams, setSearchParams] = useSearchParams();
    const [activeTab, setActiveTab] = useState<null | string>(null);

    // Determine default tab based on priority: manual selection > URL parameter > auto-detection
    const defaultTab = useMemo(() => {
        // If user manually selected a tab, use it
        if (activeTab) {
            return activeTab;
        }

        // Check URL parameter
        const tabParam = searchParams.get('tab');

        if (tabParam === 'automation' || tabParam === 'assistant' || tabParam === 'dashboard') {
            return tabParam;
        }

        // Auto-detect: switch to assistant tab if flow is loaded and messageLogs are empty
        if (!isLoading && !flowData?.messageLogs?.length) {
            return 'assistant';
        }

        return 'automation';
    }, [activeTab, searchParams, isLoading, flowData?.messageLogs]);

    // Handle tab change - update both state and URL parameter
    const handleTabChange = (tab: string) => {
        setActiveTab(tab);
        setSearchParams({ tab });
    };

    return (
        <Tabs
            className="flex size-full flex-col"
            onValueChange={handleTabChange}
            value={defaultTab}
        >
            <div className="max-w-full">
                <ScrollArea className="w-full pb-2">
                    <TabsList className="flex w-fit">
                        <TabsTrigger value="automation">Automation</TabsTrigger>
                        <TabsTrigger value="assistant">Assistant</TabsTrigger>
                        <TabsTrigger value="dashboard">Dashboard</TabsTrigger>
                    </TabsList>
                    <ScrollBar orientation="horizontal" />
                </ScrollArea>
            </div>

            <TabsContent
                className="mt-2 flex-1 overflow-auto pr-4"
                value="automation"
            >
                <FlowAutomationMessages />
            </TabsContent>
            <TabsContent
                className="mt-2 flex-1 overflow-auto pr-4"
                value="assistant"
            >
                <FlowAssistantMessages />
            </TabsContent>
            <TabsContent
                className="mt-2 flex-1 overflow-auto pr-4"
                value="dashboard"
            >
                <FlowDashboard />
            </TabsContent>
        </Tabs>
    );
};

export default FlowCentralTabs;
