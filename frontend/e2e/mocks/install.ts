import type { BrowserContext, Route } from '@playwright/test';

import type { MockWorld } from './world.ts';

import { installWsMock } from './ws-graphql.ts';

const API_GLOB = '**/api/v1/**';
const GRAPHQL_PATH = '/api/v1/graphql';

export const installMockRoutes = async (context: BrowserContext, world: MockWorld): Promise<void> => {
    await installWsMock(context, world);

    // Single catch-all over the whole API surface: anything without a cassette
    // entry fails fast (501 + recorded in world.unmatched, asserted at teardown)
    // instead of leaking through the vite preview proxy to a live backend.
    const serve = async (route: Route) => {
        const request = route.request();
        const { pathname, searchParams } = new URL(request.url());

        if (pathname === GRAPHQL_PATH && request.method() === 'POST') {
            const body = request.postDataJSON() as {
                operationName?: string;
                variables?: Record<string, unknown>;
            };
            const operationName = body.operationName ?? '(anonymous)';
            const entry = world.matchGraphQL(operationName, body.variables);

            if (entry) {
                await route.fulfill({
                    json: entry.body ?? {
                        ...(entry.data === undefined ? {} : { data: entry.data }),
                        ...(entry.errors ? { errors: entry.errors } : {}),
                    },
                    status: entry.status ?? 200,
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
        const entry = world.matchRest(request.method(), pathname, body ?? undefined, searchParams);

        if (entry) {
            await route.fulfill(
                entry.contentType
                    ? {
                          body: entry.body as Buffer | string,
                          contentType: entry.contentType,
                          status: entry.status ?? 200,
                      }
                    : { json: entry.body ?? {}, status: entry.status ?? 200 },
            );

            return;
        }

        world.reportUnmatched(`${request.method()} ${pathname}${searchParams.size ? `?${searchParams}` : ''}`);
        await route.fulfill({
            json: { error: `e2e: no cassette entry for ${request.method()} ${pathname}`, status: 'error' },
            status: 501,
        });
    };

    // A page route is not inherited by the popups that page opens, so arm every page the context gets.
    context.on('page', (opened) => void opened.route(API_GLOB, serve));

    await Promise.all(context.pages().map((existing) => existing.route(API_GLOB, serve)));
};
