import type { Dispatch, SetStateAction } from 'react';
import { memo, useEffect, useMemo, useRef } from 'react';

import { ScrollArea, ScrollBar } from '@/components/ui/scroll-area';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import type { AssistantFragmentFragment, AssistantLogFragmentFragment, FlowQuery } from '@/graphql/types';
import { useBreakpoint } from '@/hooks/use-breakpoint';

import ChatAgents from './ChatAgents';
import ChatAssistantMessages from './ChatAssistantMessages';
import ChatAutomationMessages from './ChatAutomationMessages';
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
const MemoizedChatAutomationMessages = memo(ChatAutomationMessages);
const MemoizedChatAssistantMessages = memo(ChatAssistantMessages);

interface ChatTabsProps {
    flowData: FlowQuery | undefined;
    selectedFlowId: string | null;
    onSubmitAutomationMessage: (message: string) => Promise<void>;
    onStopAutomationFlow?: (flowId: string) => Promise<void>;
    assistants: AssistantFragmentFragment[];
    assistantLogs?: AssistantLogFragmentFragment[];
    selectedAssistantId?: string | null;
    selectedProvider: string;
    providers: string[];
    onSelectAssistant?: (assistantId: string | null) => void;
    onCreateAssistant?: () => void;
    onDeleteAssistant?: (assistantId: string) => void;
    onSubmitAssistantMessage?: (assistantId: string, message: string, useAgents: boolean) => Promise<void>;
    onCreateNewAssistant?: (message: string, useAgents: boolean) => Promise<void>;
    onStopAssistant?: (assistantId: string) => Promise<void>;
    activeTab: string;
    onTabChange: Dispatch<SetStateAction<string>>;
}

const ChatTabs = ({
    flowData,
    selectedFlowId,
    onSubmitAutomationMessage,
    onStopAutomationFlow,
    assistants,
    assistantLogs,
    selectedAssistantId,
    selectedProvider,
    providers,
    onSelectAssistant,
    onCreateAssistant,
    onDeleteAssistant,
    onSubmitAssistantMessage,
    onCreateNewAssistant,
    onStopAssistant,
    activeTab,
    onTabChange,
}: ChatTabsProps) => {
    const { isDesktop } = useBreakpoint();
    const previousActiveTabRef = useRef<string>(activeTab);

    useEffect(() => {
        // Only handle actual tab changes
        if (activeTab === previousActiveTabRef.current) {
            return;
        }

        previousActiveTabRef.current = activeTab;
    }, [activeTab]);

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
            onValueChange={onTabChange}
            className="flex size-full flex-col"
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
                    value="automation"
                    className="mt-2 flex-1 overflow-auto"
                >
                    <MemoizedChatAutomationMessages
                        logs={messageLogs}
                        className="pr-4"
                        flowData={flowData}
                        selectedFlowId={selectedFlowId}
                        onSubmitMessage={onSubmitAutomationMessage}
                        onStopFlow={onStopAutomationFlow}
                    />
                </TabsContent>
            )}
            {!isDesktop && (
                <TabsContent
                    value="assistant"
                    className="mt-2 flex-1 overflow-auto"
                >
                    <MemoizedChatAssistantMessages
                        logs={assistantLogs}
                        className="pr-4"
                        selectedFlowId={selectedFlowId}
                        assistants={assistants}
                        selectedAssistantId={selectedAssistantId}
                        selectedProvider={selectedProvider}
                        providers={providers}
                        onSelectAssistant={onSelectAssistant}
                        onCreateAssistant={onCreateAssistant}
                        onDeleteAssistant={onDeleteAssistant}
                        onSubmitMessage={onSubmitAssistantMessage}
                        onCreateNewAssistant={onCreateNewAssistant}
                        onStopAssistant={onStopAssistant}
                    />
                </TabsContent>
            )}

            {/* Desktop and Mobile Tabs */}
            <TabsContent
                value="terminal"
                className="mt-2 flex-1 overflow-auto"
            >
                <MemoizedChatTerminal logs={terminalLogs} selectedFlowId={selectedFlowId} />
            </TabsContent>

            <TabsContent
                value="tasks"
                className="mt-2 flex-1 overflow-auto pr-4"
            >
                <MemoizedChatTasks tasks={tasks} selectedFlowId={selectedFlowId} flow={flowData?.flow} />
            </TabsContent>

            <TabsContent
                value="agents"
                className="mt-2 flex-1 overflow-auto pr-4"
            >
                <MemoizedChatAgents logs={agentLogs} selectedFlowId={selectedFlowId} />
            </TabsContent>

            <TabsContent
                value="tools"
                className="mt-2 flex-1 overflow-auto pr-4"
            >
                <MemoizedChatTools logs={searchLogs} selectedFlowId={selectedFlowId} />
            </TabsContent>

            <TabsContent
                value="vectorStores"
                className="mt-2 flex-1 overflow-auto pr-4"
            >
                <MemoizedChatVectorStores logs={vectorStoreLogs} selectedFlowId={selectedFlowId} />
            </TabsContent>

            <TabsContent
                value="screenshots"
                className="mt-2 flex-1 overflow-auto pr-4"
            >
                <MemoizedChatScreenshots screenshots={screenshots} selectedFlowId={selectedFlowId} />
            </TabsContent>
        </Tabs>
    );
};

export default memo(ChatTabs);
