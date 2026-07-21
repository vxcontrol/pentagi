import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const { navigate, userApi } = vi.hoisted(() => ({
    navigate: vi.fn(),
    userApi: {
        authInfo: null as unknown,
        isAuthenticated: vi.fn(() => false),
        login: vi.fn(
            async () => ({ success: true }) as { error?: string; passwordChangeRequired?: boolean; success: boolean },
        ),
        loginWithOAuth: vi.fn(async () => ({ success: true })),
        setAuth: vi.fn(),
    },
}));

vi.mock('react-router-dom', () => ({ useNavigate: () => navigate }));
vi.mock('@/providers/user-provider', () => ({ useUser: () => userApi }));

import LoginForm from './login-form';

const renderLogin = () =>
    render(
        <LoginForm
            providers={[]}
            returnUrl="/after-login"
        />,
    );

const fillCredentials = async (user: ReturnType<typeof userEvent.setup>) => {
    await user.type(screen.getByPlaceholderText('Enter your email'), 'admin@example.com');
    await user.type(screen.getByPlaceholderText('Enter your password'), 'secret-pass');
};

beforeEach(() => {
    vi.clearAllMocks();
    userApi.authInfo = null;
    userApi.isAuthenticated.mockReturnValue(false);
    userApi.login.mockResolvedValue({ success: true });
});

describe('LoginForm validation convention', () => {
    it('stays silent until the first submit, even after typing and blurring', async () => {
        const user = userEvent.setup();
        renderLogin();

        expect(screen.queryByText('Login is required')).not.toBeInTheDocument();
        expect(screen.queryByText('Password is required')).not.toBeInTheDocument();

        await user.type(screen.getByPlaceholderText('Enter your email'), 'x');
        await user.tab();

        expect(screen.queryByText('Invalid login')).not.toBeInTheDocument();
        expect(screen.queryByText('Password is required')).not.toBeInTheDocument();
    });

    it('surfaces field errors on submit and does not call login', async () => {
        const user = userEvent.setup();
        renderLogin();

        await user.click(screen.getByRole('button', { name: 'Sign in' }));

        expect(await screen.findByText('Login is required')).toBeInTheDocument();
        expect(screen.getByText('Password is required')).toBeInTheDocument();
        expect(userApi.login).not.toHaveBeenCalled();
    });

    it('re-validates live once submitted: fixing a field clears its error', async () => {
        const user = userEvent.setup();
        renderLogin();

        await user.click(screen.getByRole('button', { name: 'Sign in' }));
        expect(await screen.findByText('Login is required')).toBeInTheDocument();

        await user.type(screen.getByPlaceholderText('Enter your email'), 'admin@example.com');

        await waitFor(() => expect(screen.queryByText('Login is required')).not.toBeInTheDocument());
    });

    it('calls login with the entered credentials and navigates on success', async () => {
        const user = userEvent.setup();
        renderLogin();

        await fillCredentials(user);
        await user.click(screen.getByRole('button', { name: 'Sign in' }));

        await waitFor(() =>
            expect(userApi.login).toHaveBeenCalledWith({ mail: 'admin@example.com', password: 'secret-pass' }),
        );
        expect(navigate).toHaveBeenCalledWith('/after-login');
    });

    it('shows the server error and stays on the form when login fails', async () => {
        const user = userEvent.setup();
        userApi.login.mockResolvedValue({ error: 'Account locked', success: false });
        renderLogin();

        await fillCredentials(user);
        await user.click(screen.getByRole('button', { name: 'Sign in' }));

        expect(await screen.findByText('Account locked')).toBeInTheDocument();
        expect(navigate).not.toHaveBeenCalled();
    });
});
