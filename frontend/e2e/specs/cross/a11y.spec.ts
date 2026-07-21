import { expect, test } from '../../fixtures/test.ts';
import { scanA11y } from '../../helpers/a11y.ts';
import { loginJourneyCassette } from '../../mocks/cassettes/smoke.ts';
import { ROUTE_MANIFEST } from '../../routes.ts';

const THEMES = ['light', 'dark'] as const;

for (const theme of THEMES) {
    test.describe(`a11y (${theme})`, { tag: '@cross' }, () => {
        if (theme === 'dark') {
            test.beforeEach(async ({ page }) => {
                await page.addInitScript(() => window.localStorage.setItem('theme', 'dark'));
            });
        }

        test.describe('login', () => {
            test.use({ cassette: loginJourneyCassette, isAuthSeeded: false });

            test('login page', async ({ page }) => {
                await page.goto('/login');
                await expect(page.getByRole('button', { name: 'Sign in' })).toBeVisible();
                await scanA11y(page, '/login');
            });
        });

        for (const entry of ROUTE_MANIFEST) {
            test.describe(entry.path, () => {
                test.use({ cassette: entry.cassette() });

                test('has no axe violations', async ({ page }) => {
                    await page.goto(entry.path);
                    await expect(entry.ready(page)).toBeVisible();
                    // Not redundant: a theme seed that silently stops applying reruns light under a dark label.
                    await expect(page.locator('html')).toHaveClass(theme === 'dark' ? /dark/ : /light/);
                    await scanA11y(page, entry.path, entry.a11yWaivers);
                });
            });
        }
    });
}
