import { cva, type VariantProps } from 'class-variance-authority';
import * as React from 'react';

import { cn } from '@/lib/utils';

const badgeVariants = cva(
    'inline-flex items-center rounded-full border px-2 py-0.5 gap-1 text-xs font-semibold transition-colors focus:outline-hidden focus:ring-2 focus:ring-ring focus:ring-offset-2',
    {
        defaultVariants: {
            variant: 'default',
        },
        variants: {
            variant: {
                blue: 'border-blue-500/20 bg-blue-500/10 text-blue-600',
                default: 'border-transparent bg-primary text-primary-foreground hover:bg-primary/80',
                destructive: 'border-transparent bg-destructive text-destructive-foreground hover:bg-destructive/80',
                green: 'border-green-500/20 bg-green-500/10 text-green-600',
                orange: 'border-orange-500/20 bg-orange-500/10 text-orange-600',
                outline: 'text-foreground',
                pink: 'border-pink-500/20 bg-pink-500/10 text-pink-600',
                purple: 'border-purple-500/20 bg-purple-500/10 text-purple-600',
                red: 'border-red-500/20 bg-red-500/10 text-red-600',
                secondary: 'border-transparent bg-secondary text-secondary-foreground hover:bg-secondary/80',
                yellow: 'border-yellow-500/20 bg-yellow-500/10 text-yellow-600',
            },
        },
    },
);

export interface BadgeProps extends React.HTMLAttributes<HTMLDivElement>, VariantProps<typeof badgeVariants> {}

export type BadgeVariant = NonNullable<VariantProps<typeof badgeVariants>['variant']>;

function Badge({ className, variant, ...props }: BadgeProps) {
    return (
        <div
            className={cn(badgeVariants({ variant }), className)}
            {...props}
        />
    );
}

export { Badge, badgeVariants };
