import type { ReactNode } from 'react';

import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes, useLocation } from 'react-router-dom';
import { describe, expect, it } from 'vitest';

import { TooltipProvider } from '@/components/ui/tooltip';

import { DetailNavigationToolbar } from './detail-navigation-toolbar';
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

const ToolbarHarness = ({ currentId = 'c', items = ITEMS }: HarnessProps) => {
    const nav = useDetailNavigation<Item>({
        currentId,
        getHref,
        getLabel,
        items,
    });

    return (
        <DetailNavigationToolbar<Item>
            controller={nav}
            sheetTitle="Items"
        />
    );
};

const renderToolbar = (props: HarnessProps = {}) => {
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

    return render(<ToolbarHarness {...props} />, { wrapper: Wrapper });
};

describe('DetailNavigationToolbar', () => {
    it('renders nothing when raw items is empty', () => {
        renderToolbar({ items: [] });
        expect(screen.queryByRole('button', { name: /Previous/i })).not.toBeInTheDocument();
        expect(screen.queryByRole('button', { name: /Next/i })).not.toBeInTheDocument();
    });

    it('composes Buttons + Sheet: position button opens the listbox', async () => {
        const user = userEvent.setup();
        renderToolbar({ currentId: 'c' });

        // Buttons present (smoke).
        expect(screen.getByRole('button', { name: /Previous/i })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /Next/i })).toBeInTheDocument();

        // Position trigger opens the sheet.
        await user.click(screen.getByRole('button', { name: /3\/4/ }));
        const listbox = await screen.findByRole('listbox', { name: 'Items' });
        expect(listbox).toBeInTheDocument();
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
        renderToolbar({ currentId: 'a', filter: 'a' });

        await user.click(screen.getByRole('button', { name: /Next/i }));

        await waitFor(() => {
            expect(screen.getByTestId('location').textContent).toContain('/items/b');
        });
        expect(screen.getByTestId('location').textContent).toContain('q=a');
    });

    it('disables the position button when the filter excludes every item', async () => {
        renderToolbar({ currentId: 'c', filter: 'qqqqqqqqq' });

        // After debounce settles, the empty subset disables the trigger.
        await waitFor(() => {
            expect(screen.getByRole('button', { name: /–\/0/ })).toBeDisabled();
        });
    });
});
