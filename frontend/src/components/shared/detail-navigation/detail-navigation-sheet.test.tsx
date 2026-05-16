import type { ReactNode } from 'react';

import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes, useLocation } from 'react-router-dom';
import { describe, expect, it } from 'vitest';

import { TooltipProvider } from '@/components/ui/tooltip';

import { DetailNavigationSheet } from './detail-navigation-sheet';
import { useDetailNavigation } from './use-detail-navigation';

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

const getHref = (item: Item) => `/items/${item.id}`;
const getLabel = (item: Item) => item.title;
const getSearchableText = (item: Item) => item.title;

const LocationReadout = () => {
    const { pathname, search } = useLocation();

    return (
        <span data-testid="location">
            {pathname}
            {search}
        </span>
    );
};

interface HarnessProps {
    currentId?: null | string;
    filter?: string;
    items?: readonly Item[];
}

/**
 * Render the sheet open by default (`defaultOpen: true`) so keyboard /
 * focus / a11y interactions can be exercised without round-tripping through
 * the toolbar's position button. Keeps each test focused on the leaf.
 */
const SheetHarness = ({ currentId = 'c', items = ITEMS }: HarnessProps) => {
    const nav = useDetailNavigation<Item>({
        currentId,
        defaultOpen: true,
        getHref,
        getLabel,
        getSearchableText,
        items,
    });

    return (
        <DetailNavigationSheet<Item>
            controller={nav}
            sheetTitle="Items"
        />
    );
};

const renderSheet = (props: HarnessProps = {}) => {
    const filter = props.filter ?? '';
    const initialId = props.currentId ?? 'c';

    const Wrapper = ({ children }: { children: ReactNode }) => (
        <MemoryRouter initialEntries={[`/items/${initialId}?q=${filter}`]}>
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

    return render(<SheetHarness {...props} />, { wrapper: Wrapper });
};

describe('DetailNavigationSheet — a11y / aria contract', () => {
    it('renders the listbox with the sheet title as accessible name (aria-describedby opt-out preserved)', async () => {
        renderSheet({ currentId: 'c' });

        const listbox = await screen.findByRole('listbox', { name: 'Items' });
        expect(listbox).toBeInTheDocument();
    });

    it('marks the current item with aria-selected', async () => {
        renderSheet({ currentId: 'c' });

        const listbox = await screen.findByRole('listbox');
        const current = within(listbox).getByRole('option', { selected: true });
        expect(current).toHaveAttribute('data-item-id', 'c');
    });
});

describe('DetailNavigationSheet — roving tabIndex', () => {
    it('only the current option carries tabIndex={0}', async () => {
        renderSheet({ currentId: 'c' });

        const listbox = await screen.findByRole('listbox');
        const options = within(listbox).getAllByRole('option');

        await waitFor(() => {
            const focusable = options.filter((option) => option.getAttribute('tabindex') === '0');
            expect(focusable).toHaveLength(1);
            expect(focusable[0]).toHaveAttribute('data-item-id', 'c');
        });

        const nonFocusable = options.filter((option) => option.getAttribute('tabindex') === '-1');
        expect(nonFocusable.length).toBe(options.length - 1);
    });

    it('falls back to the first filtered option when current is outside the subset', async () => {
        renderSheet({ currentId: 'zzz' });

        const listbox = await screen.findByRole('listbox');
        await waitFor(() => {
            const focused = within(listbox)
                .getAllByRole('option')
                .find((option) => option.getAttribute('tabindex') === '0');
            expect(focused).toHaveAttribute('data-item-id', 'a');
        });
    });
});

describe('DetailNavigationSheet — keyboard navigation', () => {
    it('ArrowDown moves roving focus to the next option', async () => {
        renderSheet({ currentId: 'b' });

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
        renderSheet({ currentId: 'a' });

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
        renderSheet({ currentId: 'a' });

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
        renderSheet({ currentId: 'd' });

        const listbox = await screen.findByRole('listbox');
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
});

describe('DetailNavigationSheet — selection', () => {
    it('clicking an option navigates and closes the sheet', async () => {
        const user = userEvent.setup();
        renderSheet({ currentId: 'c' });

        const listbox = await screen.findByRole('listbox');
        await user.click(within(listbox).getByRole('option', { name: 'Alpha' }));

        await waitFor(() => {
            expect(screen.queryByRole('listbox', { name: 'Items' })).not.toBeInTheDocument();
        });
        expect(screen.getByTestId('location').textContent).toContain('/items/a');
    });

    it('narrows the listbox to filtered items', async () => {
        renderSheet({ currentId: 'a', filter: 'pha' });

        const listbox = await screen.findByRole('listbox', { name: 'Items' });

        await waitFor(() => {
            const labels = within(listbox)
                .getAllByRole('option')
                .map((option) => option.textContent ?? '');
            expect(labels).toEqual(['Alpha']);
        });
    });
});
