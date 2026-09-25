import type { ReactElement } from 'react';

import { render as renderUi, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';

import type { Provider } from '@/models/provider';

import { TooltipProvider } from '@/components/ui/tooltip';
import { ProviderType } from '@/graphql/types';

import { FlowProviderLabel } from './flow-provider-label';

const openai: Provider = { name: 'openai', type: ProviderType.Openai };
const custom: Provider = { name: 'my-gateway', type: ProviderType.Custom };

const render = (ui: ReactElement) => renderUi(<TooltipProvider delayDuration={0}>{ui}</TooltipProvider>);
const unavailableMark = () => screen.queryByRole('img', { name: 'Unavailable' });

describe('FlowProviderLabel', () => {
    it('leaves a flow alone while its provider is still configured', () => {
        render(
            <FlowProviderLabel
                isOwnFlow
                provider={openai}
                providers={[openai, custom]}
            />,
        );

        expect(screen.getByText('openai')).toBeInTheDocument();
        expect(unavailableMark()).not.toBeInTheDocument();
        expect(screen.getByTestId('flow-provider-label').className).not.toContain('opacity-50');
    });

    it('marks a flow whose provider is gone from the configured list', () => {
        render(
            <FlowProviderLabel
                isOwnFlow
                provider={custom}
                providers={[openai]}
            />,
        );

        expect(screen.getByText('my-gateway')).toBeInTheDocument();
        expect(unavailableMark()).toBeInTheDocument();
        expect(screen.getByTestId('flow-provider-label').className).toContain('opacity-50');
    });

    it('says why the flow is marked when the mark is hovered', async () => {
        const user = userEvent.setup();

        render(
            <FlowProviderLabel
                isOwnFlow
                provider={custom}
                providers={[openai]}
            />,
        );

        await user.hover(screen.getByRole('img', { name: 'Unavailable' }));

        expect(await screen.findByRole('tooltip')).toHaveTextContent('Unavailable');
    });

    it('says the same to a keyboard user who tabs onto the mark', async () => {
        const user = userEvent.setup();

        render(
            <FlowProviderLabel
                isOwnFlow
                provider={custom}
                providers={[openai]}
            />,
        );

        await user.tab();

        expect(unavailableMark()).toHaveFocus();
        expect(await screen.findByRole('tooltip')).toHaveTextContent('Unavailable');
    });

    it('says nothing while the provider list has not arrived', () => {
        render(
            <FlowProviderLabel
                isOwnFlow
                provider={custom}
                providers={[]}
            />,
        );

        expect(unavailableMark()).not.toBeInTheDocument();
    });

    it('marks a provider whose name survived but whose type changed', () => {
        render(
            <FlowProviderLabel
                isOwnFlow
                provider={custom}
                providers={[{ name: 'my-gateway', type: ProviderType.Openai }]}
            />,
        );

        expect(unavailableMark()).toBeInTheDocument();
    });

    it("says nothing about a flow that is not the viewer's", () => {
        render(
            <FlowProviderLabel
                isOwnFlow={false}
                provider={custom}
                providers={[openai]}
            />,
        );

        expect(screen.getByText('my-gateway')).toBeInTheDocument();
        expect(
            unavailableMark(),
            'the provider is missing from the viewer list, not from its owner list',
        ).not.toBeInTheDocument();
        expect(screen.getByTestId('flow-provider-label').className).not.toContain('opacity-50');
    });
});
