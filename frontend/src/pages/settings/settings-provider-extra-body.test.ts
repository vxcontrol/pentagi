import { describe, expect, it } from 'vitest';

import type { ProviderType } from '@/graphql/types';

import { normalizeGraphQLData, optionalJsonObject, transformFormToGraphQL } from './settings-provider';

describe('extraBody JSON validation', () => {
    it.each([
        ['undefined', undefined],
        ['empty string', ''],
        ['whitespace only', '   \n  '],
        ['empty object', '{}'],
        ['nested object', '{"chat_template_kwargs": {"enable_thinking": false}}'],
    ])('accepts %s', (_label, value) => {
        expect(optionalJsonObject.safeParse(value).success).toBe(true);
    });

    it.each([
        ['an array', '[1, 2]'],
        ['a bare string', '"hello"'],
        ['a bare number', '42'],
        ['malformed JSON', '{not: json}'],
        ['a trailing comma', '{"a": 1,}'],
    ])('rejects %s', (_label, value) => {
        expect(optionalJsonObject.safeParse(value).success).toBe(false);
    });
});

const formWithExtraBody = (extraBody: string) => ({
    agents: { simple: { extraBody, model: 'm' } },
    name: 'provider',
    type: 'openai' as ProviderType,
});

describe('transformFormToGraphQL extraBody', () => {
    it('parses the JSON text into an object for the mutation', () => {
        const { agents } = transformFormToGraphQL(formWithExtraBody('{"a": 1, "b": {"c": 2}}'));

        expect(agents.simple.extraBody).toEqual({ a: 1, b: { c: 2 } });
    });

    it('sends null for empty or whitespace-only text', () => {
        expect(transformFormToGraphQL(formWithExtraBody('')).agents.simple.extraBody).toBeNull();
        expect(transformFormToGraphQL(formWithExtraBody('  ')).agents.simple.extraBody).toBeNull();
    });
});

describe('normalizeGraphQLData extraBody', () => {
    it('serializes an object into indented JSON text for the field', () => {
        const result = normalizeGraphQLData({ extraBody: { enable_thinking: false } }) as { extraBody: string };

        expect(result.extraBody).toBe('{\n    "enable_thinking": false\n}');
    });

    it('renders a missing extraBody as an empty string', () => {
        expect((normalizeGraphQLData({ extraBody: null }) as { extraBody: string }).extraBody).toBe('');
    });

    it('does not recurse into the user JSON (keeps a nested __typename key)', () => {
        const result = normalizeGraphQLData({ extraBody: { __typename: 'kept' } }) as { extraBody: string };

        expect(result.extraBody).toContain('__typename');
    });
});

describe('extraBody load → submit round-trip', () => {
    it('preserves the object through normalize then transform', () => {
        const original = { chat_template_kwargs: { enable_thinking: false }, top_k: 20 };
        const normalized = normalizeGraphQLData({ simple: { extraBody: original, model: 'm' } }) as {
            simple: { extraBody: string; model: string };
        };

        const { agents } = transformFormToGraphQL({
            agents: normalized,
            name: 'provider',
            type: 'openai' as ProviderType,
        });

        expect(agents.simple.extraBody).toEqual(original);
    });
});
