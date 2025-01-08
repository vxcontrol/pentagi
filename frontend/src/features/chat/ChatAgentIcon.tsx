import type { LucideIcon } from 'lucide-react';
import {
    Airplay,
    Brain,
    Code2,
    FileCode,
    FileText,
    HardDrive,
    HardDriveDownload,
    HelpCircle,
    RefreshCw,
    Search,
    Settings,
    Sigma,
    Skull,
} from 'lucide-react';

import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { AgentType } from '@/graphql/types';
import { cn } from '@/lib/utils';
import { formatName } from '@/lib/utils/format';

interface ChatAgentIconProps {
    type?: AgentType;
    className?: string;
    tooltip?: string;
}

const icons: Record<AgentType, LucideIcon> = {
    [AgentType.Adviser]: HelpCircle,
    [AgentType.Coder]: Code2,
    [AgentType.Enricher]: HardDriveDownload,
    [AgentType.Generator]: FileCode,
    [AgentType.Installer]: Settings,
    [AgentType.Memorist]: HardDrive,
    [AgentType.Pentester]: Skull,
    [AgentType.PrimaryAgent]: Brain,
    [AgentType.Refiner]: RefreshCw,
    [AgentType.Reflector]: Airplay,
    [AgentType.Reporter]: FileText,
    [AgentType.Searcher]: Search,
    [AgentType.Summarizer]: Sigma,
};
const defaultIcon = HelpCircle;

const ChatAgentIcon = ({ type, className, tooltip = type }: ChatAgentIconProps) => {
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
