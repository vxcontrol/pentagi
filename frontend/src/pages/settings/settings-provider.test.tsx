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
        openai: [{ name: 'gpt-e2e', reasoning: { efforts: ['low', 'high'], supported: true } }],
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

import SettingsProvider, { agentConfigSchema, MAX_REASONING_TOKENS } from './settings-provider';

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

describe('SettingsProvider test feedback', () => {
    const renderCreateForm = () => {
        setSearch('type=openai');
        render(<SettingsProvider />);
    };

    const testControls = () => {
        const controls = screen.getAllByRole('button', { name: 'Test' });

        expect(controls).toHaveLength(2);

        return controls;
    };

    const blankName = '\u2022 Name: Provider name is required';

    it('names the offending field the first time Test is pressed', async () => {
        renderCreateForm();

        const [headerTest] = testControls();

        fireEvent.click(headerTest!);

        await waitFor(() => expect(toast.error).toHaveBeenCalledTimes(1));
        expect(toast.error).toHaveBeenCalledWith(expect.stringContaining(blankName));
        expect(mutate).not.toHaveBeenCalled();
    });

    it("names the offending field the first time an agent's own Test is pressed", async () => {
        renderCreateForm();

        const [, agentTest] = testControls();

        fireEvent.click(agentTest!);

        await waitFor(() => expect(toast.error).toHaveBeenCalledTimes(1));
        expect(toast.error).toHaveBeenCalledWith(expect.stringContaining(blankName));
        expect(mutate).not.toHaveBeenCalled();
    });

    it('toasts again when Test is pressed twice on the same invalid form', async () => {
        renderCreateForm();

        const [headerTest] = testControls();

        fireEvent.click(headerTest!);
        await waitFor(() => expect(toast.error).toHaveBeenCalledTimes(1));

        fireEvent.click(headerTest!);
        await waitFor(() => expect(toast.error).toHaveBeenCalledTimes(2));

        expect(vi.mocked(toast.error).mock.calls[1]?.[0]).toBe(vi.mocked(toast.error).mock.calls[0]?.[0]);
    });

    it('says the same thing the save path says', async () => {
        renderCreateForm();

        const [headerTest] = testControls();

        fireEvent.click(headerTest!);
        await waitFor(() => expect(toast.error).toHaveBeenCalledTimes(1));

        const fromTest = vi.mocked(toast.error).mock.calls[0]?.[0];

        vi.mocked(toast.error).mockClear();
        fireEvent.click(screen.getByRole('button', { name: 'Create' }));
        await waitFor(() => expect(toast.error).toHaveBeenCalledTimes(1));

        expect(vi.mocked(toast.error).mock.calls[0]?.[0]).toBe(fromTest);
    });

    it("sends the form's Simple agent with an agent's own Test", async () => {
        renderCreateForm();
        fireEvent.change(screen.getByPlaceholderText('Enter provider name'), { target: { value: 'My Provider' } });

        const [, agentTest] = testControls();

        fireEvent.click(agentTest!);

        await waitFor(() => expect(mutate).toHaveBeenCalledTimes(1));
        expect(mutate).toHaveBeenCalledWith(
            expect.objectContaining({
                variables: expect.objectContaining({ simple: expect.objectContaining({ model: 'gpt-e2e' }) }),
            }),
        );
    });
});

describe('SettingsProvider name validation', () => {
    const submitWithName = (name: string) => {
        setSearch('type=openai');
        render(<SettingsProvider />);
        fireEvent.change(screen.getByPlaceholderText('Enter provider name'), { target: { value: name } });
        fireEvent.click(screen.getByRole('button', { name: 'Create' }));
    };

    const expectRefused = async (message: string) => {
        await waitFor(() => expect(toast.error).toHaveBeenCalledWith(expect.stringContaining(message)));
        expect(mutate).not.toHaveBeenCalled();
    };

    const expectSaved = async (name: string) => {
        await waitFor(() => expect(mutate).toHaveBeenCalledTimes(1));
        expect(mutate).toHaveBeenCalledWith(expect.objectContaining({ variables: expect.objectContaining({ name }) }));
        expect(toast.error).not.toHaveBeenCalled();
    };

    it('refuses a blank name', async () => {
        submitWithName('');

        await expectRefused('Provider name is required');
    });

    it('refuses a name that is only whitespace', async () => {
        submitWithName('   ');

        await expectRefused('Provider name is required');
    });

    it('refuses a name one character over the 50-character cap', async () => {
        submitWithName('a'.repeat(51));

        await expectRefused('Maximum 50 characters allowed');
    });

    it('saves a name at the 50-character boundary', async () => {
        submitWithName('a'.repeat(50));

        await expectSaved('a'.repeat(50));
    });

    it('counts astral name characters once, the way the endpoint does', async () => {
        submitWithName('\u{1f512}'.repeat(50));

        await expectSaved('\u{1f512}'.repeat(50));
    });

    it('refuses an astral name one character over the cap', async () => {
        submitWithName('\u{1f512}'.repeat(51));

        await expectRefused('Maximum 50 characters allowed');
    });
});

describe('agentConfigSchema model', () => {
    const parse = (model: string) => agentConfigSchema.safeParse({ model });

    it('rejects a model made of whitespace', () => {
        expect(parse('   ').success).toBe(false);
        expect(parse('\t').success).toBe(false);
        expect(parse('\n').success).toBe(false);
    });

    it('keeps a named model', () => {
        expect(parse('gpt-5').success).toBe(true);
    });
});

describe('agentConfigSchema reasoning budget', () => {
    const parse = (maxTokens: number) => agentConfigSchema.safeParse({ model: 'gpt-5', reasoning: { maxTokens } });

    it('accepts the budget the endpoint allows', () => {
        expect(parse(MAX_REASONING_TOKENS).success).toBe(true);
    });

    it('rejects one token past it', () => {
        expect(parse(MAX_REASONING_TOKENS + 1).success).toBe(false);
    });

    it('keeps a stored provider loadable at the old client cap', () => {
        expect(parse(32001).success).toBe(true);
    });

    const parseMode = (reasoning: Record<string, unknown>) =>
        agentConfigSchema.safeParse({ model: 'gpt-5', reasoning });

    it('rejects budget mode left without a budget', () => {
        expect(parseMode({ mode: 'budget' }).success).toBe(false);
        expect(parseMode({ maxTokens: null, mode: 'budget' }).success).toBe(false);
        expect(parseMode({ effort: 'low', mode: 'budget' }).success).toBe(false);
    });

    it('accepts budget mode once a budget is set', () => {
        expect(parseMode({ maxTokens: 4096, mode: 'budget' }).success).toBe(true);
    });

    it('leaves the other reasoning modes free of a budget', () => {
        expect(parseMode({ mode: 'off' }).success).toBe(true);
        expect(parseMode({ mode: 'adaptive' }).success).toBe(true);
    });
});

