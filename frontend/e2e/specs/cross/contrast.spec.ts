import type { BadgeVariant } from '@/components/ui/badge';

import { badgeVariants } from '@/components/ui/badge';
import { buttonVariants } from '@/components/ui/button';

import { expect, test } from '../../fixtures/test.ts';
import { AA_NORMAL, measureContrast, mountContrastProbes, mountEditorProbes } from '../../helpers/contrast.ts';
import { flowsCassette } from '../../mocks/cassettes/flows.ts';

const EDITOR_TOKENS = {
    'editor-tag': 'template-tag',
    'editor-variable': 'template-variable',
} as const;

const BADGE_VARIANTS = [
    'blue',
    'default',
    'destructive',
    'green',
    'orange',
    'outline',
    'pink',
    'purple',
    'red',
    'secondary',
    'yellow',
] as const satisfies readonly BadgeVariant[];

const THEMES = ['light', 'dark'] as const;

for (const theme of THEMES) {
    test.describe(`contrast (${theme})`, { tag: '@cross' }, () => {
        test.use({ cassette: flowsCassette() });

        if (theme === 'dark') {
            test.beforeEach(async ({ page }) => {
                await page.addInitScript(() => window.localStorage.setItem('theme', 'dark'));
            });
        }

        test('every badge variant clears AA at rest and on hover', async ({ page }) => {
            await page.goto('/flows');
            await expect(page.getByRole('row', { name: /E2E Alpha/ })).toBeVisible();

            await mountContrastProbes(
                page,
                Object.fromEntries(BADGE_VARIANTS.map((variant) => [variant, badgeVariants({ variant })])),
            );

            for (const variant of BADGE_VARIANTS) {
                expect
                    .soft(await measureContrast(page, variant), `badge ${variant} at rest`)
                    .toBeGreaterThanOrEqual(AA_NORMAL);

                await page.locator(`[data-contrast="${variant}"]`).hover();
                expect
                    .soft(await measureContrast(page, variant), `badge ${variant} on hover`)
                    .toBeGreaterThanOrEqual(AA_NORMAL);
            }
        });

        test('the destructive button clears AA at rest and on hover', async ({ page }) => {
            await page.goto('/flows');
            await expect(page.getByRole('row', { name: /E2E Alpha/ })).toBeVisible();

            await mountContrastProbes(page, { destructive: buttonVariants({ variant: 'destructive' }) });

            expect.soft(await measureContrast(page, 'destructive'), 'button at rest').toBeGreaterThanOrEqual(AA_NORMAL);

            await page.locator('[data-contrast="destructive"]').hover();
            expect
                .soft(await measureContrast(page, 'destructive'), 'button on hover')
                .toBeGreaterThanOrEqual(AA_NORMAL);
        });

        test('editor highlight tokens clear AA on the editor surface', async ({ page }) => {
            await page.goto('/flows');
            await expect(page.getByRole('row', { name: /E2E Alpha/ })).toBeVisible();

            await mountEditorProbes(page, EDITOR_TOKENS);

            for (const token of Object.keys(EDITOR_TOKENS)) {
                expect
                    .soft(await measureContrast(page, token), `${token} (${theme})`)
                    .toBeGreaterThanOrEqual(AA_NORMAL);
            }
        });
    });
}
