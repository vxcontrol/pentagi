import { test as base, expect } from '@playwright/test';

import type { Cassette } from '../mocks/cassette.ts';

import { installMockRoutes } from '../mocks/install.ts';
import { MockWorld } from '../mocks/world.ts';

export interface BackendConfig {
    installMocks: boolean;
    tier: BackendTier;
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

export const test = base.extend<BackendFixtures & BackendOptions & { _installMocks: void }>({
    _installMocks: [
        async ({ backend, page, world }, use) => {
            if (!backend.installMocks) {
                await use();

                return;
            }

            await page.clock.install({ time: CASSETTE_EPOCH });
            await installMockRoutes(page, world);
            await use();
            expect(world.unmatched, 'every API call the app made must have a cassette entry').toEqual([]);
        },
        { auto: true },
    ],
    backend: [{ installMocks: true, tier: 'mock' }, { option: true }],
    cassette: [{}, { option: true }],
    world: async ({ cassette }, use) => {
        await use(new MockWorld(cassette));
    },
});
