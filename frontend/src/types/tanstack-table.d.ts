import '@tanstack/react-table';

declare module '@tanstack/react-table' {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    interface ColumnMeta<TData, TValue> {
        cellClassName?: string;
        /** Label for the Columns visibility menu when `column.id` is not human-readable. */
        columnMenuLabel?: string;
        headerClassName?: string;
        preventRowClick?: boolean;
    }
}
