import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const SYSTEM_TYPE = 'primary_agent_system';
const HUMAN_TYPE = 'primary_agent_human';

const CUSTOM_SYSTEM_ID = 'user-prompt-1';

const defaults = {
    agents: {
        primaryAgent: {
            human: { template: 'human default', type: HUMAN_TYPE },
            system: { template: 'system default', type: SYSTEM_TYPE },
        },
    },
    tools: {},
};

const makeData = (userDefined: { id: string; type: string }[]) => ({
    settingsPrompts: { default: defaults, userDefined },
});

const queryResult = vi.hoisted(() => ({
    current: { data: undefined, error: undefined, loading: false, refetch: () => {} } as Record<string, unknown>,
}));

const { mutate } = vi.hoisted(() => ({ mutate: vi.fn(async () => ({ data: {} })) }));

vi.mock('@apollo/client/react', () => ({
    useMutation: () => [mutate, { loading: false }],
    useQuery: () => queryResult.current,
}));

vi.mock('react-router-dom', async (importOriginal) => ({
    ...(await importOriginal<typeof import('react-router-dom')>()),
    useNavigate: () => vi.fn(),
}));

vi.mock('@/components/layouts/app/app-header', () => {
    const Pass = ({ children }: { children?: React.ReactNode }) => <div>{children}</div>;

    return { AppHeader: Pass, AppHeaderContent: Pass, AppHeaderTitle: Pass };
});

import SettingsPrompts from './settings-prompts';

const openRowMenu = async (user: ReturnType<typeof userEvent.setup>) => {
    render(
        <MemoryRouter>
            <SettingsPrompts />
        </MemoryRouter>,
    );

    const triggers = screen.getAllByRole('button', { name: /open menu/i });

    await user.click(triggers[0] as HTMLElement);
};

describe('resetting a prompt to its default', () => {
    beforeEach(() => {
        mutate.mockClear();
    });

    it('offers no reset while every prompt of the row is still the default', async () => {
        queryResult.current = { data: makeData([]), error: undefined, loading: false, refetch: () => {} };

        await openRowMenu(userEvent.setup());

        expect(screen.getByRole('menuitem', { name: /edit/i })).toBeInTheDocument();
        expect(screen.queryByRole('menuitem', { name: 'Reset System' })).not.toBeInTheDocument();
        expect(screen.queryByRole('menuitem', { name: 'Reset All' })).not.toBeInTheDocument();
    });

    it('offers the reset once the row carries a custom prompt', async () => {
        queryResult.current = {
            data: makeData([{ id: CUSTOM_SYSTEM_ID, type: SYSTEM_TYPE }]),
            error: undefined,
            loading: false,
            refetch: () => {},
        };

        await openRowMenu(userEvent.setup());

        expect(screen.getByRole('menuitem', { name: 'Reset System' })).toBeInTheDocument();
        expect(screen.getByRole('menuitem', { name: 'Reset All' })).toBeInTheDocument();
        expect(
            screen.queryByRole('menuitem', { name: 'Reset Human' }),
            'the human prompt is still the default',
        ).not.toBeInTheDocument();
    });

    it('drops the custom prompt only after the confirmation is given', async () => {
        queryResult.current = {
            data: makeData([{ id: CUSTOM_SYSTEM_ID, type: SYSTEM_TYPE }]),
            error: undefined,
            loading: false,
            refetch: () => {},
        };
        const user = userEvent.setup();

        await openRowMenu(user);
        await user.click(screen.getByRole('menuitem', { name: 'Reset System' }));

        expect(screen.getByRole('dialog')).toHaveTextContent(/reset the system prompt/i);
        expect(mutate, 'the dialog is a question, not the action').not.toHaveBeenCalled();

        await user.click(screen.getByRole('button', { name: 'Reset' }));

        expect(mutate).toHaveBeenCalledWith(expect.objectContaining({ variables: { promptId: CUSTOM_SYSTEM_ID } }));
    });

    it('asks about both prompts when the whole row is reset', async () => {
        queryResult.current = {
            data: makeData([{ id: CUSTOM_SYSTEM_ID, type: SYSTEM_TYPE }]),
            error: undefined,
            loading: false,
            refetch: () => {},
        };
        const user = userEvent.setup();

        await openRowMenu(user);
        await user.click(screen.getByRole('menuitem', { name: 'Reset All' }));

        expect(screen.getByRole('dialog')).toHaveTextContent(/reset all prompts/i);
    });
});
