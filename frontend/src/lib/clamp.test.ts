import { describe, expect, it } from 'vitest';

import { clamp } from './clamp';

describe('clamp', () => {
    it('keeps a value inside the range', () => {
        expect(clamp(0, 5, 10)).toBe(5);
    });

    it('returns the bounds themselves', () => {
        expect(clamp(0, -1, 10)).toBe(0);
        expect(clamp(0, 11, 10)).toBe(10);
        expect(clamp(0, 0, 10)).toBe(0);
        expect(clamp(0, 10, 10)).toBe(10);
    });

    it('answers the lower bound when the range is inverted', () => {
        expect(clamp(10, 5, 0)).toBe(10);
    });
});
