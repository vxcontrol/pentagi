import type { ReactNode } from 'react';

import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { BrowserRouter, useLocation } from 'react-router-dom';
import { afterEach, describe, expect, it, vi } from 'vitest';

vi.mock('@/providers/flow-provider', () => ({
    useFlow: () => ({ flowData: { messageLogs: [{ id: '1' }] }, isLoading: false }),
}));

vi.mock('@/features/flows/dashboard/flow-dashboard', () => ({ default: () => <div data-slot="panel-dashboard" /> }));
vi.mock('@/features/flows/messages/flow-assistant-messages', () => ({
    default: () => <div data-slot="panel-assistant" />,
}));
vi.mock('@/features/flows/messages/flow-automation-messages', () => ({
    default: () => <div data-slot="panel-automation" />,
}));

import FlowCentralTabs from './flow-central-tabs';

function SearchProbe() {
    return <span data-slot="probe-search">{useLocation().search}</span>;
}

const renderTabs = (entry: string) => {
    window.history.replaceState(null, '', entry);

    const wrapper = ({ children }: { children: ReactNode }) => <BrowserRouter>{children}</BrowserRouter>;

    return render(
        <>
            <SearchProbe />
            <FlowCentralTabs />
        </>,
        { wrapper },
    );
};

const search = () => screen.getByTestId('probe-search').textContent ?? '';

afterEach(() => {
    window.history.replaceState(null, '', '/');
});

describe('FlowCentralTabs', () => {
    it('moves focus along the strip without opening every panel it passes', async () => {
        renderTabs('/flows/1?tab=automation');

        const entriesBefore = window.history.length;

        await userEvent.tab();
        expect(screen.getByRole('tab', { name: 'Automation' })).toHaveFocus();

        await userEvent.keyboard('{ArrowRight}{ArrowRight}');

        expect(screen.getByRole('tab', { name: 'Dashboard' })).toHaveFocus();
        expect(search()).toContain('tab=automation');
        expect(window.history.length).toBe(entriesBefore);
        expect(screen.queryByTestId('panel-dashboard')).not.toBeInTheDocument();
    });

    it('opens the panel the reader asks for', async () => {
        renderTabs('/flows/1?tab=automation');

        await userEvent.tab();
        await userEvent.keyboard('{ArrowRight}{ArrowRight}{Enter}');

        expect(search()).toContain('tab=dashboard');
        expect(screen.getByTestId('panel-dashboard')).toBeInTheDocument();
    });
});
