import type { ResultOf } from '@graphql-typed-document-node/core';

import type {
    FlowsDocument,
    FlowTemplatesDocument,
    ProvidersDocument,
    SettingsDocument,
    SettingsUserDocument,
} from '@/graphql/types';

import type { Cassette } from '../cassette.ts';

import { authenticatedInfoEntry, guestInfoEntry } from '../../fixtures/auth.ts';

const flows: ResultOf<typeof FlowsDocument> = { flows: [] };

const flowTemplates: ResultOf<typeof FlowTemplatesDocument> = { flowTemplates: [] };

const providers: ResultOf<typeof ProvidersDocument> = { providers: [] };

const settings: ResultOf<typeof SettingsDocument> = {
    settings: {
        askUser: false,
        assistantUseAgents: false,
        debug: false,
        dockerInside: false,
        isDevelopMode: false,
        version: 'e2e',
    },
};

const settingsUser: ResultOf<typeof SettingsUserDocument> = {
    settingsUser: {
        favoriteFlows: [],
        id: '1',
    },
};

const queries: Cassette['queries'] = {
    flows: [{ data: flows }],
    flowTemplates: [{ data: flowTemplates }],
    providers: [{ data: providers }],
    settings: [{ data: settings }],
    settingsUser: [{ data: settingsUser }],
};

const resourcesEntry = { body: { data: { items: [] }, status: 'success' } };

/** Seeded-session smoke: every request answers as an authenticated user. */
export const smokeCassette: Cassette = {
    queries,
    rest: {
        'GET /api/v1/info': [authenticatedInfoEntry()],
        'GET /api/v1/resources/': [resourcesEntry],
    },
};

/** Login-form journey: /info serves guest until the login mutation raises the flag. */
export const loginJourneyCassette: Cassette = {
    queries,
    rest: {
        'GET /api/v1/info': [guestInfoEntry(), { ...authenticatedInfoEntry(), whenFlag: 'authenticated' }],
        'GET /api/v1/resources/': [resourcesEntry],
        'POST /api/v1/auth/login': [{ body: { data: {}, status: 'success' }, setFlag: 'authenticated' }],
    },
};
