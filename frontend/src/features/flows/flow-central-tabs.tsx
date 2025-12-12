import { useEffect, useState } from 'react';

import { ScrollArea, ScrollBar } from '@/components/ui/scroll-area';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import FlowAssistantMessages from '@/features/flows/messages/flow-assistant-messages';
import FlowAutomationMessages from '@/features/flows/messages/flow-automation-messages';
import { useFlow } from '@/providers/flow-provider';

const FlowCentralTabs = () => {
    const { flowData, isLoading } = useFlow();
    const [activeTab, setActiveTab] = useState<string>('automation');

    // Switch to assistant tab if flow is loaded and messageLogs are empty
    useEffect(() => {
        if (!isLoading && !flowData?.messageLogs?.length) {
            setActiveTab('assistant');
        }
    }, [isLoading, flowData?.messageLogs]);

    return (
        <Tabs
            className="flex size-full flex-col"
            onValueChange={setActiveTab}
            value={activeTab}
        >
            <div className="max-w-full">
                <ScrollArea className="w-full pb-2">
                    <TabsList className="flex w-fit">
                        <TabsTrigger value="automation">Automation</TabsTrigger>
                        <TabsTrigger value="assistant">Assistant</TabsTrigger>
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
        </Tabs>
    );
};

export default FlowCentralTabs;
