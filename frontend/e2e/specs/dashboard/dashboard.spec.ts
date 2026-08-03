import type { Page } from '@playwright/test';

import { expect, test } from '../../fixtures/test.ts';
import { expectCleanPage } from '../../helpers/errors.ts';
import { dashboardCassette } from '../../mocks/cassettes/dashboard.ts';

const CHART_TITLES = ['Flows Activity Over Time', 'Tool Calls Over Time', 'Token Usage Over Time', 'Cost Over Time'];

const cardWith = (page: Page, heading: string) =>
    page.locator('[data-slot="card"]', { has: page.getByRole('heading', { exact: true, name: heading }) });

test.describe('dashboard', { tag: '@coverage' }, () => {
    test.describe('happy path', () => {
        test.use({ cassette: dashboardCassette() });

        test('analytics tab renders the four chart cards', async ({ page, pageErrorLog }) => {
            await page.goto('/dashboard');

            for (const title of CHART_TITLES) {
                await expect(page.getByRole('heading', { exact: true, name: title })).toBeVisible();
            }

            await expect(page.getByText('E2E Alpha')).toBeVisible();
            expectCleanPage(pageErrorLog);
        });

        test('overview tab loads lazily and renders metrics and usage tables', async ({ page, pageErrorLog }) => {
            await page.goto('/dashboard');
            // The route is lazy, so this must wait for the Analytics panel first: a bare negative
            // assert is satisfied by "nothing has rendered yet" and can never fail.
            await expect(page.getByRole('heading', { name: 'Flows Activity Over Time' })).toBeVisible();
            // not.toBeAttached, not toBeHidden — the latter also passes for a mounted-but-hidden node,
            // so it would not catch the panel being eagerly mounted (all its queries firing).
            await expect(page.getByRole('heading', { exact: true, name: 'Total Flows' })).not.toBeAttached();
            await page.getByRole('tab', { name: 'Overview' }).click();

            for (const title of ['Total Flows', 'Tool Calls', 'Total Tokens', 'Total Cost']) {
                await expect(page.getByRole('heading', { exact: true, name: title })).toBeVisible();
            }

            await expect(page.getByRole('heading', { exact: true, name: 'Usage by Provider' })).toBeVisible();
            await expect(page.getByRole('heading', { exact: true, name: 'Usage by Model' })).toBeVisible();
            await expect(page.getByRole('cell', { exact: true, name: 'e2e-provider' })).toBeVisible();
            await expect(page.getByRole('cell', { name: 'e2e-model (e2e-provider)' })).toBeVisible();
            expectCleanPage(pageErrorLog);
        });

        test('period switch refetches and renders the month range', async ({ page, pageErrorLog }) => {
            await page.goto('/dashboard');

            await expect(page.getByRole('heading', { exact: true, name: 'Flows Activity Over Time' })).toBeVisible();

            const periodCards = CHART_TITLES.map((title) => cardWith(page, title));

            for (const card of periodCards) {
                await expect(card.getByText('Jan 15', { exact: true })).toBeVisible();
            }

            await expect(page.getByText('E2E Alpha')).toBeVisible();

            const monthRequest = page.waitForRequest((request) => {
                if (request.method() !== 'POST' || !request.url().includes('/api/v1/graphql')) {
                    return false;
                }

                const { operationName, variables } = request.postDataJSON() as {
                    operationName?: string;
                    variables?: { period?: string };
                };

                return operationName === 'usageStatsByPeriod' && variables?.period === 'month';
            });

            await page.getByRole('tab', { name: 'Month' }).click();
            await monthRequest;

            await expect(page.getByRole('tab', { name: 'Month' })).toHaveAttribute('aria-selected', 'true');

            for (const card of periodCards) {
                await expect(card.getByText('Aug 15', { exact: true })).toBeVisible();
                await expect(card.getByText('Jan 15', { exact: true })).toBeHidden();
            }

            // The execution breakdown carries no dates; its month dataset is a different flow.
            await expect(page.getByText('E2E Beta')).toBeVisible();
            await expect(page.getByText('E2E Alpha')).toBeHidden();
            expectCleanPage(pageErrorLog);
        });
    });

    test.describe('partial errors', () => {
        test.use({
            cassette: dashboardCassette({
                queries: {
                    flowsStatsTotal: [{ errors: [{ message: 'e2e induced failure' }] }],
                    usageStatsByPeriod: [{ errors: [{ message: 'e2e induced failure' }] }],
                },
            }),
        });

        test('a failed query degrades only its own card', async ({ page, pageErrorLog }) => {
            await page.goto('/dashboard');

            await expect(cardWith(page, 'Token Usage Over Time').getByText("Couldn't load")).toBeVisible();
            await expect(cardWith(page, 'Flows Activity Over Time').getByText("Couldn't load")).toBeHidden();

            await page.getByRole('tab', { name: 'Overview' }).click();

            const totalFlowsCard = cardWith(page, 'Total Flows');

            await expect(totalFlowsCard.getByText("Couldn't load")).toBeVisible();
            await expect(totalFlowsCard.getByText('—')).toBeVisible();

            const toolCallsCard = cardWith(page, 'Tool Calls');

            await expect(toolCallsCard.getByText('42')).toBeVisible();
            await expect(toolCallsCard.getByText("Couldn't load")).toBeHidden();
            expectCleanPage(pageErrorLog);
        });
    });
});
