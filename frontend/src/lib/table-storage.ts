import type { SortingState, VisibilityState } from '@tanstack/react-table';

import { z } from 'zod';

import { getStorageItem, setStorageItem } from './local-storage';

const sortingSchema = z.array(z.object({ desc: z.boolean(), id: z.string() }));

const visibilitySchema = z.record(z.string(), z.boolean());

const pageStateSchema = z.object({ page: z.number(), pageSize: z.number() });

export type StoredPageState = z.infer<typeof pageStateSchema>;

export const loadSorting = (key: string): null | SortingState => getStorageItem(key, sortingSchema);

export const loadColumnVisibility = (key: string): null | VisibilityState => getStorageItem(key, visibilitySchema);

export const loadPageState = (key: string): null | StoredPageState => getStorageItem(key, pageStateSchema);

export const saveSorting = (key: string, sorting: SortingState): void => setStorageItem(key, sorting);

export const saveColumnVisibility = (key: string, visibility: VisibilityState): void =>
    setStorageItem(key, visibility);

export const savePageState = (key: string, state: StoredPageState): void => setStorageItem(key, state);
