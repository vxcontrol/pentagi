import { expect, test } from '../../fixtures/test.ts';
import { expectCleanPage } from '../../helpers/errors.ts';
import {
    assertDisjoint,
    assertNoDuplicates,
    attachIdSet,
    extractMessageIds,
    MESSAGE_ID_TESTID,
} from '../../helpers/subscriptions.ts';
import {
    FLOW_A_INITIAL_IDS,
    FLOW_A_STREAMED_IDS,
    FLOW_B_INITIAL_IDS,
    FLOW_B_STREAMED_IDS,
    flowsCassette,
} from '../../mocks/cassettes/flows.ts';

const FLOW_A_IDS = [...FLOW_A_INITIAL_IDS, ...FLOW_A_STREAMED_IDS];
const FLOW_B_IDS = [...FLOW_B_INITIAL_IDS, ...FLOW_B_STREAMED_IDS];

test.describe('flow subscriptions', { tag: ['@flows', '@smoke'] }, () => {
    test.use({ cassette: flowsCassette() });

    test('streams messages without duplicates within a flow', async ({ page, pageErrorLog }, testInfo) => {
        await page.goto('/flows');
        await page.getByRole('row', { name: /E2E Alpha/ }).click();

        await expect(page.getByTestId(MESSAGE_ID_TESTID)).toHaveCount(FLOW_A_IDS.length);

        const ids = await extractMessageIds(page.locator('body'));

        assertNoDuplicates(ids);
        expect([...ids].sort()).toEqual([...FLOW_A_IDS].sort());
        await attachIdSet(testInfo, 'flow-5-message-ids', ids);
        expectCleanPage(pageErrorLog);
    });

    test('keeps concurrent flows exact-set disjoint across a round trip', async ({ page, pageErrorLog }, testInfo) => {
        await page.goto('/flows');
        await page.getByRole('row', { name: /E2E Alpha/ }).click();
        await expect(page.getByTestId(MESSAGE_ID_TESTID)).toHaveCount(FLOW_A_IDS.length);

        const idsA = await extractMessageIds(page.locator('body'));

        await page.goBack();
        await page.getByRole('row', { name: /E2E Beta/ }).click();
        await expect(page.getByTestId(MESSAGE_ID_TESTID)).toHaveCount(FLOW_B_IDS.length);

        const idsB = await extractMessageIds(page.locator('body'));

        assertNoDuplicates(idsB);
        assertDisjoint(idsA, idsB);
        await attachIdSet(testInfo, 'flow-5-message-ids', idsA);
        await attachIdSet(testInfo, 'flow-6-message-ids', idsB);

        await page.goBack();
        await page.getByRole('row', { name: /E2E Alpha/ }).click();
        await expect(page.getByTestId(MESSAGE_ID_TESTID)).toHaveCount(FLOW_A_IDS.length);

        const idsAgain = await extractMessageIds(page.locator('body'));

        expect([...idsAgain].sort(), 'round trip must not change the set').toEqual([...idsA].sort());
        expectCleanPage(pageErrorLog);
    });
});
