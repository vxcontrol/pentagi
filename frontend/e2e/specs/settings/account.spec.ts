import type { Locator, Page } from '@playwright/test';

import type { Cassette } from '../../mocks/cassette.ts';

import { SEEDED_USER } from '../../fixtures/auth.ts';
import { expect, test } from '../../fixtures/test.ts';
import { expectCleanPage } from '../../helpers/errors.ts';
import { baseQueries, baseRest } from '../../mocks/cassettes/base.ts';

const CURRENT_PASSWORD = 'CurrentPass1!';
// The server-side half of this boundary lives in specs/real/account-password.spec.ts.
const MAX_LENGTH_PASSWORD = 'n'.repeat(72);

const accountCassette = (rest: Cassette['rest'] = {}): Cassette => ({
    queries: baseQueries(),
    rest: { ...baseRest(), ...rest },
});

const section = (page: Page, title: string): Locator =>
    page.locator('[data-slot="card"]').filter({ has: page.locator('[data-slot="card-title"]', { hasText: title }) });

const openSection = async (page: Page, title: string): Promise<Locator> => {
    const card = section(page, title);

    await card.getByRole('button', { name: 'Change' }).click();

    return card;
};

const fillPasswordForm = async (page: Page, next: string, confirm = next): Promise<void> => {
    await page.getByLabel('Current Password').fill(CURRENT_PASSWORD);
    await page.getByLabel('New Password', { exact: true }).fill(next);
    await page.getByLabel('Confirm New Password').fill(confirm);
};

const submitPasswordForm = (page: Page): Promise<void> => page.getByRole('button', { name: 'Update Password' }).click();

test.describe('account settings', { tag: '@settings' }, () => {
    test.describe('read', () => {
        test.use({ cassette: accountCassette() });

        test('renders the signed-in identity and every editable section', async ({ page, pageErrorLog }) => {
            await page.goto('/settings/account');

            await expect(page.getByRole('heading', { exact: true, name: SEEDED_USER.name })).toBeVisible();
            await expect(page.getByText('Member since January 2026')).toBeVisible();
            await expect(page.getByText('Local account')).toBeVisible();
            await expect(section(page, 'Display name').getByText(SEEDED_USER.name)).toBeVisible();
            await expect(section(page, 'Email address').getByText(SEEDED_USER.mail)).toBeVisible();
            await expect(section(page, 'Password').getByText('••••••••••••')).toBeVisible();

            expectCleanPage(pageErrorLog);
        });
    });

    test.describe('password validation', () => {
        test.use({ cassette: accountCassette() });

        // The cassette deliberately holds no `PUT /user/password` entry: adding one to "fix" a
        // failure here would destroy the oracle.
        test('rejects each invalid password without touching the network', async ({ page, pageErrorLog }) => {
            const attempts: string[] = [];

            page.on('request', (request) => {
                if (request.method() === 'PUT' && new URL(request.url()).pathname === '/api/v1/user/password') {
                    attempts.push(request.url());
                }
            });

            await page.goto('/settings/account');
            await openSection(page, 'Password');

            // Only the first row clicks: the submit disables itself once a submitted form is invalid,
            // and every later row is then validated live on change.
            await fillPasswordForm(page, 'NewPass1!abc', 'OtherPass1!abc');
            await submitPasswordForm(page);
            await expect(page.getByText("Passwords don't match")).toBeVisible();

            await fillPasswordForm(page, 'newpass1');
            await expect(page.getByText('Password must be either longer than 15 characters')).toBeVisible();

            await fillPasswordForm(page, 'short1!');
            await expect(page.getByText('Password must be at least 8 characters')).toBeVisible();

            await fillPasswordForm(page, `${MAX_LENGTH_PASSWORD}x`);
            await expect(page.getByText('Password must not exceed 72 characters')).toBeVisible();

            await fillPasswordForm(page, CURRENT_PASSWORD);
            await expect(page.getByText('New password must be different from current password')).toBeVisible();

            expect(attempts, 'no invalid password may reach the backend').toEqual([]);
            expectCleanPage(pageErrorLog);
        });
    });

    test.describe('password submit', () => {
        test.use({
            cassette: accountCassette({
                'PUT /api/v1/user/password': [
                    {
                        body: { data: {}, status: 'success' },
                        bodySubset: {
                            confirm_password: MAX_LENGTH_PASSWORD,
                            current_password: CURRENT_PASSWORD,
                            password: MAX_LENGTH_PASSWORD,
                        },
                    },
                ],
            }),
        });

        test('sends the three snake_case fields and closes the section', async ({ page, pageErrorLog }) => {
            await page.goto('/settings/account');
            await openSection(page, 'Password');
            await fillPasswordForm(page, MAX_LENGTH_PASSWORD);
            await submitPasswordForm(page);

            await expect(page.getByText('Password successfully changed')).toBeVisible();
            await expect(section(page, 'Password').getByText('••••••••••••')).toBeVisible();

            expectCleanPage(pageErrorLog);
        });
    });

    test.describe('password rejection', () => {
        test.use({
            cassette: accountCassette({
                'PUT /api/v1/user/password': [
                    {
                        body: {
                            code: 'Users.ChangePasswordCurrentUser.InvalidCurrentPassword',
                            msg: 'invalid current password',
                            status: 'error',
                        },
                        status: 403,
                    },
                ],
            }),
        });

        test('translates the backend error code instead of showing the raw message', async ({ page, pageErrorLog }) => {
            await page.goto('/settings/account');
            await openSection(page, 'Password');
            await fillPasswordForm(page, 'NewPass1!abc');
            await submitPasswordForm(page);

            await expect(page.getByText('Current password is incorrect')).toBeVisible();
            await expect(page.getByText('invalid current password')).toBeHidden();
            await expect(page.getByRole('button', { name: 'Update Password' })).toBeVisible();

            expect(pageErrorLog.pageErrors, 'no uncaught page errors').toEqual([]);
            expect(
                pageErrorLog.consoleErrors.filter((text) => !text.includes('403')),
                'the rejected request is the only console error',
            ).toEqual([]);
        });
    });

    test.describe('name and email', () => {
        test.use({
            cassette: accountCassette({
                'PUT /api/v1/user/email': [
                    {
                        body: { data: {}, status: 'success' },
                        bodySubset: { current_password: CURRENT_PASSWORD, mail: 'renamed@pentagi.com' },
                    },
                ],
                'PUT /api/v1/user/name': [
                    { body: { data: {}, status: 'success' }, bodySubset: { name: 'E2E renamed admin' } },
                ],
            }),
        });

        test('sends the display name and the credential-guarded email change', async ({ page, pageErrorLog }) => {
            await page.goto('/settings/account');

            const nameCard = await openSection(page, 'Display name');

            await nameCard.getByLabel('Display name').fill('E2E renamed admin');
            await nameCard.getByRole('button', { name: 'Update Name' }).click();
            await expect(page.getByText('Name successfully updated')).toBeVisible();

            const emailCard = await openSection(page, 'Email address');

            await emailCard.getByLabel('Current Password').fill(CURRENT_PASSWORD);
            await emailCard.getByLabel('New Email').fill('renamed@pentagi.com');
            await emailCard.getByRole('button', { name: 'Update Email' }).click();
            await expect(page.getByText('Email successfully updated')).toBeVisible();

            expectCleanPage(pageErrorLog);
        });
    });
});
