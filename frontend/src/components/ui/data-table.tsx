import {
    type Column,
    type ColumnDef,
    type ExpandedState,
    flexRender,
    getCoreRowModel,
    getExpandedRowModel,
    getFilteredRowModel,
    getPaginationRowModel,
    getSortedRowModel,
    type Table as ReactTable,
    type Row,
    type SortingState,
    type Updater,
    useReactTable,
    type VisibilityState,
} from '@tanstack/react-table';
import {
    ArrowDown,
    ArrowUp,
    ChevronDown,
    ChevronLeft,
    ChevronRight,
    ChevronsLeft,
    ChevronsRight,
    ListFilter,
    Search,
    X,
} from 'lucide-react';
import { Fragment, type ReactElement, type ReactNode, useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useLocation } from 'react-router-dom';

import { Button } from '@/components/ui/button';
import { ContextMenu, ContextMenuContent, ContextMenuTrigger } from '@/components/ui/context-menu';
import {
    DropdownMenu,
    DropdownMenuCheckboxItem,
    DropdownMenuContent,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupInput } from '@/components/ui/input-group';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { useEffectAfterMount } from '@/hooks/use-effect-after-mount';
import { useLatestRef } from '@/hooks/use-latest-ref';
import { usePageStorageKeys } from '@/hooks/use-page-storage-keys';
import { migrateLegacyTableState, updateTableState } from '@/lib/table-state';
import { cn } from '@/lib/utils';

interface DataTableProps<TData, TValue = unknown> {
    columns: ColumnDef<TData, TValue>[];
    columnVisibility?: VisibilityState;
    data: TData[];
    /**
     * Search target(s) for the filter input. Three modes:
     * - `string` (legacy single-column): the input searches only this column;
     *   the column-picker dropdown is not rendered. Backward-compatible with
     *   pre-multi-column call sites.
     * - `string[]` (explicit multi-column): the input searches across all
     *   listed columns with OR semantics; a "Search in" dropdown lets the
     *   user narrow the set.
     * - `undefined` (zero-config multi-column): candidate columns are picked
     *   from those with `columnDef.meta.searchable === true`. If none match,
     *   the search input is not rendered at all.
     *
     * When provided, every column id must exist in `columns`.
     */
    filterColumn?: string | string[];
    filterPlaceholder?: string;
    /**
     * Controlled filter value. When provided together with `onFilterChange`
     * the parent owns the source of truth — typically `useTableQueryFilter`
     * for URL/storage-backed filters. The value flows through TanStack's
     * `state.globalFilter`, so `DataTableFilter` stays uniform regardless
     * of whether the table is single- or multi-column.
     */
    filterValue?: string;
    initialPageSize?: number;
    initialSorting?: SortingState;
    onColumnVisibilityChange?: (visibility: VisibilityState) => void;
    onFilterChange?: (value: string) => void;
    onPageChange?: (pageIndex: number) => void;
    onRowClick?: (row: TData) => void;
    pageIndex?: number;
    renderRowContextMenu?: (row: TData) => ReactNode;
    renderSubComponent?: (props: { row: Row<TData> }) => ReactElement;
}

const PAGE_SIZE_OPTIONS = [10, 15, 20, 50, 100] as const;

const columnPickerLabel = <TData,>(column: Column<TData, unknown>): string =>
    column.columnDef.meta?.columnMenuLabel ?? column.id;

interface DataTableFilterProps<TData> {
    placeholder: string;
    table: ReactTable<TData>;
}

/**
 * Search input bound to TanStack's `state.globalFilter`. Reads/writes through
 * `table.setGlobalFilter` exclusively — `DataTable`'s `onGlobalFilterChange`
 * funnels the write to the parent (`onFilterChange`) when the table is in
 * controlled-filter mode, so this component stays uniform regardless of
 * single- vs multi-column mode.
 */
