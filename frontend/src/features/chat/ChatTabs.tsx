import { useState } from 'react';

import { ScrollArea, ScrollBar } from '@/components/ui/scroll-area';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import type { FlowQuery } from '@/graphql/types';
import { useBreakpoint } from '@/hooks/use-breakpoint';

import ChatAgents from './ChatAgents';
import ChatForm from './ChatForm';
import ChatMessages from './ChatMessages';
import ChatScreenshots from './ChatScreenshots';
import ChatTasks from './ChatTasks';
import ChatTerminal from './ChatTerminal';
import ChatTools from './ChatTools';
import ChatVectorStores from './ChatVectorStores';

interface ChatTabsProps {
    flowData: FlowQuery | undefined;
    selectedFlowId: string | null;
    onMessageSubmit: (message: string) => Promise<void>;
}

const ChatTabs = ({ flowData, selectedFlowId, onMessageSubmit }: ChatTabsProps) => {
    const { isDesktop } = useBreakpoint();
    const [activeTab, setActiveTab] = useState<string>(!isDesktop ? 'messages' : 'terminal');

    return (
        <Tabs
            value={activeTab}
            onValueChange={setActiveTab}
            className="flex size-full flex-col"
        >
            <div className="max-w-full pr-4">
                <ScrollArea className="w-full pb-2">
                    <TabsList className="flex w-fit">
                        {!isDesktop && <TabsTrigger value="messages">Messages</TabsTrigger>}
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

            {!isDesktop && (
                <TabsContent
                    value="messages"
                    className="-mb-4 -ml-4 mt-2 flex-1 flex-col overflow-hidden data-[state=active]:flex"
                >
                    <div className="-mt-4 flex flex-1 flex-col overflow-y-auto">
                        <div className="flex-1">
                            <ChatMessages
                                logs={flowData?.messageLogs ?? []}
                                className="px-4"
                            />
                        </div>
                        <div className="sticky bottom-0 border-t bg-background p-4">
                            <ChatForm
                                selectedFlowId={selectedFlowId}
                                flowStatus={flowData?.flow?.status}
                                onSubmit={onMessageSubmit}
                            />
                        </div>
                    </div>
                </TabsContent>
            )}

            <TabsContent
                value="terminal"
                className="mt-2 flex-1 overflow-auto"
            >
                <ChatTerminal logs={flowData?.terminalLogs ?? []} />
            </TabsContent>

            <TabsContent
                value="tasks"
                className="mt-2 flex-1 overflow-auto pr-4"
            >
                <ChatTasks tasks={flowData?.tasks ?? []} />
            </TabsContent>

            <TabsContent
                value="agents"
                className="mt-2 flex-1 overflow-auto pr-4"
            >
                <ChatAgents logs={flowData?.agentLogs ?? []} />
            </TabsContent>

            <TabsContent
                value="tools"
                className="mt-2 flex-1 overflow-auto pr-4"
            >
                <ChatTools logs={flowData?.searchLogs ?? []} />
            </TabsContent>

            <TabsContent
                value="vectorStores"
                className="mt-2 flex-1 overflow-auto pr-4"
            >
                <ChatVectorStores logs={flowData?.vectorStoreLogs ?? []} />
            </TabsContent>

            <TabsContent
                value="screenshots"
                className="mt-2 flex-1 overflow-auto pr-4"
            >
                <ChatScreenshots screenshots={flowData?.screenshots ?? []} />
            </TabsContent>
        </Tabs>
    );
};

export default ChatTabs;
