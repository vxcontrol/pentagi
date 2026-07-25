import type { Page } from '@playwright/test';

import { expect, test } from '../../fixtures/test.ts';
import { expectCleanPage } from '../../helpers/errors.ts';
import {
    PROMPT_DETAIL_AGENT,
    promptDetailCassette,
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
