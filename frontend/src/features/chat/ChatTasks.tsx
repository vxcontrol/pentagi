import { memo } from 'react';

import type { TaskFragmentFragment } from '@/graphql/types';

import ChatTask from './ChatTask';

interface ChatTasksProps {
    tasks: TaskFragmentFragment[];
}

const ChatTasks = ({ tasks }: ChatTasksProps) => {
    const sortedTasks = [...(tasks || [])].sort((a, b) => +a.id - +b.id);
    const hasTasks = tasks && tasks.length > 0;

    return (
        <div className="flex h-full flex-col">
            {hasTasks ? (
                <div className="flex-1 space-y-4 overflow-auto pb-4">
                    {sortedTasks.map((task) => (
                        <ChatTask
                            key={task.id}
                            task={task}
                        />
                    ))}
                </div>
            ) : (
                <div className="flex flex-1 items-center justify-center">
                    <div className="flex flex-col items-center gap-2 text-muted-foreground">
                        <p>No tasks found for this flow</p>
                        <p className="text-xs">Tasks will appear here once the agent starts working</p>
                    </div>
                </div>
            )}
        </div>
    );
};

export default memo(ChatTasks);
