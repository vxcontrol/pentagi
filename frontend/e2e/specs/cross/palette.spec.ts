import type { Page } from '@playwright/test';

import type { BadgeVariant } from '@/components/ui/badge';

import { badgeVariants } from '@/components/ui/badge';
import { buttonVariants } from '@/components/ui/button';
import { routes } from '@/lib/routes';

import { expect, test } from '../../fixtures/test.ts';
import { populatedSettingsProvidersCassette } from '../../mocks/cassettes/settings-providers.ts';
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

// A hard-coded colour (`text-green-800`, `bg-[#16a34a]`, `text-white`) rather than a semantic token
// (`bg-primary`, `text-muted-foreground`) or a non-colour utility (`text-xs`). white/black/transparent
// are absolute too: an overlay like `text-white` on a coloured badge is off-palette, not a token.
const PALETTE_UTILITY =
    /^(?:[a-z-]+:)*(?:bg|text|border|ring|from|via|to|fill|stroke|shadow|outline|decoration|divide|accent|caret|placeholder)-(?:white|black|transparent|\[[^\]]*\]|(?:slate|gray|zinc|neutral|stone|red|orange|amber|yellow|lime|green|emerald|teal|cyan|sky|blue|indigo|violet|purple|fuchsia|pink|rose)-\d{2,3})(?:\/\d{1,3})?$/;

/**
 * Per-variant, not one union: a flat union sanctions every variant's hue on every badge and button,
 * so the wrong-variant reuse this gate exists for reads as legal.
 */
const paletteTokensOf = (classes: string) => classes.split(/\s+/).filter((token) => PALETTE_UTILITY.test(token));

const SANCTIONED_VARIANTS = [
    ...BADGE_VARIANTS.map((variant) => ({
        slot: 'badge',
        tokens: paletteTokensOf(badgeVariants({ variant })),
        variant,
    })),
    ...BUTTON_VARIANTS.map((variant) => ({
        slot: 'button',
        tokens: paletteTokensOf(buttonVariants({ variant })),
        variant,
    })),
];

/**
 * Off-palette colours already on the page, keyed by scan surface. A route's default view is keyed by
 * its path; a tab panel by `${path} [${tab}]`. Exact strings: they waive one node, not a rule.
 */
const ACCEPTED: Record<string, string[]> = {
    // file-manager.tsx expand-all control, on /resources and embedded in the flow Files tab.
    '/resources': ['button: hover:text-blue-400'],
    [`${routes.flow('5')} [Files]`]: ['button: hover:text-blue-400'],
};

const scanOffenders = (page: Page) =>
    page.evaluate(
        ({ pattern, variants }) => {
            const palette = new RegExp(pattern);

            return [...document.querySelectorAll('[data-slot="badge"],[data-slot="button"]')].flatMap((element) => {
                const slot = element.getAttribute('data-slot') ?? '';
                const used = [...element.classList].filter((token) => palette.test(token));

                if (!used.length) {
                    return [];
                }

                // Equality, not containment: a single token borrowed from another variant is a
                // subset of that variant's set, so containment would wave through exactly the
                // wrong-variant reuse this gate exists to catch.
                const covered = variants.some(
                    (candidate) =>
                        candidate.slot === slot &&
                        candidate.tokens.length === used.length &&
                        used.every((token) => candidate.tokens.includes(token)),
                );

                return covered ? [] : used.map((token) => `${slot}: ${token}`);
            });
        },
        { pattern: PALETTE_UTILITY.source, variants: SANCTIONED_VARIANTS },
    );

test.describe('palette compliance', { tag: '@cross' }, () => {
    // The manifest sweeps this route with the empty seed, so the provider cards — the surface the
    // shipped badge defect actually lived on — render nowhere in the loop below. The a11y and visual
    // gates each carry the same dedicated populated sweep.
    test.describe('settings providers (populated)', () => {
        test.use({ cassette: populatedSettingsProvidersCassette() });

        test('provider cards carry no off-palette colour', async ({ page }) => {
            await page.goto(routes.settings.providers);
            await expect(page.getByText('My Custom Endpoint')).toBeVisible();

            const offenders = await scanOffenders(page);
            const key = `${routes.settings.providers} (populated)`;

            expect(offenders.sort(), `off-palette colours on ${key}`).toEqual([...(ACCEPTED[key] ?? [])].sort());
        });
    });

    for (const entry of ROUTE_MANIFEST) {
        test.describe(entry.path, () => {
            test.use({ cassette: entry.cassette() });

            test('rendered badges and buttons carry no off-palette colour', async ({ page }) => {
                await page.goto(entry.path);
                await expect(entry.ready(page)).toBeVisible();

                const offenders = await scanOffenders(page);

                expect(offenders.sort(), `off-palette colours on ${entry.path}`).toEqual(
                    [...(ACCEPTED[entry.path] ?? [])].sort(),
                );
            });

            // Tabs mount their panels lazily (Radix unmounts inactive ones), so the default-view scan
            // above never sees them — sweep each panel like the a11y gate does.
            for (const tab of entry.tabs ?? []) {
                test(`tab "${tab.name}" carries no off-palette colour`, async ({ page }) => {
                    await page.goto(entry.path);
                    await expect(entry.ready(page)).toBeVisible();
                    await page.getByRole('tab', { name: tab.name }).click();
                    await expect(tab.ready(page)).toBeVisible();

                    const offenders = await scanOffenders(page);
                    const key = `${entry.path} [${tab.name}]`;

                    expect(offenders.sort(), `off-palette colours on ${key}`).toEqual(
                        [...(ACCEPTED[key] ?? [])].sort(),
                    );
                });
            }
        });
    }
});
