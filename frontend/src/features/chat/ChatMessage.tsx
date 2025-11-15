import { Copy } from 'lucide-react';
import { memo, useCallback, useEffect, useMemo, useState } from 'react';

import Markdown from '@/components/shared/Markdown';
import Terminal from '@/components/shared/Terminal';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import type { AssistantLogFragmentFragment, MessageLogFragmentFragment } from '@/graphql/types';
import { MessageLogType, ResultFormat } from '@/graphql/types';
import { cn } from '@/lib/utils';
import { formatDate } from '@/lib/utils/format';
import { copyMessageToClipboard } from '@/lib/сlipboard';

import ChatMessageTypeIcon from './ChatMessageTypeIcon';

interface ChatMessageProps {
    log: MessageLogFragmentFragment | AssistantLogFragmentFragment;
    searchValue?: string;
}

// Helper function to check if text contains search value (case-insensitive)
const containsSearchValue = (text: string | null | undefined, searchValue: string): boolean => {
    if (!text || !searchValue.trim()) {
        return false;
    }
    return text.toLowerCase().includes(searchValue.toLowerCase().trim());
};

const ChatMessage = ({ log, searchValue = '' }: ChatMessageProps) => {
    const { type, createdAt, message, thinking, result, resultFormat = ResultFormat.Plain } = log;
    const isReportMessage = type === MessageLogType.Report;

    // Memoize search checks to avoid recalculating on every render
    const searchChecks = useMemo(() => {
        const trimmedSearch = searchValue.trim();
        if (!trimmedSearch) {
            return { hasThinkingMatch: false, hasResultMatch: false };
        }

        return {
            hasThinkingMatch: containsSearchValue(thinking, trimmedSearch),
            hasResultMatch: containsSearchValue(result, trimmedSearch),
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
            thinking,
            message,
            result,
            resultFormat,
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
        if (!isDetailsVisible) return null;

        return (
            <>
                <div className="my-2 border-t dark:border-gray-700" />
                {resultFormat === ResultFormat.Plain && (
                    <Markdown
                        className="prose-xs prose-fixed break-words text-sm text-accent-foreground"
                        searchValue={searchValue}
                    >
                        {result}
                    </Markdown>
                )}
                {resultFormat === ResultFormat.Markdown && (
                    <Markdown
                        className="prose-xs prose-fixed break-words"
                        searchValue={searchValue}
                    >
                        {result}
                    </Markdown>
                )}
                {resultFormat === ResultFormat.Terminal && (
                    <Terminal
                        logs={[result as string]}
                        className="h-[240px] w-full bg-card py-1 pl-1"
                    />
                )}
            </>
        );
    };

    const renderThinkingContent = () => {
        if (!shouldShowThinking) return null;

        return (
            <>
                <div className="mb-3 border-l-2 border-muted pl-3">
                    <Markdown
                        className="prose-xs prose-fixed break-words text-muted-foreground/80"
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
                    'max-w-[90%] rounded-lg bg-accent p-3 text-accent-foreground',
                    resultFormat === ResultFormat.Terminal && isDetailsVisible ? 'w-full' : '',
                )}
            >
                {/* Thinking toggle button */}
                {shouldShowThinkingToggle && (
                    <div className="mb-2 text-xs text-muted-foreground">
                        <div
                            onClick={toggleThinking}
                            className="cursor-pointer"
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
                        className="prose-xs prose-fixed break-words"
                        searchValue={searchValue}
                    >
                        {message}
                    </Markdown>
                )}

                {/* Result details */}
                {result && (
                    <div className="mt-2 text-xs text-muted-foreground">
                        <div
                            onClick={toggleDetails}
                            className="cursor-pointer"
                        >
                            {isDetailsVisible ? 'Hide details' : 'Show details'}
                        </div>
                        {renderDetailsContent()}
                    </div>
                )}
            </div>
            <div
                className={`mt-1 flex items-center gap-1 px-1 text-xs text-muted-foreground ${
                    type === MessageLogType.Input ? 'flex-row-reverse' : 'flex-row'
                }`}
            >
                <ChatMessageTypeIcon type={type} />
                <Tooltip>
                    <TooltipTrigger asChild>
                        <Copy
                            className="mx-1 size-3 shrink-0 cursor-pointer transition-colors hover:text-foreground"
                            onClick={handleCopy}
                        />
                    </TooltipTrigger>
                    <TooltipContent>Copy</TooltipContent>
                </Tooltip>
                <span className="text-muted-foreground/50">{formatDate(new Date(createdAt))}</span>
            </div>
        </div>
    );
};

export default memo(ChatMessage);
