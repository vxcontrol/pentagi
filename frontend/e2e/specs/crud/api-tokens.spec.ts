import type { ResultOf } from '@graphql-typed-document-node/core';

import type { CreateApiTokenDocument, DeleteApiTokenDocument } from '@/graphql/types';

import { expect, test } from '../../fixtures/test.ts';
import { expectCleanPage } from '../../helpers/errors.ts';
import {
    apiTokensCassette,
    CREATED_TOKEN,
    CREATED_TOKEN_SECRET,
    CREATED_TOKEN_WITH_SECRET,
    DOOMED_TOKEN,
    SEED_TOKEN,
    tokensList,
} from '../../mocks/cassettes/api-tokens.ts';

test.describe('api tokens crud', { tag: '@crud' }, () => {
    test.describe('create', () => {
        const created: ResultOf<typeof CreateApiTokenDocument> = { createAPIToken: CREATED_TOKEN_WITH_SECRET };

        test.use({
            cassette: apiTokensCassette({
                mutations: {
                    createAPIToken: [
                        // `ttl` is derived from the clock at submit time and drifts by a second
                        // between runs, so only the operator-entered name is pinned.
                        {
                            data: created,
                            setFlag: 'token-created',
                            variables: { input: { name: 'E2E created token' } },
                        },
                    ],
                },
                queries: {
                    apiTokens: [
                        { data: tokensList(SEED_TOKEN) },
                        { data: tokensList(SEED_TOKEN, CREATED_TOKEN), whenFlag: 'token-created' },
                    ],
                },
            }),
        });

        test('creates a token through the inline row and reveals the secret', async ({ page, pageErrorLog }) => {
            await page.goto('/settings/api-tokens');

            await expect(page.getByRole('row', { name: /E2E seed token/ })).toBeVisible();

            await page.getByRole('button', { name: 'Create Token' }).click();
            await page.getByPlaceholder('Token name (optional)').fill('E2E created token');

            const submit = page.getByRole('button', { name: 'Submit' });

            await expect(submit).toBeDisabled();

            await page.getByRole('button', { name: 'Pick date' }).click();
            await page.getByRole('button', { name: 'Tuesday, January 20th, 2026' }).click();
            await page.keyboard.press('Escape');

            await expect(submit).toBeEnabled();
            await submit.click();

            const dialog = page.getByRole('dialog');

            await expect(dialog.getByText('API Token Created')).toBeVisible();
            await expect(dialog.getByText(CREATED_TOKEN_SECRET)).toBeVisible();

            await page.keyboard.press('Escape');

            await expect(dialog).toBeHidden();
            await expect(page.getByRole('row', { name: /E2E created token/ })).toBeVisible();
            expectCleanPage(pageErrorLog);
        });
    });

    test.describe('delete', () => {
        const deleted: ResultOf<typeof DeleteApiTokenDocument> = { deleteAPIToken: true };

        test.use({
            cassette: apiTokensCassette({
                mutations: {
                    deleteAPIToken: [
                        { data: deleted, setFlag: 'token-deleted', variables: { tokenId: DOOMED_TOKEN.tokenId } },
                    ],
                },
                queries: {
                    apiTokens: [
                        { data: tokensList(SEED_TOKEN, DOOMED_TOKEN) },
                        { data: tokensList(SEED_TOKEN), whenFlag: 'token-deleted' },
                    ],
                },
            }),
        });

        test('deletes a token via the row menu and the destructive confirm', async ({ page, pageErrorLog }) => {
            await page.goto('/settings/api-tokens');

            const doomedRow = page.getByRole('row', { name: /E2E doomed token/ });

            await doomedRow.hover();
            await doomedRow.getByRole('button', { name: 'Open menu' }).click();
            await page.getByRole('menuitem', { name: 'Delete' }).click();

            const dialog = page.getByRole('dialog');

            await expect(dialog.getByText('Delete token')).toBeVisible();
            await dialog.getByRole('button', { name: 'Delete' }).click();

            await expect(doomedRow).toBeHidden();
            await expect(page.getByRole('row', { name: /E2E seed token/ })).toBeVisible();
            expectCleanPage(pageErrorLog);
        });
    });
});
