import { expect, test } from '../../fixtures/test.ts';
import { ROUTE_MANIFEST } from '../../routes.ts';

const THEMES = ['light', 'dark'] as const;

const slugify = (path: string) => path.replaceAll('/', '-').replace(/^-/, '') || 'root';

test.describe('visual', { tag: '@visual' }, () => {
    for (const entry of ROUTE_MANIFEST) {
        for (const theme of THEMES) {
            test.describe(`${entry.path} ${theme}`, () => {
                test.use({ cassette: entry.cassette() });

                test('matches the baseline', async ({ page }) => {
                    await page.addInitScript((value) => window.localStorage.setItem('theme', String(value)), theme);
                    await page.goto(entry.path);
                    await expect(entry.ready(page)).toBeVisible();
                    await page.evaluate(() => document.fonts.ready);

                    await expect(page).toHaveScreenshot(`${slugify(entry.path)}-${theme}.png`, {
                        fullPage: true,
                        // The terminal is a WebGL canvas: SwiftShader pixels differ
                        // between runners, so its region is masked, not compared.
                        mask: [page.locator('.xterm')],
                    });
                });
            });
        }
    }
});
