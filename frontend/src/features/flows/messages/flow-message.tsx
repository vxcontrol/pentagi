import { Copy } from 'lucide-react';
import { memo, useCallback, useEffect, useMemo, useState } from 'react';

import type { AssistantLogFragmentFragment, MessageLogFragmentFragment } from '@/graphql/types';

import Markdown from '@/components/shared/markdown';
import Terminal from '@/components/shared/terminal';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { MessageLogType, ResultFormat } from '@/graphql/types';
import { cn } from '@/lib/utils';
import { formatDate } from '@/lib/utils/format';
import { copyMessageToClipboard } from '@/lib/сlipboard';

import FlowMessageTypeIcon from './flow-message-type-icon';

interface FlowMessageProps {
    log: AssistantLogFragmentFragment | MessageLogFragmentFragment;
    searchValue?: string;
}

// Helper function to check if text contains search value (case-insensitive)
const containsSearchValue = (text: null | string | undefined, searchValue: string): boolean => {
    if (!text || !searchValue.trim()) {
        return false;
    }

    return text.toLowerCase().includes(searchValue.toLowerCase().trim());
};

const FlowMessage = ({ log, searchValue = '' }: FlowMessageProps) => {
    const { createdAt, message, result, resultFormat = ResultFormat.Plain, thinking, type } = log;
    const isReportMessage = type === MessageLogType.Report;

    // Memoize search checks to avoid recalculating on every render
    const searchChecks = useMemo(() => {
        const trimmedSearch = searchValue.trim();

        if (!trimmedSearch) {
            return { hasResultMatch: false, hasThinkingMatch: false };
        }

        return {
            hasResultMatch: containsSearchValue(result, trimmedSearch),
            hasThinkingMatch: containsSearchValue(thinking, trimmedSearch),
        };
    }, [searchValue, thinking, result]);

    const [isDetailsVisible, setIsDetailsVisible] = useState(isReportMessage);
    const [isThinkingVisible, setIsThinkingVisible] = useState(false);

    // Auto-expand blocks if they contain search matches
    useEffect(() => {
        const trimmedSearch = searchValue.trim();

        if (trimmedSearch) {
            // Expand thinking block only if it contains the search term
            if (searchChecks.hasThinkingMatch) {
                setIsThinkingVisible(true);
            }

            // Expand result block only if it contains the search term
            if (searchChecks.hasResultMatch) {
                setIsDetailsVisible(true);
            }
        } else {
            // Reset to default state when search is cleared
            setIsDetailsVisible(isReportMessage);
            setIsThinkingVisible(false);
        }
    }, [searchValue, searchChecks.hasThinkingMatch, searchChecks.hasResultMatch, isReportMessage]);

    // Use useCallback to memoize the toggle functions
    const toggleDetails = useCallback(() => {
        setIsDetailsVisible((prev) => !prev);
    }, []);

    const toggleThinking = useCallback(() => {
        setIsThinkingVisible((prev) => !prev);
    }, []);

    const handleCopy = useCallback(async () => {
        await copyMessageToClipboard({
            message,
            result,
            resultFormat,
            thinking,
        });
    }, [thinking, message, result, resultFormat]);

    // Determine if thinking should be shown
    // Show thinking if: thinking exists AND (message is empty OR thinking is manually toggled visible)
    const shouldShowThinking = thinking && (!message || isThinkingVisible);

    // Determine if thinking toggle button should be shown
    // Show button only if thinking exists AND message is not empty
    const shouldShowThinkingToggle = thinking && message;

    // Only render details content when it's visible to reduce DOM nodes
    const renderDetailsContent = () => {
        if (!isDetailsVisible) {
            return null;
        }

        return (
            <>
                <div className="my-2 border-t" />
                {resultFormat === ResultFormat.Plain && (
                    <Markdown
                        className="prose-xs prose-fixed text-accent-foreground text-sm wrap-break-word"
                        searchValue={searchValue}
                    >
                        {result}
                    </Markdown>
                )}
                {resultFormat === ResultFormat.Markdown && (
                    <Markdown
                        className="prose-xs prose-fixed wrap-break-word"
                        searchValue={searchValue}
                    >
                        {result}
                    </Markdown>
                )}
                {resultFormat === ResultFormat.Terminal && (
                    <Terminal
                        className="bg-card h-[240px] w-full py-1 pl-1"
                        logs={[result as string]}
                    />
                )}
            </>
        );
    };

    const renderThinkingContent = () => {
        if (!shouldShowThinking) {
            return null;
        }

        return (
            <>
                <div className="border-muted mb-3 border-l-2 pl-3">
                    <Markdown
                        className="prose-xs prose-fixed text-muted-foreground/80 wrap-break-word"
                        searchValue={searchValue}
                    >
                        {thinking}
                    </Markdown>
                </div>
            </>
        );
    };

    return (
        <div className={`flex flex-col ${type === MessageLogType.Input ? 'items-end' : 'items-start'}`}>
            <div
                className={cn(
                    'bg-card text-card-foreground max-w-[90%] rounded-xl border p-3',
                    resultFormat === ResultFormat.Terminal && isDetailsVisible ? 'w-full' : '',
                )}
            >
                {/* Thinking toggle button */}
                {shouldShowThinkingToggle && (
                    <div className="text-muted-foreground mb-2 text-xs">
                        <div
                            className="cursor-pointer"
                            onClick={toggleThinking}
                        >
                            {isThinkingVisible ? 'Hide thinking' : 'Show thinking'}
                        </div>
                    </div>
                )}

                {/* Thinking content */}
                {renderThinkingContent()}

                {/* Main message content */}
                {message && (
                    <Markdown
                        className="prose-xs prose-fixed wrap-break-word"
                        searchValue={searchValue}
                    >
                        {message}
                    </Markdown>
                )}

                {/* Result details */}
                {result && (
                    <div className="text-muted-foreground mt-2 text-xs">
                        <div
                            className="cursor-pointer"
                            onClick={toggleDetails}
                        >
                            {isDetailsVisible ? 'Hide details' : 'Show details'}
                        </div>
                        {renderDetailsContent()}
                    </div>
                )}
            </div>
            <div
                className={`text-muted-foreground mt-1 flex items-center gap-1 px-1 text-xs ${
                    type === MessageLogType.Input ? 'flex-row-reverse' : 'flex-row'
                }`}
            >
                <FlowMessageTypeIcon type={type} />
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
                <span className="text-muted-foreground/50">{log.id}</span>
            </div>
        </div>
    );
};

export default memo(FlowMessage);
