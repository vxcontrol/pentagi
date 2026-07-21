import * as ContextMenuPrimitive from '@radix-ui/react-context-menu';
import { Check, ChevronRight, Circle } from 'lucide-react';
import * as React from 'react';

import { cn } from '@/lib/utils';

function ContextMenu({ ...props }: React.ComponentProps<typeof ContextMenuPrimitive.Root>) {
    return (
        <ContextMenuPrimitive.Root
            data-slot="context-menu"
            {...props}
        />
    );
}

function ContextMenuCheckboxItem({
    checked,
    children,
    className,
    ...props
}: React.ComponentProps<typeof ContextMenuPrimitive.CheckboxItem>) {
    return (
        <ContextMenuPrimitive.CheckboxItem
            checked={checked}
            className={cn(
                'focus:bg-accent focus:text-accent-foreground relative flex cursor-default items-center rounded-sm py-1.5 pr-2 pl-8 text-sm outline-none select-none data-[disabled]:pointer-events-none data-[disabled]:opacity-50',
                className,
            )}
            data-slot="context-menu-checkbox-item"
            {...props}
        >
            <span className="absolute left-2 flex h-3.5 w-3.5 items-center justify-center">
                <ContextMenuPrimitive.ItemIndicator>
                    <Check className="h-4 w-4" />
                </ContextMenuPrimitive.ItemIndicator>
            </span>
            {children}
        </ContextMenuPrimitive.CheckboxItem>
    );
}

function ContextMenuContent({ className, ...props }: React.ComponentProps<typeof ContextMenuPrimitive.Content>) {
    return (
        <ContextMenuPrimitive.Portal>
            <ContextMenuPrimitive.Content
                className={cn(
                    'bg-popover text-popover-foreground data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2 z-50 max-h-[--radix-context-menu-content-available-height] min-w-[8rem] origin-[--radix-context-menu-content-transform-origin] overflow-x-hidden overflow-y-auto rounded-md border p-1 shadow-md',
                    className,
                )}
                data-slot="context-menu-content"
                {...props}
            />
        </ContextMenuPrimitive.Portal>
    );
}

function ContextMenuGroup({ ...props }: React.ComponentProps<typeof ContextMenuPrimitive.Group>) {
    return (
        <ContextMenuPrimitive.Group
            data-slot="context-menu-group"
            {...props}
        />
    );
}

function ContextMenuItem({
    className,
    inset,
    variant = 'default',
    ...props
}: React.ComponentProps<typeof ContextMenuPrimitive.Item> & {
    inset?: boolean;
    variant?: 'default' | 'destructive';
}) {
    return (
        <ContextMenuPrimitive.Item
            className={cn(
                "focus:bg-accent focus:text-accent-foreground data-[variant=destructive]:text-destructive data-[variant=destructive]:focus:bg-destructive/10 data-[variant=destructive]:focus:text-destructive dark:data-[variant=destructive]:focus:bg-destructive/20 relative flex cursor-default items-center gap-2 rounded-sm px-2 py-1.5 text-sm outline-none select-none data-[disabled]:pointer-events-none data-[disabled]:opacity-50 [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4",
                inset && 'pl-8',
                className,
            )}
            data-slot="context-menu-item"
            data-variant={variant}
            {...props}
        />
    );
}

function ContextMenuLabel({
    className,
    inset,
    ...props
}: React.ComponentProps<typeof ContextMenuPrimitive.Label> & {
    inset?: boolean;
}) {
    return (
        <ContextMenuPrimitive.Label
            className={cn('text-foreground px-2 py-1.5 text-sm font-semibold', inset && 'pl-8', className)}
            data-slot="context-menu-label"
            {...props}
        />
    );
}

function ContextMenuPortal({ ...props }: React.ComponentProps<typeof ContextMenuPrimitive.Portal>) {
    return (
        <ContextMenuPrimitive.Portal
            data-slot="context-menu-portal"
            {...props}
        />
    );
}

