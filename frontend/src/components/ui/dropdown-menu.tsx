import * as DropdownMenuPrimitive from '@radix-ui/react-dropdown-menu';
import { Check, ChevronRight, Circle } from 'lucide-react';
import * as React from 'react';

import { cn } from '@/lib/utils';

function DropdownMenu(props: React.ComponentProps<typeof DropdownMenuPrimitive.Root>) {
    return (
        <DropdownMenuPrimitive.Root
            data-slot="dropdown-menu"
            {...props}
        />
    );
}

function DropdownMenuCheckboxItem({
    checked,
    children,
    className,
    ...props
}: React.ComponentProps<typeof DropdownMenuPrimitive.CheckboxItem>) {
    return (
        <DropdownMenuPrimitive.CheckboxItem
            checked={checked}
            className={cn(
                'focus:bg-accent focus:text-accent-foreground relative flex cursor-default items-center rounded-sm py-1.5 pr-2 pl-8 text-sm outline-hidden transition-colors select-none data-disabled:pointer-events-none data-disabled:opacity-50',
                className,
            )}
            data-slot="dropdown-menu-checkbox-item"
            {...props}
        >
            <span className="absolute left-2 flex h-3.5 w-3.5 items-center justify-center">
                <DropdownMenuPrimitive.ItemIndicator>
                    <Check className="h-4 w-4" />
                </DropdownMenuPrimitive.ItemIndicator>
            </span>
            {children}
        </DropdownMenuPrimitive.CheckboxItem>
    );
}

function DropdownMenuContent({
    className,
    sideOffset = 4,
    ...props
}: React.ComponentProps<typeof DropdownMenuPrimitive.Content>) {
    return (
        <DropdownMenuPrimitive.Portal>
            <DropdownMenuPrimitive.Content
                className={cn(
                    'bg-popover text-popover-foreground z-50 max-h-(--radix-dropdown-menu-content-available-height) min-w-32 origin-(--radix-dropdown-menu-content-transform-origin) overflow-x-hidden overflow-y-auto rounded-md border p-1 shadow-md',
                    'data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2',
                    className,
                )}
                data-slot="dropdown-menu-content"
                sideOffset={sideOffset}
                {...props}
            />
        </DropdownMenuPrimitive.Portal>
    );
}

function DropdownMenuGroup(props: React.ComponentProps<typeof DropdownMenuPrimitive.Group>) {
    return (
        <DropdownMenuPrimitive.Group
            data-slot="dropdown-menu-group"
            {...props}
        />
    );
}

function DropdownMenuItem({
    className,
    inset,
    variant = 'default',
    ...props
}: React.ComponentProps<typeof DropdownMenuPrimitive.Item> & {
    inset?: boolean;
    variant?: 'default' | 'destructive';
}) {
    return (
        <DropdownMenuPrimitive.Item
            className={cn(
                'focus:bg-accent focus:text-accent-foreground data-[variant=destructive]:text-destructive data-[variant=destructive]:focus:bg-destructive/10 data-[variant=destructive]:focus:text-destructive dark:data-[variant=destructive]:focus:bg-destructive/20 relative flex cursor-default items-center gap-2 rounded-sm px-2 py-1.5 text-sm outline-hidden transition-colors select-none data-disabled:pointer-events-none data-disabled:opacity-50 [&>svg]:size-4 [&>svg]:shrink-0',
                inset && 'pl-8',
                className,
            )}
            data-slot="dropdown-menu-item"
            data-variant={variant}
            {...props}
        />
    );
}

function DropdownMenuLabel({
    className,
    inset,
    ...props
}: React.ComponentProps<typeof DropdownMenuPrimitive.Label> & {
    inset?: boolean;
}) {
    return (
        <DropdownMenuPrimitive.Label
            className={cn('px-2 py-1.5 text-sm font-semibold', inset && 'pl-8', className)}
            data-slot="dropdown-menu-label"
            {...props}
        />
    );
}

function DropdownMenuPortal(props: React.ComponentProps<typeof DropdownMenuPrimitive.Portal>) {
    return (
        <DropdownMenuPrimitive.Portal
            data-slot="dropdown-menu-portal"
            {...props}
        />
    );
}

