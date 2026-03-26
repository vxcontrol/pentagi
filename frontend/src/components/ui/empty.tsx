import { cva, type VariantProps } from 'class-variance-authority';

import { cn } from '@/lib/utils';

function Empty({ className, ...props }: React.ComponentProps<'div'>) {
    return (
        <div
            className={cn(
                'flex min-w-0 flex-1 flex-col items-center justify-center gap-4 text-balance rounded-lg border-dashed p-6 text-center md:p-12',
                className,
            )}
            data-slot="empty"
            {...props}
        />
    );
}

function EmptyHeader({ className, ...props }: React.ComponentProps<'div'>) {
    return (
        <div
            className={cn('flex max-w-sm flex-col items-center gap-1 text-center', className)}
            data-slot="empty-header"
            {...props}
        />
    );
}

const emptyMediaVariants = cva(
    'mb-2 flex shrink-0 items-center justify-center [&_svg]:pointer-events-none [&_svg]:shrink-0',
    {
        defaultVariants: {
            variant: 'default',
        },
        variants: {
            variant: {
                default: 'bg-transparent',
                icon: "bg-muted text-foreground flex size-10 shrink-0 items-center justify-center rounded-lg [&_svg:not([class*='size-'])]:size-6",
            },
        },
    },
);

function EmptyContent({ className, ...props }: React.ComponentProps<'div'>) {
    return (
        <div
            className={cn('flex w-full min-w-0 max-w-sm flex-col items-center gap-2 text-balance text-sm', className)}
            data-slot="empty-content"
            {...props}
        />
    );
}

function EmptyDescription({ className, ...props }: React.ComponentProps<'p'>) {
    return (
        <div
            className={cn(
                'text-sm/relaxed text-muted-foreground [&>a:hover]:text-primary [&>a]:underline [&>a]:underline-offset-4',
                className,
            )}
            data-slot="empty-description"
            {...props}
        />
    );
}

function EmptyMedia({
    className,
    variant = 'default',
    ...props
}: React.ComponentProps<'div'> & VariantProps<typeof emptyMediaVariants>) {
    return (
        <div
            className={cn(emptyMediaVariants({ className, variant }))}
            data-slot="empty-icon"
            data-variant={variant}
            {...props}
        />
    );
}

function EmptyTitle({ className, ...props }: React.ComponentProps<'div'>) {
    return (
        <div
            className={cn('text-lg font-medium tracking-tight', className)}
            data-slot="empty-title"
            {...props}
        />
    );
}

export { Empty, EmptyContent, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle };
