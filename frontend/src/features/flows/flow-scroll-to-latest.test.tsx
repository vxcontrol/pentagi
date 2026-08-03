import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';

import { TooltipProvider } from '@/components/ui/tooltip';
import { StatusType } from '@/graphql/types';

const flowState = vi.hoisted(() => ({ current: {} as Record<string, unknown> }));

vi.mock('@/hooks/use-auto-scroll', () => ({
    useAutoScroll: () => ({
        containerRef: () => {},
        endRef: { current: null },
        hasNewMessages: false,
        isScrolledToBottom: false,
        scrollToEnd: vi.fn(),
    }),
}));

vi.mock('@/providers/flow-provider', () => ({ useFlow: () => flowState.current }));
vi.mock('@/providers/providers-provider', () => ({ useProviders: () => ({ providers: [] }) }));
vi.mock('@/providers/system-settings-provider', () => ({ useSystemSettings: () => ({ settings: null }) }));

vi.mock('./flow-form', () => ({ FlowForm: () => <div /> }));
vi.mock('./flow-tasks-dropdown', () => ({ default: () => <div /> }));
vi.mock('./agents/flow-agent', () => ({ default: () => <div /> }));
vi.mock('./tools/flow-tool', () => ({ default: () => <div /> }));
vi.mock('./vector-stores/flow-vector-store', () => ({ default: () => <div /> }));
vi.mock('./tasks/flow-task', () => ({ default: () => <div /> }));
vi.mock('./screenshots/flow-screenshot', () => ({ default: () => <div /> }));
vi.mock('./messages/flow-message', () => ({ default: () => <div /> }));

const { default: FlowAgents } = await import('./agents/flow-agents');
const { default: FlowScreenshots } = await import('./screenshots/flow-screenshots');
const { default: FlowTasks } = await import('./tasks/flow-tasks');
const { default: FlowTools } = await import('./tools/flow-tools');
const { default: FlowVectorStores } = await import('./vector-stores/flow-vector-stores');
const { default: FlowAssistantMessages } = await import('./messages/flow-assistant-messages');
const { default: FlowAutomationMessages } = await import('./messages/flow-automation-messages');

const item = { id: '1' };

const assistantFlowState = {
    assistantLogs: [{ assistantId: '5', id: '1' }],
    assistants: [{ id: '5', provider: { name: 'openai' }, status: StatusType.Running, title: 'Recon assistant' }],
    createAssistant: vi.fn(),
    deleteAssistant: vi.fn(),
    flowId: '42',
    flowStatus: StatusType.Running,
    initiateAssistantCreation: vi.fn(),
    selectAssistant: vi.fn(),
    selectedAssistantId: '5',
    stopAssistant: vi.fn(),
    submitAssistantMessage: vi.fn(),
};

const renderWithFlow = (state: Record<string, unknown>, ui: React.ReactElement) => {
    flowState.current = { flowId: '42', ...state };

    return render(<TooltipProvider>{ui}</TooltipProvider>);
};

describe('scroll-to-latest buttons carry an accessible name', () => {
    it.each([
        ['Scroll to latest agent log', { flowData: { agentLogs: [item] } }, <FlowAgents key="agents" />],
        ['Scroll to latest tool log', { flowData: { searchLogs: [item] } }, <FlowTools key="tools" />],
        [
            'Scroll to latest vector store log',
            { flowData: { vectorStoreLogs: [item] } },
            <FlowVectorStores key="vector-stores" />,
        ],
        ['Scroll to latest task', { flowData: { tasks: [item] } }, <FlowTasks key="tasks" />],
        ['Scroll to latest screenshot', { flowData: { screenshots: [item] } }, <FlowScreenshots key="screenshots" />],
        [
            'Scroll to latest message',
            {
                flowData: { messageLogs: [item] },
                flowStatus: StatusType.Running,
                stopAutomation: vi.fn(),
                submitAutomationMessage: vi.fn(),
            },
            <FlowAutomationMessages key="automation" />,
        ],
        ['Scroll to latest message', assistantFlowState, <FlowAssistantMessages key="assistant" />],
    ])('names the %s button', (name, state, ui) => {
        renderWithFlow(state, ui);

        expect(screen.getByRole('button', { name })).toBeInTheDocument();
    });
});

describe('assistant list', () => {
    it('names the per-assistant delete button after the assistant', async () => {
        const user = userEvent.setup({ delay: null });
        renderWithFlow(assistantFlowState, <FlowAssistantMessages />);

        await user.click(screen.getByRole('button', { name: 'Select assistant' }));

        expect(await screen.findByRole('button', { name: 'Delete Recon assistant' })).toBeInTheDocument();
    });
});
