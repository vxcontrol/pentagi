import * as React from 'react';

import { cn } from '@/lib/utils';

export type InputProps = React.ComponentProps<'input'>;

function Input({ className, type, ...props }: InputProps) {
    return (
        <input
            className={cn(
                'aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive border-input dark:bg-input/30 file:text-foreground placeholder:text-muted-foreground focus-visible:ring-ring flex h-9 w-full min-w-0 rounded-md border bg-transparent px-3 py-1 text-base shadow-xs transition-colors file:border-0 file:bg-transparent file:text-sm file:font-medium focus-visible:ring-1 focus-visible:outline-hidden disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-50 md:text-sm',
                type === 'number' &&
                    '[appearance:textfield] [&::-webkit-inner-spin-button]:m-0 [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:m-0 [&::-webkit-outer-spin-button]:appearance-none',
                className,
            )}
            data-slot="input"
            type={type}
            {...props}
        />
    );
}

export { Input };
