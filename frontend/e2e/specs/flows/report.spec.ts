import type { Page } from '@playwright/test';

import { expect, test } from '../../fixtures/test.ts';
import { expectCleanPage } from '../../helpers/errors.ts';
import { inspectPdf } from '../../helpers/pdf.ts';
import { flowReportCassette } from '../../mocks/cassettes/flows.ts';

const REPORT_FILE = /^report_flow_5_e2e_alpha_2026\d{10}\.md$/;
const REPORT_PDF = /^report_flow_5_e2e_alpha_2026\d{10}\.pdf$/;

const openReportMenu = async (page: Page) => {
    await page.goto('/flows/5');
    await page.getByRole('button', { name: 'Report' }).click();
};

test.describe('flow report', { tag: '@flows' }, () => {
    test.use({ cassette: flowReportCassette() });

    test('the standalone route renders the generated markdown', async ({ page, pageErrorLog }) => {
        await page.goto('/flows/5/report');

        await expect(page.getByRole('heading', { name: /E2E Alpha/ })).toBeVisible();
        await expect(page.getByRole('heading', { level: 3, name: /E2E Task Alpha/ })).toBeVisible();
        await expect(page.getByText('Ports 22, 80 open')).toBeVisible();
        // The contents links are slugs of the rendered headings; a mismatch scrolls nowhere.
        await expect(page.getByRole('link', { name: /E2E Subtask Scan/ })).toHaveAttribute('href', /#.+/);

        expectCleanPage(pageErrorLog);
    });

    test('"Open web view" opens the report in its own tab', async ({ page, pageErrorLog }) => {
        await openReportMenu(page);

        const popup = page.waitForEvent('popup');

        await page.getByRole('menuitem', { name: 'Open web view' }).click();

        const opened = await popup;

        expect(new URL(opened.url()).pathname).toBe('/flows/5/report');
        // This tab is mocked only because the fixture re-arms the route on every page the context
        // opens; without that it talks to the preview proxy and the report never arrives.
        await expect(opened.getByRole('heading', { name: /E2E Alpha/ })).toBeVisible();
        await opened.close();
        expectCleanPage(pageErrorLog);
    });

    test('"Download PDF" opens the silent print route', async ({ page, pageErrorLog }) => {
        await openReportMenu(page);

        const popup = page.waitForEvent('popup');

        await page.getByRole('menuitem', { name: 'Download PDF' }).click();

        const opened = await popup;
        const url = new URL(opened.url());

        expect(url.pathname).toBe('/flows/5/report');
        // `download` starts the PDF and `silent` closes the tab; without it the export looks hung.
        expect(url.searchParams.get('download')).toBe('true');
        expect(url.searchParams.get('silent')).toBe('true');
        await opened.close();
        expectCleanPage(pageErrorLog);
    });

    test('"Download MD" writes the report to a named file', async ({ page, pageErrorLog }) => {
        await openReportMenu(page);

        const download = page.waitForEvent('download');

        await page.getByRole('menuitem', { name: 'Download MD' }).click();

        const file = await download;

        expect(file.suggestedFilename()).toMatch(REPORT_FILE);

        const stream = await file.createReadStream();
        const body = (await stream.toArray()).join('');

        expect(body).toContain('E2E Task Alpha');
        expect(body).toContain('Ports 22, 80 open');
        expectCleanPage(pageErrorLog);
    });

    test('"Copy to clipboard" puts the same report on the clipboard', async ({ context, page, pageErrorLog }) => {
        await context.grantPermissions(['clipboard-read', 'clipboard-write']);
        await openReportMenu(page);
        await page.getByRole('menuitem', { name: 'Copy to clipboard' }).click();

        await expect(page.getByText('Report copied to clipboard')).toBeVisible();
        expect(await page.evaluate(() => navigator.clipboard.readText())).toContain('E2E Task Alpha');
        expectCleanPage(pageErrorLog);
    });

    test('the print route writes a real PDF, not just a file with the right name', async ({ page, pageErrorLog }) => {
        const download = page.waitForEvent('download');

        await page.goto('/flows/5/report?download=true');

        const file = await download;

        expect(file.suggestedFilename()).toMatch(REPORT_PDF);

        const stream = await file.createReadStream();
        const chunks: Buffer[] = [];

        for await (const chunk of stream) {
            chunks.push(Buffer.from(chunk));
        }

        const shape = inspectPdf(Buffer.concat(chunks));

        // A generator that silently produced a blank document, or an error page saved under a .pdf
        // name, passes a filename check and fails every line of this one.
        expect(shape.isPdf && shape.trailerComplete, 'a complete PDF document').toBe(true);
        expect(shape.pages, 'the report paginated').toBeGreaterThan(0);
        // The seeded report is small; the threshold only has to be above what a title-only or blank
        // document produces, which is what a silently failing generator emits.
        expect(shape.textRuns, 'glyphs were actually drawn').toBeGreaterThan(5);
        expect(shape.fonts, 'a font subset was embedded').toBeGreaterThan(0);
        expectCleanPage(pageErrorLog);
    });
});
