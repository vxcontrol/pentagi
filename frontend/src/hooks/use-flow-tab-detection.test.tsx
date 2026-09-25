import type { ReactNode } from 'react';

import { act, renderHook, waitFor } from '@testing-library/react';
import { BrowserRouter, useLocation, useNavigate } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';

const currentFlow = { id: '1', messageLogs: [{ id: '1' }] as { id: string }[] };

vi.mock('@/providers/flow-provider', () => ({
    useFlow: () => ({
        flowData: { flow: { id: currentFlow.id }, messageLogs: currentFlow.messageLogs },
        isLoading: false,
    }),
}));

import flowTabsSource from '@/features/flows/flow-tabs.tsx?raw';

import { FLOW_TAB_VALUES, useFlowSideTab, useFlowTabDetection } from './use-flow-tab-detection';

const renderDetection = (entry: string, tabValues?: readonly string[]) => {
    window.history.replaceState(null, '', entry);

    return renderHook(
        () => ({ detection: useFlowTabDetection(tabValues), navigate: useNavigate(), search: useLocation().search }),
        {
            wrapper: ({ children }: { children: ReactNode }) => <BrowserRouter>{children}</BrowserRouter>,
        },
    );
};

const renderFlowPage = (entry: string) => {
    window.history.replaceState(null, '', entry);

    return renderHook(
        () => ({
            central: useFlowTabDetection(),
            mobile: useFlowTabDetection(FLOW_TAB_VALUES),
            navigate: useNavigate(),
            search: useLocation().search,
            side: useFlowSideTab(),
        }),
        {
            wrapper: ({ children }: { children: ReactNode }) => <BrowserRouter>{children}</BrowserRouter>,
        },
    );
};

describe('useFlowTabDetection', () => {
    it('resolves the tab the URL names', () => {
        expect(renderDetection('/flows/1?tab=dashboard').result.current.detection.resolvedTab).toBe('dashboard');
    });

    it('leaves the inner-tab parameters alone when the central tab changes', () => {
        const { result } = renderDetection('/flows/1?tab=dashboard&view=analytics&graph=full-attack-chain');

        act(() => result.current.detection.handleTabChange('terminal'));

        expect(result.current.search).toBe('?tab=terminal&view=analytics&graph=full-attack-chain');
    });

    it('follows the browser Back button to the tab the address bar shows', async () => {
        const { result } = renderDetection('/flows/1?tab=automation');

        act(() => result.current.detection.handleTabChange('dashboard'));

        expect(result.current.detection.resolvedTab).toBe('dashboard');
        expect(result.current.search).toBe('?tab=dashboard');

        await act(async () => {
            window.history.back();

            await new Promise((settle) => {
                setTimeout(settle, 0);
            });
        });

        await waitFor(() => expect(result.current.search).toBe('?tab=automation'));
        expect(result.current.detection.resolvedTab).toBe('automation');
    });

    it('keeps a parameter that reached the address bar before the router saw it', () => {
        const { result } = renderDetection('/flows/1?tab=dashboard');

        window.history.replaceState(null, '', '/flows/1?tab=dashboard&graph=full-attack-chain');

        act(() => result.current.detection.handleTabChange('terminal'));

        expect(result.current.search).toContain('graph=full-attack-chain');
    });

    it('restores a tab outside the central three for the consumer that renders it', () => {
        const { result } = renderDetection('/flows/1?tab=terminal', FLOW_TAB_VALUES);

        expect(result.current.detection.resolvedTab).toBe('terminal');
    });

    it('ignores a tab the consumer cannot render', () => {
        expect(renderDetection('/flows/1?tab=terminal').result.current.detection.resolvedTab).toBe('automation');
    });

    it('knows every tab the flow tabs put on screen', () => {
        const onScreen = [...flowTabsSource.matchAll(/<TabsTrigger value="([^"]+)"/g)].map(([, value]) => value);

        expect([...onScreen].sort()).toEqual([...FLOW_TAB_VALUES].sort());
    });

    it('does not carry a tab picked on one flow over to the next', () => {
        currentFlow.id = '1';

        const { rerender, result } = renderDetection('/flows/1');

        act(() => result.current.detection.handleTabChange('dashboard'));
        expect(result.current.detection.resolvedTab).toBe('dashboard');

        // the sidebar link carries no ?tab=, and a change of :flowId does not remount the route element
        act(() => result.current.navigate('/flows/2'));
        currentFlow.id = '2';
        rerender();

        expect(result.current.detection.resolvedTab, 'the tab chosen on the previous flow decided this one').not.toBe(
            'dashboard',
        );

        currentFlow.id = '1';
    });

    it('does not move the tab when the first automation log arrives', () => {
        currentFlow.id = '9';
        currentFlow.messageLogs = [];

        const { rerender, result } = renderDetection('/flows/9');

        expect(result.current.detection.resolvedTab).toBe('assistant');

        // the log arrives over a subscription while the reader sits on the tab
        currentFlow.messageLogs = [{ id: 'first-automation-log' }];
        rerender();

        expect(result.current.detection.resolvedTab, 'the tab moved under the reader').toBe('assistant');

        currentFlow.id = '1';
        currentFlow.messageLogs = [{ id: '1' }];
    });
});

