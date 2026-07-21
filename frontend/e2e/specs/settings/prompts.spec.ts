import { PromptType } from '@/graphql/types';

import { expect, test } from '../../fixtures/test.ts';
import { expectCleanPage } from '../../helpers/errors.ts';
import { promptTemplate, settingsPromptsCassette } from '../../mocks/cassettes/settings-prompts.ts';

test.describe('settings prompts', { tag: '@coverage' }, () => {
    test.use({ cassette: settingsPromptsCassette() });

    test('renders agent and tool tables and expands an agent row to its templates', async ({ page, pageErrorLog }) => {
        await page.goto('/settings/prompts');

        await expect(page.getByRole('heading', { name: 'Agent Prompts' })).toBeVisible();
        await expect(page.getByRole('heading', { name: 'Tool Prompts' })).toBeVisible();

        const pentesterRow = page.getByRole('row', { name: /Pentester/ });

        await expect(pentesterRow).toBeVisible();
        await expect(page.getByRole('row', { name: /Get Flow Description/ })).toBeVisible();

        await pentesterRow.click();

        await expect(page.getByRole('heading', { name: 'System Prompt' })).toBeVisible();
        await expect(page.getByRole('heading', { name: 'Human Prompt' })).toBeVisible();
        await expect(page.getByText(promptTemplate(PromptType.Pentester))).toBeVisible();
        await expect(page.getByText(promptTemplate(PromptType.QuestionPentester))).toBeVisible();

        expectCleanPage(pageErrorLog);
    });
});
