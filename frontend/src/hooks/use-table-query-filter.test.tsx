import type { ReactNode } from 'react';

import { act, renderHook, waitFor } from '@testing-library/react';
import { MemoryRouter, useLocation } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';

import { useTableQueryFilter, useTableQueryFilterReader } from './use-table-query-filter';

const STORAGE_KEY = 'table_4_/flows';
const SHORT_DEBOUNCE_MS = 5;

interface RenderResult {
    debouncedFilter: string;
    filter: string;
    resetFilter: () => void;
    search: string;
    setFilter: (value: string) => void;
}

const useFilterWithLocation = (options: { debounceMs?: number; storageKey?: string } = {}): RenderResult => {
    const { debouncedFilter, filter, resetFilter, setFilter } = useTableQueryFilter({
        debounceMs: options.debounceMs ?? SHORT_DEBOUNCE_MS,
        storageKey: options.storageKey ?? STORAGE_KEY,
    });
    const { search } = useLocation();

    return { debouncedFilter, filter, resetFilter, search, setFilter };
};

const renderWithRouter = (initialEntries: string[], options?: { debounceMs?: number; storageKey?: string }) => {
    const Wrapper = ({ children }: { children: ReactNode }) => (
        <MemoryRouter initialEntries={initialEntries}>{children}</MemoryRouter>
    );

    return renderHook(() => useFilterWithLocation(options), { wrapper: Wrapper });
};

beforeEach(() => {
    localStorage.clear();
});

afterEach(() => {
    localStorage.clear();
});

const readStoredFilter = (key: string): null | string => {
    const raw = localStorage.getItem(key);

    if (raw === null) {
        return null;
    }

    try {
        const parsed = JSON.parse(raw);

        return typeof parsed?.filter === 'string' ? parsed.filter : null;
    } catch {
        return null;
    }
};

describe('useTableQueryFilter — URL ↔ storage roundtrip', () => {
    it('reads the initial filter from `?q=` when present', () => {
        const { result } = renderWithRouter(['/flows?q=alpha']);
        expect(result.current.filter).toBe('alpha');
    });

    it('returns an empty filter when neither URL nor storage have a value', () => {
        const { result } = renderWithRouter(['/flows']);
        expect(result.current.filter).toBe('');
    });

    // The restore-from-storage path lives in a `useEffect` that issues
    // `setSearchParams` on mount. Under React 19 + react-router 7 the
    // resulting URL transition lands one micro-task after the initial
    // `act(render)` flush, so we drain one round of pending updates
    // through an empty async `act` before asserting.
    it('restores `?q=` from storage when the URL is empty', async () => {
        localStorage.setItem(STORAGE_KEY, JSON.stringify({ filter: 'alpha' }));
        const { result } = renderWithRouter(['/flows']);

        await act(async () => {
            await Promise.resolve();
        });

        expect(new URLSearchParams(result.current.search).get('q')).toBe('alpha');
        expect(result.current.filter).toBe('alpha');
    });

    it('lets the URL win over storage when both have a value', async () => {
        localStorage.setItem(STORAGE_KEY, JSON.stringify({ filter: 'beta' }));
        const { result } = renderWithRouter(['/flows?q=alpha']);

        expect(result.current.filter).toBe('alpha');

        // The hook should sync the URL value into storage so a subsequent
        // tab without `?q=` resumes from the URL's intent, not the old one.
        await waitFor(() => {
            expect(readStoredFilter(STORAGE_KEY)).toBe('alpha');
        });
    });

    it('setFilter writes the value into the URL', () => {
        const { result } = renderWithRouter(['/flows']);

        act(() => result.current.setFilter('gamma'));

        expect(new URLSearchParams(result.current.search).get('q')).toBe('gamma');
    });

    it('setFilter("") drops the URL param entirely (no `?q=` empty entry)', () => {
        const { result } = renderWithRouter(['/flows?q=alpha']);

        act(() => result.current.setFilter(''));

        expect(result.current.search).toBe('');
    });

    it('setFilter("") clears the storage entry through the effect-driven write', async () => {
        localStorage.setItem(STORAGE_KEY, JSON.stringify({ filter: 'alpha' }));
        const { result } = renderWithRouter(['/flows?q=alpha']);

        act(() => result.current.setFilter(''));

        await waitFor(() => {
            expect(readStoredFilter(STORAGE_KEY)).toBeNull();
        });
    });

    it('persists the URL filter to storage once typing settles', async () => {
        const { result } = renderWithRouter(['/flows']);

        act(() => result.current.setFilter('persisted'));

        await waitFor(() => {
            expect(readStoredFilter(STORAGE_KEY)).toBe('persisted');
        });
    });

    it('resetFilter is equivalent to setFilter("")', () => {
        const { result } = renderWithRouter(['/flows?q=alpha']);

        act(() => result.current.resetFilter());

        expect(result.current.search).toBe('');
        expect(result.current.filter).toBe('');
    });

    it('debouncedFilter eventually catches up with filter', async () => {
        const { result } = renderWithRouter(['/flows']);

        act(() => result.current.setFilter('typed'));
        expect(result.current.filter).toBe('typed');

        await waitFor(() => {
            expect(result.current.debouncedFilter).toBe('typed');
        });
    });

    it('does not clobber unrelated table state when writing the filter', async () => {
        localStorage.setItem(
            STORAGE_KEY,
            JSON.stringify({ columnVisibility: { name: false }, sorting: [{ desc: true, id: 'createdAt' }] }),
        );
        const { result } = renderWithRouter(['/flows']);

        act(() => result.current.setFilter('foo'));

        await waitFor(() => {
            const stored = JSON.parse(localStorage.getItem(STORAGE_KEY) ?? '{}');
            expect(stored.filter).toBe('foo');
            expect(stored.sorting).toEqual([{ desc: true, id: 'createdAt' }]);
            expect(stored.columnVisibility).toEqual({ name: false });
        });
    });

    it('re-restores when storageKey rotates while the hook stays mounted', async () => {
        const FIRST_KEY = 'table_4_/flows';
        const SECOND_KEY = 'table_4_/templates';
        localStorage.setItem(SECOND_KEY, JSON.stringify({ filter: 'second-key' }));

        const Wrapper = ({ children }: { children: ReactNode }) => (
            <MemoryRouter initialEntries={['/flows']}>{children}</MemoryRouter>
        );
        const { rerender, result } = renderHook(
            ({ key }) => {
                const filter = useTableQueryFilter({ debounceMs: SHORT_DEBOUNCE_MS, storageKey: key });
                const { search } = useLocation();

                return { ...filter, search };
            },
            { initialProps: { key: FIRST_KEY }, wrapper: Wrapper },
        );

        await act(async () => {
            await Promise.resolve();
        });
        expect(result.current.filter).toBe('');

        rerender({ key: SECOND_KEY });

        await waitFor(() => {
            expect(new URLSearchParams(result.current.search).get('q')).toBe('second-key');
        });
    });
});

