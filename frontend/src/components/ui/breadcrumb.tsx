import { ChevronRightIcon, DotsHorizontalIcon } from '@radix-ui/react-icons';
import { Slot } from '@radix-ui/react-slot';
import * as React from 'react';

import { cn } from '@/lib/utils';

const Breadcrumb = ({
    ref,
    ...props
}: React.ComponentPropsWithoutRef<'nav'> & {
    ref?: React.Ref<HTMLElement>;
    separator?: React.ReactNode;
}) => (
    <nav
        aria-label="breadcrumb"
        ref={ref}
        {...props}
    />
);
Breadcrumb.displayName = 'Breadcrumb';

const BreadcrumbList = ({
    className,
    ref,
    ...props
}: React.ComponentPropsWithoutRef<'ol'> & { ref?: React.Ref<HTMLOListElement> }) => (
    <ol
        className={cn(
            'text-muted-foreground flex flex-wrap items-center gap-1.5 text-sm wrap-break-word sm:gap-2.5',
            className,
        )}
        ref={ref}
        {...props}
    />
);
BreadcrumbList.displayName = 'BreadcrumbList';

const BreadcrumbItem = ({
    className,
    ref,
    ...props
}: React.ComponentPropsWithoutRef<'li'> & { ref?: React.Ref<HTMLLIElement> }) => (
    <li
        className={cn('inline-flex items-center gap-1.5', className)}
        ref={ref}
        {...props}
    />
);
BreadcrumbItem.displayName = 'BreadcrumbItem';

const BreadcrumbLink = ({
    asChild,
    className,
    ref,
    ...props
}: React.ComponentPropsWithoutRef<'a'> & {
    asChild?: boolean;
    ref?: React.Ref<HTMLAnchorElement>;
}) => {
    const Comp = asChild ? Slot : 'a';

    return (
        <Comp
            className={cn('hover:text-foreground transition-colors', className)}
            ref={ref}
            {...props}
        />
    );
};

BreadcrumbLink.displayName = 'BreadcrumbLink';

const BreadcrumbPage = ({
    className,
    ref,
    ...props
}: React.ComponentPropsWithoutRef<'span'> & { ref?: React.Ref<HTMLSpanElement> }) => (
    <span
        aria-current="page"
        aria-disabled="true"
        className={cn('text-foreground font-normal', className)}
        ref={ref}
        role="link"
        {...props}
    />
);
BreadcrumbPage.displayName = 'BreadcrumbPage';

const BreadcrumbSeparator = ({ children, className, ...props }: React.ComponentProps<'li'>) => (
    <li
        aria-hidden="true"
        className={cn('[&>svg]:h-3.5 [&>svg]:w-3.5', className)}
        role="presentation"
        {...props}
    >
        {children ?? <ChevronRightIcon />}
    </li>
);
BreadcrumbSeparator.displayName = 'BreadcrumbSeparator';

const BreadcrumbEllipsis = ({ className, ...props }: React.ComponentProps<'span'>) => (
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
BreadcrumbEllipsis.displayName = 'BreadcrumbElipssis';

export {
    Breadcrumb,
    BreadcrumbEllipsis,
    BreadcrumbItem,
    BreadcrumbLink,
    BreadcrumbList,
    BreadcrumbPage,
    BreadcrumbSeparator,
};
