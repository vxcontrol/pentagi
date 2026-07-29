import type { Page } from '@playwright/test';

import { expect, test } from '../../fixtures/test.ts';
import { expectCleanPage } from '../../helpers/errors.ts';
import {
    agentTestResult,
    OTHER_PROVIDER,
    OTHER_PROVIDER_ROW,
    populatedSettingsProvidersCassette,
    providersList,
    providerTestResult,
    renamedProviderRow,
    SEEDED_PROVIDER,
    seededProviderRow,
    settingsProvidersCassette,
} from '../../mocks/cassettes/settings-providers.ts';

/** The per-agent Test control repeats its name for all 13 agents, so it is addressed through the
 *  accordion item that names the agent. */
const agentRow = (page: Page, agentName: string) =>
    page.locator('[data-slot="accordion-item"]').filter({ hasText: agentName });

test.describe('settings providers', { tag: '@coverage' }, () => {
    test.use({ cassette: settingsProvidersCassette() });

    test('shows the empty state when no providers are configured', async ({ page, pageErrorLog }) => {
        await page.goto('/settings/providers');

        await expect(page.getByText('No providers configured')).toBeVisible();
        await expect(page.getByRole('button', { name: 'Add Provider' })).toBeVisible();

        expectCleanPage(pageErrorLog);
    });

    test('opens the create form from the empty state', async ({ page, pageErrorLog }) => {
        await page.goto('/settings/providers');
        await page.getByRole('button', { name: 'Add Provider' }).click();

        await expect(page).toHaveURL(/\/settings\/providers\/new$/);
        await expect(page.getByRole('heading', { name: 'Create a new provider' })).toBeVisible();
        await expect(page.getByText('Type', { exact: true })).toBeVisible();
        await expect(page.getByRole('button', { name: 'Select provider' })).toBeVisible();
        await expect(page.getByLabel('Name', { exact: true })).toBeVisible();

        expectCleanPage(pageErrorLog);
    });
});

test.describe('settings providers create form', { tag: '@coverage' }, () => {
    // Needs a cassette where the type is enabled: the create form bounces `?type=` for a disabled
    // one, so the empty seed would just redirect to the list.
    test.use({ cassette: populatedSettingsProvidersCassette() });

    // A selected type seeds each agent's model from that type's catalog. The catalog must carry the
    // shape the backend actually emits (a non-empty model list for every type, keyed or not); an
    // empty catalog makes this whole path unreachable and any assertion on it vacuous. `?type=` is
    // the URL the type picker produces, driven directly to avoid the dirty-form navigation blocker.
    test('seeds the agent model from the selected type catalog', async ({ page, pageErrorLog }) => {
        await page.goto('/settings/providers/new?type=anthropic');

        await page.getByRole('button', { name: /^Adviser/ }).click();
        await page.getByRole('button', { name: 'Open model list' }).first().click();

        // A catalog-only model (not the seeded agent default), so it can only appear if the type's
        // model list reached the dropdown — the path an empty catalog leaves unreachable.
        await expect(page.getByRole('option', { name: 'e2e-anthropic-model-mini' })).toBeVisible();

        expectCleanPage(pageErrorLog);
    });
});

test.describe('settings providers write path', { tag: '@coverage' }, () => {
    const CREATED_NAME = 'E2E created provider';

    test.use({
        cassette: populatedSettingsProvidersCassette({
            mutations: {
                createProvider: [
                    {
                        data: { createProvider: { __typename: 'ResultType', result: 'success' } } as never,
                        variables: { name: CREATED_NAME, type: 'anthropic' },
                    },
                ],
            },
        }),
    });

    // The create form is seeded from the type's defaults and then serialises the whole agents map back.
    // Fields the user never touches are exactly the ones that get dropped on the way to the wire — this
    // asserts the payload, because the UI looks identical either way.
    test('create sends every agent field it was seeded with, zero-ish values included', async ({
        page,
        pageErrorLog,
    }) => {
        await page.goto('/settings/providers/new?type=anthropic');

        await expect(page.getByLabel('Name', { exact: true })).toBeVisible();
        await page.getByLabel('Name', { exact: true }).fill(CREATED_NAME);

        const request = page.waitForRequest(
            (candidate) =>
                candidate.method() === 'POST' && candidate.postDataJSON()?.operationName === 'createProvider',
        );

        await page.getByRole('button', { exact: true, name: 'Create' }).click();

        const { variables } = (await request).postDataJSON();

        expect(variables.name).toBe(CREATED_NAME);
        expect(variables.type).toBe('anthropic');
        expect(Object.keys(variables.agents), 'every agent is serialised, not just the edited one').toHaveLength(13);
        expect(variables.agents.simpleJson).toMatchObject({
            json: true,
            maxTokens: 4096,
            model: 'e2e-anthropic-model',
            n: 1,
        });
        expectCleanPage(pageErrorLog);
    });
});

