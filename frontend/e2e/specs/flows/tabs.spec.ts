import { expect, test } from '../../fixtures/test.ts';
import { expectCleanPage } from '../../helpers/errors.ts';
import {
    flowTabsCassette,
    TABS_FILE_NAME,
    TABS_SCREENSHOT_NAME,
    TABS_SCREENSHOT_URL,
} from '../../mocks/cassettes/flows.ts';

const TABS = [
    { marker: 'E2E Task Alpha', name: 'Tasks' },
    { marker: 'E2E agent reconnaissance', name: 'Agents' },
    { marker: 'E2E search for the CVE', name: 'Searches' },
    { marker: 'E2E recall prior findings', name: 'Vector Store' },
    { marker: TABS_FILE_NAME, name: 'Files' },
    { marker: TABS_SCREENSHOT_URL, name: 'Screenshots' },
] as const;

test.describe('flow detail tabs', { tag: '@flows' }, () => {
    test.use({ cassette: flowTabsCassette() });

    test('each tab renders its populated content', async ({ page, pageErrorLog }) => {
        await page.goto('/flows/5');
        await expect(page.locator('header').getByRole('button', { name: 'Toggle favorite' })).toBeEnabled();

        for (const { marker, name } of TABS) {
            await page.getByRole('tab', { name }).click();
            await expect(page.getByText(marker)).toBeVisible();
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
