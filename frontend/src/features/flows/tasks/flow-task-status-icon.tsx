import type { LucideIcon } from 'lucide-react';

import { CheckCircle2, CircleDashed, CircleX, Clock, Loader2, PlayCircle } from 'lucide-react';

import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { StatusType } from '@/graphql/types';
import { cn } from '@/lib/utils';
import { formatName } from '@/lib/utils/format';

interface FlowTaskStatusIconProps {
    className?: string;
    status?: StatusType;
    tooltip?: string;
}

const statusIcons: Record<StatusType, { className: string; icon: LucideIcon }> = {
    [StatusType.Created]: { className: 'text-blue-500', icon: PlayCircle },
    [StatusType.Failed]: { className: 'text-red-500', icon: CircleX },
    [StatusType.Finished]: { className: 'text-green-500', icon: CheckCircle2 },
    [StatusType.Running]: { className: 'animate-spin text-purple-500', icon: Loader2 },
    [StatusType.Waiting]: { className: 'text-yellow-500', icon: Clock },
};
const defaultIcon = { className: 'text-muted-foreground', icon: CircleDashed };

const FlowTaskStatusIcon = ({ className, status, tooltip }: FlowTaskStatusIconProps) => {
    const { className: defaultClassName, icon: Icon } = status ? statusIcons[status] || defaultIcon : defaultIcon;
    const iconElement = (
        <Icon className={cn('size-4 shrink-0', defaultClassName, tooltip && 'cursor-pointer', className)} />
    );

    if (!tooltip) {
        return iconElement;
    }

    return (
        <Tooltip>
            <TooltipTrigger asChild>{iconElement}</TooltipTrigger>
            <TooltipContent>{formatName(tooltip)}</TooltipContent>
        </Tooltip>
    );
};

export default FlowTaskStatusIcon;
