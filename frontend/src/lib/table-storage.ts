import type { SortingState, VisibilityState } from '@tanstack/react-table';
import { z } from 'zod';

const sortingSchema = z.array(z.object({ desc: z.boolean(), id: z.string() }));

const visibilitySchema = z.record(z.string(), z.boolean());

const pageStateSchema = z.object({ page: z.number(), pageSize: z.number() });

export type StoredPageState = z.infer<typeof pageStateSchema>;

function loadFromStorage<T>(key: string, schema: z.ZodType<T>): T | null {
    try {
        const raw = localStorage.getItem(key);

        if (raw === null) {
            return null;
        }

        const result = schema.safeParse(JSON.parse(raw));

        return result.success ? result.data : null;
    } catch {
        return null;
    }
}

function saveToStorage(key: string, value: unknown): void {
    try {
        localStorage.setItem(key, JSON.stringify(value));
    } catch {
        /* localStorage may be unavailable */
    }
}

export const loadSorting = (key: string): SortingState | null => loadFromStorage(key, sortingSchema);

export const loadColumnVisibility = (key: string): VisibilityState | null =>
    loadFromStorage(key, visibilitySchema);

export const loadPageState = (key: string): StoredPageState | null =>
    loadFromStorage(key, pageStateSchema);

export const saveSorting = (key: string, sorting: SortingState): void =>
    saveToStorage(key, sorting);

export const saveColumnVisibility = (key: string, visibility: VisibilityState): void =>
    saveToStorage(key, visibility);

export const savePageState = (key: string, state: StoredPageState): void =>
    saveToStorage(key, state);
