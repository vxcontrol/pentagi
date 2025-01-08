import { useState } from 'react';

import Markdown from '@/components/Markdown';
import Terminal from '@/components/Terminal';
import type { MessageLogFragmentFragment } from '@/graphql/types';
import { MessageLogType, ResultFormat } from '@/graphql/types';
import { cn } from '@/lib/utils';
import { formatDate } from '@/lib/utils/format';

import ChatMessageTypeIcon from './ChatMessageTypeIcon';

interface ChatMessageProps {
    log: MessageLogFragmentFragment;
}

const ChatMessage = ({ log }: ChatMessageProps) => {
    const { type, createdAt, message, result, resultFormat = ResultFormat.Plain } = log;
    const [isDetailsVisible, setIsDetailsVisible] = useState(false);

    return (
        <div className={`flex flex-col ${type === MessageLogType.Input ? 'items-end' : 'items-start'}`}>
            <div
                className={cn(
                    'max-w-[90%] rounded-lg bg-accent p-3 text-accent-foreground',
                    resultFormat === ResultFormat.Terminal && isDetailsVisible ? 'w-full' : '',
                )}
            >
                <Markdown className="prose-xs prose-fixed break-words">{message}</Markdown>
                {result && (
                    <div className="text-xs text-muted-foreground">
                        <div
                            onClick={() => setIsDetailsVisible(!isDetailsVisible)}
                            className="cursor-pointer"
                        >
                            {isDetailsVisible ? 'Hide details' : 'Show details'}
                        </div>
                        {isDetailsVisible && (
                            <>
                                <div className="my-2 border-t dark:border-gray-700" />
                                {resultFormat === ResultFormat.Plain && (
                                    <div className="text-sm text-accent-foreground">{result}</div>
                                )}
                                {resultFormat === ResultFormat.Markdown && (
                                    <Markdown className="prose-xs prose-fixed break-words">{result}</Markdown>
                                )}
                                {resultFormat === ResultFormat.Terminal && (
                                    <Terminal
                                        logs={[result]}
                                        className="h-[240px] w-full bg-card py-1 pl-1"
                                    />
                                )}
                            </>
                        )}
                    </div>
                )}
            </div>
            <div
                className={`mt-1 flex items-center gap-1 px-1 text-xs text-muted-foreground ${
                    type === MessageLogType.Input ? 'flex-row-reverse' : 'flex-row'
                }`}
            >
                <ChatMessageTypeIcon type={type} />
                <span className="text-muted-foreground/50">{formatDate(new Date(createdAt))}</span>
            </div>
        </div>
    );
};

export default ChatMessage;
