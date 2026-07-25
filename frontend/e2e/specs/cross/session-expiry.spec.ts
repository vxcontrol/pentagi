import type { Page } from '@playwright/test';

import type { Cassette } from '../../mocks/cassette.ts';

import { authenticatedInfoEntry, guestInfoEntry } from '../../fixtures/auth.ts';
import { expect, test } from '../../fixtures/test.ts';
import { baseQueries, baseRest } from '../../mocks/cassettes/base.ts';
import { flowsCassette } from '../../mocks/cassettes/flows.ts';

const EXPIRED = 'session-expired';

const infoTurningGuest = (): Cassette['rest'] => ({
    'GET /api/v1/info': [authenticatedInfoEntry(), { ...guestInfoEntry(), whenFlag: EXPIRED }],
});

const storedUser = (page: Page) =>
    page.evaluate(() => JSON.parse(window.localStorage.getItem('auth') ?? 'null')?.user ?? null);

test.describe('session expiry', { tag: '@cross' }, () => {
    test.describe('a REST 401', () => {
        test.use({
            cassette: {
                queries: baseQueries(),
                rest: {
                    ...baseRest(),
                    ...infoTurningGuest(),
                    'GET /api/v1/resources/': [
                        {
                            body: { code: 'AuthRequired', msg: 'auth required', status: 'error' },
                            setFlag: EXPIRED,
                            status: 401,
                        },
                    ],
                },
            },
        });

        test('bounces to the login page with a way back and drops the stored session', async ({ page }) => {
            const rejected: string[] = [];

            page.on('request', (request) => {
                if (new URL(request.url()).pathname === '/api/v1/resources/') {
                    rejected.push(request.url());
                }
            });

            await page.goto('/resources');

            await expect(page).toHaveURL(/\/login\?returnUrl=%2Fresources/);
            await expect(page.getByRole('button', { name: 'Sign in' })).toBeVisible();
            // The login page's own /info rewrites the key, so only the identity inside it matters.
            expect(await storedUser(page), 'the expired identity must not survive in storage').toBeNull();

            // A login page that kept honouring the return URL would land back on the 401 and loop.
            expect(rejected, 'the rejected endpoint is called once').toHaveLength(1);
        });
    });

    test.describe('a GraphQL auth error', () => {
        test.use({
            cassette: flowsCassette({
                queries: { flows: [{ errors: [{ message: 'auth required' }], setFlag: EXPIRED }] },
                rest: { ...baseRest(), ...infoTurningGuest() },
            }),
        });

        test('re-checks the session and sends the operator to login', async ({ page }) => {
            await page.goto('/flows');

            await expect(page).toHaveURL(/\/login\?returnUrl=%2Fflows/);
            await expect(page.getByRole('button', { name: 'Sign in' })).toBeVisible();
        });
    });
});
