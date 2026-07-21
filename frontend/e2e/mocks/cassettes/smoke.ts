import type { Cassette } from '../cassette.ts';

import { authenticatedInfoEntry, guestInfoEntry, SEEDED_USER } from '../../fixtures/auth.ts';
import { baseQueries, baseRest } from './base.ts';

export const smokeCassette: Cassette = {
    queries: baseQueries(),
    rest: baseRest(),
};

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

// The real backend answers 401 here (not a 200 error envelope); the axios
// interceptor clears the auth key and rejects, so the app hits its catch branch.
export const loginFailCassette: Cassette = {
    queries: baseQueries(),
    rest: {
        ...baseRest(),
        'GET /api/v1/info': [guestInfoEntry()],
        'POST /api/v1/auth/login': [
            {
                body: { code: 'Auth.InvalidCredentials', msg: 'invalid login or password', status: 'error' },
                bodySubset: { mail: SEEDED_USER.mail, password: 'wrong-password' },
                status: 401,
            },
        ],
    },
};
