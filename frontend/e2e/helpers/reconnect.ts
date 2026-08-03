import type { Page } from '@playwright/test';

import type { MockWorld } from '../mocks/world.ts';

/** Cassette entries gated on this serve only after a drop — see `dropAndReconnect`. */
export const RECONNECTED_FLAG = 'reconnected';

/**
 * Drops every live mock socket and fast-forwards the page clock past
 * graphql-ws's jittered retryWait (1–4s for the first retry), so the reconnect
 * happens deterministically instead of on a random timer.
 */
export const dropAndReconnect = async (page: Page, world: MockWorld): Promise<void> => {
    world.raiseFlag(RECONNECTED_FLAG);
    world.dropSockets();
    await page.clock.fastForward(5_000);
};
