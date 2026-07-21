import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, useLocation, useNavigate } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const { get } = vi.hoisted(() => ({ get: vi.fn() }));

vi.mock('@/lib/axios', async (importOriginal) => {
    const actual = await importOriginal<typeof import('@/lib/axios')>();

    return { ...actual, api: { ...actual.api, get } };
});

vi.mock('sonner', () => ({ toast: { error: vi.fn(), info: vi.fn(), success: vi.fn() } }));

import { AUTH_STORAGE_KEY, UserProvider, useUser } from './user-provider';

const session = {
    expires_at: '2099-01-01T00:00:00Z',
    type: 'user',
    user: { mail: 'me@example.com', name: 'Me', type: 'local' },
};

function Probe() {
    const { authInfo, refreshAuthInfo } = useUser();
    const location = useLocation();
    const navigate = useNavigate();

    return (
        <div>
            <span data-testid="mail">{authInfo?.user?.mail ?? 'none'}</span>
            <span data-testid="path">{location.pathname}</span>
            <button onClick={() => void refreshAuthInfo()}>refresh</button>
            <button onClick={() => navigate('/dashboard')}>go</button>
        </div>
    );
}

function renderProvider(initialPath: string) {
    return render(
        <MemoryRouter initialEntries={[initialPath]}>
            <UserProvider>
                <Probe />
            </UserProvider>
        </MemoryRouter>,
    );
}

beforeEach(() => {
    localStorage.clear();
    get.mockReset().mockRejectedValue(new Error('network'));
    localStorage.setItem(AUTH_STORAGE_KEY, JSON.stringify(session));
});

describe('UserProvider transient /info handling', () => {
    it('keeps the session when refreshAuthInfo hits a transient /info failure', async () => {
        const user = userEvent.setup();
        // /oauth/result is public, so neither the navigation refresh nor the /login effect fires —
        // refreshAuthInfo runs only on our explicit click.
        renderProvider('/oauth/result');

        expect(await screen.findByText('me@example.com')).toBeInTheDocument();

        await user.click(screen.getByRole('button', { name: 'refresh' }));
        await waitFor(() => expect(get).toHaveBeenCalled());

        expect(screen.getByTestId('mail')).toHaveTextContent('me@example.com');
    });

    it('keeps the session and stays on the page when the navigation refresh fails transiently', async () => {
        const user = userEvent.setup();
        renderProvider('/oauth/result');

        expect(await screen.findByText('me@example.com')).toBeInTheDocument();

        await user.click(screen.getByRole('button', { name: 'go' }));
        await waitFor(() => expect(get).toHaveBeenCalled());

        expect(screen.getByTestId('path')).toHaveTextContent('/dashboard');
        expect(screen.getByTestId('mail')).toHaveTextContent('me@example.com');
    });
});
