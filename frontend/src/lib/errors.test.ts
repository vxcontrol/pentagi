import { describe, expect, it } from 'vitest';

import { isNotFoundError } from './errors';

describe('isNotFoundError', () => {
    // The two shapes the backend uses for a genuinely missing record — these redirect to the list.
    it.each(['no rows in result set', 'flow not found', 'Record Not Found'])('treats %j as not-found', (message) => {
        expect(isNotFoundError(new Error(message))).toBe(true);
    });

    // Everything else is a real load failure — the detail page must keep the user behind Retry,
    // not bounce them, so these must NOT read as not-found.
    it.each(['network error', 'Failed to fetch', 'connection refused', 'internal server error', 'permission denied'])(
        'treats %j as a real failure',
        (message) => {
            expect(isNotFoundError(new Error(message))).toBe(false);
        },
    );
});
