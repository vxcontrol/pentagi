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
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';

// Extend ColumnMeta interface from @tanstack/react-table
declare module '@tanstack/react-table' {
    interface ColumnMeta<TData, TValue> {
        cellClassName?: string;
        headerClassName?: string;
    }
}

interface DataTableProps<TData, TValue> {
    columns: ColumnDef<TData, TValue>[];
    data: TData[];
    filterColumn?: string;
    filterPlaceholder?: string;
    initialPageSize?: number;
    onRowClick?: (row: TData) => void;
    renderSubComponent?: (props: { row: any }) => React.ReactElement;
}

const DataTable = React.forwardRef<HTMLDivElement, DataTableProps<any, any>>(
    (
        {
            columns,
            data,
            filterColumn = 'name',
            filterPlaceholder = 'Filter...',
            initialPageSize = 10,
            onRowClick,
            renderSubComponent,
        },
        ref,
    ) => {
        const [sorting, setSorting] = React.useState<SortingState>([]);
        const [columnFilters, setColumnFilters] = React.useState<ColumnFiltersState>([]);
        const [columnVisibility, setColumnVisibility] = React.useState<VisibilityState>({});
        const [rowSelection, setRowSelection] = React.useState({});
        const [expanded, setExpanded] = React.useState<ExpandedState>({});

        const table = useReactTable({
            columns,
            data,
            getCoreRowModel: getCoreRowModel(),
            getExpandedRowModel: getExpandedRowModel(),
            getFilteredRowModel: getFilteredRowModel(),
            getPaginationRowModel: getPaginationRowModel(),
            getSortedRowModel: getSortedRowModel(),
            initialState: {
                pagination: {
                    pageSize: initialPageSize,
                },
            },
            onColumnFiltersChange: setColumnFilters,
            onColumnVisibilityChange: setColumnVisibility,
            onExpandedChange: setExpanded,
            onRowSelectionChange: setRowSelection,
            onSortingChange: setSorting,
            state: {
                columnFilters,
                columnVisibility,
                expanded,
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
                <div className="flex items-center justify-end space-x-2 py-4">
                    {!!table.getFilteredSelectedRowModel().rows.length && (
                        <div className="text-muted-foreground flex-1 text-sm">
                            {table.getFilteredSelectedRowModel().rows.length} of{' '}
                            {table.getFilteredRowModel().rows.length} row(s) selected.
                        </div>
                    )}
                    {(table.getCanPreviousPage() || table.getCanNextPage()) && (
                        <div className="space-x-2">
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
        );
    },
);
DataTable.displayName = 'DataTable';

export { DataTable };
