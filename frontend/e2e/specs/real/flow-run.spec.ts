import { expect, test } from '@playwright/test';

import { MESSAGE_ID_SELECTOR } from '../../helpers/subscriptions.ts';
import { readTerminalBuffer } from '../../helpers/terminal.ts';

test.describe('real backend flow run', { tag: '@real' }, () => {
    test('runs a flow end-to-end through the mock LLM', async ({ page }) => {
        // Above the sum of the step timeouts below (60+30+90+90+90+60 = 420s), so a slow-but-passing
        // run is not killed mid-step with a generic timeout that masks the real state.
        test.setTimeout(450_000);

        const pageErrors: string[] = [];

        page.on('pageerror', (error) => pageErrors.push(String(error)));

        await page.goto('/flows/new');
        await page.getByPlaceholder(/Describe what you would like PentAGI to test/).fill('Say hello');

        // The form is invalid until the providers query lands, so a cold stack keeps Submit
        // disabled for a while — clicking straight away spends the whole test timeout on it.
        const submit = page.getByRole('button', { name: 'Submit' });

        await expect(submit).toBeEnabled({ timeout: 60_000 });
        await submit.click();

        await expect(page).toHaveURL(/\/flows\/\d+/, { timeout: 30_000 });
        // Scope the list assertion to this attempt's flow id: a retry runs
        // against a DB that already holds the previous attempt's identically
        // named flow, and an unscoped row match would be a strict-mode
        // violation.
        const flowId = new URL(page.url()).pathname.match(/\d+/)?.[0] ?? '';

        await expect(page.locator(MESSAGE_ID_SELECTOR).first()).toBeVisible({ timeout: 90_000 });
        await expect(page.getByText('Hello from the e2e mock LLM!').first()).toBeVisible({ timeout: 90_000 });

        // The scenario's `uname -a` exec must stream its sandbox output back
        // into the xterm buffer — `Linux` appears only in the command's real
        // output, never in the echoed command line.
        await expect(async () => {
            expect(await readTerminalBuffer(page)).toMatch(/Linux/);
        }).toPass({ timeout: 90_000 });

        // A completed task settles the flow into Waiting (ready for the next
        // input) — Finished only happens via the explicit user action.
        await page.goto('/flows');
        await expect(
            page
                .getByRole('row', { name: new RegExp(`${flowId}.*Say Hello Flow`) })
                .getByText('Waiting', { exact: true }),
        ).toBeVisible({ timeout: 60_000 });
        expect(pageErrors, 'uncaught errors during the flow run').toEqual([]);
    });
});