describe('useTableQueryFilter — page param interaction', () => {
    it('resets `?page=` when setFilter narrows the result set', () => {
        const { result } = renderWithRouter(['/flows?page=3']);

        act(() => result.current.setFilter('foo'));

        const params = new URLSearchParams(result.current.search);
        expect(params.has('page')).toBe(false);
        expect(params.get('q')).toBe('foo');
    });

    it('keeps `?page=` when clearPageParamOnChange is false', () => {
        const Wrapper = ({ children }: { children: ReactNode }) => (
            <MemoryRouter initialEntries={['/flows?page=3']}>{children}</MemoryRouter>
        );
        const { result } = renderHook(
            () => {
                const filter = useTableQueryFilter({
                    clearPageParamOnChange: false,
                    debounceMs: SHORT_DEBOUNCE_MS,
                    storageKey: STORAGE_KEY,
                });
                const { search } = useLocation();

                return { ...filter, search };
            },
            { wrapper: Wrapper },
        );

        act(() => result.current.setFilter('foo'));

        const params = new URLSearchParams(result.current.search);
        expect(params.get('page')).toBe('3');
    });
});

describe('useTableQueryFilterReader', () => {
    it('observes the URL filter without writing storage', () => {
        const Wrapper = ({ children }: { children: ReactNode }) => (
            <MemoryRouter initialEntries={['/flows/abc?q=alpha']}>{children}</MemoryRouter>
        );
        const { result } = renderHook(
            () => useTableQueryFilterReader({ debounceMs: SHORT_DEBOUNCE_MS, storageKey: STORAGE_KEY }),
            { wrapper: Wrapper },
        );

        expect(result.current.filter).toBe('alpha');
        expect(localStorage.getItem(STORAGE_KEY)).toBeNull();
    });

    it('does not replay a stored filter into a clean URL', () => {
        localStorage.setItem(STORAGE_KEY, JSON.stringify({ filter: 'alpha' }));
        const Wrapper = ({ children }: { children: ReactNode }) => (
            <MemoryRouter initialEntries={['/flows/abc']}>{children}</MemoryRouter>
        );
        const { result } = renderHook(
            () => {
                const reader = useTableQueryFilterReader({
                    debounceMs: SHORT_DEBOUNCE_MS,
                    storageKey: STORAGE_KEY,
                });
                const { search } = useLocation();

                return { ...reader, search };
            },
            { wrapper: Wrapper },
        );

        // A shared `/flows/abc` link explicitly clears the filter on entry —
        // detail-page subscribers must not inject the previous tab's `?q=`
        // back into the URL.
        expect(result.current.search).toBe('');
        expect(result.current.filter).toBe('');
    });
});
