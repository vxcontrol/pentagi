import type { Column, ColumnDef } from '@tanstack/react-table';
import type { ReactNode } from 'react';

import { act, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { useState } from 'react';
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

    it('invokes `onFilterChange` with the typed value after the debounce settles', async () => {
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

        // The input debounces its commit to the parent — intermediate values
        // never reach `onFilterChange`. After the debounce settles, the parent
        // sees exactly one call carrying the final value.
        await waitFor(() => {
            expect(onFilterChange).toHaveBeenCalled();
        });
        const lastCall = onFilterChange.mock.calls.at(-1);
        expect(lastCall?.[0]).toBe('al');
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

    // The next four tests exercise every race window the inline-arrow +
    // useDebouncedValue + two-effects design used to leak through. They
    // share a `Host` shape — a parent that mirrors the controlled value to
    // local state and hands `DataTableFilter` a fresh inline `onFilterChange`
    // per render (the same pattern real callers use through
    // `useTableQueryFilter`). The tests assert on the sequence of emitted
    // values rather than on internal state, so they remain valid even if
    // the underlying implementation changes shape again.

    interface FilterHostProps {
        emitted: string[];
        initialValue?: string;
        onSetValueRef?: (setter: (next: string) => void) => void;
    }

    const FilterHost = ({ emitted, initialValue = '', onSetValueRef }: FilterHostProps) => {
        const [value, setValue] = useState(initialValue);

        // Expose `setValue` for tests that need to mutate the controlled
        // prop without typing into the input (simulating an external source
        // like back-button or sibling control).
        if (onSetValueRef) {
            onSetValueRef(setValue);
        }

        return (
            <DataTable<Row>
                columns={COLUMNS}
                data={ROWS}
                filterColumn="name"
                filterPlaceholder="Filter..."
                filterValue={value}
                onFilterChange={(next) => {
                    emitted.push(next);
                    setValue(next);
                }}
            />
        );
    };

    it('does not resurrect the previous query after the X clears past a settled debounce', async () => {
        // Regression for the X-clear race seen in the browser:
        // typed → debounce flushes → URL=?q=alpha → X click. The old design
        // re-emitted 'alpha' on the next parent render (new inline handler,
        // stale `debouncedValue`). With the new single-timer path, X clear
        // cancels any in-flight emit synchronously.
        const emitted: string[] = [];
        const user = userEvent.setup();

        render(<FilterHost emitted={emitted} />, { wrapper: Wrapper });

        const input = screen.getByPlaceholderText('Filter...');
        await user.type(input, 'alpha');

        await waitFor(() => expect(emitted.at(-1)).toBe('alpha'));

        const inputGroup = input.closest('[data-slot="input-group"]');
        const clearButton = within(inputGroup as HTMLElement).getByRole('button');

        await user.click(clearButton);

        await new Promise((resolve) => setTimeout(resolve, 400));

        const after = emitted.slice(emitted.indexOf('alpha') + 1);
        expect(after).not.toContain('alpha');
        expect(after.at(-1)).toBe('');
        expect((input as HTMLInputElement).value).toBe('');
    });

    it('drops a pending typed emit when the X is clicked before the debounce flushes', async () => {
        // The pre-refactor code relied on `useDebouncedValue`'s internal
        // setTimeout, which `handleClear` had no way to cancel. So if a user
        // typed quickly and immediately clicked X, the still-pending timer
        // would land *after* the clear, re-emitting the typed value. The
        // refactor cancels the pending timer inside `handleClear` → `emit`.
        const emitted: string[] = [];
        const user = userEvent.setup();

        render(<FilterHost emitted={emitted} />, { wrapper: Wrapper });

        const input = screen.getByPlaceholderText('Filter...');
        await user.type(input, 'bravo');
        // No `waitFor` here — we want the X click to happen mid-debounce,
        // before any emit has reached the parent.
        expect(emitted).toEqual([]);

        const inputGroup = input.closest('[data-slot="input-group"]');
        const clearButton = within(inputGroup as HTMLElement).getByRole('button');

        await user.click(clearButton);

        // Wait long enough for any stale timer to have fired by now.
        await new Promise((resolve) => setTimeout(resolve, 400));

        // The clear may emit '' (if the parent's filterValue was previously
        // non-empty — irrelevant here, but harmless), but it must never
        // emit any partial of 'bravo'.
        expect(emitted.some((value) => value.length > 0)).toBe(false);
        expect((input as HTMLInputElement).value).toBe('');
    });

    it('lets an external change win when it lands mid-debounce', async () => {
        // If the parent flips `filterValue` while we're still holding a
        // pending debounce of locally-typed text, the external value wins
        // and the pending emit is dropped. Without this guarantee, the
        // pending timer would later fire and clobber the external value.
        const emitted: string[] = [];
        const user = userEvent.setup();
        let setValueExternal: ((next: string) => void) | undefined;

        render(
            <FilterHost
                emitted={emitted}
                onSetValueRef={(setter) => {
                    setValueExternal = setter;
                }}
            />,
            { wrapper: Wrapper },
        );

        const input = screen.getByPlaceholderText('Filter...');
        await user.type(input, 'cha');

        // While the debounce is still pending, simulate an external write
        // (e.g. back-button popping to a different `?q=`).
        expect(setValueExternal).toBeDefined();
        act(() => setValueExternal!('external'));

        // Give any leftover timer enough time to fire.
        await new Promise((resolve) => setTimeout(resolve, 400));

        // The pending typed value must never have escaped upstream.
        expect(emitted).not.toContain('cha');
        // The input snaps to the external value.
        expect((input as HTMLInputElement).value).toBe('external');
    });

    it('honours typed input when it eventually matches an unrelated external value', async () => {
        // A subtler invariant: if the user types something that happens to
        // equal a value the parent previously set externally, the typed
        // emit must still fire — we shouldn't dedupe against a value the
        // parent already knows about, only against the value we last sent.
        const emitted: string[] = [];
        const user = userEvent.setup();

        render(
            <FilterHost
                emitted={emitted}
                initialValue="seed"
            />,
            { wrapper: Wrapper },
        );

        const input = screen.getByPlaceholderText('Filter...');
        // Clear what's in the input and type the same string fresh.
        await user.clear(input);
        await waitFor(() => expect(emitted.at(-1)).toBe(''));
        await user.type(input, 'seed');

        await waitFor(() => expect(emitted.at(-1)).toBe('seed'));
        expect((input as HTMLInputElement).value).toBe('seed');
    });

    it('mounts with a deep-linked filterValue and shows the X clear button immediately', () => {
        // Deep links land on the page with `?q=alpha` already in the URL —
        // the parent passes that value down at first render, the input should
        // be populated, and the trailing X should be available right away.
        const emitted: string[] = [];

        render(
            <FilterHost
                emitted={emitted}
                initialValue="alpha"
            />,
            { wrapper: Wrapper },
        );

        const input = screen.getByPlaceholderText('Filter...') as HTMLInputElement;
        expect(input.value).toBe('alpha');
        // No emission happened — we accepted the prop, we didn't echo it back.
        expect(emitted).toEqual([]);

        // X button is rendered because the input is non-empty.
        const inputGroup = input.closest('[data-slot="input-group"]');
        expect(within(inputGroup as HTMLElement).getByRole('button')).toBeInTheDocument();
    });

    it('coalesces burst typing into a single emit with the last-typed value', async () => {
        // We type six characters back-to-back. The 150ms debounce should
        // collapse the burst to a single outbound call — sequential prefix
        // commits would defeat the debounce and hammer the URL writer.
        const emitted: string[] = [];
        const user = userEvent.setup();

        render(<FilterHost emitted={emitted} />, { wrapper: Wrapper });

        const input = screen.getByPlaceholderText('Filter...');
        await user.type(input, 'foobar');

        await waitFor(() => expect(emitted.at(-1)).toBe('foobar'));

        // No intermediate prefixes — only the final value lands.
        expect(emitted).toEqual(['foobar']);
    });

    it('does not write to the upstream on unmount with a pending debounce', async () => {
        // If the user types and immediately navigates away, the in-flight
        // timer should be cancelled by `useDebouncedCallback` so the dead
        // component can't push a stale value into the parent (which by then
        // may belong to a different page).
        const emitted: string[] = [];
        const user = userEvent.setup();

        const { unmount } = render(<FilterHost emitted={emitted} />, { wrapper: Wrapper });

        await user.type(screen.getByPlaceholderText('Filter...'), 'gone');

        // Unmount before the debounce settles.
        unmount();

        // Give any leftover timer time to misfire.
        await new Promise((resolve) => setTimeout(resolve, 400));

        // Nothing made it upstream — the timer was cancelled by the hook's
        // unmount cleanup.
        expect(emitted).toEqual([]);
    });

    it('keeps two DataTables on one page independent (settings/prompts shape)', async () => {
        // `/settings/prompts` mounts an Agent table and a Tool table on the
        // same route. Typing in one input must not leak into the other —
        // both DataTableFilter instances own private state, with separate
        // `lastEmittedReference`s and timers.
        const emittedAgent: string[] = [];
        const emittedTool: string[] = [];
        const user = userEvent.setup();

        const TwoTablesHost = () => {
            const [agent, setAgent] = useState('');
            const [tool, setTool] = useState('');

            return (
                <>
                    <DataTable<Row>
                        columns={COLUMNS}
                        data={ROWS}
                        filterColumn="name"
                        filterPlaceholder="Agent filter..."
                        filterValue={agent}
                        onFilterChange={(next) => {
                            emittedAgent.push(next);
                            setAgent(next);
                        }}
                    />
                    <DataTable<Row>
                        columns={COLUMNS}
                        data={ROWS}
                        filterColumn="name"
                        filterPlaceholder="Tool filter..."
                        filterValue={tool}
                        onFilterChange={(next) => {
                            emittedTool.push(next);
                            setTool(next);
                        }}
                    />
                </>
            );
        };

        render(<TwoTablesHost />, { wrapper: Wrapper });

        const agentInput = screen.getByPlaceholderText('Agent filter...');
        const toolInput = screen.getByPlaceholderText('Tool filter...');

        await user.type(agentInput, 'agX');
        await waitFor(() => expect(emittedAgent.at(-1)).toBe('agX'));

        // Tool table never saw anything.
        expect(emittedTool).toEqual([]);
        expect((toolInput as HTMLInputElement).value).toBe('');

        // Now type in the tool input; agent must stay frozen.
        await user.type(toolInput, 'toY');
        await waitFor(() => expect(emittedTool.at(-1)).toBe('toY'));

        expect((agentInput as HTMLInputElement).value).toBe('agX');
        expect(emittedAgent).toEqual(['agX']);
    });

    it('accepts a back-button-style external clear without re-emitting the old value', async () => {
        // History pop: user typed `'foo'`, then hit back. The router resets
        // `filterValue` to `''`. The input should snap to `''` without
        // re-emitting `'foo'` from a stale timer or sync effect.
        const emitted: string[] = [];
        const user = userEvent.setup();
        let setExternal: ((next: string) => void) | undefined;

        render(
            <FilterHost
                emitted={emitted}
                onSetValueRef={(setter) => {
                    setExternal = setter;
                }}
            />,
            { wrapper: Wrapper },
        );

        const input = screen.getByPlaceholderText('Filter...');
        await user.type(input, 'foo');
        await waitFor(() => expect(emitted.at(-1)).toBe('foo'));

        // Simulate the browser back button clearing `?q=`. The external
        // setter bypasses `onFilterChange` (which is the realistic shape —
        // a real router pop doesn't echo through our handler), so we measure
        // the absence of further emits rather than a `''` arrival.
        const indexOfFoo = emitted.lastIndexOf('foo');
        expect(setExternal).toBeDefined();
        act(() => setExternal!(''));

        await new Promise((resolve) => setTimeout(resolve, 400));

        // The input cleared; no leftover timer pushed `'foo'` back upstream.
        expect((input as HTMLInputElement).value).toBe('');
        expect(emitted.slice(indexOfFoo + 1)).not.toContain('foo');
    });

    it('survives rapid alternation between typing and X clicks', async () => {
        // Stress: type, clear, type, clear — three rounds back to back.
        // Each round must end with a clean `''` upstream and an empty input.
        const emitted: string[] = [];
        const user = userEvent.setup();

        render(<FilterHost emitted={emitted} />, { wrapper: Wrapper });

        const input = screen.getByPlaceholderText('Filter...');

        for (const round of ['one', 'two', 'three']) {
            await user.type(input, round);
            await waitFor(() => expect(emitted.at(-1)).toBe(round));

            const inputGroup = input.closest('[data-slot="input-group"]');
            const clearButton = within(inputGroup as HTMLElement).getByRole('button');
            await user.click(clearButton);
            await waitFor(() => expect((input as HTMLInputElement).value).toBe(''));
        }

        await new Promise((resolve) => setTimeout(resolve, 400));

        // Final state: input empty, last emit is the clear from round three.
        expect((input as HTMLInputElement).value).toBe('');
        expect(emitted.at(-1)).toBe('');
        // Every round's value made it through cleanly.
        expect(emitted).toContain('one');
        expect(emitted).toContain('two');
        expect(emitted).toContain('three');
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

        // Wait for the input's debounced commit; Bravo was in the initial
        // render, so the only authoritative signal is the disappearance of
        // the non-matching row.
        await waitFor(() => {
            expect(screen.queryByText('Alpha')).not.toBeInTheDocument();
        });
        expect(screen.getByText('Bravo')).toBeInTheDocument();
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
    it('asks the parent to clamp an out-of-range controlled pageIndex on mount, with replace semantics', () => {
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
        // is out of range AND the internal mirror (initialized from the prop)
        // is out of range too — both sources agree, so the parent must be told
        // to drop to 0. The `replace: true` flag is critical: the user never
        // intentionally landed on the out-of-range URL, so we don't want a
        // history entry that would trap the back-button on the bad URL.
        expect(onPageChange).toHaveBeenCalledWith(0, { replace: true });
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
        // hear about pageIndex 0 so `?page=2` is dropped from the URL. This
        // is an intentional user action — no `replace` flag, so back-button
        // can step out of the new page-size choice.
        expect(onPageChange).toHaveBeenCalledWith(0);
        // And it must be the *only* call. Regression: an earlier version
        // raced the prop-sync effect against the clamping effect and emitted
        // a second `onPageChange(lastPageIndex, { replace: true })` from the
        // clamping path while the URL was still pointing at the stale page —
        // which oscillated `/flows?page=N` → `/flows` → `/flows?page=N` on
        // every render.
        expect(onPageChange).toHaveBeenCalledTimes(1);
    });

    it('does not echo a stale pageIndex back to the URL after the parent navigates to a valid page', async () => {
        // Regression for the original infinite-render bug: on browser-back
        // from `?page=2` to `/flows`, the controlled `pageIndex` prop drops
        // from 1 → 0 in the parent, but the table's internal mirror is still
        // 1 for one render (the prop-sync effect runs after commit). A naive
        // clamping effect compared the URL (0) to its internal mirror (1) and
        // pushed onPageChange(1), which sent the URL back to `?page=2` — and
        // then the cycle repeated. With the "both agree on out-of-range"
        // guard we must NOT call `onPageChange` here, since neither source
        // is actually out of range.
        const manyRows: Row[] = Array.from({ length: 25 }, (_, index) => ({
            id: String(index + 1),
            name: `Row ${index + 1}`,
        }));
        const onPageChange = vi.fn();

        const { rerender } = render(
            <DataTable<Row>
                columns={COLUMNS}
                data={manyRows}
                onPageChange={onPageChange}
                pageIndex={1}
            />,
            { wrapper: Wrapper },
        );

        // No spurious reconcile on the initial mount — page 2 of 3 is valid.
        expect(onPageChange).not.toHaveBeenCalled();

        // Simulate the popstate landing in the parent: re-render with the
        // new in-range `pageIndex={0}`. The internal mirror is still 1 here.
        rerender(
            <DataTable<Row>
                columns={COLUMNS}
                data={manyRows}
                onPageChange={onPageChange}
                pageIndex={0}
            />,
        );

        // Flush all pending effects and microtasks so both the prop-sync and
        // clamping effects get a chance to fire. If the clamping logic still
        // raced the prop-sync, it would have called `onPageChange(1)` by now.
        await act(async () => {
            await Promise.resolve();
        });

        expect(onPageChange).not.toHaveBeenCalled();
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

    it('renders the bare "No results." cell when `empty.entityName` is omitted (legacy fallback)', () => {
        render(
            <DataTable<Row>
                columns={COLUMNS}
                data={[]}
            />,
            { wrapper: Wrapper },
        );

        expect(screen.getByText('No results.')).toBeInTheDocument();
        // The shadcn Empty title is NOT rendered without `entityName`.
        expect(screen.queryByText('No matches')).not.toBeInTheDocument();
    });

    it('renders a data-empty Empty block ("No <entity> yet") when `entityName` is set and no filter is active', () => {
        render(
            <DataTable<Row>
                columns={COLUMNS}
                data={[]}
                empty={{ entityName: 'flows' }}
            />,
            { wrapper: Wrapper },
        );

        expect(screen.getByText('No flows yet')).toBeInTheDocument();
        // No filter ⇒ no "match" description.
        expect(screen.queryByText(/Try a different query/)).not.toBeInTheDocument();
        // Pagination footer also uses the entity copy.
        expect(screen.getByText('No flows')).toBeInTheDocument();
    });

    it('renders a filter-empty Empty block ("No <entity> match …") when `entityName` is set and a filter is active', async () => {
        const user = userEvent.setup();
        render(
            <DataTable<Row>
                columns={COLUMNS}
                data={ROWS}
                empty={{ entityName: 'flows' }}
                filterColumn="name"
                filterPlaceholder="Filter..."
            />,
            { wrapper: Wrapper },
        );

        const input = screen.getByPlaceholderText('Filter...');
        // No row in ROWS has the substring "ZZZZZZ" — guarantees an empty subset.
        await user.type(input, 'ZZZZZZ');

        expect(await screen.findByText('No matches')).toBeInTheDocument();
        // Description cites the query and offers the next-step hint.
        const description = await screen.findByText(/Try a different query/);
        expect(description).toHaveTextContent('ZZZZZZ');
        expect(description.textContent).toMatch(/^No flows match/);
    });

    it('assigns a non-empty unique id and matching name to the filter field', () => {
        const { unmount } = render(
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

        const firstInput = screen.getByRole('textbox');
        const firstId = firstInput.getAttribute('id');
        // `useId` returns a non-empty stable string; `name` mirrors it so pages
        // with multiple DataTables don't collide on either attribute.
        expect(firstId).toBeTruthy();
        expect(firstInput).toHaveAttribute('name', firstId);
        unmount();

        // Render a second instance and confirm the id is different — proves
        // multi-table pages get unique ids per instance.
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
        const secondInput = screen.getByRole('textbox');
        expect(secondInput.getAttribute('id')).not.toBe(firstId);
    });

    it('caps the filter input length so a paste of multi-KB content cannot blow past URL limits', () => {
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

        const input = screen.getByRole('textbox') as HTMLInputElement;
        // The DOM `maxLength` is the only choke point we need — `<input>`
        // truncates both typing and paste at this boundary, which keeps
        // shared `?q=` URLs under the practical reverse-proxy limit (~2–4 KB).
        expect(input.maxLength).toBe(200);
    });
});

interface MultiRow {
    id: string;
    name: string;
    role: string;
}

const MULTI_ROWS: MultiRow[] = [
    { id: 'a', name: 'Alpha', role: 'admin' },
    { id: 'b', name: 'Bravo', role: 'user' },
    { id: 'c', name: 'Charlie', role: 'reader' },
];

const MULTI_COLUMNS: ColumnDef<MultiRow>[] = [
    { accessorKey: 'id', header: 'ID' },
    { accessorKey: 'name', header: 'Name', meta: { searchable: true } },
    { accessorKey: 'role', header: 'Role', meta: { searchable: true } },
];

describe('DataTable — multi-column search', () => {
    it('searches across all candidate columns with OR semantics (meta.searchable opt-in)', async () => {
        const user = userEvent.setup();
        render(
            <DataTable<MultiRow>
                columns={MULTI_COLUMNS}
                data={MULTI_ROWS}
                filterPlaceholder="Filter..."
            />,
            { wrapper: Wrapper },
        );

        const input = screen.getByPlaceholderText('Filter...');
        // "reader" only appears in the `role` column — multi-column search
        // must surface Charlie even though her `name` doesn't contain it.
        await user.type(input, 'reader');

        // Wait for the non-matching rows to disappear; Alpha and Bravo are
        // present in the initial render, so finding Charlie is not enough —
        // we need to confirm the filter actually narrowed the row set.
        await waitFor(() => {
            expect(screen.queryByText('Alpha')).not.toBeInTheDocument();
        });
        expect(screen.queryByText('Bravo')).not.toBeInTheDocument();
        expect(screen.getByText('Charlie')).toBeInTheDocument();
    });

    it('narrows the search when the picker disables a candidate', async () => {
        const user = userEvent.setup();
        render(
            <DataTable<MultiRow>
                columns={MULTI_COLUMNS}
                data={MULTI_ROWS}
                filterPlaceholder="Filter..."
            />,
            { wrapper: Wrapper },
        );

        await user.click(screen.getByRole('button', { name: /Search in/ }));
        // Uncheck the `Role` column so "reader" can no longer match Charlie.
        await user.click(await screen.findByRole('menuitemcheckbox', { name: /role/i }));

        // Close the dropdown so it doesn't intercept subsequent input focus
        // events. Pressing Escape is the user-facing way out.
        await user.keyboard('{Escape}');

        const input = screen.getByPlaceholderText('Filter...');
        await user.type(input, 'reader');

        expect(await screen.findByText('No results.')).toBeInTheDocument();
    });

    it('persists the narrowed search column set to the unified storage slot', async () => {
        const user = userEvent.setup();
        render(
            <DataTable<MultiRow>
                columns={MULTI_COLUMNS}
                data={MULTI_ROWS}
            />,
            { wrapper: Wrapper },
        );

        await user.click(screen.getByRole('button', { name: /Search in/ }));
        await user.click(await screen.findByRole('menuitemcheckbox', { name: /role/i }));

        const stored = JSON.parse(localStorage.getItem('table_4_/flows') ?? '{}');
        expect(stored.searchColumns).toEqual(['name']);
    });

    it('refilters immediately when the picker disables a column that an already-typed query matched (regression for d4c1b13)', async () => {
        const user = userEvent.setup();
        render(
            <DataTable<MultiRow>
                columns={MULTI_COLUMNS}
                data={MULTI_ROWS}
                filterPlaceholder="Filter..."
            />,
            { wrapper: Wrapper },
        );

        // "admin" appears only in `role` on Alpha. Confirm the multi-column
        // search surfaces her *before* we narrow the picker.
        await user.type(screen.getByPlaceholderText('Filter...'), 'admin');
        expect(await screen.findByText('Alpha')).toBeInTheDocument();

        // Now uncheck `Role`. The previous closure-based implementation kept
        // Alpha visible here because `state.globalFilter` stayed equal and
        // TanStack short-circuited the filter pipeline. The composite-value
        // implementation produces a new `globalFilter` object reference, so
        // the pipeline re-runs and Alpha must disappear without retyping.
        await user.click(screen.getByRole('button', { name: /Search in/ }));
        await user.click(await screen.findByRole('menuitemcheckbox', { name: /role/i }));

        await waitFor(() => {
            expect(screen.queryByText('Alpha')).not.toBeInTheDocument();
        });
    });

    it('prunes stale ids from persisted searchColumns when the candidate set shrinks', () => {
        // Pre-seed storage as if a previous version of this page exposed a
        // `deleted` column the user had selected. Today's columns no longer
        // include it — the rebase effect must trim the persisted entry.
        localStorage.setItem('table_4_/flows', JSON.stringify({ searchColumns: ['name', 'role', 'deleted'] }));

        render(
            <DataTable<MultiRow>
                columns={MULTI_COLUMNS}
                data={MULTI_ROWS}
            />,
            { wrapper: Wrapper },
        );

        const stored = JSON.parse(localStorage.getItem('table_4_/flows') ?? '{}');
        expect(stored.searchColumns).toEqual(['name', 'role']);
    });

    it('activates multi-column mode when filterColumn is an array (without meta.searchable)', async () => {
        const user = userEvent.setup();
        const PLAIN_COLUMNS: ColumnDef<MultiRow>[] = [
            { accessorKey: 'id', header: 'ID' },
            { accessorKey: 'name', header: 'Name' },
            { accessorKey: 'role', header: 'Role' },
        ];

        render(
            <DataTable<MultiRow>
                columns={PLAIN_COLUMNS}
                data={MULTI_ROWS}
                filterColumn={['name', 'role']}
                filterPlaceholder="Filter..."
            />,
            { wrapper: Wrapper },
        );

        // Picker is visible — confirms multi-mode activation via the array prop.
        expect(screen.getByRole('button', { name: /Search in/ })).toBeInTheDocument();

        const input = screen.getByPlaceholderText('Filter...');
        await user.type(input, 'user');

        // "user" matches Bravo's role; Alpha and Charlie don't contain it.
        // Wait for them to disappear instead of relying on `Bravo` being
        // present (it was in the initial render too).
        await waitFor(() => {
            expect(screen.queryByText('Alpha')).not.toBeInTheDocument();
        });
        expect(screen.queryByText('Charlie')).not.toBeInTheDocument();
        expect(screen.getByText('Bravo')).toBeInTheDocument();
    });

    it('hides the input entirely when filterColumn is undefined and no column opts in', () => {
        const PLAIN_COLUMNS: ColumnDef<MultiRow>[] = [
            { accessorKey: 'id', header: 'ID' },
            { accessorKey: 'name', header: 'Name' },
            { accessorKey: 'role', header: 'Role' },
        ];

        render(
            <DataTable<MultiRow>
                columns={PLAIN_COLUMNS}
                data={MULTI_ROWS}
            />,
            { wrapper: Wrapper },
        );

        expect(screen.queryByRole('textbox')).not.toBeInTheDocument();
        expect(screen.queryByRole('button', { name: /Search in/ })).not.toBeInTheDocument();
    });

    it('does not render the picker in legacy single-column mode', () => {
        render(
            <DataTable<MultiRow>
                columns={MULTI_COLUMNS}
                data={MULTI_ROWS}
                filterColumn="name"
            />,
            { wrapper: Wrapper },
        );

        expect(screen.getByRole('textbox')).toBeInTheDocument();
        expect(screen.queryByRole('button', { name: /Search in/ })).not.toBeInTheDocument();
        // The "Columns" trigger still renders.
        expect(screen.getByRole('button', { name: /Columns/ })).toBeInTheDocument();
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

describe('DataTable — virtualization', () => {
    // 60 rows on a single page (pageSize 100) so the rendered set clears the
    // 50-row threshold and `isVirtualized` actually activates.
    const VIRTUAL_ROWS: Row[] = Array.from({ length: 60 }, (_, index) => ({
        id: String(index + 1),
        name: `Row ${index + 1}`,
    }));

    let rectSpy: ReturnType<typeof vi.spyOn>;

    beforeEach(() => {
        // jsdom reports a zero rect for everything. Pin a realistic row height
        // so the window virtualizer culls a deterministic subset instead of
        // collapsing zero-height rows and rendering all of them.
        rectSpy = vi.spyOn(HTMLElement.prototype, 'getBoundingClientRect').mockReturnValue({
            bottom: 53,
            height: 53,
            left: 0,
            right: 800,
            toJSON: () => ({}),
            top: 0,
            width: 800,
            x: 0,
            y: 0,
        } as DOMRect);
    });

    afterEach(() => {
        rectSpy.mockRestore();
    });

    // Body rows carry `data-index`; padding spacers and the header row don't.
    const dataRows = () => document.querySelectorAll('tbody tr[data-index]');

    it('renders only a subset of rows once past the threshold', async () => {
        render(
            <DataTable<Row>
                columns={COLUMNS}
                data={VIRTUAL_ROWS}
                initialPageSize={100}
                isVirtualized
            />,
            { wrapper: Wrapper },
        );

        await waitFor(() => expect(dataRows().length).toBeGreaterThan(0));
        // The viewport can't hold all 60 estimated rows, so the window virtualizer
        // culls the rest — far fewer rows reach the DOM than exist in the data.
        expect(dataRows().length).toBeLessThan(VIRTUAL_ROWS.length);
    });

    it('exposes the true row total via aria-rowcount while virtualized', async () => {
        render(
            <DataTable<Row>
                columns={COLUMNS}
                data={VIRTUAL_ROWS}
                initialPageSize={100}
                isVirtualized
            />,
            { wrapper: Wrapper },
        );

        await waitFor(() => expect(dataRows().length).toBeGreaterThan(0));
        // 60 data rows + 1 header row — what a screen reader announces as the total.
        expect(document.querySelector('table')).toHaveAttribute('aria-rowcount', '61');
        // Header occupies aria-rowindex 1; each data row is its data index + 2
        // (1-based, past the header). Assert the relationship rather than a fixed
        // index so the check is independent of which window the virtualizer renders.
        expect(document.querySelector('thead tr')).toHaveAttribute('aria-rowindex', '1');
        const firstDataRow = dataRows()[0];

        if (!firstDataRow) {
            throw new Error('no row');
        }

        const dataIndex = Number(firstDataRow.getAttribute('data-index'));
        expect(firstDataRow).toHaveAttribute('aria-rowindex', String(dataIndex + 2));
    });

    it('keeps data-index on rows wrapped in a context menu (Radix asChild path)', async () => {
        // Production /flows pairs `isVirtualized` with `renderRowContextMenu`, so
        // every virtual row is wrapped in <ContextMenuTrigger asChild>. The
        // measureElement ref and `data-index` must survive Radix's prop/ref
        // merge or dynamic measurement silently breaks.
        render(
            <DataTable<Row>
                columns={COLUMNS}
                data={VIRTUAL_ROWS}
                initialPageSize={100}
                isVirtualized
                renderRowContextMenu={(row) => <div>menu for {row.name}</div>}
            />,
            { wrapper: Wrapper },
        );

        await waitFor(() => expect(dataRows().length).toBeGreaterThan(0));
        dataRows().forEach((row) => expect(row).toHaveAttribute('data-index'));
    });

    it('does not virtualize at or below the threshold (full DOM, no aria-rowcount)', () => {
        const fewRows: Row[] = Array.from({ length: 30 }, (_, index) => ({
            id: String(index + 1),
            name: `Row ${index + 1}`,
        }));

        render(
            <DataTable<Row>
                columns={COLUMNS}
                data={fewRows}
                initialPageSize={100}
                isVirtualized
            />,
            { wrapper: Wrapper },
        );

        // Every row is in the DOM, so native enumeration works — no spacer, no
        // virtualization wiring, no aria-rowcount override.
        expect(screen.getByText('Row 1')).toBeInTheDocument();
        expect(screen.getByText('Row 30')).toBeInTheDocument();
        expect(dataRows().length).toBe(0);
        expect(document.querySelector('tbody tr[aria-hidden]')).not.toBeInTheDocument();
        expect(document.querySelector('table')).not.toHaveAttribute('aria-rowcount');
    });

    it('disables virtualization when renderSubComponent is set, even past the threshold', () => {
        render(
            <DataTable<Row>
                columns={COLUMNS}
                data={VIRTUAL_ROWS}
                initialPageSize={100}
                isVirtualized
                renderSubComponent={({ row }) => <div>expanded {row.original.name}</div>}
            />,
            { wrapper: Wrapper },
        );

        // Expanded-row support wins over virtualization: the full set renders
        // and no measurement wiring is attached.
        expect(dataRows().length).toBe(0);
        expect(screen.getByText('Row 1')).toBeInTheDocument();
        expect(screen.getByText('Row 60')).toBeInTheDocument();
        expect(document.querySelector('table')).not.toHaveAttribute('aria-rowcount');
    });
});
