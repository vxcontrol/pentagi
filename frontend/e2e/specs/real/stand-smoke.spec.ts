import { expect, test } from '@playwright/test';

// LLM-independent integration smoke for a live backend (Tier 2 or a real
// stand): the authenticated shell renders and each core route loads without an
// uncaught error. No seeded data and no agent run — safe against a shared stand.
const ROUTES = ['/flows', '/templates', '/knowledges', '/settings/prompts', '/settings/api-tokens'];

test.describe('stand smoke', { tag: '@stand' }, () => {
    for (const route of ROUTES) {
        test(`renders ${route} without a page error`, async ({ page }) => {
            const pageErrors: string[] = [];

            page.on('pageerror', (error) => pageErrors.push(String(error)));

            await page.goto(route);
            // Staying on the route (not bounced to /login) proves the reused
            // session authenticated — no dependency on the stand's username.
            await expect(page).toHaveURL(new RegExp(route.replace(/\//g, '\\/')));
            await expect(page.getByRole('link', { name: 'Templates' })).toBeVisible();
            expect(pageErrors, `uncaught errors on ${route}`).toEqual([]);
        });
    }
});
