import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const { put } = vi.hoisted(() => ({ put: vi.fn() }));

// Keep the real error-mapping helper (`resolveApiErrorMessage`) — only the network call is stubbed.
vi.mock('@/lib/axios', async (importOriginal) => {
    const actual = await importOriginal<typeof import('@/lib/axios')>();

    return { ...actual, api: { ...actual.api, put } };
});

vi.mock('sonner', () => ({ toast: { error: vi.fn(), success: vi.fn() } }));

import { PasswordChangeForm } from './password-change-form';

const apiError = (code: string, msg: string) => ({ response: { data: { code, msg, status: 'error' } } });

beforeEach(() => {
    put.mockReset().mockResolvedValue({ status: 'success' });
});

describe('PasswordChangeForm', () => {
    it('toggles password visibility via the InputPassword control', async () => {
        const user = userEvent.setup();
        render(<PasswordChangeForm />);

        const current = screen.getByPlaceholderText('Enter your current password') as HTMLInputElement;
        expect(current.type).toBe('password');

        const [showButton] = screen.getAllByRole('button', { name: 'Show password' });

        if (!showButton) {
            throw new Error('expected a Show password toggle');
        }

        await user.click(showButton);

        expect(current.type).toBe('text');
    });

    it('submits the snake_case payload — proves RHF ref/onChange survive the InputPassword hop', async () => {
        const user = userEvent.setup();
        const onSuccess = vi.fn();
        render(<PasswordChangeForm onSuccess={onSuccess} />);

        await user.type(screen.getByPlaceholderText('Enter your current password'), 'Oldpass0!');
        await user.type(screen.getByPlaceholderText('Enter your new password'), 'Abcdef1!gh');
        await user.type(screen.getByPlaceholderText('Confirm your new password'), 'Abcdef1!gh');
        await user.click(screen.getByRole('button', { name: 'Update Password' }));

        await waitFor(() => expect(onSuccess).toHaveBeenCalledOnce());
        expect(put).toHaveBeenCalledWith('/user/password', {
            confirm_password: 'Abcdef1!gh',
            current_password: 'Oldpass0!',
            password: 'Abcdef1!gh',
        });
    });

    it('renders Skip only with onSkip, and puts submit before skip in the vertical layout', () => {
        const { rerender } = render(<PasswordChangeForm />);
        expect(screen.queryByRole('button', { name: 'Skip for now' })).not.toBeInTheDocument();

        rerender(
            <PasswordChangeForm
                buttonSize="default"
                layout="vertical"
                onSkip={vi.fn()}
            />,
        );

        const submit = screen.getByRole('button', { name: 'Update Password' });
        const skip = screen.getByRole('button', { name: 'Skip for now' });
        expect(submit.compareDocumentPosition(skip) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
    });

    it('maps a backend error code to friendly copy instead of the raw msg', async () => {
        const user = userEvent.setup();
        put.mockRejectedValueOnce(
            apiError('Users.ChangePasswordCurrentUser.InvalidCurrentPassword', 'invalid current password'),
        );
        render(<PasswordChangeForm />);

        await user.type(screen.getByPlaceholderText('Enter your current password'), 'Oldpass0!');
        await user.type(screen.getByPlaceholderText('Enter your new password'), 'Abcdef1!gh');
        await user.type(screen.getByPlaceholderText('Confirm your new password'), 'Abcdef1!gh');
        await user.click(screen.getByRole('button', { name: 'Update Password' }));

        expect(await screen.findByText('Current password is incorrect')).toBeInTheDocument();
        expect(screen.queryByText('invalid current password')).not.toBeInTheDocument();
    });
});
