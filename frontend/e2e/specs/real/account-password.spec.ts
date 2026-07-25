import type { APIRequestContext } from '@playwright/test';

import { request as apiRequest, expect, test } from '@playwright/test';

const INITIAL_PASSWORD = 'InitPass1!e2e';
const MAX_BYTES = 72;
const AT_LIMIT = 'a'.repeat(MAX_BYTES);
const OVER_LIMIT = 'a'.repeat(MAX_BYTES + 1);
const MULTIBYTE_OVER_LIMIT = 'é'.repeat(40);

const changePassword = (
    api: APIRequestContext,
    body: { confirm_password?: string; current_password: string; password: string },
) => api.put('/api/v1/user/password', { data: { confirm_password: body.password, ...body } });

const login = (api: APIRequestContext, mail: string, password: string) =>
    api.post('/api/v1/auth/login', { data: { mail, password } });

const ADMIN_USER = process.env.E2E_USER ?? 'admin@pentagi.com';
const ADMIN_PASSWORD = process.env.E2E_PASSWORD ?? 'admin';

test.describe('account password at the trust boundary', { tag: '@real' }, () => {
    // Deliberately bypasses the form: a client-side guard is one DevTools click from gone, so the
    // matrix has to reach the endpoint the way an attacker would.
    test('enforces the password policy on the endpoint, not just in the form', async ({ baseURL }) => {
        const admin = await apiRequest.newContext({ baseURL, ignoreHTTPSErrors: true });
        const victim = await apiRequest.newContext({ baseURL, ignoreHTTPSErrors: true });

        const adminLogin = await login(admin, ADMIN_USER, ADMIN_PASSWORD);

        expect(adminLogin.status(), 'seeded admin credentials').toBe(200);

        const mail = `e2e-pwd-${Date.now()}@pentagi.com`;
        const created = await admin.post('/api/v1/users/', {
            data: {
                mail,
                name: `e2epwd${Date.now()}`,
                password: INITIAL_PASSWORD,
                role_id: 2,
                status: 'active',
                type: 'local',
            },
        });

        expect(created.status(), 'throwaway user created').toBe(201);

        const hash = (await created.json()).data.hash as string;

        try {
            const tooLongAtCreate = await admin.post('/api/v1/users/', {
                data: {
                    mail: `e2e-long-${Date.now()}@pentagi.com`,
                    name: `e2elong${Date.now()}`,
                    password: OVER_LIMIT,
                    role_id: 2,
                    status: 'active',
                    type: 'local',
                },
            });

            expect(tooLongAtCreate.status(), 'over-length password at user creation').toBe(400);

            expect((await login(victim, mail, INITIAL_PASSWORD)).status()).toBe(200);

            for (const [name, body, wantStatus, wantCode] of [
                [
                    'wrong current password',
                    { current_password: 'NotTheOne1!', password: 'NewPass1!abc' },
                    403,
                    'Users.ChangePasswordCurrentUser.InvalidCurrentPassword',
                ],
                [
                    'confirmation mismatch',
                    {
                        confirm_password: 'NewPass1!abcd',
                        current_password: INITIAL_PASSWORD,
                        password: 'NewPass1!abc',
                    },
                    400,
                    'Users.ChangePasswordCurrentUser.InvalidPassword',
                ],
                [
                    'seven characters with every class',
                    { current_password: INITIAL_PASSWORD, password: 'Ab1$cde' },
                    400,
                    'Users.ChangePasswordCurrentUser.InvalidPassword',
                ],
                [
                    'fifteen characters without a special',
                    { current_password: INITIAL_PASSWORD, password: 'Abcdefghij12345' },
                    400,
                    'Users.ChangePasswordCurrentUser.InvalidPassword',
                ],
                [
                    'one byte over the bcrypt limit',
                    { current_password: INITIAL_PASSWORD, password: OVER_LIMIT },
                    400,
                    'Users.ChangePasswordCurrentUser.InvalidPassword',
                ],
                [
                    'within the rune count but over the byte limit',
                    { current_password: INITIAL_PASSWORD, password: MULTIBYTE_OVER_LIMIT },
                    400,
                    'Users.ChangePasswordCurrentUser.InvalidPassword',
                ],
            ] as const) {
                const response = await changePassword(victim, body);

                expect([name, response.status(), (await response.json()).code]).toEqual([name, wantStatus, wantCode]);
            }

            expect((await login(victim, mail, INITIAL_PASSWORD)).status(), 'no rejected row was saved').toBe(200);

            expect(
                (
                    await changePassword(victim, { current_password: INITIAL_PASSWORD, password: 'Abcdefghij123$4' })
                ).status(),
                'fifteen characters with every class is the short-branch maximum',
            ).toBe(200);
            expect(
                (await changePassword(victim, { current_password: 'Abcdefghij123$4', password: AT_LIMIT })).status(),
                'exactly the bcrypt limit is accepted',
            ).toBe(200);

            expect((await login(victim, mail, AT_LIMIT)).status(), 'the accepted password authenticates').toBe(200);
            expect((await login(victim, mail, INITIAL_PASSWORD)).status(), 'the replaced password does not').not.toBe(
                200,
            );
        } finally {
            await admin.delete(`/api/v1/users/${hash}`);
            await admin.dispose();
            await victim.dispose();
        }
    });
});
