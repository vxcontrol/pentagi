import * as SeparatorPrimitive from '@radix-ui/react-separator';
import * as React from 'react';

import { cn } from '@/lib/utils';

function Separator({
    className,
    decorative = true,
    orientation = 'horizontal',
    ...props
}: React.ComponentProps<typeof SeparatorPrimitive.Root>) {
    return (
        <SeparatorPrimitive.Root
            className={cn(
                'bg-border shrink-0',
                orientation === 'horizontal' ? 'h-px w-full' : 'h-full w-px',
                className,
            )}
            decorative={decorative}
            orientation={orientation}
            {...props}
        />
    );
}

export { Separator };
