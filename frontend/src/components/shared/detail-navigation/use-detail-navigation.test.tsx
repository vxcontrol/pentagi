import type { ReactNode } from 'react';

import { act, renderHook, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes, useLocation, useNavigate } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

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

        expect(result.current.getId(ITEMS[1])).toBe('b');
    });
});

describe('useDetailNavigation — identity stability', () => {
    it('keeps the controller reference stable across re-renders with unchanged inputs', () => {
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

        const first = result.current;
        rerender();
        // Downstream leaf components rely on this identity stability.
        expect(result.current).toBe(first);
    });

    it('keeps `goToPrev` / `goToNext` / `handleItemSelect` identity-stable when nothing relevant changes', () => {
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

        const firstGoPrev = result.current.goToPrev;
        const firstGoNext = result.current.goToNext;
        const firstSelect = result.current.handleItemSelect;
        const firstOpen = result.current.openSheet;
        const firstSet = result.current.setSheetOpen;
        rerender();

        expect(result.current.goToPrev).toBe(firstGoPrev);
        expect(result.current.goToNext).toBe(firstGoNext);
        expect(result.current.handleItemSelect).toBe(firstSelect);
        expect(result.current.openSheet).toBe(firstOpen);
        expect(result.current.setSheetOpen).toBe(firstSet);
    });
});

describe('useDetailNavigation — filter forwarding', () => {
    it('exposes the URL `?q=` value through `controller.debouncedFilter` (debounced)', async () => {
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

        // `debouncedFilter` settles after the default 200ms debounce.
        await waitFor(() => {
            expect(result.current.debouncedFilter).toBe('alpha');
        });
    });

    it('narrows `filteredItems` to the matching subset once the filter settles', async () => {
        const { result } = renderHook(
            () =>
                useDetailNavigation<Item>({
                    currentId: 'a',
                    getHref,
                    getLabel,
                    getSearchableText,
                    items: ITEMS,
                }),
            { wrapper: renderInRoute(['/items/a?q=pha']) },
        );

        await waitFor(() => {
            expect(result.current.filteredItems.map((item) => item.id)).toEqual(['a']);
        });
        expect(result.current.total).toBe(1);
        expect(result.current.currentIndex).toBe(0);
    });
});

describe('useDetailNavigation — derived state', () => {
    it('reports prev/next neighbours for a middle item', () => {
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

        expect(result.current.currentIndex).toBe(1);
        expect(result.current.prevId).toBe('a');
        expect(result.current.nextId).toBe('c');
        expect(result.current.total).toBe(3);
        expect(result.current.hasEntries).toBe(true);
        expect(result.current.itemsEmpty).toBe(false);
        expect(result.current.positionLabel).toBe('2/3');
    });

    it('reports `-1` index and `–/total` label when currentId is not in subset', () => {
        const { result } = renderHook(
            () =>
                useDetailNavigation<Item>({
                    currentId: 'zzz',
                    getHref,
                    getLabel,
                    getSearchableText,
                    items: ITEMS,
                }),
            { wrapper: renderInRoute(['/items/zzz']) },
        );

        expect(result.current.currentIndex).toBe(-1);
        expect(result.current.prevId).toBeNull();
        expect(result.current.nextId).toBeNull();
        expect(result.current.positionLabel).toBe('–/3');
    });

    it('reports `itemsEmpty=true` and `–/0` when input list is empty', () => {
        const { result } = renderHook(
            () =>
                useDetailNavigation<Item>({
                    currentId: 'b',
                    getHref,
                    getLabel,
                    getSearchableText,
                    items: [],
                }),
            { wrapper: renderInRoute(['/items/b']) },
        );

        expect(result.current.itemsEmpty).toBe(true);
        expect(result.current.total).toBe(0);
        expect(result.current.hasEntries).toBe(false);
        expect(result.current.positionLabel).toBe('–/0');
    });
});

