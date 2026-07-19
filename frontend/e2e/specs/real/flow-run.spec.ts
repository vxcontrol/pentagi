import { expect, test } from '@playwright/test';

test.describe('real backend flow run', { tag: '@real' }, () => {
    test('runs a flow end-to-end through the mock LLM', async ({ page }) => {
        test.setTimeout(240_000);

        await page.goto('/flows/new');
        await page.getByPlaceholder(/Describe what you would like PentAGI to test/).fill('Say hello');
        await page.getByRole('button', { name: 'Submit' }).click();

        await expect(page).toHaveURL(/\/flows\/\d+/, { timeout: 30_000 });

        await expect(page.getByTestId('flow-message-id').first()).toBeVisible({ timeout: 90_000 });
        await expect(page.getByText('Hello from the e2e mock LLM!').first()).toBeVisible({ timeout: 90_000 });

        // A completed task settles the flow into Waiting (ready for the next
        // input) — Finished only happens via the explicit user action.
        await page.goto('/flows');
        await expect(
            page.getByRole('row', { name: /Say Hello Flow/ }).getByText('Waiting', { exact: true }),
        ).toBeVisible({ timeout: 60_000 });
    });
});
