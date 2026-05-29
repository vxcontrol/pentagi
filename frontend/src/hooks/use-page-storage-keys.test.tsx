import type { ReactNode } from 'react';

import { renderHook } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it } from 'vitest';

import { usePageStorageKeys } from './use-page-storage-keys';

const renderWithRouter = (initialEntries: string[], options?: Parameters<typeof usePageStorageKeys>[0]) => {
    const Wrapper = ({ children }: { children: ReactNode }) => (
        <MemoryRouter initialEntries={initialEntries}>{children}</MemoryRouter>
    );

    return renderHook(() => usePageStorageKeys(options), { wrapper: Wrapper });
};

describe('usePageStorageKeys — defaults', () => {
    it('builds keys from the live pathname when no override is provided', () => {
        const { result } = renderWithRouter(['/flows']);

        expect(result.current.table).toBe('table_4_/flows');
        expect(result.current.period).toBe('period_4_/flows');
        expect(result.current.viewOptions).toBe('viewOptions_4_/flows');
    });

    it('reflects the entire pathname — does NOT strip params by default', () => {
        // A detail page that mounts the hook directly (no override) gets a
        // unique key per id. This is the correct behaviour for tables on
        // distinct subroutes; detail pages that want to share with the list
        // must pass `useTopLevel: true` or an explicit override.
        const { result } = renderWithRouter(['/flows/abc-123']);

        expect(result.current.table).toBe('table_4_/flows/abc-123');
    });
});

describe('usePageStorageKeys — useTopLevel', () => {
    it('strips id suffixes when `useTopLevel: true`', () => {
        const { result } = renderWithRouter(['/flows/abc-123'], { useTopLevel: true });

        expect(result.current.table).toBe('table_4_/flows');
    });

    it('handles a deep path with `useTopLevel: true` — only the first segment survives', () => {
        const { result } = renderWithRouter(['/knowledges/abc/foo/bar'], { useTopLevel: true });

        expect(result.current.table).toBe('table_4_/knowledges');
    });

    it('falls back to an empty path for the root', () => {
        const { result } = renderWithRouter(['/'], { useTopLevel: true });

        // `getTopLevelPath('/')` returns `''`, so the key is `table_4_`.
        // Documented as the deliberate behaviour — root-route detail pages
        // would all collide on this key, but no such page exists.
        expect(result.current.table).toBe('table_4_');
    });
});

describe('usePageStorageKeys — explicit pathname override', () => {
    it('uses the override verbatim, ignoring the live pathname', () => {
        const { result } = renderWithRouter(['/flows/abc-123'], { pathname: '/admin/flows' });

        expect(result.current.table).toBe('table_4_/admin/flows');
    });

    it('override + useTopLevel uses the top-level segment of the override', () => {
        const { result } = renderWithRouter(['/flows/abc'], {
            pathname: '/admin/flows/xyz',
            useTopLevel: true,
        });

        expect(result.current.table).toBe('table_4_/admin');
    });
});

describe('usePageStorageKeys — reactivity', () => {
    it('returns the same memoized object across re-renders when inputs do not change', () => {
        const { rerender, result } = renderWithRouter(['/flows']);

        const first = result.current;

        rerender();

        // Memoization guarantee — downstream effects depend on this for
        // their identity-stable dep arrays.
        expect(result.current).toBe(first);
    });
});
