'use client';

import * as SheetPrimitive from '@radix-ui/react-dialog';
import { cva, type VariantProps } from 'class-variance-authority';
import { X } from 'lucide-react';
import * as React from 'react';

import { cn } from '@/lib/utils';

function Sheet({ ...props }: React.ComponentProps<typeof SheetPrimitive.Root>) {
    return (
        <SheetPrimitive.Root
            data-slot="sheet"
            {...props}
        />
    );
}

function SheetClose({ ...props }: React.ComponentProps<typeof SheetPrimitive.Close>) {
    return (
        <SheetPrimitive.Close
            data-slot="sheet-close"
            {...props}
        />
    );
}

function SheetOverlay({ className, ...props }: React.ComponentProps<typeof SheetPrimitive.Overlay>) {
    return (
        <SheetPrimitive.Overlay
            className={cn(
                'data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 fixed inset-0 z-50 bg-black/80',
                className,
            )}
            data-slot="sheet-overlay"
            {...props}
        />
    );
}

function SheetPortal({ ...props }: React.ComponentProps<typeof SheetPrimitive.Portal>) {
    return (
        <SheetPrimitive.Portal
            data-slot="sheet-portal"
            {...props}
        />
    );
}

function SheetTrigger({ ...props }: React.ComponentProps<typeof SheetPrimitive.Trigger>) {
    return (
        <SheetPrimitive.Trigger
            data-slot="sheet-trigger"
            {...props}
        />
    );
}

const sheetVariants = cva(
    'fixed z-50 gap-4 bg-background p-6 shadow-lg transition ease-in-out data-[state=closed]:duration-300 data-[state=open]:duration-500 data-[state=open]:animate-in data-[state=closed]:animate-out',
    {
        defaultVariants: {
            side: 'right',
        },
        variants: {
            side: {
                bottom: 'inset-x-0 bottom-0 border-t data-[state=closed]:slide-out-to-bottom data-[state=open]:slide-in-from-bottom',
                left: 'inset-y-0 left-0 h-full w-3/4 border-r data-[state=closed]:slide-out-to-left data-[state=open]:slide-in-from-left sm:max-w-sm',
                right: 'inset-y-0 right-0 h-full w-3/4 border-l data-[state=closed]:slide-out-to-right data-[state=open]:slide-in-from-right sm:max-w-sm',
                top: 'inset-x-0 top-0 border-b data-[state=closed]:slide-out-to-top data-[state=open]:slide-in-from-top',
            },
        },
    },
);

interface SheetContentProps
    extends React.ComponentProps<typeof SheetPrimitive.Content>, VariantProps<typeof sheetVariants> {
    container?: HTMLElement | null;
    overlay?: boolean;
}

function SheetContent({ children, className, container, overlay = true, side = 'right', ...props }: SheetContentProps) {
    return (
        <SheetPortal container={container ?? undefined}>
            {overlay && <SheetOverlay />}
            <SheetPrimitive.Content
                className={cn(sheetVariants({ side }), className)}
                data-slot="sheet-content"
                {...props}
            >
                <SheetPrimitive.Close className="ring-offset-background focus:ring-ring data-[state=open]:bg-secondary absolute top-4 right-4 rounded-sm opacity-70 transition-opacity hover:opacity-100 focus:ring-2 focus:ring-offset-2 focus:outline-hidden disabled:pointer-events-none">
                    <X className="h-4 w-4" />
                    <span className="sr-only">Close</span>
                </SheetPrimitive.Close>
                {children}
            </SheetPrimitive.Content>
        </SheetPortal>
    );
}

function SheetDescription({ className, ...props }: React.ComponentProps<typeof SheetPrimitive.Description>) {
    return (
        <SheetPrimitive.Description
            className={cn('text-muted-foreground text-sm', className)}
            data-slot="sheet-description"
            {...props}
        />
    );
}

function SheetFooter({ className, ...props }: React.ComponentProps<'div'>) {
    return (
        <div
            className={cn('flex flex-col-reverse gap-2 sm:flex-row sm:justify-end', className)}
            data-slot="sheet-footer"
            {...props}
        />
    );
}

function SheetHeader({ className, ...props }: React.ComponentProps<'div'>) {
    return (
        <div
            className={cn('flex flex-col gap-2 text-center sm:text-left', className)}
            data-slot="sheet-header"
            {...props}
        />
    );
}

function SheetTitle({ className, ...props }: React.ComponentProps<typeof SheetPrimitive.Title>) {
    return (
        <SheetPrimitive.Title
            className={cn('text-foreground text-lg font-semibold', className)}
            data-slot="sheet-title"
            {...props}
        />
    );
}

export {
    Sheet,
    SheetClose,
    SheetContent,
    SheetDescription,
    SheetFooter,
    SheetHeader,
    SheetOverlay,
    SheetPortal,
    SheetTitle,
    SheetTrigger,
};