describe('useFlowSideTab', () => {
    it('opens the side tab a link names, whatever strip the layout put it in', () => {
        expect(renderFlowPage('/flows/1?tab=screenshots').result.current.side.resolvedTab).toBe('screenshots');
    });

    it('holds its own tab while the central one moves', () => {
        const { result } = renderFlowPage('/flows/1?tab=dashboard&side=files');

        expect(result.current.side.resolvedTab).toBe('files');
        expect(result.current.central.resolvedTab).toBe('dashboard');
    });

    it('writes the pick into the address so a reload comes back to it', () => {
        const { result } = renderFlowPage('/flows/1?tab=dashboard');

        act(() => result.current.side.handleTabChange('tools'));

        expect(result.current.search).toContain('side=tools');
        expect(result.current.side.resolvedTab).toBe('tools');
        expect(result.current.central.resolvedTab, 'the central panel kept its own tab').toBe('dashboard');
    });

    it('opens the terminal when the address names no side tab', () => {
        expect(renderFlowPage('/flows/1').result.current.side.resolvedTab).toBe('terminal');
    });

    it('refuses a central value planted in its own parameter', () => {
        expect(renderFlowPage('/flows/1?side=dashboard').result.current.side.resolvedTab).toBe('terminal');
    });

    it('keeps a linked side tab when the central strip rewrites its own parameter', async () => {
        const { result } = renderFlowPage('/flows/1?tab=screenshots');

        await waitFor(() => expect(result.current.search).toContain('side=screenshots'));

        act(() => result.current.central.handleTabChange('assistant'));

        expect(result.current.side.resolvedTab, 'the link named screenshots and nothing unsaid it').toBe('screenshots');
    });

    it('replaces its own address entry instead of stacking one per pick', () => {
        const { result } = renderFlowPage('/flows/1');
        const before = window.history.length;

        for (const tab of ['tasks', 'agents', 'tools', 'files']) {
            act(() => result.current.side.handleTabChange(tab));
        }

        expect(result.current.side.resolvedTab).toBe('files');
        expect(window.history.length - before, 'Back must leave the flow, not walk the tabs').toBe(0);
    });
});

describe('crossing the two-panel boundary', () => {
    it('carries a side tab picked on the single strip into the two-panel layout', () => {
        const { result } = renderFlowPage('/flows/1');

        act(() => result.current.mobile.handleTabChange('screenshots'));

        expect(result.current.side.resolvedTab).toBe('screenshots');
    });

    it('keeps the strip and the address agreeing after crossing twice', () => {
        const { result } = renderFlowPage('/flows/1');

        act(() => result.current.mobile.handleTabChange('files'));
        act(() => result.current.central.handleTabChange('dashboard'));

        expect(result.current.side.resolvedTab, 'the wide layout kept the side pick').toBe('files');
        expect(result.current.mobile.resolvedTab, 'the narrow strip reads what the address says').toBe('dashboard');
        expect(result.current.search).toContain('tab=dashboard');
    });

    it('follows an address change that no Back button caused', () => {
        const { result } = renderFlowPage('/flows/1?tab=automation');

        act(() => result.current.central.handleTabChange('dashboard'));
        expect(result.current.central.resolvedTab).toBe('dashboard');

        act(() => result.current.navigate('/flows/1?tab=assistant'));

        expect(result.current.central.resolvedTab, 'the pick outlived the address that replaced it').toBe('assistant');
    });

    it('reads the tab of the flow the address names, not the one being left', async () => {
        currentFlow.id = '1';

        const { rerender, result } = renderFlowPage('/flows/1?tab=dashboard');

        act(() => result.current.navigate('/flows/2?tab=assistant'));
        currentFlow.id = '2';
        rerender();

        expect(result.current.central.resolvedTab).toBe('assistant');

        await act(async () => {
            window.history.back();

            await new Promise((settle) => {
                setTimeout(settle, 0);
            });
        });

        currentFlow.id = '1';
        rerender();

        await waitFor(() => expect(result.current.central.resolvedTab).toBe('dashboard'));
    });
});
