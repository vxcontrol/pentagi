import '@xterm/xterm/css/xterm.css';

import { FitAddon } from '@xterm/addon-fit';
import { SearchAddon } from '@xterm/addon-search';
import { Unicode11Addon } from '@xterm/addon-unicode11';
import { WebLinksAddon } from '@xterm/addon-web-links';
import { WebglAddon } from '@xterm/addon-webgl';
import type { ITerminalOptions, ITheme } from '@xterm/xterm';
import { Terminal as XTerminal } from '@xterm/xterm';
import debounce from 'lodash/debounce';
import { useEffect, useImperativeHandle, useRef, useState } from 'react';

import { Log } from '@/lib/log';
import { cn } from '@/lib/utils';
import { useTheme } from '@/providers/ThemeProvider';

const terminalOptions: ITerminalOptions = {
    convertEol: true,
    fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace',
    fontSize: 12,
    fontWeight: 600,
    allowTransparency: true,
    cursorBlink: false,
    disableStdin: true,
    allowProposedApi: true,
    customGlyphs: true,
    scrollback: 2500,
    screenReaderMode: false,
    fastScrollSensitivity: 10,
    fastScrollModifier: 'alt',
    smoothScrollDuration: 0, // Disable smooth scrolling
} as const;

// Search decoration styles for dark theme - using HEX format as required
const darkSearchDecorations = {
    matchBackground: '#666666',
    matchOverviewRuler: '#000000',
    activeMatchBackground: '#AAAAAA',
    activeMatchColorOverviewRuler: '#000000',
} as const;

// Search decoration styles for light theme - using HEX format as required
const lightSearchDecorations = {
    matchBackground: '#000000',
    matchOverviewRuler: '#000000',
    activeMatchBackground: '#555555',
    activeMatchColorOverviewRuler: '#000000',
} as const;

const darkTheme: ITheme = {
    background: '#020817',
    foreground: '#f4f4f5',
    cursor: '#f4f4f5',
    cursorAccent: '#f4f4f5',
    selectionBackground: 'rgba(96, 165, 250, 0.2)',
    black: '#f4f4f5',
    red: '#f87171',
    green: '#4ade80',
    yellow: '#facc15',
    blue: '#60a5fa',
    magenta: '#c084fc',
    cyan: '#22d3ee',
    white: '#020817',
    brightBlack: '#e4e4e7',
    brightRed: '#fca5a5',
    brightGreen: '#86efac',
    brightYellow: '#fde047',
    brightBlue: '#93c5fd',
    brightMagenta: '#d8b4fe',
    brightCyan: '#67e8f9',
    brightWhite: '#71717a',
} as const;

const lightTheme: ITheme = {
    background: '#ffffff',
    foreground: '#020817',
    cursor: '#020817',
    cursorAccent: '#020817',
    selectionBackground: 'rgba(59, 130, 246, 0.1)',
    black: '#020817',
    red: '#ef4444',
    green: '#22c55e',
    yellow: '#eab308',
    blue: '#3b82f6',
    magenta: '#a855f7',
    cyan: '#06b6d4',
    white: '#e2e8f0',
    brightBlack: '#64748b',
    brightRed: '#f87171',
    brightGreen: '#4ade80',
    brightYellow: '#facc15',
    brightBlue: '#60a5fa',
    brightMagenta: '#c084fc',
    brightCyan: '#22d3ee',
    brightWhite: '#f1f5f9',
} as const;

interface TerminalProps {
    logs: string[];
    className?: string;
    searchValue?: string;
}

interface TerminalRef {
    findNext: () => void;
    findPrevious: () => void;
}

