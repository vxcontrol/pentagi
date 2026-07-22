import type { Page } from '@playwright/test';

import { expect, test } from '../../fixtures/test.ts';
import { expectCleanPage } from '../../helpers/errors.ts';
import { RICH_TEMPLATE_TEXT, TEMPLATE_DETAIL, templateDetailCassette } from '../../mocks/cassettes/templates.ts';

// The other route whose editor loads server content: the templates list spec only exercises
// create mode, so nothing covered the load path in the production bundle this tier runs.
test.describe('template detail', { tag: '@coverage' }, () => {
    test.use({ cassette: templateDetailCassette() });

    const EDITOR = 'Template content';

    // The raw/rich switch lives inside the actions menu, which stays open on select.
    const switchToRaw = async (page: Page) => {
        await page.getByRole('button', { name: 'Template actions' }).click();
        await page.getByLabel('Raw source').click();
        await page.keyboard.press('Escape');
    };

    // This route has no visual baseline at all, so the header order is asserted nowhere else.
    test('keeps the pager between save and the actions menu', async ({ page }) => {
        await page.goto(`/templates/${TEMPLATE_DETAIL.id}`);
        await expect(page.getByRole('button', { name: 'Template actions' })).toBeVisible();

        const labels = await page
            .locator('header button')
            .evaluateAll((buttons) => buttons.map((button) => button.getAttribute('aria-label') ?? button.textContent));
        const positionOf = (label: string) => labels.findIndex((candidate) => (candidate ?? '').includes(label));

        expect(positionOf('Save')).toBeLessThan(positionOf('Previous'));
        expect(positionOf('Next')).toBeLessThan(positionOf('Template actions'));
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
});