const DataTableFilter = <TData,>({ placeholder, table }: DataTableFilterProps<TData>) => {
    const filterValue = (table.getState().globalFilter as string | undefined) ?? '';

    return (
        <InputGroup className="max-w-sm">
            <InputGroupAddon>
                <Search />
            </InputGroupAddon>
            <InputGroupInput
                aria-label={placeholder}
                autoComplete="off"
                id="data-table-search"
                name="search"
                onChange={(event) => table.setGlobalFilter(event.target.value)}
                placeholder={placeholder}
                type="text"
                value={filterValue}
            />
            {filterValue ? (
                <InputGroupAddon align="inline-end">
                    <InputGroupButton
                        onClick={() => table.setGlobalFilter('')}
                        type="button"
                    >
                        <X />
                    </InputGroupButton>
                </InputGroupAddon>
            ) : null}
        </InputGroup>
    );
};

interface DataTableColumnHeaderProps<TData, TValue> {
    /** TanStack column. Sorting state and the cycle action both target it. */
    column: Column<TData, TValue>;
    /** Visible label rendered as the button text. Named after the shadcn convention. */
    title: ReactNode;
}

/**
 * Cycle a TanStack column through `none → asc → desc → none`. Pure with
 * respect to React (no hooks called) so a header `onClick` can invoke it
 * directly without `useCallback`/`useMemo` ceremony. Exported alongside
 * {@link DataTableColumnHeader} so a custom header can drive the same
 * sort cycle without duplicating the if/else.
 */
export function cycleColumnSort<TData, TValue = unknown>(column: Column<TData, TValue>): void {
    const sorted = column.getIsSorted();

    if (sorted === 'asc') {
        column.toggleSorting(true);

        return;
    }

    if (sorted === 'desc') {
        column.clearSorting();

        return;
    }

    column.toggleSorting(false);
}