const Terminal = ({ ref, logs, className, searchValue }: TerminalProps & { ref?: React.RefObject<TerminalRef | null> }) => {
    const terminalRef = useRef<HTMLDivElement>(null);
    const xtermRef = useRef<XTerminal | null>(null);
    const fitAddonRef = useRef<FitAddon | null>(null);
    const searchAddonRef = useRef<SearchAddon | null>(null);
    const lastLogIndexRef = useRef<number>(0);
    const webglAddonRef = useRef<WebglAddon | null>(null);
    const resizeObserverRef = useRef<ResizeObserver | null>(null);
    const debouncedFitRef = useRef<ReturnType<typeof debounce> | null>(null);
    const { theme } = useTheme();
    const [isTerminalOpened, setIsTerminalOpened] = useState(false);
    const [isTerminalReady, setIsTerminalReady] = useState(false);
    const isTerminalReadyRef = useRef(false);
    const prevLogsLengthRef = useRef<number>(0);
    const terminalInitializedRef = useRef(false);
    const isMountedRef = useRef(true);
    const initTimeoutRef = useRef<NodeJS.Timeout | null>(null);
    const fitTimeoutRef = useRef<NodeJS.Timeout | null>(null);

    // Get search decorations based on current theme
    const getSearchDecorations = () => {
        return theme === 'dark' ? darkSearchDecorations : lightSearchDecorations;
    };

    // Expose methods to parent component via ref
    useImperativeHandle(
        ref,
        () => ({
            findNext: () => {
                if (searchAddonRef.current && searchValue?.trim()) {
                    try {
                        searchAddonRef.current.findNext(searchValue.trim(), {
                            caseSensitive: false,
                            wholeWord: false,
                            regex: false,
                            decorations: getSearchDecorations(),
                        });
                    } catch (error: unknown) {
                        Log.error('Terminal findNext failed:', error);
                    }
                }
            },
            findPrevious: () => {
                if (searchAddonRef.current && searchValue?.trim()) {
                    try {
                        searchAddonRef.current.findPrevious(searchValue.trim(), {
                            caseSensitive: false,
                            wholeWord: false,
                            regex: false,
                            decorations: getSearchDecorations(),
                        });
                    } catch (error: unknown) {
                        Log.error('Terminal findPrevious failed:', error);
                    }
                }
            },
        }),
        [searchValue, theme],
    );

    // Safe terminal operations
    const safeTerminalOperation = (operation: () => void) => {
        try {
            if (isMountedRef.current && xtermRef.current) {
                operation();
            }
        } catch (error: unknown) {
            Log.error('Terminal operation failed:', error);
        }
    };

    // Safe fit
    const safeFit = () => {
        try {
            if (
                isMountedRef.current &&
                fitAddonRef.current &&
                terminalRef.current &&
                terminalRef.current.offsetHeight > 0 &&
                xtermRef.current
            ) {
                fitAddonRef.current.fit();
            }
        } catch (error: unknown) {
            Log.error('Terminal fit failed:', error);
        }
    };

    // Clear all timeouts
    const clearAllTimeouts = () => {
        if (initTimeoutRef.current) {
            clearTimeout(initTimeoutRef.current);
            initTimeoutRef.current = null;
        }
        if (fitTimeoutRef.current) {
            clearTimeout(fitTimeoutRef.current);
            fitTimeoutRef.current = null;
        }
    };

    // Track component mount/unmount
    useEffect(() => {
        isMountedRef.current = true;

        return () => {
            isMountedRef.current = false;
            clearAllTimeouts();
        };
    }, []);

    // Initialize terminal - only once
    useEffect(() => {
        if (!terminalRef.current || terminalInitializedRef.current || !isMountedRef.current) {
            return;
        }

        terminalInitializedRef.current = true;

        try {
            // Create terminal instance with optimized settings
            const terminal = new XTerminal({
                ...terminalOptions,
                theme: theme === 'dark' ? darkTheme : lightTheme,
            });

            xtermRef.current = terminal;

            // Add addons before opening terminal
            const fitAddon = new FitAddon();
            fitAddonRef.current = fitAddon;
            terminal.loadAddon(fitAddon);

            const searchAddon = new SearchAddon();
            searchAddonRef.current = searchAddon;
            terminal.loadAddon(searchAddon);

            const unicodeAddon = new Unicode11Addon();
            terminal.loadAddon(unicodeAddon);
            terminal.unicode.activeVersion = '11';

            const webLinksAddon = new WebLinksAddon();
            terminal.loadAddon(webLinksAddon);

            // Add WebGL addon last (and optionally)
            try {
                const webglAddon = new WebglAddon();
                webglAddonRef.current = webglAddon;
                terminal.loadAddon(webglAddon);
                webglAddon.onContextLoss(() => {
                    if (isMountedRef.current && webglAddonRef.current) {
                        webglAddonRef.current.dispose();
                    }
                });
            } catch {
                // Ignore WebGL errors
            }

            // Set up resize handler
            const debouncedFit = debounce(() => {
                if (isMountedRef.current && isTerminalReadyRef.current) {
                    safeFit();
                }
            }, 150);

            debouncedFitRef.current = debouncedFit;

            const resizeObserver = new ResizeObserver(() => {
                if (isMountedRef.current && isTerminalReadyRef.current) {
                    debouncedFit();
                }
            });

            resizeObserverRef.current = resizeObserver;

            // Open terminal with delay
            // This approach ensures the DOM is ready for rendering
            initTimeoutRef.current = setTimeout(() => {
                if (!isMountedRef.current || !terminalRef.current || !xtermRef.current) {
                    return;
                }

                try {
                    terminal.open(terminalRef.current);
                    setIsTerminalOpened(true);

                    // Observe size changes only after successful terminal opening
                    if (terminalRef.current && resizeObserverRef.current) {
                        resizeObserverRef.current.observe(terminalRef.current);
                    }

                    // Set size with delay to allow DOM to render terminal
                    fitTimeoutRef.current = setTimeout(() => {
                        if (isMountedRef.current) {
                            safeFit();
                            // Mark terminal as fully ready only after successful fit()
                            isTerminalReadyRef.current = true;
                            setIsTerminalReady(true);
                        }
                    }, 200);
                } catch (error: unknown) {
                    Log.error('Failed to open terminal:', error);
                }
            }, 100);

            return () => {
                // Cleanup on unmount
                if (initTimeoutRef.current) clearTimeout(initTimeoutRef.current);
                if (fitTimeoutRef.current) clearTimeout(fitTimeoutRef.current);
                clearAllTimeouts();

                if (resizeObserverRef.current) {
                    resizeObserverRef.current.disconnect();
                    resizeObserverRef.current = null;
                }

                if (debouncedFitRef.current) {
                    debouncedFitRef.current.cancel();
                    debouncedFitRef.current = null;
                }

                if (searchAddonRef.current) {
                    try {
                        searchAddonRef.current.dispose();
                    } catch {
                        // Ignore errors during disposal
                    }
                    searchAddonRef.current = null;
                }

                if (webglAddonRef.current) {
                    try {
                        webglAddonRef.current.dispose();
                    } catch {
                        // Ignore errors during disposal
                    }
                    webglAddonRef.current = null;
                }

                if (fitAddonRef.current) {
                    try {
                        fitAddonRef.current.dispose();
                    } catch {
                        // Ignore errors during disposal
                    }
                    fitAddonRef.current = null;
                }

                if (xtermRef.current) {
                    try {
                        xtermRef.current.dispose();
                    } catch {
                        // Ignore errors during disposal
                    }
                    xtermRef.current = null;
                }

                lastLogIndexRef.current = 0;
                prevLogsLengthRef.current = 0;
                terminalInitializedRef.current = false;
                setIsTerminalOpened(false);
                isTerminalReadyRef.current = false;
                setIsTerminalReady(false);
            };
        } catch (error: unknown) {
            Log.error('Terminal initialization failed:', error);
            terminalInitializedRef.current = false;
            return;
        }
    }, [theme]);

    // Handle search functionality with decorations
    useEffect(() => {
        if (!searchAddonRef.current || !isTerminalReady || !isMountedRef.current) {
            return;
        }

        const searchAddon = searchAddonRef.current;

        try {
            if (searchValue && searchValue.trim()) {
                // Perform search with theme-appropriate decorations
                searchAddon.findNext(searchValue.trim(), {
                    caseSensitive: false,
                    wholeWord: false,
                    regex: false,
                    decorations: getSearchDecorations(),
                });
            } else {
                // Clear search highlighting when search value is empty
                searchAddon.clearDecorations();
            }
        } catch (error: unknown) {
            Log.error('Terminal search failed:', error);
        }
    }, [searchValue, isTerminalReady, theme]);

    // Update theme
    useEffect(() => {
        safeTerminalOperation(() => {
            if (xtermRef.current) {
                xtermRef.current.options.theme = theme === 'dark' ? darkTheme : lightTheme;
            }
        });
    }, [theme]);

    // Update logs only when terminal is fully ready
    useEffect(() => {
        if (!isMountedRef.current || !xtermRef.current || !isTerminalOpened || !isTerminalReady) {
            return;
        }

        const terminal = xtermRef.current;

        try {
            if (logs?.length === 0 && prevLogsLengthRef.current > 0) {
                safeTerminalOperation(() => {
                    terminal.clear();
                });
                lastLogIndexRef.current = 0;
                prevLogsLengthRef.current = 0;
                return;
            }

            if (!logs?.length) {
                return;
            }

            if (logs.length >= lastLogIndexRef.current) {
                const newLogs = logs.slice(lastLogIndexRef.current);

                if (newLogs.length === 0) {
                    return;
                }

                // Add logs in batch for performance optimization
                safeTerminalOperation(() => {
                    for (const log of newLogs.filter(Boolean)) {
                        terminal.writeln(log);
                    }

                    // Scroll down only once after adding all logs
                    if (newLogs.length > 0) {
                        terminal.scrollToBottom();
                    }
                });

                lastLogIndexRef.current = logs.length;
                prevLogsLengthRef.current = logs.length;
            } else {
                // If logs were reset (became fewer)
                safeTerminalOperation(() => {
                    terminal.clear();

                    // Add all logs in batch again
                    for (const log of logs.filter(Boolean)) {
                        terminal.writeln(log);
                    }

                    terminal.scrollToBottom();
                });

                lastLogIndexRef.current = logs.length;
                prevLogsLengthRef.current = logs.length;
            }
        } catch (error) {
            Log.error('Terminal log update failed:', error);
        }
    }, [logs, isTerminalOpened, isTerminalReady]);

    return (
        <div
            ref={terminalRef}
            className={cn('overflow-hidden', className)}
        />
    );
};

Terminal.displayName = 'Terminal';

export default Terminal;
