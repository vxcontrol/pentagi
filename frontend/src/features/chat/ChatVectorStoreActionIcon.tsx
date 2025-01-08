import type { LucideIcon } from 'lucide-react';
import { HardDrive, HardDriveDownload, HardDriveUpload } from 'lucide-react';

import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { VectorStoreAction } from '@/graphql/types';
import { cn } from '@/lib/utils';
import { formatName } from '@/lib/utils/format';

interface ChatVectorStoreActionIconProps {
    action?: VectorStoreAction;
    className?: string;
    tooltip?: string;
}

const icons: Record<VectorStoreAction, LucideIcon> = {
    [VectorStoreAction.Store]: HardDriveDownload,
    [VectorStoreAction.Retrieve]: HardDriveUpload,
};
const defaultIcon = HardDrive;

const ChatVectorStoreActionIcon = ({ action, className, tooltip = action }: ChatVectorStoreActionIconProps) => {
    const Icon = action ? icons[action] || defaultIcon : defaultIcon;
    const iconElement = <Icon className={cn('size-3 shrink-0', tooltip && 'cursor-pointer', className)} />;

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

export default ChatVectorStoreActionIcon;
