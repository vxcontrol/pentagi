import * as AvatarPrimitive from '@radix-ui/react-avatar';
import * as React from 'react';

import { cn } from '@/lib/utils';

function Avatar({ className, ...props }: React.ComponentProps<typeof AvatarPrimitive.Root>) {
    return (
        <AvatarPrimitive.Root
            className={cn('relative flex h-10 w-10 shrink-0 overflow-hidden rounded-full', className)}
            data-slot="avatar"
            {...props}
        />
    );
}

function AvatarFallback({ className, ...props }: React.ComponentProps<typeof AvatarPrimitive.Fallback>) {
    return (
        <AvatarPrimitive.Fallback
            className={cn('bg-muted flex h-full w-full items-center justify-center rounded-full', className)}
            data-slot="avatar-fallback"
            {...props}
        />
    );
}

function AvatarImage({ className, ...props }: React.ComponentProps<typeof AvatarPrimitive.Image>) {
    return (
        <AvatarPrimitive.Image
            className={cn('aspect-square h-full w-full', className)}
            data-slot="avatar-image"
            {...props}
        />
    );
}

export { Avatar, AvatarFallback, AvatarImage };
