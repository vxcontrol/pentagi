import { cva, type VariantProps } from 'class-variance-authority';
import * as React from 'react';

import { Button, buttonVariants } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { TextareaAutosize } from '@/components/ui/textarea-autosize';
import { cn } from '@/lib/utils';

function InputGroup({ className, ...props }: React.ComponentProps<'div'>) {
    return (
        <div
            className={cn(
                'group/input-group shadow-2xs relative flex w-full items-center rounded-md border border-input outline-hidden transition-[color,box-shadow] dark:bg-input/30',
                'h-9 has-[>textarea]:h-auto',

                // Variants based on alignment.
                '[&>input]:has-[>[data-align=inline-start]]:pl-2',
                '[&>input]:has-[>[data-align=inline-end]]:pr-2',
                'has-[>[data-align=block-start]]:h-auto has-[>[data-align=block-start]]:flex-col [&>input]:has-[>[data-align=block-start]]:pb-3',
                'has-[>[data-align=block-end]]:h-auto has-[>[data-align=block-end]]:flex-col [&>input]:has-[>[data-align=block-end]]:pt-3',

                // Focus state.
                'has-[[data-slot=input-group-control]:focus-visible]:ring-1 has-[[data-slot=input-group-control]:focus-visible]:ring-ring',

                // Error state.
                'has-[[data-slot][aria-invalid=true]]:border-destructive has-[[data-slot][aria-invalid=true]]:ring-destructive/20 dark:has-[[data-slot][aria-invalid=true]]:ring-destructive/40',

                className,
            )}
            data-slot="input-group"
            role="group"
            {...props}
        />
    );
}

const inputGroupAddonVariants = cva(
    "text-muted-foreground flex h-auto cursor-text select-none items-center justify-center gap-2 py-1.5 text-sm font-medium group-data-[disabled=true]/input-group:opacity-50 [&>kbd]:rounded-[calc(var(--radius)-5px)] [&>svg:not([class*='size-'])]:size-4",
    {
        defaultVariants: {
            align: 'inline-start',
        },
        variants: {
            align: {
                'block-end':
                    '[.border-t]:pt-3 order-last w-full justify-start px-3 pb-3 group-has-[>input]/input-group:pb-2.5',
                'block-start':
                    '[.border-b]:pb-3 order-first w-full justify-start px-3 pt-3 group-has-[>input]/input-group:pt-2.5',
                'inline-end': 'order-last pr-3 has-[>button]:mr-[-0.4rem] has-[>kbd]:mr-[-0.35rem]',
                'inline-start': 'order-first pl-3 has-[>button]:ml-[-0.45rem] has-[>kbd]:ml-[-0.35rem]',
            },
        },
    },
);

function InputGroupAddon({
    align = 'inline-start',
    className,
    ...props
}: React.ComponentProps<'div'> & VariantProps<typeof inputGroupAddonVariants>) {
    return (
        <div
            className={cn(inputGroupAddonVariants({ align }), className)}
            data-align={align}
            data-slot="input-group-addon"
            onClick={(e) => {
                if ((e.target as HTMLElement).closest('button')) {
                    return;
                }

                e.currentTarget.parentElement?.querySelector('input')?.focus();
            }}
            role="group"
            {...props}
        />
    );
}

function InputGroupButton({
    className,
    size = 'xs',
    type = 'button',
    variant = 'ghost',
    ...props
}: Omit<React.ComponentProps<typeof Button>, 'size'> & VariantProps<typeof buttonVariants>) {
    return (
        <Button
            className={cn('shadow-none', className)}
            data-size={size}
            size={size}
            type={type}
            variant={variant}
            {...props}
        />
    );
}

function InputGroupInput({ className, ...props }: React.ComponentProps<'input'>) {
    return (
        <Input
            className={cn(
                'flex-1 rounded-none border-0 bg-transparent shadow-none focus-visible:ring-0 dark:bg-transparent',
                className,
            )}
            data-slot="input-group-control"
            {...props}
        />
    );
}

function InputGroupText({ className, ...props }: React.ComponentProps<'span'>) {
    return (
        <span
            className={cn(
                "flex items-center gap-2 text-sm text-muted-foreground [&_svg:not([class*='size-'])]:size-4 [&_svg]:pointer-events-none",
                className,
            )}
            {...props}
        />
    );
}

function InputGroupTextarea({ className, ...props }: React.ComponentProps<'textarea'>) {
    return (
        <Textarea
            className={cn(
                'flex-1 resize-none rounded-none border-0 bg-transparent py-3 shadow-none focus-visible:ring-0 dark:bg-transparent',
                className,
            )}
            data-slot="input-group-control"
            {...props}
        />
    );
}

function InputGroupTextareaAutosize({ className, ...props }: React.ComponentPropsWithoutRef<typeof TextareaAutosize>) {
    return (
        <TextareaAutosize
            className={cn(
                'flex-1 resize-none rounded-none border-0 bg-transparent py-3 shadow-none focus-visible:ring-0 dark:bg-transparent',
                className,
            )}
            data-slot="input-group-control"
            {...props}
        />
    );
}

export {
    InputGroup,
    InputGroupAddon,
    InputGroupButton,
    InputGroupInput,
    InputGroupText,
    InputGroupTextarea,
    InputGroupTextareaAutosize,
};
