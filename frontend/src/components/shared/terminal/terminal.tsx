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

const Terminal = ({
    className,
    logs,
    ref,
    searchValue,
}: TerminalProps & { ref?: React.RefObject<null | TerminalRef> }) => {
    const { theme } = useTheme();
    const { clear, containerRef, isReady, scrollToBottom, searchAddon, write } = useXterm({ theme });
    const { findNext, findPrevious } = useTerminalSearch(searchAddon, isReady, searchValue, theme);

    const lastLogIndexRef = useRef(0);
    const prevLogsLengthRef = useRef(0);

    useImperativeHandle(ref, () => ({ findNext, findPrevious }), [findNext, findPrevious]);

    useEffect(() => {
        if (!isReady) {
            return;
        }

        if (logs.length === 0 && prevLogsLengthRef.current > 0) {
            clear();
            lastLogIndexRef.current = 0;
            prevLogsLengthRef.current = 0;

            return;
        }

        if (logs.length === 0) {
            return;
        }

        if (logs.length >= lastLogIndexRef.current) {
            const newLogs = logs.slice(lastLogIndexRef.current);

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

        lastLogIndexRef.current = logs.length;
        prevLogsLengthRef.current = logs.length;
    }, [logs, isReady, write, clear, scrollToBottom]);

    return (
        <div
            className={cn('overflow-hidden', className)}
            ref={containerRef}
            style={{ contain: 'strict' }}
        />
    );
};

Terminal.displayName = 'Terminal';

export type { TerminalRef };
export default Terminal;
