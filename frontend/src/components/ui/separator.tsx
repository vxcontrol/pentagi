import * as SeparatorPrimitive from '@radix-ui/react-separator';
import * as React from 'react';

import { cn } from '@/lib/utils';

const Separator = ({
    className,
    decorative = true,
    orientation = 'horizontal',
    ref,
    ...props
}: React.ComponentPropsWithoutRef<typeof SeparatorPrimitive.Root> & {
    ref?: React.Ref<React.ElementRef<typeof SeparatorPrimitive.Root>>;
}) => (
    <SeparatorPrimitive.Root
        className={cn('bg-border shrink-0', orientation === 'horizontal' ? 'h-px w-full' : 'h-full w-px', className)}
        decorative={decorative}
        orientation={orientation}
        ref={ref}
        {...props}
    />
);
Separator.displayName = SeparatorPrimitive.Root.displayName;

export { Separator };
