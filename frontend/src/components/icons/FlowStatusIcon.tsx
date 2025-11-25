import type { LucideIcon } from 'lucide-react';

import { CircleCheck, CircleDashed, CircleOff, CircleX, Loader2 } from 'lucide-react';

import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { StatusType } from '@/graphql/types';
import { cn } from '@/lib/utils';

interface FlowStatusIconProps {
    className?: string;
    status?: null | StatusType | undefined;
    tooltip?: string;
}

const statusIcons: Record<StatusType, { className: string; icon: LucideIcon }> = {
    [StatusType.Created]: { className: 'text-blue-500', icon: CircleDashed },
    [StatusType.Failed]: { className: 'text-red-500', icon: CircleX },
    [StatusType.Finished]: { className: 'text-green-500', icon: CircleCheck },
    [StatusType.Running]: { className: 'animate-spin text-purple-500', icon: Loader2 },
    [StatusType.Waiting]: { className: 'text-yellow-500', icon: CircleDashed },
};
const defaultIcon = { className: 'text-muted-foreground', icon: CircleOff };

export const FlowStatusIcon = ({ className = 'size-4', status, tooltip }: FlowStatusIconProps) => {
    if (!status) {
        return null;
    }

    const { className: defaultClassName, icon: Icon } = statusIcons[status] || defaultIcon;
    const iconElement = <Icon className={cn('shrink-0', defaultClassName, className, tooltip && 'cursor-pointer')} />;

    if (!tooltip) {
        return iconElement;
    }

    return (
        <Tooltip>
            <TooltipTrigger asChild>{iconElement}</TooltipTrigger>
            <TooltipContent>{tooltip}</TooltipContent>
        </Tooltip>
    );
};
