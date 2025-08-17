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
        headerClassName?: string;
        cellClassName?: string;
    }
}

interface DataTableProps<TData, TValue> {
    columns: ColumnDef<TData, TValue>[];
    data: TData[];
    renderSubComponent?: (props: { row: any }) => React.ReactElement;
    initialPageSize?: number;
    filterColumn?: string;
    filterPlaceholder?: string;
}

const DataTable = React.forwardRef<HTMLDivElement, DataTableProps<any, any>>(
    (
        {
            columns,
            data,
            renderSubComponent,
            initialPageSize = 10,
            filterColumn = 'name',
            filterPlaceholder = 'Filter...',
        },
        ref,
    ) => {
        const [sorting, setSorting] = React.useState<SortingState>([]);
        const [columnFilters, setColumnFilters] = React.useState<ColumnFiltersState>([]);
        const [columnVisibility, setColumnVisibility] = React.useState<VisibilityState>({});
        const [rowSelection, setRowSelection] = React.useState({});
        const [expanded, setExpanded] = React.useState<ExpandedState>({});

        const table = useReactTable({
            data,
            columns,
            onSortingChange: setSorting,
            onColumnFiltersChange: setColumnFilters,
            getCoreRowModel: getCoreRowModel(),
            getPaginationRowModel: getPaginationRowModel(),
            getSortedRowModel: getSortedRowModel(),
            getFilteredRowModel: getFilteredRowModel(),
            onColumnVisibilityChange: setColumnVisibility,
            onRowSelectionChange: setRowSelection,
            onExpandedChange: setExpanded,
            getExpandedRowModel: getExpandedRowModel(),
            initialState: {
                pagination: {
                    pageSize: initialPageSize,
                },
            },
            state: {
                sorting,
                columnFilters,
                columnVisibility,
                rowSelection,
                expanded,
            },
        });

        return (
            <div className="w-full">
                <div className="flex items-center gap-4 py-4">
                    {filterColumn && (
                        <Input
                            placeholder={filterPlaceholder}
                            value={(table.getColumn(filterColumn)?.getFilterValue() as string) ?? ''}
                            onChange={(event) => table.getColumn(filterColumn)?.setFilterValue(event.target.value)}
                            className="max-w-sm"
                        />
                    )}
                    <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                            <Button
                                variant="outline"
                                className="ml-auto"
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
                                            key={column.id}
                                            className="capitalize"
                                            checked={column.getIsVisible()}
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
                                                key={header.id}
                                                className={header.column.columnDef.meta?.headerClassName}
                                                style={
                                                    header.column.columnDef.size
                                                        ? {
                                                              width: header.column.columnDef.size,
                                                              minWidth: header.column.columnDef.size,
                                                              maxWidth: header.column.columnDef.size,
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
                                            data-state={row.getIsSelected() && 'selected'}
                                            className="cursor-pointer hover:bg-muted/50"
                                            onClick={() => row.toggleExpanded()}
                                        >
                                            {row.getVisibleCells().map((cell) => (
                                                <TableCell
                                                    key={cell.id}
                                                    className={cell.column.columnDef.meta?.cellClassName}
                                                    style={
                                                        cell.column.columnDef.size
                                                            ? {
                                                                  width: cell.column.columnDef.size,
                                                                  minWidth: cell.column.columnDef.size,
                                                                  maxWidth: cell.column.columnDef.size,
                                                              }
                                                            : undefined
                                                    }
                                                    onClick={(e) => {
                                                        // Prevent row expansion when clicking on action buttons
                                                        if (cell.column.id === 'actions') {
                                                            e.stopPropagation();
                                                        }
                                                    }}
                                                >
                                                    {flexRender(cell.column.columnDef.cell, cell.getContext())}
                                                </TableCell>
                                            ))}
                                        </TableRow>
                                        {row.getIsExpanded() && renderSubComponent && (
                                            <TableRow className="hover:bg-transparent border-0 cursor-default">
                                                <TableCell
                                                    colSpan={row.getVisibleCells().length}
                                                    className="p-0"
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
                                        colSpan={columns.length}
                                        className="h-24 text-center"
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
                        <div className="flex-1 text-sm text-muted-foreground">
                            {table.getFilteredSelectedRowModel().rows.length} of{' '}
                            {table.getFilteredRowModel().rows.length} row(s) selected.
                        </div>
                    )}
                    {(table.getCanPreviousPage() || table.getCanNextPage()) && (
                        <div className="space-x-2">
                            <Button
                                variant="outline"
                                size="sm"
                                onClick={() => table.previousPage()}
                                disabled={!table.getCanPreviousPage()}
                            >
                                Previous
                            </Button>
                            <Button
                                variant="outline"
                                size="sm"
                                onClick={() => table.nextPage()}
                                disabled={!table.getCanNextPage()}
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
