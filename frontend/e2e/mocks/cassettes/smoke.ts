import type { Cassette } from '../cassette.ts';

import { authenticatedInfoEntry, guestInfoEntry, SEEDED_USER } from '../../fixtures/auth.ts';
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
        'POST /api/v1/auth/login': [
            {
                body: { data: {}, status: 'success' },
                bodySubset: { mail: SEEDED_USER.mail, password: 'e2e-password' },
                setFlag: 'authenticated',
            },
        ],
    },
};

// A rejected login: the backend answers 200 with an error envelope (not a 401,
// which the axios interceptor would turn into a redirect), the session stays
// guest, and the form surfaces the error.
export const loginFailCassette: Cassette = {
    queries: baseQueries(),
    rest: {
        ...baseRest(),
        'GET /api/v1/info': [guestInfoEntry()],
        'POST /api/v1/auth/login': [{ body: { error: 'invalid credentials', status: 'error' } }],
    },
};
