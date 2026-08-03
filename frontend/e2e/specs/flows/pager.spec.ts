import { expect, test } from '../../fixtures/test.ts';
import { expectCleanPage } from '../../helpers/errors.ts';
import { flowsCassette, flowTabsCassette } from '../../mocks/cassettes/flows.ts';

test.describe('flow pager', { tag: ['@flows', '@smoke'] }, () => {
    test.use({ cassette: flowsCassette() });

    test.describe('with a report to show', () => {
        test.use({ cassette: flowTabsCassette() });

        test('keeps the variable Report action left of the pager and fixed actions', async ({ page }) => {
            await page.goto('/flows/5');
            await expect(page.locator('header').getByRole('button', { name: 'Report' })).toBeVisible();

            const labels = await page
                .locator('header button')
                .evaluateAll((buttons) => buttons.map((button) => button.getAttribute('aria-label') ?? ''));
            const positionOf = (label: string) => labels.findIndex((candidate) => candidate.startsWith(label));

            // findIndex returns -1 for an absent label, and -1 < any real index, so the ordering below
            // passes vacuously when a button is missing. Require presence first.
            for (const label of ['Report', 'Toggle favorite', 'Previous', 'Next', 'Flow actions']) {
                expect(positionOf(label), `header is missing the "${label}" button`).toBeGreaterThanOrEqual(0);
            }

            expect(positionOf('Report')).toBeLessThan(positionOf('Previous'));
            expect(positionOf('Previous')).toBeLessThan(positionOf('Next'));
            expect(positionOf('Next')).toBeLessThan(positionOf('Toggle favorite'));
            expect(positionOf('Toggle favorite')).toBeLessThan(positionOf('Flow actions'));
        });
    });

    test('steps to the sibling flow and back without passing through the list', async ({ page, pageErrorLog }) => {
        const header = page.locator('header');

        await page.addInitScript(() => {
            const trail: string[] = [];

            (window as unknown as { __routeTrail: string[] }).__routeTrail = trail;

            for (const method of ['pushState', 'replaceState'] as const) {
                const original = history[method].bind(history);

                history[method] = (state: unknown, unused: string, url?: null | string | URL) => {
                    original(state, unused, url);
                    trail.push(window.location.pathname);
                };
            }
        });

        await page.goto('/flows/5');
        await expect(header.getByText('E2E Alpha')).toBeVisible();

        // Drop the router's own normalising replaceState from the initial load.
        await page.evaluate(() => {
            (window as unknown as { __routeTrail: string[] }).__routeTrail.length = 0;
        });

        await header.getByRole('button', { name: 'Next' }).click();

        await expect(page).toHaveURL(/\/flows\/6$/);
        await expect(header.getByText('E2E Beta')).toBeVisible();
        await expect(header.getByRole('button', { name: 'Next' })).toBeVisible();

        await header.getByRole('button', { name: 'Previous' }).click();

        await expect(page).toHaveURL(/\/flows\/5$/);
        await expect(header.getByText('E2E Alpha')).toBeVisible();

        const trail = await page.evaluate(() => (window as unknown as { __routeTrail: string[] }).__routeTrail);

        expect(trail, 'the pager must step straight between siblings').toEqual(['/flows/6', '/flows/5']);
        expectCleanPage(pageErrorLog);
    });
});
