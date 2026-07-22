import type { BadgeVariant } from '@/components/ui/badge';

import { badgeVariants } from '@/components/ui/badge';
import { buttonVariants } from '@/components/ui/button';

import { expect, test } from '../../fixtures/test.ts';
import { ROUTE_MANIFEST } from '../../routes.ts';

// Keyed off the unions so a new variant joins the sanctioned set without a manual edit.
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

const BUTTON_VARIANTS = ['default', 'destructive', 'ghost', 'link', 'outline', 'secondary'] as const;

const SANCTIONED = new Set(
    [
        ...BADGE_VARIANTS.map((variant) => badgeVariants({ variant })),
        ...BUTTON_VARIANTS.map((variant) => buttonVariants({ variant })),
    ]
        .join(' ')
        .split(/\s+/),
);

// A hard-coded hue (`text-green-800`, `bg-[#16a34a]`) rather than a semantic token
// (`bg-primary`, `text-muted-foreground`) or a non-colour utility (`text-xs`).
const PALETTE_UTILITY =
    /^(?:[a-z-]+:)*(?:bg|text|border|ring|from|via|to|fill|stroke|shadow|outline|decoration|divide|accent|caret|placeholder)-(?:\[[^\]]*\]|(?:slate|gray|zinc|neutral|stone|red|orange|amber|yellow|lime|green|emerald|teal|cyan|sky|blue|indigo|violet|purple|fuchsia|pink|rose)-\d{2,3}(?:\/\d{1,3})?)$/;

/** Off-palette colours already on the page, by route. Exact strings: they waive one node, not a rule. */
const ACCEPTED: Record<string, string[]> = {
    // file-manager.tsx expand-all control; changing it moves pixels, so it goes with the design pass.
    '/resources': ['button: hover:text-blue-400'],
};

/**
 * The contrast gate probes cva output, so a colour written straight into a page — the shape of the
 * one badge defect that shipped — is outside it by construction. This walks the rendered DOM
 * instead: every badge and button must draw its colour from the variant set or from a token.
 */
test.describe('palette compliance', { tag: '@cross' }, () => {
    for (const entry of ROUTE_MANIFEST) {
        test.describe(entry.path, () => {
            test.use({ cassette: entry.cassette() });

            test('rendered badges and buttons carry no off-palette colour', async ({ page }) => {
                await page.goto(entry.path);
                await expect(entry.ready(page)).toBeVisible();

                const offenders = await page.evaluate(
                    ({ pattern, sanctioned }) => {
                        const palette = new RegExp(pattern);
                        const allowed = new Set(sanctioned);

                        return [...document.querySelectorAll('[data-slot="badge"],[data-slot="button"]')].flatMap(
                            (element) =>
                                [...element.classList]
                                    .filter((token) => palette.test(token) && !allowed.has(token))
                                    .map((token) => `${element.getAttribute('data-slot')}: ${token}`),
                        );
                    },
                    { pattern: PALETTE_UTILITY.source, sanctioned: [...SANCTIONED] },
                );

                const accepted = ACCEPTED[entry.path] ?? [];

                expect([...new Set(offenders)], `off-palette colours on ${entry.path}`).toEqual(accepted);
            });
        });
    }
});
