import { expect, test } from '../../fixtures/test.ts';
import { expectCleanPage } from '../../helpers/errors.ts';
import { dropAndReconnect } from '../../helpers/reconnect.ts';
import { assertNoDuplicates, extractMessageIds, MESSAGE_ID_TESTID } from '../../helpers/subscriptions.ts';
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
        await page.goto('/flows');
        await page.getByRole('row', { name: /E2E Alpha/ }).click();
        await expect(page.getByTestId(MESSAGE_ID_TESTID)).toHaveCount(BEFORE_DROP.length);

        await dropAndReconnect(page, world);

        await expect(page.getByTestId(MESSAGE_ID_TESTID)).toHaveCount(AFTER_RECONNECT.length);

        world.raiseFlag(REPLAY_FLAG);

        await expect(page.getByTestId(MESSAGE_ID_TESTID)).toHaveCount(AFTER_REPLAY.length);

        const ids = await extractMessageIds(page.locator('body'));

        assertNoDuplicates(ids);
        expect([...ids].sort()).toEqual([...AFTER_REPLAY].sort());
        expectCleanPage(pageErrorLog);
    });
});
