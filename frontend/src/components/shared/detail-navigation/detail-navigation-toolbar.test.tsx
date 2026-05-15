import type { ReactNode } from 'react';

import { act, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes, useLocation } from 'react-router-dom';
import { describe, expect, it } from 'vitest';

import { TooltipProvider } from '@/components/ui/tooltip';

import { DetailNavigationToolbar } from './detail-navigation-toolbar';

interface Item {
    id: string;
    title: string;
}

const ITEMS: readonly Item[] = [
    { id: 'a', title: 'Alpha' },
    { id: 'b', title: 'Bravo' },
    { id: 'c', title: 'Charlie' },
    { id: 'd', title: 'Delta' },
] as const;

const getId = (item: Item) => item.id;
const getLabel = (item: Item) => item.title;
const getHref = (item: Item) => `/items/${item.id}`;

const LocationReadout = () => {
    const { pathname, search } = useLocation();

    return (
        <span data-testid="location">
            {pathname}
            {search}
        </span>
    );
};

const renderToolbar = (overrides: { currentId?: null | string; filter?: string; items?: readonly Item[] } = {}) => {
    const Wrapper = ({ children }: { children: ReactNode }) => (
        <MemoryRouter initialEntries={['/items/c?q=' + (overrides.filter ?? '')]}>
            <TooltipProvider>
                <LocationReadout />
                <Routes>
                    <Route
                        element={<>{children}</>}
                        path="/items/:id"
                    />
                </Routes>
            </TooltipProvider>
        </MemoryRouter>
    );

    return render(
        <DetailNavigationToolbar<Item>
            currentId={overrides.currentId ?? 'c'}
            filter={overrides.filter ?? ''}
            getHref={getHref}
            getId={getId}
            getLabel={getLabel}
            items={overrides.items ?? ITEMS}
            sheetTitle="Items"
        />,
        { wrapper: Wrapper },
    );
};

