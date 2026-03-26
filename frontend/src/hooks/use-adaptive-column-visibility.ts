import type { VisibilityState } from '@tanstack/react-table';

import { useEffect, useMemo, useState } from 'react';

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

export const useAdaptiveColumnVisibility = ({
    breakpoints = DEFAULT_BREAKPOINTS,
    columns,
    tableKey,
}: UseAdaptiveColumnVisibilityOptions) => {
    const [windowWidth, setWindowWidth] = useState(typeof window !== 'undefined' ? window.innerWidth : 1400);

    const localStorageKey = `table-column-visibility-${tableKey}`;

    const getUserPreferences = (): Record<string, boolean> => {
        try {
            const stored = localStorage.getItem(localStorageKey);

            return stored ? JSON.parse(stored) : {};
        } catch {
            return {};
        }
    };

    const [userPreferences, setUserPreferences] = useState<Record<string, boolean>>(getUserPreferences);

    const saveUserPreferences = (preferences: Record<string, boolean>) => {
        try {
            localStorage.setItem(localStorageKey, JSON.stringify(preferences));
            setUserPreferences(preferences);
        } catch (error) {
            console.error('Failed to save column visibility preferences:', error);
        }
    };

    useEffect(() => {
        const handleResize = () => {
            setWindowWidth(window.innerWidth);
        };

        window.addEventListener('resize', handleResize);

        return () => window.removeEventListener('resize', handleResize);
    }, []);

    const columnVisibility = useMemo((): VisibilityState => {
        const activeBreakpoint = breakpoints.find((bp) => windowWidth >= bp.width) ||
            breakpoints[breakpoints.length - 1] || { hiddenPriorities: [], width: 0 };

        const visibility: VisibilityState = {};

        columns.forEach((column) => {
            if (column.alwaysVisible) {
                visibility[column.id] = true;

                return;
            }

            const shouldHideByWidth = activeBreakpoint.hiddenPriorities.includes(column.priority);
            const userPreference = userPreferences[column.id];

            if (userPreference !== undefined) {
                visibility[column.id] = shouldHideByWidth ? false : userPreference;
            } else {
                visibility[column.id] = !shouldHideByWidth;
            }
        });

        return visibility;
    }, [windowWidth, userPreferences, columns, breakpoints]);

    const updateColumnVisibility = (columnId: string, visible: boolean) => {
        const newPreferences = {
            ...userPreferences,
            [columnId]: visible,
        };
        saveUserPreferences(newPreferences);
    };

    return {
        columnVisibility,
        updateColumnVisibility,
        userPreferences,
    };
};
