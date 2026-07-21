import { Slot } from '@radix-ui/react-slot';
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
                blue: 'border-blue-500/20 bg-blue-500/10 text-blue-800 hover:bg-blue-500/20 dark:text-blue-400',
                default: 'border-transparent bg-primary text-primary-foreground hover:bg-primary/80',
                destructive:
                    'border-transparent bg-destructive text-destructive-foreground hover:bg-destructive/90 dark:bg-destructive/80 dark:hover:bg-destructive/70',
                green: 'border-green-500/20 bg-green-500/10 text-green-800 hover:bg-green-500/20 dark:text-green-400',
                orange: 'border-orange-500/20 bg-orange-500/10 text-orange-800 hover:bg-orange-500/20 dark:text-orange-400',
                outline: 'text-foreground',
                pink: 'border-pink-500/20 bg-pink-500/10 text-pink-800 hover:bg-pink-500/20 dark:text-pink-400',
                purple: 'border-purple-500/20 bg-purple-500/10 text-purple-800 hover:bg-purple-500/20 dark:text-purple-400',
                red: 'border-red-500/20 bg-red-500/10 text-red-800 hover:bg-red-500/20 dark:text-red-400',
                secondary: 'border-transparent bg-secondary text-secondary-foreground hover:bg-secondary/80',
                yellow: 'border-yellow-500/20 bg-yellow-500/10 text-yellow-800 hover:bg-yellow-500/20 dark:text-yellow-400',
            },
        },
    },
);

export interface BadgeProps extends React.HTMLAttributes<HTMLDivElement>, VariantProps<typeof badgeVariants> {
    asChild?: boolean;
}

export type BadgeVariant = NonNullable<VariantProps<typeof badgeVariants>['variant']>;

function Badge({ asChild = false, className, variant, ...props }: BadgeProps) {
    const Comp = asChild ? Slot : 'div';

    return (
        <Comp
            className={cn(badgeVariants({ variant }), className)}
            data-slot="badge"
            {...props}
        />
    );
}

export { Badge, badgeVariants };
