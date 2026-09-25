import { render } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { TooltipProvider } from '@/components/ui/tooltip';
import { StatusType } from '@/graphql/types';

const { flowState, formProps, providersState } = vi.hoisted(() => ({
    flowState: { current: {} as Record<string, unknown> },
    formProps: [] as { defaultValues?: { providerName?: string } }[],
    providersState: { current: {} as Record<string, unknown> },
}));

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
vi.mock('@/providers/providers-provider', () => ({ useProviders: () => providersState.current }));
vi.mock('@/providers/system-settings-provider', () => ({ useSystemSettings: () => ({ settings: null }) }));

vi.mock('../flow-form', () => ({
    FlowForm: (props: { defaultValues?: { providerName?: string } }) => {
        formProps.push(props);

        return <div />;
    },
}));
vi.mock('./flow-message', () => ({ default: () => <div /> }));

const { default: FlowAssistantMessages } = await import('./flow-assistant-messages');

const renderComposer = (state: Record<string, unknown>) => {
    formProps.length = 0;
    providersState.current = {
        providers: [{ name: 'anthropic', type: 'anthropic' }],
        selectedProvider: { name: 'anthropic', type: 'anthropic' },
    };
    flowState.current = {
        assistantLogs: [],
        assistants: [],
        createAssistant: vi.fn(),
        deleteAssistant: vi.fn(),
        flowId: '42',
        flowStatus: StatusType.Running,
        initiateAssistantCreation: vi.fn(),
        selectAssistant: vi.fn(),
        selectedAssistantId: null,
        stopAssistant: vi.fn(),
        submitAssistantMessage: vi.fn(),
        ...state,
    };

    render(
        <TooltipProvider>
            <FlowAssistantMessages />
        </TooltipProvider>,
    );

    return formProps.at(-1)?.defaultValues?.providerName;
};

describe('the assistant composer defaults its provider the way /flows/new does', () => {
    it('falls back to the remembered provider when no assistant is selected yet', () => {
        expect(renderComposer({})).toBe('anthropic');
    });

    it('keeps the selected assistant own provider rather than the remembered one', () => {
        const providerName = renderComposer({
            assistants: [
                { id: '5', provider: { name: 'openai' }, status: StatusType.Running, title: 'Recon assistant' },
            ],
            selectedAssistantId: '5',
        });

        expect(providerName).toBe('openai');
    });
});
