import { ChevronRightIcon, DotsHorizontalIcon } from '@radix-ui/react-icons';
import { Slot } from '@radix-ui/react-slot';
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
            {...props}
        />
    );
}

function BreadcrumbEllipsis({ className, ...props }: React.ComponentProps<'span'>) {
    return (
        <span
            aria-hidden="true"
            className={cn('flex h-9 w-9 items-center justify-center', className)}
            role="presentation"
            {...props}
        >
            <DotsHorizontalIcon className="h-4 w-4" />
            <span className="sr-only">More</span>
        </span>
    );
}

function BreadcrumbItem({ className, ...props }: React.ComponentProps<'li'>) {
    return (
        <li
            className={cn('inline-flex items-center gap-1.5', className)}
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
            role="presentation"
            {...props}
        >
            {children ?? <ChevronRightIcon />}
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
