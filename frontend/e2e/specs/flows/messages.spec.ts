import { expect, test } from '../../fixtures/test.ts';
import { expectCleanPage } from '../../helpers/errors.ts';
import { MESSAGE_ID_SELECTOR } from '../../helpers/subscriptions.ts';
import { variedMessagesCassette } from '../../mocks/cassettes/flows.ts';

type XtermHost = {
    xterm?: { buffer: { active: { getLine: (row: number) => undefined | XtermLine; length: number } } };
};
type XtermLine = { translateToString: (trim: boolean) => string };

const anyTerminalContains = (marker: string): boolean => {
    const hosts = document.querySelectorAll('.xterm');

    return Array.from(hosts).some((el) => {
        const terminal = (el.parentElement as null | XtermHost)?.xterm;

        if (!terminal) {
            return false;
        }

        const buffer = terminal.buffer.active;
        let text = '';

        for (let row = 0; row < buffer.length; row += 1) {
            text += `${buffer.getLine(row)?.translateToString(true) ?? ''}\n`;
        }

        return text.includes(marker);
    });
};

test.describe('flow message rendering', { tag: '@flows' }, () => {
    test.use({ cassette: variedMessagesCassette() });

    test('renders the thinking toggle, auto-expanded report, terminal, and input alignment', async ({
        page,
        pageErrorLog,
    }) => {
        await page.goto('/flows/5?tab=automation');
        await expect(page.locator(MESSAGE_ID_SELECTOR)).toHaveCount(4);

        const thinkingToggle = page.getByText('Show thinking');

        await expect(thinkingToggle).toBeVisible();
        // Collapsed first, so an always-expanded regression fails instead of passing.
        await expect(page.getByText('internal reasoning about the plan')).toBeHidden();
        await thinkingToggle.click();
        await expect(page.getByText('Hide thinking')).toBeVisible();
        await expect(page.getByText('internal reasoning about the plan')).toBeVisible();

        await expect(page.getByRole('heading', { name: 'Report' })).toBeVisible();
        await expect(page.getByText('Hide details').first()).toBeVisible();

        // Only a report auto-expands, so the terminal message starts collapsed and its xterm is
        // not mounted — the panel's is the only one on screen until the toggle is used.
        await expect(page.locator('.xterm')).toHaveCount(1);
        await page.getByText('Show details').click();

        await expect(page.locator('.xterm')).toHaveCount(2);
        await expect(async () => {
            expect(await page.evaluate(anyTerminalContains, 'e2e-terminal-marker')).toBe(true);
        }).toPass();

        const rightAligned = page.locator('.items-end');

        await expect(rightAligned).toHaveCount(1);
        await expect(rightAligned).toContainText('run the smoke command');

        expectCleanPage(pageErrorLog);
    });
});
