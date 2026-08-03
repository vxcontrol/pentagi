import { expect, test } from '@playwright/test';

// LLM-independent integration smoke for a live backend (Tier 2 or a real
// stand): the authenticated shell renders and each core route loads without an
// uncaught error. No seeded data and no agent run — safe against a shared stand.
const ROUTES = [
    { emptyTitle: 'No flows found', path: '/flows', title: 'Flows' },
    { emptyTitle: 'No templates yet', path: '/templates', title: 'Templates' },
    { emptyTitle: 'No knowledge documents yet', path: '/knowledges', title: 'Knowledges' },
    { emptyTitle: 'No prompts available', path: '/settings/prompts', title: 'Prompts' },
    { emptyTitle: 'No API tokens configured', path: '/settings/api-tokens', title: 'API Tokens' },
];

// The data assertion waits on a real backend query, so it must not run on the 5s default expect
// timeout — a healthy but slow stand would flake. The config overrides only the screenshot timeout.
const STAND_DATA_TIMEOUT = 30_000;

test.describe('stand smoke', { tag: '@stand' }, () => {
    for (const { emptyTitle, path, title } of ROUTES) {
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
            // Either branch: a query error renders neither, and a stand may hold no rows.
            await expect(page.locator('[data-slot="table"]').first().or(page.getByText(emptyTitle))).toBeVisible({
                timeout: STAND_DATA_TIMEOUT,
            });
            expect(pageErrors, `uncaught errors on ${path}`).toEqual([]);
        });
    }
});
