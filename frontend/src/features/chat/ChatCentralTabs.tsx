import { useState } from 'react';

import { ScrollArea, ScrollBar } from '@/components/ui/scroll-area';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';

import ChatAssistantMessages from './ChatAssistantMessages';
import ChatAutomationMessages from './ChatAutomationMessages';

const ChatCentralTabs = () => {
    const [activeTab, setActiveTab] = useState<string>('automation');

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
                <ChatAutomationMessages />
            </TabsContent>
            <TabsContent
                className="mt-2 flex-1 overflow-auto pr-4"
                value="assistant"
            >
                <ChatAssistantMessages />
            </TabsContent>
        </Tabs>
    );
};

export default ChatCentralTabs;
