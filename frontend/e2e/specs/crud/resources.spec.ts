import type { ResultOf } from '@graphql-typed-document-node/core';
import type { Page } from '@playwright/test';

import type { ResourceAddedDocument } from '@/graphql/types';

import { expect, test } from '../../fixtures/test.ts';
import { clipboardWrites, recordClipboardWrites } from '../../helpers/clipboard.ts';
import { expectCleanPage } from '../../helpers/errors.ts';
import {
    COPY_DESTINATION,
    CREATED_FOLDER,
    emptyResourcesCassette,
    FILE_RESOURCE,
    FOLDER_RESOURCE,
    NESTED_FOLDER,
    NESTED_RESOURCE,
    nestedResourcesCassette,
    RENAMED_PATH,
    resourcesCassette,
    resourceWrites,
} from '../../mocks/cassettes/resources.ts';

interface DownloadClick {
    download: string;
    href: string;
}

const BULK_SELECTION = [FILE_RESOURCE.path, NESTED_RESOURCE.path];

// A nested entry on purpose: for a top-level row the name and the path are the same string, and
// every payload that should carry the path would read the same with the name in its place.
const selectBoth = async (page: Page) => {
    await page.getByRole('button', { name: 'Expand all' }).click();
    await page.getByRole('checkbox', { name: `Select ${NESTED_RESOURCE.name}` }).click();
    await page.getByRole('checkbox', { name: `Select ${FILE_RESOURCE.name}` }).click();
};

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
            const dropZone = page.locator('[data-slot="file-drop-zone"]');

            await expect(dropZone.getByRole('button', { name: 'Upload files' })).toBeVisible();
            expectCleanPage(pageErrorLog);
        });
    });

    // Every write here is a different verb on a path that also serves the SPA: a wrong method answers
    // 200 with HTML instead of 404, so these tests pin the method and the payload, not just the toast.
    test.describe('bulk verbs', () => {
        test.use({
            cassette: nestedResourcesCassette({
                rest: {
                    'DELETE /api/v1/resources/': [
                        {
                            body: { data: {}, status: 'success' },
                            querySubset: { 'paths[]': BULK_SELECTION },
                        },
                    ],
                    'POST /api/v1/resources/copy': [
                        {
                            body: { data: {}, status: 'success' },
                            bodySubset: { destination: 'backup', force: false, sources: BULK_SELECTION },
                        },
                    ],
                    'PUT /api/v1/resources/move': [
                        {
                            body: { data: {}, status: 'success' },
                            bodySubset: { destination: 'archive', force: false, sources: BULK_SELECTION },
                        },
                    ],
                },
            }),
        });

        const BULK_MOVE_DIRECTORY = 'archive';
        const BULK_COPY_DIRECTORY = 'backup';

        test('move sends one destination directory for the whole selection', async ({ page, pageErrorLog }) => {
            await page.goto('/resources');
            await selectBoth(page);
            await page.getByRole('button', { exact: true, name: 'Move to…' }).click();

            const dialog = page.getByRole('dialog');

            await expect(dialog.getByRole('heading', { name: 'Move 2 items' })).toBeVisible();
            await dialog.getByLabel('Destination directory').fill(BULK_MOVE_DIRECTORY);

            const request = page.waitForRequest(
                (candidate) =>
                    candidate.method() === 'PUT' && new URL(candidate.url()).pathname === '/api/v1/resources/move',
            );

            await dialog.getByRole('button', { exact: true, name: 'Move' }).click();

            expect((await request).postDataJSON()).toMatchObject({
                destination: BULK_MOVE_DIRECTORY,
                force: false,
                sources: BULK_SELECTION,
            });
            await expect(page.getByText(`Moved 2 items into /${BULK_MOVE_DIRECTORY}`)).toBeVisible();
            expectCleanPage(pageErrorLog);
        });

        test('copy sends the same selection to its own destination', async ({ page, pageErrorLog }) => {
            await page.goto('/resources');
            await selectBoth(page);
            await page.getByRole('button', { name: 'More actions' }).click();
            await page.getByRole('menuitem', { name: 'Copy to…' }).click();

            const dialog = page.getByRole('dialog');

            await expect(dialog.getByRole('heading', { name: 'Copy 2 items' })).toBeVisible();
            await dialog.getByLabel('Destination directory').fill(BULK_COPY_DIRECTORY);

            const request = page.waitForRequest(
                (candidate) =>
                    candidate.method() === 'POST' && new URL(candidate.url()).pathname === '/api/v1/resources/copy',
            );

            await dialog.getByRole('button', { exact: true, name: 'Copy' }).click();

            expect((await request).postDataJSON()).toMatchObject({
                destination: BULK_COPY_DIRECTORY,
                force: false,
                sources: BULK_SELECTION,
            });
            await expect(page.getByText(`Copied 2 items into /${BULK_COPY_DIRECTORY}`)).toBeVisible();
            expectCleanPage(pageErrorLog);
        });

        test('delete sends every checked path in one request', async ({ page, pageErrorLog }) => {
            await page.goto('/resources');
            await selectBoth(page);

            const request = page.waitForRequest(
                (candidate) =>
                    candidate.method() === 'DELETE' && new URL(candidate.url()).pathname === '/api/v1/resources/',
            );

            await page.getByRole('button', { name: 'Delete' }).click();
            await page.getByRole('dialog').getByRole('button', { exact: true, name: 'Delete' }).click();

            const sent = new URL((await request).url()).searchParams.getAll('paths[]');

            expect(sent).toEqual(BULK_SELECTION);
            expectCleanPage(pageErrorLog);
        });
    });

    test.describe('drag onto a directory', () => {
        test.use({
            cassette: nestedResourcesCassette({
                rest: {
                    'PUT /api/v1/resources/move': [
                        {
                            body: { data: {}, status: 'success' },
                            bodySubset: {
                                destination: NESTED_FOLDER.path,
                                force: false,
                                sources: [NESTED_RESOURCE.path],
                            },
                        },
                    ],
                },
            }),
        });

        test('moves the dragged row into the folder it was dropped on', async ({ page, pageErrorLog }) => {
            await page.goto('/resources');
            await page.getByRole('button', { name: 'Expand all' }).click();

            const request = page.waitForRequest(
                (candidate) =>
                    candidate.method() === 'PUT' && new URL(candidate.url()).pathname === '/api/v1/resources/move',
            );

            await page
                .getByRole('treeitem', { name: new RegExp(NESTED_RESOURCE.name) })
                .dragTo(page.getByRole('treeitem', { name: new RegExp(NESTED_FOLDER.name) }));

            expect((await request).postDataJSON()).toMatchObject({
                destination: NESTED_FOLDER.path,
                force: false,
                sources: [NESTED_RESOURCE.path],
            });
            await expect(page.getByText(`Moved to /${NESTED_FOLDER.path}`)).toBeVisible();
            expectCleanPage(pageErrorLog);
        });
    });

    test.describe('copy paths', () => {
        test.use({ cassette: nestedResourcesCassette() });

        test('puts the row path on the clipboard', async ({ page, pageErrorLog }) => {
            await recordClipboardWrites(page);
            await page.goto('/resources');

            await page.getByRole('button', { name: 'Expand all' }).click();
            await page
                .getByRole('treeitem', { name: new RegExp(NESTED_RESOURCE.name) })
                .getByRole('button', { name: 'Row actions' })
                .click();
            await page.getByRole('menuitem', { name: 'Copy path' }).click();

            await expect(page.getByText('Path copied to clipboard')).toBeVisible();
            expect(await clipboardWrites(page)).toEqual([NESTED_RESOURCE.path]);
            expectCleanPage(pageErrorLog);
        });
    });

    test.describe('write verbs', () => {
        test.use({ cassette: nestedResourcesCassette({ rest: resourceWrites() }) });

        const rowActions = (page: Page, name: string) =>
            page.getByRole('treeitem', { name: new RegExp(name) }).getByRole('button', { name: 'Row actions' });

        const openNestedRowActions = async (page: Page) => {
            await page.getByRole('button', { name: 'Expand all' }).click();
            await rowActions(page, NESTED_RESOURCE.name).click();
        };

        test('rename issues a PUT to /resources/move carrying the typed destination', async ({
            page,
            pageErrorLog,
        }) => {
            await page.goto('/resources');

            const request = page.waitForRequest(
                (candidate) =>
                    candidate.method() === 'PUT' && new URL(candidate.url()).pathname === '/api/v1/resources/move',
            );

            await openNestedRowActions(page);
            await page.getByRole('menuitem', { name: 'Rename or move' }).click();

            const dialog = page.getByRole('dialog');

            await dialog.getByLabel('New path').fill(RENAMED_PATH);
            await dialog.getByRole('button', { exact: true, name: 'Move' }).click();

            expect((await request).postDataJSON()).toMatchObject({
                destination: RENAMED_PATH,
                sources: [NESTED_RESOURCE.path],
            });
            expectCleanPage(pageErrorLog);
        });

        test('copy issues a POST to /resources/copy', async ({ page, pageErrorLog }) => {
            await page.goto('/resources');

            const request = page.waitForRequest(
                (candidate) =>
                    candidate.method() === 'POST' && new URL(candidate.url()).pathname === '/api/v1/resources/copy',
            );

            await openNestedRowActions(page);
            await page.getByRole('menuitem', { name: 'Copy to…' }).click();

            const dialog = page.getByRole('dialog');

            await dialog.getByLabel('Destination path').fill(COPY_DESTINATION);
            await dialog.getByRole('button', { exact: true, name: 'Copy' }).click();

            expect((await request).postDataJSON()).toMatchObject({
                destination: COPY_DESTINATION,
                sources: [NESTED_RESOURCE.path],
            });
            expectCleanPage(pageErrorLog);
        });

        // Copying onto an existing name must stop at a confirmation instead of overwriting silently:
        // the request may not leave until the user takes the second decision.
        test('a colliding destination raises the replace guard before any request', async ({ page }) => {
            await page.goto('/resources');

            let copyRequests = 0;

            page.on('request', (candidate) => {
                if (new URL(candidate.url()).pathname === '/api/v1/resources/copy') {
                    copyRequests += 1;
                }
            });

            await rowActions(page, FILE_RESOURCE.name).click();
            await page.getByRole('menuitem', { name: 'Copy to…' }).click();
            await page.getByRole('dialog').getByLabel('Destination path').fill(FOLDER_RESOURCE.name);
            await page.getByRole('dialog').getByRole('button', { exact: true, name: 'Copy' }).click();

            await expect(page.getByRole('dialog', { name: 'Replace existing item?' })).toBeVisible();
            expect(copyRequests, 'nothing is sent while the guard is open').toBe(0);

            await page.getByRole('button', { name: 'Cancel' }).click();
            expect(copyRequests, 'cancelling the guard sends nothing at all').toBe(0);
        });

        test('delete issues a DELETE whose query carries the row path', async ({ page, pageErrorLog }) => {
            await page.goto('/resources');

            const request = page.waitForRequest(
                (candidate) =>
                    candidate.method() === 'DELETE' && new URL(candidate.url()).pathname === '/api/v1/resources/',
            );

            await openNestedRowActions(page);
            await page.getByRole('menuitem', { name: 'Delete' }).click();
            await page.getByRole('dialog').getByRole('button', { exact: true, name: 'Delete' }).click();

            expect(new URL((await request).url()).searchParams.getAll('paths[]')).toEqual([NESTED_RESOURCE.path]);
            expectCleanPage(pageErrorLog);
        });

        // An `<a download>` transfer is carried out by the browser outside the page, so no Playwright
        // route — page- or context-scoped — is ever offered it: one started here leaves for whatever
        // the preview proxy targets. The bytes are asserted in specs/real/resources-download.spec.ts.
        test('the row download action links at the row, saved under its name', async ({ page, pageErrorLog }) => {
            await page.goto('/resources');
            await openNestedRowActions(page);

            const link = page.getByRole('menuitem', { name: 'Download' });

            await expect(link).toHaveAttribute('download', NESTED_RESOURCE.name);

            const href = new URL((await link.getAttribute('href')) ?? '', page.url());

            expect(href.pathname).toBe('/api/v1/resources/download');
            expect(href.searchParams.getAll('paths[]')).toEqual([NESTED_RESOURCE.path]);
            expectCleanPage(pageErrorLog);
        });

        // Multi-select is a different code path: one transfer carrying every selected path, saved
        // under an archive name. Neither is derivable from the single-row case.
        test('the bulk action asks for every selected path in one archive', async ({ page, pageErrorLog }) => {
            // Its anchor is created, clicked and dropped inside one handler, so nothing survives in
            // the DOM to read — record the click and swallow it instead of starting a transfer.
            await page.addInitScript(() => {
                const recorded: DownloadClick[] = [];

                (window as unknown as { e2eDownloads: DownloadClick[] }).e2eDownloads = recorded;

                const { click } = HTMLAnchorElement.prototype;

                HTMLAnchorElement.prototype.click = function (this: HTMLAnchorElement) {
                    if (!this.hasAttribute('download')) {
                        click.call(this);

                        return;
                    }

                    recorded.push({ download: this.download, href: this.href });
                };
            });
            await page.goto('/resources');
            await selectBoth(page);
            await page.getByRole('checkbox', { name: `Select ${NESTED_FOLDER.name}` }).click();
            await page.getByRole('button', { exact: true, name: 'Download' }).click();

            const clicks = await page.evaluate(
                () => (window as unknown as { e2eDownloads: DownloadClick[] }).e2eDownloads,
            );

            expect(clicks, 'one transfer for the whole selection').toHaveLength(1);

            const { download, href } = clicks[0]!;

            expect(new URL(href).pathname).toBe('/api/v1/resources/download');
            expect(new URL(href).searchParams.getAll('paths[]').sort()).toEqual(
                [...BULK_SELECTION, NESTED_FOLDER.path].sort(),
            );
            expect(download).toMatch(/\.zip$/);
            expectCleanPage(pageErrorLog);
        });
    });
});
