import type { Page } from '@playwright/test';

import type { MockWorld } from './world.ts';

import { installWsMock } from './ws-graphql.ts';

const GRAPHQL_PATH = '/api/v1/graphql';

export const installMockRoutes = async (page: Page, world: MockWorld): Promise<void> => {
    await installWsMock(page, world);

    // Single catch-all over the whole API surface: anything without a cassette
    // entry fails fast (501 + recorded in world.unmatched, asserted at teardown)
    // instead of leaking through the vite preview proxy to a live backend.
    await page.route('**/api/v1/**', async (route) => {
        const request = route.request();
        const { pathname } = new URL(request.url());

        if (pathname === GRAPHQL_PATH && request.method() === 'POST') {
            const body = request.postDataJSON() as {
                operationName?: string;
                variables?: Record<string, unknown>;
            };
            const operationName = body.operationName ?? '(anonymous)';
            const entry = world.matchGraphQL(operationName, body.variables);

            if (entry) {
                await route.fulfill({
                    json: entry.errors ? { errors: entry.errors } : { data: entry.data },
                });

                return;
            }

            world.reportUnmatched(`graphql ${operationName} ${JSON.stringify(body.variables ?? {})}`);
            await route.fulfill({
                json: { errors: [{ message: `e2e: no cassette entry for operation "${operationName}"` }] },
                status: 501,
            });

            return;
        }

        // postDataJSON throws on non-JSON bodies (multipart uploads) — only
        // parse what declares itself JSON.
        const contentType = request.headers()['content-type'] ?? '';
        const body = contentType.includes('application/json')
            ? (request.postDataJSON() as Record<string, unknown> | undefined)
            : undefined;
        const entry = world.matchRest(request.method(), pathname, body ?? undefined);

        if (entry) {
            await route.fulfill({
                json: entry.body ?? {},
                status: entry.status ?? 200,
            });

            return;
        }

        world.reportUnmatched(`${request.method()} ${pathname}`);
        await route.fulfill({
            json: { error: `e2e: no cassette entry for ${request.method()} ${pathname}`, status: 'error' },
            status: 501,
        });
    });
};
