import type { Page } from '@playwright/test';

// Structural mirror of the XtermHostElement handle attached in
// src/components/shared/terminal/use-xterm.ts — importing that module here
// would drag vite-only ambient types into the e2e tsconfig program.
interface XtermHost extends HTMLElement {
    xterm?: {
        buffer: {
            active: {
                getLine: (row: number) => undefined | { translateToString: (trim: boolean) => string };
                length: number;
            };
        };
    };
}

export const readTerminalBuffer = async (page: Page): Promise<string> =>
    page
        .locator('.xterm')
        .first()
        .evaluate((element) => {
            const host = element.parentElement as null | XtermHost;
            const terminal = host?.xterm;

            if (!terminal) {
                return '';
            }

            const buffer = terminal.buffer.active;
            const lines: string[] = [];

            for (let row = 0; row < buffer.length; row += 1) {
                lines.push(buffer.getLine(row)?.translateToString(true) ?? '');
            }

            return lines.join('\n').trimEnd();
        });