describe('useDetailNavigation — navigation actions', () => {
    it('goToNext() navigates to the next sibling preserving `?q=`', async () => {
        const LocationProbe = ({ onChange }: { onChange: (loc: string) => void }) => {
            const { pathname, search } = useLocation();
            onChange(`${pathname}${search}`);

            return null;
        };

        const seen: string[] = [];
        const Wrapper = ({ children }: { children: ReactNode }) => (
            <MemoryRouter initialEntries={['/items/a?q=a']}>
                <LocationProbe onChange={(loc) => seen.push(loc)} />
                <Routes>
                    <Route
                        element={<>{children}</>}
                        path="/items/:id"
                    />
                </Routes>
            </MemoryRouter>
        );

        const { result } = renderHook(
            () =>
                useDetailNavigation<Item>({
                    currentId: 'a',
                    getHref,
                    getLabel,
                    getSearchableText,
                    items: ITEMS,
                }),
            { wrapper: Wrapper },
        );

        await waitFor(() => {
            expect(result.current.debouncedFilter).toBe('a');
        });

        act(() => {
            result.current.goToNext();
        });

        await waitFor(() => {
            expect(seen.at(-1)).toContain('/items/b');
        });
        expect(seen.at(-1)).toContain('q=a');
    });

    it('goToPrev() is a no-op when prevId is null (no navigate)', async () => {
        const LocationProbe = ({ onChange }: { onChange: (loc: string) => void }) => {
            const { pathname } = useLocation();
            onChange(pathname);

            return null;
        };

        const seen: string[] = [];
        const Wrapper = ({ children }: { children: ReactNode }) => (
            <MemoryRouter initialEntries={['/items/a']}>
                <LocationProbe onChange={(loc) => seen.push(loc)} />
                <Routes>
                    <Route
                        element={<>{children}</>}
                        path="/items/:id"
                    />
                </Routes>
            </MemoryRouter>
        );

        const { result } = renderHook(
            () =>
                useDetailNavigation<Item>({
                    currentId: 'a',
                    getHref,
                    getLabel,
                    getSearchableText,
                    items: ITEMS,
                }),
            { wrapper: Wrapper },
        );

        expect(result.current.prevId).toBeNull();
        const before = seen.length;

        act(() => {
            result.current.goToPrev();
        });

        // No path change emitted.
        expect(seen.length).toBe(before);
    });

    it('handleItemSelect closes the sheet *before* navigating', async () => {
        const Wrapper = ({ children }: { children: ReactNode }) => (
            <MemoryRouter initialEntries={['/items/a']}>
                <Routes>
                    <Route
                        element={<>{children}</>}
                        path="/items/:id"
                    />
                </Routes>
            </MemoryRouter>
        );

        const { result } = renderHook(
            () =>
                useDetailNavigation<Item>({
                    currentId: 'a',
                    defaultOpen: true,
                    getHref,
                    getLabel,
                    getSearchableText,
                    items: ITEMS,
                }),
            { wrapper: Wrapper },
        );

        expect(result.current.isSheetOpen).toBe(true);

        act(() => {
            result.current.handleItemSelect(ITEMS[2]);
        });

        // Sheet closes synchronously inside the same call — the navigate
        // that follows can't race with a still-mounted sheet.
        expect(result.current.isSheetOpen).toBe(false);
    });
});

describe('useDetailNavigation — controlled sheet mode', () => {
    it('respects `open={true}` even when `setSheetOpen(false)` is called', () => {
        const onOpenChange = vi.fn();

        const { result } = renderHook(
            () =>
                useDetailNavigation<Item>({
                    currentId: 'b',
                    getHref,
                    getLabel,
                    getSearchableText,
                    items: ITEMS,
                    onOpenChange,
                    open: true,
                }),
            { wrapper: renderInRoute(['/items/b']) },
        );

        expect(result.current.isSheetOpen).toBe(true);

        act(() => {
            result.current.setSheetOpen(false);
        });

        // Parent owns the state — controller doesn't flip without a prop change.
        expect(result.current.isSheetOpen).toBe(true);
        // But `onOpenChange` fires so the parent can observe the request.
        expect(onOpenChange).toHaveBeenLastCalledWith(false);
    });

    it('toggles internal state and fires onOpenChange in uncontrolled mode', () => {
        const onOpenChange = vi.fn();

        const { result } = renderHook(
            () =>
                useDetailNavigation<Item>({
                    currentId: 'b',
                    getHref,
                    getLabel,
                    getSearchableText,
                    items: ITEMS,
                    onOpenChange,
                }),
            { wrapper: renderInRoute(['/items/b']) },
        );

        expect(result.current.isSheetOpen).toBe(false);

        act(() => {
            result.current.openSheet();
        });

        expect(result.current.isSheetOpen).toBe(true);
        expect(onOpenChange).toHaveBeenLastCalledWith(true);

        act(() => {
            result.current.closeSheet();
        });

        expect(result.current.isSheetOpen).toBe(false);
        expect(onOpenChange).toHaveBeenLastCalledWith(false);
    });

    it('honours defaultOpen=true on first render in uncontrolled mode', () => {
        const { result } = renderHook(
            () =>
                useDetailNavigation<Item>({
                    currentId: 'b',
                    defaultOpen: true,
                    getHref,
                    getLabel,
                    getSearchableText,
                    items: ITEMS,
                }),
            { wrapper: renderInRoute(['/items/b']) },
        );

        expect(result.current.isSheetOpen).toBe(true);
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

        expect(result.current.currentId).toBeNull();
        expect(result.current.currentIndex).toBe(-1);
        expect(result.current.filteredItems).toEqual(ITEMS);
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
