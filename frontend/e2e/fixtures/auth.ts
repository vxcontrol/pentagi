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

export const seededAuthInfo = (): AuthInfo => ({
    expires_at: new Date(CASSETTE_EPOCH.getTime() + 12 * 60 * 60 * 1000).toISOString(),
    privileges: [],
    role: { id: 1, name: 'admin' },
    type: 'user',
    user: SEEDED_USER,
});

export const authenticatedInfoEntry = (): RestCassetteEntry => ({
    body: { data: seededAuthInfo(), status: 'success' },
});

export const guestInfoEntry = (): RestCassetteEntry => ({
    body: { data: { type: 'guest' }, status: 'success' },
});

export const seedAuthenticated = async (page: Page): Promise<void> => {
    await page.addInitScript(
        ([key, value]) => {
            window.localStorage.setItem(String(key), String(value));
        },
        [AUTH_STORAGE_KEY, JSON.stringify(seededAuthInfo())] as const,
    );
};
