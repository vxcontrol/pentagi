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
    // A private endpoint answers an expired session with 403 + code AuthRequired, never 401 — pinning
    // 401 here would drive an interceptor branch production never reaches.
    test.describe('a REST 403', () => {
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
                            status: 403,
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

            // A login page that kept honouring the return URL would land back on the 403 and loop.
            expect(rejected, 'the rejected endpoint is called once').toHaveLength(1);
        });
    });

    // `/api/v1/graphql` is registered on the same private group as the REST endpoints, so an expired
    // session is refused by the middleware and gqlgen never runs: the client sees a transport 403
    // carrying the REST error body, not a GraphQL `errors` array.
    test.describe('a GraphQL 403', () => {
        test.use({
            cassette: flowsCassette({
                queries: {
                    flows: [
                        {
                            body: { code: 'AuthRequired', msg: 'auth required', status: 'error' },
                            setFlag: EXPIRED,
                            status: 403,
                        },
                    ],
                },
                rest: { ...baseRest(), ...infoTurningGuest() },
            }),
        });

        test('re-checks the session and sends the operator to login', async ({ page }) => {
            await page.goto('/flows');

            await expect(page).toHaveURL(/\/login\?returnUrl=%2Fflows/);
            await expect(page.getByRole('button', { name: 'Sign in' })).toBeVisible();
        });
    });

    test.describe('a GraphQL auth error', () => {
        test.use({
            cassette: flowsCassette({
                queries: { flows: [{ errors: [{ message: 'auth required' }], setFlag: EXPIRED }] },
                rest: { ...baseRest(), ...infoTurningGuest() },
            }),
        });

        test('re-checks the session on an auth-worded resolver error too', async ({ page }) => {
            await page.goto('/flows');

            await expect(page).toHaveURL(/\/login\?returnUrl=%2Fflows/);
            await expect(page.getByRole('button', { name: 'Sign in' })).toBeVisible();
        });
    });

    test.describe('the interceptor itself', () => {
        test.use({
            cassette: {
                queries: baseQueries(),
                rest: {
                    ...baseRest(),
                    // The login page re-reads /info and would write the key back; failing that read is
                    // the only way to observe what the interceptor left behind.
                    'GET /api/v1/info': [
                        authenticatedInfoEntry(),
                        { body: { code: 'Internal', status: 'error' }, status: 500, whenFlag: EXPIRED },
                    ],
                    'GET /api/v1/resources/': [
                        {
                            body: { code: 'AuthRequired', msg: 'auth required', status: 'error' },
                            setFlag: EXPIRED,
                            status: 403,
                        },
                    ],
                },
            },
        });

        test('clears the stored session outright, not just its user', async ({ page, pageErrorLog }) => {
            await page.goto('/resources');

            await expect(page).toHaveURL(/\/login/);
            expect(
                await page.evaluate(() => window.localStorage.getItem('auth')),
                'the key is removed, not overwritten',
            ).toBeNull();

            expect(pageErrorLog.pageErrors, 'no uncaught page errors').toEqual([]);
            expect(
                pageErrorLog.consoleErrors.filter((text) => !/403|500/.test(text)),
                'only the rejected requests are logged',
            ).toEqual([]);
        });
    });
});
