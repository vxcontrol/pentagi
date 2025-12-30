import { ListCheck, ListTodo } from 'lucide-react';
import { memo, useEffect, useMemo, useState } from 'react';

import type { SubtaskFragmentFragment } from '@/graphql/types';

import Markdown from '@/components/shared/markdown';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

import FlowTaskStatusIcon from './flow-task-status-icon';

interface FlowSubtaskProps {
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

const FlowSubtask = ({ searchValue = '', subtask }: FlowSubtaskProps) => {
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
        <div className="group relative flex gap-2.5 pb-4 pl-0.5">
            <FlowTaskStatusIcon
                className="bg-background ring-border ring-background relative z-1 mt-px rounded-full ring-3"
                status={status}
                tooltip={`Subtask ID: ${id}`}
            />
            <div className="flex flex-1 flex-col gap-2">
                <div className="text-sm">
                    <Markdown
                        className="prose-fixed prose-sm wrap-break-word *:m-0 [&>p]:leading-tight"
                        searchValue={searchValue}
                    >
                        {title}
                    </Markdown>
                </div>

                {hasDetails && (
                    <div className="text-muted-foreground text-xs">
                        <div
                            className="cursor-pointer hover:underline"
                            onClick={() => setIsDetailsVisible(!isDetailsVisible)}
                        >
                            {isDetailsVisible ? 'Hide details' : 'Show details'}
                        </div>
                        {isDetailsVisible && (
                            <div className="mt-4 flex flex-col gap-4">
                                {description && (
                                    <Card>
                                        <CardHeader className="p-3">
                                            <CardTitle className="flex items-center gap-2">
                                                <ListTodo className="size-4 shrink-0" /> Description
                                            </CardTitle>
                                        </CardHeader>
                                        <CardContent className="p-3 pt-0">
                                            <hr className="mt-0 mb-3" />
                                            <Markdown
                                                className="prose-xs prose-fixed wrap-break-word"
                                                searchValue={searchValue}
                                            >
                                                {description}
                                            </Markdown>
                                        </CardContent>
                                    </Card>
                                )}
                                {result && (
                                    <Card>
                                        <CardHeader className="p-3">
                                            <CardTitle className="flex items-center gap-2">
                                                <ListCheck className="size-4 shrink-0" /> Result
                                            </CardTitle>
                                        </CardHeader>
                                        <CardContent className="p-3 pt-0">
                                            <hr className="mt-0 mb-3" />
                                            <Markdown
                                                className="prose-xs prose-fixed wrap-break-word"
                                                searchValue={searchValue}
                                            >
                                                {result}
                                            </Markdown>
                                        </CardContent>
                                    </Card>
                                )}
                            </div>
                        )}
                    </div>
                )}
            </div>
            <div className="absolute top-0 left-[calc((--spacing(2.5))-0.5px)] h-full border-l group-last:hidden"></div>
        </div>
    );
};

export default memo(FlowSubtask);
