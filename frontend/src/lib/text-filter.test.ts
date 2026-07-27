import { describe, expect, it } from 'vitest';

import { createTextMatcher, matchesTextFilter, normalizeForFilter } from './text-filter';

describe('normalizeForFilter', () => {
    it('lowercases ASCII text', () => {
        expect(normalizeForFilter('FooBar')).toBe('foobar');
    });

    it('strips combining diacritics so accented characters fold to plain', () => {
        expect(normalizeForFilter('café')).toBe('cafe');
        expect(normalizeForFilter('résumé')).toBe('resume');
        expect(normalizeForFilter('naïve')).toBe('naive');
    });

    it('is idempotent', () => {
        const once = normalizeForFilter('Café');
        expect(normalizeForFilter(once)).toBe(once);
    });
});

describe('createTextMatcher', () => {
    it('returns a matcher that accepts everything for an empty query', () => {
        const matcher = createTextMatcher('');

        expect(matcher('anything')).toBe(true);
        expect(matcher('')).toBe(true);
        expect(matcher(null)).toBe(true);
        expect(matcher(undefined)).toBe(true);
    });

    it('matches case-insensitively', () => {
        const matcher = createTextMatcher('Foo');

        expect(matcher('foo')).toBe(true);
        expect(matcher('FOO')).toBe(true);
        expect(matcher('hello FOO bar')).toBe(true);
        expect(matcher('bar')).toBe(false);
    });

    it('matches across diacritic-folded forms', () => {
        const matcher = createTextMatcher('cafe');

        expect(matcher('café')).toBe(true);
        expect(matcher('Café au lait')).toBe(true);
        expect(matcher('CAFÉ')).toBe(true);
    });

    it('folds diacritics in the query too', () => {
        const matcher = createTextMatcher('Café');

        expect(matcher('cafe')).toBe(true);
        expect(matcher('CAFE')).toBe(true);
    });

    it('returns false for null and undefined text when query is non-empty', () => {
        const matcher = createTextMatcher('foo');

        expect(matcher(null)).toBe(false);
        expect(matcher(undefined)).toBe(false);
    });

    it('does substring matching, not whole-word or prefix matching', () => {
        const matcher = createTextMatcher('ell');

        expect(matcher('hello')).toBe(true);
        expect(matcher('shell shocked')).toBe(true);
        expect(matcher('apricot')).toBe(false);
    });

    it('preserves whitespace exactly — does not trim the query', () => {
        const matcher = createTextMatcher(' foo ');

        expect(matcher(' foo ')).toBe(true);
        expect(matcher('foo')).toBe(false);
    });

    it('returns the same matcher behaviour across many invocations (no internal state leak)', () => {
        const matcher = createTextMatcher('abc');

        for (let i = 0; i < 10; i += 1) {
            expect(matcher('xxabcyy')).toBe(true);
            expect(matcher('xyz')).toBe(false);
        }
    });
});

describe('matchesTextFilter', () => {
    it('delegates to createTextMatcher and produces the same answers', () => {
        expect(matchesTextFilter('hello', 'ell')).toBe(true);
        expect(matchesTextFilter('hello', 'world')).toBe(false);
        expect(matchesTextFilter(null, 'foo')).toBe(false);
        expect(matchesTextFilter(null, '')).toBe(true);
    });

    it('folds diacritics symmetrically on both sides', () => {
        expect(matchesTextFilter('résumé.pdf', 'resume')).toBe(true);
        expect(matchesTextFilter('resume.pdf', 'résumé')).toBe(true);
    });
});
