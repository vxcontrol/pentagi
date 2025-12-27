import { memo, useEffect, useMemo, useState } from 'react';

import type { TaskFragmentFragment } from '@/graphql/types';

import Markdown from '@/components/shared/markdown';
import { Progress } from '@/components/ui/progress';
import { StatusType } from '@/graphql/types';

import FlowSubtask from './flow-subtask';
import FlowTaskStatusIcon from './flow-task-status-icon';

interface FlowTaskProps {
    searchValue?: string;
    task: TaskFragmentFragment;
}

// Helper function to check if text contains search value (case-insensitive)
const containsSearchValue = (text: null | string | undefined, searchValue: string): boolean => {
    if (!text || !searchValue.trim()) {
        return false;
    }

    return text.toLowerCase().includes(searchValue.toLowerCase().trim());
};

const FlowTask = ({ searchValue = '', task }: FlowTaskProps) => {
    const { id, result, status, subtasks, title } = task;
    const [isDetailsVisible, setIsDetailsVisible] = useState(false);

    // Memoize search checks to avoid recalculating on every render
    const searchChecks = useMemo(() => {
        const trimmedSearch = searchValue.trim();

        if (!trimmedSearch) {
            return { hasResultMatch: false };
        }

        return {
            hasResultMatch: containsSearchValue(result, trimmedSearch),
        };
    }, [searchValue, result]);

    // Auto-expand details if they contain search matches
    useEffect(() => {
        const trimmedSearch = searchValue.trim();

        if (trimmedSearch) {
            // Expand result block only if it contains the search term
            if (searchChecks.hasResultMatch) {
                setIsDetailsVisible(true);
            }
        } else {
            // Reset to default state when search is cleared
            setIsDetailsVisible(false);
        }
    }, [searchValue, searchChecks.hasResultMatch]);

    const sortedSubtasks = [...(subtasks || [])].sort((a, b) => +a.id - +b.id);
    const hasSubtasks = subtasks && subtasks.length > 0;

    // Calculate completed subtasks count
    const completedSubtasksCount = useMemo(() => {
        if (!subtasks?.length) {
            return 0;
        }

        return subtasks.filter((subtask) => subtask.status === StatusType.Finished).length;
    }, [subtasks]);

    // Calculate progress based on completed subtasks
    const progress = useMemo(() => {
        if (!subtasks?.length) {
            return 0;
        }

        return Math.round((completedSubtasksCount / subtasks.length) * 100);
    }, [subtasks, completedSubtasksCount]);

    return (
        <div className="relative">
            <div className="relative flex gap-2 pb-4">
                <FlowTaskStatusIcon
                    className="bg-background ring-border ring-background relative z-1 mt-px size-5 rounded-full ring-3"
                    status={status}
                    tooltip={`Task ID: ${id}`}
                />
                <div className="flex flex-1 flex-col gap-2">
                    <div className="font-semibold">
                        <Markdown
                            className="prose-fixed prose-sm wrap-break-word *:m-0 [&>p]:leading-tight"
                            searchValue={searchValue}
                        >
                            {title}
                        </Markdown>
                    </div>

                    {hasSubtasks && (
                        <div className="flex items-center gap-2">
                            <Progress
                                className="h-1.5 flex-1"
                                value={progress}
                            />
                            <div className="text-muted-foreground shrink-0 text-xs text-nowrap">
                                {progress}% completed ({completedSubtasksCount} of {subtasks?.length})
                            </div>
                        </div>
                    )}

                    {result && (
                        <div className="text-muted-foreground text-xs">
                            <div
                                className="cursor-pointer"
                                onClick={() => setIsDetailsVisible(!isDetailsVisible)}
                            >
                                {isDetailsVisible ? 'Hide details' : 'Show details'}
                            </div>
                            {isDetailsVisible && (
                                <>
                                    <div className="border-border my-3 border-t" />
                                    <Markdown
                                        className="prose-xs prose-fixed wrap-break-word"
                                        searchValue={searchValue}
                                    >
                                        {result}
                                    </Markdown>
                                </>
                            )}
                        </div>
                    )}
                </div>
                <div className="border-red absolute top-0 left-[calc((--spacing(2.5))-0.5px)] h-full border-l"></div>
            </div>

            {hasSubtasks ? (
                <div className="flex flex-col">
                    {sortedSubtasks.map((subtask) => (
                        <FlowSubtask
                            key={subtask.id}
                            searchValue={searchValue}
                            subtask={subtask}
                        />
                    ))}
                </div>
            ) : (
                <div className="text-muted-foreground mt-2 ml-6 text-xs">Waiting for subtasks to be created...</div>
            )}
        </div>
    );
};

export default memo(FlowTask);
