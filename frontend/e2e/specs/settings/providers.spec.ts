import { expect, test } from '../../fixtures/test.ts';
import { expectCleanPage } from '../../helpers/errors.ts';
import {
    populatedSettingsProvidersCassette,
    settingsProvidersCassette,
} from '../../mocks/cassettes/settings-providers.ts';

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

test.describe('settings providers create form', { tag: '@coverage' }, () => {
    // Needs a cassette where the type is enabled: the create form bounces `?type=` for a disabled
    // one, so the empty seed would just redirect to the list.
    test.use({ cassette: populatedSettingsProvidersCassette() });

    // A selected type seeds each agent's model from that type's catalog. The catalog must carry the
    // shape the backend actually emits (a non-empty model list for every type, keyed or not); an
    // empty catalog makes this whole path unreachable and any assertion on it vacuous. `?type=` is
    // the URL the type picker produces, driven directly to avoid the dirty-form navigation blocker.
    test('seeds the agent model from the selected type catalog', async ({ page, pageErrorLog }) => {
        await page.goto('/settings/providers/new?type=anthropic');

        await page.getByRole('button', { name: /^Adviser/ }).click();
        await page.getByRole('button', { name: 'Open model list' }).first().click();

        // A catalog-only model (not the seeded agent default), so it can only appear if the type's
        // model list reached the dropdown — the path an empty catalog leaves unreachable.
        await expect(page.getByRole('option', { name: 'e2e-anthropic-model-mini' })).toBeVisible();

        expectCleanPage(pageErrorLog);
    });
});
