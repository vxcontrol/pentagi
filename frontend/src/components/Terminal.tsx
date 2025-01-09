import '@xterm/xterm/css/xterm.css';

import { FitAddon } from '@xterm/addon-fit';
import { Unicode11Addon } from '@xterm/addon-unicode11';
import { WebLinksAddon } from '@xterm/addon-web-links';
import { WebglAddon } from '@xterm/addon-webgl';
import type { ITerminalOptions, ITheme } from '@xterm/xterm';
import { Terminal as XTerminal } from '@xterm/xterm';
import debounce from 'lodash/debounce';
import { useEffect, useRef, useState } from 'react';

import { cn } from '@/lib/utils';
import { useThemeStore } from '@/store/theme-store';

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
    scrollback: 10000,
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
}

const Terminal = ({ logs, className }: TerminalProps) => {
    const terminalRef = useRef<HTMLDivElement>(null);
    const xtermRef = useRef<XTerminal | null>(null);
    const fitAddonRef = useRef<FitAddon | null>(null);
    const lastLogIndexRef = useRef<number>(0);
    const theme = useThemeStore((store) => store.theme);
    const [isTerminalOpened, setIsTerminalOpened] = useState(false);

    useEffect(() => {
        if (!terminalRef.current || !xtermRef) {
            return;
        }

        const terminal = new XTerminal(terminalOptions);
        const fitAddon = new FitAddon();

        fitAddonRef.current = fitAddon;
        terminal.loadAddon(fitAddon);

        try {
            const webglAddon = new WebglAddon();
            terminal.loadAddon(webglAddon);
            webglAddon.onContextLoss(() => {
                webglAddon.dispose();
            });
        } catch {
            // ignore
        }

        const unicodeAddon = new Unicode11Addon();
        terminal.loadAddon(unicodeAddon);
        terminal.unicode.activeVersion = '11';

        const webLinksAddon = new WebLinksAddon();
        terminal.loadAddon(webLinksAddon);

        if (terminalRef.current && terminalRef.current.offsetHeight > 0) {
            fitAddon.fit();
        }

        xtermRef.current = terminal;

        const debouncedFit = debounce(() => {
            const height = terminalRef.current?.offsetHeight;

            if (height) {
                fitAddon.fit();
            }
        }, 100);

        const resizeObserver = new ResizeObserver(debouncedFit);

        if (terminalRef.current) {
            resizeObserver.observe(terminalRef.current);
        }

        return () => {
            resizeObserver.disconnect();
            debouncedFit.cancel();
            terminal.dispose();
        };
    }, []);

    useEffect(() => {
        if (!xtermRef?.current) {
            return;
        }

        xtermRef.current.options.theme = theme === 'dark' ? darkTheme : lightTheme;
    }, [theme]);

    useEffect(() => {
        const terminal = xtermRef.current;
        const container = terminalRef.current;

        if (!container || !terminal || isTerminalOpened) {
            return;
        }

        terminal.open(container);
        setIsTerminalOpened(true);
    }, [isTerminalOpened]);

    useEffect(() => {
        const terminal = xtermRef.current;

        if (!terminal || !isTerminalOpened) {
            return;
        }

        if (!logs?.length) {
            terminal.clear();
            lastLogIndexRef.current = 0;
            return;
        }

        logs.slice(lastLogIndexRef.current)
            .filter(Boolean)
            .forEach((log) => terminal.writeln(log));

        lastLogIndexRef.current = logs.length;

        terminal.scrollToBottom();
    }, [logs, isTerminalOpened]);

    return (
        <div
            ref={terminalRef}
            className={cn('overflow-hidden', className)}
        />
    );
};

export default Terminal;
