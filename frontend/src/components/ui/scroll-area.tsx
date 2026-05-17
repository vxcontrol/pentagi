import * as ScrollAreaPrimitive from '@radix-ui/react-scroll-area';
import * as React from 'react';

import { cn } from '@/lib/utils';

const ScrollArea = ({
    children,
    className,
    ref,
    ...props
}: React.ComponentPropsWithoutRef<typeof ScrollAreaPrimitive.Root> & {
    ref?: React.Ref<React.ElementRef<typeof ScrollAreaPrimitive.Root>>;
}) => (
    <ScrollAreaPrimitive.Root
        className={cn('relative overflow-hidden', className)}
        ref={ref}
        {...props}
    >
        <ScrollAreaPrimitive.Viewport className="size-full rounded-[inherit]">{children}</ScrollAreaPrimitive.Viewport>
        <ScrollBar />
        <ScrollAreaPrimitive.Corner />
    </ScrollAreaPrimitive.Root>
);
ScrollArea.displayName = ScrollAreaPrimitive.Root.displayName;

const ScrollBar = ({
    className,
    orientation = 'vertical',
    ref,
    ...props
}: React.ComponentPropsWithoutRef<typeof ScrollAreaPrimitive.ScrollAreaScrollbar> & {
    ref?: React.Ref<React.ElementRef<typeof ScrollAreaPrimitive.ScrollAreaScrollbar>>;
}) => (
    <ScrollAreaPrimitive.ScrollAreaScrollbar
        className={cn(
            'flex touch-none transition-colors select-none',
            orientation === 'vertical' && 'h-full w-2.5 border-l border-l-transparent p-px',
            orientation === 'horizontal' && 'h-2.5 flex-col border-t border-t-transparent p-px',
            className,
        )}
        orientation={orientation}
        ref={ref}
        {...props}
    >
        <ScrollAreaPrimitive.ScrollAreaThumb className="bg-border relative flex-1 rounded-full" />
    </ScrollAreaPrimitive.ScrollAreaScrollbar>
);
ScrollBar.displayName = ScrollAreaPrimitive.ScrollAreaScrollbar.displayName;

export { ScrollArea, ScrollBar };
