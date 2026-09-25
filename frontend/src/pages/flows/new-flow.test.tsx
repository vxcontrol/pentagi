import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const { createFlow } = vi.hoisted(() => ({ createFlow: vi.fn() }));

vi.mock('@/components/layouts/app/app-header', () => {
    const Pass = ({ children }: { children?: React.ReactNode }) => <div>{children}</div>;

    return { AppHeader: Pass, AppHeaderContent: Pass, AppHeaderTitle: Pass };
});

vi.mock('@/providers/flows-provider', () => ({
    useFlows: () => ({ createFlow, createFlowWithAssistant: vi.fn() }),
}));

vi.mock('@/providers/providers-provider', () => ({
    useProviders: () => ({
        providers: [{ name: 'openai', type: 'openai' }],
        selectedProvider: { name: 'openai', type: 'openai' },
        setSelectedProvider: vi.fn(),
    }),
}));

vi.mock('@/providers/system-settings-provider', () => ({
    useSystemSettings: () => ({ settings: { assistantUseAgents: false } }),
}));

vi.mock('@/providers/templates-provider', () => ({
    useTemplates: () => ({ templates: [] }),
}));

vi.mock('@/providers/resources-provider', () => ({
    useResources: () => ({ resources: [] }),
}));

vi.mock('@/features/resources/use-resources-upload', () => ({
    useResourcesUpload: () => ({ isUploading: false, uploadFiles: vi.fn() }),
}));

const { default: NewFlow } = await import('./new-flow');

const renderPage = () =>
    render(
        <MemoryRouter initialEntries={['/flows/new']}>
            <Routes>
                <Route
                    element={<NewFlow />}
                    path="/flows/new"
                />
                <Route
                    element={<p>flow page</p>}
                    path="/flows/:flowId"
                />
            </Routes>
        </MemoryRouter>,
    );

beforeEach(() => {
    createFlow.mockReset();
});

describe('NewFlow', () => {
    it('keeps the typed text when the flow is not created', async () => {
        const user = userEvent.setup({ delay: null });
        createFlow.mockResolvedValue(null);

        renderPage();

        await user.type(screen.getByRole('textbox'), 'scan 10.0.0.5');
        await user.click(screen.getByRole('button', { name: 'Submit' }));

        await waitFor(() => expect(createFlow).toHaveBeenCalledOnce());
        await waitFor(() => expect(screen.getByRole('textbox')).not.toBeDisabled());
        expect(screen.getByRole('textbox')).toHaveValue('scan 10.0.0.5');
    });

    it('opens the created flow', async () => {
        const user = userEvent.setup({ delay: null });
        createFlow.mockResolvedValue('42');

        renderPage();

        await user.type(screen.getByRole('textbox'), 'scan 10.0.0.5');
        await user.click(screen.getByRole('button', { name: 'Submit' }));

        expect(await screen.findByText('flow page')).toBeInTheDocument();
    });
});
