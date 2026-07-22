import { expect, test } from '../../fixtures/test.ts';
import { expectCleanPage } from '../../helpers/errors.ts';
import { flowsCassette, flowTabsCassette } from '../../mocks/cassettes/flows.ts';

// Nothing else in the suite presses Prev/Next, which is how a redirect that fired mid-switch
// shipped: every other flow spec reaches a detail page through the list.
test.describe('flow pager', { tag: ['@flows', '@smoke'] }, () => {
    test.use({ cassette: flowsCassette() });

    test.describe('with a report to show', () => {
        test.use({ cassette: flowTabsCassette() });

        // Not covered by the pixel baselines: the header cluster is well under the visual
        // project's maxDiffPixelRatio, so a reordering passes there unnoticed.
        test('keeps the variable action left of the fixed ones', async ({ page }) => {
            await page.goto('/flows/5');
            await expect(page.locator('header').getByRole('button', { name: 'Report' })).toBeVisible();

            const labels = await page
                .locator('header button')
                .evaluateAll((buttons) => buttons.map((button) => button.getAttribute('aria-label') ?? ''));
            const positionOf = (label: string) => labels.findIndex((candidate) => candidate.startsWith(label));

            // Report comes and goes with the task list; anchored leftmost, its arrival cannot shift
            // the controls that are always there.
            expect(positionOf('Report')).toBeLessThan(positionOf('Toggle favorite'));
            expect(positionOf('Toggle favorite')).toBeLessThan(positionOf('Previous'));
            expect(positionOf('Next')).toBeLessThan(positionOf('Flow actions'));
        });
    });

    test('steps to the sibling flow and back without passing through the list', async ({ page, pageErrorLog }) => {
        const header = page.locator('header');

        await page.goto('/flows/5');
        await expect(header.getByText('E2E Alpha')).toBeVisible();

        // Hold the sibling's data back so the in-flight window is wide enough to sample: the
        // cassette answers in single-digit milliseconds, which hides an unmount entirely.
        await page.route('**/graphql', async (route) => {
            await new Promise((resolve) => setTimeout(resolve, 300));
            await route.fallback();
        });

        await header.getByRole('button', { name: 'Next' }).click();

        const samples = await page.evaluate(async () => {
            const taken = [];

            for (let index = 0; index < 10; index += 1) {
                taken.push({
                    hasPager: !!document.querySelector('header button[aria-label="Next"]'),
                    path: window.location.pathname,
                });
                await new Promise((resolve) => setTimeout(resolve, 30));
            }

            return taken;
        });

        await expect(page).toHaveURL(/\/flows\/6$/);
        await expect(header.getByText('E2E Beta')).toBeVisible();

        // Sampled rather than awaited: `toBeVisible` retries, so it passes even if the cluster
        // unmounts for the length of the fetch and comes back — which is what gating the pager
        // on the loaded flow did, costing the user a round trip per step.
        expect(samples.every((sample) => sample.hasPager)).toBe(true);
        expect(samples.map((sample) => sample.path)).not.toContain('/flows');

        await header.getByRole('button', { name: 'Previous' }).click();

        await expect(page).toHaveURL(/\/flows\/5$/);
        expectCleanPage(pageErrorLog);
    });
});
