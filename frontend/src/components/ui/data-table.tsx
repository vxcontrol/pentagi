import {
    type Column,
    type ColumnDef,
    type ExpandedState,
    type FilterFn,
    flexRender,
    getCoreRowModel,
    getExpandedRowModel,
    getFilteredRowModel,
    getPaginationRowModel,
    getSortedRowModel,
    type Row,
    type SortingState,
    useReactTable,
    type VisibilityState,
} from '@tanstack/react-table';
import {
    ArrowDown,
    ArrowUp,
    ChevronLeft,
    ChevronRight,
    ChevronsLeft,
    ChevronsRight,
    ColumnsSettings,
    Inbox,
    ListFilter,
    Search,
    X,
} from 'lucide-react';
import {
    type ChangeEvent,
    Fragment,
    type ReactElement,
    type ReactNode,
    useCallback,
    useDeferredValue,
    useEffect,
    useId,
    useMemo,
    useRef,
    useState,
} from 'react';
import { useLocation } from 'react-router-dom';
import { useDebouncedCallback } from 'use-debounce';

import { Button } from '@/components/ui/button';
import { ContextMenu, ContextMenuContent, ContextMenuTrigger } from '@/components/ui/context-menu';
import {
    DropdownMenu,
    DropdownMenuCheckboxItem,
    DropdownMenuContent,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';
import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupInput } from '@/components/ui/input-group';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { useEffectAfterMount } from '@/hooks/use-effect-after-mount';
import { useLatestRef } from '@/hooks/use-latest-ref';
import { usePageStorageKeys } from '@/hooks/use-page-storage-keys';
import { migrateLegacyTableState, updateTableState } from '@/lib/table-state';
import { cn } from '@/lib/utils';

/**
 * Composite value stored in TanStack's `state.globalFilter`. Bundling `query`
 * and the active `columns` set together makes the predicate self-contained:
 * any change to either field produces a new object reference, which is what
 * TanStack watches to re-run the filter pipeline. This sidesteps the trap
 * where a closure-only narrowing (`getColumnCanGlobalFilter`) updates silently
 * and the rows stay stale.
 */
type DataTableGlobalFilter = { columns: string[]; query: string };

interface DataTableProps<TData, TValue = unknown> {
    columns: ColumnDef<TData, TValue>[];
    columnVisibility?: VisibilityState;
    data: TData[];
    /** Plural lowercase, e.g. `"flows"`, `"knowledge documents"`, `"API tokens"`. */
    empty?: { entityName?: string };
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
    onPageChange?: (pageIndex: number, options?: { replace?: boolean }) => void;
    onRowClick?: (row: TData) => void;
    pageIndex?: number;
    renderRowContextMenu?: (row: TData) => ReactNode;
    renderSubComponent?: (props: { row: Row<TData> }) => ReactElement;
    /**
     * Storage slot for sorting / column visibility / page size / search-column
     * narrowing. Defaults to `usePageStorageKeys().table` — i.e.
     * `table_4_<pathname>` — which is correct for any route that owns exactly
     * one DataTable. Pages that mount multiple DataTables on the same route
     * (e.g. `/settings/prompts`) must pass distinct keys per instance, or
     * their persisted state will alias and overwrite each other.
     *
     * Recommended composition for multi-table routes — take the route base
     * through `usePageStorageKeys` and append a per-table suffix instead of
     * hard-coding the `table_4_<path>` prefix:
     *
     * ```tsx
     * const { table: base } = usePageStorageKeys();
     * <DataTable storageKey={`${base}:agents`} … />
     * <DataTable storageKey={`${base}:tools`} … />
     * ```
     */
    storageKey?: string;
}

const PAGE_SIZE_OPTIONS = [10, 15, 20, 50, 100] as const;

const columnPickerLabel = <TData,>(column: Column<TData, unknown>): string =>
    column.columnDef.meta?.columnMenuLabel ?? column.id;

/**
 * Resolve a `ColumnDef`'s id the way TanStack does internally: explicit `id`
 * wins, then `accessorKey` when it's a plain string. Display columns and
 * `accessorFn` columns without an explicit id resolve to `undefined` —
 * callers filter those out before passing the id to APIs that require one.
 */
const getColumnId = <TData, TValue>(column: ColumnDef<TData, TValue>): string | undefined => {
    const withId = column as { id?: string };

    if (withId.id) {
        return withId.id;
    }

    const withAccessor = column as { accessorKey?: string };

    return typeof withAccessor.accessorKey === 'string' ? withAccessor.accessorKey : undefined;
};

interface DataTableEmptyStateProps {
    entityName?: string;
    filterValue: string;
}

interface DataTableFilterProps {
    onQueryChange: (value: string) => void;
    placeholder: string;
    query: string;
}

