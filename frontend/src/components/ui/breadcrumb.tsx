import { ChevronRightIcon, DotsHorizontalIcon } from '@radix-ui/react-icons';
import { Slot } from '@radix-ui/react-slot';
import { CircleCheck, CircleDashed, CircleOff, CircleX, Loader2 } from 'lucide-react';
import * as React from 'react';

import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { cn } from '@/lib/utils';
import { getProviderIcon, getProviderTooltip, type Provider } from '@/models/Provider';

const Breadcrumb = React.forwardRef<
    HTMLElement,
    React.ComponentPropsWithoutRef<'nav'> & {
        separator?: React.ReactNode;
    }
>(({ ...props }, ref) => (
    <nav
        aria-label="breadcrumb"
        ref={ref}
        {...props}
    />
));
Breadcrumb.displayName = 'Breadcrumb';

const BreadcrumbList = React.forwardRef<HTMLOListElement, React.ComponentPropsWithoutRef<'ol'>>(
    ({ className, ...props }, ref) => (
        <ol
            className={cn(
                'flex flex-wrap items-center gap-1.5 break-words text-sm text-muted-foreground sm:gap-2.5',
                className,
            )}
            ref={ref}
            {...props}
        />
    ),
);
BreadcrumbList.displayName = 'BreadcrumbList';

const BreadcrumbItem = React.forwardRef<HTMLLIElement, React.ComponentPropsWithoutRef<'li'>>(
    ({ className, ...props }, ref) => (
        <li
            className={cn('inline-flex items-center gap-1.5', className)}
            ref={ref}
            {...props}
        />
    ),
);
BreadcrumbItem.displayName = 'BreadcrumbItem';

const BreadcrumbLink = React.forwardRef<
    HTMLAnchorElement,
    React.ComponentPropsWithoutRef<'a'> & {
        asChild?: boolean;
    }
>(({ asChild, className, ...props }, ref) => {
    const Comp = asChild ? Slot : 'a';

    return (
        <Comp
            className={cn('transition-colors hover:text-foreground', className)}
            ref={ref}
            {...props}
        />
    );
});
BreadcrumbLink.displayName = 'BreadcrumbLink';

const BreadcrumbStatus = React.forwardRef<
    HTMLSpanElement,
    React.ComponentPropsWithoutRef<'span'> & {
        status?: null | string;
    }
>(({ className, status, ...props }, ref) => {
    const renderStatusIcon = () => {
        if (!status) {
            return null;
        }

        switch (status) {
            case 'created':
                return (
                    <CircleDashed
                        aria-label="created"
                        className="h-4 w-4 text-blue-500"
                    />
                );
            case 'failed':
                return (
                    <CircleX
                        aria-label="failed"
                        className="h-4 w-4 text-red-500"
                    />
                );
            case 'finished':
                return (
                    <CircleCheck
                        aria-label="finished"
                        className="h-4 w-4 text-green-500"
                    />
                );
            case 'running':
                return (
                    <Loader2
                        aria-label="running"
                        className="h-4 w-4 animate-spin text-purple-500"
                    />
                );
            case 'waiting':
                return (
                    <CircleDashed
                        aria-label="waiting"
                        className="h-4 w-4 text-yellow-500"
                    />
                );
            default:
                return (
                    <CircleOff
                        aria-label="unknown status"
                        className="h-4 w-4 text-muted-foreground"
                    />
                );
        }
    };

    const iconElement = (
        <span
            className={cn('mr-2 inline-flex cursor-pointer items-center', className)}
            ref={ref}
            {...props}
        >
            {renderStatusIcon()}
        </span>
    );

    if (!status) {
        return null;
    }

    return (
        <Tooltip>
            <TooltipTrigger asChild>{iconElement}</TooltipTrigger>
            <TooltipContent>{status?.toLowerCase()}</TooltipContent>
        </Tooltip>
    );
});
BreadcrumbStatus.displayName = 'BreadcrumbStatus';

const BreadcrumbProvider = React.forwardRef<
    HTMLSpanElement,
    React.ComponentPropsWithoutRef<'span'> & {
        provider?: null | Provider;
    }
>(({ className, provider, ...props }, ref) => {
    const actualProvider = provider;

    const iconElement = (
        <span
            className={cn('mr-2 inline-flex cursor-pointer items-center', className)}
            ref={ref}
            {...props}
        >
            {actualProvider && getProviderIcon(actualProvider)}
        </span>
    );

    if (!actualProvider) {
        return null;
    }

    return (
        <Tooltip>
            <TooltipTrigger asChild>{iconElement}</TooltipTrigger>
            <TooltipContent>{getProviderTooltip(actualProvider)}</TooltipContent>
        </Tooltip>
    );
});
BreadcrumbProvider.displayName = 'BreadcrumbProvider';

const BreadcrumbPage = React.forwardRef<HTMLSpanElement, React.ComponentPropsWithoutRef<'span'>>(
    ({ className, ...props }, ref) => (
        <span
            aria-current="page"
            aria-disabled="true"
            className={cn('font-normal text-foreground', className)}
            ref={ref}
            role="link"
            {...props}
        />
    ),
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
    BreadcrumbProvider,
    BreadcrumbSeparator,
    BreadcrumbStatus,
};
