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
        // Type a name distinct from the dialog's default so the request body proves the
        // typed value reached it — a value equal to the default would match even if the
        // input→payload binding were broken.
        const TYPED_PATH = 'e2e-typed-folder';
        const added: ResultOf<typeof ResourceAddedDocument> = {
            resourceAdded: { ...CREATED_FOLDER, name: TYPED_PATH, path: TYPED_PATH },
        };

        test.use({
            cassette: resourcesCassette({
                rest: {
                    'POST /api/v1/resources/mkdir': [
                        {
                            body: { data: {}, status: 'success' },
                            bodySubset: { path: TYPED_PATH },
                            setFlag: 'folder-created',
                        },
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
            await dialog.getByLabel('Path').fill(TYPED_PATH);
            await dialog.getByRole('button', { name: 'Create' }).click();

            await expect(page.getByText('Directory created')).toBeVisible();
            await expect(page.getByRole('treeitem', { name: new RegExp(TYPED_PATH) })).toBeVisible();
            expectCleanPage(pageErrorLog);
        });
    });

    test.describe('empty state', () => {
        test.use({ cassette: emptyResourcesCassette() });

        test('shows the upload call to action', async ({ page, pageErrorLog }) => {
            await page.goto('/resources');

            await expect(page.getByText('No resources yet')).toBeVisible();
            // The CTA, not just the title — the empty state's whole point is the upload affordance.
            // Scope to the drop zone (its hint is unique) so the toolbar's Upload button is excluded.
            const dropZone = page.locator('div').filter({ hasText: 'Up to 300 MB per file' }).last();

            await expect(dropZone.getByRole('button', { name: 'Upload files' })).toBeVisible();
            expectCleanPage(pageErrorLog);
        });
    });
});
