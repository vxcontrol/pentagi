import * as React from 'react';

import { cn } from '@/lib/utils';

const Table = ({
    className,
    ref,
    ...props
}: React.HTMLAttributes<HTMLTableElement> & { ref?: React.Ref<HTMLTableElement> }) => (
    <div className="relative w-full overflow-auto">
        <table
            className={cn('w-full caption-bottom text-sm', className)}
            ref={ref}
            {...props}
        />
    </div>
);
Table.displayName = 'Table';

const TableHeader = ({
    className,
    ref,
    ...props
}: React.HTMLAttributes<HTMLTableSectionElement> & { ref?: React.Ref<HTMLTableSectionElement> }) => (
    <thead
        className={cn('[&_tr]:border-b', className)}
        ref={ref}
        {...props}
    />
);
TableHeader.displayName = 'TableHeader';

const TableBody = ({
    className,
    ref,
    ...props
}: React.HTMLAttributes<HTMLTableSectionElement> & { ref?: React.Ref<HTMLTableSectionElement> }) => (
    <tbody
        className={cn('[&_tr:last-child]:border-0', className)}
        ref={ref}
        {...props}
    />
);
TableBody.displayName = 'TableBody';

const TableFooter = ({
    className,
    ref,
    ...props
}: React.HTMLAttributes<HTMLTableSectionElement> & { ref?: React.Ref<HTMLTableSectionElement> }) => (
    <tfoot
        className={cn('bg-muted/50 border-t font-medium last:[&>tr]:border-b-0', className)}
        ref={ref}
        {...props}
    />
);
TableFooter.displayName = 'TableFooter';

const TableRow = ({
    className,
    ref,
    ...props
}: React.HTMLAttributes<HTMLTableRowElement> & { ref?: React.Ref<HTMLTableRowElement> }) => (
    <tr
        className={cn('hover:bg-muted/50 data-[state=selected]:bg-muted border-b transition-colors', className)}
        ref={ref}
        {...props}
    />
);
TableRow.displayName = 'TableRow';

const TableHead = ({
    className,
    ref,
    ...props
}: React.ThHTMLAttributes<HTMLTableCellElement> & { ref?: React.Ref<HTMLTableCellElement> }) => (
    <th
        className={cn(
            'text-muted-foreground h-12 px-4 text-left align-middle font-medium [&:has([role=checkbox])]:pr-0',
            className,
        )}
        ref={ref}
        {...props}
    />
);
TableHead.displayName = 'TableHead';

const TableCell = ({
    className,
    ref,
    ...props
}: React.TdHTMLAttributes<HTMLTableCellElement> & { ref?: React.Ref<HTMLTableCellElement> }) => (
    <td
        className={cn('p-4 align-middle [&:has([role=checkbox])]:pr-0', className)}
        ref={ref}
        {...props}
    />
);
TableCell.displayName = 'TableCell';

const TableCaption = ({
    className,
    ref,
    ...props
}: React.HTMLAttributes<HTMLTableCaptionElement> & { ref?: React.Ref<HTMLTableCaptionElement> }) => (
    <caption
        className={cn('text-muted-foreground mt-4 text-sm', className)}
        ref={ref}
        {...props}
    />
);
TableCaption.displayName = 'TableCaption';

export { Table, TableBody, TableCaption, TableCell, TableFooter, TableHead, TableHeader, TableRow };
