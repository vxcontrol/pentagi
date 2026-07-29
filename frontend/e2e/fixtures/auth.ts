import type { Page } from '@playwright/test';

import type { AuthInfo } from '@/models/info';
import type { User } from '@/models/user';

import type { RestCassetteEntry } from '../mocks/cassette.ts';

import { CASSETTE_EPOCH } from './backend.ts';

// Mirrors AUTH_STORAGE_KEY in src/providers/user-provider.tsx.
const AUTH_STORAGE_KEY = 'auth';

export const SEEDED_USER = {
    created_at: '2026-01-10T09:00:00Z',
    hash: 'e2e-user-hash',
    id: 1,
    mail: 'admin@pentagi.com',
    name: 'admin',
    password_change_required: false,
    role_id: 1,
    status: 'active',
    type: 'local',
} satisfies User;

// The admin role's privilege set as /api/v1/info returns it. An empty list is a state the
// backend never returns — it renders every privilege-gated control in its denied variant.
const ADMIN_PRIVILEGES = [
    'users.create',
    'users.delete',
    'users.edit',
    'users.view',
    'roles.view',
    'providers.view',
    'usage.admin',
    'usage.view',
    'screenshots.admin',
    'screenshots.view',
    'screenshots.download',
    'screenshots.subscribe',
    'msglogs.admin',
    'msglogs.view',
    'msglogs.subscribe',
    'termlogs.admin',
    'termlogs.view',
    'termlogs.subscribe',
    'flows.admin',
    'flows.create',
    'flows.delete',
    'flows.edit',
    'flows.view',
    'flows.subscribe',
    'tasks.admin',
    'tasks.view',
    'tasks.subscribe',
    'subtasks.admin',
    'subtasks.view',
    'containers.admin',
    'containers.view',
    'agentlogs.admin',
    'agentlogs.view',
    'agentlogs.subscribe',
    'vecstorelogs.admin',
    'vecstorelogs.view',
    'vecstorelogs.subscribe',
    'searchlogs.admin',
    'searchlogs.view',
    'searchlogs.subscribe',
    'assistants.admin',
    'assistants.create',
    'assistants.delete',
    'assistants.edit',
    'assistants.view',
    'assistants.subscribe',
    'assistantlogs.admin',
    'assistantlogs.view',
    'assistantlogs.subscribe',
    'settings.admin',
    'settings.view',
    'settings.providers.admin',
    'settings.providers.view',
    'settings.providers.edit',
    'settings.providers.subscribe',
    'settings.prompts.admin',
    'settings.prompts.view',
    'settings.prompts.edit',
    'settings.tokens.admin',
    'settings.tokens.create',
    'settings.tokens.view',
    'settings.tokens.edit',
    'settings.tokens.delete',
    'settings.tokens.subscribe',
    'settings.user.admin',
    'settings.user.view',
    'settings.user.edit',
    'settings.user.subscribe',
    'templates.admin',
    'templates.create',
    'templates.view',
    'templates.edit',
    'templates.delete',
    'templates.subscribe',
    'flow_files.admin',
    'flow_files.view',
    'flow_files.upload',
    'flow_files.edit',
    'flow_files.delete',
    'flow_files.download',
    'flow_files.subscribe',
    'resources.admin',
    'resources.view',
    'resources.upload',
    'resources.edit',
    'resources.delete',
    'resources.download',
    'resources.subscribe',
    'knowledge.admin',
    'knowledge.view',
    'knowledge.create',
    'knowledge.edit',
    'knowledge.delete',
    'knowledge.search',
    'knowledge.subscribe',
    'anonymize.call',
    'toolcalls.admin',
    'toolcalls.view',
];

const OAUTH_PROVIDERS = ['google', 'github'];

export const seededAuthInfo = (): AuthInfo => ({
    expires_at: new Date(CASSETTE_EPOCH.getTime() + 12 * 60 * 60 * 1000).toISOString(),
    oauth: false,
    privileges: ADMIN_PRIVILEGES,
    providers: OAUTH_PROVIDERS,
    role: { id: 1, name: 'admin' },
    type: 'user',
    user: SEEDED_USER,
});

export const authenticatedInfoEntry = (): RestCassetteEntry => ({
    body: { data: seededAuthInfo(), status: 'success' },
});

export const guestInfoEntry = (): RestCassetteEntry => ({
    body: { data: { providers: OAUTH_PROVIDERS, type: 'guest' }, status: 'success' },
});

export const seedAuthenticated = async (page: Page): Promise<void> => {
    await page.addInitScript(
        ([key, value, sentinel]) => {
            // Init scripts run on every document, so an unguarded seed would resurrect the session
            // the app just cleared and no logout/expiry spec could observe the signed-out state.
            if (window.sessionStorage.getItem(String(sentinel))) {
                return;
            }

            window.sessionStorage.setItem(String(sentinel), '1');
            window.localStorage.setItem(String(key), String(value));
        },
        [AUTH_STORAGE_KEY, JSON.stringify(seededAuthInfo()), 'e2e-auth-seeded'] as const,
    );
};
