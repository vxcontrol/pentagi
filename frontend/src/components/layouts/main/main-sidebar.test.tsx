import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes, useLocation } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const LOCAL_USER = { mail: 'me@example.com', name: 'Test User', type: 'local' };
const auth = vi.hoisted(() => ({
    value: { user: { mail: 'me@example.com', name: 'Test User', type: 'local' } } as { user: Record<string, unknown> },
}));

vi.mock('@/providers/user-provider', () => ({
    useUser: () => ({ authInfo: auth.value, logout: vi.fn() }),
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
            build: 'b1d7c0de',
            checkedAt: null,
            current: '2.1.0',
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

    return <span data-slot="probe-origin">{(location.state as null | { from?: string })?.from ?? 'none'}</span>;
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

beforeEach(() => {
    auth.value = { user: { ...LOCAL_USER } };
});

describe('MainSidebar account badge', () => {
    const openMenu = async (name: RegExp) => {
        const user = userEvent.setup();

        renderSidebar();
        await user.click(screen.getByRole('button', { name }));
    };

    it('calls a local account local', async () => {
        await openMenu(/Test User/);

        expect(screen.getByText('Local')).toBeInTheDocument();
    });

    it('names the provider a federated account came from', async () => {
        auth.value = { user: { mail: 'gh@example.com', name: 'GH', provider: 'github', type: 'oauth' } };
        await openMenu(/GH/);

        expect(screen.getByText('GitHub')).toBeInTheDocument();
    });

    it('falls back to the kind of account when the record names no provider', async () => {
        auth.value = { user: { mail: 'x@example.com', name: 'X', type: 'oauth' } };
        await openMenu(/x@example.com/);

        expect(screen.getByText('OAuth')).toBeInTheDocument();
    });
});

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

        expect(screen.getByTestId('probe-origin')).toHaveTextContent('/dashboard');
    });

    it('the Profile menu item carries the current path as the return origin', async () => {
        const user = userEvent.setup();
        renderSidebar();

        await user.click(screen.getByRole('button', { name: /Test User/ }));
        await user.click(screen.getByRole('menuitem', { name: 'Profile' }));

        expect(screen.getByTestId('probe-origin')).toHaveTextContent('/dashboard');
    });
});

describe('MainSidebar rows that carry an action', () => {
    it('marks the row so its label keeps clear of the action button', () => {
        renderSidebar();

        expect(screen.getByRole('link', { name: 'Flows' }).closest('li')).toHaveAttribute('data-has-action', 'true');
    });

    it('leaves a row without an action unmarked, so its label spans the full width', () => {
        renderSidebar();

        expect(screen.getByRole('link', { name: 'Dashboard' }).closest('li')).not.toHaveAttribute('data-has-action');
    });
});
