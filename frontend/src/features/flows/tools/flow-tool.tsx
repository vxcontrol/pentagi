import { Copy, Hammer } from 'lucide-react';
import { memo, useCallback, useEffect, useMemo, useState } from 'react';

import type { SearchLogFragmentFragment } from '@/graphql/types';

import Markdown from '@/components/shared/markdown';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import FlowAgentIcon from '@/features/flows/agents/flow-agent-icon';
import { formatDate, formatName } from '@/lib/utils/format';
import { copyMessageToClipboard } from '@/lib/сlipboard';

interface FlowToolProps {
    log: SearchLogFragmentFragment;
    searchValue?: string;
}

// Helper function to check if text contains search value (case-insensitive)
const containsSearchValue = (text: null | string | undefined, searchValue: string): boolean => {
    if (!text || !searchValue.trim()) {
        return false;
    }

    return text.toLowerCase().includes(searchValue.toLowerCase().trim());
};

const FlowTool = ({ log, searchValue = '' }: FlowToolProps) => {
    const { createdAt, engine, executor, initiator, query, result, subtaskId, taskId } = log;

    // Memoize search checks to avoid recalculating on every render
    const searchChecks = useMemo(() => {
        const trimmedSearch = searchValue.trim();

        if (!trimmedSearch) {
            return { hasQueryMatch: false, hasResultMatch: false };
        }

        return {
            hasQueryMatch: containsSearchValue(query, trimmedSearch),
            hasResultMatch: containsSearchValue(result, trimmedSearch),
        };
    }, [searchValue, query, result]);

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

    const handleCopy = useCallback(async () => {
        await copyMessageToClipboard({
            message: query,
            result: result || undefined,
        });
    }, [query, result]);

    return (
        <div className="flex flex-col items-start">
            <div className="bg-card text-card-foreground max-w-full rounded-xl border p-3">
                <div className="flex flex-col">
                    <div className="cursor-pointer text-sm font-semibold">
                        <Tooltip>
                            <TooltipTrigger asChild>
                                <span className="inline-flex items-center gap-1">
                                    <Hammer className="text-muted-foreground size-4" />
                                    <span>{formatName(engine)}</span>
                                </span>
                            </TooltipTrigger>
                            <TooltipContent>Tool name</TooltipContent>
                        </Tooltip>
                    </div>

                    <Markdown
                        className="prose-xs prose-fixed wrap-break-word"
                        searchValue={searchValue}
                    >
                        {query}
                    </Markdown>
                </div>
                {result && (
                    <div className="text-muted-foreground mt-2 text-xs">
                        <div
                            className="cursor-pointer"
                            onClick={() => setIsDetailsVisible(!isDetailsVisible)}
                        >
                            {isDetailsVisible ? 'Hide details' : 'Show details'}
                        </div>
                        {isDetailsVisible && (
                            <>
                                <div className="my-2 border-t dark:border-gray-700" />
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
};

export default memo(FlowTool);
