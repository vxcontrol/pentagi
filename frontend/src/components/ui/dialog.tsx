import * as DialogPrimitive from '@radix-ui/react-dialog';
import { X } from 'lucide-react';
import * as React from 'react';

import { cn } from '@/lib/utils';

function Dialog({ ...props }: React.ComponentProps<typeof DialogPrimitive.Root>) {
    return (
        <DialogPrimitive.Root
            data-slot="dialog"
            {...props}
        />
    );
}

function DialogClose({ ...props }: React.ComponentProps<typeof DialogPrimitive.Close>) {
    return (
        <DialogPrimitive.Close
            data-slot="dialog-close"
            {...props}
        />
    );
}

function DialogContent({ children, className, ...props }: React.ComponentProps<typeof DialogPrimitive.Content>) {
    return (
        <DialogPortal>
            <DialogOverlay />
            <DialogPrimitive.Content
                className={cn(
                    'bg-background data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 fixed top-[50%] left-[50%] z-50 grid w-full max-w-[calc(100%-2rem)] translate-x-[-50%] translate-y-[-50%] gap-4 rounded-lg border p-6 shadow-lg outline-0 duration-200 sm:max-w-lg',
                    className,
                )}
                data-slot="dialog-content"
                {...props}
            >
                {children}
                <DialogPrimitive.Close
                    className="ring-offset-background focus:ring-ring data-[state=open]:bg-accent data-[state=open]:text-muted-foreground absolute top-4 right-4 rounded-sm opacity-70 transition-opacity hover:opacity-100 focus:ring-2 focus:ring-offset-2 focus:outline-hidden disabled:pointer-events-none"
                    data-slot="dialog-close"
                >
                    <X className="h-4 w-4" />
                    <span className="sr-only">Close</span>
                </DialogPrimitive.Close>
            </DialogPrimitive.Content>
        </DialogPortal>
    );
}

function DialogDescription({ className, ...props }: React.ComponentProps<typeof DialogPrimitive.Description>) {
    return (
        <DialogPrimitive.Description
            className={cn('text-muted-foreground text-sm', className)}
            data-slot="dialog-description"
            {...props}
        />
    );
}

function DialogFooter({ className, ...props }: React.ComponentProps<'div'>) {
    return (
        <div
            className={cn('flex flex-col-reverse gap-2 sm:flex-row sm:justify-end', className)}
            data-slot="dialog-footer"
            {...props}
        />
    );
}

function DialogHeader({ className, ...props }: React.ComponentProps<'div'>) {
    return (
        <div
            className={cn('flex flex-col gap-1.5 text-center sm:text-left', className)}
            data-slot="dialog-header"
            {...props}
        />
    );
}

function DialogOverlay({ className, ...props }: React.ComponentProps<typeof DialogPrimitive.Overlay>) {
    return (
        <DialogPrimitive.Overlay
            className={cn(
                'data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 bg-background/80 fixed inset-0 z-50 backdrop-blur-xs',
                className,
            )}
            data-slot="dialog-overlay"
            {...props}
        />
    );
}

function DialogPortal({ ...props }: React.ComponentProps<typeof DialogPrimitive.Portal>) {
    return (
        <DialogPrimitive.Portal
            data-slot="dialog-portal"
            {...props}
        />
    );
}

function DialogTitle({ className, ...props }: React.ComponentProps<typeof DialogPrimitive.Title>) {
    return (
        <DialogPrimitive.Title
            className={cn('text-lg leading-none font-semibold tracking-tight', className)}
            data-slot="dialog-title"
            {...props}
        />
    );
}

function DialogTrigger({ ...props }: React.ComponentProps<typeof DialogPrimitive.Trigger>) {
    return (
        <DialogPrimitive.Trigger
            data-slot="dialog-trigger"
            {...props}
        />
    );
}

export {
    Dialog,
    DialogClose,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
};
