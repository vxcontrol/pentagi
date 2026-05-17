import * as React from 'react';

import { cn } from '@/lib/utils';

function Card({ className, ...props }: React.ComponentProps<'div'>) {
    return (
        <div
            className={cn('bg-card text-card-foreground rounded-xl border shadow-sm', className)}
            {...props}
        />
    );
}

function CardContent({ className, ...props }: React.ComponentProps<'div'>) {
    return (
        <div
            className={cn('p-4 pt-0', className)}
            {...props}
        />
    );
}

function CardDescription({ className, ...props }: React.ComponentProps<'p'>) {
    return (
        <p
            className={cn('text-muted-foreground text-sm', className)}
            {...props}
        />
    );
}

function CardFooter({ className, ...props }: React.ComponentProps<'div'>) {
    return (
        <div
            className={cn('flex items-center p-4 pt-0', className)}
            {...props}
        />
    );
}

function CardHeader({ className, ...props }: React.ComponentProps<'div'>) {
    return (
        <div
            className={cn('flex flex-col gap-1.5 p-4', className)}
            {...props}
        />
    );
}

function CardTitle({ className, ...props }: React.ComponentProps<'h3'>) {
    return (
        <h3
            className={cn('leading-none font-semibold tracking-tight', className)}
            {...props}
        />
    );
}

export { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle };
