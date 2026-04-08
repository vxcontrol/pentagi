import '@xterm/xterm/css/xterm.css';
import type { ILinkHandler } from '@xterm/xterm';

import { FitAddon } from '@xterm/addon-fit';
import { SearchAddon } from '@xterm/addon-search';
import { Unicode11Addon } from '@xterm/addon-unicode11';
import { WebLinksAddon } from '@xterm/addon-web-links';
import { WebglAddon } from '@xterm/addon-webgl';
import { Terminal } from '@xterm/xterm';
import { useCallback, useEffect, useRef, useState } from 'react';

import { Log } from '@/lib/log';
import { isMac } from '@/lib/utils/platform';

import { getTerminalTheme, isDarkMode, TERMINAL_OPTIONS } from './terminal-config';
import { SAFE_PROTOCOLS } from './terminal-sanitizer';

const FLOW_CONTROL_CHUNK_SIZE = 64 * 1024;
const DEBOUNCE_DELAY = 150;
const TOOLTIP_CLASS = 'terminal-link-tooltip';

export interface UseXtermResult {
    clear: () => void;
    containerRef: React.RefObject<HTMLDivElement | null>;
    isReady: boolean;
    scrollToBottom: () => void;
    searchAddon: null | SearchAddon;
    write: (data: string) => void;
}

/**
 * Manages the full xterm.js lifecycle: creation, addon loading,
 * resize handling, WebGL fallback, theme sync, and cleanup.
 *
 * Implements chunked flow control for large writes per
 * https://xtermjs.org/docs/guides/flowcontrol/
 *
 * Requires Ctrl+Click (Cmd+Click on Mac) to open links per
 * https://xtermjs.org/docs/guides/link-handling/
 */
export function useXterm({ theme }: { theme: 'dark' | 'light' | 'system' }): UseXtermResult {
    const containerRef = useRef<HTMLDivElement | null>(null);
    const terminalRef = useRef<null | Terminal>(null);
    const themeRef = useRef(theme);
    const [isReady, setIsReady] = useState(false);
    const [searchAddon, setSearchAddon] = useState<null | SearchAddon>(null);

    useEffect(() => {
        themeRef.current = theme;
    }, [theme]);

    const write = useCallback((data: string) => {
        const terminal = terminalRef.current;

        if (!terminal) {
            return;
        }

        try {
            writeWithFlowControl(terminal, data);
        } catch (error: unknown) {
            Log.error('Terminal write failed:', error);
        }
    }, []);

    const clear = useCallback(() => {
        try {
            terminalRef.current?.clear();
        } catch (error: unknown) {
            Log.error('Terminal clear failed:', error);
        }
    }, []);

    const scrollToBottom = useCallback(() => {
        try {
            terminalRef.current?.scrollToBottom();
        } catch (error: unknown) {
            Log.error('Terminal scrollToBottom failed:', error);
        }
    }, []);

    useEffect(() => {
        const container = containerRef.current;

        if (!container) {
            return;
        }

        let mounted = true;

        const terminal = new Terminal({
            ...TERMINAL_OPTIONS,
            theme: getTerminalTheme(isDarkMode(themeRef.current)),
        });

        const fitAddon = new FitAddon();
        const search = new SearchAddon();
        const unicodeAddon = new Unicode11Addon();

        const mac = isMac();

        const openLink = (event: MouseEvent, uri: string) => {
            const uriLower = uri.toLowerCase();

            if (
                (mac ? event.metaKey : event.ctrlKey) &&
                SAFE_PROTOCOLS.some((p) => uriLower.startsWith(p))
            ) {
                window.open(uri, '_blank', 'noopener,noreferrer');
            }
        };

        const linkHandler: ILinkHandler = {
            activate: (event, text) => openLink(event, text),
            allowNonHttpProtocols: true,
            hover: (event, text) => showLinkTooltip(container, event, text, mac),
            leave: () => removeLinkTooltip(container),
        };

        terminal.options.linkHandler = linkHandler;

        const webLinksAddon = new WebLinksAddon(openLink, {
            hover: (event, text) => showLinkTooltip(container, event, text, mac),
            leave: () => removeLinkTooltip(container),
        });

        terminal.loadAddon(fitAddon);
        terminal.loadAddon(search);
        terminal.loadAddon(unicodeAddon);
        terminal.unicode.activeVersion = '11';
        terminal.loadAddon(webLinksAddon);

        const disposables: Array<{ dispose: () => void }> = [unicodeAddon, webLinksAddon];

        try {
            const webglAddon = new WebglAddon();
            terminal.loadAddon(webglAddon);
            disposables.push(webglAddon);
            disposables.push(
                webglAddon.onContextLoss(() => {
                    try {
                        webglAddon.dispose();
                    } catch {
                        /* ignore */
                    }
                }),
            );
        } catch {
            /* WebGL not available — canvas renderer is used as fallback */
        }

        const safeFit = () => {
            try {
                if (container.offsetHeight > 0) {
                    fitAddon.fit();
                }
            } catch (error: unknown) {
                Log.error('Terminal fit failed:', error);
            }
        };

        let debounceTimer: null | ReturnType<typeof setTimeout> = null;
        const resizeObserver = new ResizeObserver(() => {
            if (debounceTimer) {
                clearTimeout(debounceTimer);
            }

            debounceTimer = setTimeout(safeFit, DEBOUNCE_DELAY);
        });

        terminalRef.current = terminal;

        const rafId = requestAnimationFrame(() => {
            if (!mounted) {
                return;
            }

            try {
                terminal.open(container);
                resizeObserver.observe(container);
                safeFit();
                setSearchAddon(search);
                setIsReady(true);
            } catch (error: unknown) {
                Log.error('Failed to open terminal:', error);
            }
        });

        return () => {
            mounted = false;
            cancelAnimationFrame(rafId);

            if (debounceTimer) {
                clearTimeout(debounceTimer);
            }

            removeLinkTooltip(container);
            resizeObserver.disconnect();

            for (const d of disposables) {
                try {
                    d.dispose();
                } catch {
                    /* ignore */
                }
            }

            try {
                search.dispose();
            } catch {
                /* ignore */
            }

            try {
                fitAddon.dispose();
            } catch {
                /* ignore */
            }

            try {
                terminal.dispose();
            } catch {
                /* ignore */
            }

            terminalRef.current = null;
            setSearchAddon(null);
            setIsReady(false);
        };
    }, []);

    useEffect(() => {
        const terminal = terminalRef.current;

        if (!terminal) {
            return;
        }

        const applyTheme = () => {
            try {
                terminal.options.theme = getTerminalTheme(isDarkMode(theme));
            } catch (error: unknown) {
                Log.error('Terminal theme update failed:', error);
            }
        };

        applyTheme();

        if (theme === 'system') {
            const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
            const handler = () => applyTheme();
            mediaQuery.addEventListener('change', handler);

            return () => {
                mediaQuery.removeEventListener('change', handler);
            };
        }
    }, [theme]);

    return { clear, containerRef, isReady, scrollToBottom, searchAddon, write };
}

