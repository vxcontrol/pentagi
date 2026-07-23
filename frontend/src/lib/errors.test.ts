import { describe, expect, it } from 'vitest';

import { isNotFoundError } from './errors';

describe('isNotFoundError', () => {
    it.each(['no rows in result set', 'flow not found', 'template not found: sql: no rows', 'Record Not Found'])(
        'treats %j as not-found',
        (message) => {
            expect(isNotFoundError(new Error(message))).toBe(true);
        },
    );

    // The authz strings are real backend messages that also contain "not found" (see errors.ts) —
    // they must stay classified as real failures, so do not drop them as odd-looking fixtures.
    it.each([
        'network error',
        'Failed to fetch',
        'connection refused',
        'internal server error',
        "requested permission 'flows.read' not found",
        'not authorized to access this token',
        'no permissions granted',
        'privileges are not set',
    ])('treats %j as a real failure', (message) => {
        expect(isNotFoundError(new Error(message))).toBe(false);
    });
});