describe('SettingsProvider model catalogue', () => {
    const customCatalogue = {
        ...settingsProviders,
        default: { ...settingsProviders.default, custom: { agents: agentDefaults('openai/gpt-x', 0.5) } },
        enabled: { ...enabled, custom: true },
        models: {
            ...settingsProviders.models,
            custom: [
                { name: 'openai/gpt-x' },
                { name: 'openai/gpt-x' },
                { name: 'openai/gpt-x' },
                { name: 'anthropic/claude-x' },
                { name: 'anthropic/claude-x' },
            ],
        },
    };

    const openModelList = () => {
        if (!screen.queryByLabelText('Temperature')) {
            fireEvent.click(screen.getByRole('button', { name: /Simple/ }));
        }

        fireEvent.click(screen.getByRole('button', { name: /open model list/i }));
    };

    beforeEach(() => {
        setSearch('type=custom');
        queryResult.data = { settingsProviders: customCatalogue };
    });

    it('offers a repeated model name once', async () => {
        render(<SettingsProvider />);
        openModelList();

        await waitFor(() => expect(screen.getAllByRole('option').length).toBeGreaterThan(0));

        const names = screen.getAllByRole('option').map((option) => option.textContent ?? '');
        const repeated = names.filter((name) => name.includes('openai/gpt-x'));
        const claude = names.filter((name) => name.includes('anthropic/claude-x'));

        expect(repeated).toHaveLength(1);
        expect(claude).toHaveLength(1);
    });

    it('keeps every distinct model the catalogue offers', async () => {
        render(<SettingsProvider />);
        openModelList();

        await waitFor(() => expect(screen.getAllByRole('option').length).toBeGreaterThan(0));

        expect(screen.getAllByRole('option')).toHaveLength(2);
    });

    const searchModels = (value: string) => {
        render(<SettingsProvider />);
        openModelList();
        fireEvent.change(screen.getByPlaceholderText('Search model...'), { target: { value } });
    };

    it('offers no custom model for a search made of whitespace', async () => {
        searchModels('   ');

        await waitFor(() => expect(screen.getByText(/no model found/i)).toBeInTheDocument());
        expect(screen.queryByRole('button', { name: /as custom model/i })).not.toBeInTheDocument();
    });

    it('carries a padded custom model to the payload trimmed', async () => {
        searchModels('  openai/gpt-new  ');

        fireEvent.click(await screen.findByRole('button', { name: /as custom model/i }));
        fireEvent.change(screen.getByPlaceholderText('Enter provider name'), { target: { value: 'My Provider' } });
        fireEvent.click(screen.getByRole('button', { name: 'Create' }));

        await waitFor(() => expect(mutate).toHaveBeenCalledTimes(1));

        expect(mutate).toHaveBeenCalledWith(
            expect.objectContaining({
                variables: expect.objectContaining({
                    agents: expect.objectContaining({ simple: expect.objectContaining({ model: 'openai/gpt-new' }) }),
                }),
            }),
        );
    });
});

describe('SettingsProvider reasoning effort levels', () => {
    const openSimpleAgent = () => fireEvent.click(screen.getByRole('button', { name: /Simple/ }));

    it('offers no effort control for a model whose catalogue declares none', () => {
        setSearch(`?type=${ProviderType.Anthropic}`);
        render(<SettingsProvider />);
        openSimpleAgent();

        expect(screen.getByText('Reasoning Configuration')).toBeInTheDocument();
        expect(screen.queryByText('Reasoning Effort')).toBeNull();
    });

    it('offers the control when the catalogue declares levels', () => {
        setSearch(`?type=${ProviderType.Openai}`);
        render(<SettingsProvider />);
        openSimpleAgent();

        expect(screen.getByText('Reasoning Effort')).toBeInTheDocument();
    });
});

describe('SettingsProvider minimal effort', () => {
    const catalogue = {
        ...settingsProviders,
        default: {
            ...settingsProviders.default,
            openai: {
                agents: {
                    simple: { maxTokens: 1000, model: 'gpt-e2e', reasoning: { effort: 'minimal' }, temperature: 0.7 },
                },
            },
        },
        models: {
            ...settingsProviders.models,
            openai: [{ name: 'gpt-e2e', reasoning: { efforts: ['minimal', 'low'], supported: true } }],
        },
    };

    beforeEach(() => {
        setSearch(`?type=${ProviderType.Openai}`);
        queryResult.data = { settingsProviders: catalogue };
        render(<SettingsProvider />);
        fireEvent.change(screen.getByPlaceholderText('Enter provider name'), { target: { value: 'My Provider' } });
    });

    it('carries the level the vendor documents through to the payload', async () => {
        fireEvent.click(screen.getByRole('button', { name: 'Create' }));

        await waitFor(() => expect(mutate).toHaveBeenCalledTimes(1));

        const [call] = mutate.mock.calls;

        expect(call?.[0].variables.agents.simple.reasoning.effort).toBe('minimal');
    });
});

