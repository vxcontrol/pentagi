import type { Page } from '@playwright/test';

import { expect, test } from '../../fixtures/test.ts';
import { expectCleanPage } from '../../helpers/errors.ts';
import {
    PROMPT_DETAIL_AGENT,
    PROMPT_OVERRIDE,
    PROMPT_RESET_FLAG,
    promptDetailCassette,
    promptOverrideCassette,
    RICH_PROMPT_TEMPLATE,
} from '../../mocks/cassettes/settings-prompts.ts';

test.describe('settings prompt detail', { tag: '@coverage' }, () => {
    test.use({
        cassette: promptDetailCassette({
            mutations: {
                createPrompt: [
                    {
                        data: {
                            createPrompt: {
                                __typename: 'UserPrompt',
                                createdAt: '2026-01-15T12:00:00Z',
                                id: '77',
                                template: 'saved',
                                type: 'pentester',
                                updatedAt: '2026-01-15T12:00:00Z',
                            },
                        } as never,
                    },
                ],
            },
        }),
    });

    const EDITOR = 'System prompt template';

    // The raw/rich switch lives inside the actions menu, which stays open on select.
    const switchToRaw = async (page: Page) => {
        await page.getByRole('button', { name: 'Prompt actions' }).click();
        await page.getByLabel('Raw source').click();
        await page.keyboard.press('Escape');
    };

    test('loads the template into the editor byte-exact', async ({ page, pageErrorLog }) => {
        await page.goto(`/settings/prompts/${PROMPT_DETAIL_AGENT}`);

        const editor = page.getByRole('textbox', { name: EDITOR });

        await expect(editor).toBeVisible();
        await expect(editor.getByRole('heading', { name: 'Pentester' })).toBeVisible();
        await expect(editor.getByText('nmap -sV {{.Target}}')).toBeVisible();

        await switchToRaw(page);

        await expect(page.getByRole('textbox', { name: EDITOR })).toHaveValue(RICH_PROMPT_TEMPLATE);
        expectCleanPage(pageErrorLog);
    });

    test('carries the template atoms through an edit in the rich editor', async ({ page, pageErrorLog }) => {
        await page.goto(`/settings/prompts/${PROMPT_DETAIL_AGENT}`);

        const editor = page.getByRole('textbox', { name: EDITOR });

        await expect(editor.getByRole('heading', { name: 'Pentester' })).toBeVisible();

        await editor.click();
        await page.keyboard.press('ControlOrMeta+End');
        await editor.pressSequentially(' E2E-MARK');

        await switchToRaw(page);

        const raw = await page.getByRole('textbox', { name: EDITOR }).inputValue();

        for (const atom of ['Pentester', '{{.Target}}', '{{.Scope}}', 'nmap -sV', 'stay inside the agreed scope']) {
            expect(raw, `"${atom}" survived the editor round-trip`).toContain(atom);
        }

        expect(raw).toContain('E2E-MARK');
        expectCleanPage(pageErrorLog);
    });

    // The editor is where a Go template's byte fidelity is decided, and Save is the only operation
    // that can destroy one. Until now no spec let it reach the wire, so the whole form → mutation hop
    // was unverified: the round-trip units cannot see it.
    test('Save sends the edited template verbatim, atoms intact', async ({ page, pageErrorLog }) => {
        await page.goto(`/settings/prompts/${PROMPT_DETAIL_AGENT}`);

        const editor = page.getByRole('textbox', { name: EDITOR });

        await expect(editor.getByRole('heading', { name: 'Pentester' })).toBeVisible();

        await editor.click();
        await page.keyboard.press('ControlOrMeta+End');
        await editor.pressSequentially(' E2E-SAVE-MARK');

        const request = page.waitForRequest(
            (candidate) => candidate.method() === 'POST' && candidate.postDataJSON()?.operationName === 'createPrompt',
        );

        await page.getByRole('button', { exact: true, name: 'Save' }).click();

        const { variables } = (await request).postDataJSON();

        expect(variables.template, 'the edit reached the wire').toContain('E2E-SAVE-MARK');

        for (const atom of ['# Pentester', '{{.Target}}', '{{.Scope}}', '```bash', '| Scope | {{.Scope}} |']) {
            expect(variables.template, `"${atom}" survived the save`).toContain(atom);
        }

        expectCleanPage(pageErrorLog);
    });
});

