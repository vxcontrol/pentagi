'use client';

import {
    type ColumnDef,
    type ColumnFiltersState,
    type ExpandedState,
    flexRender,
    getCoreRowModel,
    getExpandedRowModel,
    getFilteredRowModel,
    getPaginationRowModel,
    getSortedRowModel,
    type SortingState,
    useReactTable,
    type VisibilityState,
} from '@tanstack/react-table';
import { ChevronDown } from 'lucide-react';
import * as React from 'react';

import { Button } from '@/components/ui/button';
import {
    DropdownMenu,
    DropdownMenuCheckboxItem,
    DropdownMenuContent,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';

// Extend ColumnMeta interface from @tanstack/react-table
declare module '@tanstack/react-table' {
    interface ColumnMeta<TData, TValue> {
        cellClassName?: string;
        headerClassName?: string;
    }
}

interface DataTableProps<TData, TValue = unknown> {
    columns: ColumnDef<TData, TValue>[];
    columnVisibility?: VisibilityState;
    data: TData[];
    filterColumn?: string;
    filterPlaceholder?: string;
    initialPageSize?: number;
    onColumnVisibilityChange?: (visibility: VisibilityState) => void;
    onPageChange?: (pageIndex: number) => void;
    onRowClick?: (row: TData) => void;
    pageIndex?: number;
    renderSubComponent?: (props: { row: unknown }) => React.ReactElement;
    tableKey?: string;
}

const PAGE_SIZE_OPTIONS = [10, 15, 20, 50, 100] as const;

function DataTableInner<TData, TValue>(props: DataTableProps<TData, TValue>) {
    const {
        columns,
        columnVisibility: externalColumnVisibility,
        data,
        filterColumn = 'name',
        filterPlaceholder = 'Filter...',
        initialPageSize = 10,
        onColumnVisibilityChange,
        onPageChange,
        onRowClick,
        pageIndex,
        renderSubComponent,
        tableKey,
    } = props;

    // Load page size from localStorage
    const getStoredPageSize = React.useCallback((): number => {
        if (!tableKey) {
            return initialPageSize;
        }

        try {
            const stored = localStorage.getItem(`table-page-size-${tableKey}`);

            if (stored) {
                const parsed = Number.parseInt(stored, 10);

                return Number.isNaN(parsed) ? initialPageSize : parsed;
            }
        } catch {
            // Ignore localStorage errors
        }

        return initialPageSize;
    }, [tableKey, initialPageSize]);

    const [sorting, setSorting] = React.useState<SortingState>([]);
    const [columnFilters, setColumnFilters] = React.useState<ColumnFiltersState>([]);
    const [internalColumnVisibility, setInternalColumnVisibility] = React.useState<VisibilityState>({});
    const [rowSelection, setRowSelection] = React.useState({});
    const [expanded, setExpanded] = React.useState<ExpandedState>({});
    const [pagination, setPagination] = React.useState({
        pageIndex: pageIndex ?? 0,
        pageSize: getStoredPageSize(),
    });

    const columnVisibility = externalColumnVisibility ?? internalColumnVisibility;
    const handleColumnVisibilityChange = React.useCallback(
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

    // Sync external pageIndex with internal state
    React.useEffect(() => {
        if (pageIndex !== undefined && pageIndex !== pagination.pageIndex) {
            setPagination((prev) => ({ ...prev, pageIndex }));
        }
    }, [pageIndex, pagination.pageIndex]);

    // Save page size to localStorage when it changes
    const handlePageSizeChange = React.useCallback(
        (newPageSize: number) => {
            setPagination(() => ({ pageIndex: 0, pageSize: newPageSize }));

            if (tableKey) {
                try {
                    localStorage.setItem(`table-page-size-${tableKey}`, String(newPageSize));
                } catch {
                    // Ignore localStorage errors
                }
            }
        },
        [tableKey],
    );

    const table = useReactTable({
        autoResetPageIndex: false,
        columns,
        data,
        getCoreRowModel: getCoreRowModel(),
        getExpandedRowModel: getExpandedRowModel(),
        getFilteredRowModel: getFilteredRowModel(),
        getPaginationRowModel: getPaginationRowModel(),
        getSortedRowModel: getSortedRowModel(),
        onColumnFiltersChange: setColumnFilters,
        onColumnVisibilityChange: handleColumnVisibilityChange,
        onExpandedChange: setExpanded,
        onPaginationChange: (updater) => {
            const newPagination = typeof updater === 'function' ? updater(pagination) : updater;
            setPagination(newPagination);

            if (onPageChange && newPagination.pageIndex !== pagination.pageIndex) {
                onPageChange(newPagination.pageIndex);
            }
        },
        onRowSelectionChange: setRowSelection,
        onSortingChange: setSorting,
        state: {
            columnFilters,
            columnVisibility,
            expanded,
            pagination,
            rowSelection,
            sorting,
        },
    });

    return (
        <div className="w-full">
            <div className="flex items-center gap-4 py-4">
                {filterColumn && (
                    <Input
                        className="max-w-sm"
                        onChange={(event) => table.getColumn(filterColumn)?.setFilterValue(event.target.value)}
                        placeholder={filterPlaceholder}
                        value={(table.getColumn(filterColumn)?.getFilterValue() as string) ?? ''}
                    />
                )}
                <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                        <Button
                            className="ml-auto"
                            variant="outline"
                        >
                            Columns <ChevronDown className="ml-2 h-4 w-4" />
                        </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                        {table
                            .getAllColumns()
                            .filter((column) => column.getCanHide())
                            .map((column) => {
                                return (
                                    <DropdownMenuCheckboxItem
                                        checked={column.getIsVisible()}
                                        className="capitalize"
                                        key={column.id}
                                        onCheckedChange={(value) => column.toggleVisibility(!!value)}
                                        onSelect={(e) => e.preventDefault()}
                                    >
                                        {column.id}
                                    </DropdownMenuCheckboxItem>
                                );
                            })}
                    </DropdownMenuContent>
                </DropdownMenu>
            </div>
            <div className="rounded-md border">
                <Table>
                    <TableHeader>
                        {table.getHeaderGroups().map((headerGroup) => (
                            <TableRow key={headerGroup.id}>
                                {headerGroup.headers.map((header) => {
                                    return (
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
                                    );
                                })}
                            </TableRow>
                        ))}
                    </TableHeader>
                    <TableBody>
                        {table.getRowModel().rows?.length ? (
                            table.getRowModel().rows.map((row) => (
                                <React.Fragment key={row.id}>
                                    <TableRow
                                        className="group hover:bg-muted/50 cursor-pointer"
                                        data-state={row.getIsSelected() && 'selected'}
                                        onClick={() => {
                                            if (onRowClick) {
                                                onRowClick(row.original);
                                            } else {
                                                row?.toggleExpanded();
                                            }
                                        }}
                                    >
                                        {row.getVisibleCells().map((cell) => (
                                            <TableCell
                                                className={cell.column.columnDef.meta?.cellClassName}
                                                key={cell.id}
                                                onClick={(e) => {
                                                    // Prevent row click handler when clicking on action buttons
                                                    if (cell.column.id === 'actions') {
                                                        e.stopPropagation();
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
                                </React.Fragment>
                            ))
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
            <div className="flex items-center justify-between gap-2 py-4">
                <div className="text-muted-foreground flex-1 text-sm">
                    {!!table.getFilteredSelectedRowModel().rows.length && (
                        <>
                            {table.getFilteredSelectedRowModel().rows.length} of{' '}
                            {table.getFilteredRowModel().rows.length} row(s) selected.
                        </>
                    )}
                </div>
                <div className="flex items-center gap-2">
                    <div className="flex items-center gap-2">
                        <span className="text-muted-foreground text-sm">Rows per page:</span>
                        <Select
                            onValueChange={(value) => {
                                const pageSize = value === 'all' ? data.length : Number.parseInt(value, 10);
                                handlePageSizeChange(pageSize);
                            }}
                            value={
                                pagination.pageSize >= data.length && data.length > 0
                                    ? 'all'
                                    : String(pagination.pageSize)
                            }
                        >
                            <SelectTrigger className="h-8 w-[70px]">
                                <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
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
                    {(table.getCanPreviousPage() || table.getCanNextPage()) && (
                        <div className="flex gap-2">
                            <Button
                                disabled={!table.getCanPreviousPage()}
                                onClick={() => table.previousPage()}
                                size="sm"
                                variant="outline"
                            >
                                Previous
                            </Button>
                            <Button
                                disabled={!table.getCanNextPage()}
                                onClick={() => table.nextPage()}
                                size="sm"
                                variant="outline"
                            >
                                Next
                            </Button>
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
}

const DataTable = DataTableInner as <TData, TValue = never>(props: DataTableProps<TData, TValue>) => React.ReactElement;

export { DataTable };
