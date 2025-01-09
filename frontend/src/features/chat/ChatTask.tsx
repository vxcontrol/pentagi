import { useState } from 'react';

import Markdown from '@/components/Markdown';
import type { TaskFragmentFragment } from '@/graphql/types';

import ChatSubtask from './ChatSubtask';
import ChatTaskStatusIcon from './ChatTaskStatusIcon';

interface ChatTaskProps {
    task: TaskFragmentFragment;
}

export const ChatTask = ({ task }: ChatTaskProps) => {
    const { id, status, title, result, subtasks } = task;
    const [isDetailsVisible, setIsDetailsVisible] = useState(false);

    const sortedSubtasks = [...(subtasks || [])].sort((a, b) => +a.id - +b.id);

    return (
        <div>
            <div className="flex gap-2">
                <ChatTaskStatusIcon
                    status={status}
                    tooltip={`Task ID: ${id}`}
                    className="mt-1"
                />
                <h3 className="font-semibold">{title}</h3>
            </div>
            {result && (
                <div className="ml-6 text-xs text-muted-foreground">
                    <div
                        onClick={() => setIsDetailsVisible(!isDetailsVisible)}
                        className="cursor-pointer"
                    >
                        {isDetailsVisible ? 'Hide details' : 'Show details'}
                    </div>
                    {isDetailsVisible && (
                        <>
                            <div className="my-2 border-t border-border" />
                            <Markdown className="prose-xs prose-fixed break-words">{result}</Markdown>
                        </>
                    )}
                </div>
            )}
            <div className="mt-2 space-y-2">
                {sortedSubtasks.map((subtask) => (
                    <ChatSubtask
                        key={subtask.id}
                        subtask={subtask}
                    />
                ))}
            </div>
        </div>
    );
};
