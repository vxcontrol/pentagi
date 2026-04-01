import '@tanstack/react-table';

declare module '@tanstack/react-table' {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    interface ColumnMeta<TData, TValue> {
        cellClassName?: string;
        headerClassName?: string;
        preventRowClick?: boolean;
    }
}