function DataTable<TData, TValue = unknown>({
    columns,
    columnVisibility: externalColumnVisibility,
    data,
    filterColumn,
    filterPlaceholder = 'Filter...',
    filterValue: externalFilterValue,
    initialPageSize = 10,
    initialSorting = [],
    onColumnVisibilityChange,
    onFilterChange,
    onPageChange,
    onRowClick,
    pageIndex: externalPageIndex,
    renderRowContextMenu,
    renderSubComponent,
}: DataTableProps<TData, TValue>) {
    const isColumnVisibilityControlled = externalColumnVisibility !== undefined;
    const isPageControlled = externalPageIndex !== undefined;
    const isFilterControlled = externalFilterValue !== undefined && onFilterChange !== undefined;
    const isRowInteractive = !!onRowClick || !!renderSubComponent;

    const { pathname } = useLocation();
    // Reuse the pathname we just read instead of letting the hook subscribe
    // independently — react-router caches `useLocation` so the cost is
    // negligible, but the explicit pass keeps the data flow obvious and
    // makes it easy to migrate the table to a different storage scope (e.g.
    // a workspace-prefixed path) without grepping for every subscription.
    const { table: tableKey } = usePageStorageKeys({ pathname });

    // Run the legacy → unified migration exactly once on mount via `useState`
    // lazy init. The lazy initializer runs in a single commit, so the
    // localStorage hit happens once even under StrictMode's double-invoke.
    // After mount, the `useEffect` below re-runs migration when `tableKey`
    // rotates (route change inside a persistent layout).
    const [initialState] = useState(() => migrateLegacyTableState(pathname, tableKey));

    const [sorting, setSorting] = useState<SortingState>(() => initialState.sorting ?? initialSorting);
    const [internalGlobalFilter, setInternalGlobalFilter] = useState<string>('');
    const [searchColumns, setSearchColumns] = useState<string[]>(() => initialState.searchColumns ?? []);
    const [internalColumnVisibility, setInternalColumnVisibility] = useState<VisibilityState>(() =>
        isColumnVisibilityControlled ? {} : (initialState.columnVisibility ?? {}),
    );
    const [pagination, setPagination] = useState(() => ({
        pageIndex: isPageControlled ? (externalPageIndex ?? 0) : 0,
        pageSize: initialState.pageSize ?? initialPageSize,
    }));
    const [rowSelection, setRowSelection] = useState({});
    const [expanded, setExpanded] = useState<ExpandedState>({});

    // Resolve the set of column ids the search input may target.
    // Priority: explicit array prop > legacy single-string prop > columns with
    // `meta.searchable === true`. Falls through to `[]` when no opt-in exists,
    // which suppresses the search input entirely (see JSX below).
    const searchCandidateIds = useMemo<string[]>(() => {
        if (Array.isArray(filterColumn)) {
            return filterColumn;
        }

        if (typeof filterColumn === 'string') {
            return [filterColumn];
        }

        return columns
            .filter((column) => column.meta?.searchable === true)
            .map((column) => {
                const withId = column as { id?: string };

                if (withId.id) {
                    return withId.id;
                }

                const withAccessor = column as { accessorKey?: string };

                return typeof withAccessor.accessorKey === 'string' ? withAccessor.accessorKey : undefined;
            })
            .filter((id): id is string => typeof id === 'string');
    }, [columns, filterColumn]);

    const isMultiMode = Array.isArray(filterColumn) || (filterColumn === undefined && searchCandidateIds.length > 0);

    // Track which tableKey we've already migrated + seeded from. When the
    // key rotates (route change inside a persistent layout, or an explicit
    // override) we re-run the migration for the new path and refresh local
    // state with whatever's stored under the new key.
    const seededForKeyReference = useRef(tableKey);

    useEffect(() => {
        if (seededForKeyReference.current === tableKey) {
            return;
        }

        seededForKeyReference.current = tableKey;
        const stored = migrateLegacyTableState(pathname, tableKey);

        setSorting(stored.sorting ?? initialSorting);
        setSearchColumns(stored.searchColumns ?? []);

        if (!isColumnVisibilityControlled) {
            setInternalColumnVisibility(stored.columnVisibility ?? {});
        }

        setPagination((previous) => ({
            ...previous,
            pageSize: stored.pageSize ?? initialPageSize,
        }));
    }, [initialPageSize, initialSorting, isColumnVisibilityControlled, pathname, tableKey]);

    // Effective global filter: controlled-mode parents own the value via
    // `useTableQueryFilter` (URL `?q=`), uncontrolled tables drive it from
    // internal state. Either way TanStack reads from `state.globalFilter`
    // exclusively.
    const effectiveGlobalFilter = isFilterControlled ? (externalFilterValue ?? '') : internalGlobalFilter;

    const externalFilterValueReference = useLatestRef(externalFilterValue);

    // Funnel every TanStack global-filter change through the right sink:
    // parent callback in controlled mode, internal state otherwise. The
    // updater signature mirrors TanStack's own — we resolve the function
    // form against the latest external value before calling `onFilterChange`,
    // so the parent always sees the next string, not a function.
    const handleGlobalFilterChange = useCallback(
        (updater: Updater<string>) => {
            if (!isFilterControlled) {
                setInternalGlobalFilter(updater);

                return;
            }

            const previous = externalFilterValueReference.current ?? '';
            const next = typeof updater === 'function' ? updater(previous) : updater;
            onFilterChange?.(next);
        },
        [externalFilterValueReference, isFilterControlled, onFilterChange],
    );

    // Persist sorting + column visibility + page size into the unified
    // `table_4_<path>` slot. Skipping the first render is intentional: a
    // fresh mount would otherwise overwrite the seeded values with the
    // same payload on the next commit.
    useEffectAfterMount(() => {
        updateTableState(tableKey, { sorting });
    }, [sorting, tableKey]);

    useEffectAfterMount(() => {
        if (isColumnVisibilityControlled) {
            return;
        }

        updateTableState(tableKey, { columnVisibility: internalColumnVisibility });
    }, [internalColumnVisibility, isColumnVisibilityControlled, tableKey]);

    // Only persist a non-default pageSize. `updateTableState` clears the
    // field when handed `undefined`, so the storage slot stays empty until
    // the user actively picks a different page size. Without this guard the
    // StrictMode dev double-invoke of `useEffectAfterMount` would seed
    // `{ pageSize: 10 }` on every fresh mount — harmless in prod, noisy in
    // dev tooling and surprising when inspecting storage.
    useEffectAfterMount(() => {
        updateTableState(tableKey, {
            pageSize: pagination.pageSize === initialPageSize ? undefined : pagination.pageSize,
        });
    }, [initialPageSize, pagination.pageSize, tableKey]);

    // Empty array is the "default for everyone" sentinel — `updateTableState`
    // collapses `[]` to a delete so the storage slot stays empty until the
    // user actively narrows the search column set.
    useEffectAfterMount(() => {
        updateTableState(tableKey, { searchColumns });
    }, [searchColumns, tableKey]);

    const columnVisibility = externalColumnVisibility ?? internalColumnVisibility;

    const handleColumnVisibilityChange = useCallback(
        (updaterOrValue: ((old: VisibilityState) => VisibilityState) | VisibilityState) => {
            if (onColumnVisibilityChange) {
                const newValue =
                    typeof updaterOrValue === 'function'
                        ? updaterOrValue(externalColumnVisibility ?? {})
                        : updaterOrValue;
                onColumnVisibilityChange(newValue);
            } else {
                setInternalColumnVisibility(updaterOrValue);
            }
        },
        [onColumnVisibilityChange, externalColumnVisibility],
    );

    const previousExternalPageIndexReference = useRef(externalPageIndex);
    useEffectAfterMount(() => {
        if (externalPageIndex !== undefined && externalPageIndex !== previousExternalPageIndexReference.current) {
            previousExternalPageIndexReference.current = externalPageIndex;
            setPagination((previous) => ({ ...previous, pageIndex: externalPageIndex }));
        }
    }, [externalPageIndex]);

    const paginationReference = useLatestRef(pagination);

    const handlePaginationChange = useCallback(
        (
            updater:
                | ((old: { pageIndex: number; pageSize: number }) => { pageIndex: number; pageSize: number })
                | { pageIndex: number; pageSize: number },
        ) => {
            const current = paginationReference.current;
            const newPagination = typeof updater === 'function' ? updater(current) : updater;
            setPagination(newPagination);

            if (onPageChange && newPagination.pageIndex !== current.pageIndex) {
                onPageChange(newPagination.pageIndex);
            }
        },
        [onPageChange, paginationReference],
    );

    // Route page-size changes through the same channel as TanStack's
    // pagination mutations so `onPageChange` fires when the URL-driven
    // pageIndex needs to drop to 0. Without this, picking "All" on a high
    // page leaves `?page=5` stranded in the URL even though the display
    // correctly clamps to "Page 1 of 1".
    const handlePageSizeChange = useCallback(
        (newPageSize: number) => {
            handlePaginationChange({ pageIndex: 0, pageSize: newPageSize });
        },
        [handlePaginationChange],
    );

    // Active column set for the global filter: when the user has narrowed via
    // the picker, honour the explicit list; otherwise fall back to all
    // candidates ("empty selection = search everywhere"). Recomputed on each
    // render — cheap because it's just a couple of arrays — and consumed by
    // the closure below.
    const getColumnCanGlobalFilter = useCallback(
        (column: Column<TData, unknown>) => {
            const active = searchColumns.length > 0 ? searchColumns : searchCandidateIds;

            return active.includes(column.id);
        },
        [searchCandidateIds, searchColumns],
    );

    const table = useReactTable({
        autoResetPageIndex: false,
        columns,
        data,
        enableSortingRemoval: true,
        getColumnCanGlobalFilter,
        getCoreRowModel: getCoreRowModel(),
        getExpandedRowModel: getExpandedRowModel(),
        getFilteredRowModel: getFilteredRowModel(),
        getPaginationRowModel: getPaginationRowModel(),
        getSortedRowModel: getSortedRowModel(),
        globalFilterFn: 'includesString',
        onColumnVisibilityChange: handleColumnVisibilityChange,
        onExpandedChange: setExpanded,
        onGlobalFilterChange: handleGlobalFilterChange,
        onPaginationChange: handlePaginationChange,
        onRowSelectionChange: setRowSelection,
        onSortingChange: setSorting,
        state: {
            columnVisibility,
            expanded,
            globalFilter: effectiveGlobalFilter,
            pagination,
            rowSelection,
            sorting,
        },
    });

    // TanStack doesn't re-run the filter pipeline when only the
    // `getColumnCanGlobalFilter` predicate's closure changes — it watches
    // `state.globalFilter`. Re-set the same value through TanStack's API on
    // every `searchColumns` change so the pipeline picks up the new predicate.
    // Skip the very first render to avoid a redundant cycle on mount.
    const isFirstRefilterRender = useRef(true);
    useEffect(() => {
        if (isFirstRefilterRender.current) {
            isFirstRefilterRender.current = false;

            return;
        }

        table.setGlobalFilter((current: string) => current);
    }, [searchColumns, table]);

    const handleRowClick = useCallback(
        (row: Row<TData>) => {
            if (onRowClick) {
                onRowClick(row.original);
            } else {
                row.toggleExpanded();
            }
        },
        [onRowClick],
    );

    const pageSizeValue = pagination.pageSize >= data.length && data.length > 0 ? 'all' : String(pagination.pageSize);

    const totalRows = table.getFilteredRowModel().rows.length;
    const pageCount = table.getPageCount();

    // `pagination.pageIndex` is mirrored from `?page=` (controlled) or the
    // user's clicks (uncontrolled). Either source can point past the actual
    // end of the dataset — hand-typed URLs, a filter narrowing results, or a
    // page-size bump to "All" while we were on a high page. Derive a clamped
    // view for the display values so the user never sees "Page 999 of 31",
    // and reconcile the source of truth via the effect below.
    const safePageIndex = pageCount > 0 ? Math.min(Math.max(0, pagination.pageIndex), pageCount - 1) : 0;
    const rangeStart = totalRows > 0 ? safePageIndex * pagination.pageSize + 1 : 0;
    const rangeEnd = Math.min((safePageIndex + 1) * pagination.pageSize, totalRows);

    // Reconcile the canonical pageIndex once we know `pageCount`. The write
    // has to happen in an effect because `setSearchParams` (controlled mode)
    // and `setPagination` (uncontrolled) both mutate state outside the render
    // pipeline, which React forbids during render. The early return makes the
    // effect a no-op on every render where the URL is already in range, so
    // the cost on the happy path is one comparison.
    //
    // Controlled mode compares the URL (`externalPageIndex`) — the canonical
    // source of truth — rather than the internal mirror, because internal
    // state can drop to a clamped value via `handlePageSizeChange` while the
    // URL still points at the stale page. Uncontrolled mode keeps comparing
    // the internal pageIndex since there's no other source.
    useEffect(() => {
        if (pageCount === 0) {
            return;
        }

        if (isPageControlled) {
            if (externalPageIndex !== undefined && externalPageIndex !== safePageIndex) {
                onPageChange?.(safePageIndex);
            }

            return;
        }

        if (pagination.pageIndex !== safePageIndex) {
            setPagination((previous) => ({ ...previous, pageIndex: safePageIndex }));
        }
    }, [externalPageIndex, isPageControlled, onPageChange, pageCount, pagination.pageIndex, safePageIndex]);

    return (
        <div className="w-full">
            <div className="flex items-center gap-4 py-4">
                {searchCandidateIds.length > 0 ? (
                    <DataTableFilter
                        placeholder={filterPlaceholder}
                        table={table}
                    />
                ) : null}
                {isMultiMode && searchCandidateIds.length > 1 ? (
                    <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                            <Button variant="outline">
                                <ListFilter className="mr-2" />
                                Search in <ChevronDown className="ml-2" />
                            </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="start">
                            {searchCandidateIds.map((id) => {
                                const column = table.getColumn(id);

                                if (!column) {
                                    return null;
                                }

                                const isChecked = searchColumns.length === 0 ? true : searchColumns.includes(id);

                                return (
                                    <DropdownMenuCheckboxItem
                                        checked={isChecked}
                                        className={column.columnDef.meta?.columnMenuLabel ? undefined : 'capitalize'}
                                        key={id}
                                        onCheckedChange={(value) => {
                                            setSearchColumns((prev) => {
                                                // Treat the empty-selection
                                                // sentinel as "all candidates"
                                                // before mutating, so the user
                                                // never lands in a state where
                                                // unchecking one box silently
                                                // re-enables every column.
                                                const base = prev.length === 0 ? [...searchCandidateIds] : prev;

                                                return value
                                                    ? Array.from(new Set([id, ...base]))
                                                    : base.filter((x) => x !== id);
                                            });
                                        }}
                                        onSelect={(event) => event.preventDefault()}
                                    >
                                        {columnPickerLabel(column)}
                                    </DropdownMenuCheckboxItem>
                                );
                            })}
                        </DropdownMenuContent>
                    </DropdownMenu>
                ) : null}
                <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                        <Button
                            className="ml-auto"
                            variant="outline"
                        >
                            Columns <ChevronDown className="ml-2" />
                        </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                        {table
                            .getAllColumns()
                            .filter((column) => column.getCanHide())
                            .map((column) => (
                                <DropdownMenuCheckboxItem
                                    checked={column.getIsVisible()}
                                    className={column.columnDef.meta?.columnMenuLabel ? undefined : 'capitalize'}
                                    key={column.id}
                                    onCheckedChange={(value) => column.toggleVisibility(!!value)}
                                    onSelect={(event) => event.preventDefault()}
                                >
                                    {columnPickerLabel(column)}
                                </DropdownMenuCheckboxItem>
                            ))}
                    </DropdownMenuContent>
                </DropdownMenu>
            </div>
            <div className="rounded-md border">
                <Table>
                    <TableHeader>
                        {table.getHeaderGroups().map((headerGroup) => (
                            <TableRow key={headerGroup.id}>
                                {headerGroup.headers.map((header) => (
                                    <TableHead
                                        className={header.column.columnDef.meta?.headerClassName}
                                        key={header.id}
                                        style={
                                            header.column.columnDef.size
                                                ? {
                                                      maxWidth: header.column.columnDef.size,
                                                      minWidth: header.column.columnDef.size,
                                                      width: header.column.columnDef.size,
                                                  }
                                                : undefined
                                        }
                                    >
                                        {header.isPlaceholder
                                            ? null
                                            : flexRender(header.column.columnDef.header, header.getContext())}
                                    </TableHead>
                                ))}
                            </TableRow>
                        ))}
                    </TableHeader>
                    <TableBody>
                        {table.getRowModel().rows.length > 0 ? (
                            table.getRowModel().rows.map((row) => {
                                const contextMenuContent = renderRowContextMenu?.(row.original);

                                const tableRow = (
                                    <TableRow
                                        className={cn('group hover:bg-muted/50', isRowInteractive && 'cursor-pointer')}
                                        data-state={row.getIsSelected() && 'selected'}
                                        onClick={() => handleRowClick(row)}
                                    >
                                        {row.getVisibleCells().map((cell) => (
                                            <TableCell
                                                className={cell.column.columnDef.meta?.cellClassName}
                                                key={cell.id}
                                                onClick={(event) => {
                                                    if (cell.column.columnDef.meta?.preventRowClick) {
                                                        event.stopPropagation();
                                                    }
                                                }}
                                                style={
                                                    cell.column.columnDef.size
                                                        ? {
                                                              maxWidth: cell.column.columnDef.size,
                                                              minWidth: cell.column.columnDef.size,
                                                              width: cell.column.columnDef.size,
                                                          }
                                                        : undefined
                                                }
                                            >
                                                {flexRender(cell.column.columnDef.cell, cell.getContext())}
                                            </TableCell>
                                        ))}
                                    </TableRow>
                                );

                                return (
                                    <Fragment key={row.id}>
                                        {contextMenuContent ? (
                                            <ContextMenu>
                                                <ContextMenuTrigger asChild>{tableRow}</ContextMenuTrigger>
                                                <ContextMenuContent>{contextMenuContent}</ContextMenuContent>
                                            </ContextMenu>
                                        ) : (
                                            tableRow
                                        )}
                                        {row.getIsExpanded() && renderSubComponent && (
                                            <TableRow className="cursor-default border-0 hover:bg-transparent">
                                                <TableCell
                                                    className="p-0"
                                                    colSpan={row.getVisibleCells().length}
                                                >
                                                    {renderSubComponent({ row })}
                                                </TableCell>
                                            </TableRow>
                                        )}
                                    </Fragment>
                                );
                            })
                        ) : (
                            <TableRow>
                                <TableCell
                                    className="h-24 text-center"
                                    colSpan={columns.length}
                                >
                                    No results.
                                </TableCell>
                            </TableRow>
                        )}
                    </TableBody>
                </Table>
            </div>
            <div className="flex flex-wrap items-center justify-between gap-4 px-4 py-4">
                <div className="text-muted-foreground flex-1 text-xs text-nowrap">
                    {totalRows > 0 ? (
                        <>
                            Showing {rangeStart}–{rangeEnd} of {totalRows}
                        </>
                    ) : (
                        'No results'
                    )}
                </div>
                <div className="flex items-center gap-2">
                    <span className="text-xs font-medium">Rows per page</span>
                    <Select
                        onValueChange={(value) => {
                            const pageSize = value === 'all' ? data.length : Number.parseInt(value, 10);
                            handlePageSizeChange(pageSize);
                        }}
                        value={pageSizeValue}
                    >
                        <SelectTrigger className="h-7 w-16 text-xs">
                            <SelectValue />
                        </SelectTrigger>
                        <SelectContent
                            className="min-w-16"
                            side="top"
                        >
                            {PAGE_SIZE_OPTIONS.map((size) => (
                                <SelectItem
                                    key={size}
                                    value={String(size)}
                                >
                                    {size}
                                </SelectItem>
                            ))}
                            <SelectItem value="all">All</SelectItem>
                        </SelectContent>
                    </Select>
                </div>
                {pageCount > 0 ? (
                    <div className="flex items-center justify-center text-xs font-medium lg:w-20">
                        Page {safePageIndex + 1} of {pageCount}
                    </div>
                ) : (
                    <div
                        aria-hidden
                        className="lg:w-20"
                    />
                )}
                <div className="flex items-center gap-1">
                    <Button
                        disabled={!table.getCanPreviousPage()}
                        onClick={() => table.firstPage()}
                        size="icon-xs"
                        variant="outline"
                    >
                        <ChevronsLeft />
                    </Button>
                    <Button
                        disabled={!table.getCanPreviousPage()}
                        onClick={() => table.previousPage()}
                        size="icon-xs"
                        variant="outline"
                    >
                        <ChevronLeft />
                    </Button>
                    <Button
                        disabled={!table.getCanNextPage()}
                        onClick={() => table.nextPage()}
                        size="icon-xs"
                        variant="outline"
                    >
                        <ChevronRight />
                    </Button>
                    <Button
                        disabled={!table.getCanNextPage()}
                        onClick={() => table.lastPage()}
                        size="icon-xs"
                        variant="outline"
                    >
                        <ChevronsRight />
                    </Button>
                </div>
            </div>
        </div>
    );
}

