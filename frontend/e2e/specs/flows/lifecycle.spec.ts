import type { ResultOf } from '@graphql-typed-document-node/core';

import type {
    AddFavoriteFlowDocument,
    DeleteFavoriteFlowDocument,
    DeleteFlowDocument,
    FinishFlowDocument,
    RenameFlowDocument,
} from '@/graphql/types';

import { ResultType, StatusType } from '@/graphql/types';

import { expect, test } from '../../fixtures/test.ts';
import { expectCleanPage } from '../../helpers/errors.ts';
import { entity } from '../../mocks/cassette.ts';
import { FLOW_A, flowsCassette, makeFlow } from '../../mocks/cassettes/flows.ts';

const openFlowA = async (page: import('@playwright/test').Page) => {
    await page.goto('/flows');
    await page.getByRole('row', { name: /E2E Alpha/ }).click();
    await expect(page.locator('header').getByRole('button', { name: 'Toggle favorite' })).toBeEnabled();
};

test.describe('flow lifecycle', { tag: '@flows' }, () => {
    test.describe('rename', () => {
        const renamed: ResultOf<typeof RenameFlowDocument> = { renameFlow: ResultType.Success };

        test.use({
            cassette: flowsCassette({
                mutations: {
                    renameFlow: [
                        {
                            data: renamed,
                            setFlag: 'flow-renamed',
                            variables: { flowId: '5', title: 'E2E Alpha Renamed' },
                        },
                    ],
                },
                subscriptions: {
                    flowUpdated: [
                        {
                            frames: [
                                {
                                    payload: { data: { flowUpdated: makeFlow('5', 'E2E Alpha Renamed') } },
                                    whenFlag: 'flow-renamed',
                                },
                            ],
                        },
                    ],
                },
            }),
        });

        test('renames from the actions menu and the title persists', async ({ page, pageErrorLog }) => {
            await openFlowA(page);
            await page.getByRole('button', { name: 'Flow actions' }).click();
            await page.getByRole('menuitem', { name: 'Rename' }).click();
            await page.getByPlaceholder('Flow title').fill('E2E Alpha Renamed');
            await page.getByPlaceholder('Flow title').press('Enter');

            await expect(page.getByText('Flow renamed successfully')).toBeVisible();
            await expect(page.locator('header').getByText('E2E Alpha Renamed', { exact: true })).toBeVisible();
            expectCleanPage(pageErrorLog);
        });
    });

    test.describe('finish', () => {
        const finished: ResultOf<typeof FinishFlowDocument> = { finishFlow: ResultType.Success };

        test.use({
            cassette: flowsCassette({
                mutations: { finishFlow: [{ data: finished, setFlag: 'flow-finished', variables: { flowId: '5' } }] },
                subscriptions: {
                    flowUpdated: [
                        {
                            frames: [
                                {
                                    payload: {
                                        data: { flowUpdated: makeFlow('5', 'E2E Alpha', StatusType.Finished) },
                                    },
                                    whenFlag: 'flow-finished',
                                },
                            ],
                        },
                    ],
                },
            }),
        });

        test('finishes a running flow from the actions menu', async ({ page, pageErrorLog }) => {
            await openFlowA(page);
            await page.getByRole('button', { name: 'Flow actions' }).click();
            await page.getByRole('menuitem', { name: 'Finish' }).click();

            await expect(page.getByText('Flow finished successfully')).toBeVisible();

            await page.getByRole('button', { name: 'Flow actions' }).click();

            // Anchor on an item that stays: absence alone also holds while the
            // reopened menu is still rendering.
            await expect(page.getByRole('menuitem', { name: 'Rename' })).toBeVisible();
            await expect(page.getByRole('menuitem', { name: 'Finish' })).toBeHidden();
            expectCleanPage(pageErrorLog);
        });
    });

    test.describe('favorite', () => {
        const favorited: ResultOf<typeof AddFavoriteFlowDocument> = { addFavoriteFlow: ResultType.Success };

        // The star is useOptimistic and reverts once the mutation settles — the
        // persisted state only arrives via the settingsUserUpdated frame. The
        // delay keeps the optimistic flash from satisfying the persisted-state
        // asserts below before the frame lands.
        test.use({
            cassette: flowsCassette({
                mutations: {
                    addFavoriteFlow: [{ data: favorited, setFlag: 'flow-favorited', variables: { flowId: '5' } }],
                },
                subscriptions: {
                    settingsUserUpdated: [
                        {
                            frames: [
                                {
                                    delayMs: 400,
                                    payload: {
                                        data: {
                                            settingsUserUpdated: entity('UserPreferences', {
                                                favoriteFlows: ['5'],
                                                id: '1',
                                            }),
                                        },
                                    },
                                    whenFlag: 'flow-favorited',
                                },
                            ],
                        },
                    ],
                },
            }),
        });

        test('toggles the favorite star and the state persists', async ({ page, pageErrorLog }) => {
            await openFlowA(page);

            const star = page.locator('header').getByRole('button', { name: 'Toggle favorite' });

            await expect(star).toHaveAttribute('aria-pressed', 'false');
            await star.click();
            await expect(star).toHaveAttribute('aria-pressed', 'true');
            await expect(page.getByText('Favorite Flows')).toBeVisible();
            await expect(star).toHaveAttribute('aria-pressed', 'true');
            expectCleanPage(pageErrorLog);
        });
    });

    test.describe('unfavorite', () => {
        const unfavorited: ResultOf<typeof DeleteFavoriteFlowDocument> = { deleteFavoriteFlow: ResultType.Success };

        test.use({
            cassette: flowsCassette({
                mutations: {
                    deleteFavoriteFlow: [
                        { data: unfavorited, setFlag: 'flow-unfavorited', variables: { flowId: '5' } },
                    ],
                },
                queries: {
                    settingsUser: [
                        {
                            data: {
                                settingsUser: entity('UserPreferences', { favoriteFlows: ['5'], id: '1' }),
                            },
                        },
                    ],
                },
                subscriptions: {
                    settingsUserUpdated: [
                        {
                            frames: [
                                {
                                    delayMs: 400,
                                    payload: {
                                        data: {
                                            settingsUserUpdated: entity('UserPreferences', {
                                                favoriteFlows: [],
                                                id: '1',
                                            }),
                                        },
                                    },
                                    whenFlag: 'flow-unfavorited',
                                },
                            ],
                        },
                    ],
                },
            }),
        });

        test('clears the star, and the group it was listed under goes with it', async ({ page, pageErrorLog }) => {
            await openFlowA(page);

            const star = page.locator('header').getByRole('button', { name: 'Toggle favorite' });

            await expect(star).toHaveAttribute('aria-pressed', 'true');
            await expect(page.getByText('Favorite Flows')).toBeVisible();

            await star.click();

            await expect(star).toHaveAttribute('aria-pressed', 'false');
            await expect(page.getByText('Favorite Flows')).toBeHidden();
            expectCleanPage(pageErrorLog);
        });
    });

    test.describe('a flow owned by someone else', () => {
        test.use({
            cassette: flowsCassette({
                queries: { flows: [{ data: { flows: [FLOW_A, makeFlow('6', 'E2E Beta', StatusType.Running, '2')] } }] },
            }),
        });

        test('offers no favorite star, because adding one is refused', async ({ page, pageErrorLog }) => {
            await page.goto('/flows');

            const mine = page.getByRole('row', { name: /E2E Alpha/ });
            const theirs = page.getByRole('row', { name: /E2E Beta/ });

            await expect(theirs).toBeVisible();
            await mine.hover();
            await expect(mine.getByRole('button', { name: 'Toggle favorite' })).toBeVisible();

            await theirs.hover();
            await expect(theirs.getByRole('button', { name: 'Toggle favorite' })).toHaveCount(0);

            await theirs.click({ button: 'right' });
            await expect(page.getByRole('menuitem', { name: 'View' })).toBeVisible();
            await expect(page.getByRole('menuitem', { name: /favorites/ })).toHaveCount(0);
            expectCleanPage(pageErrorLog);
        });
    });

    test.describe('delete', () => {
        const deleted: ResultOf<typeof DeleteFlowDocument> = { deleteFlow: ResultType.Success };

        test.use({
            cassette: flowsCassette({
                mutations: { deleteFlow: [{ data: deleted, setFlag: 'flow-deleted', variables: { flowId: '5' } }] },
                subscriptions: {
                    flowDeleted: [
                        {
                            frames: [{ payload: { data: { flowDeleted: FLOW_A } }, whenFlag: 'flow-deleted' }],
                        },
                    ],
                },
            }),
        });

        test('deletes via the destructive confirm and drops off the list', async ({ page, pageErrorLog }) => {
            await openFlowA(page);
            await page.getByRole('button', { name: 'Flow actions' }).click();
            await page.getByRole('menuitem', { name: 'Delete' }).click();

            const dialog = page.getByRole('dialog');

            await expect(dialog.getByText('Delete flow')).toBeVisible();
            // Anchored on the variant's own fill: every shadcn Button carries
            // `border-destructive`/`ring-destructive` in its base for the invalid
            // state, so a bare /destructive/ matches whatever variant is set.
            await expect(dialog.getByRole('button', { name: 'Delete' })).toHaveClass(/(^|\s)bg-destructive(\s|$)/);
            await dialog.getByRole('button', { name: 'Delete' }).click();

            await expect(page.getByText('Flow deleted successfully')).toBeVisible();
            await expect(page).toHaveURL(/\/flows$/);
            await expect(page.getByRole('row', { name: /E2E Beta/ })).toBeVisible();
            await expect(page.getByRole('row', { name: /E2E Alpha/ })).toBeHidden();
            expectCleanPage(pageErrorLog);
        });
    });
});
