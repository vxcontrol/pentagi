import type { ResultOf } from '@graphql-typed-document-node/core';

import { differenceInSeconds } from 'date-fns';

import type { CreateApiTokenDocument, DeleteApiTokenDocument, UpdateApiTokenDocument } from '@/graphql/types';

import { TokenStatus } from '@/graphql/types';

import { CASSETTE_EPOCH } from '../../fixtures/backend.ts';
import { expect, test } from '../../fixtures/test.ts';
import { clipboardWrites, recordClipboardWrites } from '../../helpers/clipboard.ts';
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

        // The picker's date and the pinned clock together fix what the form must put on the wire.
        const EXPIRES_AT = new Date('2026-01-20T00:00:00Z');
        const EXPECTED_TTL = differenceInSeconds(EXPIRES_AT, CASSETTE_EPOCH);

        test('creates a token through the inline row and reveals the secret', async ({ page, pageErrorLog }) => {
            await recordClipboardWrites(page);
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

            // The cassette cannot pin a clock-derived value, so the expiry the operator picked is
            // asserted on the request itself: without this the mutation matches with any ttl at all,
            // including the 60s floor a broken calculateTTL would send.
            const createRequest = page.waitForRequest(
                (request) => request.method() === 'POST' && request.postDataJSON()?.operationName === 'createAPIToken',
            );

            await submit.click();

            const { variables } = (await createRequest).postDataJSON();

            expect(variables.input.ttl).toBeLessThanOrEqual(EXPECTED_TTL);
            expect(variables.input.ttl).toBeGreaterThan(EXPECTED_TTL - 1_000);

            const dialog = page.getByRole('dialog');

            await expect(dialog.getByText('API Token Created')).toBeVisible();
            await expect(dialog.getByText(CREATED_TOKEN_SECRET)).toBeVisible();

            await dialog.getByRole('button', { name: 'Copy Token' }).click();

            await expect(page.getByText('Token copied to clipboard')).toBeVisible();
            expect(await clipboardWrites(page)).toEqual([CREATED_TOKEN_SECRET]);

            await page.keyboard.press('Escape');

            await expect(dialog).toBeHidden();
            await expect(page.getByRole('row', { name: /E2E created token/ })).toBeVisible();
            expectCleanPage(pageErrorLog);
        });
    });

    test.describe('clipboard', () => {
        test.use({ cassette: apiTokensCassette() });

        test('copies the token id from the row menu', async ({ page, pageErrorLog }) => {
            await recordClipboardWrites(page);
            await page.goto('/settings/api-tokens');

            await page
                .getByRole('row', { name: /E2E seed token/ })
                .getByRole('button', { name: 'Open menu' })
                .click();
            await page.getByRole('menuitem', { name: 'Copy Token ID' }).click();

            await expect(page.getByText('Token ID copied to clipboard')).toBeVisible();
            expect(await clipboardWrites(page)).toEqual([SEED_TOKEN.tokenId]);
            expectCleanPage(pageErrorLog);
        });
    });

    test.describe('update', () => {
        const revoked = { ...SEED_TOKEN, name: 'E2E renamed token', status: TokenStatus.Revoked };
        const updated: ResultOf<typeof UpdateApiTokenDocument> = { updateAPIToken: revoked };

        test.use({
            cassette: apiTokensCassette({
                mutations: {
                    updateAPIToken: [
                        {
                            data: updated,
                            setFlag: 'token-updated',
                            variables: {
                                input: { name: 'E2E renamed token', status: TokenStatus.Revoked },
                                tokenId: SEED_TOKEN.tokenId,
                            },
                        },
                    ],
                },
                queries: {
                    apiTokens: [
                        { data: tokensList(SEED_TOKEN) },
                        { data: tokensList(revoked), whenFlag: 'token-updated' },
                    ],
                },
            }),
        });

        test('renames a token and revokes it from the inline edit row', async ({ page, pageErrorLog }) => {
            await page.goto('/settings/api-tokens');

            const row = page.getByRole('row', { name: /E2E seed token/ });

            await row.hover();
            await row.getByRole('button', { name: 'Open menu' }).click();
            await page.getByRole('menuitem', { name: 'Edit' }).click();

            await page.getByPlaceholder('Token name (optional)').fill('E2E renamed token');
            await page.getByRole('combobox', { name: 'Token status' }).click();
            await page.getByRole('option', { name: 'revoked' }).click();
            await page.getByRole('button', { name: 'Submit' }).click();

            await expect(page.getByRole('row', { name: /E2E renamed token/ })).toContainText('revoked');
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
