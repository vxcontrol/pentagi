'use client';

import * as ToggleGroupPrimitive from '@radix-ui/react-toggle-group';
import { type VariantProps } from 'class-variance-authority';
import * as React from 'react';

import { toggleVariants } from '@/components/ui/toggle';
import { cn } from '@/lib/utils';

const ToggleGroupContext = React.createContext<VariantProps<typeof toggleVariants>>({
    size: 'default',
    variant: 'default',
});

const ToggleGroup = ({
    children,
    className,
    ref,
    size,
    variant,
    ...props
}: React.ComponentPropsWithoutRef<typeof ToggleGroupPrimitive.Root> &
    VariantProps<typeof toggleVariants> & {
        ref?: React.Ref<React.ElementRef<typeof ToggleGroupPrimitive.Root>>;
    }) => (
    <ToggleGroupPrimitive.Root
        className={cn('flex items-center justify-center gap-1', className)}
        ref={ref}
        {...props}
    >
        <ToggleGroupContext.Provider value={{ size, variant }}>{children}</ToggleGroupContext.Provider>
    </ToggleGroupPrimitive.Root>
);

ToggleGroup.displayName = ToggleGroupPrimitive.Root.displayName;

const ToggleGroupItem = ({
    children,
    className,
    ref,
    size,
    variant,
    ...props
}: React.ComponentPropsWithoutRef<typeof ToggleGroupPrimitive.Item> &
    VariantProps<typeof toggleVariants> & {
        ref?: React.Ref<React.ElementRef<typeof ToggleGroupPrimitive.Item>>;
    }) => {
    const context = React.useContext(ToggleGroupContext);

    return (
        <ToggleGroupPrimitive.Item
            className={cn(
                toggleVariants({
                    size: context.size || size,
                    variant: context.variant || variant,
                }),
                className,
            )}
            ref={ref}
            {...props}
        >
            {children}
        </ToggleGroupPrimitive.Item>
    );
};

ToggleGroupItem.displayName = ToggleGroupPrimitive.Item.displayName;

export { ToggleGroup, ToggleGroupItem };
