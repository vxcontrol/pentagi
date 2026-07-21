import { routes } from '@/lib/routes';

import type { A11yAllowlist } from '../../helpers/a11y.ts';

import { expect, test } from '../../fixtures/test.ts';
import { scanA11y } from '../../helpers/a11y.ts';
import { loginJourneyCassette } from '../../mocks/cassettes/smoke.ts';
import { ROUTE_MANIFEST } from '../../routes.ts';

const ALLOWLIST: A11yAllowlist = {
    // Debt surfaced once this sweep started covering the whole manifest.
    // The period switcher drives Tabs as a segmented control with no TabsContent,
    // so the active trigger's aria-controls names an element that never exists.
    [routes.dashboard]: [{ rule: 'aria-valid-attr-value', target: /radix-.*-trigger-/ }],

    // Message metadata (date + ID) intentionally renders at 50% opacity — a
    // design decision, not a regression; revisit with the design pass. Axe
    // targets are class chains, so the pattern pins the 50%-muted class; any
    // other contrast violation on the page still fails.
    [routes.flow('5')]: [{ rule: 'color-contrast', target: /text-muted-foreground\\?\/50/ }],

    // Same 50%-muted decision as the flow messages above, one step lighter: the
    // per-row size and modified-at metadata renders at 80% and misses AA. The
    // tree's expand toggle and row checkboxes sit under the 24px pointer-target
    // floor — widening them is a density decision for the file manager.
    [routes.resources]: [
        { rule: 'color-contrast', target: /text-muted-foreground\\?\/80/ },
        { rule: 'target-size', target: /\.rounded|aria-label="Select / },
    ],
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

        // Outside the manifest: the only unauthenticated route.
        test.describe('login', () => {
            test.use({ cassette: loginJourneyCassette, isAuthSeeded: false });

            test('login page', async ({ page }) => {
                await page.goto('/login');
                await expect(page.getByRole('button', { name: 'Sign in' })).toBeVisible();
                await scanA11y(page, '/login', ALLOWLIST);
            });
        });

        // Driven off the manifest so a newly swept route is scanned without a
        // second, hand-maintained route list drifting behind it.
        for (const entry of ROUTE_MANIFEST) {
            test.describe(entry.path, () => {
                test.use({ cassette: entry.cassette() });

                test('has no axe violations', async ({ page }) => {
                    await page.goto(entry.path);
                    await expect(entry.ready(page)).toBeVisible();
                    await scanA11y(page, entry.path, ALLOWLIST);
                });
            });
        }
    });
}
