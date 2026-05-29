import { z } from 'zod';

import { getStorageItem, setStorageItem } from '@/lib/local-storage';

import type { FileManagerSortState } from './file-manager-types';

/**
 * Storage shape for the FileManager active sort. Mirrors the runtime
 * `FileManagerSortState` (a `null` payload means "no sort, insertion order
 * preserved"). The schema is the source of truth for the on-disk format —
 * if the runtime type changes, this schema must change too so existing
 * payloads either upgrade cleanly or get rejected (and discarded by the
 * loader's safeParse fallback).
 */
const fileManagerSortingSchema = z
    .object({
        column: z.enum(['modified', 'name', 'size']),
        direction: z.enum(['asc', 'desc']),
    })
    .nullable();

/**
 * Loader for the FileManager sort descriptor. Both "nothing stored" and
 * "stored as null" surface as `null` (no-sort) — the two cases are
 * indistinguishable from the consumer's point of view.
 */
export const loadFileManagerSorting = (key: string): FileManagerSortState =>
    getStorageItem(key, fileManagerSortingSchema);

export const saveFileManagerSorting = (key: string, sorting: FileManagerSortState): void =>
    setStorageItem(key, sorting);
