import type { ReactNode } from 'react';

import { Breadcrumb, BreadcrumbItem, BreadcrumbList, BreadcrumbPage } from '@/components/ui/breadcrumb';
import { Button, type ButtonProps } from '@/components/ui/button';
import { Separator } from '@/components/ui/separator';
import { SidebarTrigger } from '@/components/ui/sidebar';
import { cn } from '@/lib/utils';

interface AppHeaderActionProps extends Omit<ButtonProps, 'children'> {
    endIcon?: ReactNode;
    icon: ReactNode;
    label: ReactNode;
}

export function AppHeader({ children, className }: { children: ReactNode; className?: string }) {
    return (
        <header
            className={cn(
                'bg-background sticky top-0 z-10 flex h-12 w-full shrink-0 items-center gap-2 border-b transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-12',
                className,
            )}
        >
            {children}
        </header>
    );
}

export function AppHeaderAction({
    'aria-label': ariaLabel,
    className,
    endIcon,
    icon,
    label,
    size = 'sm',
    ...props
}: AppHeaderActionProps) {
    const accessibleLabel = ariaLabel ?? (typeof label === 'string' ? label : undefined);

    return (
        <Button
            aria-label={accessibleLabel}
            className={cn('w-8 px-0 md:w-auto md:px-3', className)}
            size={size}
            {...props}
        >
            {icon}
            <span className="hidden md:inline">{label}</span>
            {endIcon ? <span className="hidden md:inline-flex">{endIcon}</span> : null}
        </Button>
    );
}

export function AppHeaderActions({ children, className }: { children: ReactNode; className?: string }) {
    return <div className={cn('flex shrink-0 items-center gap-2 px-4', className)}>{children}</div>;
}

export function AppHeaderContent({ children, className }: { children: ReactNode; className?: string }) {
    return (
        <div className={cn('flex min-w-0 flex-1 items-center gap-2 px-4', className)}>
            <SidebarTrigger className="-ml-1 shrink-0" />
            <Separator
                className="h-4 shrink-0"
                orientation="vertical"
            />
            {children}
        </div>
    );
}

export function AppHeaderTitle({
    children,
    className,
    icon,
}: {
    children: ReactNode;
    className?: string;
    icon?: ReactNode;
}) {
    return (
        <Breadcrumb className={cn('min-w-0 flex-1', className)}>
            <BreadcrumbList className="min-w-0 flex-nowrap">
                <BreadcrumbItem className="min-w-0 gap-2">
                    {icon}
                    <BreadcrumbPage className="min-w-0 truncate">{children}</BreadcrumbPage>
                </BreadcrumbItem>
            </BreadcrumbList>
        </Breadcrumb>
    );
}
