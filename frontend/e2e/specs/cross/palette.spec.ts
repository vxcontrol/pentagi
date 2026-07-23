import type { Page } from '@playwright/test';

import type { BadgeVariant } from '@/components/ui/badge';

import { badgeVariants } from '@/components/ui/badge';
import { buttonVariants } from '@/components/ui/button';
import { routes } from '@/lib/routes';

import { expect, test } from '../../fixtures/test.ts';
import { ROUTE_MANIFEST } from '../../routes.ts';

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

/**
 * Off-palette colours already on the page, keyed by scan surface. A route's default view is keyed by
 * its path; a tab panel by `${path} [${tab}]`. Exact strings: they waive one node, not a rule.
 */
const ACCEPTED: Record<string, string[]> = {
    // file-manager.tsx expand-all control; changing it moves pixels, so it goes with the design pass.
    // Same control renders in the flow Files tab, so it is waived there under the same rationale.
    '/resources': ['button: hover:text-blue-400'],
    [`${routes.flow('5')} [Files]`]: ['button: hover:text-blue-400'],
};

const scanOffenders = (page: Page) =>
    page.evaluate(
        ({ pattern, sanctioned }) => {
            const palette = new RegExp(pattern);
            const allowed = new Set(sanctioned);

            return [...document.querySelectorAll('[data-slot="badge"],[data-slot="button"]')].flatMap((element) =>
                [...element.classList]
                    .filter((token) => palette.test(token) && !allowed.has(token))
                    .map((token) => `${element.getAttribute('data-slot')}: ${token}`),
            );
        },
        { pattern: PALETTE_UTILITY.source, sanctioned: [...SANCTIONED] },
    );

test.describe('palette compliance', { tag: '@cross' }, () => {
    for (const entry of ROUTE_MANIFEST) {
        test.describe(entry.path, () => {
            test.use({ cassette: entry.cassette() });

            test('rendered badges and buttons carry no off-palette colour', async ({ page }) => {
                await page.goto(entry.path);
                await expect(entry.ready(page)).toBeVisible();

                const offenders = await scanOffenders(page);

                expect([...new Set(offenders)], `off-palette colours on ${entry.path}`).toEqual(
                    ACCEPTED[entry.path] ?? [],
                );
            });

            // Tabs mount their panels lazily (Radix unmounts inactive ones), so the default-view scan
            // above never sees them — sweep each panel like the a11y gate does.
            for (const tab of entry.tabs ?? []) {
                test(`tab "${tab}" carries no off-palette colour`, async ({ page }) => {
                    await page.goto(entry.path);
                    await expect(entry.ready(page)).toBeVisible();
                    await page.getByRole('tab', { name: tab }).click();

                    const offenders = await scanOffenders(page);
                    const key = `${entry.path} [${tab}]`;

                    expect([...new Set(offenders)], `off-palette colours on ${key}`).toEqual(ACCEPTED[key] ?? []);
                });
            }
        });
    }
});
