import { ListCheck, ListTodo } from 'lucide-react';
import { useState } from 'react';

import Markdown from '@/components/Markdown';
import type { SubtaskFragmentFragment } from '@/graphql/types';

import ChatTaskStatusIcon from './ChatTaskStatusIcon';

interface ChatSubtaskProps {
    subtask: SubtaskFragmentFragment;
}

const ChatSubtask = ({ subtask }: ChatSubtaskProps) => {
    const { id, status, title, description, result } = subtask;
    const [isDetailsVisible, setIsDetailsVisible] = useState(false);

    return (
        <div className="border-l pl-4">
            <div className="flex gap-2">
                <ChatTaskStatusIcon
                    status={status}
                    tooltip={`Subtask ID: ${id}`}
                    className="mt-0.5"
                />
                <p className="text-sm">{title}</p>
            </div>
            {(description || result) && (
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
                            {description && (
                                <>
                                    <div className="flex items-center gap-2 text-muted-foreground">
                                        <ListTodo className="size-4" />
                                        <span className="text-xs">Description</span>
                                    </div>
                                    <Markdown className="prose-xs prose-fixed ml-6 break-words">{description}</Markdown>
                                    {result && <div className="my-2 border-t border-border" />}
                                </>
                            )}
                            {result && (
                                <>
                                    <div className="flex items-center gap-2 text-muted-foreground">
                                        <ListCheck className="size-4" />
                                        <span className="text-xs">Result</span>
                                    </div>
                                    <Markdown className="prose-xs prose-fixed ml-6 break-words">{result}</Markdown>
                                </>
                            )}
                        </>
                    )}
                </div>
            )}
        </div>
    );
};

export default ChatSubtask;
