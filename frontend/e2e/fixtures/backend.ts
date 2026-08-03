import { test as base, expect } from '@playwright/test';

import type { Cassette } from '../mocks/cassette.ts';

import { installMockRoutes } from '../mocks/install.ts';
import { MockWorld } from '../mocks/world.ts';
import { seedAuthenticated } from './auth.ts';

export interface BackendConfig {
    installMocks: boolean;
}

export interface BackendOptions {
    backend: BackendConfig;
    cassette: Cassette;
}

export type BackendTier = 'local' | 'mock' | 'stand';

/**
 * All cassette timestamps must stay on this (UTC) day: `formatDate` branches on
 * isToday/isThisYear, so an unpinned clock rots date renders within 24h.
 */
export const CASSETTE_EPOCH = new Date('2026-01-15T12:00:00Z');

interface BackendFixtures {
    world: MockWorld;
}

export const test = base.extend<BackendFixtures & BackendOptions & { _installMocks: void; isAuthSeeded: boolean }>({
    // Seeding client-side auth is a mock-tier shortcut only: against a real
    // backend a forged localStorage session has no cookie behind it, and the
    // first 401 wipes it and hard-redirects to login. Real tiers authenticate
    // once in the setup project and reuse storageState.
    _installMocks: [
        async ({ backend, context, isAuthSeeded, page, world }, use, testInfo) => {
            if (!backend.installMocks) {
                await use();

                return;
            }

            await page.clock.install({ time: CASSETTE_EPOCH });

            if (isAuthSeeded) {
                await seedAuthenticated(page);
            }

            await installMockRoutes(context, world);
            await use();

            // Silently-dead subscriptions are legal (cassettes mock only the
            // relevant ones), but on a failure they are the first suspect — a
            // typo'd subscription key otherwise surfaces as a UI timeout.
            if (testInfo.status !== testInfo.expectedStatus && world.unmatchedSubscriptions.length) {
                await testInfo.attach('unmatched-subscriptions', {
                    body: world.unmatchedSubscriptions.join('\n'),
                    contentType: 'text/plain',
                });
            }

            expect(world.unmatched, 'every API call the app made must have a cassette entry').toEqual([]);
        },
        { auto: true },
    ],
    backend: [{ installMocks: true }, { option: true }],
    cassette: [{}, { option: true }],
    isAuthSeeded: [true, { option: true }],
    world: async ({ cassette }, use) => {
        await use(new MockWorld(cassette));
    },
});
