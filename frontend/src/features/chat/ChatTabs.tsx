import type { Dispatch, SetStateAction } from 'react';

import { useEffect, useRef } from 'react';

import { ScrollArea, ScrollBar } from '@/components/ui/scroll-area';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { useBreakpoint } from '@/hooks/use-breakpoint';

import ChatAgents from './ChatAgents';
import ChatAssistantMessages from './ChatAssistantMessages';
import ChatAutomationMessages from './ChatAutomationMessages';
import ChatScreenshots from './ChatScreenshots';
import ChatTasks from './ChatTasks';
import ChatTerminal from './ChatTerminal';
import ChatTools from './ChatTools';
import ChatVectorStores from './ChatVectorStores';

interface ChatTabsProps {
    activeTab: string;
    onTabChange: Dispatch<SetStateAction<string>>;
}

const ChatTabs = ({ activeTab, onTabChange }: ChatTabsProps) => {
    const { isDesktop } = useBreakpoint();

    const previousActiveTabRef = useRef<string>(activeTab);

    useEffect(() => {
        // Only handle actual tab changes
        if (activeTab === previousActiveTabRef.current) {
            return;
        }

        previousActiveTabRef.current = activeTab;
    }, [activeTab]);

    return (
        <Tabs
            className="flex size-full flex-col"
            onValueChange={onTabChange}
            value={activeTab}
        >
            <div className="max-w-full pr-4">
                <ScrollArea className="w-full pb-2">
                    <TabsList className="flex w-fit">
                        {!isDesktop && <TabsTrigger value="automation">Automation</TabsTrigger>}
                        {!isDesktop && <TabsTrigger value="assistant">Assistant</TabsTrigger>}
                        <TabsTrigger value="terminal">Terminal</TabsTrigger>
                        <TabsTrigger value="tasks">Tasks</TabsTrigger>
                        <TabsTrigger value="agents">Agents</TabsTrigger>
                        <TabsTrigger value="tools">Searches</TabsTrigger>
                        <TabsTrigger value="vectorStores">Vector Store</TabsTrigger>
                        <TabsTrigger value="screenshots">Screenshots</TabsTrigger>
                    </TabsList>
                    <ScrollBar orientation="horizontal" />
                </ScrollArea>
            </div>

            {/* Mobile Tabs only */}
            {!isDesktop && (
                <TabsContent
                    className="mt-2 flex-1 overflow-auto"
                    value="automation"
                >
                    <ChatAutomationMessages className="pr-4" />
                </TabsContent>
            )}
            {!isDesktop && (
                <TabsContent
                    className="mt-2 flex-1 overflow-auto"
                    value="assistant"
                >
                    <ChatAssistantMessages className="pr-4" />
                </TabsContent>
            )}

            {/* Desktop and Mobile Tabs */}
            <TabsContent
                className="mt-2 flex-1 overflow-auto"
                value="terminal"
            >
                <ChatTerminal />
            </TabsContent>

            <TabsContent
                className="mt-2 flex-1 overflow-auto pr-4"
                value="tasks"
            >
                <ChatTasks />
            </TabsContent>

            <TabsContent
                className="mt-2 flex-1 overflow-auto pr-4"
                value="agents"
            >
                <ChatAgents />
            </TabsContent>

            <TabsContent
                className="mt-2 flex-1 overflow-auto pr-4"
                value="tools"
            >
                <ChatTools />
            </TabsContent>

            <TabsContent
                className="mt-2 flex-1 overflow-auto pr-4"
                value="vectorStores"
            >
                <ChatVectorStores />
            </TabsContent>

            <TabsContent
                className="mt-2 flex-1 overflow-auto pr-4"
                value="screenshots"
            >
                <ChatScreenshots />
            </TabsContent>
        </Tabs>
    );
};

export default ChatTabs;
