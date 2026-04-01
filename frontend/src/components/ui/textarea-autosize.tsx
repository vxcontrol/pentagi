import * as React from 'react';
import ReactTextareaAutosize from 'react-textarea-autosize';

import { cn } from '@/lib/utils';

function TextareaAutosize({ className, ...props }: React.ComponentProps<typeof ReactTextareaAutosize>) {
    return (
        <ReactTextareaAutosize
            className={cn(
                'aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive border-input placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-ring/50 dark:bg-input/30 flex min-h-16 w-full resize-none rounded-md border bg-transparent px-3 py-2 text-base shadow-2xs outline-hidden transition-[color,box-shadow] focus-visible:ring-[3px] disabled:cursor-not-allowed disabled:opacity-50 md:text-sm',
                className,
            )}
            data-slot="textarea-autosize"
            {...props}
        />
    );
}

export { TextareaAutosize };
