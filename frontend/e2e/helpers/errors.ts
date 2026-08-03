import type { Page } from '@playwright/test';

import { expect } from '@playwright/test';

export interface PageErrorLog {
    consoleErrors: string[];
    pageErrors: string[];
}

/**
 * The mock tier runs the production bundle where terser drops app console
 * output, so `pageerror` (uncaught exceptions + unhandled rejections) is the
 * load-bearing assert; console errors still catch browser-generated failures
 * such as failed resource loads.
 */
export const expectCleanPage = (log: PageErrorLog): void => {
    expect(log.pageErrors, 'no uncaught page errors').toEqual([]);
    expect(log.consoleErrors, 'no browser console errors').toEqual([]);
};

export const watchPageErrors = (page: Page): PageErrorLog => {
    const log: PageErrorLog = { consoleErrors: [], pageErrors: [] };

    page.on('pageerror', (error) => log.pageErrors.push(String(error)));
    page.on('console', (message) => {
        if (message.type() === 'error') {
            log.consoleErrors.push(message.text());
        }
    });

    return log;
};
