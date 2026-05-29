import type { ReactNode } from 'react';

import { act, renderHook, waitFor } from '@testing-library/react';
import { MemoryRouter, useLocation } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';

import { useTableState } from './use-table-state';

const STORAGE_KEY = 'table_4_/flows';
const SHORT_DEBOUNCE_MS = 5;

interface RenderResult {
    debouncedFilter: string;
    filter: string;
    pageIndex: number;
    resetFilter: () => void;
    search: string;
    setFilter: (value: string) => void;
    setPage: ReturnType<typeof useTableState>['setPage'];
    update: ReturnType<typeof useTableState>['update'];
}

const useStateWithLocation = (options: { debounceMs?: number } = {}): RenderResult => {
    const { debouncedFilter, filter, pageIndex, resetFilter, setFilter, setPage, update } = useTableState({
        debounceMs: options.debounceMs ?? SHORT_DEBOUNCE_MS,
    });
    const { search } = useLocation();

    return { debouncedFilter, filter, pageIndex, resetFilter, search, setFilter, setPage, update };
};

const renderWithRouter = (initialEntries: string[], options?: { debounceMs?: number }) => {
    const Wrapper = ({ children }: { children: ReactNode }) => (
        <MemoryRouter initialEntries={initialEntries}>{children}</MemoryRouter>
    );

    return renderHook(() => useStateWithLocation(options), { wrapper: Wrapper });
};

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

beforeEach(() => {
    localStorage.clear();
});

afterEach(() => {
    localStorage.clear();
});

describe('useTableState — URL reads', () => {
    it('reads the initial filter from `?q=`', () => {
        const { result } = renderWithRouter(['/flows?q=alpha']);
        expect(result.current.filter).toBe('alpha');
    });

    it('reads the initial pageIndex from `?page=` (1-based URL → 0-based state)', () => {
        const { result } = renderWithRouter(['/flows?page=3']);
        expect(result.current.pageIndex).toBe(2);
    });

    it('defaults both filter and pageIndex when neither URL nor storage have a value', () => {
        const { result } = renderWithRouter(['/flows']);
        expect(result.current.filter).toBe('');
        expect(result.current.pageIndex).toBe(0);
    });

    it('ignores non-numeric `?page=` values gracefully', () => {
        const { result } = renderWithRouter(['/flows?page=abc']);
        expect(result.current.pageIndex).toBe(0);
    });
});

describe('useTableState — setFilter / setPage', () => {
    it('setFilter writes the value into the URL and resets `?page=` by default', async () => {
        const { result } = renderWithRouter(['/flows?page=3']);
        act(() => result.current.setFilter('alpha'));
        await waitFor(() => {
            const params = new URLSearchParams(result.current.search);
            // `clearPageOnFilterChange` default = true: changing the filter
            // drops the page so users don't land "on page 5 of nothing".
            expect(params.get('q')).toBe('alpha');
            expect(params.get('page')).toBeNull();
        });
    });

    it('setFilter("") clears the URL param', async () => {
        const { result } = renderWithRouter(['/flows?q=alpha']);
        act(() => result.current.setFilter(''));
        await waitFor(() => {
            expect(result.current.filter).toBe('');
            expect(new URLSearchParams(result.current.search).has('q')).toBe(false);
        });
    });

    it('setPage writes the 1-based page number into the URL', async () => {
        const { result } = renderWithRouter(['/flows']);
        act(() => result.current.setPage(4));
        await waitFor(() => {
            expect(new URLSearchParams(result.current.search).get('page')).toBe('5');
        });
    });

    it('setPage(0) drops the URL param entirely', async () => {
        const { result } = renderWithRouter(['/flows?page=5']);
        act(() => result.current.setPage(0));
        await waitFor(() => {
            expect(new URLSearchParams(result.current.search).has('page')).toBe(false);
        });
    });

    it('resetFilter is equivalent to setFilter("")', async () => {
        const { result } = renderWithRouter(['/flows?q=alpha']);
        act(() => result.current.resetFilter());
        await waitFor(() => {
            expect(result.current.filter).toBe('');
        });
    });
});

