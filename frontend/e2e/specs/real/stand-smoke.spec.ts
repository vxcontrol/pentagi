import { expect, test } from '@playwright/test';

// LLM-independent integration smoke for a live backend (Tier 2 or a real
// stand): the authenticated shell renders and each core route loads without an
// uncaught error. No seeded data and no agent run — safe against a shared stand.
const ROUTES = [
    { path: '/flows', title: 'Flows' },
    { path: '/templates', title: 'Templates' },
    { path: '/knowledges', title: 'Knowledges' },
    { path: '/settings/prompts', title: 'Prompts' },
    { path: '/settings/api-tokens', title: 'API Tokens' },
];

test.describe('stand smoke', { tag: '@stand' }, () => {
    for (const { path, title } of ROUTES) {
        test(`renders ${path} without a page error`, async ({ page }) => {
            const pageErrors: string[] = [];

            page.on('pageerror', (error) => pageErrors.push(String(error)));

            await page.goto(path);
            // Staying on the route (not bounced to /login) proves the reused
            // session authenticated — no dependency on the stand's username.
            await expect(page).toHaveURL((url) => url.pathname === path);
            // The header title is the only anchor every route owns in both
            // layouts — the main and settings sidebars have disjoint links, so
            // any shared sidebar anchor fails on one layout or the other. The
            // span tag matters: NavLink puts aria-current="page" on the active
            // settings-sidebar <a> at runtime, the breadcrumb title is a span.
            await expect(page.locator('span[aria-current="page"]')).toHaveText(title);
            expect(pageErrors, `uncaught errors on ${path}`).toEqual([]);
        });
    }
});
