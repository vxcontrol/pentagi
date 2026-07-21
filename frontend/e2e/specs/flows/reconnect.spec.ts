import { expect, test } from '../../fixtures/test.ts';
import { expectCleanPage } from '../../helpers/errors.ts';
import { dropAndReconnect } from '../../helpers/reconnect.ts';
import { assertNoDuplicates, extractMessageIds, MESSAGE_ID_TESTID } from '../../helpers/subscriptions.ts';
import {
    FLOW_A_INITIAL_IDS,
    FLOW_A_RECONNECT_ID,
    FLOW_A_STREAMED_IDS,
    flowsCassette,
} from '../../mocks/cassettes/flows.ts';

const BEFORE_DROP = [...FLOW_A_INITIAL_IDS, ...FLOW_A_STREAMED_IDS];
const AFTER_RECONNECT = [...BEFORE_DROP, FLOW_A_RECONNECT_ID];

test.describe('flow reconnect', { tag: ['@flows', '@smoke'] }, () => {
    test.use({ cassette: flowsCassette() });

    test('reconciles the missed delta exactly once after a socket drop', async ({ page, pageErrorLog, world }) => {
        await page.goto('/flows');
        await page.getByRole('row', { name: /E2E Alpha/ }).click();
        await expect(page.getByTestId(MESSAGE_ID_TESTID)).toHaveCount(BEFORE_DROP.length);

        await dropAndReconnect(page, world);

        await expect(page.getByTestId(MESSAGE_ID_TESTID)).toHaveCount(AFTER_RECONNECT.length);

        const ids = await extractMessageIds(page.locator('body'));

        assertNoDuplicates(ids);
        expect([...ids].sort()).toEqual([...AFTER_RECONNECT].sort());
        expectCleanPage(pageErrorLog);
    });
});
