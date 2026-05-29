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
    defaultSearchQuery?: string;
    filter?: string;
    hasSearch?: boolean;
    items?: readonly Item[];
    searchPlaceholder?: string;
}

/**
 * Render the sheet open by default (`defaultOpen: true`) so keyboard /
 * focus / a11y interactions can be exercised without round-tripping through
 * the toolbar's position button. Keeps each test focused on the leaf.
 */
const SheetHarness = ({
    currentId = 'c',
    defaultSearchQuery,
    hasSearch,
    items = ITEMS,
    searchPlaceholder,
}: HarnessProps) => {
    const nav = useDetailNavigation<Item>({
        currentId,
        defaultOpen: true,
        defaultSearchQuery,
        getHref,
        getLabel,
        getSearchableText,
        items,
        // Skip the debounce so each test sees the post-typing subset on the
        // next paint instead of having to wait `>150ms` per keystroke.
        searchDebounceMs: 0,
    });

    return (
        <DetailNavigationSheet<Item>
            controller={nav}
            hasSearch={hasSearch}
            searchPlaceholder={searchPlaceholder}
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

describe('DetailNavigationSheet — search input', () => {
    it('renders the search input by default (hasSearch=true)', async () => {
        renderSheet({ currentId: 'c' });

        const input = await screen.findByRole('textbox');
        expect(input).toBeInTheDocument();
        expect(input).toHaveValue('');
    });

    it('hides the search input when hasSearch is false (opt-out)', async () => {
        renderSheet({ currentId: 'c', hasSearch: false });

        await screen.findByRole('listbox');
        expect(screen.queryByRole('textbox')).not.toBeInTheDocument();
    });

    it('uses the supplied placeholder', async () => {
        renderSheet({ currentId: 'c', searchPlaceholder: 'Find item' });

        const input = await screen.findByPlaceholderText('Find item');
        expect(input).toBeInTheDocument();
    });

    it('typing narrows the listbox immediately when searchDebounceMs=0', async () => {
        renderSheet({ currentId: 'a' });

        const input = await screen.findByRole('textbox');
        // `fireEvent.change` — `user.type` races with the Radix focus trap (see
        // the URL-filter AND test below for the same sidestep).
        fireEvent.change(input, { target: { value: 'cha' } });

        const listbox = await screen.findByRole('listbox');

        await waitFor(() => {
            const labels = within(listbox)
                .getAllByRole('option')
                .map((option) => option.textContent ?? '');
            expect(labels).toEqual(['Charlie']);
        });
    });

    it('renders a clear button while the query is non-empty; clicking it resets', async () => {
        const user = userEvent.setup();
        renderSheet({ currentId: 'a', defaultSearchQuery: 'cha' });

        const clearButton = await screen.findByRole('button', { name: 'Clear search' });
        await user.click(clearButton);

        const input = await screen.findByRole('textbox');
        expect(input).toHaveValue('');

        const listbox = await screen.findByRole('listbox');

        await waitFor(() => {
            const labels = within(listbox)
                .getAllByRole('option')
                .map((option) => option.textContent ?? '');
            expect(labels).toEqual(['Alpha', 'Bravo', 'Charlie', 'Delta']);
        });
    });

    it('Escape clears a non-empty query without closing the sheet', async () => {
        const user = userEvent.setup();
        renderSheet({ currentId: 'a', defaultSearchQuery: 'cha' });

        const input = await screen.findByRole('textbox');
        input.focus();
        await user.keyboard('{Escape}');

        expect(input).toHaveValue('');
        // Sheet stayed mounted because the keydown was prevented + stopped.
        expect(screen.queryByRole('listbox', { name: 'Items' })).toBeInTheDocument();
    });

    it('shows a query-specific empty state when nothing matches', async () => {
        renderSheet({ currentId: 'a', defaultSearchQuery: 'zzzzz' });

        await waitFor(() => {
            expect(screen.getByText(/No items match "zzzzz"\./)).toBeInTheDocument();
        });
        expect(screen.queryByRole('listbox', { name: 'Items' })).not.toBeInTheDocument();
    });

    it('ANDs the in-sheet search with the URL ?q= filter', async () => {
        // URL filter narrows to Alpha + Bravo + Charlie (substring "a"); the
        // local "bra" then narrows further to just Bravo. Use a single
        // `fireEvent.change` instead of `user.type` so the assertion does
        // not race with React's per-keystroke re-renders.
        renderSheet({ currentId: 'b', filter: 'a' });

        const input = await screen.findByRole('textbox');
        fireEvent.change(input, { target: { value: 'bra' } });

        const listbox = await screen.findByRole('listbox', { name: 'Items' });

        await waitFor(() => {
            const labels = within(listbox)
                .getAllByRole('option')
                .map((option) => option.textContent ?? '');
            expect(labels).toEqual(['Bravo']);
        });
    });

    it('ArrowDown from the input moves roving focus into the listbox', async () => {
        // Start at currentId 'b' so the initial roving focus is 'b'; that way
        // we can prove ArrowDown from the *input* targeted the first item ('a')
        // rather than just no-op'ing on already-focused 'a'.
        renderSheet({ currentId: 'b' });

        const input = await screen.findByRole('textbox');
        fireEvent.keyDown(input, { key: 'ArrowDown' });

        const listbox = await screen.findByRole('listbox');

        await waitFor(() => {
            const focused = within(listbox)
                .getAllByRole('option')
                .find((option) => option.getAttribute('tabindex') === '0');
            expect(focused).toHaveAttribute('data-item-id', 'a');
        });
        // ArrowDown jumps DOM focus to the new option synchronously — the
        // auto-focus effect skips when focus is on the search input, so this
        // handler has to move focus itself. Without it, the user would have
        // to click the option to interact with it after typing.
        expect(document.activeElement).toHaveAttribute('data-item-id', 'a');
    });

    it('does not yank focus onto a listbox option when filteredItems shrinks (regression)', async () => {
        // Reproduces the bug from the original implementation: typing a
        // query that drops the current item from the subset triggered the
        // auto-focus effect to grab focus onto the new `focusedId` (the
        // first survivor) at the debounce boundary, interrupting typing.
        //
        // `currentId='d'` + query `'rav'` → Delta is gone, subset becomes
        // [Bravo], render-phase reconciliation sets `focusedId='b'`. With
        // the fix the effect observes the search input is focused and
        // refuses to steal focus onto the option.
        //
        // (The exact post-shrink `activeElement` is sensitive to Radix's
        // JSDOM-only focus-trap fallback behaviour — in a real browser
        // focus stays on the input, but JSDOM may bounce it onto the
        // dialog container if the previously-focused option unmounts. The
        // bug is specifically about the listbox option stealing focus, so
        // we assert against *that*, not against equality with the input.)
        renderSheet({ currentId: 'd' });

        const input = await screen.findByRole('textbox');

        // Radix Dialog runs its own open-time focus trap. Wait for it to
        // settle, then put focus on the input as the user would after
        // clicking it.
        await waitFor(() => {
            input.focus();
            expect(document.activeElement).toBe(input);
        });

        fireEvent.change(input, { target: { value: 'rav' } });

        await waitFor(() => {
            const labels = within(screen.getByRole('listbox'))
                .getAllByRole('option')
                .map((option) => option.textContent ?? '');
            expect(labels).toEqual(['Bravo']);
        });

        // The exact failure mode of the original bug: focus jumps onto a
        // `role="option"` button when the list shrinks. With the fix it
        // never lands there until the user explicitly arrows into the list.
        expect(document.activeElement?.getAttribute('role')).not.toBe('option');
    });
});
