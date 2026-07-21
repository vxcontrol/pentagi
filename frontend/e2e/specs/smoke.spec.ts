import { SEEDED_USER } from '../fixtures/auth.ts';
import { expect, test } from '../fixtures/test.ts';
import { expectCleanPage } from '../helpers/errors.ts';
import { loginFailCassette, loginJourneyCassette, smokeCassette } from '../mocks/cassettes/smoke.ts';

test.describe('smoke', { tag: '@smoke' }, () => {
    test.describe('unauthenticated', () => {
        test.use({ cassette: loginJourneyCassette, isAuthSeeded: false });

        test('redirects a protected route to /login with returnUrl', async ({ page }) => {
            await page.goto('/flows');

            await expect(page).toHaveURL(/\/login\?returnUrl=%2Fflows/);
            await expect(page.getByRole('button', { name: 'Sign in' })).toBeVisible();
        });

        test('logs in through the form and lands on the flows list', async ({ page, pageErrorLog }) => {
            await page.goto('/flows');
            await page.getByLabel('Login').fill(SEEDED_USER.mail);
            await page.getByRole('textbox', { name: 'Password' }).fill('e2e-password');
            await page.getByRole('button', { name: 'Sign in' }).click();

            await expect(page).toHaveURL(/\/flows$/);
            await expect(page.getByText(SEEDED_USER.mail)).toBeVisible();
            expectCleanPage(pageErrorLog);
        });
    });

    test.describe('rejected login', () => {
        test.use({ cassette: loginFailCassette, isAuthSeeded: false });

        test('surfaces the error and re-disables Sign in until a field changes', async ({ page }) => {
            await page.goto('/login');
            await page.getByLabel('Login').fill(SEEDED_USER.mail);
            await page.getByRole('textbox', { name: 'Password' }).fill('wrong-password');
            await page.getByRole('button', { name: 'Sign in' }).click();

            await expect(page.getByText('Login failed. Please try again.').first()).toBeVisible();
            await expect(page).toHaveURL(/\/login/);
            // react-hook-form leaves Submit disabled after a failed submit until
            // an input changes.
            await expect(page.getByRole('button', { name: 'Sign in' })).toBeDisabled();
        });
    });

    test.describe('authenticated', () => {
        test.use({ cassette: smokeCassette });

        test('renders the flows list with a clean page and no unmatched calls', async ({ page, pageErrorLog }) => {
            await page.goto('/flows');

            await expect(page).toHaveURL(/\/flows$/);
            await expect(page.getByText(SEEDED_USER.mail)).toBeVisible();
            expectCleanPage(pageErrorLog);
        });
    });
});