describe('SettingsProvider reasoning mode against extra body', () => {
    const typeExtraBody = (value: string) =>
        fireEvent.change(screen.getByLabelText('Extra Body (JSON)'), { target: { value } });

    beforeEach(() => {
        setSearch(`?type=${ProviderType.Openai}`);
        render(<SettingsProvider />);
        fireEvent.click(screen.getByRole('button', { name: /Simple/ }));
    });

    it('explains the mode is overwritten when extra body pins thinking', () => {
        typeExtraBody('{"thinking":{"type":"enabled"}}');

        expect(screen.getByText(/Extra Body sets "thinking"/)).toBeInTheDocument();
    });

    it('names the qwen spelling of the same conflict', () => {
        typeExtraBody('{"enable_thinking":true}');

        expect(screen.getByText(/Extra Body sets "enable_thinking"/)).toBeInTheDocument();
    });

    it('stays quiet for an extra body that leaves thinking alone', () => {
        typeExtraBody('{"tool_choice":"auto"}');

        expect(screen.queryByText(/overwrites this choice/)).toBeNull();
    });

    it('stays quiet for a thinking object that only carries vendor keys', () => {
        typeExtraBody('{"thinking":{"clear_thinking":false}}');

        expect(screen.queryByText(/overwrites this choice/)).toBeNull();
    });

    it('stays quiet while the json is still half-typed', () => {
        typeExtraBody('{"thinking":');

        expect(screen.queryByText(/overwrites this choice/)).toBeNull();
    });
});

describe('SettingsProvider effort across a model change', () => {
    const catalogue = {
        ...settingsProviders,
        default: {
            ...settingsProviders.default,
            openai: {
                agents: {
                    simple: { maxTokens: 1000, model: 'gpt-e2e', reasoning: { effort: 'high' }, temperature: 0.7 },
                },
            },
        },
        models: {
            ...settingsProviders.models,
            openai: [
                { name: 'gpt-e2e', reasoning: { efforts: ['low', 'high'], supported: true } },
                { name: 'gpt-keeps-high', reasoning: { efforts: ['high'], supported: true } },
                { name: 'gpt-low-only', reasoning: { efforts: ['low'], supported: true } },
            ],
        },
    };

    const pickModel = async (name: string) => {
        fireEvent.click(screen.getByRole('button', { name: /Simple/ }));
        fireEvent.click(screen.getByRole('button', { name: /open model list/i }));

        await waitFor(() => expect(screen.getAllByRole('option').length).toBeGreaterThan(0));

        fireEvent.click(screen.getByRole('option', { name: new RegExp(name) }));
    };

    const savedEffort = async () => {
        fireEvent.click(screen.getByRole('button', { name: 'Create' }));

        await waitFor(() => expect(mutate).toHaveBeenCalledTimes(1));

        const [call] = mutate.mock.calls;

        return call?.[0].variables.agents.simple.reasoning.effort;
    };

    beforeEach(() => {
        setSearch('type=openai');
        queryResult.data = { settingsProviders: catalogue };
        render(<SettingsProvider />);
        fireEvent.change(screen.getByPlaceholderText('Enter provider name'), { target: { value: 'My Provider' } });
    });

    it('keeps the level when the new model declares it', async () => {
        await pickModel('gpt-keeps-high');

        expect(await savedEffort()).toBe('high');
    });

    it('clears the level the new model does not declare', async () => {
        await pickModel('gpt-low-only');

        expect(await savedEffort()).toBeNull();
    });
});

describe('SettingsProvider output ceiling across a model change', () => {
    const catalogue = {
        ...settingsProviders,
        default: {
            ...settingsProviders.default,
            openai: {
                agents: {
                    simple: { maxTokens: 8000, model: 'gpt-e2e', temperature: 0.7 },
                },
            },
        },
        models: {
            ...settingsProviders.models,
            openai: [
                { name: 'gpt-e2e', reasoning: { efforts: ['low'], supported: true } },
                { maxOutputTokens: 4096, name: 'gpt-narrow', reasoning: { efforts: ['low'], supported: true } },
                { maxOutputTokens: 16384, name: 'gpt-wide', reasoning: { efforts: ['low'], supported: true } },
                { name: 'gpt-silent', reasoning: { efforts: ['low'], supported: true } },
            ],
        },
    };

    const pickModel = async (name: string) => {
        fireEvent.click(screen.getByRole('button', { name: /Simple/ }));
        fireEvent.click(screen.getByRole('button', { name: /open model list/i }));

        await waitFor(() => expect(screen.getAllByRole('option').length).toBeGreaterThan(0));

        fireEvent.click(screen.getByRole('option', { name: new RegExp(name) }));
    };

    const savedMaxTokens = async () => {
        fireEvent.click(screen.getByRole('button', { name: 'Create' }));

        await waitFor(() => expect(mutate).toHaveBeenCalledTimes(1));

        const [call] = mutate.mock.calls;

        return call?.[0].variables.agents.simple.maxTokens;
    };

    beforeEach(() => {
        setSearch('type=openai');
        queryResult.data = { settingsProviders: catalogue };
        render(<SettingsProvider />);
        fireEvent.change(screen.getByPlaceholderText('Enter provider name'), { target: { value: 'My Provider' } });
    });

    it('lowers a value the new model cannot accept', async () => {
        await pickModel('gpt-narrow');

        expect(await savedMaxTokens()).toBe(4096);
    });

    it('keeps a value that fits under the new ceiling', async () => {
        await pickModel('gpt-wide');

        expect(await savedMaxTokens()).toBe(8000);
    });

    it('keeps the value when the model publishes no ceiling', async () => {
        await pickModel('gpt-silent');

        expect(await savedMaxTokens()).toBe(8000);
    });
});

