import { readFileSync } from 'node:fs';

import type { BadgeVariant } from '@/components/ui/badge';

import { badgeVariants } from '@/components/ui/badge';
import { buttonVariants } from '@/components/ui/button';

import type { EditorProbe } from '../../helpers/contrast.ts';

import { expect, test } from '../../fixtures/test.ts';
import { AA_NORMAL, measureContrast, mountContrastProbes, mountEditorProbes } from '../../helpers/contrast.ts';
import { flowsCassette } from '../../mocks/cassettes/flows.ts';

const EDITOR_PROBES = {
    'editor-accent': { tag: 'a' },
    'editor-code': { tag: 'code' },
    'editor-tag': { className: 'template-tag', tag: 'span' },
    'editor-variable': { className: 'template-variable', tag: 'span' },
} satisfies Record<string, EditorProbe>;

/** The stylesheet is the token list, so a newly declared `--editor-*` has to be probed or exempted. */
const declaredEditorTokens = (): string[] => {
    const css = readFileSync(new URL('../../../src/styles/index.css', import.meta.url), 'utf8');

    return [...new Set(css.match(/--editor-[\w-]+/g) ?? [])].map((token) => token.slice(2));
};

// Keyed off the union so a newly added variant fails to compile until it is probed.
const BADGE_VARIANTS = Object.keys({
    blue: true,
    default: true,
    destructive: true,
    green: true,
    orange: true,
    outline: true,
    pink: true,
    purple: true,
    red: true,
    secondary: true,
    yellow: true,
} satisfies Record<BadgeVariant, true>) as BadgeVariant[];

const THEMES = ['light', 'dark'] as const;

test('every editor colour token is probed', { tag: '@cross' }, () => {
    const probed = new Set(Object.keys(EDITOR_PROBES));
    // A `-bg` token is the paired ground of its base probe, measured with it.
    const uncovered = declaredEditorTokens().filter(
        (token) => !probed.has(token) && !(token.endsWith('-bg') && probed.has(token.slice(0, -3))),
    );

    expect(uncovered).toEqual([]);
});

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
            await expect(page.locator('html')).toHaveClass(theme === 'dark' ? /dark/ : /light/);

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
            await expect(page.locator('html')).toHaveClass(theme === 'dark' ? /dark/ : /light/);

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
            await expect(page.locator('html')).toHaveClass(theme === 'dark' ? /dark/ : /light/);

            await mountEditorProbes(page, EDITOR_PROBES);

            for (const token of Object.keys(EDITOR_PROBES)) {
                expect
                    .soft(await measureContrast(page, token), `${token} (${theme})`)
                    .toBeGreaterThanOrEqual(AA_NORMAL);
            }
        });
    });
}
