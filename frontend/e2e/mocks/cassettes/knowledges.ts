import type { ResultOf } from '@graphql-typed-document-node/core';

import type {
    KnowledgeDocumentDocument,
    KnowledgeDocumentFragmentFragment,
    KnowledgeDocumentsDocument,
} from '@/graphql/types';

import { KnowledgeAnswerType, KnowledgeDocType } from '@/graphql/types';

import type { Cassette } from '../cassette.ts';

import { entity, mergeCassettes } from '../cassette.ts';
import { baseQueries, baseRest } from './base.ts';

export const makeKnowledge = (id: string, question: string): KnowledgeDocumentFragmentFragment =>
    entity('KnowledgeDocument', {
        answerType: KnowledgeAnswerType.Other,
        codeLang: null,
        content: `Content for ${question}`,
        description: null,
        docType: KnowledgeDocType.Answer,
        flowId: null,
        guideType: null,
        id,
        manual: true,
        partSize: 1024,
        question,
        subtaskId: null,
        taskId: null,
        totalSize: 1024,
        userId: '1',
    });

export const KNOWLEDGE_DOC = makeKnowledge('7', 'E2E Seed Question');

// The list query is withContent:false; the backend returns content:'' in that case (schema.graphqls).
const knowledgeDocuments: ResultOf<typeof KnowledgeDocumentsDocument> = {
    knowledgeDocuments: [{ ...KNOWLEDGE_DOC, content: '' }],
};

export const knowledgesCassette = (override: Cassette = {}): Cassette =>
    mergeCassettes(
        {
            queries: {
                ...baseQueries(),
                knowledgeDocuments: [{ data: knowledgeDocuments, variables: { withContent: false } }],
            },
            rest: baseRest(),
        },
        override,
    );

const knowledgeDocument: ResultOf<typeof KnowledgeDocumentDocument> = { knowledgeDocument: KNOWLEDGE_DOC };

/** Adds the single-document query the detail route issues for KNOWLEDGE_DOC. */
export const knowledgeDetailCassette = (override: Cassette = {}): Cassette =>
    knowledgesCassette(
        mergeCassettes(
            {
                queries: {
                    knowledgeDocument: [{ data: knowledgeDocument, variables: { id: KNOWLEDGE_DOC.id } }],
                },
            },
            override,
        ),
    );
