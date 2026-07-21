import { expect, test } from '../../fixtures/test.ts';
import { expectCleanPage } from '../../helpers/errors.ts';
import { settingsProvidersCassette } from '../../mocks/cassettes/settings-providers.ts';

test.describe('settings providers', { tag: '@coverage' }, () => {
    test.use({ cassette: settingsProvidersCassette() });

    test('shows the empty state when no providers are configured', async ({ page, pageErrorLog }) => {
        await page.goto('/settings/providers');

        await expect(page.getByText('No providers configured')).toBeVisible();
        await expect(page.getByRole('button', { name: 'Add Provider' })).toBeVisible();

        expectCleanPage(pageErrorLog);
    });

    test('opens the create form from the empty state', async ({ page, pageErrorLog }) => {
        await page.goto('/settings/providers');
        await page.getByRole('button', { name: 'Add Provider' }).click();

        await expect(page).toHaveURL(/\/settings\/providers\/new$/);
        await expect(page.getByRole('heading', { name: 'Create a new provider' })).toBeVisible();
        await expect(page.getByText('Type', { exact: true })).toBeVisible();
        await expect(page.getByRole('button', { name: 'Select provider' })).toBeVisible();
        await expect(page.getByLabel('Name', { exact: true })).toBeVisible();

        expectCleanPage(pageErrorLog);
    });
});
