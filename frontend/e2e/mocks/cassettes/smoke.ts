import type { Cassette } from '../cassette.ts';

import { authenticatedInfoEntry, guestInfoEntry } from '../../fixtures/auth.ts';
import { baseQueries, baseRest } from './base.ts';

/** Seeded-session smoke: every request answers as an authenticated user. */
export const smokeCassette: Cassette = {
    queries: baseQueries(),
    rest: baseRest(),
};

/** Login-form journey: /info serves guest until the login mutation raises the flag. */
export const loginJourneyCassette: Cassette = {
    queries: baseQueries(),
    rest: {
        ...baseRest(),
        'GET /api/v1/info': [guestInfoEntry(), { ...authenticatedInfoEntry(), whenFlag: 'authenticated' }],
        'POST /api/v1/auth/login': [{ body: { data: {}, status: 'success' }, setFlag: 'authenticated' }],
    },
};