describe('SettingsProvider reasoning on the ollama catalogue', () => {
    const ollamaCatalogue = (model: string, mode?: string) => ({
        ...settingsProviders,
        default: {
            ...settingsProviders.default,
            ollama: {
                agents: {
                    simple: { maxTokens: 1000, model, temperature: 0.7, ...(mode ? { reasoning: { mode } } : {}) },
                },
            },
        },
        models: {
            ...settingsProviders.models,
            ollama: [
                { name: 'deepseek-r1:8b', reasoning: { cannotDisable: false, supported: true }, thinking: true },
                { name: 'gpt-oss:120b', reasoning: { cannotDisable: true, supported: true }, thinking: true },
            ],
        },
    });

    const renderWith = (model: string, mode?: string) => {
        setSearch(`?type=${ProviderType.Ollama}`);
        queryResult.data = { settingsProviders: ollamaCatalogue(model, mode) };
        render(<SettingsProvider />);
        fireEvent.change(screen.getByPlaceholderText('Enter provider name'), { target: { value: 'My Ollama' } });
        fireEvent.click(screen.getByRole('button', { name: /Simple/ }));
    };

    const savedMode = async () => {
        fireEvent.click(screen.getByRole('button', { name: 'Create' }));

        await waitFor(() => expect(mutate).toHaveBeenCalledTimes(1));

        const [call] = mutate.mock.calls;

        return call?.[0].variables.agents.simple.reasoning?.mode ?? null;
    };

    it('offers a reasoning mode for a model whose door can switch thinking off', () => {
        renderWith('deepseek-r1:8b');

        expect(screen.getByText('Reasoning Mode')).toBeInTheDocument();
    });

    it('offers no Off for a model that always thinks', async () => {
        renderWith('gpt-oss:120b');
        fireEvent.keyDown(screen.getByRole('combobox', { name: 'Reasoning Mode' }), { key: 'Enter' });

        const options = (await screen.findAllByRole('option')).map((option) => option.textContent);

        expect(options).not.toContain('Off (no thinking)');
    });

    it('keeps a saved Off the model can honour', async () => {
        renderWith('deepseek-r1:8b', 'off');

        expect(screen.queryByRole('status')).toBeNull();
        expect(await savedMode()).toBe('off');
    });

    it('says so when it resets an Off the model cannot honour', async () => {
        renderWith('gpt-oss:120b', 'off');

        expect(await screen.findByRole('status')).toHaveTextContent(/gpt-oss:120b cannot disable, so it was reset/);
        expect(await savedMode()).toBeNull();
    });

    it.each([
        ['Create', () => screen.getByRole('button', { name: 'Create' })],
        ['Test', () => screen.getAllByRole('button', { name: 'Test' })[0]],
        ['the agent Test', () => screen.getAllByRole('button', { name: 'Test' })[1]],
    ])('%s sends no Off the model cannot honour for an agent left collapsed', async (_, button) => {
        setSearch(`?type=${ProviderType.Ollama}`);
        queryResult.data = { settingsProviders: ollamaCatalogue('gpt-oss:120b', 'off') };
        render(<SettingsProvider />);
        fireEvent.change(screen.getByPlaceholderText('Enter provider name'), { target: { value: 'My Ollama' } });
        fireEvent.click(button() as HTMLElement);

        await waitFor(() => expect(mutate).toHaveBeenCalledTimes(1));

        const [call] = mutate.mock.calls;
        const sent = call?.[0].variables.agent ?? call?.[0].variables.agents.simple;

        expect(sent?.reasoning?.mode).toBeNull();
    });

    it('sends a collapsed Off the model can honour', async () => {
        setSearch(`?type=${ProviderType.Ollama}`);
        queryResult.data = { settingsProviders: ollamaCatalogue('deepseek-r1:8b', 'off') };
        render(<SettingsProvider />);
        fireEvent.change(screen.getByPlaceholderText('Enter provider name'), { target: { value: 'My Ollama' } });

        expect(await savedMode()).toBe('off');
    });

    it('stops saying so once the model picked again can honour Off', async () => {
        const pick = async (model: string) => {
            fireEvent.click(screen.getByRole('button', { name: /open model list/i }));
            fireEvent.click(await screen.findByRole('option', { name: new RegExp(model) }));
        };

        renderWith('deepseek-r1:8b', 'off');
        await pick('gpt-oss:120b');

        expect(await screen.findByRole('status')).toHaveTextContent(/gpt-oss:120b cannot disable/);

        await pick('deepseek-r1:8b');

        await waitFor(() => expect(screen.queryByRole('status')).toBeNull());
    });
});

