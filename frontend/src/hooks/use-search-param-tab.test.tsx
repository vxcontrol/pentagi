import type { ReactNode } from 'react';

import { act, renderHook, waitFor } from '@testing-library/react';
import { BrowserRouter, useLocation } from 'react-router-dom';
import { describe, expect, it } from 'vitest';

import { useSearchParamTab } from './use-search-param-tab';

const VALUES = ['overview', 'analytics'] as const;

const wrapperAt = (entry: string) => {
    window.history.replaceState(null, '', entry);

    return function Wrapper({ children }: { children: ReactNode }) {
        return <BrowserRouter>{children}</BrowserRouter>;
    };
};

const renderTab = (entry: string, values: readonly string[] = VALUES) =>
    renderHook(
        () => ({
            search: useLocation().search,
            tab: useSearchParamTab({ fallback: 'overview', key: 'view', values }),
        }),
        { wrapper: wrapperAt(entry) },
    );

const renderVolatileTab = (entry: string) =>
    renderHook(
        ({ values }: { values: readonly string[] }) =>
            useSearchParamTab({
                fallback: values[0] ?? '',
                hasVolatileValues: true,
                key: 'graph',
                values,
            }),
        { initialProps: { values: ['infrastructure'] as readonly string[] }, wrapper: wrapperAt(entry) },
    );

describe('useSearchParamTab', () => {
    it('reads the tab the URL names', () => {
        expect(renderTab('/flows/1?tab=dashboard&view=analytics').result.current.tab[0]).toBe('analytics');
    });

    it('falls back when the URL names a tab that is not on offer', () => {
        expect(renderTab('/flows/1?view=analytics', ['overview']).result.current.tab[0]).toBe('overview');
    });

    it('keeps the rest of the query string when the tab changes', () => {
        const { result } = renderTab('/flows/1?tab=dashboard&graph=full-attack-chain');

        act(() => result.current.tab[1]('analytics'));

        expect(result.current.search).toBe('?tab=dashboard&graph=full-attack-chain&view=analytics');
        expect(result.current.tab[0]).toBe('analytics');
    });

    it('holds a volatile tab where the reader found it when a later one gains data', async () => {
        const { rerender, result } = renderVolatileTab('/flows/1?tab=dashboard');

        await waitFor(() => expect(result.current[0]).toBe('infrastructure'));

        rerender({ values: ['overview', 'infrastructure'] });

        expect(result.current[0]).toBe('infrastructure');
    });

    // The graph tabs appear as their data arrives, so a value the URL names may be legitimate and
    // simply not loaded yet.
    it('keeps a volatile tab the URL names until its data arrives', async () => {
        const { rerender, result } = renderHook(
            ({ values }: { values: readonly string[] }) =>
                useSearchParamTab({ fallback: values[0] ?? '', hasVolatileValues: true, key: 'graph', values }),
            {
                initialProps: { values: ['overview'] as readonly string[] },
                wrapper: wrapperAt('/flows/1?graph=shortest-path'),
            },
        );

        await waitFor(() => expect(result.current[0]).toBe('overview'));

        rerender({ values: ['overview', 'shortest-path'] });

        expect(result.current[0]).toBe('shortest-path');
    });

    // The reader asked for a tab this flow has no data for yet, so they are looking at whatever
    // was available. A graph that arrives later must not take the panel from under them.
    it('does not move the reader between tabs they did not ask for', async () => {
        const { rerender, result } = renderHook(
            ({ values }: { values: readonly string[] }) =>
                useSearchParamTab({ fallback: values[0] ?? '', hasVolatileValues: true, key: 'graph', values }),
            {
                initialProps: { values: ['infrastructure'] as readonly string[] },
                wrapper: wrapperAt('/flows/2?graph=shortest-path'),
            },
        );

        await waitFor(() => expect(result.current[0]).toBe('infrastructure'));

        rerender({ values: ['overview', 'infrastructure'] });

        expect(result.current[0]).toBe('infrastructure');

        rerender({ values: ['overview', 'infrastructure', 'shortest-path'] });

        // The tab they did ask for arrives: that switch is the one they wanted.
        expect(result.current[0]).toBe('shortest-path');
    });

    it('pins a choice that happens to equal the fallback', () => {
        const { result } = renderTab('/flows/1?tab=dashboard&view=analytics');

        act(() => result.current.tab[1]('overview'));

        expect(result.current.search).toBe('?tab=dashboard&view=overview');
        expect(result.current.tab[0]).toBe('overview');
    });
});
