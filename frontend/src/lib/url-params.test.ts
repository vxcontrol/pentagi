import { describe, expect, it } from 'vitest';

import { mergeHrefWithSearchParams } from './url-params';

describe('mergeHrefWithSearchParams', () => {
    it('appends incoming params to a path-only href', () => {
        const result = mergeHrefWithSearchParams('/flows/1', new URLSearchParams('q=foo'));
        expect(result).toBe('/flows/1?q=foo');
    });

    it('lets the href keep its own query params when keys collide', () => {
        const result = mergeHrefWithSearchParams('/flows/1?bar=keep', new URLSearchParams('bar=overridden&q=x'));
        expect(result).toBe('/flows/1?bar=keep&q=x');
    });

    it('preserves the hash fragment', () => {
        const result = mergeHrefWithSearchParams('/flows/1#tab=logs', new URLSearchParams('q=foo'));
        expect(result).toBe('/flows/1?q=foo#tab=logs');
    });

    it('returns the href unchanged when nothing is incoming', () => {
        expect(mergeHrefWithSearchParams('/flows/1', new URLSearchParams())).toBe('/flows/1');
        expect(mergeHrefWithSearchParams('/flows/1?bar=1', new URLSearchParams())).toBe('/flows/1?bar=1');
    });

    it('accepts an entries iterable as well as URLSearchParams', () => {
        const entries: [string, string][] = [
            ['q', 'foo'],
            ['x', '1'],
        ];

        const result = mergeHrefWithSearchParams('/flows/1', entries);
        expect(result).toBe('/flows/1?q=foo&x=1');
    });

    it('round-trips through URLSearchParams encoding (spaces ↔ +, special chars escaped)', () => {
        // URLSearchParams decodes `+` as space on read, then re-encodes spaces
        // as `+` on toString — so a literal `+` in the input is interpreted
        // as a space. We document the actual behaviour here so callers know
        // what to expect.
        const result = mergeHrefWithSearchParams('/flows/1', new URLSearchParams('q=a b&keep=%23anchor'));
        expect(result).toBe('/flows/1?q=a+b&keep=%23anchor');
    });
});
