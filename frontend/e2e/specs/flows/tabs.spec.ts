import { expect, test } from '../../fixtures/test.ts';
import { expectCleanPage } from '../../helpers/errors.ts';
import { flowTabsCassette, TABS_SCREENSHOT_NAME } from '../../mocks/cassettes/flows.ts';
import { FLOW_DETAIL_TABS } from '../../routes.ts';

test.describe('flow detail tabs', { tag: '@flows' }, () => {
    test.use({ cassette: flowTabsCassette() });

    test('each tab renders its populated content', async ({ page, pageErrorLog }) => {
        await page.goto('/flows/5');
        await expect(page.locator('header').getByRole('button', { name: 'Toggle favorite' })).toBeEnabled();

        for (const { name, ready } of FLOW_DETAIL_TABS) {
            // Pins that the click is what reveals the panel. Ordering this sweep so a tab is already
            // open when its turn comes (the flow auto-opens Assistant) makes its iteration a no-op.
            await expect(ready(page), `"${name}" must not be on screen before its own click`).not.toBeAttached();
            await page.getByRole('tab', { name }).click();
            await expect(ready(page)).toBeVisible();
        }

        expectCleanPage(pageErrorLog);
    });

    test('the screenshot image decodes from its REST endpoint', async ({ page, pageErrorLog }) => {
        await page.goto('/flows/5');
        await expect(page.locator('header').getByRole('button', { name: 'Toggle favorite' })).toBeEnabled();
        await page.getByRole('tab', { name: 'Screenshots' }).click();

        const image = page.getByRole('img', { name: TABS_SCREENSHOT_NAME });
        await image.scrollIntoViewIfNeeded();

        // toBeVisible passes on an undecoded element; only the width proves the blob arrived.
        await expect
            .poll(async () => image.evaluate((element: HTMLImageElement) => element.naturalWidth))
            .toBeGreaterThan(0);

        expectCleanPage(pageErrorLog);
    });
});
