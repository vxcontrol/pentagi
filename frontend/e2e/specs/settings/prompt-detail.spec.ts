import type { Page } from '@playwright/test';

import { expect, test } from '../../fixtures/test.ts';
import { expectCleanPage } from '../../helpers/errors.ts';
import {
    PROMPT_DETAIL_AGENT,
    promptDetailCassette,
    RICH_PROMPT_TEMPLATE,
} from '../../mocks/cassettes/settings-prompts.ts';

test.describe('settings prompt detail', { tag: '@coverage' }, () => {
    test.use({ cassette: promptDetailCassette() });

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
});
