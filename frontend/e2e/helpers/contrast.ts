import type { Page } from '@playwright/test';

export const AA_NORMAL = 4.5;

const HOST_ID = 'contrast-probes';

export const mountContrastProbes = async (page: Page, probes: Record<string, string>): Promise<void> => {
    await page.evaluate(
        ({ entries, hostId }) => {
            document.getElementById(hostId)?.remove();
            const host = document.createElement('div');

            host.id = hostId;
            host.style.cssText =
                'position:fixed;inset:0 auto auto 0;z-index:2147483647;display:flex;flex-wrap:wrap;gap:8px;padding:12px;background:var(--card)';

            for (const [name, className] of entries) {
                const probe = document.createElement('span');

                probe.className = className;
                probe.dataset.contrast = name;
                // The variants carry `transition-colors`, so a measurement taken right after
                // hover() would sample a colour mid-interpolation, near the rest state.
                probe.style.transition = 'none';
                probe.textContent = 'Sample';
                host.append(probe);
            }

            document.body.append(host);
        },
        { entries: Object.entries(probes), hostId: HOST_ID },
    );
};

export interface EditorProbe {
    className?: string;
    tag: string;
}

// Editor highlight tokens are styled through `.tiptap-content .ProseMirror <selector>`, so a flat
// probe span (as in mountContrastProbes) would pick up none of it and measure nothing. Reproduce
// the real ancestor chain on the shipped element. Surface is --card, within ~0.005 L of the
// editor's real `dark:bg-input/30` ground and conservative for the floor.
export const mountEditorProbes = async (page: Page, probes: Record<string, EditorProbe>): Promise<void> => {
    await page.evaluate(
        ({ entries, hostId }) => {
            document.getElementById(hostId)?.remove();
            const host = document.createElement('div');

            host.id = hostId;
            host.className = 'tiptap-content';
            host.style.cssText = 'position:fixed;inset:0 auto auto 0;z-index:2147483647;padding:12px';

            const surface = document.createElement('div');

            surface.className = 'ProseMirror';
            surface.style.cssText = 'background:var(--card);display:flex;gap:8px;padding:8px';

            for (const [name, { className, tag }] of entries) {
                const probe = document.createElement(tag);

                probe.className = className ?? '';
                probe.dataset.contrast = name;
                probe.textContent = 'Sample';
                surface.append(probe);
            }

            host.append(surface);
            document.body.append(host);

            // A probe still wearing the surface's own colour matched no rule. Without this the
            // measurement falls back to the page's default text pair, which clears AA silently.
            const inherited = getComputedStyle(surface).color;

            for (const probe of surface.children) {
                if (getComputedStyle(probe).color === inherited) {
                    throw new Error(`editor contrast probe "${probe.getAttribute('data-contrast')}" matched no rule`);
                }
            }
        },
        { entries: Object.entries(probes), hostId: HOST_ID },
    );
};

export const parseColorAlpha = (color: string): number => {
    const value = color.trim().toLowerCase();

    if (value === 'transparent') {
        return 0;
    }

    if (value.startsWith('#')) {
        const digits = value.slice(1);
        const hex = digits.length === 4 ? digits.slice(3).repeat(2) : digits.length === 8 ? digits.slice(6) : '';

        return hex ? parseInt(hex, 16) / 255 : 1;
    }

    const args = /^[a-z-]+\(([^)]*)\)$/.exec(value)?.[1];

    if (args === undefined) {
        return 1;
    }

    const raw = (args.includes('/') ? args.split('/').at(-1) : args.split(',').at(3))?.trim();

    if (raw === undefined) {
        return 1;
    }

    const alpha = raw.endsWith('%') ? Number(raw.slice(0, -1)) / 100 : Number(raw);

    return Number.isFinite(alpha) ? Math.min(Math.max(alpha, 0), 1) : 1;
};

// Walk to the first opaque ancestor, not just the parent: a probe inside a code block sits on
// `pre.hljs` through a transparent `code`, and stopping at the parent would measure it against
// nothing. The stack is painted onto a cleared canvas, so a translucent layer with no opaque base
// under it keeps its own colour and reads as a contrast it does not have.
export const resolveGroundLayers = (ancestorBackgrounds: string[]): string[] => {
    const stack: string[] = [];

    for (const background of ancestorBackgrounds) {
        const alpha = parseColorAlpha(background);

        if (alpha === 0) {
            continue;
        }

        stack.unshift(background);

        if (alpha === 1) {
            break;
        }
    }

    return stack;
};

