import type { ResultOf } from '@graphql-typed-document-node/core';

import type { ApiTokenFragmentFragment, ApiTokensDocument, ApiTokenWithSecretFragmentFragment } from '@/graphql/types';

import { TokenStatus } from '@/graphql/types';

import type { Cassette } from '../cassette.ts';

import { entity, mergeCassettes } from '../cassette.ts';
import { baseQueries, baseRest } from './base.ts';

const T = '2026-01-15T08:00:00Z';
const THIRTY_DAYS_SECONDS = 30 * 24 * 60 * 60;

const tokenFields = (id: string, name: string) => ({
    createdAt: T,
    id,
    name,
    roleId: '1',
    status: TokenStatus.Active,
    tokenId: `tok-${id}`,
    ttl: THIRTY_DAYS_SECONDS,
    updatedAt: T,
    userId: '1',
});

const makeToken = (id: string, name: string): ApiTokenFragmentFragment => entity('APIToken', tokenFields(id, name));

export const SEED_TOKEN = makeToken('1', 'E2E seed token');
export const DOOMED_TOKEN = makeToken('2', 'E2E doomed token');
export const CREATED_TOKEN = makeToken('3', 'E2E created token');
export const CREATED_TOKEN_SECRET = 'e2e-fake-token-secret';

export const CREATED_TOKEN_WITH_SECRET: ApiTokenWithSecretFragmentFragment = entity('APITokenWithSecret', {
    ...tokenFields('3', 'E2E created token'),
    token: CREATED_TOKEN_SECRET,
});

export const tokensList = (...tokens: ApiTokenFragmentFragment[]): ResultOf<typeof ApiTokensDocument> => ({
    apiTokens: tokens,
});

export const apiTokensCassette = (override: Cassette = {}): Cassette =>
    mergeCassettes(
        {
            queries: {
                ...baseQueries(),
                apiTokens: [{ data: tokensList(SEED_TOKEN) }],
            },
            rest: baseRest(),
        },
        override,
    );
