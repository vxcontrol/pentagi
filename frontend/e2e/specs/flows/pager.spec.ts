import { StatusType } from '@/graphql/types';

import { expect, test } from '../../fixtures/test.ts';
import { expectCleanPage } from '../../helpers/errors.ts';
import { FLOW_A, flowQueryData, flowsCassette, flowTabsCassette, makeFlow } from '../../mocks/cassettes/flows.ts';

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

    test('keeps the header actions where they are while the flow is still loading', async ({ page, pageErrorLog }) => {
        let releaseFlow = () => {};

        const flowHeld = new Promise<void>((resolve) => {
            releaseFlow = resolve;
        });
        await page.route('**/api/v1/graphql', async (route) => {
            if (route.request().postDataJSON()?.operationName === 'flow') {
                await flowHeld;
            }

            await route.fallback();
        });

        const header = page.locator('header');
        const star = header.getByRole('button', { name: 'Toggle favorite' });
        const actionPositions = () =>
            Promise.all(
                ['Previous', 'Next', 'Toggle favorite', 'Flow actions'].map(async (label) =>
                    header.getByRole('button', { name: label }).evaluate((button) => button.getBoundingClientRect().x),
                ),
            );

        await page.goto('/flows/5');
        await expect(star).toBeDisabled();

        const whileLoading = await actionPositions();
        releaseFlow();

        await expect(star).toBeEnabled();
        await expect(header.getByText('E2E Alpha')).toBeVisible();
        expect(await actionPositions()).toEqual(whileLoading);
        expectCleanPage(pageErrorLog);
    });

    test.describe('onto a flow of another user', () => {
        const FOREIGN_FLOW = makeFlow('6', 'E2E Beta', StatusType.Running, '2');

        test.use({
            cassette: flowsCassette({
                queries: {
                    flow: [
                        { data: flowQueryData(FLOW_A, []), variables: { id: '5' } },
                        { data: flowQueryData(FOREIGN_FLOW, []), variables: { id: '6' } },
                    ],
                    flows: [{ data: { flows: [FLOW_A, FOREIGN_FLOW] } }],
                },
            }),
        });

        test('shows no favorite toggle while that flow loads', async ({ page, pageErrorLog }) => {
            const header = page.locator('header');
            const star = header.getByRole('button', { name: 'Toggle favorite' });
            const actionPositions = () =>
                Promise.all(
                    ['Previous', 'Next', 'Flow actions'].map(async (label) =>
                        header
                            .getByRole('button', { name: label })
                            .evaluate((button) => button.getBoundingClientRect().x),
                    ),
                );

            await page.goto('/flows/5');
            await expect(header.getByText('E2E Alpha')).toBeVisible();

            let releaseForeignFlow = () => {};

            const foreignFlowHeld = new Promise<void>((resolve) => {
                releaseForeignFlow = resolve;
            });
            await page.route('**/api/v1/graphql', async (route) => {
                const request = route.request().postDataJSON();

                if (request?.operationName === 'flow' && request?.variables?.id === '6') {
                    await foreignFlowHeld;
                }

                await route.fallback();
            });

            await header.getByRole('button', { name: 'Next' }).click();
            await expect(page).toHaveURL(/\/flows\/6$/);
            await expect(star).toHaveCount(0);

            const whileLoading = await actionPositions();
            releaseForeignFlow();

            await expect(header.getByText('E2E Beta')).toBeVisible();
            await expect(star).toHaveCount(0);
            expect(await actionPositions()).toEqual(whileLoading);
            expectCleanPage(pageErrorLog);
        });
    });
});
