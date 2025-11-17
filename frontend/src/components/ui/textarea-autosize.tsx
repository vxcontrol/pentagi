import * as React from 'react';
import ReactTextareaAutosize from 'react-textarea-autosize';

import { cn } from '@/lib/utils';

function TextareaAutosize({ className, ...props }: React.ComponentProps<typeof ReactTextareaAutosize>) {
    return (
        <ReactTextareaAutosize
            className={cn(
                'aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive shadow-xs flex min-h-16 w-full resize-none rounded-md border border-input bg-transparent px-3 py-2 text-base outline-none transition-[color,box-shadow] placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50 disabled:cursor-not-allowed disabled:opacity-50 dark:bg-input/30 md:text-sm',
                className,
            )}
            data-slot="textarea-autosize"
            {...props}
        />
    );
}

export { TextareaAutosize };
