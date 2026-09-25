import { PromptType } from '@/graphql/types';

import { expect, test } from '../../fixtures/test.ts';
import { expectCleanPage } from '../../helpers/errors.ts';
import {
    PROMPT_OVERRIDE,
    PROMPT_RESET_FLAG,
    promptOverrideCassette,
    promptTemplate,
    settingsPromptsCassette,
} from '../../mocks/cassettes/settings-prompts.ts';

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

test.describe('settings prompts, resetting from the list', { tag: '@coverage' }, () => {
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

    test('offers Reset System on the overridden agent and sends the override id', async ({ page, pageErrorLog }) => {
        await page.goto('/settings/prompts');

        const row = page.getByRole('row', { name: /Pentester/ }).first();

        await row.getByRole('button', { name: 'Open menu' }).click();

        const reset = page.getByRole('menuitem', { name: 'Reset System' });

        await expect(reset).toBeVisible();
        await reset.click();

        const request = page.waitForRequest(
            (candidate) => candidate.method() === 'POST' && candidate.postDataJSON()?.operationName === 'deletePrompt',
        );

        await page.getByRole('dialog').getByRole('button', { exact: true, name: 'Reset' }).click();

        expect((await request).postDataJSON().variables.promptId).toBe(PROMPT_OVERRIDE.id);
        expectCleanPage(pageErrorLog);
    });

    test('offers no reset on an agent the operator never overrode', async ({ page }) => {
        await page.goto('/settings/prompts');

        const row = page.getByRole('row', { name: /Generator/ }).first();

        await row.getByRole('button', { name: 'Open menu' }).click();

        await expect(page.getByRole('menuitem', { name: 'Reset System' })).toHaveCount(0);
    });
});
