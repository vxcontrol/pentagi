import type { VisibilityState } from '@tanstack/react-table';

import { useCallback, useEffect, useMemo, useState } from 'react';

import { getColumnStorageKey } from '@/lib/storage-keys';

export interface ColumnPriority {
    alwaysVisible?: boolean;
    id: string;
    priority: number;
}

interface UseAdaptiveColumnVisibilityOptions {
    breakpoints?: { hiddenPriorities: number[]; width: number }[];
    columns: ColumnPriority[];
    tableKey: string;
}

const DEFAULT_BREAKPOINTS = [
    { hiddenPriorities: [], width: 1400 },
    { hiddenPriorities: [5], width: 1200 },
    { hiddenPriorities: [4, 5], width: 1000 },
    { hiddenPriorities: [3, 4, 5], width: 800 },
    { hiddenPriorities: [2, 3, 4, 5], width: 600 },
    { hiddenPriorities: [1, 2, 3, 4, 5], width: 0 },
];

function loadUserPreferences(key: string): Record<string, boolean> {
    try {
        const stored = localStorage.getItem(key);

        return stored ? JSON.parse(stored) : {};
    } catch {
        return {};
    }
}

export const useAdaptiveColumnVisibility = ({
    breakpoints = DEFAULT_BREAKPOINTS,
    columns,
    tableKey,
}: UseAdaptiveColumnVisibilityOptions) => {
    const [windowWidth, setWindowWidth] = useState(typeof window !== 'undefined' ? window.innerWidth : 1400);

    const localStorageKey = useMemo(() => getColumnStorageKey(tableKey), [tableKey]);

    const [userPreferences, setUserPreferences] = useState<Record<string, boolean>>(() =>
        loadUserPreferences(localStorageKey),
    );

    const saveUserPreferences = useCallback(
        (preferences: Record<string, boolean>) => {
            try {
                localStorage.setItem(localStorageKey, JSON.stringify(preferences));
                setUserPreferences(preferences);
            } catch {
                /* localStorage may be unavailable */
            }
        },
        [localStorageKey],
    );

    useEffect(() => {
        const handleResize = () => setWindowWidth(window.innerWidth);

        window.addEventListener('resize', handleResize);

        return () => window.removeEventListener('resize', handleResize);
    }, []);

    const columnVisibility = useMemo((): VisibilityState => {
        const activeBreakpoint = breakpoints.find((breakpoint) => windowWidth >= breakpoint.width) ??
            breakpoints.at(-1) ?? { hiddenPriorities: [], width: 0 };

        return Object.fromEntries(
            columns.map((column) => {
                if (column.alwaysVisible) {
                    return [column.id, true];
                }

                const shouldHideByWidth = activeBreakpoint.hiddenPriorities.includes(column.priority);
                const userPreference = userPreferences[column.id];

                const isVisible =
                    userPreference !== undefined
                        ? !shouldHideByWidth && userPreference
                        : !shouldHideByWidth;

                return [column.id, isVisible];
            }),
        );
    }, [windowWidth, userPreferences, columns, breakpoints]);

    const updateColumnVisibility = useCallback(
        (columnId: string, visible: boolean) => {
            saveUserPreferences({
                ...userPreferences,
                [columnId]: visible,
            });
        },
        [userPreferences, saveUserPreferences],
    );

    return {
        columnVisibility,
        updateColumnVisibility,
        userPreferences,
    };
};
