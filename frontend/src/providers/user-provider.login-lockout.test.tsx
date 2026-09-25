import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const { get, post } = vi.hoisted(() => ({ get: vi.fn(), post: vi.fn() }));

vi.mock('@/lib/axios', async (importOriginal) => {
    const actual = await importOriginal<typeof import('@/lib/axios')>();

    return { ...actual, api: { ...actual.api, get, post } };
});

const { error } = vi.hoisted(() => ({ error: vi.fn() }));

vi.mock('sonner', () => ({ toast: { error, info: vi.fn(), success: vi.fn(), warning: vi.fn() } }));

import { LOGIN_LOCKED_MESSAGE, UserProvider, useUser } from './user-provider';

function Probe() {
    const { login } = useUser();

    return (
        <button
            onClick={() =>
                void login({ mail: 'me@example.com', password: 'whatever' }).then((result) => {
                    document.title = result.error ?? 'no error';
                })
            }
        >
            sign in
        </button>
    );
}

const rejectWith = (status: number, code?: string) =>
    Object.assign(new Error('request failed'), {
        response: { data: code ? { code, msg: 'refused', status: 'error' } : undefined, status },
    });

beforeEach(() => {
    localStorage.clear();
    error.mockReset();
    get.mockReset().mockRejectedValue(new Error('network'));
    post.mockReset();
});

describe('UserProvider login refusals', () => {
    it('names the lockout when the endpoint answers Auth.TooManyAttempts', async () => {
        const user = userEvent.setup();
        post.mockRejectedValue(rejectWith(429, 'Auth.TooManyAttempts'));

        render(
            <MemoryRouter initialEntries={['/oauth/result']}>
                <UserProvider>
                    <Probe />
                </UserProvider>
            </MemoryRouter>,
        );

        await user.click(screen.getByRole('button', { name: 'sign in' }));

        await waitFor(() => expect(document.title).toBe(LOGIN_LOCKED_MESSAGE));
        expect(error).toHaveBeenCalledWith(LOGIN_LOCKED_MESSAGE);
    });

    it('keeps the generic wording for a refusal that carries no code', async () => {
        const user = userEvent.setup();
        post.mockRejectedValue(rejectWith(401));

        render(
            <MemoryRouter initialEntries={['/oauth/result']}>
                <UserProvider>
                    <Probe />
                </UserProvider>
            </MemoryRouter>,
        );

        await user.click(screen.getByRole('button', { name: 'sign in' }));

        await waitFor(() => expect(document.title).toBe('Login failed. Please try again.'));
    });
});
