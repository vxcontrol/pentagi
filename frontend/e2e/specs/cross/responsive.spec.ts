import { devices } from '@playwright/test';

import { expect, test } from '../../fixtures/test.ts';
import { expectCleanPage } from '../../helpers/errors.ts';
import { flowsCassette } from '../../mocks/cassettes/flows.ts';

const hasHorizontalOverflow = () => document.documentElement.scrollWidth > document.documentElement.clientWidth + 1;

const { defaultBrowserType: _defaultBrowserType, ...pixel7 } = devices['Pixel 7'];

test.describe('responsive', { tag: '@cross' }, () => {
    test.use({ cassette: flowsCassette() });

    test.describe('mobile (real touch device profile)', () => {
        test.use(pixel7);

        test('collapses to the mobile shell without horizontal overflow', async ({ page, pageErrorLog }) => {
            await page.goto('/flows');

            await expect(page.getByRole('button', { name: 'Toggle Sidebar' })).toBeVisible();
            expect(await page.evaluate(hasHorizontalOverflow)).toBe(false);

            await page.getByRole('row', { name: /E2E Alpha/ }).click();
            await expect(page.getByRole('button', { name: 'Flow actions' })).toBeVisible();
            expect(await page.evaluate(hasHorizontalOverflow)).toBe(false);
            expectCleanPage(pageErrorLog);
        });
    });

    test.describe('split boundary', () => {
        test('1280 shows the resizable split, 1279 merges into one tab row', async ({ page, pageErrorLog }) => {
            await page.setViewportSize({ height: 800, width: 1280 });
            await page.goto('/flows');
            await page.getByRole('row', { name: /E2E Alpha/ }).click();
            await expect(page.getByRole('separator').first()).toBeVisible();

            await page.setViewportSize({ height: 800, width: 1279 });
            await expect(page.getByRole('separator')).toBeHidden();
            expect(await page.evaluate(hasHorizontalOverflow)).toBe(false);
            expectCleanPage(pageErrorLog);
        });
    });
});
