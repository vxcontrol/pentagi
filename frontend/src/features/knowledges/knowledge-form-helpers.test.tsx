import { describe, expect, it } from 'vitest';

import type { KnowledgeDocumentFragmentFragment } from '@/graphql/types';

import { KnowledgeAnswerType, KnowledgeDocType, KnowledgeGuideType } from '@/graphql/types';

import type { DirtyFlags, FormValues } from './knowledge-form';

import { documentToFormValues, formSchema, formValuesToCreateInput, formValuesToUpdateInput } from './knowledge-form';

const baseValues: FormValues = {
    answerType: KnowledgeAnswerType.Other,
    codeLang: '',
    content: 'some content',
    description: '',
    docType: KnowledgeDocType.Answer,
    guideType: undefined,
    question: 'a question',
};

const baseFragment: KnowledgeDocumentFragmentFragment = {
    answerType: KnowledgeAnswerType.Other,
    codeLang: null,
    content: 'frag content',
    description: null,
    docType: KnowledgeDocType.Answer,
    flowId: null,
    guideType: null,
    id: '42',
    manual: true,
    partSize: 1,
    question: 'frag question',
    subtaskId: null,
    taskId: null,
    totalSize: 1,
    userId: '7',
};

describe('formValuesToCreateInput', () => {
    it('sends all required fields', () => {
        const input = formValuesToCreateInput({
            ...baseValues,
            codeLang: 'python',
            description: 'desc',
            docType: KnowledgeDocType.Code,
        });

        expect(input).toEqual({
            answerType: KnowledgeAnswerType.Other,
            codeLang: 'python',
            content: 'some content',
            description: 'desc',
            docType: KnowledgeDocType.Code,
            guideType: undefined,
            question: 'a question',
        });
    });

    it('maps empty optional fields to undefined, not ""', () => {
        const input = formValuesToCreateInput({ ...baseValues, codeLang: '', description: '' });

        expect(input.codeLang).toBeUndefined();
        expect(input.description).toBeUndefined();
    });
});

describe('formValuesToUpdateInput', () => {
    it('always sends content and omits untouched fields', () => {
        const input = formValuesToUpdateInput(baseValues, {});

        expect(input).toEqual({ content: 'some content' });
        expect('description' in input).toBe(false);
        expect('question' in input).toBe(false);
        expect('docType' in input).toBe(false);
    });

    it('emits a dirty+cleared description as "" (not undefined)', () => {
        const input = formValuesToUpdateInput({ ...baseValues, description: '' }, { description: true });

        expect(input.description).toBe('');
    });

    it('emits a dirty+cleared codeLang as ""', () => {
        const input = formValuesToUpdateInput({ ...baseValues, codeLang: '' }, { codeLang: true });

        expect(input.codeLang).toBe('');
    });

    it.each<[keyof DirtyFlags, Partial<FormValues>, unknown]>([
        ['docType', { docType: KnowledgeDocType.Guide }, KnowledgeDocType.Guide],
        ['question', { question: 'new q' }, 'new q'],
        ['guideType', { guideType: KnowledgeGuideType.Pentest }, KnowledgeGuideType.Pentest],
        ['answerType', { answerType: KnowledgeAnswerType.Tool }, KnowledgeAnswerType.Tool],
    ])('includes %s only when dirty', (field, overrides, expected) => {
        const values = { ...baseValues, ...overrides };

        expect(formValuesToUpdateInput(values, {})[field]).toBeUndefined();
        expect(formValuesToUpdateInput(values, { [field]: true })[field]).toBe(expected);
    });
});

describe('documentToFormValues', () => {
    it('coerces nullish subtypes to undefined and nullish text to ""', () => {
        expect(documentToFormValues(baseFragment)).toEqual({
            answerType: KnowledgeAnswerType.Other,
            codeLang: '',
            content: 'frag content',
            description: '',
            docType: KnowledgeDocType.Answer,
            guideType: undefined,
            question: 'frag question',
        });
    });

    it('preserves present values', () => {
        const values = documentToFormValues({
            ...baseFragment,
            codeLang: 'go',
            description: 'a desc',
            guideType: KnowledgeGuideType.Install,
        });

        expect(values.codeLang).toBe('go');
        expect(values.description).toBe('a desc');
        expect(values.guideType).toBe(KnowledgeGuideType.Install);
    });
});

describe('formSchema', () => {
    const valid: FormValues = {
        answerType: KnowledgeAnswerType.Other,
        codeLang: '',
        content: 'content',
        description: '',
        docType: KnowledgeDocType.Answer,
        guideType: undefined,
        question: 'question',
    };

    it('accepts a valid Answer document', () => {
        expect(formSchema.safeParse(valid).success).toBe(true);
    });

    it('requires content', () => {
        const result = formSchema.safeParse({ ...valid, content: '' });

        expect(result.success).toBe(false);
        expect(result.error?.issues.find((i) => i.path[0] === 'content')?.message).toBe('Content is required');
    });

    it('requires question', () => {
        const result = formSchema.safeParse({ ...valid, question: '' });

        expect(result.success).toBe(false);
        expect(result.error?.issues.find((i) => i.path[0] === 'question')?.message).toBe('Question is required');
    });

    it('enforces the question max-length message', () => {
        const result = formSchema.safeParse({ ...valid, question: 'x'.repeat(2049) });

        expect(result.success).toBe(false);
        expect(result.error?.issues.find((i) => i.path[0] === 'question')?.message).toBe(
            'Question must be 2048 characters or fewer',
        );
    });

    it('enforces the description max-length message', () => {
        const result = formSchema.safeParse({ ...valid, description: 'x'.repeat(1001) });

        expect(result.success).toBe(false);
        expect(result.error?.issues.find((i) => i.path[0] === 'description')?.message).toBe(
            'Description must be 1000 characters or fewer',
        );
    });

    it.each<[KnowledgeDocType, keyof FormValues, string]>([
        [KnowledgeDocType.Answer, 'answerType', 'Answer type is required'],
        [KnowledgeDocType.Code, 'codeLang', 'Code language is required'],
        [KnowledgeDocType.Guide, 'guideType', 'Guide type is required'],
    ])('requires the %s subtype field', (docType, field, message) => {
        const result = formSchema.safeParse({
            ...valid,
            answerType: undefined,
            codeLang: '',
            docType,
            guideType: undefined,
        });

        expect(result.success).toBe(false);
        const issue = result.error?.issues.find((i) => i.path[0] === field);
        expect(issue?.message).toBe(message);
    });

    it.each<[KnowledgeDocType, Partial<FormValues>]>([
        [KnowledgeDocType.Answer, { answerType: KnowledgeAnswerType.Tool }],
        [KnowledgeDocType.Code, { codeLang: 'rust' }],
        [KnowledgeDocType.Guide, { guideType: KnowledgeGuideType.Use }],
    ])('passes the superRefine for %s when its subtype is present', (docType, overrides) => {
        const result = formSchema.safeParse({
            ...valid,
            answerType: undefined,
            codeLang: '',
            docType,
            guideType: undefined,
            ...overrides,
        });

        expect(result.success).toBe(true);
    });
});
