import { memo, useCallback, useEffect, useMemo, useState } from 'react';

import Markdown from '@/components/shared/Markdown';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import type { AgentLogFragmentFragment } from '@/graphql/types';
import { formatDate } from '@/lib/utils/format';
import { copyMessageToClipboard } from '@/lib/сlipboard';
import { Copy } from 'lucide-react';

import ChatAgentIcon from './ChatAgentIcon';

const taskPreviewLength = 500;

interface ChatAgentProps {
    log: AgentLogFragmentFragment;
    searchValue?: string;
}

// Helper function to check if text contains search value (case-insensitive)
const containsSearchValue = (text: string | null | undefined, searchValue: string): boolean => {
    if (!text || !searchValue.trim()) {
        return false;
    }
    return text.toLowerCase().includes(searchValue.toLowerCase().trim());
};

const ChatAgent = ({ log, searchValue = '' }: ChatAgentProps) => {
    const { executor, initiator, task, result, taskId, subtaskId, createdAt } = log;

    // Memoize search checks to avoid recalculating on every render
    const searchChecks = useMemo(() => {
        const trimmedSearch = searchValue.trim();
        if (!trimmedSearch) {
            return { hasTaskMatch: false, hasResultMatch: false };
        }

        return {
            hasTaskMatch: containsSearchValue(task, trimmedSearch),
            hasResultMatch: containsSearchValue(result, trimmedSearch),
        };
    }, [searchValue, task, result]);

    const [isDetailsVisible, setIsDetailsVisible] = useState(false);

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

    // Determine if we should show full task or preview
    // Show full task if: search found in task OR details are manually visible OR task is short
    const shouldShowFullTask = searchChecks.hasTaskMatch || isDetailsVisible || task.length <= taskPreviewLength;
    const taskToShow = shouldShowFullTask ? task : task.slice(0, taskPreviewLength) + '...';

    // Determine if we should show details toggle
    // Show toggle if: result exists OR task is longer than preview length
    const shouldShowDetailsToggle = result || task.length > taskPreviewLength;

    const handleCopy = useCallback(async () => {
        await copyMessageToClipboard({
            message: task,
            result: result || undefined,
        });
    }, [task, result]);

    return (
        <div className="flex flex-col items-start">
            <div className="max-w-full rounded-lg bg-accent p-3 text-accent-foreground">
                <Markdown
                    className="prose-xs prose-fixed break-words"
                    searchValue={searchValue}
                >
                    {taskToShow}
                </Markdown>
                {shouldShowDetailsToggle && (
                    <div className="mt-2 text-xs text-muted-foreground">
                        <div
                            onClick={() => setIsDetailsVisible(!isDetailsVisible)}
                            className="cursor-pointer"
                        >
                            {isDetailsVisible ? 'Hide details' : 'Show details'}
                        </div>
                        {isDetailsVisible && result && (
                            <>
                                <div className="my-2 border-t dark:border-gray-700" />
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
            </div>
            <div className="mt-1 flex items-center gap-1 px-1 text-xs text-muted-foreground">
                <span className="flex items-center gap-0.5">
                    <ChatAgentIcon
                        type={initiator}
                        className="text-muted-foreground"
                    />
                    <span className="text-muted-foreground/50">→</span>
                    <ChatAgentIcon
                        type={executor}
                        className="text-muted-foreground"
                    />
                </span>
                <Tooltip>
                    <TooltipTrigger asChild>
                        <Copy
                            className="size-3 shrink-0 cursor-pointer hover:text-foreground ml-1 mr-1 transition-colors"
                            onClick={handleCopy}
                        />
                    </TooltipTrigger>
                    <TooltipContent>Copy</TooltipContent>
                </Tooltip>
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

export default memo(ChatAgent);
