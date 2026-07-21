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
                probe.textContent = 'Sample';
                host.append(probe);
            }

            document.body.append(host);
        },
        { entries: Object.entries(probes), hostId: HOST_ID },
    );
};

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
