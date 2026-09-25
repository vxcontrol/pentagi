import type { Page } from '@playwright/test';

import { expect, test } from '../../fixtures/test.ts';
import { flowsCassette } from '../../mocks/cassettes/flows.ts';

const SERVICE_DOWN = 'a required service is unavailable';

const countInfoCalls = (page: Page) => {
    const calls: string[] = [];

    page.on('request', (request) => {
        if (new URL(request.url()).pathname === '/api/v1/info') {
            calls.push(request.url());
        }
    });

    return calls;
};

test.describe('the flows list when its query fails', { tag: ['@flows', '@cross'] }, () => {
    test.describe('an unreachable database', () => {
        test.use({
            cassette: flowsCassette({
                queries: { flows: [{ errors: [{ extensions: { code: 'INTERNAL' }, message: SERVICE_DOWN }] }] },
            }),
        });

        test('shows the failure and offers a retry instead of an empty table', async ({ page }) => {
            const infoCalls = countInfoCalls(page);

            await page.goto('/flows');

            await expect(page.getByText(SERVICE_DOWN)).toBeVisible();
            await expect(page.getByRole('button', { name: /try again/i })).toBeVisible();
            expect(infoCalls, 'a storage failure is not an auth question').toHaveLength(0);
        });
    });

    test.describe('a permission denial', () => {
        test.use({
            cassette: flowsCassette({
                queries: {
                    flows: [{ errors: [{ extensions: { code: 'FORBIDDEN' }, message: 'not permitted' }] }],
                },
            }),
        });

        test('re-checks the session but keeps the operator on the page', async ({ page }) => {
            const infoCalls = countInfoCalls(page);

            await page.goto('/flows');

            await expect(page.getByText('not permitted')).toBeVisible();
            await expect
                .poll(() => infoCalls.length, { message: 'the denial asks /info whether the session is still good' })
                .toBe(1);
            await expect(page).toHaveURL(/\/flows$/);
        });
    });

    test.describe('a refusal that carries only prose', () => {
        test.use({
            cassette: flowsCassette({
                queries: {
                    flows: [{ errors: [{ message: 'unauthorized: non-user session is not allowed to view flows' }] }],
                },
            }),
        });

        test('is shown, not read as a session question', async ({ page }) => {
            const infoCalls = countInfoCalls(page);

            await page.goto('/flows');

            await expect(page.getByText(/non-user session is not allowed/)).toBeVisible();
            expect(infoCalls, 'the wording is not what decides an auth error').toHaveLength(0);
            await expect(page).toHaveURL(/\/flows$/);
        });
    });
});
