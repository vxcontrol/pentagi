import { Slot } from '@radix-ui/react-slot';
import { ChevronRight, MoreHorizontal } from 'lucide-react';
import * as React from 'react';

import { cn } from '@/lib/utils';

function Breadcrumb({
    ...props
}: React.ComponentProps<'nav'> & {
    separator?: React.ReactNode;
}) {
    return (
        <nav
            aria-label="breadcrumb"
            data-slot="breadcrumb"
            {...props}
        />
    );
}

function BreadcrumbEllipsis({ className, ...props }: React.ComponentProps<'span'>) {
    return (
        <span
            aria-hidden="true"
            className={cn('flex h-9 w-9 items-center justify-center', className)}
            data-slot="breadcrumb-ellipsis"
            role="presentation"
            {...props}
        >
            <MoreHorizontal className="h-4 w-4" />
            <span className="sr-only">More</span>
        </span>
    );
}

function BreadcrumbItem({ className, ...props }: React.ComponentProps<'li'>) {
    return (
        <li
            className={cn('inline-flex items-center gap-1.5', className)}
            data-slot="breadcrumb-item"
            {...props}
        />
    );
}

function BreadcrumbLink({
    asChild,
    className,
    ...props
}: React.ComponentProps<'a'> & {
    asChild?: boolean;
}) {
    const Comp = asChild ? Slot : 'a';

    return (
        <Comp
            className={cn('hover:text-foreground transition-colors', className)}
            data-slot="breadcrumb-link"
            {...props}
        />
    );
}

function BreadcrumbList({ className, ...props }: React.ComponentProps<'ol'>) {
    return (
        <ol
            className={cn(
                'text-muted-foreground flex flex-wrap items-center gap-1.5 text-sm wrap-break-word sm:gap-2.5',
                className,
            )}
            data-slot="breadcrumb-list"
            {...props}
        />
    );
}

function BreadcrumbPage({ className, ...props }: React.ComponentProps<'span'>) {
    return (
        <span
            aria-current="page"
            aria-disabled="true"
            className={cn('text-foreground font-normal', className)}
            data-slot="breadcrumb-page"
            role="link"
            {...props}
        />
    );
}

function BreadcrumbSeparator({ children, className, ...props }: React.ComponentProps<'li'>) {
    return (
        <li
            aria-hidden="true"
            className={cn('[&>svg]:h-3.5 [&>svg]:w-3.5', className)}
            data-slot="breadcrumb-separator"
            role="presentation"
            {...props}
        >
            {children ?? <ChevronRight />}
        </li>
    );
}

export {
    Breadcrumb,
    BreadcrumbEllipsis,
    BreadcrumbItem,
    BreadcrumbLink,
    BreadcrumbList,
    BreadcrumbPage,
    BreadcrumbSeparator,
};