function removeLinkTooltip(container: HTMLElement): void {
    const existing = container.querySelector(`.${TOOLTIP_CLASS}`);

    if (existing) {
        existing.remove();
    }
}

function showLinkTooltip(container: HTMLElement, event: MouseEvent, uri: string, mac: boolean): void {
    removeLinkTooltip(container);

    const tooltip = document.createElement('div');
    tooltip.className = `${TOOLTIP_CLASS} xterm-hover`;
    tooltip.style.cssText =
        'position:absolute;z-index:10;padding:4px 8px;max-width:80%;font-size:12px;' +
        'line-height:1.4;border-radius:4px;pointer-events:none;word-break:break-all;' +
        'background:var(--popover);color:var(--popover-foreground);' +
        'border:1px solid var(--border);box-shadow:0 2px 8px rgba(0,0,0,.15)';

    const urlSpan = document.createElement('span');
    urlSpan.textContent = uri;
    tooltip.appendChild(urlSpan);

    const hint = document.createElement('div');
    hint.style.cssText = 'opacity:0.6;font-size:11px;margin-top:2px';
    hint.textContent = `${mac ? 'Cmd' : 'Ctrl'}+Click to open`;
    tooltip.appendChild(hint);

    container.style.position = 'relative';
    container.appendChild(tooltip);

    const rect = container.getBoundingClientRect();
    const x = event.clientX - rect.left;
    let y = event.clientY - rect.top + 20;

    tooltip.style.left = `${Math.min(x, rect.width - tooltip.offsetWidth - 8)}px`;

    if (y + tooltip.offsetHeight > rect.height) {
        y = event.clientY - rect.top - tooltip.offsetHeight - 8;
    }

    tooltip.style.top = `${Math.max(0, y)}px`;
}

/**
 * Writes data to the terminal with chunked flow control.
 * For large payloads, splits into chunks and uses xterm's write callback
 * to avoid overwhelming the parser and blocking the UI thread.
 *
 * Chunk boundaries are adjusted to avoid splitting UTF-16 surrogate pairs:
 * if the last code unit of a chunk is a high surrogate (0xD800-0xDBFF),
 * the boundary is moved back by one so the pair stays intact.
 *
 * See: https://xtermjs.org/docs/guides/flowcontrol/
 */
function writeWithFlowControl(terminal: Terminal, data: string): void {
    if (data.length <= FLOW_CONTROL_CHUNK_SIZE) {
        terminal.write(data);

        return;
    }

    let offset = 0;

    const writeNext = () => {
        let end = Math.min(offset + FLOW_CONTROL_CHUNK_SIZE, data.length);

        if (end < data.length) {
            const code = data.charCodeAt(end - 1);

            if (code >= 0xd800 && code <= 0xdbff) {
                end--;
            }
        }

        const chunk = data.slice(offset, end);
        offset = end;

        if (offset < data.length) {
            terminal.write(chunk, writeNext);
        } else {
            terminal.write(chunk);
        }
    };

    writeNext();
}
