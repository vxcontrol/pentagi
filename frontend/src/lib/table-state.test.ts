import { afterEach, beforeEach, describe, expect, it } from 'vitest';

import { loadTableState, migrateLegacyTableState, updateTableState } from './table-state';

const PATH = '/flows';
const UNIFIED_KEY = `table_4_${PATH}`;

const LEGACY_COLUMN_KEY = `column_4_${PATH}`;
const LEGACY_SORTING_KEY = `sorting_4_${PATH}`;
const LEGACY_FILTER_KEY = `filter_4_${PATH}`;
const LEGACY_PAGE_KEY = `page_4_${PATH}`;

beforeEach(() => {
    localStorage.clear();
});

afterEach(() => {
    localStorage.clear();
});

describe('loadTableState', () => {
    it('returns an empty object when the key is missing', () => {
        expect(loadTableState(UNIFIED_KEY)).toEqual({});
    });

    it('returns an empty object when the payload is invalid JSON', () => {
        localStorage.setItem(UNIFIED_KEY, 'not json');
        expect(loadTableState(UNIFIED_KEY)).toEqual({});
    });

    it('returns an empty object when the payload fails the schema', () => {
        localStorage.setItem(UNIFIED_KEY, JSON.stringify({ filter: 42 }));
        expect(loadTableState(UNIFIED_KEY)).toEqual({});
    });

    it('round-trips a full state object', () => {
        const state = {
            columnVisibility: { name: false },
            filter: 'alpha',
            pageSize: 25,
            sorting: [{ desc: true, id: 'createdAt' }],
        };
        localStorage.setItem(UNIFIED_KEY, JSON.stringify(state));
        expect(loadTableState(UNIFIED_KEY)).toEqual(state);
    });
});

describe('updateTableState', () => {
    it('writes a partial patch on top of the existing state', () => {
        localStorage.setItem(UNIFIED_KEY, JSON.stringify({ filter: 'alpha', pageSize: 25 }));
        updateTableState(UNIFIED_KEY, { sorting: [{ desc: false, id: 'name' }] });

        const stored = JSON.parse(localStorage.getItem(UNIFIED_KEY) ?? '{}');
        expect(stored.filter).toBe('alpha');
        expect(stored.pageSize).toBe(25);
        expect(stored.sorting).toEqual([{ desc: false, id: 'name' }]);
    });

    it('clears a field via `undefined`', () => {
        localStorage.setItem(UNIFIED_KEY, JSON.stringify({ filter: 'alpha', pageSize: 25 }));
        updateTableState(UNIFIED_KEY, { filter: undefined });

        const stored = JSON.parse(localStorage.getItem(UNIFIED_KEY) ?? '{}');
        expect('filter' in stored).toBe(false);
        expect(stored.pageSize).toBe(25);
    });

    it('collapses empty strings / arrays / records to "no value"', () => {
        localStorage.setItem(
            UNIFIED_KEY,
            JSON.stringify({ columnVisibility: { name: false }, filter: 'foo', sorting: [{ desc: true, id: 'a' }] }),
        );

        updateTableState(UNIFIED_KEY, { columnVisibility: {}, filter: '', sorting: [] });

        const stored = JSON.parse(localStorage.getItem(UNIFIED_KEY) ?? '{}');
        expect(stored).toEqual({});
    });

    it('removes the storage key when the merged state is empty', () => {
        localStorage.setItem(UNIFIED_KEY, JSON.stringify({ filter: 'alpha' }));

        updateTableState(UNIFIED_KEY, { filter: undefined });

        expect(localStorage.getItem(UNIFIED_KEY)).toBeNull();
    });

    it('returns the merged state for the caller', () => {
        const next = updateTableState(UNIFIED_KEY, { pageSize: 50 });
        expect(next).toEqual({ pageSize: 50 });
    });
});