function DataTableEmptyState({ entityName, filterValue }: DataTableEmptyStateProps) {
    if (!entityName) {
        return <>No results.</>;
    }

    const hasFilter = filterValue.length > 0;
    const Icon = hasFilter ? Search : Inbox;

    return (
        <Empty className="border-0 p-0 md:p-0">
            <EmptyHeader>
                <EmptyMedia variant="icon">
                    <Icon />
                </EmptyMedia>
                <EmptyTitle>{hasFilter ? 'No matches' : `No ${entityName} yet`}</EmptyTitle>
                {hasFilter ? (
                    <EmptyDescription>
                        No {entityName} match <code>{filterValue}</code>. Try a different query.
                    </EmptyDescription>
                ) : null}
            </EmptyHeader>
        </Empty>
    );
}

const FILTER_DEBOUNCE_MS = 150;
// Hard cap on the filter query length. 200 chars is more than any realistic
// search term and protects against pathological inputs (paste of a multi-KB
// chunk) that would otherwise blow past URL limits — browsers handle ~8 KB,
// reverse-proxies typically cap at 2–4 KB, so a 5 KB share-link becomes
// unreliable. `<input maxLength>` truncates typing and paste at the DOM
// boundary, which is the only entry point users have here.
const FILTER_MAX_LENGTH = 200;

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
    empty,
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
    storageKey: explicitStorageKey,
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
    //
    // When `storageKey` is passed by the parent it wins — multi-table routes
    // (e.g. /settings/prompts) need distinct slots per instance, otherwise
    // their sorting / visibility / search-column narrowing alias and
    // overwrite each other under the shared `table_4_<path>` key.
    const { table: defaultTableKey } = usePageStorageKeys({ pathname });
    const tableKey = explicitStorageKey ?? defaultTableKey;

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
            .map(getColumnId)
            .filter((id): id is string => typeof id === 'string');
    }, [columns, filterColumn]);

    const isMultiMode = Array.isArray(filterColumn) || (filterColumn === undefined && searchCandidateIds.length > 0);

    // Rebase `searchColumns` whenever the set of candidates changes (column
    // reconfiguration, HMR, or a hydrated localStorage entry that references
    // ids the current page no longer exposes). The empty-array sentinel — our
    // "search everywhere" state — is preserved as-is; non-empty selections are
    // pruned to the still-valid intersection. Return `prev` unchanged when the
    // intersection equals the input so we don't trigger a spurious storage
    // write through the persistence effect below.
    useEffect(() => {
        setSearchColumns((prev) => {
            if (prev.length === 0) {
                return prev;
            }

            const filtered = prev.filter((id) => searchCandidateIds.includes(id));

            return filtered.length === prev.length ? prev : filtered;
        });
    }, [searchCandidateIds]);

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

    // Compose the query and the active column set into a single TanStack
    // state value. New object identity on any change is exactly what TanStack
    // watches to re-run the filter pipeline — no imperative `setGlobalFilter`
    // pokes, no closure-only narrowing that updates silently. The `columns`
    // field follows the same "empty selection = search everywhere" sentinel
    // as `searchColumns`, resolved once here for the predicate's convenience.
    const effectiveQuery = isFilterControlled ? (externalFilterValue ?? '') : internalGlobalFilter;

    const effectiveGlobalFilter = useMemo<DataTableGlobalFilter>(
        () => ({
            columns: searchColumns.length > 0 ? searchColumns : searchCandidateIds,
            query: effectiveQuery,
        }),
        [effectiveQuery, searchCandidateIds, searchColumns],
    );

    const effectiveGlobalFilterReference = useLatestRef(effectiveGlobalFilter);

    // Hand the filter pipeline a deferred snapshot so TanStack's row-model
    // recomputation runs at low priority. With large datasets (≈350 flows,
    // 180 knowledges) every keystroke through the debounce previously committed
    // the heavy `getFilteredRowModel()` pass on the urgent track, occasionally
    // dropping the next keystroke. `useDeferredValue` lets React surface the
    // stale rows + responsive UI while the new filter resolves in the
    // background; once it commits, the deferred snapshot catches up.
    const deferredGlobalFilter = useDeferredValue(effectiveGlobalFilter);

    // Funnel every TanStack global-filter change through the right sink. The
    // updater arrives in three shapes:
    //   * a raw string from `DataTableFilter` (input.onChange → setGlobalFilter)
    //   * a function `(prev: DataTableGlobalFilter) => DataTableGlobalFilter`
    //     from any TanStack-internal mutation (e.g. `table.resetGlobalFilter`)
    //   * a bare composite object from a programmatic write
    // We resolve it down to the `query` string at the boundary — controlled
    // parents and internal state both speak strings.
    const handleGlobalFilterChange = useCallback(
        (updater: unknown) => {
            const resolveQuery = (value: unknown): string => {
                if (typeof value === 'string') {
                    return value;
                }

                if (value && typeof value === 'object' && 'query' in value) {
                    return String((value as { query: unknown }).query ?? '');
                }

                return '';
            };

            const nextRaw =
                typeof updater === 'function'
                    ? (updater as (previous: DataTableGlobalFilter) => unknown)(effectiveGlobalFilterReference.current)
                    : updater;
            const nextQuery = resolveQuery(nextRaw);

            if (isFilterControlled) {
                onFilterChange?.(nextQuery);
            } else {
                setInternalGlobalFilter(nextQuery);
            }
        },
        [effectiveGlobalFilterReference, isFilterControlled, onFilterChange],
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

    // Case-insensitive substring predicate that consults the composite
    // filter for both "what to look for" and "where to look". Returning
    // `true` for the empty query keeps the default "no filter = all rows"
    // behaviour identical to TanStack's built-in `includesString`.
    const globalFilterFn = useCallback<FilterFn<TData>>((row, columnId, filter: DataTableGlobalFilter) => {
        if (!filter.query) {
            return true;
        }

        if (!filter.columns.includes(columnId)) {
            return false;
        }

        const value = row.getValue(columnId);

        return String(value ?? '')
            .toLowerCase()
            .includes(filter.query.toLowerCase());
    }, []);

    const table = useReactTable({
        autoResetPageIndex: false,
        columns,
        data,
        enableSortingRemoval: true,
        getCoreRowModel: getCoreRowModel(),
        getExpandedRowModel: getExpandedRowModel(),
        getFilteredRowModel: getFilteredRowModel(),
        getPaginationRowModel: getPaginationRowModel(),
        getSortedRowModel: getSortedRowModel(),
        globalFilterFn,
        onColumnVisibilityChange: handleColumnVisibilityChange,
        onExpandedChange: setExpanded,
        onGlobalFilterChange: handleGlobalFilterChange,
        onPaginationChange: handlePaginationChange,
        onRowSelectionChange: setRowSelection,
        onSortingChange: setSorting,
        state: {
            columnVisibility,
            expanded,
            globalFilter: deferredGlobalFilter,
            pagination,
            rowSelection,
            sorting,
        },
    });

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

    // Controlled clamping requires both URL and internal mirror to agree on
    // out-of-range. The two update asynchronously in opposite orders (popstate
    // URL-first, page-size mirror-first), so a single-source check echoes the
    // stale value over the fresh one and oscillates.
    useEffect(() => {
        if (pageCount === 0) {
            return;
        }

        const lastPageIndex = pageCount - 1;

        if (isPageControlled) {
            if (externalPageIndex !== undefined) {
                const externalOutOfRange = externalPageIndex < 0 || externalPageIndex > lastPageIndex;
                const internalOutOfRange = pagination.pageIndex < 0 || pagination.pageIndex > lastPageIndex;

                if (externalOutOfRange && internalOutOfRange && externalPageIndex !== lastPageIndex) {
                    onPageChange?.(lastPageIndex, { replace: true });
                }
            }

            return;
        }

        if (pagination.pageIndex !== safePageIndex) {
            setPagination((previous) => ({ ...previous, pageIndex: safePageIndex }));
        }
    }, [externalPageIndex, isPageControlled, onPageChange, pageCount, pagination.pageIndex, safePageIndex]);

    return (
        <div className="w-full">
            <div className="flex items-center gap-2 py-4">
                {searchCandidateIds.length > 0 ? (
                    <DataTableFilter
                        onQueryChange={(value) => table.setGlobalFilter(value)}
                        placeholder={filterPlaceholder}
                        query={effectiveQuery}
                    />
                ) : null}
                {isMultiMode && searchCandidateIds.length > 1 ? (
                    <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                            <Button
                                aria-label="Search in"
                                className="shrink-0"
                                size="icon"
                                variant="outline"
                            >
                                <ListFilter />
                            </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
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
                            aria-label="Columns"
                            className="ml-auto shrink-0"
                            size="icon"
                            variant="outline"
                        >
                            <ColumnsSettings />
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
                                        className={cn(
                                            'group hover:bg-muted/50 data-[state=open]:bg-muted/50 has-[[data-state=open]]:bg-muted/50',
                                            isRowInteractive && 'cursor-pointer',
                                            contextMenuContent &&
                                                'pointer-coarse:select-none pointer-coarse:[-webkit-touch-callout:none]',
                                        )}
                                        {...(row.getIsSelected() ? { 'data-state': 'selected' } : {})}
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
                                    className={cn('text-center', empty?.entityName ? 'py-12' : 'h-24')}
                                    colSpan={columns.length}
                                >
                                    <DataTableEmptyState
                                        entityName={empty?.entityName}
                                        filterValue={effectiveQuery.trim()}
                                    />
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
                    ) : empty?.entityName ? (
                        `No ${empty.entityName}`
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
                    <div className="flex items-center justify-center text-xs font-medium lg:w-24">
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
                        aria-label="First page"
                        disabled={!table.getCanPreviousPage()}
                        onClick={() => table.firstPage()}
                        size="icon-xs"
                        variant="outline"
                    >
                        <ChevronsLeft />
                    </Button>
                    <Button
                        aria-label="Previous page"
                        disabled={!table.getCanPreviousPage()}
                        onClick={() => table.previousPage()}
                        size="icon-xs"
                        variant="outline"
                    >
                        <ChevronLeft />
                    </Button>
                    <Button
                        aria-label="Next page"
                        disabled={!table.getCanNextPage()}
                        onClick={() => table.nextPage()}
                        size="icon-xs"
                        variant="outline"
                    >
                        <ChevronRight />
                    </Button>
                    <Button
                        aria-label="Last page"
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

/**
 * Search input for the table's global filter.
 *
 * Controlled debounced input — the kind the React docs warn you cannot
 * model with `useDeferredValue` alone, because we want to throttle the
 * *side-effect* (URL writes via the parent's `onQueryChange`) rather than
 * the render of the filtered rows (`useDeferredValue` already handles the
 * latter at the table level — see `deferredGlobalFilter` above). React's
 * own recommendation for this case is plain debouncing.
 *
 * The whole component is two state cells and one effect:
 *
 *   1. `localValue` (`useState`) — drives the controlled input, updates
 *      synchronously on every keystroke so the caret never lags.
 *   2. `lastEmittedReference` (`useRef`) — records the most recent value
 *      we've handed upstream so the external-sync effect can distinguish
 *      "this is our own commit coming back through the router" (skip)
 *      from "something outside changed the source of truth" (accept,
 *      override any in-flight typing).
 *   3. One `useEffect` reconciling external `query` with local state.
 *
 * Everything timer-related lives inside `useDebouncedCallback`, which
 * exposes a single `.cancel()` we call from the only two places that need
 * to override the schedule — external sync and `handleClear`. There is no
 * stale `debouncedValue` cell for an effect to read out of phase, which
 * was the entire class of races the previous design carried.
 */
function DataTableFilter({ onQueryChange, placeholder, query }: DataTableFilterProps) {
    const [localValue, setLocalValue] = useState(query);
    const lastEmittedReference = useRef(query);
    // Generated per-instance so pages with multiple DataTables (e.g.
    // /settings/prompts) don't end up with duplicate `id` attributes — that
    // breaks `getElementById`, a11y semantics, and any test selector that
    // relies on the input id.
    const fieldId = useId();

    const debouncedEmit = useDebouncedCallback((next: string) => {
        if (next === lastEmittedReference.current) {
            return;
        }

        lastEmittedReference.current = next;
        onQueryChange(next);
    }, FILTER_DEBOUNCE_MS);

    // External → local sync. Runs only when `query` is something we didn't
    // emit ourselves: back button, programmatic clear, sibling control.
    // Any in-flight debounce of locally-typed text is cancelled
    // synchronously; the external value wins, by design.
    useEffect(() => {
        if (query === lastEmittedReference.current) {
            return;
        }

        debouncedEmit.cancel();
        lastEmittedReference.current = query;
        setLocalValue(query);
    }, [query, debouncedEmit]);

    const handleChange = useCallback(
        (event: ChangeEvent<HTMLInputElement>) => {
            setLocalValue(event.target.value);
            debouncedEmit(event.target.value);
        },
        [debouncedEmit],
    );

    const handleClear = useCallback(() => {
        setLocalValue('');
        debouncedEmit.cancel();

        if (lastEmittedReference.current !== '') {
            lastEmittedReference.current = '';
            onQueryChange('');
        }
    }, [debouncedEmit, onQueryChange]);

    return (
        <InputGroup className="max-w-sm">
            <InputGroupInput
                aria-label={placeholder}
                autoComplete="off"
                id={fieldId}
                maxLength={FILTER_MAX_LENGTH}
                name={fieldId}
                onChange={handleChange}
                placeholder={placeholder}
                type="text"
                value={localValue}
            />
            {localValue ? (
                <InputGroupAddon align="inline-end">
                    <InputGroupButton
                        onClick={handleClear}
                        type="button"
                    >
                        <X />
                    </InputGroupButton>
                </InputGroupAddon>
            ) : null}
        </InputGroup>
    );
}

export { DataTable, DataTableColumnHeader };
