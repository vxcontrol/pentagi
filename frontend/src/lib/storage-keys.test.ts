import { describe, expect, it } from 'vitest';

import {
    getPeriodStorageKey,
    getStorageKey,
    getTableStorageKey,
    getTopLevelPath,
    getViewOptionsStorageKey,
} from './storage-keys';

describe('getTopLevelPath', () => {
    it('returns the first segment with a leading slash for top-level paths', () => {
        expect(getTopLevelPath('/flows')).toBe('/flows');
    });

    it('strips the id from detail-page paths', () => {
        expect(getTopLevelPath('/flows/abc-123')).toBe('/flows');
        expect(getTopLevelPath('/knowledges/abc/foo')).toBe('/knowledges');
    });

    it('returns an empty string for the root', () => {
        expect(getTopLevelPath('/')).toBe('');
        expect(getTopLevelPath('')).toBe('');
    });

    it('handles paths without a leading slash', () => {
        expect(getTopLevelPath('flows/123')).toBe('/flows');
    });

    it('skips repeated leading slashes', () => {
        expect(getTopLevelPath('//flows/123')).toBe('/flows');
    });
});

describe('getStorageKey', () => {
    it('joins type and path with the `_4_` separator', () => {
        expect(getStorageKey('table', '/flows')).toBe('table_4_/flows');
        expect(getStorageKey('period', '/dashboard')).toBe('period_4_/dashboard');
    });
});

describe('typed key helpers', () => {
    it('getTableStorageKey delegates to the unified `table` namespace', () => {
        expect(getTableStorageKey('/flows')).toBe('table_4_/flows');
    });

    it('getPeriodStorageKey delegates to the `period` namespace', () => {
        expect(getPeriodStorageKey('/dashboard')).toBe('period_4_/dashboard');
    });

    it('getViewOptionsStorageKey delegates to the `viewOptions` namespace', () => {
        expect(getViewOptionsStorageKey('/resources')).toBe('viewOptions_4_/resources');
    });
});
