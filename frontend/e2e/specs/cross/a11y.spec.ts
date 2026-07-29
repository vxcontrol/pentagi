import { expect, test } from '../../fixtures/test.ts';
import { scanA11y, waiversForScan } from '../../helpers/a11y.ts';
import { expectCleanPage } from '../../helpers/errors.ts';
import { FILE_RESOURCE, resourcesCassette } from '../../mocks/cassettes/resources.ts';
import { populatedSettingsProvidersCassette } from '../../mocks/cassettes/settings-providers.ts';
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
                // Same guard as the manifest scan: a dark seed that stops applying must not
                // silently re-scan the light page under a dark label.
                await expect(page.locator('html')).toHaveClass(theme === 'dark' ? /dark/ : /light/);
                await scanA11y(page, '/login');
            });
        });

        test.describe('settings providers (populated)', () => {
            test.use({ cassette: populatedSettingsProvidersCassette() });

            // The manifest scan sweeps this route empty, so it renders no provider cards.
            test('provider cards have no axe violations', async ({ page }) => {
                await page.goto('/settings/providers');
                await expect(page.getByText('My Custom Endpoint')).toBeVisible();
                await expect(page.locator('html')).toHaveClass(theme === 'dark' ? /dark/ : /light/);
                await scanA11y(page, '/settings/providers (populated)');
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
                    await scanA11y(page, entry.path, waiversForScan(entry.a11yWaivers));
                });

                for (const tab of entry.tabs ?? []) {
                    test(`tab "${tab.name}" has no axe violations`, async ({ page }) => {
                        await page.goto(entry.path);
                        await expect(entry.ready(page)).toBeVisible();
                        await expect(page.locator('html')).toHaveClass(theme === 'dark' ? /dark/ : /light/);
                        await page.getByRole('tab', { name: tab.name }).click();
                        await expect(tab.ready(page)).toBeVisible();
                        await scanA11y(
                            page,
                            `${entry.path} [${tab.name}]`,
                            waiversForScan(entry.a11yWaivers, tab.name),
                        );
                    });
                }
            });
        }
    });
}

test.describe('dialog keyboard contract', { tag: '@cross' }, () => {
    test.use({ cassette: resourcesCassette() });

    // Closing a dialog must hand focus back to whatever opened it. Radix only does that for its own
    // DialogTrigger, and this app opens dialogs from controlled state, so without help focus lands on
    // <body> and a keyboard user starts again from the top of the page.
    test('Escape returns focus to the control that opened the dialog', async ({ page, pageErrorLog }) => {
        await page.goto('/resources');

        const opener = page.getByRole('button', { name: 'New folder' });

        await opener.focus();
        await page.keyboard.press('Enter');
        await expect(page.getByRole('dialog')).toBeVisible();

        await page.keyboard.press('Escape');
        await expect(page.getByRole('dialog')).toBeHidden();

        await expect(opener).toBeFocused();
        expectCleanPage(pageErrorLog);
    });

    // The other dialog family: ConfirmationDialog renders its DialogContent unconditionally, so a
    // focus hook living on the wrapper would capture whatever had focus when the PAGE mounted and
    // restore it on navigation instead. This case is what tells the two apart.
    test('Escape returns focus for a dialog whose content is always mounted', async ({ page, pageErrorLog }) => {
        await page.goto('/resources');
        await page.getByRole('checkbox', { name: `Select ${FILE_RESOURCE.name}` }).click();

        const opener = page.getByRole('button', { exact: true, name: 'Delete' });

        await opener.click();
        await expect(page.getByRole('dialog')).toBeVisible();

        await page.keyboard.press('Escape');
        await expect(page.getByRole('dialog')).toBeHidden();

        await expect(opener).toBeFocused();
        expectCleanPage(pageErrorLog);
    });
});
