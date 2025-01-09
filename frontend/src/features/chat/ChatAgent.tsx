import { useState } from 'react';

import Markdown from '@/components/Markdown';
import type { AgentLogFragmentFragment } from '@/graphql/types';
import { formatDate } from '@/lib/utils/format';

import ChatAgentIcon from './ChatAgentIcon';

const taskPreviewLength = 500;

interface ChatAgentProps {
    log: AgentLogFragmentFragment;
}

const ChatAgent = ({ log }: ChatAgentProps) => {
    const { executor, initiator, task, result, taskId, subtaskId, createdAt } = log;
    const [isDetailsVisible, setIsDetailsVisible] = useState(false);

    const taskPreview = task.slice(0, taskPreviewLength) + (task.length > taskPreviewLength ? '...' : '');

    return (
        <div className="flex flex-col items-start">
            <div className="max-w-full rounded-lg bg-accent p-3 text-accent-foreground">
                <Markdown className="prose-xs prose-fixed break-words">
                    {isDetailsVisible ? task : taskPreview}
                </Markdown>
                {(result || task.length > taskPreviewLength) && (
                    <div className="text-xs text-muted-foreground">
                        <div
                            onClick={() => setIsDetailsVisible(!isDetailsVisible)}
                            className="cursor-pointer"
                        >
                            {isDetailsVisible ? 'Hide details' : 'Show details'}
                        </div>
                        {isDetailsVisible && result && (
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

export default ChatAgent;
