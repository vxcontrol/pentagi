import * as DialogPrimitive from '@radix-ui/react-dialog';
import { X } from 'lucide-react';
import * as React from 'react';

import { cn } from '@/lib/utils';

const Dialog = DialogPrimitive.Root;

const DialogTrigger = DialogPrimitive.Trigger;

const DialogPortal = DialogPrimitive.Portal;

const DialogClose = DialogPrimitive.Close;

const DialogOverlay = ({
    className,
    ref,
    ...props
}: React.ComponentPropsWithoutRef<typeof DialogPrimitive.Overlay> & {
    ref?: React.Ref<React.ElementRef<typeof DialogPrimitive.Overlay>>;
}) => (
    <DialogPrimitive.Overlay
        className={cn(
            'data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 bg-background/80 fixed inset-0 z-50 backdrop-blur-xs',
            className,
        )}
        ref={ref}
        {...props}
    />
);
DialogOverlay.displayName = DialogPrimitive.Overlay.displayName;

const DialogContent = ({
    children,
    className,
    ref,
    ...props
}: React.ComponentPropsWithoutRef<typeof DialogPrimitive.Content> & {
    ref?: React.Ref<React.ElementRef<typeof DialogPrimitive.Content>>;
}) => (
    <DialogPortal>
        <DialogOverlay />
        <DialogPrimitive.Content
            className={cn(
                'bg-background data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 fixed top-[50%] left-[50%] z-50 grid w-full max-w-[calc(100%-2rem)] translate-x-[-50%] translate-y-[-50%] gap-4 rounded-lg border p-6 shadow-lg outline-0 duration-200 sm:max-w-lg',
                className,
            )}
            ref={ref}
            {...props}
        >
            {children}
            <DialogPrimitive.Close className="ring-offset-background focus:ring-ring data-[state=open]:bg-accent data-[state=open]:text-muted-foreground absolute top-4 right-4 rounded-sm opacity-70 transition-opacity hover:opacity-100 focus:ring-2 focus:ring-offset-2 focus:outline-hidden disabled:pointer-events-none">
                <X className="h-4 w-4" />
                <span className="sr-only">Close</span>
            </DialogPrimitive.Close>
        </DialogPrimitive.Content>
    </DialogPortal>
);
DialogContent.displayName = DialogPrimitive.Content.displayName;

const DialogHeader = ({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) => (
    <div
        className={cn('flex flex-col gap-1.5 text-center sm:text-left', className)}
        {...props}
    />
);
DialogHeader.displayName = 'DialogHeader';

const DialogFooter = ({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) => (
    <div
        className={cn('flex flex-col-reverse gap-2 sm:flex-row sm:justify-end', className)}
        {...props}
    />
);
DialogFooter.displayName = 'DialogFooter';

const DialogTitle = ({
    className,
    ref,
    ...props
}: React.ComponentPropsWithoutRef<typeof DialogPrimitive.Title> & {
    ref?: React.Ref<React.ElementRef<typeof DialogPrimitive.Title>>;
}) => (
    <DialogPrimitive.Title
        className={cn('text-lg leading-none font-semibold tracking-tight', className)}
        ref={ref}
        {...props}
    />
);
DialogTitle.displayName = DialogPrimitive.Title.displayName;

const DialogDescription = ({
    className,
    ref,
    ...props
}: React.ComponentPropsWithoutRef<typeof DialogPrimitive.Description> & {
    ref?: React.Ref<React.ElementRef<typeof DialogPrimitive.Description>>;
}) => (
    <DialogPrimitive.Description
        className={cn('text-muted-foreground text-sm', className)}
        ref={ref}
        {...props}
    />
);
DialogDescription.displayName = DialogPrimitive.Description.displayName;

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