describe('migrateLegacyTableState', () => {
    it('short-circuits to loadTableState when no legacy keys exist', () => {
        localStorage.setItem(UNIFIED_KEY, JSON.stringify({ filter: 'unified' }));

        const result = migrateLegacyTableState(PATH, UNIFIED_KEY);

        expect(result).toEqual({ filter: 'unified' });
        // Storage should remain untouched.
        expect(JSON.parse(localStorage.getItem(UNIFIED_KEY) ?? '{}')).toEqual({ filter: 'unified' });
    });

    it('folds legacy sorting / column / filter / page-size into the unified slot', () => {
        localStorage.setItem(LEGACY_SORTING_KEY, JSON.stringify([{ desc: true, id: 'createdAt' }]));
        localStorage.setItem(LEGACY_COLUMN_KEY, JSON.stringify({ name: false }));
        localStorage.setItem(LEGACY_FILTER_KEY, JSON.stringify('alpha'));
        localStorage.setItem(LEGACY_PAGE_KEY, JSON.stringify({ page: 4, pageSize: 25 }));

        const result = migrateLegacyTableState(PATH, UNIFIED_KEY);

        expect(result).toEqual({
            columnVisibility: { name: false },
            filter: 'alpha',
            pageSize: 25,
            sorting: [{ desc: true, id: 'createdAt' }],
        });

        // Legacy keys must be deleted after the fold so they never resurface.
        expect(localStorage.getItem(LEGACY_SORTING_KEY)).toBeNull();
        expect(localStorage.getItem(LEGACY_COLUMN_KEY)).toBeNull();
        expect(localStorage.getItem(LEGACY_FILTER_KEY)).toBeNull();
        expect(localStorage.getItem(LEGACY_PAGE_KEY)).toBeNull();
    });

    it('drops the legacy `page` field — only `pageSize` survives the migration', () => {
        // The page index now lives in `?page=`, not storage. Carrying the
        // old `page` field forward would put it back into a place the rest
        // of the app no longer reads from.
        localStorage.setItem(LEGACY_PAGE_KEY, JSON.stringify({ page: 4, pageSize: 25 }));

        const result = migrateLegacyTableState(PATH, UNIFIED_KEY);

        expect(result).toEqual({ pageSize: 25 });
        expect((result as { page?: number }).page).toBeUndefined();
    });

    it('is idempotent — a second invocation reads the unified slot only', () => {
        localStorage.setItem(LEGACY_SORTING_KEY, JSON.stringify([{ desc: true, id: 'x' }]));

        migrateLegacyTableState(PATH, UNIFIED_KEY);
        const second = migrateLegacyTableState(PATH, UNIFIED_KEY);

        expect(second).toEqual({ sorting: [{ desc: true, id: 'x' }] });
    });

    it('skips fields whose legacy payload fails the schema', () => {
        // A legacy `column_4_` key that holds non-boolean values (e.g.
        // because the format changed mid-history). The migration should
        // silently drop it instead of polluting the unified state.
        localStorage.setItem(LEGACY_COLUMN_KEY, JSON.stringify({ name: 'visible' }));
        localStorage.setItem(LEGACY_FILTER_KEY, JSON.stringify('alpha'));

        const result = migrateLegacyTableState(PATH, UNIFIED_KEY);

        expect(result).toEqual({ filter: 'alpha' });
        // Legacy keys must still be deleted — they're invalid either way.
        expect(localStorage.getItem(LEGACY_COLUMN_KEY)).toBeNull();
        expect(localStorage.getItem(LEGACY_FILTER_KEY)).toBeNull();
    });

    it('migrates legacy keys stored under the trailing-slash pathname variant', () => {
        const pathNoSlash = '/flows';
        const pathSlash = `${pathNoSlash}/`;
        const unifiedKey = `table_4_${pathNoSlash}`;
        localStorage.setItem(`sorting_4_${pathSlash}`, JSON.stringify([{ desc: false, id: 'title' }]));

        const result = migrateLegacyTableState(pathNoSlash, unifiedKey);

        expect(result.sorting).toEqual([{ desc: false, id: 'title' }]);
        expect(localStorage.getItem(`sorting_4_${pathSlash}`)).toBeNull();
        expect(localStorage.getItem(`sorting_4_${pathNoSlash}`)).toBeNull();
    });

    it('migrates legacy keys from the canonical path when migration is invoked with a trailing slash', () => {
        const pathSlash = '/flows/';
        const pathCanonical = '/flows';
        const unifiedKey = `table_4_${pathSlash}`;
        localStorage.setItem(`filter_4_${pathCanonical}`, JSON.stringify('needle'));

        const result = migrateLegacyTableState(pathSlash, unifiedKey);

        expect(result.filter).toBe('needle');
        expect(localStorage.getItem(`filter_4_${pathCanonical}`)).toBeNull();
        expect(localStorage.getItem(`filter_4_${pathSlash}`)).toBeNull();
    });
});
