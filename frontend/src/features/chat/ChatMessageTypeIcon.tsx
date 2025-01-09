import type { LucideIcon } from 'lucide-react';
import { Bot, Brain, CheckSquare, FileText, Globe, HelpCircle, Search, Terminal, User as UserIcon } from 'lucide-react';

import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { MessageLogType } from '@/graphql/types';
import { cn } from '@/lib/utils';
import { formatName } from '@/lib/utils/format';

interface MessageTypeIconProps {
    type?: MessageLogType;
    className?: string;
    tooltip?: string;
}

const messageTypeIcons: Record<MessageLogType, LucideIcon> = {
    [MessageLogType.Input]: UserIcon,
    [MessageLogType.Browser]: Globe,
    [MessageLogType.Search]: Search,
    [MessageLogType.Terminal]: Terminal,
    [MessageLogType.File]: FileText,
    [MessageLogType.Ask]: HelpCircle,
    [MessageLogType.Done]: CheckSquare,
    [MessageLogType.Thoughts]: Brain,
    [MessageLogType.Advice]: Bot,
};
const defaultIcon = Brain;

const ChatMessageTypeIcon = ({ type, className, tooltip = type }: MessageTypeIconProps) => {
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

export default ChatMessageTypeIcon;
