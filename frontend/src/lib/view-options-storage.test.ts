import { afterEach, beforeEach, describe, expect, it } from 'vitest';

import { loadViewOptions, migrateLegacyViewOptions, saveViewOptions } from './view-options-storage';

const PATH = '/resources';
const UNIFIED_KEY = `viewOptions_4_${PATH}`;
const LEGACY_KEY = `column_4_${PATH}`;

beforeEach(() => {
    localStorage.clear();
});

afterEach(() => {
    localStorage.clear();
});

describe('loadViewOptions', () => {
    it('returns an empty object when the key is missing', () => {
        expect(loadViewOptions(UNIFIED_KEY)).toEqual({});
    });

    it('returns an empty object when the payload is invalid JSON', () => {
        localStorage.setItem(UNIFIED_KEY, 'not json');
        expect(loadViewOptions(UNIFIED_KEY)).toEqual({});
    });

    it('returns an empty object when a value is not boolean', () => {
        // The schema is `Record<string, boolean>` — a numeric value should
        // be rejected wholesale rather than silently coerced.
        localStorage.setItem(UNIFIED_KEY, JSON.stringify({ foldersFirst: 1 }));
        expect(loadViewOptions(UNIFIED_KEY)).toEqual({});
    });

    it('round-trips a multi-flag payload', () => {
        const value = { foldersFirst: true, relativeTimestamp: false };
        localStorage.setItem(UNIFIED_KEY, JSON.stringify(value));
        expect(loadViewOptions(UNIFIED_KEY)).toEqual(value);
    });
});

describe('saveViewOptions', () => {
    it('writes the payload as JSON', () => {
        saveViewOptions(UNIFIED_KEY, { foldersFirst: true });

        const raw = localStorage.getItem(UNIFIED_KEY);
        expect(raw).not.toBeNull();
        expect(JSON.parse(raw ?? 'null')).toEqual({ foldersFirst: true });
    });

    it('removes the storage key when the payload is empty — keeps storage tidy', () => {
        localStorage.setItem(UNIFIED_KEY, JSON.stringify({ foldersFirst: true }));

        saveViewOptions(UNIFIED_KEY, {});

        expect(localStorage.getItem(UNIFIED_KEY)).toBeNull();
    });

    it('replaces (does not merge) the existing payload', () => {
        // Deliberately a setter, not a patcher — the caller must merge for partial updates.
        saveViewOptions(UNIFIED_KEY, { foldersFirst: true, relativeTimestamp: true });
        saveViewOptions(UNIFIED_KEY, { foldersFirst: false });

        expect(loadViewOptions(UNIFIED_KEY)).toEqual({ foldersFirst: false });
    });
});

describe('migrateLegacyViewOptions', () => {
    it('short-circuits to loadViewOptions when no legacy key exists', () => {
        localStorage.setItem(UNIFIED_KEY, JSON.stringify({ foldersFirst: true }));

        const result = migrateLegacyViewOptions(PATH, UNIFIED_KEY);

        expect(result).toEqual({ foldersFirst: true });
        // The unified slot must remain untouched.
        expect(JSON.parse(localStorage.getItem(UNIFIED_KEY) ?? '{}')).toEqual({ foldersFirst: true });
    });

    it('folds the legacy `column_4_<path>` payload into the unified key', () => {
        localStorage.setItem(LEGACY_KEY, JSON.stringify({ foldersFirst: true, relativeTimestamp: false }));

        const result = migrateLegacyViewOptions(PATH, UNIFIED_KEY);

        expect(result).toEqual({ foldersFirst: true, relativeTimestamp: false });
        expect(localStorage.getItem(LEGACY_KEY)).toBeNull();
    });

    it('merges legacy values on top of pre-existing unified values', () => {
        // Possible if a partial migration happened on a different tab. The
        // legacy payload represents the most recent user intent (it was
        // never cleared), so we accept it as the winner on overlap.
        localStorage.setItem(UNIFIED_KEY, JSON.stringify({ foldersFirst: false, modified: true }));
        localStorage.setItem(LEGACY_KEY, JSON.stringify({ foldersFirst: true }));

        const result = migrateLegacyViewOptions(PATH, UNIFIED_KEY);

        expect(result).toEqual({ foldersFirst: true, modified: true });
        expect(localStorage.getItem(LEGACY_KEY)).toBeNull();
    });

    it('still deletes the legacy key when its payload is invalid', () => {
        // An invalid legacy payload contributes nothing, but we must not
        // leave it in storage — a future call would otherwise repeatedly
        // re-parse the same garbage.
        localStorage.setItem(LEGACY_KEY, JSON.stringify({ foldersFirst: 'truthy' }));

        const result = migrateLegacyViewOptions(PATH, UNIFIED_KEY);

        expect(result).toEqual({});
        expect(localStorage.getItem(LEGACY_KEY)).toBeNull();
    });

    it('is idempotent — a second invocation reads only the unified key', () => {
        localStorage.setItem(LEGACY_KEY, JSON.stringify({ foldersFirst: true }));

        migrateLegacyViewOptions(PATH, UNIFIED_KEY);
        const second = migrateLegacyViewOptions(PATH, UNIFIED_KEY);

        expect(second).toEqual({ foldersFirst: true });
    });
});
