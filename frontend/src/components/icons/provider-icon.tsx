import type { ComponentType } from 'react';

import type { Provider } from '@/models/provider';

import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { ProviderType } from '@/graphql/types';
import { cn } from '@/lib/utils';

import Anthropic from './anthropic';
import Bedrock from './bedrock';
import Custom from './custom';
import DeepSeek from './deepseek';
import Gemini from './gemini';
import GLM from './glm';
import Kimi from './kimi';
import Ollama from './ollama';
import OpenAi from './open-ai';
import Qwen from './qwen';

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
    [ProviderType.Deepseek]: { className: 'text-blue-600', icon: DeepSeek },
    [ProviderType.Gemini]: { className: 'text-blue-500', icon: Gemini },
    [ProviderType.Glm]: { className: 'text-violet-500', icon: GLM },
    [ProviderType.Kimi]: { className: 'text-sky-500', icon: Kimi },
    [ProviderType.Ollama]: { className: 'text-blue-500', icon: Ollama },
    [ProviderType.Openai]: { className: 'text-blue-500', icon: OpenAi },
    [ProviderType.Qwen]: { className: 'text-orange-500', icon: Qwen },
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
