import * as AccordionPrimitive from '@radix-ui/react-accordion';
import { ChevronDown } from 'lucide-react';
import * as React from 'react';

import { cn } from '@/lib/utils';

function Accordion({ ...props }: React.ComponentProps<typeof AccordionPrimitive.Root>) {
    return (
        <AccordionPrimitive.Root
            data-slot="accordion"
            {...props}
        />
    );
}

function AccordionContent({ children, className, ...props }: React.ComponentProps<typeof AccordionPrimitive.Content>) {
    return (
        <AccordionPrimitive.Content
            className="data-[state=closed]:animate-accordion-up data-[state=open]:animate-accordion-down overflow-hidden text-sm"
            data-slot="accordion-content"
            {...props}
        >
            <div className={cn('pt-0 pb-4', className)}>{children}</div>
        </AccordionPrimitive.Content>
    );
}

function AccordionItem({ className, ...props }: React.ComponentProps<typeof AccordionPrimitive.Item>) {
    return (
        <AccordionPrimitive.Item
            className={cn('border-b', className)}
            data-slot="accordion-item"
            {...props}
        />
    );
}

function AccordionTrigger({ children, className, ...props }: React.ComponentProps<typeof AccordionPrimitive.Trigger>) {
    return (
        <AccordionPrimitive.Header className="flex">
            <AccordionPrimitive.Trigger
                className={cn(
                    'flex flex-1 items-center justify-between py-4 text-left text-sm font-medium transition-all hover:underline [&[data-state=open]>svg]:rotate-180',
                    className,
                )}
                data-slot="accordion-trigger"
                {...props}
            >
                {children}
                <ChevronDown className="text-muted-foreground h-4 w-4 shrink-0 transition-transform duration-200" />
            </AccordionPrimitive.Trigger>
        </AccordionPrimitive.Header>
    );
}

export { Accordion, AccordionContent, AccordionItem, AccordionTrigger };
