import type { Page } from '@playwright/test';

import { expect, test } from '../../fixtures/test.ts';
import { expectCleanPage } from '../../helpers/errors.ts';
import { resourcesCassette, UPLOADED_FRAGMENT, UPLOADED_RESOURCE } from '../../mocks/cassettes/resources.ts';

const UPLOAD_ENDPOINT = '/api/v1/resources/';

const pickFiles = (page: Page, files: Array<{ buffer: Buffer; mimeType: string; name: string }>) =>
    // The sidebar's quick-upload mounts an input of its own; this is the library page's.
    page.locator('main input[name="resource-upload"]').setInputFiles(files);

const textFile = (name: string, body: string) => ({
    buffer: Buffer.from(body),
    mimeType: 'text/plain',
    name,
});

const uploadRequests = (page: Page): string[] => {
    const seen: string[] = [];

    page.on('request', (request) => {
        if (request.method() === 'POST' && new URL(request.url()).pathname === UPLOAD_ENDPOINT) {
            seen.push(request.url());
        }
    });

    return seen;
};

test.describe('resources upload', { tag: '@crud' }, () => {
    test.describe('accepted', () => {
        test.use({
            cassette: resourcesCassette({
                rest: {
                    'POST /api/v1/resources/': [
                        {
                            body: { data: { items: [UPLOADED_RESOURCE], total: 1 }, status: 'success' },
                            setFlag: 'resource-uploaded',
                        },
                    ],
                },
                subscriptions: {
                    resourceAdded: [
                        {
                            frames: [
                                {
                                    payload: { data: { resourceAdded: UPLOADED_FRAGMENT } },
                                    whenFlag: 'resource-uploaded',
                                },
                            ],
                        },
                    ],
                },
            }),
        });

        test('sends the picked file as multipart and shows it in the library', async ({ page, pageErrorLog }) => {
            await page.goto('/resources');
            await expect(page.getByRole('treeitem', { name: /notes\.txt/ })).toBeVisible();

            const posted = page.waitForRequest(
                (request) => request.method() === 'POST' && new URL(request.url()).pathname === UPLOAD_ENDPOINT,
            );

            await pickFiles(page, [textFile(UPLOADED_RESOURCE.name, 'e2e upload payload')]);

            const body = (await posted).postData() ?? '';

            expect((await posted).headers()['content-type']).toMatch(/^multipart\/form-data; boundary=/);
            // The field name is the contract with the Go handler: rename it and the server sees
            // an upload with no files at all.
            expect(body).toContain('name="files"');
            expect(body).toContain(`filename="${UPLOADED_RESOURCE.name}"`);
            expect(body).toContain('e2e upload payload');

            await expect(page.getByText('File uploaded')).toBeVisible();
            await expect(page.getByRole('treeitem', { name: new RegExp(UPLOADED_RESOURCE.name) })).toBeVisible();
            expectCleanPage(pageErrorLog);
        });
    });

    test.describe('rejected before the wire', () => {
        test.use({ cassette: resourcesCassette() });

        test('an empty file never becomes a request', async ({ page, pageErrorLog }) => {
            await page.goto('/resources');
            await expect(page.getByRole('treeitem', { name: /notes\.txt/ })).toBeVisible();

            const attempts = uploadRequests(page);

            await pickFiles(page, [textFile('hollow.txt', '')]);

            await expect(page.getByText('is empty')).toBeVisible();
            expect(attempts, 'the guard runs before the upload is built').toEqual([]);
            expectCleanPage(pageErrorLog);
        });
    });

    test.describe('refused by the backend', () => {
        test.use({
            cassette: resourcesCassette({
                rest: {
                    'POST /api/v1/resources/': [
                        {
                            // No `msg`: the client's own 409 copy is only reached when the server sends none.
                            body: { code: 'Resources.AlreadyExists', status: 'error' },
                            status: 409,
                        },
                    ],
                },
            }),
        });

        test('a name clash is explained instead of failing silently', async ({ page, pageErrorLog }) => {
            await page.goto('/resources');
            await expect(page.getByRole('treeitem', { name: /notes\.txt/ })).toBeVisible();

            await pickFiles(page, [textFile('notes.txt', 'same name, new bytes')]);

            await expect(page.getByText('Upload failed')).toBeVisible();
            await expect(page.getByText(/already exists/i)).toBeVisible();

            expect(pageErrorLog.pageErrors, 'no uncaught page errors').toEqual([]);
            expect(pageErrorLog.consoleErrors.filter((text) => !text.includes('409'))).toEqual([]);
        });
    });
});
