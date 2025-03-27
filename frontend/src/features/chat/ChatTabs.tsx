import { memo, useEffect, useMemo, useRef, useState } from 'react';

import { ScrollArea, ScrollBar } from '@/components/ui/scroll-area';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import type { FlowQuery } from '@/graphql/types';
import { useBreakpoint } from '@/hooks/use-breakpoint';
import { client } from '@/lib/apollo';

import ChatAgents from './ChatAgents';
import ChatForm from './ChatForm';
import ChatMessages from './ChatMessages';
import ChatScreenshots from './ChatScreenshots';
import ChatTasks from './ChatTasks';
import ChatTerminal from './ChatTerminal';
import ChatTools from './ChatTools';
import ChatVectorStores from './ChatVectorStores';

// Memoized components to prevent unnecessary re-renders
const MemoizedChatTerminal = memo(ChatTerminal);
const MemoizedChatTasks = memo(ChatTasks);
const MemoizedChatAgents = memo(ChatAgents);
const MemoizedChatTools = memo(ChatTools);
const MemoizedChatVectorStores = memo(ChatVectorStores);
const MemoizedChatScreenshots = memo(ChatScreenshots);
const MemoizedChatMessages = memo(ChatMessages);

interface ChatTabsProps {
    flowData: FlowQuery | undefined;
    selectedFlowId: string | null;
    onMessageSubmit: (message: string) => Promise<void>;
}

const ChatTabs = ({ flowData, selectedFlowId, onMessageSubmit }: ChatTabsProps) => {
    const { isDesktop } = useBreakpoint();
    const [activeTab, setActiveTab] = useState<string>(!isDesktop ? 'messages' : 'terminal');
    const previousActiveTabRef = useRef<string>(activeTab);
    const dataRefreshTimeoutRef = useRef<NodeJS.Timeout | null>(null);

    // Only refetch data when switching to tasks tab if we haven't done so recently
    useEffect(() => {
        // Only handle actual tab changes
        if (activeTab === previousActiveTabRef.current) {
            return;
        }

        previousActiveTabRef.current = activeTab;

        // When switching to tasks tab, quietly refresh data in the background
        if (activeTab === 'tasks' && selectedFlowId && selectedFlowId !== 'new') {
            // Clear any pending refresh
            if (dataRefreshTimeoutRef.current) {
                clearTimeout(dataRefreshTimeoutRef.current);
            }

            // Delay the refresh slightly to avoid UI jank during tab transition
            dataRefreshTimeoutRef.current = setTimeout(() => {
                // Silently refetch without triggering loading states
                client.refetchQueries({
                    include: ['flow'],
                });
            }, 100);
        }

        return () => {
            if (dataRefreshTimeoutRef.current) {
                clearTimeout(dataRefreshTimeoutRef.current);
            }
        };
    }, [activeTab, selectedFlowId]);

    // Load tasks data once when a new flow ID is selected
    useEffect(() => {
        if (selectedFlowId && selectedFlowId !== 'new') {
            // Silently clear the tasks cache when flow ID changes
            client.cache.evict({ fieldName: 'tasks' });

            // Refresh in the background
            client.refetchQueries({
                include: ['flow'],
            });
        }
    }, [selectedFlowId]);

    // Cache data references to prevent unnecessary re-renders
    const terminalLogs = useMemo(() => flowData?.terminalLogs ?? [], [flowData?.terminalLogs]);
    const messageLogs = useMemo(() => flowData?.messageLogs ?? [], [flowData?.messageLogs]);
    const tasks = useMemo(() => flowData?.tasks ?? [], [flowData?.tasks]);
    const agentLogs = useMemo(() => flowData?.agentLogs ?? [], [flowData?.agentLogs]);
    const searchLogs = useMemo(() => flowData?.searchLogs ?? [], [flowData?.searchLogs]);
    const vectorStoreLogs = useMemo(() => flowData?.vectorStoreLogs ?? [], [flowData?.vectorStoreLogs]);
    const screenshots = useMemo(() => flowData?.screenshots ?? [], [flowData?.screenshots]);

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
                            <MemoizedChatMessages
                                logs={messageLogs}
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
                <MemoizedChatTerminal logs={terminalLogs} />
            </TabsContent>

            <TabsContent
                value="tasks"
                className="mt-2 flex-1 overflow-auto pr-4"
            >
                <MemoizedChatTasks tasks={tasks} />
            </TabsContent>

            <TabsContent
                value="agents"
                className="mt-2 flex-1 overflow-auto pr-4"
            >
                <MemoizedChatAgents logs={agentLogs} />
            </TabsContent>

            <TabsContent
                value="tools"
                className="mt-2 flex-1 overflow-auto pr-4"
            >
                <MemoizedChatTools logs={searchLogs} />
            </TabsContent>

            <TabsContent
                value="vectorStores"
                className="mt-2 flex-1 overflow-auto pr-4"
            >
                <MemoizedChatVectorStores logs={vectorStoreLogs} />
            </TabsContent>

            <TabsContent
                value="screenshots"
                className="mt-2 flex-1 overflow-auto pr-4"
            >
                <MemoizedChatScreenshots screenshots={screenshots} />
            </TabsContent>
        </Tabs>
    );
};

// Using React.memo to prevent unnecessary rerenders
export default memo(ChatTabs);
