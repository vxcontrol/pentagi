import * as React from 'react';

import { cn } from '@/lib/utils';

const Table = React.forwardRef<HTMLTableElement, React.HTMLAttributes<HTMLTableElement>>(
    ({ className, ...props }, ref) => (
        <div className="relative w-full overflow-auto">
            <table
                className={cn('w-full caption-bottom text-sm', className)}
                ref={ref}
                {...props}
            />
        </div>
    ),
);
Table.displayName = 'Table';

const TableHeader = React.forwardRef<HTMLTableSectionElement, React.HTMLAttributes<HTMLTableSectionElement>>(
    ({ className, ...props }, ref) => (
        <thead
            className={cn('[&_tr]:border-b', className)}
            ref={ref}
            {...props}
        />
    ),
);
TableHeader.displayName = 'TableHeader';

const TableBody = React.forwardRef<HTMLTableSectionElement, React.HTMLAttributes<HTMLTableSectionElement>>(
    ({ className, ...props }, ref) => (
        <tbody
            className={cn('[&_tr:last-child]:border-0', className)}
            ref={ref}
            {...props}
        />
    ),
);
TableBody.displayName = 'TableBody';

const TableFooter = React.forwardRef<HTMLTableSectionElement, React.HTMLAttributes<HTMLTableSectionElement>>(
    ({ className, ...props }, ref) => (
        <tfoot
            className={cn('bg-muted/50 border-t font-medium last:[&>tr]:border-b-0', className)}
            ref={ref}
            {...props}
        />
    ),
);
TableFooter.displayName = 'TableFooter';

const TableRow = React.forwardRef<HTMLTableRowElement, React.HTMLAttributes<HTMLTableRowElement>>(
    ({ className, ...props }, ref) => (
        <tr
            className={cn('hover:bg-muted/50 data-[state=selected]:bg-muted border-b transition-colors', className)}
            ref={ref}
            {...props}
        />
    ),
);
TableRow.displayName = 'TableRow';

const TableHead = React.forwardRef<HTMLTableCellElement, React.ThHTMLAttributes<HTMLTableCellElement>>(
    ({ className, ...props }, ref) => (
        <th
            className={cn(
                'text-muted-foreground h-12 px-4 text-left align-middle font-medium [&:has([role=checkbox])]:pr-0',
                className,
            )}
            ref={ref}
            {...props}
        />
    ),
);
TableHead.displayName = 'TableHead';

const TableCell = React.forwardRef<HTMLTableCellElement, React.TdHTMLAttributes<HTMLTableCellElement>>(
    ({ className, ...props }, ref) => (
        <td
            className={cn('p-4 align-middle [&:has([role=checkbox])]:pr-0', className)}
            ref={ref}
            {...props}
        />
    ),
);
TableCell.displayName = 'TableCell';

const TableCaption = React.forwardRef<HTMLTableCaptionElement, React.HTMLAttributes<HTMLTableCaptionElement>>(
    ({ className, ...props }, ref) => (
        <caption
            className={cn('text-muted-foreground mt-4 text-sm', className)}
            ref={ref}
            {...props}
        />
    ),
);
TableCaption.displayName = 'TableCaption';

export { Table, TableBody, TableCaption, TableCell, TableFooter, TableHead, TableHeader, TableRow };
