import { expect, test } from '../../fixtures/test.ts';
import { populatedSettingsProvidersCassette } from '../../mocks/cassettes/settings-providers.ts';

const THEMES = ['light', 'dark'] as const;

test.describe('visual: settings providers (populated)', { tag: '@visual' }, () => {
    for (const theme of THEMES) {
        test.describe(theme, () => {
            test.use({ cassette: populatedSettingsProvidersCassette() });

            test('matches the baseline', async ({ page }) => {
                await page.addInitScript((value) => window.localStorage.setItem('theme', String(value)), theme);
                await page.goto('/settings/providers');
                await expect(page.getByText('My Custom Endpoint')).toBeVisible();
                await page.evaluate(() => document.fonts.ready);

                await expect(page).toHaveScreenshot(`settings-providers-populated-${theme}.png`, {
                    fullPage: true,
                    mask: [page.locator('.xterm')],
                });
            });
        });
    }
});
