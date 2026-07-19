import { expect, test } from '../../fixtures/test.ts';
import { expectCleanPage } from '../../helpers/errors.ts';
import { MESSAGE_ID_TESTID } from '../../helpers/subscriptions.ts';
import { flowsCassette } from '../../mocks/cassettes/flows.ts';

test.describe('themes', { tag: '@cross' }, () => {
    test.use({ cassette: flowsCassette() });

    test('renders the flow detail in seeded dark theme', async ({ page, pageErrorLog }) => {
        await page.addInitScript(() => window.localStorage.setItem('theme', 'dark'));
        await page.goto('/flows');
        await page.getByRole('row', { name: /E2E Alpha/ }).click();

        await expect(page.locator('html')).toHaveClass(/dark/);
        await expect(page.getByTestId(MESSAGE_ID_TESTID).first()).toBeVisible();
        await expect(page.locator('.xterm')).toBeVisible();
        expectCleanPage(pageErrorLog);
    });

    test('switches themes from the sidebar user menu', async ({ page, pageErrorLog }) => {
        await page.goto('/flows');

        await expect(page.locator('html')).not.toHaveClass(/dark/);

        await page.getByRole('button', { name: /admin@pentagi\.com/ }).click();
        await page.getByRole('tab', { name: 'Dark theme' }).click();
        await expect(page.locator('html')).toHaveClass(/dark/);

        await page.getByRole('tab', { name: 'Light theme' }).click();
        await expect(page.locator('html')).not.toHaveClass(/dark/);
        expectCleanPage(pageErrorLog);
    });
});
