import { expect, test } from '../../fixtures/test.ts';
import { expectCleanPage } from '../../helpers/errors.ts';
import { entity } from '../../mocks/cassette.ts';
import { apiTokensCassette, makeToken, SEED_TOKEN, tokensList } from '../../mocks/cassettes/api-tokens.ts';
import { KNOWLEDGE_DOC, knowledgesCassette, makeKnowledge } from '../../mocks/cassettes/knowledges.ts';
import { FILE_RESOURCE, RENAMED_PATH, resourcesCassette } from '../../mocks/cassettes/resources.ts';
import { makeTemplate, TEMPLATE_SEED, templatesCassette } from '../../mocks/cassettes/templates.ts';

test.describe('a change made outside the tab', { tag: '@cross' }, () => {
    test.describe('flowTemplateUpdated', () => {
        test.use({
            cassette: templatesCassette({
                subscriptions: {
                    flowTemplateUpdated: [
                        {
                            frames: [
                                {
                                    delayMs: 50,
                                    payload: {
                                        data: {
                                            flowTemplateUpdated: makeTemplate(
                                                TEMPLATE_SEED.id,
                                                'Renamed elsewhere',
                                                TEMPLATE_SEED.text,
                                            ),
                                        },
                                    },
                                },
                            ],
                        },
                    ],
                },
            }),
        });

        test('retitles the row', async ({ page, pageErrorLog }) => {
            await page.goto('/templates');

            await expect(page.getByRole('row', { name: /Renamed elsewhere/ })).toBeVisible();
            await expect(page.getByRole('row', { name: /E2E Seed Template/ })).toBeHidden();
            expectCleanPage(pageErrorLog);
        });
    });

    test.describe('knowledgeDocumentCreated', () => {
        test.use({
            cassette: knowledgesCassette({
                subscriptions: {
                    knowledgeDocumentCreated: [
                        {
                            frames: [
                                {
                                    delayMs: 50,
                                    payload: {
                                        data: {
                                            knowledgeDocumentCreated: makeKnowledge('808', 'Added elsewhere'),
                                        },
                                    },
                                },
                            ],
                        },
                    ],
                },
            }),
        });

        test('adds the row', async ({ page, pageErrorLog }) => {
            await page.goto('/knowledges');

            await expect(page.getByRole('row', { name: /E2E Seed Question/ })).toBeVisible();
            await expect(page.getByRole('row', { name: /Added elsewhere/ })).toBeVisible();
            expectCleanPage(pageErrorLog);
        });
    });

    test.describe('knowledgeDocumentUpdated', () => {
        test.use({
            cassette: knowledgesCassette({
                subscriptions: {
                    knowledgeDocumentUpdated: [
                        {
                            frames: [
                                {
                                    delayMs: 50,
                                    payload: {
                                        data: {
                                            knowledgeDocumentUpdated: makeKnowledge(
                                                KNOWLEDGE_DOC.id,
                                                'Rephrased elsewhere',
                                            ),
                                        },
                                    },
                                },
                            ],
                        },
                    ],
                },
            }),
        });

        test('rewrites the row', async ({ page, pageErrorLog }) => {
            await page.goto('/knowledges');

            await expect(page.getByRole('row', { name: /Rephrased elsewhere/ })).toBeVisible();
            expectCleanPage(pageErrorLog);
        });
    });

    const renamedResource = entity('UserResource', {
        createdAt: '2026-01-15T09:00:00Z',
        id: String(FILE_RESOURCE.id),
        isDir: false,
        name: RENAMED_PATH,
        path: RENAMED_PATH,
        size: FILE_RESOURCE.size,
        updatedAt: '2026-01-15T12:00:00Z',
        userId: '1',
    });

    test.describe('resourceUpdated', () => {
        test.use({
            cassette: resourcesCassette({
                subscriptions: {
                    resourceUpdated: [
                        { frames: [{ delayMs: 50, payload: { data: { resourceUpdated: renamedResource } } }] },
                    ],
                },
            }),
        });

        test('renames the entry', async ({ page, pageErrorLog }) => {
            await page.goto('/resources');

            await expect(page.getByRole('treeitem', { name: new RegExp(RENAMED_PATH) })).toBeVisible();
            expectCleanPage(pageErrorLog);
        });
    });

    test.describe('resourceDeleted', () => {
        test.use({
            cassette: resourcesCassette({
                subscriptions: {
                    resourceDeleted: [
                        {
                            frames: [
                                {
                                    delayMs: 50,
                                    payload: { data: { resourceDeleted: { ...renamedResource, name: 'notes.txt' } } },
                                },
                            ],
                        },
                    ],
                },
            }),
        });

        test('drops the entry', async ({ page, pageErrorLog }) => {
            await page.goto('/resources');

            await expect(page.getByRole('treeitem', { name: /reports/ })).toBeVisible();
            await expect(page.getByRole('treeitem', { name: /notes\.txt/ })).toBeHidden();
            expectCleanPage(pageErrorLog);
        });
    });
});

