import type { Dispatch, SetStateAction } from 'react';

import { memo, useEffect, useMemo, useRef } from 'react';

import type { AssistantFragmentFragment, AssistantLogFragmentFragment, FlowQuery } from '@/graphql/types';
import type { Provider } from '@/models/Provider';

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
    activeTab: string;
    assistantLogs?: AssistantLogFragmentFragment[];
    assistants: AssistantFragmentFragment[];
    flowData: FlowQuery | undefined;
    onCreateAssistant?: () => void;
    onCreateNewAssistant?: (message: string, useAgents: boolean) => Promise<void>;
    onDeleteAssistant?: (assistantId: string) => void;
    onSelectAssistant?: (assistantId: null | string) => void;
    onStopAssistant?: (assistantId: string) => Promise<void>;
    onStopAutomationFlow?: (flowId: string) => Promise<void>;
    onSubmitAssistantMessage?: (assistantId: string, message: string, useAgents: boolean) => Promise<void>;
    onSubmitAutomationMessage: (message: string) => Promise<void>;
    onTabChange: Dispatch<SetStateAction<string>>;
    providers: Provider[];
    selectedAssistantId?: null | string;
    selectedFlowId: null | string;
    selectedProvider: null | Provider;
}

const ChatTabs = ({
    activeTab,
    assistantLogs,
    assistants,
    flowData,
    onCreateAssistant,
    onCreateNewAssistant,
    onDeleteAssistant,
    onSelectAssistant,
    onStopAssistant,
    onStopAutomationFlow,
    onSubmitAssistantMessage,
    onSubmitAutomationMessage,
    onTabChange,
    providers,
    selectedAssistantId,
    selectedFlowId,
    selectedProvider,
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
                    <MemoizedChatAutomationMessages
                        className="pr-4"
                        flowData={flowData}
                        logs={messageLogs}
                        onStopFlow={onStopAutomationFlow}
                        onSubmitMessage={onSubmitAutomationMessage}
                        selectedFlowId={selectedFlowId}
                    />
                </TabsContent>
            )}
            {!isDesktop && (
                <TabsContent
                    className="mt-2 flex-1 overflow-auto"
                    value="assistant"
                >
                    <MemoizedChatAssistantMessages
                        assistants={assistants}
                        className="pr-4"
                        logs={assistantLogs}
                        onCreateAssistant={onCreateAssistant}
                        onCreateNewAssistant={onCreateNewAssistant}
                        onDeleteAssistant={onDeleteAssistant}
                        onSelectAssistant={onSelectAssistant}
                        onStopAssistant={onStopAssistant}
                        onSubmitMessage={onSubmitAssistantMessage}
                        providers={providers}
                        selectedAssistantId={selectedAssistantId}
                        selectedFlowId={selectedFlowId}
                        selectedProvider={selectedProvider}
                    />
                </TabsContent>
            )}

            {/* Desktop and Mobile Tabs */}
            <TabsContent
                className="mt-2 flex-1 overflow-auto"
                value="terminal"
            >
                <MemoizedChatTerminal
                    logs={terminalLogs}
                    selectedFlowId={selectedFlowId}
                />
            </TabsContent>

            <TabsContent
                className="mt-2 flex-1 overflow-auto pr-4"
                value="tasks"
            >
                <MemoizedChatTasks
                    flow={flowData?.flow}
                    selectedFlowId={selectedFlowId}
                    tasks={tasks}
                />
            </TabsContent>

            <TabsContent
                className="mt-2 flex-1 overflow-auto pr-4"
                value="agents"
            >
                <MemoizedChatAgents
                    logs={agentLogs}
                    selectedFlowId={selectedFlowId}
                />
            </TabsContent>

            <TabsContent
                className="mt-2 flex-1 overflow-auto pr-4"
                value="tools"
            >
                <MemoizedChatTools
                    logs={searchLogs}
                    selectedFlowId={selectedFlowId}
                />
            </TabsContent>

            <TabsContent
                className="mt-2 flex-1 overflow-auto pr-4"
                value="vectorStores"
            >
                <MemoizedChatVectorStores
                    logs={vectorStoreLogs}
                    selectedFlowId={selectedFlowId}
                />
            </TabsContent>

            <TabsContent
                className="mt-2 flex-1 overflow-auto pr-4"
                value="screenshots"
            >
                <MemoizedChatScreenshots
                    screenshots={screenshots}
                    selectedFlowId={selectedFlowId}
                />
            </TabsContent>
        </Tabs>
    );
};

export default memo(ChatTabs);