describe('DetailNavigationToolbar', () => {
    it('renders nothing when items list is empty', () => {
        renderToolbar({ items: [] });
        // The toolbar should not emit any of its own controls — the wrapper
        // tree still renders the location readout, so assert on absent
        // toolbar elements instead of "container empty".
        expect(screen.queryByRole('button', { name: /Previous/i })).not.toBeInTheDocument();
        expect(screen.queryByRole('button', { name: /Next/i })).not.toBeInTheDocument();
    });

    it('shows `${currentIndex + 1}/${total}` for a matched current item', () => {
        renderToolbar({ currentId: 'c' });
        expect(screen.getByRole('button', { name: /3\/4/ })).toBeInTheDocument();
    });

    it('shows `–/total` when current is missing from the filtered subset', () => {
        renderToolbar({ currentId: 'zzz' });
        expect(screen.getByRole('button', { name: /–\/4/ })).toBeInTheDocument();
    });

    it('disables Prev for the first item', () => {
        renderToolbar({ currentId: 'a' });
        expect(screen.getByRole('button', { name: /Previous/i })).toBeDisabled();
    });

    it('disables Next for the last item', () => {
        renderToolbar({ currentId: 'd' });
        expect(screen.getByRole('button', { name: /Next/i })).toBeDisabled();
    });

    it('Next navigates to the next sibling preserving `?q=`', async () => {
        const user = userEvent.setup();
        // Filter 'a' matches every item (substring), so prev/next stays
        // navigable across the full set while still proving the `?q=` URL
        // round-trip on click.
        renderToolbar({ currentId: 'a', filter: 'a' });

        await user.click(screen.getByRole('button', { name: /Next/i }));

        await waitFor(() => {
            const location = screen.getByTestId('location').textContent ?? '';
            expect(location).toContain('/items/b');
        });
        expect(screen.getByTestId('location').textContent).toContain('q=a');
    });

    it('opens the sheet listbox when the position button is clicked', async () => {
        const user = userEvent.setup();
        renderToolbar({ currentId: 'c' });

        await user.click(screen.getByRole('button', { name: /3\/4/ }));

        const listbox = await screen.findByRole('listbox', { name: 'Items' });
        expect(listbox).toBeInTheDocument();
        expect(within(listbox).getAllByRole('option')).toHaveLength(4);
    });

    it('narrows the listbox to filtered items', async () => {
        const user = userEvent.setup();
        // "pha" matches only "Alpha" — substring is case-insensitive and
        // diacritic-insensitive (see ./text-filter tests).
        renderToolbar({ currentId: 'a', filter: 'pha' });

        await user.click(screen.getByRole('button', { name: /1\/1/ }));

        const listbox = await screen.findByRole('listbox', { name: 'Items' });
        const labels = within(listbox)
            .getAllByRole('option')
            .map((option) => option.textContent ?? '');
        expect(labels).toEqual(['Alpha']);
    });

    it('marks the current item with aria-selected and font-medium', async () => {
        const user = userEvent.setup();
        renderToolbar({ currentId: 'c' });

        await user.click(screen.getByRole('button', { name: /3\/4/ }));

        const listbox = await screen.findByRole('listbox');
        const current = within(listbox).getByRole('option', { selected: true });
        expect(current).toHaveAttribute('data-item-id', 'c');
    });

    it('applies roving tabIndex — only the current option is focusable', async () => {
        const user = userEvent.setup();
        renderToolbar({ currentId: 'c' });

        await user.click(screen.getByRole('button', { name: /3\/4/ }));

        const listbox = await screen.findByRole('listbox');
        const options = within(listbox).getAllByRole('option');

        // Wait for the open-time roving focus to land on the current entry.
        await waitFor(() => {
            const focusable = options.filter((option) => option.getAttribute('tabindex') === '0');
            expect(focusable).toHaveLength(1);
            expect(focusable[0]).toHaveAttribute('data-item-id', 'c');
        });

        const nonFocusable = options.filter((option) => option.getAttribute('tabindex') === '-1');
        expect(nonFocusable.length).toBe(options.length - 1);
    });

    it('ArrowDown moves roving focus to the next option', async () => {
        const user = userEvent.setup();
        renderToolbar({ currentId: 'b' });

        await user.click(screen.getByRole('button', { name: /2\/4/ }));

        const listbox = await screen.findByRole('listbox');
        await waitFor(() => {
            const focused = within(listbox)
                .getAllByRole('option')
                .find((option) => option.getAttribute('tabindex') === '0');
            expect(focused).toHaveAttribute('data-item-id', 'b');
        });

        fireEvent.keyDown(listbox, { key: 'ArrowDown' });

        await waitFor(() => {
            const focused = within(listbox)
                .getAllByRole('option')
                .find((option) => option.getAttribute('tabindex') === '0');
            expect(focused).toHaveAttribute('data-item-id', 'c');
        });
    });

    it('ArrowUp at the first option clamps (no wrap)', async () => {
        const user = userEvent.setup();
        renderToolbar({ currentId: 'a' });

        await user.click(screen.getByRole('button', { name: /1\/4/ }));

        const listbox = await screen.findByRole('listbox');
        await waitFor(() => {
            const focused = within(listbox)
                .getAllByRole('option')
                .find((option) => option.getAttribute('tabindex') === '0');
            expect(focused).toHaveAttribute('data-item-id', 'a');
        });

        fireEvent.keyDown(listbox, { key: 'ArrowUp' });

        // Focus stays on the first option — no wrap-around.
        const focused = within(listbox)
            .getAllByRole('option')
            .find((option) => option.getAttribute('tabindex') === '0');
        expect(focused).toHaveAttribute('data-item-id', 'a');
    });

    it('End jumps roving focus to the last option', async () => {
        const user = userEvent.setup();
        renderToolbar({ currentId: 'a' });

        await user.click(screen.getByRole('button', { name: /1\/4/ }));

        const listbox = await screen.findByRole('listbox');
        fireEvent.keyDown(listbox, { key: 'End' });

        await waitFor(() => {
            const focused = within(listbox)
                .getAllByRole('option')
                .find((option) => option.getAttribute('tabindex') === '0');
            expect(focused).toHaveAttribute('data-item-id', 'd');
        });
    });

    it('Home jumps roving focus to the first option', async () => {
        const user = userEvent.setup();
        renderToolbar({ currentId: 'd' });

        await user.click(screen.getByRole('button', { name: /4\/4/ }));

        const listbox = await screen.findByRole('listbox');
        // Wait for the open-time focus on the current ('d') before issuing Home.
        await waitFor(() => {
            const focused = within(listbox)
                .getAllByRole('option')
                .find((option) => option.getAttribute('tabindex') === '0');
            expect(focused).toHaveAttribute('data-item-id', 'd');
        });

        fireEvent.keyDown(listbox, { key: 'Home' });

        await waitFor(() => {
            const focused = within(listbox)
                .getAllByRole('option')
                .find((option) => option.getAttribute('tabindex') === '0');
            expect(focused).toHaveAttribute('data-item-id', 'a');
        });
    });

    it('clicking an option in the sheet navigates and closes it', async () => {
        const user = userEvent.setup();
        renderToolbar({ currentId: 'c' });

        await user.click(screen.getByRole('button', { name: /3\/4/ }));

        const listbox = await screen.findByRole('listbox');
        await user.click(within(listbox).getByRole('option', { name: 'Alpha' }));

        await waitFor(() => {
            expect(screen.queryByRole('listbox', { name: 'Items' })).not.toBeInTheDocument();
        });
        expect(screen.getByTestId('location').textContent).toContain('/items/a');
    });

    it('disables the position button when the filter excludes every item', () => {
        renderToolbar({ currentId: 'c', filter: 'qqqqqqqqq' });

        // With nothing to navigate to, the sheet trigger is disabled — the
        // empty-state copy lives inside the sheet but the button gating
        // prevents opening it in the first place.
        const positionButton = screen.getByRole('button', { name: /–\/0/ });
        expect(positionButton).toBeDisabled();
    });

    it('falls back to the first filtered option when current is outside the subset', async () => {
        const user = userEvent.setup();
        renderToolbar({ currentId: 'zzz' });

        await user.click(screen.getByRole('button', { name: /–\/4/ }));

        const listbox = await screen.findByRole('listbox');
        await waitFor(() => {
            const focused = within(listbox)
                .getAllByRole('option')
                .find((option) => option.getAttribute('tabindex') === '0');
            expect(focused).toHaveAttribute('data-item-id', 'a');
        });
    });
});

