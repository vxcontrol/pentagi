import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { toast } from 'sonner';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ProviderType } from '@/graphql/types';
import { routes } from '@/lib/routes';

const { mutate, navigate } = vi.hoisted(() => ({ mutate: vi.fn(), navigate: vi.fn() }));

const state = vi.hoisted(() => ({ params: new URLSearchParams(), providerId: 'new' }));

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

const userDefined = [
    {
        agents: {},
        createdAt: '',
        id: 'disabled-1',
        name: 'My MiniMax',
        type: ProviderType.Minimax,
        updatedAt: '',
    },
    {
        agents: {},
        createdAt: '',
        id: 'edit-1',
        name: 'Persisted Name',
        type: ProviderType.Anthropic,
        updatedAt: '',
    },
];

// The create form seeds agents from the type's defaults, and bails when the type has no models —
// so a fixture with an empty catalogue would make the seeding tests vacuously green.
const agentDefaults = (model: string, temperature: number) => ({
    simple: { maxTokens: 1000, model, temperature },
});

const settingsProviders = {
    default: {
        anthropic: { agents: agentDefaults('claude-e2e', 0.2) },
        openai: { agents: agentDefaults('gpt-e2e', 0.7) },
    },
    enabled,
    models: {
        anthropic: [{ name: 'claude-e2e' }],
        minimax: [],
        openai: [{ name: 'gpt-e2e' }],
    },
    userDefined,
};

// A stable `data` identity matters: the page's seeding effect lists `data` as a
// dependency, so a fresh object each render would loop it (Apollo returns a
// cached reference in production).
const queryResult = { data: { settingsProviders }, error: undefined as Error | undefined, loading: false };

vi.mock('@apollo/client/react', () => ({
    useMutation: () => [mutate, {}],
    useQuery: () => queryResult,
}));

// `useBlocker` throws outside a data router; the guard only needs a stable inert
// blocker for these render/guard tests.
vi.mock('react-router-dom', async (importOriginal) => ({
    ...(await importOriginal<typeof import('react-router-dom')>()),
    useBlocker: () => ({ proceed: undefined, reset: undefined, state: 'unblocked' }),
    useNavigate: () => navigate,
    useParams: () => ({ providerId: state.providerId }),
    useSearchParams: () => [state.params, vi.fn()],
}));

vi.mock('@/hooks/use-breakpoint', () => ({
    useBreakpoint: () => ({ isDesktop: true, isMobile: false }),
}));

// AppHeader pulls in SidebarTrigger (needs a SidebarProvider context); stub the
// whole family so the guard effect under test renders without that scaffolding.
vi.mock('@/components/layouts/app/app-header', () => {
    const Pass = ({ children }: { children?: React.ReactNode }) => <div>{children}</div>;
    const Action = ({
        icon: _icon,
        label,
        loading: _loading,
        ...props
    }: React.ComponentProps<'button'> & {
        icon?: React.ReactNode;
        label?: React.ReactNode;
        loading?: boolean;
    }) => <button {...props}>{label}</button>;

    return {
        AppHeader: Pass,
        AppHeaderAction: Action,
        AppHeaderActions: Pass,
        AppHeaderContent: Pass,
        AppHeaderTitle: Pass,
    };
});

vi.mock('sonner', () => ({ toast: { error: vi.fn(), info: vi.fn(), success: vi.fn() } }));

import SettingsProvider from './settings-provider';

