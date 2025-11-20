import type { Dispatch, SetStateAction } from 'react';

import { memo, useMemo } from 'react';

import type {
    AssistantFragmentFragment,
    AssistantLogFragmentFragment,
    FlowQuery,
    MessageLogFragmentFragment,
} from '@/graphql/types';
import type { Provider } from '@/models/Provider';

import { ScrollArea, ScrollBar } from '@/components/ui/scroll-area';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';

import type { FlowFormValues } from '../flows/FlowForm';

import ChatAssistantMessages from './ChatAssistantMessages';
import ChatAutomationMessages from './ChatAutomationMessages';

const MemoizedChatAutomationMessages = memo(ChatAutomationMessages);
const MemoizedChatAssistantMessages = memo(ChatAssistantMessages);

interface ChatCentralTabsProps {
    activeTab: string;
    assistantLogs?: AssistantLogFragmentFragment[];
    assistants: AssistantFragmentFragment[];
    flowData?: FlowQuery;
    onCreateAssistant?: () => void;
    onCreateNewAssistant?: (message: string, useAgents: boolean) => Promise<void>;
    onDeleteAssistant?: (assistantId: string) => void;
    onSelectAssistant?: (assistantId: null | string) => void;
    onStopAssistant?: (assistantId: string) => Promise<void>;
    onStopAutomationFlow?: (flowId: string) => Promise<void>;
    onSubmitAssistantMessage?: (assistantId: string, message: string, useAgents: boolean) => Promise<void>;
    onSubmitAutomationMessage: (values: FlowFormValues) => Promise<void>;
    onTabChange: Dispatch<SetStateAction<string>>;
    providers: Provider[];
    selectedAssistantId?: null | string;
    selectedFlowId: null | string;
    selectedProvider: null | Provider;
}

const ChatCentralTabs = ({
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
}: ChatCentralTabsProps) => {
    const messageLogs = useMemo<MessageLogFragmentFragment[]>(
        () => flowData?.messageLogs ?? [],
        [flowData?.messageLogs],
    );

    return (
        <Tabs
            className="flex size-full flex-col"
            onValueChange={onTabChange}
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
                <MemoizedChatAutomationMessages
                    flowData={flowData}
                    logs={messageLogs}
                    onStopFlow={onStopAutomationFlow}
                    onSubmitMessage={onSubmitAutomationMessage}
                    selectedFlowId={selectedFlowId}
                />
            </TabsContent>
            <TabsContent
                className="mt-2 flex-1 overflow-auto pr-4"
                value="assistant"
            >
                <MemoizedChatAssistantMessages
                    assistants={assistants}
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
        </Tabs>
    );
};

export default memo(ChatCentralTabs);
