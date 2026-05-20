import { describe, expect, it } from 'vitest';

import { computeNavigation } from './use-navigation';

interface Row {
    id: string;
    title: string;
}

const getId = (row: Row) => row.id;
const getTitle = (row: Row) => row.title;

const ROWS: readonly Row[] = [
    { id: 'a', title: 'Alpha' },
    { id: 'b', title: 'Bravo' },
    { id: 'c', title: 'Charlie' },
    { id: 'd', title: 'Delta' },
] as const;

describe('computeNavigation', () => {
    it('returns prev/next neighbours for a middle item', () => {
        const result = computeNavigation({
            currentId: 'c',
            getId,
            items: ROWS,
        });

        expect(result.currentIndex).toBe(2);
        expect(result.prevId).toBe('b');
        expect(result.nextId).toBe('d');
        expect(result.total).toBe(4);
    });

    it('returns null prev for the first item', () => {
        const result = computeNavigation({
            currentId: 'a',
            getId,
            items: ROWS,
        });

        expect(result.currentIndex).toBe(0);
        expect(result.prevId).toBeNull();
        expect(result.nextId).toBe('b');
    });

    it('returns null next for the last item', () => {
        const result = computeNavigation({
            currentId: 'd',
            getId,
            items: ROWS,
        });

        expect(result.currentIndex).toBe(3);
        expect(result.prevId).toBe('c');
        expect(result.nextId).toBeNull();
    });

    it('reports currentIndex=-1 when the current item is missing from the filtered subset', () => {
        const result = computeNavigation({
            currentId: 'zzz',
            getId,
            items: ROWS,
        });

        expect(result.currentIndex).toBe(-1);
        expect(result.currentItem).toBeNull();
        expect(result.prevId).toBeNull();
        expect(result.nextId).toBeNull();
        expect(result.total).toBe(4);
    });

    it('reports currentIndex=-1 when currentId is null or undefined', () => {
        const nullResult = computeNavigation({
            currentId: null,
            getId,
            items: ROWS,
        });

        expect(nullResult.currentIndex).toBe(-1);
        expect(nullResult.prevId).toBeNull();
        expect(nullResult.nextId).toBeNull();

        const undefinedResult = computeNavigation({
            currentId: undefined,
            getId,
            items: ROWS,
        });

        expect(undefinedResult.currentIndex).toBe(-1);
    });

    it('honours `query` when narrowing the subset', () => {
        const result = computeNavigation({
            currentId: 'c',
            getId,
            getSearchableText: getTitle,
            items: ROWS,
            // Substring "c" matches "Charlie" (case-insensitive).
            query: 'c',
        });

        expect(result.filteredItems.map(getId)).toEqual(['c']);
        expect(result.currentIndex).toBe(0);
        expect(result.prevId).toBeNull();
        expect(result.nextId).toBeNull();
        expect(result.total).toBe(1);
    });

    it('drops the current item from the result when it does not match the query', () => {
        const result = computeNavigation({
            currentId: 'a',
            getId,
            getSearchableText: getTitle,
            items: ROWS,
            query: 'Bravo',
        });

        expect(result.filteredItems.map(getId)).toEqual(['b']);
        expect(result.currentIndex).toBe(-1);
        expect(result.prevId).toBeNull();
        expect(result.nextId).toBeNull();
    });

    it('treats an empty/undefined query as "no filter" even when getSearchableText is provided', () => {
        const empty = computeNavigation({
            currentId: 'c',
            getId,
            getSearchableText: getTitle,
            items: ROWS,
            query: '',
        });

        expect(empty.filteredItems.map(getId)).toEqual(['a', 'b', 'c', 'd']);

        const missing = computeNavigation({
            currentId: 'c',
            getId,
            getSearchableText: getTitle,
            items: ROWS,
        });

        expect(missing.filteredItems.map(getId)).toEqual(['a', 'b', 'c', 'd']);
    });

    it('skips filtering silently when query is non-empty but getSearchableText is missing', () => {
        // Documents the boundary: without a haystack accessor we cannot evaluate
        // the query against rows, so we degrade to "no filter" instead of
        // throwing — keeps the hook usable while a caller forgets to wire one up.
        const result = computeNavigation({
            currentId: 'c',
            getId,
            items: ROWS,
            query: 'pha',
        });

        expect(result.filteredItems.map(getId)).toEqual(['a', 'b', 'c', 'd']);
    });

    it('preserves input order when no sortFn is provided', () => {
        const reversed = [...ROWS].reverse();
        const result = computeNavigation({
            currentId: 'c',
            getId,
            items: reversed,
        });

        expect(result.filteredItems.map(getId)).toEqual(['d', 'c', 'b', 'a']);
        expect(result.currentIndex).toBe(1);
        expect(result.prevId).toBe('d');
        expect(result.nextId).toBe('b');
    });

    it('applies sortFn to the filtered subset', () => {
        const result = computeNavigation({
            currentId: 'b',
            getId,
            items: ROWS,
            sortFn: (a, b) => b.title.localeCompare(a.title),
        });

        expect(result.filteredItems.map(getId)).toEqual(['d', 'c', 'b', 'a']);
        expect(result.currentIndex).toBe(2);
        expect(result.prevId).toBe('c');
        expect(result.nextId).toBe('a');
    });

    it('handles an empty items array', () => {
        const result = computeNavigation({
            currentId: 'anything',
            getId,
            items: [],
        });

        expect(result.filteredItems).toEqual([]);
        expect(result.total).toBe(0);
        expect(result.currentIndex).toBe(-1);
        expect(result.prevId).toBeNull();
        expect(result.nextId).toBeNull();
    });

    it('returns the current item when present', () => {
        const result = computeNavigation({
            currentId: 'c',
            getId,
            items: ROWS,
        });

        expect(result.currentItem).toEqual({ id: 'c', title: 'Charlie' });
    });

    it('does not mutate the input array even when sorting', () => {
        const before = ROWS.map(getId);
        computeNavigation({
            currentId: 'a',
            getId,
            items: ROWS,
            sortFn: (a, b) => b.title.localeCompare(a.title),
        });

        expect(ROWS.map(getId)).toEqual(before);
    });

    it('handles "item deleted, currentId still points to it"', () => {
        // The user was viewing item `b`, then `b` was deleted (removed from
        // `items`) — currentId stays `b` until the route updates. The hook
        // must surface this as currentIndex=-1 / no neighbours so the UI
        // disables Prev/Next rather than jumping to an unrelated row.
        const trimmed = ROWS.filter((row) => row.id !== 'b');
        const result = computeNavigation({
            currentId: 'b',
            getId,
            items: trimmed,
        });

        expect(result.filteredItems.map(getId)).toEqual(['a', 'c', 'd']);
        expect(result.currentIndex).toBe(-1);
        expect(result.prevId).toBeNull();
        expect(result.nextId).toBeNull();
        expect(result.total).toBe(3);
    });

    it('folds diacritics + case the same way the list filter does', () => {
        const accented: readonly Row[] = [
            { id: '1', title: 'Café' },
            { id: '2', title: 'naïve' },
            { id: '3', title: 'resume' },
        ];

        const result = computeNavigation({
            currentId: '1',
            getId,
            getSearchableText: getTitle,
            items: accented,
            query: 'cafe',
        });

        expect(result.filteredItems.map(getId)).toEqual(['1']);
        expect(result.currentIndex).toBe(0);
    });

    it('treats an empty array query as "no filter"', () => {
        const result = computeNavigation({
            currentId: 'c',
            getId,
            getSearchableText: getTitle,
            items: ROWS,
            query: [],
        });

        expect(result.filteredItems.map(getId)).toEqual(['a', 'b', 'c', 'd']);
        expect(result.currentIndex).toBe(2);
    });

    it('ANDs every non-empty term when `query` is an array', () => {
        // Both terms must match the haystack — substring "a" matches Alpha,
        // Bravo, Charlie; substring "ph" then narrows to just Alpha.
        const result = computeNavigation({
            currentId: 'a',
            getId,
            getSearchableText: getTitle,
            items: ROWS,
            query: ['a', 'ph'],
        });

        expect(result.filteredItems.map(getId)).toEqual(['a']);
        expect(result.currentIndex).toBe(0);
        expect(result.prevId).toBeNull();
        expect(result.nextId).toBeNull();
    });

    it('drops empty strings from an array query without throwing', () => {
        // Callers can pass `[urlFilter, localSearch]` straight through; either
        // one being empty must not turn the whole filter off — the other term
        // is still expected to narrow.
        const result = computeNavigation({
            currentId: 'b',
            getId,
            getSearchableText: getTitle,
            items: ROWS,
            query: ['', 'bravo', ''],
        });

        expect(result.filteredItems.map(getId)).toEqual(['b']);
    });

    it('falls back to "no filter" when every array term is empty', () => {
        const result = computeNavigation({
            currentId: 'a',
            getId,
            getSearchableText: getTitle,
            items: ROWS,
            query: ['', '', ''],
        });

        expect(result.filteredItems.map(getId)).toEqual(['a', 'b', 'c', 'd']);
    });

    it('matches numeric item ids with string currentId', () => {
        type NumRow = { id: number; title: string };
        const rows: readonly NumRow[] = [
            { id: 10, title: 'A' },
            { id: 20, title: 'B' },
            { id: 30, title: 'C' },
        ];
        const getNumericId = (row: NumRow) => String(row.id);

        const result = computeNavigation({
            currentId: '20',
            getId: getNumericId,
            items: rows,
        });

        expect(result.currentIndex).toBe(1);
        expect(result.prevId).toBe('10');
        expect(result.nextId).toBe('30');
    });
});