test.describe('settings provider edit paths', { tag: '@coverage' }, () => {
    const RENAMED = 'E2E renamed endpoint';

    const editCassette = (override: Parameters<typeof populatedSettingsProvidersCassette>[0] = {}) =>
        populatedSettingsProvidersCassette(override);

    const openProvider = async (page: Page) => {
        await page.goto(`/settings/providers/${SEEDED_PROVIDER.id}`);
        await expect(page.getByLabel('Name', { exact: true })).toHaveValue(SEEDED_PROVIDER.name);
    };

    const mutationRequest = (page: Page, operationName: string) =>
        page.waitForRequest(
            (candidate) => candidate.method() === 'POST' && candidate.postDataJSON()?.operationName === operationName,
        );

    test.describe('update', () => {
        test.use({
            cassette: editCassette({
                mutations: {
                    updateProvider: [
                        {
                            data: { updateProvider: renamedProviderRow(RENAMED) } as never,
                            variables: { name: RENAMED, providerId: SEEDED_PROVIDER.id },
                        },
                    ],
                },
            }),
        });

        test('Save carries the id of the provider being edited', async ({ page, pageErrorLog }) => {
            await openProvider(page);
            await page.getByLabel('Name', { exact: true }).fill(RENAMED);

            const request = mutationRequest(page, 'updateProvider');

            await page.getByRole('button', { exact: true, name: 'Save' }).click();

            const { variables } = (await request).postDataJSON();

            expect(variables.providerId).toBe(SEEDED_PROVIDER.id);
            expect(variables.name).toBe(RENAMED);
            expect(Object.keys(variables.agents), 'the whole agents map is resubmitted').toHaveLength(13);
            await expect(page, 'a saved provider returns to the list').toHaveURL(/\/settings\/providers$/);
            expectCleanPage(pageErrorLog);
        });
    });

    test.describe('delete from the editor', () => {
        test.use({
            cassette: editCassette({
                mutations: {
                    deleteProvider: [
                        {
                            data: { deleteProvider: 'success' } as never,
                            variables: { providerId: SEEDED_PROVIDER.id },
                        },
                    ],
                },
            }),
        });

        test('Delete confirms by name, sends the id, and returns to the list', async ({ page, pageErrorLog }) => {
            await openProvider(page);
            await page.getByRole('button', { name: 'Provider actions' }).click();
            await page.getByRole('menuitem', { name: 'Delete' }).click();

            const dialog = page.getByRole('dialog');

            // A confirm dialog wired to the wrong row shows the wrong name here instead of deleting silently.
            await expect(dialog).toContainText(SEEDED_PROVIDER.name);

            const request = mutationRequest(page, 'deleteProvider');

            await dialog.getByRole('button', { name: 'Delete' }).click();

            expect((await request).postDataJSON().variables).toEqual({ providerId: SEEDED_PROVIDER.id });
            await expect(page).toHaveURL(/\/settings\/providers$/);
            expectCleanPage(pageErrorLog);
        });
    });

    test.describe('delete from the list', () => {
        test.use({
            cassette: editCassette({
                mutations: {
                    deleteProvider: [
                        {
                            data: { deleteProvider: 'success' } as never,
                            setFlag: 'provider-deleted',
                            variables: { providerId: OTHER_PROVIDER.id },
                        },
                    ],
                },
                queries: {
                    settingsProviders: [
                        { data: providersList(seededProviderRow(), OTHER_PROVIDER_ROW()) },
                        { data: providersList(seededProviderRow()), whenFlag: 'provider-deleted' },
                    ],
                },
            }),
        });

        test('the row menu deletes the row it belongs to', async ({ page, pageErrorLog }) => {
            await page.goto('/settings/providers');

            // The second row on purpose: a menu wired to a fixed index would still send the first
            // row's id, and a spec that operated row one could not tell the two apart.
            const row = page.getByRole('row', { name: new RegExp(OTHER_PROVIDER.name) });

            await expect(row).toBeVisible();
            await row.hover();
            await row.getByRole('button', { name: 'Open menu' }).click();
            await page.getByRole('menuitem', { name: 'Delete' }).click();

            const request = mutationRequest(page, 'deleteProvider');

            await page.getByRole('dialog').getByRole('button', { name: 'Delete' }).click();

            expect((await request).postDataJSON().variables).toEqual({ providerId: OTHER_PROVIDER.id });
            // The row goes only because the refetch answers without it, and the sibling proves the
            // table itself survived — a vanished table would satisfy a bare toBeHidden() too.
            await expect(row).toBeHidden();
            await expect(page.getByRole('row', { name: new RegExp(SEEDED_PROVIDER.name) })).toBeVisible();
            expectCleanPage(pageErrorLog);
        });
    });

    test.describe('test one agent', () => {
        test.use({
            cassette: editCassette({
                mutations: {
                    testAgent: [
                        {
                            data: { testAgent: agentTestResult() } as never,
                            variables: { agentType: 'adviser', type: SEEDED_PROVIDER.type },
                        },
                    ],
                },
            }),
        });

        test("a single agent's Test sends that agent, its type, and shows the verdict", async ({
            page,
            pageErrorLog,
        }) => {
            await openProvider(page);

            const request = mutationRequest(page, 'testAgent');

            await agentRow(page, 'Adviser').getByRole('button', { exact: true, name: 'Test' }).click();

            const { variables } = (await request).postDataJSON();

            expect(variables.agentType, 'the enum the backend keys the agent by').toBe('adviser');
            expect(variables.type).toBe(SEEDED_PROVIDER.type);
            expect(variables.agent, 'the edited agent config travels with the request').toMatchObject({
                json: true,
                maxTokens: 4096,
                model: 'e2e-model',
                n: 1,
            });

            const dialog = page.getByRole('dialog');

            // Only the agent that was tested appears — the dialog is keyed by the clicked agent, so a
            // handler that tested the wrong one would name it here.
            await expect(dialog.getByRole('heading', { name: 'Adviser 1/1 passed' })).toBeVisible();
            await expect(dialog.getByRole('heading', { name: /passed/ })).toHaveCount(1);
            expectCleanPage(pageErrorLog);
        });
    });

    test.describe('test the whole provider', () => {
        test.use({
            cassette: editCassette({
                mutations: {
                    testProvider: [
                        {
                            data: { testProvider: providerTestResult() } as never,
                            variables: { type: SEEDED_PROVIDER.type },
                        },
                    ],
                },
            }),
        });

        test('the header Test sends every agent and renders a verdict per agent', async ({ page, pageErrorLog }) => {
            await openProvider(page);

            const request = mutationRequest(page, 'testProvider');

            await page.locator('header').getByRole('button', { exact: true, name: 'Test' }).click();

            const { variables } = (await request).postDataJSON();

            expect(Object.keys(variables.agents), 'the whole map is tested, not one agent').toHaveLength(13);
            expect(variables.type).toBe(SEEDED_PROVIDER.type);

            const dialog = page.getByRole('dialog');

            await expect(dialog.getByRole('heading', { name: /passed/ }), 'one row per agent').toHaveCount(13);
            await expect(dialog.getByRole('heading', { name: 'Pentester 0/1 passed' })).toBeVisible();

            await dialog.getByRole('button', { name: 'Pentester 0/1 passed' }).click();

            // The failing agent's own error must reach the operator; a dialog that renders only the
            // verdict leaves them with nothing to act on.
            await expect(dialog.getByText('e2e simulated failure')).toBeVisible();
            expectCleanPage(pageErrorLog);
        });
    });
});
