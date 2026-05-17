import * as React from 'react';

import { cn } from '@/lib/utils';

const Card = ({
    className,
    ref,
    ...props
}: React.HTMLAttributes<HTMLDivElement> & { ref?: React.Ref<HTMLDivElement> }) => (
    <div
        className={cn('bg-card text-card-foreground rounded-xl border shadow-sm', className)}
        ref={ref}
        {...props}
    />
);
Card.displayName = 'Card';

const CardHeader = ({
    className,
    ref,
    ...props
}: React.HTMLAttributes<HTMLDivElement> & { ref?: React.Ref<HTMLDivElement> }) => (
    <div
        className={cn('flex flex-col gap-1.5 p-4', className)}
        ref={ref}
        {...props}
    />
);
CardHeader.displayName = 'CardHeader';

const CardTitle = ({
    className,
    ref,
    ...props
}: React.HTMLAttributes<HTMLHeadingElement> & { ref?: React.Ref<HTMLParagraphElement> }) => (
    <h3
        className={cn('leading-none font-semibold tracking-tight', className)}
        ref={ref}
        {...props}
    />
);
CardTitle.displayName = 'CardTitle';

const CardDescription = ({
    className,
    ref,
    ...props
}: React.HTMLAttributes<HTMLParagraphElement> & { ref?: React.Ref<HTMLParagraphElement> }) => (
    <p
        className={cn('text-muted-foreground text-sm', className)}
        ref={ref}
        {...props}
    />
);
CardDescription.displayName = 'CardDescription';

const CardContent = ({
    className,
    ref,
    ...props
}: React.HTMLAttributes<HTMLDivElement> & { ref?: React.Ref<HTMLDivElement> }) => (
    <div
        className={cn('p-4 pt-0', className)}
        ref={ref}
        {...props}
    />
);
CardContent.displayName = 'CardContent';

const CardFooter = ({
    className,
    ref,
    ...props
}: React.HTMLAttributes<HTMLDivElement> & { ref?: React.Ref<HTMLDivElement> }) => (
    <div
        className={cn('flex items-center p-4 pt-0', className)}
        ref={ref}
        {...props}
    />
);
CardFooter.displayName = 'CardFooter';

export { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle };
