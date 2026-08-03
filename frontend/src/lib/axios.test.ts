import { describe, expect, it } from 'vitest';

import { resolveApiErrorMessage } from './axios';

const apiError = (code: string, msg: string) => ({ response: { data: { code, msg, status: 'error' } } });
const messages = { 'Users.NotFound': 'User not found' };

describe('resolveApiErrorMessage', () => {
    it('prefers a mapped code over the raw server message', () => {
        expect(resolveApiErrorMessage(apiError('Users.NotFound', 'user not found'), messages, 'fallback')).toBe(
            'User not found',
        );
    });

    it('falls back to the server message for an unmapped code', () => {
        expect(resolveApiErrorMessage(apiError('Some.Other.Code', 'raw server message'), messages, 'fallback')).toBe(
            'raw server message',
        );
    });

    it('uses the fallback when there is neither a mapped code nor a server message', () => {
        expect(resolveApiErrorMessage({}, messages, 'fallback')).toBe('fallback');
    });
});
