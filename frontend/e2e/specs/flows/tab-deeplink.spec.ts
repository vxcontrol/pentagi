import type { Page } from '@playwright/test';

import { expect, test } from '../../fixtures/test.ts';
import { expectCleanPage } from '../../helpers/errors.ts';
import { FLOW_A, flowQueryData, flowsCassette } from '../../mocks/cassettes/flows.ts';

/** A flow whose message list stays empty, so the auto-detect resolves to Assistant. */
const emptyFlowCassette = () =>
    flowsCassette({
        queries: { flow: [{ data: flowQueryData(FLOW_A, []), variables: { id: '5' } }] },
        subscriptions: { messageLogAdded: [{ frames: [], variables: { flowId: '5' } }] },
    });

const expectSelectedTab = async (page: Page, name: string) => {
    await expect(page.getByRole('tab', { name })).toHaveAttribute('aria-selected', 'true');
};

test.describe('flow tab deep link', { tag: '@flows' }, () => {
    test.describe('over a populated flow', () => {
        test.use({ cassette: flowsCassette() });

        test('?tab=assistant beats the auto-detect that would have shown Automation', async ({
            page,
            pageErrorLog,
        }) => {
            await page.goto('/flows/5?tab=assistant');
            await expectSelectedTab(page, 'Assistant');
            await expect(page.getByText('New assistant', { exact: true })).toBeVisible();

            expectCleanPage(pageErrorLog);
        });
    });

    test.describe('over a flow with no messages', () => {
        test.use({ cassette: emptyFlowCassette() });

        test('?tab=automation beats the auto-detect that would have shown Assistant', async ({
            page,
            pageErrorLog,
        }) => {
            await page.goto('/flows/5?tab=automation');
            await expectSelectedTab(page, 'Automation');
            await expect(page.getByRole('tabpanel', { name: 'Automation' })).toBeVisible();

            expectCleanPage(pageErrorLog);
        });

        test('a tab click records itself in the URL', async ({ page, pageErrorLog }) => {
            await page.goto('/flows/5');
            await expectSelectedTab(page, 'Assistant');

            await page.getByRole('tab', { name: 'Automation' }).click();

            await expect(page).toHaveURL(/[?&]tab=automation/);
            await expectSelectedTab(page, 'Automation');

            expectCleanPage(pageErrorLog);
        });

        test('an unknown tab falls back to the auto-detect', async ({ page, pageErrorLog }) => {
            await page.goto('/flows/5?tab=not-a-tab');
            await expectSelectedTab(page, 'Assistant');
            // The bogus value is left alone rather than rewritten, so a shared link keeps working
            // once the tab it names exists again.
            await expect(page).toHaveURL(/[?&]tab=not-a-tab/);

            expectCleanPage(pageErrorLog);
        });
    });
});
