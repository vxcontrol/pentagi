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

test.describe('flow side panel tab', { tag: '@flows' }, () => {
    test.use({ cassette: flowsCassette() });

    test('a link naming a side tab opens it in the two-panel layout', async ({ page, pageErrorLog }) => {
        await page.setViewportSize({ height: 800, width: 1440 });
        await page.goto('/flows/5?tab=screenshots');

        await expectSelectedTab(page, 'Screenshots');
        await expect(page, 'the link is read, not rewritten').toHaveURL(/[?&]tab=screenshots/);
        expectCleanPage(pageErrorLog);
    });

    test('a side pick lives in the address and survives a reload', async ({ page, pageErrorLog }) => {
        await page.setViewportSize({ height: 800, width: 1440 });
        await page.goto('/flows/5?tab=assistant');

        await page.getByRole('tab', { name: 'Searches' }).click();

        await expect(page).toHaveURL(/[?&]side=tools/);
        await expectSelectedTab(page, 'Searches');

        await page.reload();

        await expectSelectedTab(page, 'Searches');
        await expectSelectedTab(page, 'Assistant');
        expectCleanPage(pageErrorLog);
    });

    test('the strip and the address still agree after crossing the two-panel boundary twice', async ({
        page,
        pageErrorLog,
    }) => {
        await page.setViewportSize({ height: 800, width: 1100 });
        await page.goto('/flows/5');

        await page.getByRole('tab', { name: 'Screenshots' }).click();
        await expect(page).toHaveURL(/[?&]tab=screenshots/);

        await page.setViewportSize({ height: 800, width: 1440 });
        await expectSelectedTab(page, 'Screenshots');

        await page.getByRole('tab', { name: 'Assistant' }).click();
        await expect(page).toHaveURL(/[?&]tab=assistant/);
        await expectSelectedTab(page, 'Screenshots');

        await page.setViewportSize({ height: 800, width: 1100 });

        await expectSelectedTab(page, 'Assistant');
        expectCleanPage(pageErrorLog);
    });
});
