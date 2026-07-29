import { describe, expect, it } from 'vitest';

import { parseColorAlpha, resolveGroundLayers } from './contrast.ts';

// Every colour below is a verbatim `getComputedStyle(...).backgroundColor` reading from the
// Chromium the e2e suite drives: Chrome leaves oklch/lab/color() in their own function.
describe('parseColorAlpha', () => {
    it('reads the alpha of the colour functions Chrome keeps unconverted', () => {
        expect(parseColorAlpha('oklch(0.7 0.1 250 / 0.3)')).toBeCloseTo(0.3);
        expect(parseColorAlpha('oklch(0.7 0.1 250)')).toBe(1);
        expect(parseColorAlpha('oklch(0.216 0.006 56.043 / 0.3)')).toBeCloseTo(0.3);
        expect(parseColorAlpha('color(srgb 1 0 0 / 0.4)')).toBeCloseTo(0.4);
        expect(parseColorAlpha('color(srgb 1 0 0)')).toBe(1);
        expect(parseColorAlpha('lab(50 40 59.5 / 0.5)')).toBeCloseTo(0.5);
    });

    it('reads the alpha of the legacy sRGB forms', () => {
        expect(parseColorAlpha('rgba(255, 0, 0, 0.5)')).toBeCloseTo(0.5);
        expect(parseColorAlpha('rgba(64, 149, 191, 0.25)')).toBeCloseTo(0.25);
        expect(parseColorAlpha('rgba(0, 0, 0, 0)')).toBe(0);
        expect(parseColorAlpha('transparent')).toBe(0);
    });

    it('treats an opaque rgb() as opaque instead of reading its blue channel as the alpha', () => {
        expect(parseColorAlpha('rgb(255, 0, 0)')).toBe(1);
        expect(parseColorAlpha('rgb(24, 24, 27)')).toBe(1);
        expect(parseColorAlpha('rgb(0, 0, 0)')).toBe(1);
    });

    it('reads alpha out of hex and percentage forms', () => {
        expect(parseColorAlpha('#ff0000')).toBe(1);
        expect(parseColorAlpha('#ff000080')).toBeCloseTo(0.5, 2);
        expect(parseColorAlpha('#f008')).toBeCloseTo(0.53, 2);
        expect(parseColorAlpha('rgb(255 0 0 / 30%)')).toBeCloseTo(0.3);
    });
});

describe('resolveGroundLayers', () => {
    it('stops on an opaque rgb() ancestor instead of skipping it as invisible', () => {
        expect(resolveGroundLayers(['rgba(0, 0, 0, 0)', 'rgb(255, 0, 0)', 'rgb(0, 0, 255)'])).toEqual([
            'rgb(255, 0, 0)',
        ]);
    });

    it('keeps walking under a translucent oklch ancestor instead of grounding on it', () => {
        expect(resolveGroundLayers(['oklch(0.7 0.1 250 / 0.3)', 'oklch(0.985 0 0)', 'oklch(0.145 0 0)'])).toEqual([
            'oklch(0.985 0 0)',
            'oklch(0.7 0.1 250 / 0.3)',
        ]);
    });

    it('skips a fully transparent ancestor and keeps the layers outermost first', () => {
        expect(
            resolveGroundLayers(['rgba(0, 0, 0, 0)', 'color(srgb 1 0 0 / 0.4)', 'rgba(0, 0, 0, 0)', '#101014']),
        ).toEqual(['#101014', 'color(srgb 1 0 0 / 0.4)']);
    });
});
