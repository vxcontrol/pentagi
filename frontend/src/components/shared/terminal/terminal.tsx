import { useEffect, useImperativeHandle, useRef } from 'react';

import { useTheme } from '@/hooks/use-theme';
import { cn } from '@/lib/utils';

import { processLog } from './terminal-sanitizer';
import { useTerminalSearch } from './use-terminal-search';
import { useXterm } from './use-xterm';

interface TerminalProps {
    className?: string;
    logs: string[];
    searchValue?: string;
}

interface TerminalRef {
    findNext: () => void;
    findPrevious: () => void;
}

function Terminal({
    className,
    logs,
    ref,
    searchValue,
}: TerminalProps & { ref?: React.RefObject<null | TerminalRef> }) {
    const { theme } = useTheme();
    const { clear, containerRef, isReady, scrollToBottom, searchAddon, write } = useXterm({ theme });
    const { findNext, findPrevious } = useTerminalSearch(searchAddon, isReady, searchValue, theme);

    const renderedLogsRef = useRef<string[]>([]);

    useImperativeHandle(ref, () => ({ findNext, findPrevious }), [findNext, findPrevious]);

    useEffect(() => {
        if (!isReady) {
            return;
        }

        const rendered = renderedLogsRef.current;

        if (logs.length === 0) {
            if (rendered.length > 0) {
                clear();
                renderedLogsRef.current = [];
            }

            return;
        }

        // Append the tail only when the on-screen lines are an exact prefix of
        // the new array. `logs` is also fed filtered subsets (flow-terminal), so
        // a length-only check reappends the tail over stale lines when it shrinks
        // then grows back.
        const isAppend = logs.length >= rendered.length && rendered.every((line, index) => line === logs[index]);

        if (isAppend) {
            const newLogs = logs.slice(rendered.length);

            if (newLogs.length > 0) {
                const batch = newLogs.filter(Boolean).map(processLog).join('\r\n');

                if (batch) {
                    write(batch + '\r\n');
                    scrollToBottom();
                }
            }
        } else {
            clear();

            const batch = logs.filter(Boolean).map(processLog).join('\r\n');

            if (batch) {
                write(batch + '\r\n');
            }

            scrollToBottom();
        }

        renderedLogsRef.current = logs;
    }, [logs, isReady, write, clear, scrollToBottom]);

    return (
        <div
            className={cn('overflow-hidden', className)}
            ref={containerRef}
            style={{ contain: 'strict' }}
        />
    );
}

export type { TerminalRef };
export default Terminal;
