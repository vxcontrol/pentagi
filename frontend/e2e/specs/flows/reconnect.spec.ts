import { expect, test } from '../../fixtures/test.ts';
import { expectCleanPage } from '../../helpers/errors.ts';
import { dropAndReconnect } from '../../helpers/reconnect.ts';
import { assertNoDuplicates, extractMessageIds, MESSAGE_ID_SELECTOR } from '../../helpers/subscriptions.ts';
import {
    FLOW_A_INITIAL_IDS,
    FLOW_A_RECONNECT_ID,
    FLOW_A_REPLAY_SENTINEL_ID,
    FLOW_A_STREAMED_IDS,
    flowsCassette,
    REPLAY_FLAG,
} from '../../mocks/cassettes/flows.ts';

const BEFORE_DROP = [...FLOW_A_INITIAL_IDS, ...FLOW_A_STREAMED_IDS];
const AFTER_RECONNECT = [...BEFORE_DROP, FLOW_A_RECONNECT_ID];
const AFTER_REPLAY = [...AFTER_RECONNECT, FLOW_A_REPLAY_SENTINEL_ID];

test.describe('flow reconnect', { tag: ['@flows', '@smoke'] }, () => {
    test.use({ cassette: flowsCassette() });

    test('reconciles the missed delta exactly once after a socket drop', async ({ page, pageErrorLog, world }) => {
        // The GraphQL sweep after a reconnect skips the resources slot (it is cache-only), so the
        // REST re-hydration rides on the `ws:reconnected` event alone. Count the calls rather than
        // arming a one-shot expectation: the mount-time hydration would satisfy that on its own.
        const resourceLoads: string[] = [];

        page.on('request', (request) => {
            if (request.method() === 'GET' && new URL(request.url()).pathname.startsWith('/api/v1/resources')) {
                resourceLoads.push(request.url());
            }
        });

        await page.goto('/flows');
        await page.getByRole('row', { name: /E2E Alpha/ }).click();
        await expect(page.locator(MESSAGE_ID_SELECTOR)).toHaveCount(BEFORE_DROP.length);

        const loadsBeforeDrop = resourceLoads.length;

        await dropAndReconnect(page, world);

        await expect
            .poll(() => resourceLoads.length, { message: 'the reconnect re-hydrates the REST resources slot' })
            .toBe(loadsBeforeDrop + 1);

        await expect(page.locator(MESSAGE_ID_SELECTOR)).toHaveCount(AFTER_RECONNECT.length);

        world.raiseFlag(REPLAY_FLAG);

        await expect(page.locator(MESSAGE_ID_SELECTOR)).toHaveCount(AFTER_REPLAY.length);

        const ids = await extractMessageIds(page.locator('body'));

        assertNoDuplicates(ids);
        expect([...ids].sort()).toEqual([...AFTER_REPLAY].sort());
        expectCleanPage(pageErrorLog);
    });
});
