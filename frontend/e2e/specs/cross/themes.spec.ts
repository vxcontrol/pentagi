import type { Page } from '@playwright/test';

import { expect, test } from '../../fixtures/test.ts';
import { expectCleanPage } from '../../helpers/errors.ts';
import { MESSAGE_ID_SELECTOR } from '../../helpers/subscriptions.ts';
import { flowsCassette } from '../../mocks/cassettes/flows.ts';

const storedTheme = (page: Page) => page.evaluate(() => window.localStorage.getItem('theme'));

test.describe('themes', { tag: '@cross' }, () => {
    test.use({ cassette: flowsCassette() });

    test('renders the flow detail in seeded dark theme', async ({ page, pageErrorLog }) => {
        await page.addInitScript(() => window.localStorage.setItem('theme', 'dark'));
        await page.goto('/flows');
        await page.getByRole('row', { name: /E2E Alpha/ }).click();

        await expect(page.locator('html')).toHaveClass(/dark/);
        await expect(page.locator(MESSAGE_ID_SELECTOR).first()).toBeVisible();
        await expect(page.locator('.xterm')).toBeVisible();
        expectCleanPage(pageErrorLog);
    });

    test('switches themes from the sidebar user menu, and the pick outlives the page', async ({
        page,
        pageErrorLog,
    }) => {
        await page.goto('/flows');

        await expect(page.locator('html')).toHaveClass(/light/);

        await page.getByRole('button', { name: /admin@pentagi\.com/ }).click();
        await page.getByRole('tab', { name: 'Dark theme' }).click();
        await expect(page.locator('html')).toHaveClass(/dark/);
        expect(await storedTheme(page), 'the pick is written, not only applied to the live document').toBe('dark');

        await page.reload();
        await expect(page.locator('html'), 'the reloaded page starts dark').toHaveClass(/dark/);

        await page.getByRole('button', { name: /admin@pentagi\.com/ }).click();
        await page.getByRole('tab', { name: 'Light theme' }).click();
        await expect(page.locator('html')).toHaveClass(/light/);
        expect(await storedTheme(page)).toBe('light');

        await page.reload();
        await expect(page.locator('html'), 'the reloaded page starts light').toHaveClass(/light/);
        expectCleanPage(pageErrorLog);
    });
});