function ContextMenuRadioGroup({ ...props }: React.ComponentProps<typeof ContextMenuPrimitive.RadioGroup>) {
    return (
        <ContextMenuPrimitive.RadioGroup
            data-slot="context-menu-radio-group"
            {...props}
        />
    );
}

function ContextMenuRadioItem({
    children,
    className,
    ...props
}: React.ComponentProps<typeof ContextMenuPrimitive.RadioItem>) {
    return (
        <ContextMenuPrimitive.RadioItem
            className={cn(
                'focus:bg-accent focus:text-accent-foreground relative flex cursor-default items-center rounded-sm py-1.5 pr-2 pl-8 text-sm outline-none select-none data-[disabled]:pointer-events-none data-[disabled]:opacity-50',
                className,
            )}
            data-slot="context-menu-radio-item"
            {...props}
        >
            <span className="absolute left-2 flex h-3.5 w-3.5 items-center justify-center">
                <ContextMenuPrimitive.ItemIndicator>
                    <Circle className="h-4 w-4 fill-current" />
                </ContextMenuPrimitive.ItemIndicator>
            </span>
            {children}
        </ContextMenuPrimitive.RadioItem>
    );
}

function ContextMenuSeparator({ className, ...props }: React.ComponentProps<typeof ContextMenuPrimitive.Separator>) {
    return (
        <ContextMenuPrimitive.Separator
            className={cn('bg-border -mx-1 my-1 h-px', className)}
            data-slot="context-menu-separator"
            {...props}
        />
    );
}

function ContextMenuShortcut({ className, ...props }: React.ComponentProps<'span'>) {
    return (
        <span
            className={cn('text-muted-foreground ml-auto text-xs tracking-widest', className)}
            data-slot="context-menu-shortcut"
            {...props}
        />
    );
}

function ContextMenuSub({ ...props }: React.ComponentProps<typeof ContextMenuPrimitive.Sub>) {
    return (
        <ContextMenuPrimitive.Sub
            data-slot="context-menu-sub"
            {...props}
        />
    );
}

function ContextMenuSubContent({ className, ...props }: React.ComponentProps<typeof ContextMenuPrimitive.SubContent>) {
    return (
        <ContextMenuPrimitive.SubContent
            className={cn(
                'bg-popover text-popover-foreground data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2 z-50 min-w-[8rem] origin-[--radix-context-menu-content-transform-origin] overflow-hidden rounded-md border p-1 shadow-lg',
                className,
            )}
            data-slot="context-menu-sub-content"
            {...props}
        />
    );
}

function ContextMenuSubTrigger({
    children,
    className,
    inset,
    ...props
}: React.ComponentProps<typeof ContextMenuPrimitive.SubTrigger> & {
    inset?: boolean;
}) {
    return (
        <ContextMenuPrimitive.SubTrigger
            className={cn(
                'focus:bg-accent focus:text-accent-foreground data-[state=open]:bg-accent data-[state=open]:text-accent-foreground flex cursor-default items-center rounded-sm px-2 py-1.5 text-sm outline-none select-none',
                inset && 'pl-8',
                className,
            )}
            data-slot="context-menu-sub-trigger"
            {...props}
        >
            {children}
            <ChevronRight className="ml-auto h-4 w-4" />
        </ContextMenuPrimitive.SubTrigger>
    );
}

function ContextMenuTrigger({ ...props }: React.ComponentProps<typeof ContextMenuPrimitive.Trigger>) {
    return (
        <ContextMenuPrimitive.Trigger
            data-slot="context-menu-trigger"
            {...props}
        />
    );
}

export {
    ContextMenu,
    ContextMenuCheckboxItem,
    ContextMenuContent,
    ContextMenuGroup,
    ContextMenuItem,
    ContextMenuLabel,
    ContextMenuPortal,
    ContextMenuRadioGroup,
    ContextMenuRadioItem,
    ContextMenuSeparator,
    ContextMenuShortcut,
    ContextMenuSub,
    ContextMenuSubContent,
    ContextMenuSubTrigger,
    ContextMenuTrigger,
};
