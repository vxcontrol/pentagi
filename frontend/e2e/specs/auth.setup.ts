import { expect, test as setup } from '@playwright/test';

import { AUTH_STATE_PATH } from '../playwright.config.ts';

const E2E_USER = process.env.E2E_USER ?? 'admin@pentagi.com';
const E2E_PASSWORD = process.env.E2E_PASSWORD ?? 'admin';

setup('authenticate', async ({ page }) => {
    // A real backend login + the websocket teardown can outrun the 30s default.
    setup.setTimeout(90_000);
    await page.goto('/login');
    await page.getByLabel('Login').fill(E2E_USER);
    await page.getByRole('textbox', { name: 'Password' }).fill(E2E_PASSWORD);
    await page.getByRole('button', { name: 'Sign in' }).click();

    // The migration-seeded admin carries password_change_required — the forced
    // change screen is skippable and must not block the suite.
    const skip = page.getByRole('button', { name: 'Skip for now' });
    const loggedIn = page.getByRole('button', { name: /admin@|flows/i });

    await expect(skip.or(loggedIn).first()).toBeVisible({ timeout: 15_000 });

    if (await skip.isVisible()) {
        await skip.click();
    }

    await expect(page).toHaveURL(/\/flows/, { timeout: 15_000 });
    await page.context().storageState({ path: AUTH_STATE_PATH });
});
