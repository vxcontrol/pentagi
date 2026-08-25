import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { SidebarProvider } from '@/components/ui/sidebar';

const { authState } = vi.hoisted(() => ({ authState: { privileges: [] as string[] } }));

vi.mock('@/providers/user-provider', () => ({
    useUser: () => ({ authInfo: { privileges: authState.privileges, type: 'user' } }),
}));

import { SettingsSidebar } from './settings-sidebar';

beforeEach(() => {
    authState.privileges = [];
});

function renderSidebar(entry: { pathname: string; state?: unknown }) {
    return render(
        <MemoryRouter initialEntries={[entry]}>
            <SidebarProvider>
                <SettingsSidebar />
            </SidebarProvider>
            <Routes>
                <Route
                    element={<div>account</div>}
                    path="/settings/account"
                />
                <Route
                    element={<div>providers</div>}
                    path="/settings/providers"
                />
            </Routes>
        </MemoryRouter>,
    );
}

const backToApp = () => screen.getByRole('link', { name: /Back to App/ });

describe('SettingsSidebar "Back to App"', () => {
    it('returns to the page the user came from', () => {
        renderSidebar({ pathname: '/settings/account', state: { from: '/dashboard' } });

        expect(backToApp()).toHaveAttribute('href', '/dashboard');
    });

    it('falls back to /flows when there is no origin', () => {
        renderSidebar({ pathname: '/settings/account' });

        expect(backToApp()).toHaveAttribute('href', '/flows');
    });

    it('keeps the origin after switching settings sub-tabs (which drop location.state)', async () => {
        const user = userEvent.setup();
        renderSidebar({ pathname: '/settings/account', state: { from: '/dashboard' } });

        await user.click(screen.getByRole('link', { name: 'Providers' }));

        expect(backToApp()).toHaveAttribute('href', '/dashboard');
    });
});

describe('SettingsSidebar privileged items', () => {
    it('hides Users from accounts without the users.view privilege', () => {
        renderSidebar({ pathname: '/settings/account' });

        expect(screen.queryByRole('link', { name: /Users/ })).not.toBeInTheDocument();
        expect(screen.getByRole('link', { name: /Account/ })).toBeInTheDocument();
    });

    it('shows Users once the account may view them', () => {
        authState.privileges = ['users.view'];
        renderSidebar({ pathname: '/settings/account' });

        expect(screen.getByRole('link', { name: /Users/ })).toHaveAttribute('href', '/settings/users');
    });
});
