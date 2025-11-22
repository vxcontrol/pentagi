import { CircleCheck, CircleDashed, CircleOff, CircleX, Loader2 } from 'lucide-react';

import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { StatusType } from '@/graphql/types';
import { cn } from '@/lib/utils';

interface FlowStatusIconProps {
    className?: string;
    hasTooltip?: boolean;
    status?: null | StatusType | undefined;
}

export const FlowStatusIcon = ({ className = 'size-4', hasTooltip = false, status }: FlowStatusIconProps) => {
    if (!status) {
        return null;
    }

    const iconClassName = cn(className, hasTooltip ? 'cursor-pointer' : '');
    const tooltip = status;

    const renderIcon = () => {
        switch (status) {
            case StatusType.Created: {
                return (
                    <CircleDashed
                        aria-label="created"
                        className={cn(iconClassName, 'text-blue-500')}
                    />
                );
            }

            case StatusType.Failed: {
                return (
                    <CircleX
                        aria-label="failed"
                        className={cn(iconClassName, 'text-red-500')}
                    />
                );
            }

            case StatusType.Finished: {
                return (
                    <CircleCheck
                        aria-label="finished"
                        className={cn(iconClassName, 'text-green-500')}
                    />
                );
            }

            case StatusType.Running: {
                return (
                    <Loader2
                        aria-label="running"
                        className={cn(iconClassName, 'animate-spin text-purple-500')}
                    />
                );
            }

            case StatusType.Waiting: {
                return (
                    <CircleDashed
                        aria-label="waiting"
                        className={cn(iconClassName, 'text-yellow-500')}
                    />
                );
            }

            default: {
                return (
                    <CircleOff
                        aria-label="unknown status"
                        className={cn(iconClassName, 'text-muted-foreground')}
                    />
                );
            }
        }
    };

    const icon = renderIcon();

    if (hasTooltip) {
        return (
            <Tooltip>
                <TooltipTrigger asChild>{icon}</TooltipTrigger>
                <TooltipContent>{tooltip}</TooltipContent>
            </Tooltip>
        );
    }

    return icon;
};