export const mountFenceProbes = async (page: Page, tokens: string[]): Promise<void> => {
    await page.evaluate(
        ({ classNames, hostId }) => {
            const fence = document.querySelector('.tiptap-content .ProseMirror pre');
            const code = fence?.querySelector('code');

            if (!fence || !code) {
                throw new Error('no code fence mounted to probe');
            }

            document.getElementById(hostId)?.remove();
            const host = document.createElement('div');

            host.id = hostId;
            host.className = 'tiptap-content';
            host.style.cssText = 'position:fixed;inset:0 auto auto 0;z-index:2147483647';

            const surface = document.createElement('div');

            surface.className = 'ProseMirror';

            const pre = fence.cloneNode(false) as HTMLElement;
            const holder = code.cloneNode(false) as HTMLElement;

            for (const className of classNames) {
                const probe = document.createElement('span');

                probe.className = className;
                probe.dataset.contrast = className;
                probe.textContent = 'Sample';
                holder.append(probe);
            }

            pre.append(holder);
            surface.append(pre);
            host.append(surface);
            document.body.append(host);

            const cloned = getComputedStyle(pre).backgroundColor;
            const real = getComputedStyle(fence).backgroundColor;

            if (cloned !== real) {
                throw new Error(`fence clone sits on ${cloned}, the editor's fence on ${real}`);
            }
        },
        { classNames: tokens, hostId: 'contrast-fence-probes' },
    );
};

export const measureContrast = async (page: Page, probe: string): Promise<number> =>
    measureContrastAt(page, `[data-contrast="${probe}"]`);

export const measureContrastAt = async (page: Page, selector: string): Promise<number> => {
    const { ancestors, backgroundColor, color } = await page.evaluate((target) => {
        const element = document.querySelector<HTMLElement>(target);

        if (!element?.parentElement) {
            throw new Error(`contrast target "${target}" is not mounted`);
        }

        const chain: string[] = [];

        for (let node: HTMLElement | null = element.parentElement; node; node = node.parentElement) {
            chain.push(getComputedStyle(node).backgroundColor);
        }

        const own = getComputedStyle(element);

        return { ancestors: chain, backgroundColor: own.backgroundColor, color: own.color };
    }, selector);

    return page.evaluate(
        ({ ground, text }) => {
            const canvas = document.createElement('canvas');

            canvas.width = 1;
            canvas.height = 1;

            const context = canvas.getContext('2d', { willReadFrequently: true });

            if (!context) {
                throw new Error('2d canvas context unavailable');
            }

            const paint = (...layers: string[]) => {
                context.clearRect(0, 0, 1, 1);

                for (const layer of layers) {
                    // An unparseable colour is a silent no-op on fillStyle, leaving the previous
                    // value — which measures as a spurious ~21:1 pass. Assign it against two
                    // different sentinels: a colour that applied lands on the same value both
                    // times, an invalid one keeps whichever sentinel preceded it.
                    context.fillStyle = '#000000';
                    context.fillStyle = layer;
                    const fromBlack = context.fillStyle;

                    context.fillStyle = '#ffffff';
                    context.fillStyle = layer;

                    if (context.fillStyle !== fromBlack) {
                        throw new Error(`contrast: colour "${layer}" did not apply`);
                    }

                    context.fillRect(0, 0, 1, 1);
                }

                const [r = 0, g = 0, b = 0] = context.getImageData(0, 0, 1, 1).data;

                return { b, g, r };
            };

            const luminance = ({ b, g, r }: { b: number; g: number; r: number }) => {
                const channel = (value: number) => {
                    const ratio = value / 255;

                    return ratio <= 0.03928 ? ratio / 12.92 : ((ratio + 0.055) / 1.055) ** 2.4;
                };

                return 0.2126 * channel(r) + 0.7152 * channel(g) + 0.0722 * channel(b);
            };

            const surface = paint(...ground);
            const label = paint(...ground, text);
            const [high = 0, low = 0] = [luminance(label), luminance(surface)].sort((a, b) => b - a);

            return (high + 0.05) / (low + 0.05);
        },
        { ground: [...resolveGroundLayers(ancestors), backgroundColor], text: color },
    );
};
