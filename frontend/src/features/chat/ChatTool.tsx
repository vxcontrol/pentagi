import { Hammer } from 'lucide-react';
import { useState } from 'react';

import Markdown from '@/components/Markdown';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import type { SearchLogFragmentFragment } from '@/graphql/types';
import { formatDate, formatName } from '@/lib/utils/format';

import ChatAgentIcon from './ChatAgentIcon';

interface ChatToolProps {
    log: SearchLogFragmentFragment;
}

const ChatTool = ({ log }: ChatToolProps) => {
    const { executor, initiator, query, result, engine, taskId, subtaskId, createdAt } = log;
    const [isDetailsVisible, setIsDetailsVisible] = useState(false);

    return (
        <div className="flex flex-col items-start">
            <div className="max-w-full rounded-lg bg-accent p-3 text-accent-foreground">
                <div className="flex flex-col">
                    <div className="cursor-pointer text-sm font-semibold">
                        <Tooltip>
                            <TooltipTrigger asChild>
                                <span className="inline-flex items-center gap-1">
                                    <Hammer className="size-4 text-muted-foreground" />
                                    <span>{formatName(engine)}</span>
                                </span>
                            </TooltipTrigger>
                            <TooltipContent>Tool name</TooltipContent>
                        </Tooltip>
                    </div>

                    <Markdown className="prose-xs prose-fixed break-words">{query}</Markdown>
                </div>
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
                                <Markdown className="prose-xs prose-fixed break-words">{result}</Markdown>
                            </>
                        )}
                    </div>
                )}
            </div>
            <div className="mt-1 flex items-center gap-1 px-1 text-xs text-muted-foreground">
                <span className="flex items-center gap-0.5">
                    <ChatAgentIcon
                        type={executor}
                        className="text-muted-foreground"
                    />
                    <span className="text-muted-foreground/50">→</span>
                    <ChatAgentIcon
                        type={initiator}
                        className="text-muted-foreground"
                    />
                </span>
                <span className="text-muted-foreground/50">{formatDate(new Date(createdAt))}</span>
                {taskId && (
                    <>
                        <span className="text-muted-foreground/50">|</span>
                        <span className="text-muted-foreground/50">Task ID: {taskId}</span>
                    </>
                )}
                {subtaskId && (
                    <>
                        <span className="text-muted-foreground/50">|</span>
                        <span className="text-muted-foreground/50">Subtask ID: {subtaskId}</span>
                    </>
                )}
            </div>
        </div>
    );
};

export default ChatTool;
