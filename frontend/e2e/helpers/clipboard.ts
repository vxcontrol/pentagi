import type { Page } from '@playwright/test';

/** Install before the first navigation. */
export const recordClipboardWrites = async (page: Page): Promise<void> => {
    await page.addInitScript(() => {
        const copied: string[] = [];

        (window as unknown as { e2eClipboard: string[] }).e2eClipboard = copied;

        Object.defineProperty(navigator, 'clipboard', {
            configurable: true,
            value: {
                writeText: (text: string) => {
                    copied.push(text);

                    return Promise.resolve();
                },
            },
        });
    });
};

export const clipboardWrites = (page: Page): Promise<string[]> =>
    page.evaluate(() => (window as unknown as { e2eClipboard: string[] }).e2eClipboard);
