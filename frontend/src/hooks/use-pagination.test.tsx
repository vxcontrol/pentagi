import type { ReactNode } from 'react';

import { act, renderHook, waitFor } from '@testing-library/react';
import { MemoryRouter, useLocation } from 'react-router-dom';
import { describe, expect, it } from 'vitest';

import { usePagination } from './use-pagination';

interface RenderResult {
    pageIndex: number;
    search: string;
    setPage: (pageIndex: number) => void;
}

// Bundle `usePagination` + `useLocation` into a single hook so `renderHook`
// observes both reactively — without this, a sibling `<LocationProbe>` lags
// one render behind the canonicalization layout effect when the URL is
// rewritten on mount.
const usePaginationWithLocation = (): RenderResult => {
    const { pageIndex, setPage } = usePagination();
    const { search } = useLocation();

    return { pageIndex, search, setPage };
};

const renderWithRouter = (initialEntries: string[]) => {
    const Wrapper = ({ children }: { children: ReactNode }) => (
        <MemoryRouter initialEntries={initialEntries}>{children}</MemoryRouter>
    );

    return renderHook(usePaginationWithLocation, { wrapper: Wrapper });
};

describe('usePagination', () => {
    it('returns pageIndex 0 when no `?page=` is present', () => {
        const { result } = renderWithRouter(['/flows']);
        expect(result.current.pageIndex).toBe(0);
    });

    it('converts a 1-based `?page=` URL param to a 0-based index', () => {
        const { result } = renderWithRouter(['/flows?page=3']);
        expect(result.current.pageIndex).toBe(2);
    });

    it('falls back to 0 on a non-numeric `?page=`', () => {
        const { result } = renderWithRouter(['/flows?page=notanumber']);
        expect(result.current.pageIndex).toBe(0);
    });

    it('canonicalizes `?page=1` away from the URL on mount', async () => {
        const { result } = renderWithRouter(['/flows?page=1']);

        // `setSearchParams` from a layout effect schedules a transition that
        // doesn't always commit in the same `act()` flush as the initial
        // mount under React 19 + react-router 7, so wait for the URL to
        // settle.
        await waitFor(() => {
            expect(result.current.search).toBe('');
        });
    });

    it('canonicalizes `?page=1` but leaves other params untouched', async () => {
        const { result } = renderWithRouter(['/flows?page=1&q=foo']);

        await waitFor(() => {
            const params = new URLSearchParams(result.current.search);
            expect(params.has('page')).toBe(false);
        });

        expect(new URLSearchParams(result.current.search).get('q')).toBe('foo');
    });

    it('does not canonicalize `?page=2` (canonical form already)', () => {
        const { result } = renderWithRouter(['/flows?page=2']);
        expect(result.current.search).toBe('?page=2');
    });

    it('setPage(0) removes the `?page=` param entirely', () => {
        const { result } = renderWithRouter(['/flows?page=5']);
        act(() => result.current.setPage(0));
        expect(result.current.search).toBe('');
    });

    it('setPage(N>0) writes 1-based number to `?page=`', () => {
        const { result } = renderWithRouter(['/flows']);
        act(() => result.current.setPage(4));
        expect(result.current.search).toBe('?page=5');
    });

    it('setPage preserves other URL params', () => {
        const { result } = renderWithRouter(['/flows?q=foo']);
        act(() => result.current.setPage(2));
        const params = new URLSearchParams(result.current.search ?? '');
        expect(params.get('q')).toBe('foo');
        expect(params.get('page')).toBe('3');
    });

    it('honors a custom paramName override', () => {
        const Wrapper = ({ children }: { children: ReactNode }) => (
            <MemoryRouter initialEntries={['/flows?p=4']}>{children}</MemoryRouter>
        );
        const { result } = renderHook(() => usePagination({ paramName: 'p' }), { wrapper: Wrapper });
        expect(result.current.pageIndex).toBe(3);
    });
});