test.describe('settings prompt detail, already overridden', { tag: '@coverage' }, () => {
    const EDITOR = 'System prompt template';

    test.describe('save', () => {
        test.use({
            cassette: promptOverrideCassette({
                mutations: {
                    updatePrompt: [
                        {
                            data: {
                                updatePrompt: {
                                    __typename: 'UserPrompt',
                                    createdAt: '2026-01-15T09:00:00Z',
                                    id: PROMPT_OVERRIDE.id,
                                    template: 'saved',
                                    type: 'pentester',
                                    updatedAt: '2026-01-15T12:00:00Z',
                                },
                            } as never,
                            variables: { promptId: PROMPT_OVERRIDE.id },
                        },
                    ],
                },
            }),
        });

        test('Save updates the existing override instead of creating another', async ({ page, pageErrorLog }) => {
            await page.goto(`/settings/prompts/${PROMPT_DETAIL_AGENT}`);

            const editor = page.getByRole('textbox', { name: EDITOR });

            await expect(editor.getByRole('heading', { name: 'Operator override' })).toBeVisible();

            await editor.click();
            await page.keyboard.press('ControlOrMeta+End');
            await editor.pressSequentially(' E2E-UPDATE-MARK');

            const request = page.waitForRequest(
                (candidate) =>
                    candidate.method() === 'POST' && candidate.postDataJSON()?.operationName === 'updatePrompt',
            );

            await page.getByRole('button', { exact: true, name: 'Save' }).click();

            const { variables } = (await request).postDataJSON();

            expect(variables.promptId, "the operator's own row is the one rewritten").toBe(PROMPT_OVERRIDE.id);
            expect(variables.template).toContain('E2E-UPDATE-MARK');
            expect(variables.template, 'the override travels whole').toContain('# Operator override');
            expectCleanPage(pageErrorLog);
        });
    });

    test.describe('reset', () => {
        test.use({
            cassette: promptOverrideCassette({
                mutations: {
                    deletePrompt: [
                        {
                            data: { deletePrompt: 'success' } as never,
                            setFlag: PROMPT_RESET_FLAG,
                            variables: { promptId: PROMPT_OVERRIDE.id },
                        },
                    ],
                },
            }),
        });

        test('Reset deletes the override and puts the default back in the editor', async ({ page, pageErrorLog }) => {
            await page.goto(`/settings/prompts/${PROMPT_DETAIL_AGENT}`);

            const editor = page.getByRole('textbox', { name: EDITOR });

            await expect(editor.getByRole('heading', { name: 'Operator override' })).toBeVisible();

            await page.getByRole('button', { name: 'Prompt actions' }).click();
            await page.getByRole('menuitem', { name: 'Reset' }).click();

            const request = page.waitForRequest(
                (candidate) =>
                    candidate.method() === 'POST' && candidate.postDataJSON()?.operationName === 'deletePrompt',
            );

            await page.getByRole('dialog').getByRole('button', { exact: true, name: 'Reset' }).click();

            expect((await request).postDataJSON().variables).toEqual({ promptId: PROMPT_OVERRIDE.id });

            // Wait for the dialog to go first: a role-based locator does not resolve while the modal
            // holds the background out of the a11y tree, and Playwright reads "0 elements" as hidden —
            // so asserting through the open dialog passes whatever the editor actually shows.
            await expect(page.getByRole('dialog')).toBeHidden();

            // Deleting the row is only half the operation: the editor has to fall back to the default,
            // or the operator is left looking at text that no longer exists anywhere.
            await expect(editor.getByRole('heading', { name: 'Pentester' })).toBeVisible();
            await expect(editor.getByRole('heading', { name: 'Operator override' })).toBeHidden();
            expectCleanPage(pageErrorLog);
        });
    });
});