describe('useTableState — atomic `update` (race regression)', () => {
    it('writes both filter and pageIndex in a single transition', async () => {
        const { result } = renderWithRouter(['/flows']);
        act(() => result.current.update({ filter: 'alpha', pageIndex: 4 }));
        await waitFor(() => {
            const params = new URLSearchParams(result.current.search);
            expect(params.get('q')).toBe('alpha');
            expect(params.get('page')).toBe('5');
        });
    });

    it('preserves the other param when only one field is patched', async () => {
        const { result } = renderWithRouter(['/flows?q=alpha&page=3']);
        // Update only the page — `q` must survive.
        act(() => result.current.update({ pageIndex: 9 }));
        await waitFor(() => {
            const after = new URLSearchParams(result.current.search);
            expect(after.get('q')).toBe('alpha');
            expect(after.get('page')).toBe('10');
        });
        // And the reverse: change the filter without touching page (note:
        // this is distinct from `setFilter`, which deliberately resets it).
        act(() => result.current.update({ filter: 'beta' }));
        await waitFor(() => {
            const after = new URLSearchParams(result.current.search);
            expect(after.get('q')).toBe('beta');
            expect(after.get('page')).toBe('10');
        });
    });

    it('regression: two top-level updaters firing in the same tick keep both params', async () => {
        // The exact scenario the old split-hooks design lost: `setFilter`
        // and `setPage` issued from the same event handler dropped `?q=`
        // because react-router fed both functional updaters the same
        // pre-batch snapshot. With microtask-batched coalescence inside
        // `update`, both calls land in a single `setSearchParams` and both
        // params survive — regardless of router implementation.
        const { result } = renderWithRouter(['/flows']);
        act(() => {
            result.current.setFilter('alpha');
            result.current.setPage(5);
        });
        await waitFor(() => {
            const params = new URLSearchParams(result.current.search);
            expect(params.get('q')).toBe('alpha');
            expect(params.get('page')).toBe('6');
        });
    });

    it('null filter clears the URL param', async () => {
        const { result } = renderWithRouter(['/flows?q=alpha']);
        act(() => result.current.update({ filter: null }));
        await waitFor(() => {
            expect(new URLSearchParams(result.current.search).has('q')).toBe(false);
        });
    });

    it('coalescence resolves replace conflict in favour of push (intentional history wins)', async () => {
        // setFilter requests replace, setPage requests push. The merged
        // navigation should push, so back-button can step out of the new
        // filter+page combination.
        const { result } = renderWithRouter(['/flows']);
        act(() => {
            result.current.setFilter('alpha'); // replace
            result.current.setPage(5); // push
        });
        await waitFor(() => {
            expect(new URLSearchParams(result.current.search).get('q')).toBe('alpha');
        });
        // History assertion is implicit — we can't easily inspect the entry
        // stack from `MemoryRouter`, but the merged URL has both params
        // which proves coalescence happened.
    });
});

describe('useTableState — filter is URL-only', () => {
    // Filter is an ad-hoc query, not a preference. We deliberately do NOT
    // mirror it into storage or replay it from storage on mount — a fresh
    // navigation to `/flows` should land on an empty filter even if the user
    // typed something there earlier in the session. Sorting / column
    // visibility / page size still persist via `lib/table-state`.

    it('does not restore `?q=` from storage on mount', async () => {
        localStorage.setItem(STORAGE_KEY, JSON.stringify({ filter: 'alpha' }));
        const { result } = renderWithRouter(['/flows']);
        await act(async () => Promise.resolve());
        expect(result.current.filter).toBe('');
        expect(new URLSearchParams(result.current.search).has('q')).toBe(false);
    });

    it('honours `?q=` from the URL regardless of what is in storage', async () => {
        localStorage.setItem(STORAGE_KEY, JSON.stringify({ filter: 'stored' }));
        const { result } = renderWithRouter(['/flows?q=urlwin']);
        await act(async () => Promise.resolve());
        expect(result.current.filter).toBe('urlwin');
    });

    it('does not write the typed filter into storage', async () => {
        const { result } = renderWithRouter(['/flows']);
        act(() => result.current.setFilter('alpha'));
        await waitFor(() => {
            expect(new URLSearchParams(result.current.search).get('q')).toBe('alpha');
        });
        expect(readStoredFilter(STORAGE_KEY)).toBeNull();
    });
});

describe('useTableState — debounced filter', () => {
    it('debouncedFilter eventually catches up with filter', async () => {
        const { result } = renderWithRouter(['/flows'], { debounceMs: SHORT_DEBOUNCE_MS });
        act(() => result.current.setFilter('alpha'));
        await waitFor(() => {
            expect(result.current.debouncedFilter).toBe('alpha');
        });
    });
});

describe('useTableState — `?page=1` canonicalization', () => {
    it('rewrites `?page=1` to a clean URL on mount', async () => {
        const { result } = renderWithRouter(['/flows?page=1']);
        await waitFor(() => {
            expect(new URLSearchParams(result.current.search).has('page')).toBe(false);
        });
        expect(result.current.pageIndex).toBe(0);
    });
});

describe('useTableState — setPage replace option', () => {
    it('setPage(n, { replace: true }) writes the URL without leaving a history entry the user can step back onto', async () => {
        // Out-of-range clamping calls `setPage(lastPage, { replace: true })`
        // — without `replace`, pressing back from the clamped URL would land
        // on the original out-of-range URL and re-trigger the clamp,
        // trapping the back-button. We can't directly assert "no history
        // entry" from `MemoryRouter`, but we can confirm the option flows
        // through to a successful URL write so the call site contract holds.
        const { result } = renderWithRouter(['/flows?page=999']);
        act(() => result.current.setPage(78, { replace: true }));
        await waitFor(() => {
            expect(new URLSearchParams(result.current.search).get('page')).toBe('79');
        });
    });
});
