import { ChevronRightIcon, DotsHorizontalIcon } from '@radix-ui/react-icons';
import { Slot } from '@radix-ui/react-slot';
import * as React from 'react';
import {
    Brain,
    CircleCheck,
    CircleDashed,
    CircleOff,
    CircleX,
    Loader2
} from 'lucide-react';

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
        ref={ref}
        aria-label="breadcrumb"
        {...props}
    />
));
Breadcrumb.displayName = 'Breadcrumb';

const BreadcrumbList = React.forwardRef<HTMLOListElement, React.ComponentPropsWithoutRef<'ol'>>(
    ({ className, ...props }, ref) => (
        <ol
            ref={ref}
            className={cn(
                'flex flex-wrap items-center gap-1.5 break-words text-sm text-muted-foreground sm:gap-2.5',
                className,
            )}
            {...props}
        />
    ),
);
BreadcrumbList.displayName = 'BreadcrumbList';

const BreadcrumbItem = React.forwardRef<HTMLLIElement, React.ComponentPropsWithoutRef<'li'>>(
    ({ className, ...props }, ref) => (
        <li
            ref={ref}
            className={cn('inline-flex items-center gap-1.5', className)}
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
            ref={ref}
            className={cn('transition-colors hover:text-foreground', className)}
            {...props}
        />
    );
});
BreadcrumbLink.displayName = 'BreadcrumbLink';

const BreadcrumbStatus = React.forwardRef<
    HTMLSpanElement,
    React.ComponentPropsWithoutRef<'span'> & {
        status?: string | null;
    }
>(({ status, className, ...props }, ref) => {
    const renderStatusIcon = () => {
        if (!status) return null;
        
        switch (status) {
            case 'failed':
                return <CircleX className="h-4 w-4 text-red-500" aria-label="failed" />;
            case 'finished':
                return <CircleCheck className="h-4 w-4 text-green-500" aria-label="finished" />;
            case 'running':
                return <Loader2 className="h-4 w-4 text-purple-500 animate-spin" aria-label="running" />;
            case 'created':
                return <CircleDashed className="h-4 w-4 text-blue-500" aria-label="created" />;
            case 'waiting':
                return <CircleDashed className="h-4 w-4 text-yellow-500" aria-label="waiting" />;
            default:
                return <CircleOff className="h-4 w-4 text-muted-foreground" aria-label="unknown status" />;
        }
    };

    const iconElement = (
        <span
            ref={ref}
            className={cn('inline-flex items-center mr-2 cursor-pointer', className)}
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
        provider?: Provider | null;
    }
>(({ provider, className, ...props }, ref) => {
    const actualProvider = provider;

    const iconElement = (
        <span
            ref={ref}
            className={cn('inline-flex items-center mr-2 cursor-pointer', className)}
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
            ref={ref}
            role="link"
            aria-disabled="true"
            aria-current="page"
            className={cn('font-normal text-foreground', className)}
            {...props}
        />
    ),
);
BreadcrumbPage.displayName = 'BreadcrumbPage';

const BreadcrumbSeparator = ({ children, className, ...props }: React.ComponentProps<'li'>) => (
    <li
        role="presentation"
        aria-hidden="true"
        className={cn('[&>svg]:h-3.5 [&>svg]:w-3.5', className)}
        {...props}
    >
        {children ?? <ChevronRightIcon />}
    </li>
);
BreadcrumbSeparator.displayName = 'BreadcrumbSeparator';

const BreadcrumbEllipsis = ({ className, ...props }: React.ComponentProps<'span'>) => (
    <span
        role="presentation"
        aria-hidden="true"
        className={cn('flex h-9 w-9 items-center justify-center', className)}
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
