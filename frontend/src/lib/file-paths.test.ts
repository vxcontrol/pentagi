import { describe, expect, it } from 'vitest';

import { getBaseName, getParentDir } from './file-paths';

describe('getParentDir', () => {
    it.each([
        ['reports/notes.txt', 'reports'],
        ['a/b/c.txt', 'a/b'],
        ['notes.txt', ''],
        ['/notes.txt', ''],
        ['', ''],
    ])('takes %j back to %j', (path, expected) => {
        expect(getParentDir(path)).toBe(expected);
    });
});

describe('getBaseName', () => {
    it.each([
        ['reports/notes.txt', 'notes.txt'],
        ['notes.txt', 'notes.txt'],
        ['reports/', ''],
        ['', ''],
    ])('takes %j down to %j', (path, expected) => {
        expect(getBaseName(path)).toBe(expected);
    });
});
