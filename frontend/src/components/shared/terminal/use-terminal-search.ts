import type { SearchAddon } from '@xterm/addon-search';

import { useCallback, useEffect } from 'react';

import { Log } from '@/lib/log';

import { getSearchDecorations, isDarkMode } from './terminal-config';

interface UseTerminalSearchResult {
    findNext: () => void;
    findPrevious: () => void;
}

export function useTerminalSearch(
    searchAddon: null | SearchAddon,
    isReady: boolean,
    searchValue: string | undefined,
    theme: 'dark' | 'light' | 'system',
): UseTerminalSearchResult {
    useEffect(() => {
        if (!searchAddon || !isReady) {
            return;
        }

        try {
            const trimmed = searchValue?.trim();

            if (trimmed) {
                searchAddon.findNext(trimmed, buildSearchOptions(isDarkMode(theme)));
            } else {
                searchAddon.clearDecorations();
            }
        } catch (error: unknown) {
            Log.error('Terminal search failed:', error);
        }
    }, [searchAddon, isReady, searchValue, theme]);

    const findNext = useCallback(() => {
        const trimmed = searchValue?.trim();

        if (!searchAddon || !trimmed) {
            return;
        }

        try {
            searchAddon.findNext(trimmed, buildSearchOptions(isDarkMode(theme)));
        } catch (error: unknown) {
            Log.error('Terminal findNext failed:', error);
        }
    }, [searchAddon, searchValue, theme]);

    const findPrevious = useCallback(() => {
        const trimmed = searchValue?.trim();

        if (!searchAddon || !trimmed) {
            return;
        }

        try {
            searchAddon.findPrevious(trimmed, buildSearchOptions(isDarkMode(theme)));
        } catch (error: unknown) {
            Log.error('Terminal findPrevious failed:', error);
        }
    }, [searchAddon, searchValue, theme]);

    return { findNext, findPrevious };
}

function buildSearchOptions(isDark: boolean) {
    return {
        caseSensitive: false,
        decorations: getSearchDecorations(isDark),
        regex: false,
        wholeWord: false,
    } as const;
}
