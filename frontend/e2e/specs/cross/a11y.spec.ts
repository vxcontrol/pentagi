import type { A11yAllowlist } from '../../helpers/a11y.ts';

import { expect, test } from '../../fixtures/test.ts';
import { scanA11y } from '../../helpers/a11y.ts';
import { apiTokensCassette } from '../../mocks/cassettes/api-tokens.ts';
import { flowsCassette } from '../../mocks/cassettes/flows.ts';
import { knowledgesCassette } from '../../mocks/cassettes/knowledges.ts';
import { loginJourneyCassette } from '../../mocks/cassettes/smoke.ts';
import { templatesCassette } from '../../mocks/cassettes/templates.ts';

const ALLOWLIST: A11yAllowlist = {
    // Message metadata (date + ID) intentionally renders at 50% opacity — a
    // design decision, not a regression; revisit with the design pass. Axe
    // targets are class chains, so the pattern pins the 50%-muted class; any
    // other contrast violation on the page still fails.
    '/flows/:flowId': [{ rule: 'color-contrast', target: /text-muted-foreground\\?\/50/ }],
    // Colored Badge variants use text-{color}-600 with no dark override, so on
    // the dark near-black background the text fails contrast (all seven color
    // variants share this; the knowledges list renders the blue one). Real
    // dark-mode debt, queued for a palette fix — remove when badge dark text
    // shades land.
    '/knowledges': [{ rule: 'color-contrast', target: /border-blue-500/ }],
};

// Both themes are scanned: contrast waivers are theme-independent, but dark mode
// recomputes every color, so a token that passes in light can still fail in dark.
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
                await scanA11y(page, '/login', ALLOWLIST);
            });
        });

        test.describe('flows', () => {
            test.use({ cassette: flowsCassette() });

            test('flows list', async ({ page }) => {
                await page.goto('/flows');
                await expect(page.getByRole('row', { name: /E2E Alpha/ })).toBeVisible();
                await scanA11y(page, '/flows', ALLOWLIST);
            });

            test('flow detail', async ({ page }) => {
                await page.goto('/flows');
                await page.getByRole('row', { name: /E2E Alpha/ }).click();
                await expect(page.getByRole('button', { name: 'Flow actions' })).toBeVisible();
                await scanA11y(page, '/flows/:flowId', ALLOWLIST);
            });
        });

        test.describe('templates', () => {
            test.use({ cassette: templatesCassette() });

            test('templates list', async ({ page }) => {
                await page.goto('/templates');
                await expect(page.getByRole('row', { name: /E2E Seed Template/ })).toBeVisible();
                await scanA11y(page, '/templates', ALLOWLIST);
            });
        });

        test.describe('knowledges', () => {
            test.use({ cassette: knowledgesCassette() });

            test('knowledges list', async ({ page }) => {
                await page.goto('/knowledges');
                await expect(page.getByRole('row', { name: /E2E Seed Question/ })).toBeVisible();
                await scanA11y(page, '/knowledges', ALLOWLIST);
            });
        });

        test.describe('api tokens', () => {
            test.use({ cassette: apiTokensCassette() });

            test('api tokens list', async ({ page }) => {
                await page.goto('/settings/api-tokens');
                await expect(page.getByRole('row', { name: /E2E seed token/ })).toBeVisible();
                await scanA11y(page, '/settings/api-tokens', ALLOWLIST);
            });
        });
    });
}