beforeEach(() => {
    mutate.mockClear();
    navigate.mockClear();
    vi.mocked(toast.error).mockClear();
    setSearch('');
    state.providerId = 'new';
    queryResult.data = { settingsProviders };
    queryResult.error = undefined;
    queryResult.loading = false;
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

    // cache-and-network means an error can arrive with cached data still present; the form must
    // survive it rather than flip to the full-page error screen.
    it('keeps the form on a refetch error while cached data is present', () => {
        setSearch('type=anthropic');
        queryResult.error = new Error('e2e induced refetch failure');
        render(<SettingsProvider />);

        expect(screen.queryByText('Error loading provider data')).not.toBeInTheDocument();
        expect(navigate).not.toHaveBeenCalled();
    });

    // The loading branch runs before the error branch, so it needs the same guard: a background
    // refetch reports loading:true with cached data and must not blank the form to the spinner.
    it('keeps the form on a background refetch while cached data is present', () => {
        setSearch('type=anthropic');
        queryResult.loading = true;
        render(<SettingsProvider />);

        expect(screen.queryByText('Loading provider data...')).not.toBeInTheDocument();
        expect(navigate).not.toHaveBeenCalled();
    });

    const expandAgent = () => {
        if (!screen.queryByLabelText('Temperature')) {
            fireEvent.click(screen.getByRole('button', { name: /Simple/ }));
        }
    };

    const temperatureInput = () => screen.getByLabelText('Temperature') as HTMLInputElement;

    it('preserves an in-flight agent edit on the create form across a background refetch', () => {
        setSearch('type=openai');
        const { rerender } = render(<SettingsProvider />);

        expandAgent();
        expect(temperatureInput().value).toBe('0.7');

        fireEvent.change(temperatureInput(), { target: { value: '1.5' } });

        // A refetch that carries no change leaves `data` referentially equal and is inert, so the
        // payload has to actually differ for this to exercise the seeding effect.
        queryResult.data = {
            settingsProviders: { ...settingsProviders, userDefined: [...userDefined] },
        };
        rerender(<SettingsProvider />);

        expect(temperatureInput().value).toBe('1.5');
    });

    it('still re-seeds the agents when the type changes mid-edit', () => {
        setSearch('type=openai');
        const { rerender } = render(<SettingsProvider />);

        expandAgent();
        fireEvent.change(temperatureInput(), { target: { value: '1.5' } });

        setSearch('type=anthropic');
        rerender(<SettingsProvider />);
        expandAgent();

        expect(temperatureInput().value).toBe('0.2');
    });

    it('preserves an in-flight edit across a background refetch', () => {
        state.providerId = 'edit-1';
        const { rerender } = render(<SettingsProvider />);

        const nameInput = screen.getByPlaceholderText('Enter provider name') as HTMLInputElement;

        expect(nameInput.value).toBe('Persisted Name');

        fireEvent.change(nameInput, { target: { value: 'My Unsaved Edit' } });

        queryResult.data = { settingsProviders };
        rerender(<SettingsProvider />);

        expect((screen.getByPlaceholderText('Enter provider name') as HTMLInputElement).value).toBe('My Unsaved Edit');
    });
});

describe('SettingsProvider save feedback', () => {
    const toggleAgent = () => fireEvent.click(screen.getByRole('button', { name: /Simple/ }));

    const saveButton = () => screen.getByRole('button', { name: 'Create' });

    const renderCreateForm = () => {
        setSearch('type=openai');
        render(<SettingsProvider />);
        fireEvent.change(screen.getByPlaceholderText('Enter provider name'), { target: { value: 'My Provider' } });
    };

    const typeInvalidExtraBody = () => {
        toggleAgent();
        fireEvent.change(screen.getByLabelText('Extra Body (JSON)'), { target: { value: '{ nope' } });
        toggleAgent();
    };

    it('toasts an agent-level error raised inside a collapsed panel', async () => {
        renderCreateForm();
        typeInvalidExtraBody();

        expect(screen.queryByLabelText('Extra Body (JSON)')).not.toBeInTheDocument();

        fireEvent.click(saveButton());

        await waitFor(() => expect(toast.error).toHaveBeenCalledTimes(1));
        expect(toast.error).toHaveBeenCalledWith(expect.stringContaining('Must be a valid JSON object'));
        expect(mutate).not.toHaveBeenCalled();
    });

    it('toasts again when the same invalid form is submitted twice', async () => {
        renderCreateForm();
        typeInvalidExtraBody();

        fireEvent.click(saveButton());
        await waitFor(() => expect(toast.error).toHaveBeenCalledTimes(1));

        fireEvent.click(saveButton());
        await waitFor(() => expect(toast.error).toHaveBeenCalledTimes(2));
    });

    it('runs the create mutation when the form is valid', async () => {
        renderCreateForm();

        fireEvent.click(saveButton());

        await waitFor(() => expect(mutate).toHaveBeenCalledTimes(1));
        expect(toast.error).not.toHaveBeenCalled();
    });
});