/**
 * Reusable header for sortable `DataTable` columns. Wraps the conventional
 * "label + asc/desc arrow + onClick → cycleColumnSort" pattern so every list
 * page reuses one component instead of repeating ~20 lines per column header.
 *
 * The header reads sort direction directly from `column.getIsSorted()` so the
 * caller does not have to plumb it. The arrow icon mirrors the TanStack
 * convention: `asc` shows ↓ ("ascending continues to grow downward") and
 * `desc` shows ↑. No arrow means "not sorted by this column".
 *
 * Naming follows the canonical shadcn `DataTableColumnHeader` (see
 * https://ui.shadcn.com/docs/components/data-table) so call sites read like
 * the upstream docs — only the click model is simpler here (none → asc →
 * desc → none toggle vs. a dropdown menu).
 */
function DataTableColumnHeader<TData, TValue = unknown>({ column, title }: DataTableColumnHeaderProps<TData, TValue>) {
    const sorted = column.getIsSorted();

    return (
        <Button
            className="text-muted-foreground hover:text-primary flex items-center gap-2 p-0 no-underline hover:no-underline"
            onClick={() => cycleColumnSort(column)}
            variant="link"
        >
            {title}
            {sorted === 'asc' ? <ArrowDown className="size-4" /> : null}
            {sorted === 'desc' ? <ArrowUp className="size-4" /> : null}
        </Button>
    );
}

export { DataTable, DataTableColumnHeader };
