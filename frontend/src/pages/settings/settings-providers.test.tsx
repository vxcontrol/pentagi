import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { format } from 'date-fns';
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

    it('dates a provider from an earlier year without the time', () => {
        queryResult.current = {
            data: makeData([{ ...seeded[0], createdAt: '2025-06-15T12:00:00Z', updatedAt: '2025-06-16T12:00:00Z' }]),
            error: undefined,
            loading: false,
            refetch: () => {},
        };

        render(
            <MemoryRouter>
                <SettingsProviders />
            </MemoryRouter>,
        );

        expect(screen.getByText(format(new Date('2025-06-15T12:00:00Z'), 'd MMM yyyy'))).toBeInTheDocument();
        expect(screen.getByText(format(new Date('2025-06-16T12:00:00Z'), 'd MMM yyyy'))).toBeInTheDocument();
    });
});

describe('SettingsProviders agent configuration panel', () => {
    const withAgent = (simple: Record<string, unknown>) => [
        {
            agents: { simple },
            createdAt: '2026-01-15T00:00:00Z',
            id: '1',
            name: 'Panel Provider',
            type: ProviderType.Custom,
            updatedAt: '2026-01-15T00:00:00Z',
        },
    ];

    const expand = async () => {
        const user = userEvent.setup();
        render(
            <MemoryRouter>
                <SettingsProviders />
            </MemoryRouter>,
        );
        await user.click(screen.getByText('Panel Provider'));
    };

    it('spells out a boolean field instead of leaving the label bare', async () => {
        queryResult.current = {
            data: makeData(withAgent({ json: true, model: 'm' })),
            error: undefined,
            loading: false,
            refetch: () => {},
        };

        await expand();

        expect(screen.getByText('Json:').parentElement).toHaveTextContent('Json: yes');
    });

    it('keeps a zero-valued field visible so it reads as zero, not as unset', async () => {
        queryResult.current = {
            data: makeData(withAgent({ model: 'm', temperature: 0 })),
            error: undefined,
            loading: false,
            refetch: () => {},
        };

        await expand();

        expect(screen.getByText('Temperature:').parentElement).toHaveTextContent('Temperature: 0');
    });

    it('drops a price the catalogue never carried instead of printing it as a zero tariff', async () => {
        queryResult.current = {
            data: makeData(withAgent({ model: 'm', price: { cacheRead: 0, cacheWrite: 0, input: 2, output: 8 } })),
            error: undefined,
            loading: false,
            refetch: () => {},
        };

        await expand();

        expect(screen.queryByText('Price Cache Read:')).not.toBeInTheDocument();
        expect(screen.queryByText('Price Cache Write:')).not.toBeInTheDocument();
        expect(screen.getByText('Price Input:').parentElement).toHaveTextContent('Price Input: 2');
        expect(screen.getByText('Price Output:').parentElement).toHaveTextContent('Price Output: 8');
    });

    it('still omits a field the provider did not set', async () => {
        queryResult.current = {
            data: makeData(withAgent({ model: 'm', temperature: null, topP: undefined })),
            error: undefined,
            loading: false,
            refetch: () => {},
        };

        await expand();

        expect(screen.queryByText('Temperature:')).not.toBeInTheDocument();
        expect(screen.queryByText('Top P:')).not.toBeInTheDocument();
    });
});
