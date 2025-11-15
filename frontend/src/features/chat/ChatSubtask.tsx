import { ListCheck, ListTodo } from 'lucide-react';
import { memo, useEffect, useMemo, useState } from 'react';

import type { SubtaskFragmentFragment } from '@/graphql/types';

import Markdown from '@/components/shared/Markdown';

import ChatTaskStatusIcon from './ChatTaskStatusIcon';

interface ChatSubtaskProps {
    searchValue?: string;
    subtask: SubtaskFragmentFragment;
}

// Helper function to check if text contains search value (case-insensitive)
const containsSearchValue = (text: null | string | undefined, searchValue: string): boolean => {
    if (!text || !searchValue.trim()) {
        return false;
    }

    return text.toLowerCase().includes(searchValue.toLowerCase().trim());
};

const ChatSubtask = ({ searchValue = '', subtask }: ChatSubtaskProps) => {
    const { description, id, result, status, title } = subtask;
    const [isDetailsVisible, setIsDetailsVisible] = useState(false);
    const hasDetails = description || result;

    // Memoize search checks to avoid recalculating on every render
    const searchChecks = useMemo(() => {
        const trimmedSearch = searchValue.trim();

        if (!trimmedSearch) {
            return { hasDescriptionMatch: false, hasResultMatch: false };
        }

        return {
            hasDescriptionMatch: containsSearchValue(description, trimmedSearch),
            hasResultMatch: containsSearchValue(result, trimmedSearch),
        };
    }, [searchValue, description, result]);

    // Auto-expand details if they contain search matches
    useEffect(() => {
        const trimmedSearch = searchValue.trim();

        if (trimmedSearch) {
            // Expand details if description or result contains the search term
            if (searchChecks.hasDescriptionMatch || searchChecks.hasResultMatch) {
                setIsDetailsVisible(true);
            }
        } else {
            // Reset to default state when search is cleared
            setIsDetailsVisible(false);
        }
    }, [searchValue, searchChecks.hasDescriptionMatch, searchChecks.hasResultMatch]);

    return (
        <div className="border-l pl-4">
            <div className="flex gap-2">
                <ChatTaskStatusIcon
                    className="mt-px"
                    status={status}
                    tooltip={`Subtask ID: ${id}`}
                />
                <div className="text-sm">
                    <Markdown
                        className="prose-fixed prose-sm break-words [&>*]:m-0 [&>p]:leading-tight"
                        searchValue={searchValue}
                    >
                        {title}
                    </Markdown>
                </div>
            </div>
            {hasDetails && (
                <div className="ml-6 text-xs text-muted-foreground">
                    <div
                        className="cursor-pointer hover:underline"
                        onClick={() => setIsDetailsVisible(!isDetailsVisible)}
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
                                    <Markdown
                                        className="prose-xs prose-fixed ml-6 break-words"
                                        searchValue={searchValue}
                                    >
                                        {description}
                                    </Markdown>
                                    {result && <div className="my-2 border-t border-border" />}
                                </>
                            )}
                            {result && (
                                <>
                                    <div className="flex items-center gap-2 text-muted-foreground">
                                        <ListCheck className="size-4" />
                                        <span className="text-xs">Result</span>
                                    </div>
                                    <Markdown
                                        className="prose-xs prose-fixed ml-6 break-words"
                                        searchValue={searchValue}
                                    >
                                        {result}
                                    </Markdown>
                                </>
                            )}
                        </>
                    )}
                </div>
            )}
        </div>
    );
};

export default memo(ChatSubtask);
