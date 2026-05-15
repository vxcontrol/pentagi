import type { ReactNode } from 'react';

import { act, renderHook, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes, useNavigate, useSearchParams } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';

import { useDetailNavigation } from './use-detail-navigation';

interface Item {
    id: string;
    title: string;
}

const ITEMS: readonly Item[] = [
    { id: 'a', title: 'Alpha' },
    { id: 'b', title: 'Bravo' },
    { id: 'c', title: 'Charlie' },
] as const;

const getHref = (item: Item) => `/items/${item.id}`;
const getLabel = (item: Item) => item.title;
const getSearchableText = (item: Item) => item.title;

const renderInRoute = (initialEntries: string[]) => {
    const Wrapper = ({ children }: { children: ReactNode }) => (
        <MemoryRouter initialEntries={initialEntries}>
            <Routes>
                <Route
                    element={<>{children}</>}
                    path="/items/:id"
                />
            </Routes>
        </MemoryRouter>
    );

    return Wrapper;
};

beforeEach(() => {
    localStorage.clear();
});

afterEach(() => {
    localStorage.clear();
});

describe('useDetailNavigation — default getId', () => {
    it('uses item.id when no `getId` override is provided', () => {
        const { result } = renderHook(
            () =>
                useDetailNavigation<Item>({
                    currentId: 'b',
                    getHref,
                    getLabel,
                    getSearchableText,
                    items: ITEMS,
                }),
            { wrapper: renderInRoute(['/items/b']) },
        );

        expect(result.current.toolbarProps.getId(ITEMS[1])).toBe('b');
    });

    it('keeps `toolbarProps` reference stable when the caller re-renders without changes', () => {
        const { rerender, result } = renderHook(
            () =>
                useDetailNavigation<Item>({
                    currentId: 'b',
                    getHref,
                    getLabel,
                    getSearchableText,
                    items: ITEMS,
                }),
            { wrapper: renderInRoute(['/items/b']) },
        );

        const first = result.current.toolbarProps;
        rerender();
        // Downstream toolbar memos rely on this identity stability.
        expect(result.current.toolbarProps).toBe(first);
    });
});

describe('useDetailNavigation — filter forwarding', () => {
    it('exposes the URL `?q=` value through `toolbarProps.filter` (debounced)', async () => {
        const { result } = renderHook(
            () =>
                useDetailNavigation<Item>({
                    currentId: 'b',
                    getHref,
                    getLabel,
                    getSearchableText,
                    items: ITEMS,
                }),
            { wrapper: renderInRoute(['/items/b?q=alpha']) },
        );

        // `toolbarProps.filter` is the debounced value — wait for the debounce
        // to settle (default 200ms) before asserting equality.
        await waitFor(() => {
            expect(result.current.toolbarProps.filter).toBe('alpha');
        });
        expect(result.current.debouncedFilter).toBe('alpha');
    });

    it('does NOT replay `localStorage` into the URL on a fresh detail mount', async () => {
        // Detail pages use `useTableQueryFilterReader` under the hood, which
        // is explicitly storage-blind. A shared `/items/abc` link should not
        // gain a stale `?q=` from a previous tab's filter.
        localStorage.setItem('table_4_/items', JSON.stringify({ filter: 'stored' }));

        const ProbeChild = ({ onMount }: { onMount: (search: string) => void }) => {
            const [searchParams] = useSearchParams();
            onMount(searchParams.toString());

            return null;
        };

        const seen: string[] = [];
        renderHook(
            () => {
                const nav = useDetailNavigation<Item>({
                    currentId: 'b',
                    getHref,
                    getLabel,
                    getSearchableText,
                    items: ITEMS,
                });

                return nav;
            },
            { wrapper: renderInRoute(['/items/b']) },
        );

        // Drain pending tasks — any storage→URL replay would land here.
        await act(async () => {
            await Promise.resolve();
        });

        renderHook(() => null, {
            wrapper: ({ children }) => (
                <MemoryRouter initialEntries={['/items/b']}>
                    <ProbeChild onMount={(search) => seen.push(search)} />
                    {children}
                </MemoryRouter>
            ),
        });

        expect(seen.every((search) => !search.includes('q='))).toBe(true);
    });
});

describe('useDetailNavigation — current item bookkeeping', () => {
    it('treats a missing currentId as "no match" without throwing', () => {
        const { result } = renderHook(
            () =>
                useDetailNavigation<Item>({
                    currentId: undefined,
                    getHref,
                    getLabel,
                    getSearchableText,
                    items: ITEMS,
                }),
            { wrapper: renderInRoute(['/items/b']) },
        );

        expect(result.current.toolbarProps.currentId).toBeUndefined();
        expect(result.current.toolbarProps.items).toBe(ITEMS);
    });
});

describe('useDetailNavigation — navigation back to list does not leak the filter', () => {
    it('keeps the URL `?q=` intact after a navigate inside the same router', async () => {
        const Harness = () => {
            const navigate = useNavigate();
            const nav = useDetailNavigation<Item>({
                currentId: 'b',
                getHref,
                getLabel,
                getSearchableText,
                items: ITEMS,
            });

            return { nav, navigate };
        };

        const { result } = renderHook(() => Harness(), {
            wrapper: renderInRoute(['/items/b?q=alpha']),
        });

        await waitFor(() => {
            expect(result.current.nav.debouncedFilter).toBe('alpha');
        });

        // Navigating to another detail page should not alter the filter
        // value the hook observes — the URL still carries it.
        act(() => {
            result.current.navigate('/items/a?q=alpha');
        });

        await waitFor(() => {
            expect(result.current.nav.debouncedFilter).toBe('alpha');
        });
    });
});
