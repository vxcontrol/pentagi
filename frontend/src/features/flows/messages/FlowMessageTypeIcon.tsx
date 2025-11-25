import type { LucideIcon } from 'lucide-react';

import {
    BotMessageSquare,
    Brain,
    CheckSquare,
    FileText,
    Globe,
    HelpCircle,
    MessageSquareReply,
    NotepadText,
    Search,
    Terminal,
    User as UserIcon,
} from 'lucide-react';

import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { MessageLogType } from '@/graphql/types';
import { cn } from '@/lib/utils';
import { formatName } from '@/lib/utils/format';

interface MessageTypeIconProps {
    className?: string;
    tooltip?: string;
    type?: MessageLogType;
}

const messageTypeIcons: Record<MessageLogType, LucideIcon> = {
    [MessageLogType.Advice]: BotMessageSquare,
    [MessageLogType.Answer]: MessageSquareReply,
    [MessageLogType.Ask]: HelpCircle,
    [MessageLogType.Browser]: Globe,
    [MessageLogType.Done]: CheckSquare,
    [MessageLogType.File]: FileText,
    [MessageLogType.Input]: UserIcon,
    [MessageLogType.Report]: NotepadText,
    [MessageLogType.Search]: Search,
    [MessageLogType.Terminal]: Terminal,
    [MessageLogType.Thoughts]: Brain,
};
const defaultIcon = Brain;

const FlowMessageTypeIcon = ({ className, type, tooltip = type }: MessageTypeIconProps) => {
    const Icon = type ? messageTypeIcons[type] || defaultIcon : defaultIcon;
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

export default FlowMessageTypeIcon;

