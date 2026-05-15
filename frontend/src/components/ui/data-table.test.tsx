import type { Column, ColumnDef } from '@tanstack/react-table';
import type { ReactNode } from 'react';

import { fireEvent, render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { TooltipProvider } from '@/components/ui/tooltip';

import { cycleColumnSort, DataTable } from './data-table';

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

describe('DataTable — controlled pageIndex reconciliation', () => {
    it('asks the parent to clamp an out-of-range controlled pageIndex on mount', () => {
        const onPageChange = vi.fn();

        render(
            <DataTable<Row>
                columns={COLUMNS}
                data={ROWS}
                onPageChange={onPageChange}
                pageIndex={9}
            />,
            { wrapper: Wrapper },
        );

        // 3 rows / default pageSize 10 → 1 page. The URL-driven pageIndex 9
        // is out of range, so the parent must be told to drop to 0 so the
        // canonical URL no longer points past the dataset.
        expect(onPageChange).toHaveBeenCalledWith(0);
    });

    it('clears the URL page when picking "All" on a high page in controlled mode', async () => {
        const user = userEvent.setup();
        // 15 rows × default pageSize 10 → 2 pages; start on page 2.
        const manyRows: Row[] = Array.from({ length: 15 }, (_, index) => ({
            id: String(index + 1),
            name: `Row ${index + 1}`,
        }));
        const onPageChange = vi.fn();

        render(
            <DataTable<Row>
                columns={COLUMNS}
                data={manyRows}
                onPageChange={onPageChange}
                pageIndex={1}
            />,
            { wrapper: Wrapper },
        );

        // Initial mirror of the controlled pageIndex shouldn't trigger a
        // reconcile — page 2 of 2 is in range.
        expect(onPageChange).not.toHaveBeenCalled();

        // Open the rows-per-page select and pick "All".
        const trigger = screen.getByRole('combobox');
        await user.click(trigger);
        const option = await screen.findByRole('option', { name: 'All' });
        await user.click(option);

        // "All" collapses the dataset to a single page; the parent has to
        // hear about pageIndex 0 so `?page=2` is dropped from the URL.
        expect(onPageChange).toHaveBeenCalledWith(0);
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

interface FakeColumn {
    clearSorting: ReturnType<typeof vi.fn>;
    getIsSorted: ReturnType<typeof vi.fn>;
    toggleSorting: ReturnType<typeof vi.fn>;
}

// Stub the subset of `Column` `cycleColumnSort` actually touches. The cast
// through `unknown as Column<...>` keeps the test free of @tanstack/react-table
// internals while preserving the typed call signature.
const makeColumn = (sorted: 'asc' | 'desc' | false): FakeColumn => ({
    clearSorting: vi.fn(),
    getIsSorted: vi.fn().mockReturnValue(sorted),
    toggleSorting: vi.fn(),
});

describe('cycleColumnSort', () => {
    it('starts the cycle on a column with no sort by setting ascending', () => {
        const column = makeColumn(false);
        cycleColumnSort(column as unknown as Column<unknown>);

        expect(column.toggleSorting).toHaveBeenCalledWith(false);
        expect(column.clearSorting).not.toHaveBeenCalled();
    });

    it('advances ascending to descending', () => {
        const column = makeColumn('asc');
        cycleColumnSort(column as unknown as Column<unknown>);

        expect(column.toggleSorting).toHaveBeenCalledWith(true);
        expect(column.clearSorting).not.toHaveBeenCalled();
    });

    it('clears the sort after descending — the third tap removes the sort entirely', () => {
        const column = makeColumn('desc');
        cycleColumnSort(column as unknown as Column<unknown>);

        expect(column.clearSorting).toHaveBeenCalledTimes(1);
        expect(column.toggleSorting).not.toHaveBeenCalled();
    });

    it('reads `getIsSorted` exactly once per call — no double-evaluation between branches', () => {
        const column = makeColumn('asc');
        cycleColumnSort(column as unknown as Column<unknown>);

        expect(column.getIsSorted).toHaveBeenCalledTimes(1);
    });
});
