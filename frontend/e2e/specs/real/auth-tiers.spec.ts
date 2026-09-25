import type { APIRequestContext } from '@playwright/test';

import { request as apiRequest, expect, test } from '@playwright/test';

const ADMIN_USER = process.env.E2E_USER ?? 'admin@pentagi.com';
const ADMIN_PASSWORD = process.env.E2E_PASSWORD ?? 'admin';

const TABLE_QUERY = 'page=1&pageSize=5&type=init';

const gql = (api: APIRequestContext, query: string) => api.post('/api/v1/graphql', { data: { query } });

test.describe('account and credential routes are session-only', { tag: '@real' }, () => {
    test('an API token cannot reach /users, /roles or /tokens', async ({ baseURL }) => {
        const session = await apiRequest.newContext({ baseURL, ignoreHTTPSErrors: true });

        const login = await session.post('/api/v1/auth/login', {
            data: { mail: ADMIN_USER, password: ADMIN_PASSWORD },
        });

        expect(login.status(), 'seeded admin credentials').toBe(200);

        const created = await gql(
            session,
            `mutation { createAPIToken(input: { name: "e2e-auth-tiers-${Date.now()}", ttl: 3600 }) { tokenId token } }`,
        );
        const createdBody = await created.json();

        expect(createdBody.errors, 'token minted through the session').toBeUndefined();

        const { token, tokenId } = createdBody.data.createAPIToken;
        const bearer = await apiRequest.newContext({
            baseURL,
            extraHTTPHeaders: { Authorization: `Bearer ${token}` },
            ignoreHTTPSErrors: true,
            storageState: { cookies: [], origins: [] },
        });

        try {
            for (const path of [
                `/api/v1/users/?${TABLE_QUERY}`,
                `/api/v1/roles/?${TABLE_QUERY}`,
                `/api/v1/tokens/?${TABLE_QUERY}`,
            ]) {
                expect([path, (await bearer.get(path)).status()], 'refused for a bearer token').toEqual([path, 403]);
            }

            const mail = `e2e-tier-${Date.now()}@pentagi.com`;
            const escalation = await bearer.post('/api/v1/users/', {
                data: {
                    mail,
                    name: `e2etier${Date.now()}`,
                    password: 'EscalateMe1!',
                    role_id: 1,
                    status: 'active',
                    type: 'local',
                },
            });

            expect(escalation.status(), 'a bearer token cannot mint an administrator').toBe(403);

            const users = await session.get(`/api/v1/users/?page=1&pageSize=1000&type=init`);
            const mails = ((await users.json()).data.users as { mail: string }[]).map((user) => user.mail);

            expect(mails, 'the refused request wrote nothing').not.toContain(mail);

            expect((await bearer.get(`/api/v1/flows/?${TABLE_QUERY}`)).status(), 'the flow API stays open').toBe(200);
            expect(users.status(), 'the session still administers users').toBe(200);
        } finally {
            await gql(session, `mutation { deleteAPIToken(tokenId: ${JSON.stringify(tokenId)}) }`);
            await bearer.dispose();
            await session.dispose();
        }
    });
});
