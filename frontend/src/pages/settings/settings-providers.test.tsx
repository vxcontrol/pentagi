import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const state = vi.hoisted(() => ({ enabled: {} as Record<string, boolean> | undefined }));

vi.mock('@apollo/client/react', () => ({
    useMutation: () => [vi.fn(), {}],
    useQuery: () => ({ data: { settingsProviders: { enabled: state.enabled } } }),
}));

vi.mock('react-router-dom', async (importOriginal) => ({
    ...(await importOriginal<typeof import('react-router-dom')>()),
    useNavigate: () => vi.fn(),
}));

import { SettingsProvidersHeader } from './settings-providers';

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

beforeEach(() => {
    state.enabled = Object.fromEntries(ALL_TYPES.map((type) => [type, type !== 'minimax' && type !== 'custom']));
});

describe('SettingsProvidersHeader create menu', () => {
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
        state.enabled = Object.fromEntries(ALL_TYPES.map((type) => [type, false]));
        const user = userEvent.setup();
        render(<SettingsProvidersHeader />);

        await user.click(screen.getByRole('button', { name: /create provider/i }));

        expect(screen.getByRole('menuitem', { name: /no available provider types/i })).toBeInTheDocument();
        expect(screen.queryByRole('menuitem', { name: /OpenAI/ })).not.toBeInTheDocument();
    });
});
