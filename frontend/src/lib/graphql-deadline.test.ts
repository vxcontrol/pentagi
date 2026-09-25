import { describe, expect, it } from 'vitest';

import { TestAgentDocument, TestProviderDocument } from '@/graphql/types';

import { isUnboundedOperation, UNBOUNDED_OPERATIONS } from './graphql-deadline';

const body = (operationName: string) => JSON.stringify({ operationName, query: '{ __typename }', variables: {} });

const operationNameOf = (document: unknown): string => {
    const [definition] = (document as { definitions: { name?: { value?: string } }[] }).definitions;

    return definition?.name?.value ?? '';
};

describe('isUnboundedOperation', () => {
    it('leaves a model run the operator asked for without a deadline of ours', () => {
        expect(isUnboundedOperation(body('testProvider'))).toBe(true);
        expect(isUnboundedOperation(body('testAgent'))).toBe(true);
    });

    it('bounds an ordinary operation', () => {
        expect(isUnboundedOperation(body('createFlow'))).toBe(false);
    });

    it('bounds anything it cannot read, rather than exempting it', () => {
        expect(isUnboundedOperation('not json at all')).toBe(false);
        expect(isUnboundedOperation(JSON.stringify({ query: '{ __typename }' }))).toBe(false);
        expect(isUnboundedOperation(undefined)).toBe(false);
        expect(isUnboundedOperation(new FormData())).toBe(false);
    });

    // The exemption is matched against the name the client actually sends, so a
    // renamed operation would exempt nothing and the probe would be cut off.
    it('names the operations the documents really carry', () => {
        expect(UNBOUNDED_OPERATIONS).toContain(operationNameOf(TestProviderDocument));
        expect(UNBOUNDED_OPERATIONS).toContain(operationNameOf(TestAgentDocument));
    });
});
