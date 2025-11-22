import type { Provider } from '@/models/Provider';

import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { ProviderType } from '@/graphql/types';
import { cn } from '@/lib/utils';

import Anthropic from './Anthropic';
import Bedrock from './Bedrock';
import Custom from './Custom';
import Gemini from './Gemini';
import Ollama from './Ollama';
import OpenAi from './OpenAi';

interface ProviderIconProps {
    className?: string;
    hasTooltip?: boolean;
    provider: null | Provider | undefined;
}

export const ProviderIcon = ({ className = 'size-4', hasTooltip = false, provider }: ProviderIconProps) => {
    if (!provider || !provider.type) {
        return null;
    }

    const iconClassName = cn(className, hasTooltip ? 'cursor-pointer' : '');
    const tooltip = provider.name === provider.type ? provider.name : `${provider.name} - ${provider.type}`;

    const renderIcon = () => {
        switch (provider.type) {
            case ProviderType.Anthropic: {
                return (
                    <Anthropic
                        aria-label="Anthropic"
                        className={cn(iconClassName, 'text-purple-500')}
                    />
                );
            }

            case ProviderType.Bedrock: {
                return (
                    <Bedrock
                        aria-label="Bedrock"
                        className={cn(iconClassName, 'text-blue-500')}
                    />
                );
            }

            case ProviderType.Gemini: {
                return (
                    <Gemini
                        aria-label="Gemini"
                        className={cn(iconClassName, 'text-blue-500')}
                    />
                );
            }

            case ProviderType.Ollama: {
                return (
                    <Ollama
                        aria-label="Ollama"
                        className={cn(iconClassName, 'text-blue-500')}
                    />
                );
            }

            case ProviderType.Openai: {
                return (
                    <OpenAi
                        aria-label="OpenAI"
                        className={cn(iconClassName, 'text-blue-500')}
                    />
                );
            }

            default: {
                return (
                    <Custom
                        aria-label="Custom provider"
                        className={cn(iconClassName, 'text-blue-500')}
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