function DropdownMenuRadioGroup(props: React.ComponentProps<typeof DropdownMenuPrimitive.RadioGroup>) {
    return (
        <DropdownMenuPrimitive.RadioGroup
            data-slot="dropdown-menu-radio-group"
            {...props}
        />
    );
}

function DropdownMenuRadioItem({
    children,
    className,
    ...props
}: React.ComponentProps<typeof DropdownMenuPrimitive.RadioItem>) {
    return (
        <DropdownMenuPrimitive.RadioItem
            className={cn(
                'focus:bg-accent focus:text-accent-foreground relative flex cursor-default items-center rounded-sm py-1.5 pr-2 pl-8 text-sm outline-hidden transition-colors select-none data-disabled:pointer-events-none data-disabled:opacity-50',
                className,
            )}
            data-slot="dropdown-menu-radio-item"
            {...props}
        >
            <span className="absolute left-2 flex h-3.5 w-3.5 items-center justify-center">
                <DropdownMenuPrimitive.ItemIndicator>
                    <Circle className="h-2 w-2 fill-current" />
                </DropdownMenuPrimitive.ItemIndicator>
            </span>
            {children}
        </DropdownMenuPrimitive.RadioItem>
    );
}

function DropdownMenuSeparator({ className, ...props }: React.ComponentProps<typeof DropdownMenuPrimitive.Separator>) {
    return (
        <DropdownMenuPrimitive.Separator
            className={cn('bg-border -mx-1 my-1 h-px', className)}
            data-slot="dropdown-menu-separator"
            {...props}
        />
    );
}

function DropdownMenuShortcut({ className, ...props }: React.ComponentProps<'span'>) {
    return (
        <span
            className={cn('ml-auto text-xs tracking-widest opacity-60', className)}
            data-slot="dropdown-menu-shortcut"
            {...props}
        />
    );
}

function DropdownMenuSub(props: React.ComponentProps<typeof DropdownMenuPrimitive.Sub>) {
    return (
        <DropdownMenuPrimitive.Sub
            data-slot="dropdown-menu-sub"
            {...props}
        />
    );
}

function DropdownMenuSubContent({
    className,
    ...props
}: React.ComponentProps<typeof DropdownMenuPrimitive.SubContent>) {
    return (
        <DropdownMenuPrimitive.SubContent
            className={cn(
                'bg-popover text-popover-foreground data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2 z-50 min-w-32 overflow-hidden rounded-md border p-1 shadow-lg',
                className,
            )}
            data-slot="dropdown-menu-sub-content"
            {...props}
        />
    );
}

function DropdownMenuSubTrigger({
    children,
    className,
    inset,
    ...props
}: React.ComponentProps<typeof DropdownMenuPrimitive.SubTrigger> & {
    inset?: boolean;
}) {
    return (
        <DropdownMenuPrimitive.SubTrigger
            className={cn(
                'focus:bg-accent data-[state=open]:bg-accent flex cursor-default items-center gap-2 rounded-sm px-2 py-1.5 text-sm outline-hidden select-none [&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0',
                inset && 'pl-8',
                className,
            )}
            data-slot="dropdown-menu-sub-trigger"
            {...props}
        >
            {children}
            <ChevronRight className="ml-auto" />
        </DropdownMenuPrimitive.SubTrigger>
    );
}

function DropdownMenuTrigger(props: React.ComponentProps<typeof DropdownMenuPrimitive.Trigger>) {
    return (
        <DropdownMenuPrimitive.Trigger
            data-slot="dropdown-menu-trigger"
            {...props}
        />
    );
}

export {
    DropdownMenu,
    DropdownMenuCheckboxItem,
    DropdownMenuContent,
    DropdownMenuGroup,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuPortal,
    DropdownMenuRadioGroup,
    DropdownMenuRadioItem,
    DropdownMenuSeparator,
    DropdownMenuShortcut,
    DropdownMenuSub,
    DropdownMenuSubContent,
    DropdownMenuSubTrigger,
    DropdownMenuTrigger,
};
