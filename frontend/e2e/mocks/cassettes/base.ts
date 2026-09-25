import type { ResultOf } from '@graphql-typed-document-node/core';

import type {
    FlowsDocument,
    FlowTemplatesDocument,
    ProvidersDocument,
    ResourcesDocument,
    SettingsDocument,
    SettingsUserDocument,
    VersionInfoDocument,
} from '@/graphql/types';

import { UpdateState } from '@/graphql/types';

import type { Cassette } from '../cassette.ts';

import { authenticatedInfoEntry } from '../../fixtures/auth.ts';
import { entity } from '../cassette.ts';

const flows: ResultOf<typeof FlowsDocument> = { flows: [] };

const flowTemplates: ResultOf<typeof FlowTemplatesDocument> = { flowTemplates: [] };

// After a websocket reconnect the app re-requests resources on the wire with {recursive:true};
// removing this entry red-lights the reconnect spec on the 501 gate.
const resources: ResultOf<typeof ResourcesDocument> = { resources: [] };

const providers: ResultOf<typeof ProvidersDocument> = { providers: [] };

const versionInfo: ResultOf<typeof VersionInfoDocument> = {
    versionInfo: entity('VersionInfo', {
        build: 'e2e00000',
        checkedAt: null,
        current: 'e2e',
        failedAt: null,
        latest: null,
        state: UpdateState.Disabled,
        strategy: 'preview',
    }),
};

const settings: ResultOf<typeof SettingsDocument> = {
    settings: entity('Settings', {
        askUser: false,
        assistantUseAgents: false,
        debug: false,
        dockerInside: false,
        isDevelopMode: false,
        version: 'e2e',
    }),
};

export const SIGNED_IN_USER_ID = '1';

const settingsUser: ResultOf<typeof SettingsUserDocument> = {
    settingsUser: entity('UserPreferences', {
        favoriteFlows: [],
        id: SIGNED_IN_USER_ID,
    }),
};

export const baseQueries = (): NonNullable<Cassette['queries']> => ({
    flows: [{ data: flows }],
    flowTemplates: [{ data: flowTemplates }],
    providers: [{ data: providers }],
    resources: [{ data: resources }],
    settings: [{ data: settings }],
    settingsUser: [{ data: settingsUser }],
    versionInfo: [{ data: versionInfo }],
});

export const baseRest = (): NonNullable<Cassette['rest']> => ({
    'GET /api/v1/info': [authenticatedInfoEntry()],
    'GET /api/v1/resources/': [{ body: { data: { items: [] }, status: 'success' } }],
});
