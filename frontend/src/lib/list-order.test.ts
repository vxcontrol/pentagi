import { describe, expect, it } from 'vitest';

import { orderOf, placeMissedRows } from './list-order';

describe('orderOf', () => {
    it('reads the declared order for a mirrored collection', () => {
        expect(orderOf('assistants')).toBe('newestFirst');
        expect(orderOf('assistantLogs')).toBe('chronological');
    });

    it('treats an unknown field as unmirrored', () => {
        expect(orderOf('somethingElse')).toBe('unmirrored');
    });
});

describe('placeMissedRows', () => {
    it('puts rows the snapshot missed in front of a newest-first list', () => {
        expect(placeMissedRows('assistants', ['b', 'c'], ['a'])).toEqual(['a', 'b', 'c']);
    });

    it('appends them to a chronological list', () => {
        expect(placeMissedRows('assistantLogs', ['a', 'b'], ['c'])).toEqual(['a', 'b', 'c']);
    });

    it('appends them for an unmirrored field too', () => {
        expect(placeMissedRows('resources', ['a'], ['b'])).toEqual(['a', 'b']);
    });
});
