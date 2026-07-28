import { readFileSync } from 'node:fs';

import type { BadgeVariant } from '@/components/ui/badge';
import type { ButtonVariant } from '@/components/ui/button';

import { badgeVariants } from '@/components/ui/badge';
import { buttonVariants } from '@/components/ui/button';

import type { EditorProbe } from '../../helpers/contrast.ts';

import { expect, test } from '../../fixtures/test.ts';
import {
    AA_NORMAL,
    measureContrast,
    measureContrastAt,
    mountContrastProbes,
    mountEditorProbes,
    mountFenceProbes,
} from '../../helpers/contrast.ts';
import { flowsCassette } from '../../mocks/cassettes/flows.ts';
import { PROMPT_DETAIL_AGENT, promptDetailCassette } from '../../mocks/cassettes/settings-prompts.ts';

const HLJS_TOKENS = [
    'hljs-comment',
    'hljs-keyword',
    'hljs-section',
    'hljs-literal',
    'hljs-string',
    'hljs-attr',
    'hljs-symbol',
    'hljs-built_in',
];

const SYNTAX_TOKEN_CLASS: Record<string, string> = {
    'editor-syntax-attr': 'hljs-attr',
    'editor-syntax-class': 'hljs-built_in',
    'editor-syntax-comment': 'hljs-comment',
    'editor-syntax-entity': 'hljs-section',
    'editor-syntax-fg': 'fence text',
    'editor-syntax-keyword': 'hljs-keyword',
    'editor-syntax-literal': 'hljs-literal',
    'editor-syntax-string': 'hljs-string',
    'editor-syntax-symbol': 'hljs-symbol',
};

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

// Keyed off the unions so a newly added variant fails to compile until it is probed.
const BUTTON_VARIANTS = Object.keys({
    default: true,
    destructive: true,
    ghost: true,
    link: true,
    outline: true,
    secondary: true,
} satisfies Record<ButtonVariant, true>) as ButtonVariant[];

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
        (token) =>
            !probed.has(token) &&
            !(token.endsWith('-bg') && probed.has(token.slice(0, -3))) &&
            !(token in SYNTAX_TOKEN_CLASS),
    );

    expect(uncovered).toEqual([]);
});

test('every syntax token is measured through a mounted hljs class', { tag: '@cross' }, () => {
    const measured = new Set([...HLJS_TOKENS, 'fence text']);
    const declared = declaredEditorTokens().filter((token) => token.startsWith('editor-syntax-'));

    expect(declared.filter((token) => !measured.has(SYNTAX_TOKEN_CLASS[token] ?? ''))).toEqual([]);
    expect(declared.sort()).toEqual(Object.keys(SYNTAX_TOKEN_CLASS).sort());
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

        test('every button variant clears AA at rest and on hover', async ({ page }) => {
            await page.goto('/flows');
            await expect(page.getByRole('row', { name: /E2E Alpha/ })).toBeVisible();
            await expect(page.locator('html')).toHaveClass(theme === 'dark' ? /dark/ : /light/);

            await mountContrastProbes(
                page,
                Object.fromEntries(BUTTON_VARIANTS.map((variant) => [variant, buttonVariants({ variant })])),
            );

            for (const variant of BUTTON_VARIANTS) {
                expect
                    .soft(await measureContrast(page, variant), `button ${variant} at rest`)
                    .toBeGreaterThanOrEqual(AA_NORMAL);

                await page.locator(`[data-contrast="${variant}"]`).hover();
                expect
                    .soft(await measureContrast(page, variant), `button ${variant} on hover`)
                    .toBeGreaterThanOrEqual(AA_NORMAL);
            }
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

    test.describe(`contrast in a code block (${theme})`, { tag: '@cross' }, () => {
        test.use({ cassette: promptDetailCassette() });

        if (theme === 'dark') {
            test.beforeEach(async ({ page }) => {
                await page.addInitScript(() => window.localStorage.setItem('theme', 'dark'));
            });
        }

        test('editor highlight tokens clear AA inside a fence', async ({ page }) => {
            await page.goto(`/settings/prompts/${PROMPT_DETAIL_AGENT}`);

            const fence = page.locator('.tiptap-content .ProseMirror pre');

            await expect(fence).toBeVisible();
            await expect(page.locator('html')).toHaveClass(theme === 'dark' ? /dark/ : /light/);

            // Resolve the token through a probe, not getPropertyValue: the minifier rewrites the authored
            // text (`oklch(0.94 …)` ships as `oklch(94% …)`), which never equals the computed colour.
            const surface = await page.evaluate(() => {
                const probe = document.createElement('div');

                probe.style.backgroundColor = 'var(--editor-code-bg)';
                document.body.append(probe);

                const resolved = getComputedStyle(probe).backgroundColor;

                probe.remove();

                return resolved;
            });

            expect(surface).not.toBe('');
            await expect(fence).toHaveCSS('background-color', surface);

            await mountFenceProbes(page, HLJS_TOKENS);

            for (const token of HLJS_TOKENS) {
                expect
                    .soft(await measureContrast(page, token), `${token} inside a code block (${theme})`)
                    .toBeGreaterThanOrEqual(AA_NORMAL);
            }

            for (const token of ['variable', 'tag'] as const) {
                await expect(fence.locator(`.template-${token}`).first()).toBeVisible();
                expect
                    .soft(
                        await measureContrastAt(page, `.tiptap-content .ProseMirror pre .template-${token}`),
                        `editor-${token} inside a code block (${theme})`,
                    )
                    .toBeGreaterThanOrEqual(AA_NORMAL);
            }
        });
    });
}
