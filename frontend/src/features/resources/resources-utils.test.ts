import { describe, expect, it } from 'vitest';

import { shouldPromptToSearchLargeLibrary } from './resources-utils';

describe('shouldPromptToSearchLargeLibrary', () => {
    it('returns false when the count is at or below the threshold', () => {
        expect(shouldPromptToSearchLargeLibrary(5000, false, 5000)).toBe(false);
        expect(shouldPromptToSearchLargeLibrary(100, false, 5000)).toBe(false);
    });

    it('returns true when the count exceeds the threshold and no query is active', () => {
        expect(shouldPromptToSearchLargeLibrary(5001, false, 5000)).toBe(true);
        expect(shouldPromptToSearchLargeLibrary(100000, false, 5000)).toBe(true);
    });

    it('returns false once a search query is active, regardless of count', () => {
        expect(shouldPromptToSearchLargeLibrary(100000, true, 5000)).toBe(false);
    });
});
