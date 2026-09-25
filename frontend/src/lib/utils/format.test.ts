import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { formatAccountProvider, formatDate, formatTableDate } from './format';

describe('formatAccountProvider', () => {
    it('names the providers the product knows', () => {
        expect(formatAccountProvider('github')).toBe('GitHub');
        expect(formatAccountProvider('google')).toBe('Google');
    });

    it('passes an unknown provider through rather than hiding it', () => {
        expect(formatAccountProvider('gitlab')).toBe('gitlab');
    });

    it('says nothing when the record does not name one', () => {
        expect(formatAccountProvider(undefined)).toBeNull();
        expect(formatAccountProvider('')).toBeNull();
    });
});

describe('formatTableDate', () => {
    beforeEach(() => {
        vi.useFakeTimers({ now: new Date('2026-09-21T15:30:00') });
    });

    afterEach(() => {
        vi.useRealTimers();
    });

    it('keeps the time for a date from this year', () => {
        const today = new Date('2026-09-21T08:39:07');
        const earlier = new Date('2026-03-04T23:59:00');

        expect(formatTableDate(today)).toBe(formatDate(today));
        expect(formatTableDate(today)).toBe('08:39:07');
        expect(formatTableDate(earlier)).toBe('23:59, 4 Mar');
    });

    it('drops the time for a date from another year', () => {
        expect(formatTableDate(new Date('2025-09-30T23:59:00'))).toBe('30 Sep 2025');
        expect(formatTableDate(new Date('2027-01-01T00:00:00'))).toBe('1 Jan 2027');
    });
});