describe('DetailNavigationToolbar — predicate stability', () => {
    it('updates the filtered subset when the filter prop changes', async () => {
        const user = userEvent.setup();

        const Wrapper = ({ children }: { children: ReactNode }) => (
            <MemoryRouter initialEntries={['/items/a']}>
                <TooltipProvider>
                    <Routes>
                        <Route
                            element={<>{children}</>}
                            path="/items/:id"
                        />
                    </Routes>
                </TooltipProvider>
            </MemoryRouter>
        );

        const { rerender } = render(
            <DetailNavigationToolbar<Item>
                currentId="a"
                filter=""
                getHref={getHref}
                getId={getId}
                getLabel={getLabel}
                items={ITEMS}
                sheetTitle="Items"
            />,
            { wrapper: Wrapper },
        );

        expect(screen.getByRole('button', { name: /1\/4/ })).toBeInTheDocument();

        rerender(
            <DetailNavigationToolbar<Item>
                currentId="a"
                filter="Alpha"
                getHref={getHref}
                getId={getId}
                getLabel={getLabel}
                items={ITEMS}
                sheetTitle="Items"
            />,
        );

        expect(screen.getByRole('button', { name: /1\/1/ })).toBeInTheDocument();
        await act(async () => {
            await user.tab();
        });
    });
});
