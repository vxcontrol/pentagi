import type { LucideIcon } from 'lucide-react';
import { CheckCircle2, CircleDashed, CircleX, Clock, Loader2, PlayCircle } from 'lucide-react';

import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { StatusType } from '@/graphql/types';
import { cn } from '@/lib/utils';
import { formatName } from '@/lib/utils/format';

interface ChatTaskStatusIconProps {
    status?: StatusType;
    className?: string;
    tooltip?: string;
}

const statusIcons: Record<StatusType, { icon: LucideIcon; className: string }> = {
    [StatusType.Finished]: { icon: CheckCircle2, className: 'text-green-500' },
    [StatusType.Failed]: { icon: CircleX, className: 'text-red-500' },
    [StatusType.Running]: { icon: Loader2, className: 'animate-spin text-purple-500' },
    [StatusType.Created]: { icon: PlayCircle, className: 'text-blue-500' },
    [StatusType.Waiting]: { icon: Clock, className: 'text-yellow-500' },
};
const defaultIcon = { icon: CircleDashed, className: 'text-muted-foreground' };

const ChatTaskStatusIcon = ({ status, className, tooltip }: ChatTaskStatusIconProps) => {
    const { icon: Icon, className: defaultClassName } = status ? statusIcons[status] || defaultIcon : defaultIcon;
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

export default ChatTaskStatusIcon;
