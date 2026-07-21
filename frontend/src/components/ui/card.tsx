import * as React from 'react';

import { cn } from '@/lib/utils';

function Card({ className, ...props }: React.ComponentProps<'div'>) {
    return (
        <div
            className={cn('bg-card text-card-foreground rounded-xl border shadow-sm', className)}
            data-slot="card"
            {...props}
        />
    );
}

function CardContent({ className, ...props }: React.ComponentProps<'div'>) {
    return (
        <div
            className={cn('p-4 pt-0', className)}
            data-slot="card-content"
            {...props}
        />
    );
}

function CardDescription({ className, ...props }: React.ComponentProps<'p'>) {
    return (
        <p
            className={cn('text-muted-foreground text-sm', className)}
            data-slot="card-description"
            {...props}
        />
    );
}

function CardFooter({ className, ...props }: React.ComponentProps<'div'>) {
    return (
        <div
            className={cn('flex items-center p-4 pt-0', className)}
            data-slot="card-footer"
            {...props}
        />
    );
}

function CardHeader({ className, ...props }: React.ComponentProps<'div'>) {
    return (
        <div
            className={cn('flex flex-col gap-1.5 p-4', className)}
            data-slot="card-header"
            {...props}
        />
    );
}

function CardTitle({ className, ...props }: React.ComponentProps<'h3'>) {
    return (
        <h3
            className={cn('leading-none font-semibold tracking-tight', className)}
            data-slot="card-title"
            {...props}
        />
    );
}

export { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle };
