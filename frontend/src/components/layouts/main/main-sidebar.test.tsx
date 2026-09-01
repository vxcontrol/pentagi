import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes, useLocation } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';

vi.mock('@/providers/user-provider', () => ({
    useUser: () => ({
        authInfo: { user: { mail: 'me@example.com', name: 'Test User', type: 'local' } },
        logout: vi.fn(),
    }),
}));
vi.mock('@/hooks/use-theme', () => ({ useTheme: () => ({ setTheme: vi.fn(), theme: 'system' }) }));
vi.mock('@/providers/favorites-provider', () => ({
    useFavorites: () => ({ addFavoriteFlow: vi.fn(), favoriteFlowIds: [], removeFavoriteFlow: vi.fn() }),
}));
vi.mock('@/providers/sidebar-flows-provider', () => ({ useSidebarFlows: () => ({ flows: [] }) }));
vi.mock('@/features/resources/use-resources-upload', () => ({
    useResourcesUpload: () => ({ fileInputKey: 'k', fileInputProps: {}, openFilePicker: vi.fn() }),
}));
vi.mock('@/providers/version-info-provider', () => ({
    useVersionInfo: () => ({
        isLoading: false,
        versionInfo: {
            checkedAt: null,
            current: '2.1.0-93e99748',
            failedAt: null,
            latest: '2.4.0',
            state: 'update_available',
            strategy: 'preview',
        },
    }),
}));

import { SidebarProvider } from '@/components/ui/sidebar';

import { MainSidebar } from './main-sidebar';

function FromProbe() {
    const location = useLocation();

    return <span data-testid="from">{(location.state as null | { from?: string })?.from ?? 'none'}</span>;
}

function renderSidebar() {
    return render(
        <MemoryRouter initialEntries={['/dashboard']}>
            <SidebarProvider>
                <MainSidebar />
            </SidebarProvider>
            <Routes>
                <Route
                    element={<div>dashboard</div>}
                    path="/dashboard"
                />
                <Route
                    element={<FromProbe />}
                    path="/settings"
                />
                <Route
                    element={<FromProbe />}
                    path="/settings/account"
                />
            </Routes>
        </MemoryRouter>,
    );
}

describe('MainSidebar version badge', () => {
    it('shows the running version under the product name, with the update indicator', () => {
        renderSidebar();

        expect(screen.getByRole('button', { name: /PentAGI v2\.1\.0/ })).toHaveTextContent('v2.1.0');
        expect(screen.getByTestId('version-indicator-update')).toBeInTheDocument();
        expect(screen.getByTestId('version-attention-dot')).toBeInTheDocument();
    });
});

describe('MainSidebar settings entry points', () => {
    it('the Settings link carries the current path as the return origin', async () => {
        const user = userEvent.setup();
        renderSidebar();

        await user.click(screen.getByRole('link', { name: 'Settings' }));

        expect(screen.getByTestId('from')).toHaveTextContent('/dashboard');
    });

    it('the Profile menu item carries the current path as the return origin', async () => {
        const user = userEvent.setup();
        renderSidebar();

        await user.click(screen.getByRole('button', { name: /Test User/ }));
        await user.click(screen.getByRole('menuitem', { name: 'Profile' }));

        expect(screen.getByTestId('from')).toHaveTextContent('/dashboard');
    });
});