describe('SettingsProvider level refused alongside tools', () => {
    interface Agent {
        maxTokens: number;
        model: string;
        reasoning?: Record<string, unknown>;
        temperature: number;
    }

    const agent = (model: string, reasoning?: Record<string, unknown>): Agent => ({
        maxTokens: 1000,
        model,
        temperature: 0.7,
        ...(reasoning ? { reasoning } : {}),
    });

    const catalogue = (agents: { [key: string]: Agent; simple: Agent }) => ({
        ...settingsProviders,
        default: { ...settingsProviders.default, openai: { agents } },
        models: {
            ...settingsProviders.models,
            openai: [
                { name: 'gpt-e2e', reasoning: { efforts: ['low', 'high'], supported: true } },
                {
                    name: 'gpt-5.6-sol',
                    reasoning: {
                        cannotDisable: false,
                        efforts: ['low', 'medium', 'high', 'xhigh'],
                        rejectsEffortWithTools: true,
                        supported: true,
                    },
                    thinking: true,
                },
            ],
        },
    });

    const renderWith = (agents: { [key: string]: Agent; simple: Agent }, agentButton?: RegExp) => {
        setSearch(`?type=${ProviderType.Openai}`);
        queryResult.data = { settingsProviders: catalogue(agents) };
        render(<SettingsProvider />);
        fireEvent.change(screen.getByPlaceholderText('Enter provider name'), { target: { value: 'My OpenAI' } });

        if (agentButton) {
            fireEvent.click(screen.getByRole('button', { name: agentButton }));
        }
    };

    const savedReasoning = async (agent: string) => {
        fireEvent.click(screen.getByRole('button', { name: 'Create' }));

        await waitFor(() => expect(mutate).toHaveBeenCalledTimes(1));

        const [call] = mutate.mock.calls;

        return call?.[0].variables.agents[agent].reasoning;
    };

    const modeOptions = async () => {
        fireEvent.keyDown(screen.getByRole('combobox', { name: 'Reasoning Mode' }), { key: 'Enter' });

        return (await screen.findAllByRole('option')).map((option) => option.textContent);
    };

    it('offers neither an effort nor a budget to an agent that calls with tools', async () => {
        renderWith({ simple: agent('gpt-5.6-sol') }, /Simple/);

        expect(screen.queryByText('Reasoning Effort')).toBeNull();
        expect(screen.queryByLabelText('Reasoning Max Tokens')).toBeNull();
        expect(screen.getByText(/gpt-5.6-sol refuses a reasoning effort or budget/)).toBeInTheDocument();
        expect(await modeOptions()).toEqual(['Not selected', 'Off (no thinking)']);
    });

    it('resets a saved effort the model refuses and says so', async () => {
        renderWith({ simple: agent('gpt-5.6-sol', { effort: 'medium' }) }, /Simple/);

        expect(await screen.findByRole('status')).toHaveTextContent(
            'Reasoning Effort was Medium, which gpt-5.6-sol refuses on requests with function tools',
        );
        expect(await savedReasoning('simple')).toEqual({ effort: null, maxTokens: null, mode: null });
    });

    it.each([{ maxTokens: 2048 }, { maxTokens: 2048, mode: 'budget' }])(
        'resets a saved budget %o the model refuses and says so',
        async (reasoning) => {
            renderWith({ simple: agent('gpt-5.6-sol', reasoning) }, /Simple/);

            expect(await screen.findByRole('status')).toHaveTextContent('Reasoning Max Tokens was 2048');
            expect(await savedReasoning('simple')).toEqual({ effort: null, maxTokens: null, mode: null });
        },
    );

    const testButtons = () => screen.getAllByRole('button', { name: 'Test' });

    it.each([
        ['Create', { effort: 'medium' }, () => screen.getByRole('button', { name: 'Create' })],
        ['Create', { maxTokens: 2048, mode: 'budget' }, () => screen.getByRole('button', { name: 'Create' })],
        ['Test', { effort: 'medium' }, () => testButtons()[0]],
        ['the agent Test', { effort: 'medium' }, () => testButtons()[1]],
    ])('%s sends no refused level %o for an agent left collapsed', async (_, reasoning, button) => {
        renderWith({ simple: agent('gpt-5.6-sol', reasoning) });
        fireEvent.click(button() as HTMLElement);

        await waitFor(() => expect(mutate).toHaveBeenCalledTimes(1));

        const [call] = mutate.mock.calls;
        const sent = call?.[0].variables.agent ?? call?.[0].variables.agents.simple;

        expect(sent?.reasoning).toEqual({ effort: null, maxTokens: null, mode: null });
    });

    it('shows a saved effort on a model the catalogue does not list, so it can be cleared', async () => {
        renderWith({ simple: agent('gpt-5.5-2026-04-23', { effort: 'high' }) }, /Simple/);

        const effort = screen.getByRole('combobox', { name: 'Reasoning Effort' });

        expect(effort).toHaveTextContent('High');

        fireEvent.keyDown(effort, { key: 'Enter' });
        fireEvent.click(await screen.findByRole('option', { name: 'Not selected' }));

        expect((await savedReasoning('simple')).effort).toBeNull();
    });

    it('keeps an Off the model can honour', async () => {
        renderWith({ simple: agent('gpt-5.6-sol', { mode: 'off' }) }, /Simple/);

        expect(screen.queryByRole('status')).toBeNull();
        expect((await savedReasoning('simple')).mode).toBe('off');
    });

    it('drops the level when the model picked refuses it', async () => {
        renderWith({ simple: agent('gpt-e2e', { effort: 'high' }) }, /Simple/);
        fireEvent.click(screen.getByRole('button', { name: /open model list/i }));
        fireEvent.click(await screen.findByRole('option', { name: /gpt-5.6-sol/ }));

        expect((await savedReasoning('simple')).effort).toBeNull();
    });

    it.each([
        ['primaryAgent', /Primary Agent/],
        ['assistant', /Assistant/],
        ['generator', /Generator/],
        ['refiner', /Refiner/],
        ['searcher', /Searcher/],
        ['enricher', /Enricher/],
        ['coder', /Coder/],
        ['installer', /Installer/],
        ['pentester', /Pentester/],
    ])('refuses the effort to %s, whose calls carry tools', async (key, button) => {
        renderWith({ [key]: agent('gpt-5.6-sol', { effort: 'medium' }), simple: agent('gpt-e2e') }, button);

        expect(screen.queryByText('Reasoning Effort')).toBeNull();
        expect(await screen.findByRole('status')).toHaveTextContent(
            'Reasoning Effort was Medium, which gpt-5.6-sol refuses on requests with function tools',
        );
        expect(await savedReasoning(key)).toEqual({ effort: null, maxTokens: null, mode: null });
    });

    it.each([
        ['simpleJson', /Simple Json/],
        ['adviser', /Adviser/],
        ['reflector', /Reflector/],
    ])('leaves the effort to %s, whose calls carry no tools', async (key, button) => {
        renderWith({ [key]: agent('gpt-5.6-sol', { effort: 'medium' }), simple: agent('gpt-e2e') }, button);

        expect(screen.getByText('Reasoning Effort')).toBeInTheDocument();
        expect(screen.queryByRole('status')).toBeNull();
        expect((await savedReasoning(key)).effort).toBe('medium');
    });
});

describe('SettingsProvider reasoning on a model that does not reason', () => {
    const catalogue = {
        ...settingsProviders,
        default: {
            ...settingsProviders.default,
            ollama: { agents: { simple: { maxTokens: 1000, model: 'llama3.1:8b', temperature: 0.7 } } },
        },
        models: {
            ...settingsProviders.models,
            ollama: [{ name: 'llama3.1:8b', thinking: false }, { name: 'qwen3:8b' }],
        },
    };

    beforeEach(() => {
        setSearch(`?type=${ProviderType.Ollama}`);
        queryResult.data = { settingsProviders: catalogue };
        render(<SettingsProvider />);
        fireEvent.click(screen.getByRole('button', { name: /Simple/ }));
    });

    it('offers no reasoning budget and says why', () => {
        expect(screen.queryByLabelText('Reasoning Max Tokens')).toBeNull();
        expect(screen.getByText('llama3.1:8b does not reason, so it has no reasoning settings.')).toBeInTheDocument();
    });

    it('keeps the budget for a model whose catalogue entry says nothing about reasoning', async () => {
        fireEvent.click(screen.getByRole('button', { name: /open model list/i }));
        fireEvent.click(await screen.findByRole('option', { name: /qwen3:8b/ }));

        expect(await screen.findByLabelText('Reasoning Max Tokens')).toBeInTheDocument();
        expect(screen.queryByText(/does not reason/)).toBeNull();
    });
});

