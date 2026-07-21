import { describe, expect, it } from 'vitest';

import { tokenNameSchema } from './settings-api-tokens';

describe('tokenNameSchema', () => {
    it('accepts a name at the 100-character boundary', () => {
        expect(tokenNameSchema.safeParse('a'.repeat(100)).success).toBe(true);
    });

    it('rejects a name one character over the 100-character cap', () => {
        expect(tokenNameSchema.safeParse('a'.repeat(101)).success).toBe(false);
    });

    it('trims surrounding whitespace before length-checking', () => {
        expect(tokenNameSchema.safeParse(`  ${'a'.repeat(100)}  `)).toMatchObject({
            data: 'a'.repeat(100),
            success: true,
        });
    });

    it('defaults an omitted name to an empty string', () => {
        expect(tokenNameSchema.safeParse(undefined)).toMatchObject({ data: '', success: true });
    });
});
