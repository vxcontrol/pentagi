import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const { patchUser, put, refreshAuthInfo } = vi.hoisted(() => ({
    patchUser: vi.fn(),
    put: vi.fn(),
    refreshAuthInfo: vi.fn().mockResolvedValue(undefined),
}));

// Keep the real error-mapping helper (`resolveApiErrorMessage`) — only the network call is stubbed.
vi.mock('@/lib/axios', async (importOriginal) => {
    const actual = await importOriginal<typeof import('@/lib/axios')>();

    return { ...actual, api: { ...actual.api, put } };
});

vi.mock('@/providers/user-provider', () => ({
    useUser: () => ({ patchUser, refreshAuthInfo }),
}));

vi.mock('sonner', () => ({ toast: { error: vi.fn(), success: vi.fn() } }));

import { EmailChangeForm } from './email-change-form';

const apiError = (code: string, msg: string) => ({ response: { data: { code, msg, status: 'error' } } });

beforeEach(() => {
    put.mockReset().mockResolvedValue({ status: 'success' });
    refreshAuthInfo.mockClear();
    patchUser.mockClear();
});

describe('EmailChangeForm', () => {
    it('submits the new email and refreshes auth before closing', async () => {
        const user = userEvent.setup();
        const onSuccess = vi.fn();
        render(<EmailChangeForm onSuccess={onSuccess} />);

        await user.type(screen.getByPlaceholderText('Enter your new email address'), 'New@Example.com');
        await user.type(screen.getByPlaceholderText('Enter your current password'), 'Oldpass0!');
        await user.click(screen.getByRole('button', { name: 'Update Email' }));

        await waitFor(() => expect(onSuccess).toHaveBeenCalledOnce());
        expect(put).toHaveBeenCalledWith('/user/email', { current_password: 'Oldpass0!', mail: 'new@example.com' });
        expect(patchUser).toHaveBeenCalledWith({ mail: 'new@example.com' });
        expect(refreshAuthInfo).toHaveBeenCalledOnce();
    });

    it('maps the email-already-exists code to friendly copy', async () => {
        const user = userEvent.setup();
        put.mockRejectedValueOnce(apiError('Users.ChangeEmailCurrentUser.EmailAlreadyExists', 'email already exists'));
        render(<EmailChangeForm />);

        await user.type(screen.getByPlaceholderText('Enter your new email address'), 'taken@example.com');
        await user.type(screen.getByPlaceholderText('Enter your current password'), 'Oldpass0!');
        await user.click(screen.getByRole('button', { name: 'Update Email' }));

        expect(await screen.findByText('Email address is already in use')).toBeInTheDocument();
    });
});
