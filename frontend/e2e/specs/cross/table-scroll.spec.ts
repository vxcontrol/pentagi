import { expect, test } from '../../fixtures/test.ts';
import { expectCleanPage } from '../../helpers/errors.ts';
import { populatedSettingsProvidersCassette } from '../../mocks/cassettes/settings-providers.ts';

const CONTAINER = '[data-slot="table-container"]';

test.describe('scrollable table keyboard contract', { tag: '@cross' }, () => {
    test.use({ cassette: populatedSettingsProvidersCassette() });

    test('a table wider than its card takes a tab stop and scrolls with the arrow keys', async ({
        page,
        pageErrorLog,
    }) => {
        await page.setViewportSize({ height: 800, width: 400 });
        await page.goto('/settings/providers');
        await expect(page.getByText('My Custom Endpoint')).toBeVisible();

        const container = page.locator(CONTAINER).first();
        const overflow = () => container.evaluate((element) => element.scrollWidth - element.clientWidth);

        await expect
            .poll(overflow, { message: 'the fixture has to overflow or the contract is untested' })
            .toBeGreaterThan(0);
        await expect(container).toHaveAttribute('tabindex', '0');

        await container.focus();
        await expect(container).toBeFocused();

        await page.keyboard.press('ArrowRight');

        await expect.poll(() => container.evaluate((element) => element.scrollLeft)).toBeGreaterThan(0);
        expectCleanPage(pageErrorLog);
    });

    test('the same table stays out of the tab order once it fits', async ({ page, pageErrorLog }) => {
        await page.setViewportSize({ height: 800, width: 1600 });
        await page.goto('/settings/providers');
        await expect(page.getByText('My Custom Endpoint')).toBeVisible();

        const container = page.locator(CONTAINER).first();

        await expect.poll(() => container.evaluate((element) => element.scrollWidth - element.clientWidth)).toBe(0);
        await expect(container).not.toHaveAttribute('tabindex');
        expectCleanPage(pageErrorLog);
    });
});
