import { GripVertical } from 'lucide-react';
import { Group, Panel, Separator } from 'react-resizable-panels';

import { cn } from '@/lib/utils';

function ResizableHandle({
    className,
    withHandle,
    ...props
}: React.ComponentProps<typeof Separator> & {
    withHandle?: boolean;
}) {
    return (
        <Separator
            className={cn(
                'bg-border focus-visible:ring-ring relative flex w-px items-center justify-center after:absolute after:inset-y-0 after:left-1/2 after:w-1 after:-translate-x-1/2 focus-visible:ring-1 focus-visible:ring-offset-1 focus-visible:outline-hidden aria-[orientation=horizontal]:h-px aria-[orientation=horizontal]:w-full aria-[orientation=horizontal]:after:left-0 aria-[orientation=horizontal]:after:h-1 aria-[orientation=horizontal]:after:w-full aria-[orientation=horizontal]:after:translate-x-0 aria-[orientation=horizontal]:after:-translate-y-1/2 [&[aria-orientation=horizontal]>div]:rotate-90',
                className,
            )}
            data-slot="resizable-handle"
            {...props}
        >
            {withHandle && (
                <div className="bg-border z-10 flex h-4 w-3 items-center justify-center rounded-sm border">
                    <GripVertical className="h-2.5 w-2.5" />
                </div>
            )}
        </Separator>
    );
}

function ResizablePanel(props: React.ComponentProps<typeof Panel>) {
    return (
        <Panel
            data-slot="resizable-panel"
            {...props}
        />
    );
}

function ResizablePanelGroup({ className, ...props }: React.ComponentProps<typeof Group>) {
    return (
        <Group
            className={cn('flex h-full w-full aria-[orientation=vertical]:flex-col', className)}
            data-slot="resizable-panel-group"
            {...props}
        />
    );
}

export { ResizableHandle, ResizablePanel, ResizablePanelGroup };
