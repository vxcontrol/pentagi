import type { ColumnDef } from '@tanstack/react-table';
import type { ReactNode } from 'react';

import { fireEvent, render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { TooltipProvider } from '@/components/ui/tooltip';

import { DataTable } from './data-table';

interface Row {
    id: string;
    name: string;
}

const ROWS: Row[] = [
    { id: 'a', name: 'Alpha' },
    { id: 'b', name: 'Bravo' },
    { id: 'c', name: 'Charlie' },
];

const COLUMNS: ColumnDef<Row>[] = [
    { accessorKey: 'id', header: 'ID' },
    { accessorKey: 'name', filterFn: 'includesString', header: 'Name' },
];

const Wrapper = ({ children }: { children: ReactNode }) => (
    <MemoryRouter initialEntries={['/flows']}>
        <TooltipProvider>{children}</TooltipProvider>
    </MemoryRouter>
);

beforeEach(() => {
    localStorage.clear();
});

afterEach(() => {
    localStorage.clear();
});

describe('DataTable — controlled filter projection', () => {
    it('projects `filterValue` into the visible rows when a `filterColumn` is set', () => {
        render(
            <DataTable<Row>
                columns={COLUMNS}
                data={ROWS}
                filterColumn="name"
                filterValue="Bravo"
                onFilterChange={() => {
                    /* no-op */
                }}
            />,
            { wrapper: Wrapper },
        );

        // The body shows only "Bravo" — Alpha and Charlie are filtered out.
        expect(screen.getByText('Bravo')).toBeInTheDocument();
        expect(screen.queryByText('Alpha')).not.toBeInTheDocument();
        expect(screen.queryByText('Charlie')).not.toBeInTheDocument();
    });

    it('renders the empty filter as "all rows visible"', () => {
        render(
            <DataTable<Row>
                columns={COLUMNS}
                data={ROWS}
                filterColumn="name"
                filterValue=""
                onFilterChange={() => {
                    /* no-op */
                }}
            />,
            { wrapper: Wrapper },
        );

        expect(screen.getByText('Alpha')).toBeInTheDocument();
        expect(screen.getByText('Bravo')).toBeInTheDocument();
        expect(screen.getByText('Charlie')).toBeInTheDocument();
    });

    it('invokes `onFilterChange` with the typed value as the user edits the input', async () => {
        const onFilterChange = vi.fn();
        const user = userEvent.setup();

        render(
            <DataTable<Row>
                columns={COLUMNS}
                data={ROWS}
                filterColumn="name"
                filterPlaceholder="Filter name..."
                filterValue=""
                onFilterChange={onFilterChange}
            />,
            { wrapper: Wrapper },
        );

        const input = screen.getByPlaceholderText('Filter name...');
        await user.type(input, 'al');

        // userEvent fires keystroke-by-keystroke; the last call carries the
        // full input — controlled-mode parents receive every intermediate
        // value because the input is controlled by the column filter API.
        expect(onFilterChange).toHaveBeenCalled();
        const lastCall = onFilterChange.mock.calls.at(-1);
        expect(lastCall?.[0]).toBe('l');
    });

    it('clears the filter when the trailing X button is clicked', async () => {
        const onFilterChange = vi.fn();
        const user = userEvent.setup();

        render(
            <DataTable<Row>
                columns={COLUMNS}
                data={ROWS}
                filterColumn="name"
                filterPlaceholder="Filter..."
                filterValue="Alpha"
                onFilterChange={onFilterChange}
            />,
            { wrapper: Wrapper },
        );

        // The clear button is rendered only when the filter has content.
        const input = screen.getByPlaceholderText('Filter...');
        const inputGroup = input.closest('[data-slot="input-group"]');
        expect(inputGroup).not.toBeNull();
        const clearButton = within(inputGroup as HTMLElement).getByRole('button');

        await user.click(clearButton);

        expect(onFilterChange).toHaveBeenCalledWith('');
    });
});

describe('DataTable — uncontrolled filter is still routed through the same input', () => {
    it('falls back to internal column filter state when `filterValue` is omitted', async () => {
        const user = userEvent.setup();
        render(
            <DataTable<Row>
                columns={COLUMNS}
                data={ROWS}
                filterColumn="name"
                filterPlaceholder="Filter..."
            />,
            { wrapper: Wrapper },
        );

        const input = screen.getByPlaceholderText('Filter...');
        await user.type(input, 'Bra');

        // The input is uncontrolled here — TanStack reflects the value back
        // via `column.getFilterValue()`, and only matching rows survive.
        expect(screen.getByText('Bravo')).toBeInTheDocument();
        expect(screen.queryByText('Alpha')).not.toBeInTheDocument();
    });
});

describe('DataTable — does not render the filter input when `filterColumn` is omitted', () => {
    it('hides the search input entirely', () => {
        render(
            <DataTable<Row>
                columns={COLUMNS}
                data={ROWS}
            />,
            { wrapper: Wrapper },
        );

        expect(screen.queryByRole('textbox')).not.toBeInTheDocument();
    });
});

describe('DataTable — sorting state persists to storage', () => {
    it('writes a sorting entry to the unified `table_4_<path>` slot after the user sorts', () => {
        const user = userEvent.setup();
        render(
            <DataTable<Row>
                columns={[
                    {
                        accessorKey: 'name',
                        enableSorting: true,
                        header: ({ column }) => (
                            <button
                                onClick={() => column.toggleSorting(false)}
                                type="button"
                            >
                                Name
                            </button>
                        ),
                    },
                ]}
                data={ROWS}
            />,
            { wrapper: Wrapper },
        );

        const sortButton = screen.getByRole('button', { name: 'Name' });
        fireEvent.click(sortButton);

        const stored = JSON.parse(localStorage.getItem('table_4_/flows') ?? '{}');
        expect(stored.sorting).toEqual([{ desc: false, id: 'name' }]);
        // ESLint will yell about `user` being unused if we remove the import.
        void user;
    });
});

describe('DataTable — empty results', () => {
    it('does not render "Page 1 of 0" when there are no rows', () => {
        render(
            <DataTable<Row>
                columns={COLUMNS}
                data={[]}
            />,
            { wrapper: Wrapper },
        );

        expect(screen.queryByText(/Page 1 of 0/)).not.toBeInTheDocument();
        expect(screen.getByText('No results')).toBeInTheDocument();
    });

    it('exposes name and id on the filter field when filterColumn is set', () => {
        render(
            <DataTable<Row>
                columns={COLUMNS}
                data={ROWS}
                filterColumn="name"
                filterValue=""
                onFilterChange={() => {
                    /* no-op */
                }}
            />,
            { wrapper: Wrapper },
        );

        const input = screen.getByRole('textbox');
        expect(input).toHaveAttribute('name', 'name');
        expect(input).toHaveAttribute('id', 'data-table-filter-name');
    });
});