describe('SettingsProvider saved effort on a model that does not reason', () => {
    const catalogue = {
        ...settingsProviders,
        default: {
            ...settingsProviders.default,
            ollama: {
                agents: {
                    simple: { maxTokens: 1000, model: 'llama3.1:8b', reasoning: { effort: 'high' }, temperature: 0.7 },
                },
            },
        },
        models: { ...settingsProviders.models, ollama: [{ name: 'llama3.1:8b', thinking: false }] },
    };

    it('stays hidden, since the model has no reasoning settings', () => {
        setSearch(`?type=${ProviderType.Ollama}`);
        queryResult.data = { settingsProviders: catalogue };
        render(<SettingsProvider />);
        fireEvent.click(screen.getByRole('button', { name: /Simple/ }));

        expect(screen.getByText('llama3.1:8b does not reason, so it has no reasoning settings.')).toBeInTheDocument();
        expect(screen.queryByText('Reasoning Effort')).toBeNull();
    });
});

describe('SettingsProvider saved level on a listed model that does not reason', () => {
    const catalogue = {
        ...settingsProviders,
        default: {
            ...settingsProviders.default,
            deepseek: {
                agents: {
                    adviser: {
                        maxTokens: 1000,
                        model: 'deepseek-chat',
                        reasoning: { effort: 'high' },
                        temperature: 0.7,
                    },
                    pentester: {
                        maxTokens: 1000,
                        model: 'deepseek-chat',
                        reasoning: { effort: 'high' },
                        temperature: 0.7,
                    },
                    simple: {
                        maxTokens: 1000,
                        model: 'deepseek-chat',
                        reasoning: { maxTokens: 2048 },
                        temperature: 0.7,
                    },
                },
            },
        },
        models: { ...settingsProviders.models, deepseek: [{ name: 'deepseek-chat', thinking: false }] },
    };

    const savedAgents = async () => {
        setSearch(`?type=${ProviderType.Deepseek}`);
        queryResult.data = { settingsProviders: catalogue };
        render(<SettingsProvider />);
        fireEvent.change(screen.getByPlaceholderText('Enter provider name'), { target: { value: 'My DeepSeek' } });
        fireEvent.click(screen.getByRole('button', { name: 'Create' }));

        await waitFor(() => expect(mutate).toHaveBeenCalledTimes(1));

        return mutate.mock.calls[0]?.[0].variables.agents;
    };

    it('sends no hidden level for an agent that calls with tools', async () => {
        const agents = await savedAgents();

        expect(agents.pentester.reasoning).toEqual({ effort: null, maxTokens: null, mode: null });
        expect(agents.simple.reasoning).toEqual({ effort: null, maxTokens: null, mode: null });
    });

    it('leaves the level to an agent whose calls carry no tools', async () => {
        const agents = await savedAgents();

        expect(agents.adviser.reasoning.effort).toBe('high');
    });
});

describe('SettingsProvider Off reset on a model that does not reason', () => {
    const catalogue = {
        ...settingsProviders,
        default: {
            ...settingsProviders.default,
            ollama: {
                agents: {
                    simple: { maxTokens: 1000, model: 'llama3.1:8b', reasoning: { mode: 'off' }, temperature: 0.7 },
                },
            },
        },
        models: {
            ...settingsProviders.models,
            ollama: [
                { name: 'gpt-oss:120b', reasoning: { cannotDisable: true, supported: true }, thinking: true },
                { name: 'llama3.1:8b', thinking: false },
            ],
        },
    };

    beforeEach(() => {
        setSearch(`?type=${ProviderType.Ollama}`);
        queryResult.data = { settingsProviders: catalogue };
        render(<SettingsProvider />);
        fireEvent.click(screen.getByRole('button', { name: /Simple/ }));
    });

    it('gives the model not reasoning as the reason', async () => {
        const notice = await screen.findByRole('status');

        expect(notice).toHaveTextContent('Reasoning Mode was Off, but llama3.1:8b does not reason, so it was reset.');
        expect(notice).not.toHaveTextContent('cannot disable');
    });

    it('keeps saying the model cannot disable where it thinks', async () => {
        fireEvent.click(screen.getByRole('button', { name: /open model list/i }));
        fireEvent.click(await screen.findByRole('option', { name: /gpt-oss:120b/ }));

        expect(await screen.findByRole('status')).toHaveTextContent('gpt-oss:120b cannot disable, so it was reset');
    });
});

describe('SettingsProvider reasoning budget follows the mode', () => {
    const catalogue = (reasoning?: Record<string, unknown>) => ({
        ...settingsProviders,
        default: {
            ...settingsProviders.default,
            ollama: {
                agents: {
                    simple: {
                        maxTokens: 1000,
                        model: 'claude-budget',
                        temperature: 0.7,
                        ...(reasoning ? { reasoning } : {}),
                    },
                },
            },
        },
        models: {
            ...settingsProviders.models,
            ollama: [
                {
                    name: 'claude-budget',
                    reasoning: { cannotDisable: false, mode: 'budget', supported: true },
                    thinking: true,
                },
                { name: 'gpt-oss:120b', reasoning: { cannotDisable: true, supported: true }, thinking: true },
            ],
        },
    });

    const renderWith = (reasoning?: Record<string, unknown>) => {
        setSearch(`?type=${ProviderType.Ollama}`);
        queryResult.data = { settingsProviders: catalogue(reasoning) };
        render(<SettingsProvider />);
        fireEvent.change(screen.getByPlaceholderText('Enter provider name'), { target: { value: 'My Ollama' } });
        fireEvent.click(screen.getByRole('button', { name: /Simple/ }));
    };

    const openModes = async () => {
        fireEvent.keyDown(screen.getByRole('combobox', { name: 'Reasoning Mode' }), { key: 'Enter' });

        return screen.findAllByRole('option');
    };

    const pickMode = async (name: string) => {
        await openModes();
        fireEvent.click(screen.getByRole('option', { name }));
    };

    const savedReasoning = async () => {
        fireEvent.click(screen.getByRole('button', { name: 'Create' }));

        await waitFor(() => expect(mutate).toHaveBeenCalledTimes(1));

        const [call] = mutate.mock.calls;

        return call?.[0].variables.agents.simple.reasoning;
    };

    it('drops the budget when the mode goes back to Not selected', async () => {
        renderWith({ maxTokens: 2048, mode: 'budget' });
        await pickMode('Not selected');

        expect(screen.getByLabelText('Reasoning Max Tokens')).toHaveValue(null);
        expect(await savedReasoning()).toEqual({ effort: null, maxTokens: null, mode: null });
    });

    it('drops the budget when thinking is turned off', async () => {
        renderWith({ maxTokens: 2048, mode: 'budget' });
        await pickMode('Off (no thinking)');

        expect(await savedReasoning()).toEqual({ effort: null, maxTokens: null, mode: 'off' });
    });

    it('takes a budget only in Budget mode', async () => {
        renderWith();

        expect(screen.getByLabelText('Reasoning Max Tokens')).toBeDisabled();

        await pickMode('Budget');

        expect(screen.getByLabelText('Reasoning Max Tokens')).toBeEnabled();
    });

    it('offers Budget where the model thinks but cannot be switched off', async () => {
        renderWith();
        fireEvent.click(screen.getByRole('button', { name: /open model list/i }));
        fireEvent.click(await screen.findByRole('option', { name: /gpt-oss:120b/ }));

        expect((await openModes()).map((option) => option.textContent)).toEqual(['Not selected', 'Budget']);
    });
});

