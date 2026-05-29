import { z } from 'zod';

import { getStorageItem, removeStorageItem, setStorageItem } from './local-storage';
import { STORAGE_KEY_SEPARATOR } from './storage-keys';

const sortingSchema = z.array(z.object({ desc: z.boolean(), id: z.string() }));
const visibilitySchema = z.record(z.string(), z.boolean());

/**
 * Unified table state schema. All per-page table preferences live under one
 * JSON key (`table_4_<path>`) instead of fanning out into `column_4_/`,
 * `sorting_4_/`, `filter_4_/` etc. Every field is optional so partial state
 * (e.g. only a filter, no custom column visibility) doesn't force defaults
 * into storage.
 *
 * `pageSize` is the user's chosen rows-per-page; the current `page` index
 * stays in the URL where it can be bookmarked.
 */
const tableStateSchema = z.object({
    columnVisibility: visibilitySchema.optional(),
    filter: z.string().optional(),
    pageSize: z.number().int().positive().optional(),
    searchColumns: z.array(z.string()).optional(),
    sorting: sortingSchema.optional(),
});

export type TableState = z.infer<typeof tableStateSchema>;

const LEGACY_TYPES = ['column', 'sorting', 'filter', 'page'] as const;
const legacyPageSchema = z.object({ page: z.number(), pageSize: z.number() });

/** Paths that historically wrote legacy keys with or without a trailing `/`. */
const legacyStoragePathVariants = (path: string): string[] => {
    if (path === '/') {
        return ['/'];
    }

    const withoutTrailing = path.replace(/\/+$/, '');
    const canonical = withoutTrailing === '' ? '/' : withoutTrailing;
    const withSlash = `${canonical}/`;

    return canonical === path ? [canonical, withSlash] : [path, canonical];
};

const legacyKeysForPath = (p: string): string[] => LEGACY_TYPES.map((type) => `${type}${STORAGE_KEY_SEPARATOR}${p}`);

/** Read the unified state. Returns `{}` when missing or invalid. */
export const loadTableState = (key: string): TableState => getStorageItem(key, tableStateSchema) ?? {};

/**
 * Merge `patch` into the stored state. `undefined` values clear the
 * corresponding field. When the result is empty, the storage key is removed
 * so an unsubscribed page leaves no residue (callers can rely on
 * `loadTableState() === {}` ≡ "no preferences set").
 */
export const updateTableState = (key: string, patch: Partial<TableState>): TableState => {
    const current = loadTableState(key);
    const merged: TableState = { ...current };

    for (const [field, value] of Object.entries(patch) as [keyof TableState, TableState[keyof TableState]][]) {
        if (value === undefined) {
            delete merged[field];
            continue;
        }

        // Empty string / empty array / empty object all collapse to "no value"
        // so storage holds the same canonical shape as `loadTableState()` of
        // a fresh user — keeps comparison ergonomic and prevents key churn.
        const isEmptyString = typeof value === 'string' && value.length === 0;
        const isEmptyArray = Array.isArray(value) && value.length === 0;
        const isEmptyRecord =
            value !== null && typeof value === 'object' && !Array.isArray(value) && Object.keys(value).length === 0;

        if (isEmptyString || isEmptyArray || isEmptyRecord) {
            delete merged[field];
            continue;
        }

        (merged as Record<string, unknown>)[field] = value;
    }

    if (Object.keys(merged).length === 0) {
        removeStorageItem(key);

        return merged;
    }

    setStorageItem(key, merged);

    return merged;
};

/**
 * One-shot migration from the pre-unified storage layout. Reads the four
 * legacy keys (`column_4_`, `sorting_4_`, `filter_4_`, `page_4_`), folds
 * them into the new unified key, and deletes the legacy slots. Returns the
 * resulting state.
 *
 * Idempotent: once the legacy keys are gone, subsequent runs short-circuit
 * to a plain `loadTableState`. Safe to call on every page mount.
 *
 * Note on `page`: the legacy schema stored `{ page, pageSize }`. We carry
 * over `pageSize` only — the current page index now lives in `?page=` so
 * the user can share a paginated view via URL.
 */
export const migrateLegacyTableState = (path: string, unifiedKey: string): TableState => {
    const variants = legacyStoragePathVariants(path);
    const allLegacyKeys = variants.flatMap(legacyKeysForPath);
    const anyLegacyPresent = allLegacyKeys.some((key) => localStorage.getItem(key) !== null);

    if (!anyLegacyPresent) {
        return loadTableState(unifiedKey);
    }

    const pickFirst = <T>(reader: (scopedPath: string) => null | T): null | T => {
        for (const scopedPath of variants) {
            const value = reader(scopedPath);

            if (value !== null) {
                return value;
            }
        }

        return null;
    };

    const legacySorting = pickFirst((p) => getStorageItem(`sorting${STORAGE_KEY_SEPARATOR}${p}`, sortingSchema));
    const legacyVisibility = pickFirst((p) => getStorageItem(`column${STORAGE_KEY_SEPARATOR}${p}`, visibilitySchema));
    const legacyFilter = pickFirst((p) => getStorageItem(`filter${STORAGE_KEY_SEPARATOR}${p}`, z.string()));
    const legacyPage = pickFirst((p) => getStorageItem(`page${STORAGE_KEY_SEPARATOR}${p}`, legacyPageSchema));

    const merged = updateTableState(unifiedKey, {
        columnVisibility: legacyVisibility ?? undefined,
        filter: legacyFilter ?? undefined,
        pageSize: legacyPage?.pageSize,
        sorting: legacySorting ?? undefined,
    });

    for (const key of allLegacyKeys) {
        removeStorageItem(key);
    }

    return merged;
};
