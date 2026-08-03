import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ProviderType } from '@/graphql/types';

const ALL_TYPES = [
    'anthropic',
    'bedrock',
    'custom',
    'deepseek',
    'gemini',
    'glm',
    'kimi',
    'minimax',
    'ollama',
    'openai',
    'qwen',
];

const emptyProvider = { agents: {} };

const makeData = (userDefined: unknown[]) => ({
    settingsProviders: {
        default: { anthropic: emptyProvider, openai: emptyProvider },
        enabled: Object.fromEntries(ALL_TYPES.map((type) => [type, type !== 'minimax' && type !== 'custom'])),
        models: {},
        userDefined,
    },
});

const queryResult = vi.hoisted(() => ({
    current: { data: undefined, error: undefined, loading: false, refetch: () => {} } as Record<string, unknown>,
}));

vi.mock('@apollo/client/react', () => ({
    useMutation: () => [vi.fn(), {}],
    useQuery: () => queryResult.current,
}));

vi.mock('react-router-dom', async (importOriginal) => ({
    ...(await importOriginal<typeof import('react-router-dom')>()),
    useNavigate: () => vi.fn(),
}));

vi.mock('@/hooks/use-table-state', () => ({
    useTableState: () => ({ filter: '', pageIndex: 0, setFilter: vi.fn(), setPage: vi.fn() }),
}));

// AppHeader pulls in SidebarTrigger (needs a SidebarProvider context); stub the family so the
// list's load-state branches render without that scaffolding. SettingsProvidersHeader builds its
// own trigger from a plain Button, so this does not touch the create-menu tests.
vi.mock('@/components/layouts/app/app-header', () => {
    const Pass = ({ children }: { children?: React.ReactNode }) => <div>{children}</div>;

    return {
        AppHeader: Pass,
        AppHeaderAction: Pass,
        AppHeaderActions: Pass,
        AppHeaderContent: Pass,
        AppHeaderTitle: Pass,
    };
});

import SettingsProviders, { SettingsProvidersHeader } from './settings-providers';

describe('SettingsProvidersHeader create menu', () => {
    beforeEach(() => {
        queryResult.current = { data: makeData([]), error: undefined, loading: false, refetch: () => {} };
    });

    it('offers only provider types whose API key is configured', async () => {
        const user = userEvent.setup();
        render(<SettingsProvidersHeader />);

        await user.click(screen.getByRole('button', { name: /create provider/i }));

        expect(screen.queryByRole('menuitem', { name: /MiniMax/ })).not.toBeInTheDocument();
        expect(screen.queryByRole('menuitem', { name: /Custom/ })).not.toBeInTheDocument();
        expect(screen.getByRole('menuitem', { name: /Anthropic/ })).toBeInTheDocument();
        expect(screen.getByRole('menuitem', { name: /OpenAI/ })).toBeInTheDocument();
    });

    it('shows a placeholder, not an empty menu, when no type is enabled', async () => {
        queryResult.current = {
            data: {
                settingsProviders: {
                    ...makeData([]).settingsProviders,
                    enabled: Object.fromEntries(ALL_TYPES.map((type) => [type, false])),
                },
            },
            error: undefined,
            loading: false,
            refetch: () => {},
        };
        const user = userEvent.setup();
        render(<SettingsProvidersHeader />);

        await user.click(screen.getByRole('button', { name: /create provider/i }));

        expect(screen.getByRole('menuitem', { name: /no available provider types/i })).toBeInTheDocument();
        expect(screen.queryByRole('menuitem', { name: /OpenAI/ })).not.toBeInTheDocument();
    });
});

describe('SettingsProviders list load states', () => {
    const seeded = [
        {
            agents: {},
            createdAt: '2026-01-15T00:00:00Z',
            id: '1',
            name: 'Seeded Provider',
            type: ProviderType.Custom,
            updatedAt: '2026-01-15T00:00:00Z',
        },
    ];

    // cache-and-network flips loading true with cached rows still present; the table must survive
    // it rather than flip to the full-page spinner.
    it('keeps the populated table on a background refetch instead of flashing the loader', () => {
        queryResult.current = { data: makeData(seeded), error: undefined, loading: true, refetch: () => {} };

        render(
            <MemoryRouter>
                <SettingsProviders />
            </MemoryRouter>,
        );

        expect(screen.getByText('Seeded Provider')).toBeInTheDocument();
        expect(screen.queryByText('Loading providers...')).not.toBeInTheDocument();
    });
});
