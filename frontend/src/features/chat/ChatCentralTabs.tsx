import type { Dispatch, SetStateAction } from 'react';

import { memo, useMemo } from 'react';

import type { MessageLogFragmentFragment } from '@/graphql/types';

import { ScrollArea, ScrollBar } from '@/components/ui/scroll-area';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { useFlow } from '@/providers/FlowProvider';

import ChatAssistantMessages from './ChatAssistantMessages';
import ChatAutomationMessages from './ChatAutomationMessages';

const MemoizedChatAutomationMessages = memo(ChatAutomationMessages);
const MemoizedChatAssistantMessages = memo(ChatAssistantMessages);

interface ChatCentralTabsProps {
    activeTab: string;
    onTabChange: Dispatch<SetStateAction<string>>;
}

const ChatCentralTabs = ({ activeTab, onTabChange }: ChatCentralTabsProps) => {
    const { flowData } = useFlow();

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
                <MemoizedChatAutomationMessages logs={messageLogs} />
            </TabsContent>
            <TabsContent
                className="mt-2 flex-1 overflow-auto pr-4"
                value="assistant"
            >
                <MemoizedChatAssistantMessages />
            </TabsContent>
        </Tabs>
    );
};

export default memo(ChatCentralTabs);
