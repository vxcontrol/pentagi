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
    // The user menu is labelled with the signed-in address, so derive the "logged in" signal
    // from E2E_USER rather than a hardcoded admin@ (there is no button named "flows").
    const loggedIn = page.getByRole('button', {
        name: new RegExp(E2E_USER.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'), 'i'),
    });

    await expect(skip.or(loggedIn).first()).toBeVisible({ timeout: 15_000 });

    // Whether the seeded admin still carries password_change_required depends
    // on backend state; the `.or()` wait above has already settled one branch.
    // eslint-disable-next-line playwright/no-conditional-in-test
    if (await skip.isVisible()) {
        await skip.click();
    }

    await expect(page).toHaveURL(/\/flows/, { timeout: 15_000 });
    await page.context().storageState({ path: AUTH_STATE_PATH });
});
