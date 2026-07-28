import type { Page } from '@playwright/test';

import { expect, test } from '../../fixtures/test.ts';
import { expectCleanPage } from '../../helpers/errors.ts';
import { RICH_TEMPLATE_TEXT, TEMPLATE_DETAIL, templateDetailCassette } from '../../mocks/cassettes/templates.ts';

test.describe('template detail', { tag: '@coverage' }, () => {
    test.use({
        cassette: templateDetailCassette({
            mutations: {
                updateFlowTemplate: [{ data: { updateFlowTemplate: TEMPLATE_DETAIL } as never }],
            },
        }),
    });

    const EDITOR = 'Template content';

    // The raw/rich switch lives inside the actions menu, which stays open on select.
    const switchToRaw = async (page: Page) => {
        await page.getByRole('button', { name: 'Template actions' }).click();
        await page.getByLabel('Raw source').click();
        await page.keyboard.press('Escape');
    };

    test('keeps the pager left of save and the actions menu', async ({ page }) => {
        await page.goto(`/templates/${TEMPLATE_DETAIL.id}`);
        await expect(page.getByRole('button', { name: 'Template actions' })).toBeVisible();

        const labels = await page
            .locator('header button')
            .evaluateAll((buttons) => buttons.map((button) => button.getAttribute('aria-label') ?? button.textContent));
        const positionOf = (label: string) => labels.findIndex((candidate) => (candidate ?? '').includes(label));

        // findIndex returns -1 for an absent label, and -1 < any real index, so the ordering below
        // passes vacuously when a button is missing. Require presence first.
        for (const label of ['Save', 'Previous', 'Next', 'Template actions']) {
            expect(positionOf(label), `header is missing the "${label}" button`).toBeGreaterThanOrEqual(0);
        }

        expect(positionOf('Previous')).toBeLessThan(positionOf('Save'));
        expect(positionOf('Next')).toBeLessThan(positionOf('Save'));
        expect(positionOf('Save')).toBeLessThan(positionOf('Template actions'));
    });

    test('loads the template body into the editor byte-exact', async ({ page, pageErrorLog }) => {
        await page.goto(`/templates/${TEMPLATE_DETAIL.id}`);

        const editor = page.getByRole('textbox', { name: EDITOR });

        await expect(editor).toBeVisible();
        await expect(editor.getByRole('heading', { name: 'Recon' })).toBeVisible();
        await expect(editor.getByText('nmap -sV {{TARGET}}')).toBeVisible();

        await switchToRaw(page);

        await expect(page.getByRole('textbox', { name: EDITOR })).toHaveValue(RICH_TEMPLATE_TEXT);
        expectCleanPage(pageErrorLog);
    });

    test('carries the placeholders through an edit in the rich editor', async ({ page, pageErrorLog }) => {
        await page.goto(`/templates/${TEMPLATE_DETAIL.id}`);

        const editor = page.getByRole('textbox', { name: EDITOR });

        await expect(editor.getByRole('heading', { name: 'Recon' })).toBeVisible();

        await editor.click();
        await page.keyboard.press('ControlOrMeta+End');
        await editor.pressSequentially(' E2E-MARK');

        await switchToRaw(page);

        const raw = await page.getByRole('textbox', { name: EDITOR }).inputValue();

        for (const atom of ['Recon', '{{TARGET}}', '{{SCOPE}}', 'nmap -sV', 'capture evidence for each finding']) {
            expect(raw, `"${atom}" survived the editor round-trip`).toContain(atom);
        }

        expect(raw).toContain('E2E-MARK');
        expectCleanPage(pageErrorLog);
    });

    // A real load failure must offer Retry in place, not the "Template not found" card that a
    // genuine 404 shows — the two used to collapse into the same dead-end.

    // Save is the only operation that can destroy a user's template, and no spec has ever let it reach
    // the wire — the round-trip units stop at the editor boundary.
    test('Save sends the edited body verbatim, placeholders intact', async ({ page, pageErrorLog }) => {
        await page.goto(`/templates/${TEMPLATE_DETAIL.id}`);

        const editor = page.getByRole('textbox', { name: EDITOR });

        await expect(editor.getByRole('heading', { name: 'Recon' })).toBeVisible();
        await editor.click();
        await page.keyboard.press('ControlOrMeta+End');
        await editor.pressSequentially(' E2E-SAVE-MARK');

        const request = page.waitForRequest(
            (candidate) =>
                candidate.method() === 'POST' && candidate.postDataJSON()?.operationName === 'updateFlowTemplate',
        );

        await page.getByRole('button', { exact: true, name: 'Save' }).click();

        const { variables } = (await request).postDataJSON();

        expect(variables.templateId).toBe(TEMPLATE_DETAIL.id);
        expect(variables.input.text).toContain('E2E-SAVE-MARK');

        for (const atom of ['# Recon', '{{TARGET}}', 'nmap -sV']) {
            expect(variables.input.text, `"${atom}" survived the save`).toContain(atom);
        }

        expectCleanPage(pageErrorLog);
    });

    test.describe('load failure', () => {
        test.use({
            cassette: templateDetailCassette({
                queries: { flowTemplate: [{ errors: [{ message: 'e2e induced load failure' }] }] },
            }),
        });

        test('shows an in-page error with Retry, not "not found"', async ({ page }) => {
            await page.goto(`/templates/${TEMPLATE_DETAIL.id}`);

            await expect(page.getByRole('button', { name: 'Try again' })).toBeVisible();
            await expect(page.getByText('Template not found')).toBeHidden();
            await expect(page).toHaveURL(new RegExp(`/templates/${TEMPLATE_DETAIL.id}$`));
        });
    });
});
