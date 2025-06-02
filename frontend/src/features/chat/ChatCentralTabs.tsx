import type { Dispatch, SetStateAction } from 'react';
import { memo, useMemo } from 'react';

import { ScrollArea, ScrollBar } from '@/components/ui/scroll-area';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import type { AssistantFragmentFragment, AssistantLogFragmentFragment, FlowQuery, MessageLogFragmentFragment } from '@/graphql/types';

import ChatAssistantMessages from './ChatAssistantMessages';
import ChatAutomationMessages from './ChatAutomationMessages';

const MemoizedChatAutomationMessages = memo(ChatAutomationMessages);
const MemoizedChatAssistantMessages = memo(ChatAssistantMessages);

interface ChatCentralTabsProps {
    selectedFlowId: string | null;
    flowData?: FlowQuery;
    assistants: AssistantFragmentFragment[];
    assistantLogs?: AssistantLogFragmentFragment[];
    selectedAssistantId?: string | null;
    selectedProvider: string;
    providers: string[];
    onSelectAssistant?: (assistantId: string | null) => void;
    onCreateAssistant?: () => void;
    onDeleteAssistant?: (assistantId: string) => void;
    onSubmitAutomationMessage: (message: string) => Promise<void>;
    onSubmitAssistantMessage?: (assistantId: string, message: string, useAgents: boolean) => Promise<void>;
    onCreateNewAssistant?: (message: string, useAgents: boolean) => Promise<void>;
    onStopAutomationFlow?: (flowId: string) => Promise<void>;
    onStopAssistant?: (assistantId: string) => Promise<void>;
    activeTab: string;
    onTabChange: Dispatch<SetStateAction<string>>;
}

const ChatCentralTabs = ({
    selectedFlowId,
    flowData,
    assistants,
    assistantLogs,
    selectedAssistantId,
    selectedProvider,
    providers,
    onSelectAssistant,
    onCreateAssistant,
    onDeleteAssistant,
    onSubmitAutomationMessage,
    onSubmitAssistantMessage,
    onCreateNewAssistant,
    onStopAutomationFlow,
    onStopAssistant,
    activeTab,
    onTabChange,
}: ChatCentralTabsProps) => {
    const messageLogs = useMemo<MessageLogFragmentFragment[]>(
        () => flowData?.messageLogs ?? [],
        [flowData?.messageLogs],
    );

    return (
        <Tabs
            value={activeTab}
            onValueChange={onTabChange}
            className="flex size-full flex-col"
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
                value="automation"
                className="mt-2 flex-1 overflow-auto pr-4"
            >
                <MemoizedChatAutomationMessages
                    logs={messageLogs}
                    selectedFlowId={selectedFlowId}
                    flowData={flowData}
                    onSubmitMessage={onSubmitAutomationMessage}
                    onStopFlow={onStopAutomationFlow}
                />
            </TabsContent>
            <TabsContent
                value="assistant"
                className="mt-2 flex-1 overflow-auto pr-4"
            >
                <MemoizedChatAssistantMessages
                    logs={assistantLogs}
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
        </Tabs>
    );
};

export default memo(ChatCentralTabs);
