import { describe, expect, it } from 'vitest';

import { formSchema } from './template';

const TITLE_LIMIT = 255;
const TEXT_LIMIT = 65536;

describe('the flow template form schema', () => {
    const parse = (values: Partial<{ text: string; title: string }>) =>
        formSchema.safeParse({ text: 'Body', title: 'Title', ...values });

    it('accepts a title at the 255-character boundary', () => {
        expect(parse({ title: 'a'.repeat(TITLE_LIMIT) }).success).toBe(true);
    });

    it('rejects a title one character over the 255-character cap', () => {
        expect(parse({ title: 'a'.repeat(TITLE_LIMIT + 1) }).success).toBe(false);
    });

    it('counts an astral title character once, the way the mutation does', () => {
        expect(parse({ title: '\u{1f512}'.repeat(TITLE_LIMIT) }).success).toBe(true);
        expect(parse({ title: '\u{1f512}'.repeat(TITLE_LIMIT + 1) }).success).toBe(false);
    });

    it('accepts a text at the 65536-character boundary', () => {
        expect(parse({ text: 'a'.repeat(TEXT_LIMIT) }).success).toBe(true);
    });

    it('rejects a text one character over the 65536-character cap', () => {
        expect(parse({ text: 'a'.repeat(TEXT_LIMIT + 1) }).success).toBe(false);
    });

    it('counts an astral text character once, the way the mutation does', () => {
        expect(parse({ text: '\u{1f512}'.repeat(TEXT_LIMIT) }).success).toBe(true);
        expect(parse({ text: '\u{1f512}'.repeat(TEXT_LIMIT + 1) }).success).toBe(false);
    });
});
