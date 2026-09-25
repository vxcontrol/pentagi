import { readFileSync } from 'node:fs';
import { describe, expect, it } from 'vitest';

import { KNOWLEDGE_LIMITS } from '@/features/knowledges/knowledge-form';

import { repoFile } from './paths.ts';

const LIMITS_GO = repoFile('backend', 'pkg', 'database', 'knowledge', 'limits', 'limits.go');
const REST_MODEL_GO = repoFile('backend', 'pkg', 'server', 'models', 'knowledge.go');

type Field = 'codeLang' | 'content' | 'description' | 'question';

const FIELDS: readonly Field[] = ['codeLang', 'content', 'description', 'question'];

const GO_CONSTANT: Readonly<Record<Field, string>> = {
    codeLang: 'MaxCodeLangLen',
    content: 'MaxContentLen',
    description: 'MaxDescriptionLen',
    question: 'MaxQuestionLen',
};

const REST_FIELD: Readonly<Record<Field, string>> = {
    codeLang: 'CodeLang',
    content: 'Content',
    description: 'Description',
    question: 'Question',
};

const REST_STRUCTS = ['CreateKnowledgeDocRequest', 'UpdateKnowledgeDocRequest'] as const;

const formLimit = (field: Field): number => KNOWLEDGE_LIMITS[field];

const graphLimit = (constant: string): number => {
    const source = readFileSync(LIMITS_GO, 'utf8');
    const match = new RegExp(`\\b${constant}\\s*=\\s*(\\d+)`).exec(source);

    expect(match, `${constant} is not declared in ${LIMITS_GO}`).toBeTruthy();

    return Number(match?.[1]);
};

const restLimit = (struct: string, field: string): number => {
    const source = readFileSync(REST_MODEL_GO, 'utf8');
    const body = new RegExp(`type ${struct} struct \\{([^}]*)\\}`).exec(source)?.[1];

    expect(body, `${struct} is not declared in ${REST_MODEL_GO}`).toBeTruthy();

    const line = (body ?? '').split('\n').find((row) => new RegExp(`^\\s*${field}\\s`).test(row));

    expect(line, `${struct} has no ${field} field`).toBeTruthy();

    const max = /validate:"[^"]*\bmax=(\d+)/.exec(line ?? '');

    expect(max, `${struct}.${field} carries no max= in its validate tag`).toBeTruthy();

    return Number(max?.[1]);
};

describe('knowledge document length limits', () => {
    it.each(FIELDS)('the Go limit and the form agree on %s', (field) => {
        expect(graphLimit(GO_CONSTANT[field])).toBe(formLimit(field));
    });

    it.each(REST_STRUCTS.flatMap((struct) => FIELDS.map((field) => [struct, field] as const)))(
        'the REST body %s and the form agree on %s',
        (struct, field) => {
            expect(restLimit(struct, REST_FIELD[field])).toBe(formLimit(field));
        },
    );
});
