import { gql, type TypedDocumentNode } from '@apollo/client';
import { describe, expect, it } from 'vitest';

import { apolloTitle, isApolloTitle } from './apollo-title';

const noopDocument = gql`
    query Noop {
        __typename
    }
` as TypedDocumentNode<unknown, Record<string, never>>;

describe('apolloTitle()', () => {
    // The single most important invariant of the module: the factory must
    // produce a value that `isApolloTitle()` recognizes. If this ever breaks
    // (e.g. someone replaces `Object.assign` with a spread that drops Symbol
    // keys, or swaps `Symbol.for` for a fresh `Symbol`), `DocumentTitle`
    // silently misclassifies reactive titles as resolvers and the next
    // production build crashes on routes like /flows/:id.
    it('returns a component recognized by isApolloTitle', () => {
        const Title = apolloTitle({
            document: noopDocument,
            select: () => 'x',
            variables: () => null,
        });

        expect(isApolloTitle(Title)).toBe(true);
    });

    it('isApolloTitle rejects plain functions and non-functions', () => {
        expect(isApolloTitle(() => 'x')).toBe(false);
        expect(isApolloTitle(function named() {})).toBe(false);
        expect(isApolloTitle('static title')).toBe(false);
        expect(isApolloTitle(null)).toBe(false);
        expect(isApolloTitle(undefined)).toBe(false);
        expect(isApolloTitle({})).toBe(false);
    });

    it('survives function name stripping (simulates Terser minification)', () => {
        const Title = apolloTitle({
            document: noopDocument,
            select: () => 'x',
            variables: () => null,
        });

        // Redefine `.name` to '' to mimic what Terser does to anonymous-in-
        // source function expressions. The marker is property-based, so the
        // guard must still pass — this is the regression that caused the
        // /flows/:id crash on production.
        Object.defineProperty(Title, 'name', { value: '' });

        expect(Title.name).toBe('');
        expect(isApolloTitle(Title)).toBe(true);
    });
});
