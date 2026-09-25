import { request as apiRequest, expect, test } from '@playwright/test';

const INITIAL_PASSWORD = 'InitPass1!e2e';
const WRONG_PASSWORD = 'NotTheOne1!e2e';
const PAIR_ATTEMPT_LIMIT = 10;

const ADMIN_USER = process.env.E2E_USER ?? 'admin@pentagi.com';
const ADMIN_PASSWORD = process.env.E2E_PASSWORD ?? 'admin';

test.describe('login throttling at the trust boundary', { tag: '@real' }, () => {
    test('locks one account out after repeated failures and leaves the others alone', async ({ baseURL }) => {
        const admin = await apiRequest.newContext({ baseURL, ignoreHTTPSErrors: true });
        const attacker = await apiRequest.newContext({ baseURL, ignoreHTTPSErrors: true });

        const login = (mail: string, password: string) =>
            attacker.post('/api/v1/auth/login', { data: { mail, password } });

        expect(
            (await admin.post('/api/v1/auth/login', { data: { mail: ADMIN_USER, password: ADMIN_PASSWORD } })).status(),
            'seeded admin credentials',
        ).toBe(200);

        const stamp = Date.now();
        const victim = `e2e-throttle-${stamp}@pentagi.com`;
        const bystander = `e2e-bystander-${stamp}@pentagi.com`;
        const hashes: string[] = [];

        for (const mail of [victim, bystander]) {
            const created = await admin.post('/api/v1/users/', {
                data: {
                    mail,
                    name: `e2ethrottle${stamp}${mail === victim ? 'v' : 'b'}`,
                    password: INITIAL_PASSWORD,
                    role_id: 2,
                    status: 'active',
                    type: 'local',
                },
            });

            expect(created.status(), `throwaway user ${mail} created`).toBe(201);
            hashes.push((await created.json()).data.hash as string);
        }

        try {
            const codes: number[] = [];

            for (let attempt = 0; attempt < PAIR_ATTEMPT_LIMIT; attempt++) {
                codes.push((await login(victim, WRONG_PASSWORD)).status());
            }

            expect(codes.slice(0, PAIR_ATTEMPT_LIMIT - 1), 'attempts under the limit are ordinary refusals').toEqual(
                Array(PAIR_ATTEMPT_LIMIT - 1).fill(401),
            );

            const locked = await login(victim, INITIAL_PASSWORD);

            expect(locked.status(), 'past the limit even the right password is refused').toBe(429);
            expect(Number(locked.headers()['retry-after']), 'the refusal says how long to wait').toBeGreaterThan(0);
            expect((await locked.json()).code).toBe('Auth.TooManyAttempts');

            expect(
                (await login(bystander, WRONG_PASSWORD)).status(),
                'the lockout is keyed on the pair, not on everyone',
            ).toBe(401);

            expect(
                (await login(bystander, INITIAL_PASSWORD)).status(),
                'and a bystander with the right password still gets in',
            ).toBe(200);
        } finally {
            for (const hash of hashes) {
                await admin.delete(`/api/v1/users/${hash}`);
            }

            await admin.dispose();
            await attacker.dispose();
        }
    });

    test('answers an unknown account no faster than a wrong password', async ({ baseURL }) => {
        const prober = await apiRequest.newContext({ baseURL, ignoreHTTPSErrors: true });

        expect(
            (
                await prober.post('/api/v1/auth/login', {
                    data: { mail: ADMIN_USER, password: ADMIN_PASSWORD },
                })
            ).status(),
            'the known account has to exist, or both samples measure a miss',
        ).toBe(200);

        const timeLogin = async (mail: string) => {
            const started = Date.now();
            const response = await prober.post('/api/v1/auth/login', { data: { mail, password: WRONG_PASSWORD } });

            expect(response.status()).toBe(401);

            return Date.now() - started;
        };

        const stamp = Date.now();

        await timeLogin(`e2e-warmup-${stamp}@pentagi.com`);

        const known = await timeLogin(ADMIN_USER);
        const unknown = await timeLogin(`e2e-absent-${stamp}@pentagi.com`);

        expect(unknown * 4, `known ${known} ms, unknown ${unknown} ms`).toBeGreaterThanOrEqual(known);

        await prober.dispose();
    });
});
