import { afterEach, beforeEach, describe, expect, it } from 'vitest';

import { defaultResourcesViewOptions, seedResourcesViewOptions } from './resources-view-options';

const STORAGE_KEY = 'viewOptions_4_/resources';

beforeEach(() => {
    localStorage.clear();
});

afterEach(() => {
    localStorage.clear();
});

describe('seedResourcesViewOptions', () => {
    it('falls back to the defaults when nothing is stored', () => {
        expect(seedResourcesViewOptions(STORAGE_KEY)).toEqual(defaultResourcesViewOptions);
    });

    it('keeps the choices a browser stored under the unprefixed names', () => {
        localStorage.setItem(STORAGE_KEY, JSON.stringify({ foldersFirst: false, modified: false, size: false }));

        expect(seedResourcesViewOptions(STORAGE_KEY)).toEqual({
            isFoldersFirst: false,
            isModifiedRelative: true,
            isModifiedVisible: false,
            isSizeVisible: false,
        });
    });

    it('prefers the prefixed name when a bag carries both', () => {
        localStorage.setItem(STORAGE_KEY, JSON.stringify({ isSizeVisible: true, size: false }));

        expect(seedResourcesViewOptions(STORAGE_KEY).isSizeVisible).toBe(true);
    });

    it('reads the choices the pre-unification key still holds', () => {
        localStorage.setItem('column_4_/resources', JSON.stringify({ size: false }));

        expect(seedResourcesViewOptions(STORAGE_KEY).isSizeVisible).toBe(false);
    });
});