test.describe('a token changed outside the tab', { tag: '@cross' }, () => {
    test.describe('apiTokenCreated', () => {
        test.use({
            cassette: apiTokensCassette({
                queries: {
                    apiTokens: [
                        { data: tokensList(SEED_TOKEN) },
                        { data: tokensList(SEED_TOKEN, makeToken('9', 'Minted elsewhere')) },
                    ],
                },
                subscriptions: {
                    apiTokenCreated: [
                        {
                            frames: [
                                {
                                    delayMs: 50,
                                    payload: { data: { apiTokenCreated: makeToken('9', 'Minted elsewhere') } },
                                },
                            ],
                        },
                    ],
                },
            }),
        });

        test('joins the list', async ({ page, pageErrorLog }) => {
            await page.goto('/settings/api-tokens');

            await expect(page.getByRole('row', { name: /E2E seed token/ })).toBeVisible();
            await expect(page.getByRole('row', { name: /Minted elsewhere/ })).toBeVisible();
            expectCleanPage(pageErrorLog);
        });
    });

    test.describe('apiTokenUpdated', () => {
        test.use({
            cassette: apiTokensCassette({
                queries: {
                    apiTokens: [
                        { data: tokensList(SEED_TOKEN) },
                        { data: tokensList(makeToken(SEED_TOKEN.id, 'Renamed elsewhere')) },
                    ],
                },
                subscriptions: {
                    apiTokenUpdated: [
                        {
                            frames: [
                                {
                                    delayMs: 50,
                                    payload: {
                                        data: { apiTokenUpdated: makeToken(SEED_TOKEN.id, 'Renamed elsewhere') },
                                    },
                                },
                            ],
                        },
                    ],
                },
            }),
        });

        test('renames the row in place', async ({ page, pageErrorLog }) => {
            await page.goto('/settings/api-tokens');

            await expect(page.getByRole('row', { name: /Renamed elsewhere/ })).toBeVisible();
            await expect(page.getByRole('row', { name: /E2E seed token/ })).toBeHidden();
            expectCleanPage(pageErrorLog);
        });
    });

    const KEPT_TOKEN = makeToken('11', 'E2E kept token');

    test.describe('apiTokenDeleted', () => {
        test.use({
            cassette: apiTokensCassette({
                queries: {
                    apiTokens: [{ data: tokensList(SEED_TOKEN, KEPT_TOKEN) }, { data: tokensList(KEPT_TOKEN) }],
                },
                subscriptions: {
                    apiTokenDeleted: [
                        {
                            frames: [{ delayMs: 50, payload: { data: { apiTokenDeleted: SEED_TOKEN } } }],
                        },
                    ],
                },
            }),
        });

        test('drops the row', async ({ page, pageErrorLog }) => {
            await page.goto('/settings/api-tokens');

            // A row that survives the delete is the landmark: until the list has rendered, the page
            // shows a loading state, and the absence of a row that never existed means nothing.
            await expect(page.getByRole('row', { name: /E2E kept token/ })).toBeVisible();
            await expect(page.getByRole('row', { name: /E2E seed token/ })).toBeHidden();
            expectCleanPage(pageErrorLog);
        });
    });
});
