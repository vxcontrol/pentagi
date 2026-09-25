import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';

vi.mock('@/providers/version-info-provider', () => ({
    useVersionInfo: () => ({
        isLoading: false,
        versionInfo: {
            build: 'b1d7c0de',
            checkedAt: null,
            current: '2.1.0',
            failedAt: null,
            latest: null,
            state: 'up_to_date',
            strategy: 'preview',
        },
    }),
}));

import { SidebarProvider } from '@/components/ui/sidebar';

import { SettingsSidebar } from './settings-sidebar';

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

describe('SettingsSidebar version', () => {
    it('names the section and the running version in the header', () => {
        renderSidebar({ pathname: '/settings/account' });

        const header = screen.getByRole('button', { name: /Settings v2\.1\.0/ });

        expect(header).toHaveTextContent('Settings');
        expect(header).toHaveTextContent('v2.1.0');
    });
});

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
