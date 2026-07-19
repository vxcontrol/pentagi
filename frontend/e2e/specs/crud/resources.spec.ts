import type { ResultOf } from '@graphql-typed-document-node/core';

import type { ResourceAddedDocument } from '@/graphql/types';

import { expect, test } from '../../fixtures/test.ts';
import { expectCleanPage } from '../../helpers/errors.ts';
import { CREATED_FOLDER, emptyResourcesCassette, resourcesCassette } from '../../mocks/cassettes/resources.ts';

test.describe('resources', { tag: '@coverage' }, () => {
    test.describe('listing', () => {
        test.use({ cassette: resourcesCassette() });

        test('renders seeded entries with sortable column headers', async ({ page, pageErrorLog }) => {
            await page.goto('/resources');

            await expect(page.getByRole('treeitem', { name: /reports/ })).toBeVisible();
            await expect(page.getByRole('treeitem', { name: /notes\.txt/ })).toBeVisible();
            await expect(page.getByRole('button', { name: 'Sort by name (ascending)' })).toBeVisible();
            await expect(page.getByRole('button', { name: 'Sort by size (ascending)' })).toBeVisible();
            await expect(page.getByRole('button', { name: 'Sort by modified date (ascending)' })).toBeVisible();
            expectCleanPage(pageErrorLog);
        });
    });

    test.describe('mkdir', () => {
        const added: ResultOf<typeof ResourceAddedDocument> = { resourceAdded: CREATED_FOLDER };

        test.use({
            cassette: resourcesCassette({
                rest: {
                    'POST /api/v1/resources/mkdir': [
                        { body: { data: {}, status: 'success' }, setFlag: 'folder-created' },
                    ],
                },
                subscriptions: {
                    resourceAdded: [{ frames: [{ payload: { data: added }, whenFlag: 'folder-created' }] }],
                },
            }),
        });

        test('creates a directory and the new row arrives via subscription', async ({ page, pageErrorLog }) => {
            await page.goto('/resources');
            await page.getByRole('button', { name: 'New folder' }).click();

            const dialog = page.getByRole('dialog');

            await expect(dialog.getByRole('heading', { name: 'Create directory' })).toBeVisible();
            await expect(dialog.getByLabel('Path')).toHaveValue('new-folder');
            await dialog.getByRole('button', { name: 'Create' }).click();

            await expect(page.getByText('Directory created')).toBeVisible();
            await expect(page.getByRole('treeitem', { name: /new-folder/ })).toBeVisible();
            expectCleanPage(pageErrorLog);
        });
    });

    test.describe('empty state', () => {
        test.use({ cassette: emptyResourcesCassette() });

        test('shows the upload call to action', async ({ page, pageErrorLog }) => {
            await page.goto('/resources');

            await expect(page.getByText('No resources yet')).toBeVisible();
            expectCleanPage(pageErrorLog);
        });
    });
});
