import { render } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ProviderType } from '@/graphql/types';
import { routes } from '@/lib/routes';

const { navigate } = vi.hoisted(() => ({ navigate: vi.fn() }));

const state = vi.hoisted(() => ({ params: new URLSearchParams() }));

const setSearch = (search: string) => {
    state.params = new URLSearchParams(search);
};

const enabled = {
    anthropic: true,
    bedrock: true,
    custom: false,
    deepseek: true,
    gemini: true,
    glm: true,
    kimi: true,
    minimax: false,
    ollama: true,
    openai: true,
    qwen: true,
};

const emptyProvider = { agents: {} };

const userDefined = [
    {
        agents: {},
        createdAt: '',
        id: 'disabled-1',
        name: 'My MiniMax',
        type: ProviderType.Minimax,
        updatedAt: '',
    },
];

const settingsProviders = {
    default: {
        anthropic: emptyProvider,
        openai: emptyProvider,
    },
    enabled,
    models: { anthropic: [], minimax: [], openai: [] },
    userDefined,
};

// A stable `data` identity matters: the page's seeding effect lists `data` as a
// dependency, so a fresh object each render would loop it (Apollo returns a
// cached reference in production).
const queryResult = { data: { settingsProviders }, error: undefined, loading: false };

vi.mock('@apollo/client/react', () => ({
    useMutation: () => [vi.fn(), {}],
    useQuery: () => queryResult,
}));

// `useBlocker` throws outside a data router; the guard only needs a stable inert
// blocker for these render/guard tests.
vi.mock('react-router-dom', async (importOriginal) => ({
    ...(await importOriginal<typeof import('react-router-dom')>()),
    useBlocker: () => ({ proceed: undefined, reset: undefined, state: 'unblocked' }),
    useNavigate: () => navigate,
    useParams: () => ({ providerId: 'new' }),
    useSearchParams: () => [state.params, vi.fn()],
}));

vi.mock('@/hooks/use-breakpoint', () => ({
    useBreakpoint: () => ({ isDesktop: true, isMobile: false }),
}));

// AppHeader pulls in SidebarTrigger (needs a SidebarProvider context); stub the
// whole family so the guard effect under test renders without that scaffolding.
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

vi.mock('sonner', () => ({ toast: { error: vi.fn(), info: vi.fn(), success: vi.fn() } }));

import SettingsProvider from './settings-provider';

beforeEach(() => {
    navigate.mockClear();
    setSearch('');
});

describe('SettingsProvider create-form type guards', () => {
    it('bounces ?type= for a disabled type', () => {
        setSearch('type=minimax');
        render(<SettingsProvider />);

        expect(navigate).toHaveBeenCalledWith(routes.settings.providers, { replace: true });
    });

    it('bounces ?type= for an unknown type', () => {
        setSearch('type=not-a-real-provider');
        render(<SettingsProvider />);

        expect(navigate).toHaveBeenCalledWith(routes.settings.providers, { replace: true });
    });

    it('bounces ?id= cloning a provider whose type is disabled', () => {
        setSearch('id=disabled-1');
        render(<SettingsProvider />);

        expect(navigate).toHaveBeenCalledWith(routes.settings.providers, { replace: true });
    });

    it('renders the create form for an enabled ?type=', () => {
        setSearch('type=anthropic');
        render(<SettingsProvider />);

        expect(navigate).not.toHaveBeenCalled();
    });
});