describe('SettingsProvider budget saved without a mode', () => {
    const catalogue = (reasoning: Record<string, unknown>) => ({
        ...settingsProviders,
        default: {
            ...settingsProviders.default,
            bedrock: {
                agents: {
                    simple: { maxTokens: 16384, model: 'claude-sonnet-budget', reasoning, temperature: 0.7 },
                },
            },
        },
        models: {
            ...settingsProviders.models,
            bedrock: [
                {
                    name: 'claude-sonnet-budget',
                    reasoning: { cannotDisable: false, efforts: ['low', 'high'], mode: 'budget', supported: true },
                    thinking: true,
                },
            ],
        },
    });

    const renderWith = (reasoning: Record<string, unknown>) => {
        setSearch(`?type=${ProviderType.Bedrock}`);
        queryResult.data = { settingsProviders: catalogue(reasoning) };
        render(<SettingsProvider />);
        fireEvent.change(screen.getByPlaceholderText('Enter provider name'), { target: { value: 'My Bedrock' } });
        fireEvent.click(screen.getByRole('button', { name: /Simple/ }));
    };

    const modeSelect = () => screen.getByRole('combobox', { name: 'Reasoning Mode' });

    it('opens as Budget, the mode the server runs it in', () => {
        renderWith({ maxTokens: 2048 });

        expect(modeSelect()).toHaveTextContent('Budget');
        expect(screen.getByLabelText('Reasoning Max Tokens')).toBeEnabled();
        expect(screen.getByLabelText('Reasoning Max Tokens')).toHaveValue(2048);
    });

    it('keeps the budget field open while its value is replaced', async () => {
        renderWith({ maxTokens: 2048 });

        const field = () => screen.getByLabelText('Reasoning Max Tokens');

        fireEvent.change(field(), { target: { value: '' } });

        expect(field()).toBeEnabled();
        expect(modeSelect()).toHaveTextContent('Budget');

        fireEvent.change(field(), { target: { value: '4096' } });
        fireEvent.click(screen.getByRole('button', { name: 'Create' }));

        await waitFor(() => expect(mutate).toHaveBeenCalledTimes(1));

        const [call] = mutate.mock.calls;

        expect(call?.[0].variables.agents.simple.reasoning).toEqual({ effort: null, maxTokens: 4096, mode: 'budget' });
    });

    it('opens as Not selected when an effort rides along, since the effort wins', () => {
        renderWith({ effort: 'low', maxTokens: 2048 });

        expect(modeSelect()).toHaveTextContent('Not selected');
    });

    const pick = async (combobox: string, option: string) => {
        fireEvent.keyDown(screen.getByRole('combobox', { name: combobox }), { key: 'Enter' });
        fireEvent.click(await screen.findByRole('option', { name: option }));
    };

    const created = async () => {
        fireEvent.click(screen.getByRole('button', { name: 'Create' }));

        await waitFor(() => expect(mutate).toHaveBeenCalledTimes(1));

        const [call] = mutate.mock.calls;

        return call?.[0].variables.agents.simple.reasoning;
    };

    it('leaves Budget for the effort picked, since the server runs the effort then', async () => {
        renderWith({ maxTokens: 2048 });
        await pick('Reasoning Effort', 'High');

        expect(modeSelect()).toHaveTextContent('Not selected');
        expect(await created()).toEqual({ effort: 'high', maxTokens: null, mode: null });
    });

    it('drops the effort when Budget is picked, since the server ignores it there', async () => {
        renderWith({ effort: 'low' });
        await pick('Reasoning Mode', 'Budget');
        fireEvent.change(screen.getByLabelText('Reasoning Max Tokens'), { target: { value: '4096' } });

        expect(screen.getByRole('combobox', { name: 'Reasoning Effort' })).toHaveTextContent('Not selected');
        expect(await created()).toEqual({ effort: null, maxTokens: 4096, mode: 'budget' });
    });

    it('shows no effort beside a saved Budget', async () => {
        renderWith({ effort: 'high', maxTokens: 2048, mode: 'budget' });

        expect(modeSelect()).toHaveTextContent('Budget');
        await waitFor(() =>
            expect(screen.getByRole('combobox', { name: 'Reasoning Effort' })).toHaveTextContent('Not selected'),
        );
    });

    it('goes to the wire without a budget once switched to Not selected', async () => {
        renderWith({ maxTokens: 2048 });
        fireEvent.keyDown(modeSelect(), { key: 'Enter' });
        fireEvent.click(await screen.findByRole('option', { name: 'Not selected' }));
        fireEvent.click(screen.getByRole('button', { name: 'Create' }));

        await waitFor(() => expect(mutate).toHaveBeenCalledTimes(1));

        const [call] = mutate.mock.calls;

        expect(call?.[0].variables.agents.simple.reasoning).toEqual({ effort: null, maxTokens: null, mode: null });
    });
});

