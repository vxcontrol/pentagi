import * as CollapsiblePrimitive from '@radix-ui/react-collapsible';
import * as React from 'react';

function Collapsible({ ...props }: React.ComponentProps<typeof CollapsiblePrimitive.Root>) {
    return (
        <CollapsiblePrimitive.Root
            data-slot="collapsible"
            {...props}
        />
    );
}

function CollapsibleContent({ ...props }: React.ComponentProps<typeof CollapsiblePrimitive.CollapsibleContent>) {
    return (
        <CollapsiblePrimitive.CollapsibleContent
            data-slot="collapsible-content"
            {...props}
        />
    );
}

function CollapsibleTrigger({ ...props }: React.ComponentProps<typeof CollapsiblePrimitive.CollapsibleTrigger>) {
    return (
        <CollapsiblePrimitive.CollapsibleTrigger
            data-slot="collapsible-trigger"
            {...props}
        />
    );
}

export { Collapsible, CollapsibleContent, CollapsibleTrigger };
