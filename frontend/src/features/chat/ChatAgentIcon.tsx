import type { LucideIcon } from 'lucide-react';

import {
    Bot,
    Brain,
    Code2,
    FileText,
    HardDrive,
    HardDriveDownload,
    HelpCircle,
    LayoutList,
    MessagesSquare,
    RefreshCw,
    Search,
    Settings,
    Sigma,
    Skull,
    Wrench,
} from 'lucide-react';

import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { AgentType } from '@/graphql/types';
import { cn } from '@/lib/utils';
import { formatName } from '@/lib/utils/format';

interface ChatAgentIconProps {
    className?: string;
    tooltip?: string;
    type?: AgentType;
}

const icons: Record<AgentType, LucideIcon> = {
    [AgentType.Adviser]: HelpCircle,
    [AgentType.Assistant]: Bot,
    [AgentType.Coder]: Code2,
    [AgentType.Enricher]: HardDriveDownload,
    [AgentType.Generator]: LayoutList,
    [AgentType.Installer]: Settings,
    [AgentType.Memorist]: HardDrive,
    [AgentType.Pentester]: Skull,
    [AgentType.PrimaryAgent]: Brain,
    [AgentType.Refiner]: RefreshCw,
    [AgentType.Reflector]: MessagesSquare,
    [AgentType.Reporter]: FileText,
    [AgentType.Searcher]: Search,
    [AgentType.Summarizer]: Sigma,
    [AgentType.ToolCallFixer]: Wrench,
};
const defaultIcon = HelpCircle;

const ChatAgentIcon = ({ className, type, tooltip = type }: ChatAgentIconProps) => {
    const Icon = type ? icons[type] || defaultIcon : defaultIcon;
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

export default ChatAgentIcon;
