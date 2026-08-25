import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { SidebarProvider } from '@/components/ui/sidebar';

const { api, authState } = vi.hoisted(() => ({
    api: {
        delete: vi.fn(),
        get: vi.fn(),
        post: vi.fn(),
        put: vi.fn(),
    },
    authState: { value: null as unknown },
}));

vi.mock('@/lib/axios', async (importOriginal) => {
    const actual = await importOriginal<typeof import('@/lib/axios')>();

    return { ...actual, api };
});
vi.mock('@/providers/user-provider', () => ({ useUser: () => ({ authInfo: authState.value }) }));

import SettingsUsers from './settings-users';

// The page renders inside the settings shell, which supplies the sidebar and router context.
const renderPage = () =>
    render(
        <MemoryRouter>
            <SidebarProvider>
                <SettingsUsers />
            </SidebarProvider>
        </MemoryRouter>,
    );

const admin = {
    hash: 'bb000000000000000000000000000001',
    id: 1,
    mail: 'admin@pentagi.com',
    name: 'admin',
    password_change_required: false,
    role_id: 1,
    status: 'active' as const,
    type: 'local' as const,
};

const ssoUser = {
    created_at: '2026-08-25T10:00:00Z',
    hash: 'aa000000000000000000000000000001',
    id: 2,
    mail: 'sso-user@example.com',
    name: 'sso-user',
    password_change_required: false,
    provider: 'oidc',
    role_id: 2,
    status: 'active' as const,
    type: 'oauth' as const,
};

const roles = [
    { id: 1, name: 'Admin' },
    { id: 2, name: 'User' },
];

const setAuth = (privileges: string[]) => {
    authState.value = {
        privileges,
        role: roles[0],
        type: 'user',
        user: { ...admin, created_at: '2026-08-01T00:00:00Z' },
    };
};

beforeEach(() => {
    vi.clearAllMocks();
    api.get.mockImplementation((url: string) =>
        url.startsWith('/users/')
            ? Promise.resolve({
                  data: { total: 2, users: [{ ...admin, created_at: '2026-08-01T00:00:00Z' }, ssoUser] },
                  status: 'success',
              })
            : Promise.resolve({ data: { roles, total: 2 }, status: 'success' }),
    );
    api.put.mockResolvedValue({ data: ssoUser, status: 'success' });
    api.delete.mockResolvedValue({ data: {}, status: 'success' });
});

describe('SettingsUsers access control', () => {
    it('refuses to show anything without the users.view privilege', async () => {
        setAuth(['flows.view']);
        renderPage();

        expect(await screen.findByText('Not available')).toBeInTheDocument();
        expect(api.get).not.toHaveBeenCalled();
    });

    it('lists users for an account holding users.view', async () => {
        setAuth(['users.view']);
        renderPage();

        expect(await screen.findByText('sso-user@example.com')).toBeInTheDocument();
        expect(screen.getByText('admin@pentagi.com')).toBeInTheDocument();
        // The sign-in column shows where the account comes from.
        expect(screen.getByText('oidc')).toBeInTheDocument();
    });

    it('keeps roles read-only without users.edit', async () => {
        setAuth(['users.view']);
        renderPage();

        await screen.findByText('sso-user@example.com');

        expect(screen.queryByRole('combobox', { name: 'Role of sso-user@example.com' })).not.toBeInTheDocument();
        expect(screen.getAllByText('User').length).toBeGreaterThan(0);
    });
});

describe('SettingsUsers role management', () => {
    it('offers a role selector for other users and saves the change', async () => {
        const user = userEvent.setup();
        setAuth(['users.view', 'users.edit']);
        renderPage();

        await screen.findByText('sso-user@example.com');

        const roleSelect = screen.getByRole('combobox', { name: 'Role of sso-user@example.com' });
        await user.click(roleSelect);
        await user.click(await screen.findByRole('option', { name: 'Admin' }));

        await waitFor(() =>
            expect(api.put).toHaveBeenCalledWith(
                `/users/${ssoUser.hash}`,
                expect.objectContaining({ hash: ssoUser.hash, role_id: 1 }),
            ),
        );
    });

    it('never offers to change your own role', async () => {
        setAuth(['users.view', 'users.edit']);
        renderPage();

        await screen.findByText('admin@pentagi.com');

        expect(screen.queryByRole('combobox', { name: 'Role of admin@pentagi.com' })).not.toBeInTheDocument();
        expect(screen.getByRole('combobox', { name: 'Role of sso-user@example.com' })).toBeInTheDocument();
    });

    it('hides row actions when the account may neither edit nor delete', async () => {
        setAuth(['users.view']);
        renderPage();

        await screen.findByText('sso-user@example.com');

        expect(screen.queryByRole('button', { name: 'Actions for sso-user@example.com' })).not.toBeInTheDocument();
    });

    it('hides the create action without users.create', async () => {
        setAuth(['users.view', 'users.edit']);
        renderPage();

        await screen.findByText('sso-user@example.com');

        expect(screen.queryByRole('button', { name: /add user/i })).not.toBeInTheDocument();
    });

    it('shows the create action for users.create', async () => {
        setAuth(['users.view', 'users.create']);
        renderPage();

        await screen.findByText('sso-user@example.com');

        expect(screen.getByRole('button', { name: /add user/i })).toBeInTheDocument();
    });
});
