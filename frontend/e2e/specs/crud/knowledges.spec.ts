import type { ResultOf } from '@graphql-typed-document-node/core';
import type { Page } from '@playwright/test';

import type {
    CreateKnowledgeDocumentDocument,
    DeleteKnowledgeDocumentDocument,
    KnowledgeDocumentDocument,
} from '@/graphql/types';

import { ResultType } from '@/graphql/types';

import { expect, test } from '../../fixtures/test.ts';
import { typeIntoEditor } from '../../helpers/editor.ts';
import { expectCleanPage } from '../../helpers/errors.ts';
import {
    KNOWLEDGE_DOC,
    knowledgeDetailCassette,
    knowledgesCassette,
    makeKnowledge,
} from '../../mocks/cassettes/knowledges.ts';

const openNewKnowledgeForm = async (page: Page) => {
    await page.goto('/knowledges');
    await page.getByRole('button', { name: 'New Knowledge' }).click();

    await expect(page).toHaveURL(/\/knowledges\/new$/);
    await expect(page.getByRole('combobox', { name: 'Answer type' })).toBeVisible();
};

test.describe('knowledges crud', { tag: '@crud' }, () => {
    test.describe('create validation', () => {
        test.use({ cassette: knowledgesCassette() });

        test('blocks submit until the answer type is picked', async ({ page, pageErrorLog }) => {
            await openNewKnowledgeForm(page);
            await page.getByRole('textbox', { name: 'Question' }).fill('What is the E2E answer?');
            await typeIntoEditor(page, 'Content', 'E2E knowledge content');
            await page.getByRole('button', { name: 'Create' }).click();

            await expect(page.getByText('Answer type is required')).toBeVisible();
            await expect(page).toHaveURL(/\/knowledges\/new$/);
            expectCleanPage(pageErrorLog);
        });
    });

    test.describe('create', () => {
        const CREATED_DOC = makeKnowledge('301', 'What is the E2E answer?');
        const created: ResultOf<typeof CreateKnowledgeDocumentDocument> = { createKnowledgeDocument: CREATED_DOC };
        const detail: ResultOf<typeof KnowledgeDocumentDocument> = { knowledgeDocument: CREATED_DOC };

        test.use({
            cassette: knowledgesCassette({
                mutations: {
                    createKnowledgeDocument: [
                        {
                            data: created,
                            variables: {
                                input: {
                                    answerType: 'other',
                                    content: 'E2E knowledge content',
                                    docType: 'answer',
                                    question: 'What is the E2E answer?',
                                },
                            },
                        },
                    ],
                },
                queries: { knowledgeDocument: [{ data: detail, variables: { id: '301' } }] },
            }),
        });

        test('creates a document and lands on its detail page', async ({ page, pageErrorLog }) => {
            await openNewKnowledgeForm(page);
            await page.getByRole('textbox', { name: 'Question' }).fill('What is the E2E answer?');
            await typeIntoEditor(page, 'Content', 'E2E knowledge content');
            await page.getByRole('combobox', { name: 'Answer type' }).click();
            await page.getByRole('option', { name: 'other' }).click();
            await page.getByRole('button', { name: 'Create' }).click();

            await expect(page).toHaveURL(/\/knowledges\/301$/);
            await expect(page.getByRole('textbox', { name: 'Question' })).toHaveValue('What is the E2E answer?');
            expectCleanPage(pageErrorLog);
        });
    });

    test.describe('delete', () => {
        const deleted: ResultOf<typeof DeleteKnowledgeDocumentDocument> = {
            deleteKnowledgeDocument: ResultType.Success,
        };

        test.use({
            cassette: knowledgesCassette({
                mutations: {
                    deleteKnowledgeDocument: [{ data: deleted, setFlag: 'knowledge-deleted', variables: { id: '7' } }],
                },
                subscriptions: {
                    knowledgeDocumentDeleted: [
                        {
                            frames: [
                                {
                                    payload: { data: { knowledgeDocumentDeleted: KNOWLEDGE_DOC } },
                                    whenFlag: 'knowledge-deleted',
                                },
                            ],
                        },
                    ],
                },
            }),
        });

        test('deletes from the row menu and drops off the list', async ({ page, pageErrorLog }) => {
            await page.goto('/knowledges');

            const row = page.getByRole('row', { name: /E2E Seed Question/ });

            await row.hover();
            await row.getByRole('button', { name: 'Open menu' }).click();
            await page.getByRole('menuitem', { name: 'Delete' }).click();

            const dialog = page.getByRole('dialog');

            await expect(dialog.getByRole('heading', { name: 'Delete knowledge document' })).toBeVisible();
            await dialog.getByRole('button', { name: 'Delete' }).click();

            await expect(page.getByRole('row', { name: /E2E Seed Question/ })).toBeHidden();
            await expect(page.getByText('No knowledge documents yet')).toBeVisible();
            expectCleanPage(pageErrorLog);
        });
    });

    // A real load failure on the detail route used to be mislabelled "not found" and bounced to
    // the list with no way back in; it must keep the user on the route behind Retry instead.
    test.describe('detail load failure', () => {
        test.use({
            cassette: knowledgesCassette({
                queries: { knowledgeDocument: [{ errors: [{ message: 'e2e induced load failure' }] }] },
            }),
        });

        test('shows an in-page error with Retry, not a bounce to the list', async ({ page }) => {
            await page.goto(`/knowledges/${KNOWLEDGE_DOC.id}`);

            await expect(page.getByRole('button', { name: 'Try again' })).toBeVisible();
            await expect(page).toHaveURL(new RegExp(`/knowledges/${KNOWLEDGE_DOC.id}$`));
        });
    });

    // The backend authz denial (graph/context.go) contains "not found", so a naive match would
    // mistake it for a missing record and bounce a user who merely lacks access.
    test.describe('detail authz denial', () => {
        test.use({
            cassette: knowledgesCassette({
                queries: {
                    knowledgeDocument: [{ errors: [{ message: "requested permission 'knowledge.read' not found" }] }],
                },
            }),
        });

        test('does not bounce a permission denial to the list', async ({ page }) => {
            await page.goto(`/knowledges/${KNOWLEDGE_DOC.id}`);

            await expect(page.getByRole('button', { name: 'Try again' })).toBeVisible();
            await expect(page).toHaveURL(new RegExp(`/knowledges/${KNOWLEDGE_DOC.id}$`));
        });
    });

    test.describe('save', () => {
        test.use({
            cassette: knowledgesCassette({
                mutations: {
                    updateKnowledgeDocument: [{ data: { updateKnowledgeDocument: KNOWLEDGE_DOC } as never }],
                },
                queries: {
                    knowledgeDocument: [
                        { data: { knowledgeDocument: KNOWLEDGE_DOC }, variables: { id: KNOWLEDGE_DOC.id } },
                    ],
                },
            }),
        });

        // The write half of the editor: until now the suite proved the document loads and round-trips
        // in memory, never that Save puts the edited bytes on the wire.
        test('Save sends the edited content and keeps the document id', async ({ page, pageErrorLog }) => {
            await page.goto(`/knowledges/${KNOWLEDGE_DOC.id}`);

            await typeIntoEditor(page, 'Content', 'E2E-SAVE-MARK');

            const request = page.waitForRequest(
                (candidate) =>
                    candidate.method() === 'POST' &&
                    candidate.postDataJSON()?.operationName === 'updateKnowledgeDocument',
            );

            await page.getByRole('button', { exact: true, name: 'Save' }).click();

            const { variables } = (await request).postDataJSON();

            expect(variables.id).toBe(KNOWLEDGE_DOC.id);
            expect(variables.input.content).toContain('E2E-SAVE-MARK');
            expectCleanPage(pageErrorLog);
        });
    });

    test.describe('detail header', () => {
        test.use({ cassette: knowledgeDetailCassette() });

        test('keeps the pager left of save and the actions menu', async ({ page }) => {
            await page.goto(`/knowledges/${KNOWLEDGE_DOC.id}`);
            await expect(page.getByRole('button', { name: 'Knowledge actions' })).toBeVisible();

            const labels = await page
                .locator('header button')
                .evaluateAll((buttons) =>
                    buttons.map((button) => button.getAttribute('aria-label') ?? button.textContent),
                );
            const positionOf = (label: string) => labels.findIndex((candidate) => (candidate ?? '').includes(label));

            // findIndex returns -1 for an absent label, and -1 < any real index, so the ordering below
            // passes vacuously when a button is missing. Require presence first.
            for (const label of ['Save', 'Previous', 'Next', 'Knowledge actions']) {
                expect(positionOf(label), `header is missing the "${label}" button`).toBeGreaterThanOrEqual(0);
            }

            expect(positionOf('Previous')).toBeLessThan(positionOf('Save'));
            expect(positionOf('Next')).toBeLessThan(positionOf('Save'));
            expect(positionOf('Save')).toBeLessThan(positionOf('Knowledge actions'));
        });
    });
});
