import type { Page } from '@playwright/test';

/** Badge and button labels are 12–13px, so the 3:1 large-text allowance never applies. */
export const AA_NORMAL = 4.5;

const HOST_ID = 'contrast-probes';

/**
 * Paints one sample per class string on a card-coloured surface, on top of the
 * page so each probe is hoverable. Class strings come from the component's own
 * cva, so a palette edit changes what is measured — a probe list copied into the
 * spec would only ever re-test itself.
 */
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
                probe.textContent = 'Sample';
                host.append(probe);
            }

            document.body.append(host);
        },
        { entries: Object.entries(probes), hostId: HOST_ID },
    );
};

/**
 * The browser resolves and composites the colours, so oklch tokens and
 * translucent fills are measured as rendered rather than as declared.
 *
 * Returned unrounded: green-700 measures 4.4991 against its tint, which two
 * decimals would report as a passing "4.5".
 */
export const measureContrast = async (page: Page, probe: string): Promise<number> =>
    page.evaluate((name) => {
        const element = document.querySelector<HTMLElement>(`[data-contrast="${name}"]`);

        if (!element?.parentElement) {
            throw new Error(`contrast probe "${name}" is not mounted`);
        }

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
                // An invalid fillStyle silently keeps the previous colour, which
                // would report a missing token as a passing measurement.
                context.fillStyle = '#000000';
                context.fillStyle = layer;
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

        const surface = paint(
            getComputedStyle(element.parentElement).backgroundColor,
            getComputedStyle(element).backgroundColor,
        );
        const label = paint(getComputedStyle(element).color);
        const [high = 0, low = 0] = [luminance(label), luminance(surface)].sort((a, b) => b - a);

        return (high + 0.05) / (low + 0.05);
    }, probe);
