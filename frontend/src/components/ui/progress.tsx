import * as ProgressPrimitive from '@radix-ui/react-progress';
import * as React from 'react';

import { cn } from '@/lib/utils';

function Progress({ className, value, ...props }: React.ComponentProps<typeof ProgressPrimitive.Root>) {
    return (
        <ProgressPrimitive.Root
            className={cn('bg-primary/20 relative h-2 w-full overflow-hidden rounded-full', className)}
            data-slot="progress"
            {...props}
        >
            <ProgressPrimitive.Indicator
                className={cn('bg-primary h-full w-full flex-1 transition-all')}
                data-slot="progress-indicator"
                style={{ transform: `translateX(-${100 - (value || 0)}%)` }}
            />
        </ProgressPrimitive.Root>
    );
}

export { Progress };