describe('SettingsProvider reasoning on an adaptive-only model', () => {
    const catalogue = (cannotDisable: boolean, reasoning?: Record<string, unknown>) => ({
        ...settingsProviders,
        default: {
            ...settingsProviders.default,
            anthropic: {
                agents: {
                    simple: {
                        maxTokens: 16384,
                        model: 'claude-adaptive',
                        temperature: 0.7,
                        ...(reasoning ? { reasoning } : {}),
                    },
                },
            },
        },
        models: {
            ...settingsProviders.models,
            anthropic: [
                {
                    name: 'claude-adaptive',
                    reasoning: { cannotDisable, mode: 'adaptive_only', supported: true },
                    thinking: true,
                },
            ],
        },
    });

    const renderWith = (cannotDisable: boolean, reasoning?: Record<string, unknown>) => {
        setSearch(`?type=${ProviderType.Anthropic}`);
        queryResult.data = { settingsProviders: catalogue(cannotDisable, reasoning) };
        render(<SettingsProvider />);
        fireEvent.change(screen.getByPlaceholderText('Enter provider name'), { target: { value: 'My Anthropic' } });
        fireEvent.click(screen.getByRole('button', { name: /Simple/ }));
    };

    it('offers no Budget and no budget field where Off works', async () => {
        renderWith(false);

        expect(screen.queryByLabelText('Reasoning Max Tokens')).toBeNull();

        fireEvent.keyDown(screen.getByRole('combobox', { name: 'Reasoning Mode' }), { key: 'Enter' });

        expect((await screen.findAllByRole('option')).map((option) => option.textContent)).toEqual([
            'Adaptive',
            'Off (no thinking)',
        ]);
    });

    it('offers no budget field and locks the mode where thinking cannot be disabled', () => {
        renderWith(true);

        expect(screen.queryByLabelText('Reasoning Max Tokens')).toBeNull();
        expect(screen.getByRole('combobox', { name: 'Reasoning Mode' })).toBeDisabled();
        expect(screen.getByRole('combobox', { name: 'Reasoning Mode' })).toHaveTextContent('Adaptive');
    });

    it('does not turn a saved budget into Budget mode', async () => {
        renderWith(false, { maxTokens: 2048 });

        expect(screen.getByRole('combobox', { name: 'Reasoning Mode' })).toHaveTextContent('Adaptive');

        fireEvent.click(screen.getByRole('button', { name: 'Create' }));

        await waitFor(() => expect(mutate).toHaveBeenCalledTimes(1));

        const [call] = mutate.mock.calls;

        expect(call?.[0].variables.agents.simple.reasoning.mode).toBeNull();
    });
});

describe('SettingsProvider effort beside a saved Budget left collapsed', () => {
    const catalogue = {
        ...settingsProviders,
        default: {
            ...settingsProviders.default,
            bedrock: {
                agents: {
                    simple: {
                        maxTokens: 16384,
                        model: 'claude-sonnet-budget',
                        reasoning: { effort: 'high', maxTokens: 2048, mode: 'budget' },
                        temperature: 0.7,
                    },
                },
            },
        },
        models: {
            ...settingsProviders.models,
            bedrock: [
                {
                    name: 'claude-sonnet-budget',
                    reasoning: { cannotDisable: false, efforts: ['low', 'high'], mode: 'budget', supported: true },
                    thinking: true,
                },
            ],
        },
    };

    it('sends no effort, since the server runs the budget alone', async () => {
        setSearch(`?type=${ProviderType.Bedrock}`);
        queryResult.data = { settingsProviders: catalogue };
        render(<SettingsProvider />);
        fireEvent.change(screen.getByPlaceholderText('Enter provider name'), { target: { value: 'My Bedrock' } });
        fireEvent.click(screen.getByRole('button', { name: 'Create' }));

        await waitFor(() => expect(mutate).toHaveBeenCalledTimes(1));

        const [call] = mutate.mock.calls;

        expect(call?.[0].variables.agents.simple.reasoning).toEqual({ effort: null, maxTokens: 2048, mode: 'budget' });
    });
});

describe('SettingsProvider budget on a model that takes none', () => {
    const catalogue = (reasoning?: Record<string, unknown>) => ({
        ...settingsProviders,
        default: {
            ...settingsProviders.default,
            glm: {
                agents: {
                    simple: {
                        maxTokens: 8192,
                        model: 'glm-5-turbo',
                        temperature: 0.7,
                        ...(reasoning ? { reasoning } : {}),
                    },
                },
            },
        },
        models: {
            ...settingsProviders.models,
            glm: [
                {
                    name: 'glm-5-turbo',
                    reasoning: { cannotDisable: false, supported: true, takesNoThinkingDepth: true },
                    thinking: true,
                },
            ],
        },
    });

    const renderWith = (reasoning?: Record<string, unknown>, isOpened = true) => {
        setSearch(`?type=${ProviderType.Glm}`);
        queryResult.data = { settingsProviders: catalogue(reasoning) };
        render(<SettingsProvider />);
        fireEvent.change(screen.getByPlaceholderText('Enter provider name'), { target: { value: 'My GLM' } });

        if (isOpened) {
            fireEvent.click(screen.getByRole('button', { name: /Simple/ }));
        }
    };

    const created = async () => {
        fireEvent.click(screen.getByRole('button', { name: 'Create' }));

        await waitFor(() => expect(mutate).toHaveBeenCalledTimes(1));

        const [call] = mutate.mock.calls;

        return call?.[0].variables.agents.simple.reasoning;
    };

    it('offers neither Budget nor a budget field', async () => {
        renderWith();

        expect(screen.queryByLabelText('Reasoning Max Tokens')).toBeNull();

        fireEvent.keyDown(screen.getByRole('combobox', { name: 'Reasoning Mode' }), { key: 'Enter' });

        expect((await screen.findAllByRole('option')).map((option) => option.textContent)).toEqual([
            'Not selected',
            'Off (no thinking)',
        ]);
    });

    it.each([{ maxTokens: 2048 }, { maxTokens: 2048, mode: 'budget' }])(
        'resets a saved budget %o and says so',
        async (reasoning) => {
            renderWith(reasoning);

            expect(await screen.findByRole('status')).toHaveTextContent(
                'Reasoning Max Tokens was 2048, which glm-5-turbo does not take, so it was reset',
            );
            expect(await created()).toEqual({ effort: null, maxTokens: null, mode: null });
        },
    );

    it.each([{ maxTokens: 2048 }, { maxTokens: 2048, mode: 'budget' }])(
        'sends no budget %o for an agent left collapsed',
        async (reasoning) => {
            renderWith(reasoning, false);

            expect(await created()).toEqual({ effort: null, maxTokens: null, mode: null });
        },
    );

    it('keeps an Off the model can honour', async () => {
        renderWith({ mode: 'off' }, false);

        expect((await created()).mode).toBe('off');
    });
});
