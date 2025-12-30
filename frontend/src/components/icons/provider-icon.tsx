import type { ComponentType } from 'react';

import type { Provider } from '@/models/provider';

import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { ProviderType } from '@/graphql/types';
import { cn } from '@/lib/utils';

import Anthropic from './anthropic';
import Bedrock from './bedrock';
import Custom from './custom';
import Gemini from './gemini';
import Ollama from './ollama';
import OpenAi from './open-ai';

interface ProviderIconConfig {
    className: string;
    icon: ComponentType<{ className?: string }>;
}

interface ProviderIconProps {
    className?: string;
    provider: null | Provider | undefined;
    tooltip?: string;
}

const providerIcons: Record<ProviderType, ProviderIconConfig> = {
    [ProviderType.Anthropic]: { className: 'text-purple-500', icon: Anthropic },
    [ProviderType.Bedrock]: { className: 'text-blue-500', icon: Bedrock },
    [ProviderType.Custom]: { className: 'text-blue-500', icon: Custom },
    [ProviderType.Gemini]: { className: 'text-blue-500', icon: Gemini },
    [ProviderType.Ollama]: { className: 'text-blue-500', icon: Ollama },
    [ProviderType.Openai]: { className: 'text-blue-500', icon: OpenAi },
};
const defaultProviderIcon: ProviderIconConfig = { className: 'text-blue-500', icon: Custom };

export const ProviderIcon = ({ className = 'size-4', provider, tooltip }: ProviderIconProps) => {
    if (!provider?.type) {
        return null;
    }

    const { className: defaultClassName, icon: Icon } = providerIcons[provider.type] || defaultProviderIcon;
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
