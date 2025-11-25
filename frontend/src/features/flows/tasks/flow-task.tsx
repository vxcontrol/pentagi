import { memo, useEffect, useMemo, useState } from 'react';

import type { TaskFragmentFragment } from '@/graphql/types';

import Markdown from '@/components/shared/markdown';

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

    return (
        <div className="rounded-lg border p-4 shadow-sm">
            <div className="flex gap-2">
                <FlowTaskStatusIcon
                    className="mt-px"
                    status={status}
                    tooltip={`Task ID: ${id}`}
                />
                <div className="font-semibold">
                    <Markdown
                        className="prose-fixed prose-sm break-words [&>*]:m-0 [&>p]:leading-tight"
                        searchValue={searchValue}
                    >
                        {title}
                    </Markdown>
                </div>
            </div>
            {result && (
                <div className="ml-6 text-xs text-muted-foreground">
                    <div
                        className="cursor-pointer"
                        onClick={() => setIsDetailsVisible(!isDetailsVisible)}
                    >
                        {isDetailsVisible ? 'Hide details' : 'Show details'}
                    </div>
                    {isDetailsVisible && (
                        <>
                            <div className="my-2 border-t border-border" />
                            <Markdown
                                className="prose-xs prose-fixed break-words"
                                searchValue={searchValue}
                            >
                                {result}
                            </Markdown>
                        </>
                    )}
                </div>
            )}
            {hasSubtasks ? (
                <div className="mt-2 space-y-2">
                    {sortedSubtasks.map((subtask) => (
                        <FlowSubtask
                            key={subtask.id}
                            searchValue={searchValue}
                            subtask={subtask}
                        />
                    ))}
                </div>
            ) : (
                <div className="ml-6 mt-2 text-xs text-muted-foreground">Waiting for subtasks to be created...</div>
            )}
        </div>
    );
};

export default memo(FlowTask);

