import type { TaskFragmentFragment } from '@/graphql/types';

import { ChatTask } from './ChatTask';

interface ChatTasksProps {
    tasks: TaskFragmentFragment[];
}

export const ChatTasks = ({ tasks }: ChatTasksProps) => {
    const sortedTasks = [...(tasks || [])].sort((a, b) => +a.id - +b.id);

    return (
        <div className="flex h-dvh">
            {tasks?.length ? (
                <div className="w-full space-y-4">
                    {sortedTasks.map((task) => (
                        <ChatTask
                            key={task.id}
                            task={task}
                        />
                    ))}
                </div>
            ) : null}
        </div>
    );
};

export default ChatTasks;
