import type { ITerminalOptions, ITheme } from '@xterm/xterm';

export const TERMINAL_OPTIONS: ITerminalOptions = {
    allowProposedApi: true,
    allowTransparency: true,
    convertEol: true,
    cursorBlink: false,
    customGlyphs: true,
    disableStdin: true,
    fastScrollSensitivity: 10,
    fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace',
    fontSize: 12,
    fontWeight: 600,
    logLevel: 'off',
    screenReaderMode: false,
    scrollback: 10_000,
    smoothScrollDuration: 0,
} as const;

const DARK_THEME: ITheme = {
    background: '#050c13',
    black: '#f4f4f5',
    blue: '#60a5fa',
    brightBlack: '#e4e4e7',
    brightBlue: '#93c5fd',
    brightCyan: '#67e8f9',
    brightGreen: '#86efac',
    brightMagenta: '#d8b4fe',
    brightRed: '#fca5a5',
    brightWhite: '#71717a',
    brightYellow: '#fde047',
    cursor: '#f4f4f5',
    cursorAccent: '#f4f4f5',
    cyan: '#22d3ee',
    foreground: '#f4f4f5',
    green: '#4ade80',
    magenta: '#c084fc',
    red: '#f87171',
    selectionBackground: 'rgba(96, 165, 250, 0.2)',
    white: '#050c13',
    yellow: '#facc15',
} as const;

const LIGHT_THEME: ITheme = {
    background: '#ffffff',
    black: '#020817',
    blue: '#3b82f6',
    brightBlack: '#64748b',
    brightBlue: '#60a5fa',
    brightCyan: '#22d3ee',
    brightGreen: '#4ade80',
    brightMagenta: '#c084fc',
    brightRed: '#f87171',
    brightWhite: '#f1f5f9',
    brightYellow: '#facc15',
    cursor: '#020817',
    cursorAccent: '#020817',
    cyan: '#06b6d4',
    foreground: '#020817',
    green: '#22c55e',
    magenta: '#a855f7',
    red: '#ef4444',
    selectionBackground: 'rgba(59, 130, 246, 0.1)',
    white: '#e2e8f0',
    yellow: '#eab308',
} as const;

const DARK_SEARCH_DECORATIONS = {
    activeMatchBackground: '#AAAAAA',
    activeMatchColorOverviewRuler: '#000000',
    matchBackground: '#666666',
    matchOverviewRuler: '#000000',
} as const;

const LIGHT_SEARCH_DECORATIONS = {
    activeMatchBackground: '#555555',
    activeMatchColorOverviewRuler: '#000000',
    matchBackground: '#000000',
    matchOverviewRuler: '#000000',
} as const;

export function getSearchDecorations(isDark: boolean) {
    return isDark ? DARK_SEARCH_DECORATIONS : LIGHT_SEARCH_DECORATIONS;
}

export function getTerminalTheme(isDark: boolean): ITheme {
    return isDark ? DARK_THEME : LIGHT_THEME;
}

export function isDarkMode(theme: 'dark' | 'light' | 'system'): boolean {
    if (theme === 'dark') {
        return true;
    }

    if (theme === 'light') {
        return false;
    }

    return window.matchMedia('(prefers-color-scheme: dark)').matches;
}
