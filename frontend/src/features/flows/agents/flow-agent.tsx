import { Copy } from 'lucide-react';
import { memo, useCallback, useMemo, useState } from 'react';

import type { AgentLogFragmentFragment } from '@/graphql/types';

import Markdown from '@/components/shared/markdown';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { copyMessageToClipboard } from '@/lib/clipboard';
import { formatDate } from '@/lib/utils/format';

import FlowAgentIcon from './flow-agent-icon';

const taskPreviewLength = 500;

interface FlowAgentProps {
    log: AgentLogFragmentFragment;
    searchValue?: string;
}

const containsSearchValue = (text: null | string | undefined, searchValue: string): boolean => {
    if (!text || !searchValue.trim()) {
        return false;
    }

    return text.toLowerCase().includes(searchValue.toLowerCase().trim());
};

function FlowAgent({ log, searchValue = '' }: FlowAgentProps) {
    const { createdAt, executor, initiator, result, subtaskId, task, taskId } = log;

    const searchChecks = useMemo(() => {
        const trimmedSearch = searchValue.trim();

        if (!trimmedSearch) {
            return { hasResultMatch: false, hasTaskMatch: false };
        }

        return {
            hasResultMatch: containsSearchValue(result, trimmedSearch),
            hasTaskMatch: containsSearchValue(task, trimmedSearch),
        };
    }, [searchValue, task, result]);

    const [isDetailsVisible, setIsDetailsVisible] = useState(false);

    const [prevSearchValue, setPrevSearchValue] = useState(searchValue);
    const [prevHasResultMatch, setPrevHasResultMatch] = useState(searchChecks.hasResultMatch);

    if (searchValue !== prevSearchValue || searchChecks.hasResultMatch !== prevHasResultMatch) {
        setPrevSearchValue(searchValue);
        setPrevHasResultMatch(searchChecks.hasResultMatch);

        const trimmedSearch = searchValue.trim();

        if (trimmedSearch) {
            if (searchChecks.hasResultMatch) {
                setIsDetailsVisible(true);
            }
        } else {
            setIsDetailsVisible(false);
        }
    }

    const shouldShowFullTask = searchChecks.hasTaskMatch || isDetailsVisible || task.length <= taskPreviewLength;
    const taskToShow = shouldShowFullTask ? task : `${task.slice(0, taskPreviewLength)}...`;

    const shouldShowDetailsToggle = result || task.length > taskPreviewLength;

    const handleCopy = useCallback(async () => {
        await copyMessageToClipboard({
            message: task,
            result: result || undefined,
        });
    }, [task, result]);

    return (
        <div className="flex flex-col items-start">
            <div className="bg-card text-card-foreground max-w-full rounded-xl border p-3 shadow-sm">
                <Markdown
                    className="prose-xs prose-fixed wrap-break-word"
                    searchValue={searchValue}
                >
                    {taskToShow}
                </Markdown>
                {shouldShowDetailsToggle && (
                    <div className="text-muted-foreground mt-2 text-xs">
                        <div
                            className="cursor-pointer"
                            onClick={() => setIsDetailsVisible(!isDetailsVisible)}
                        >
                            {isDetailsVisible ? 'Hide details' : 'Show details'}
                        </div>
                        {isDetailsVisible && result && (
                            <>
                                <div className="my-3 border-t" />
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
            <div className="text-muted-foreground mt-1 flex items-center gap-1 px-1 text-xs">
                <span className="flex items-center gap-0.5">
                    <FlowAgentIcon
                        className="text-muted-foreground"
                        type={initiator}
                    />
                    <span className="text-muted-foreground/50">→</span>
                    <FlowAgentIcon
                        className="text-muted-foreground"
                        type={executor}
                    />
                </span>
                <Tooltip>
                    <TooltipTrigger asChild>
                        <Copy
                            className="hover:text-foreground mx-1 size-3 shrink-0 cursor-pointer transition-colors"
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
}

export default memo(FlowAgent);
