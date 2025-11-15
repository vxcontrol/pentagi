import Anthropic from '@/components/icons/Anthropic';
import Bedrock from '@/components/icons/Bedrock';
import Custom from '@/components/icons/Custom';
import Gemini from '@/components/icons/Gemini';
import Ollama from '@/components/icons/Ollama';
import OpenAi from '@/components/icons/OpenAi';
import { ProviderType } from '@/graphql/types';

export interface Provider {
    name: string;
    type: ProviderType;
}

/**
 * Generates a display name for a provider
 * If the name matches the type, only the name is returned
 * Otherwise, returns "name - type"
 */
export const getProviderDisplayName = (provider: Provider): string => {
    return provider.name;
};

/**
 * Generates a tooltip for a provider
 * If the name matches the type, only the name is returned
 * Otherwise, returns "name - type"
 */
export const getProviderTooltip = (provider: Provider): string => {
    if (provider.name === provider.type) {
        return provider.name;
    }

    return `${provider.name} - ${provider.type}`;
};

/**
 * Checks if a provider exists in the list of providers
 */
export const isProviderValid = (provider: Provider, providers: Provider[]): boolean => {
    return providers.some((p) => p.name === provider.name && p.type === provider.type);
};

/**
 * Finds a provider by name and type
 */
export const findProvider = (provider: Provider, providers: Provider[]): Provider | undefined => {
    return providers.find((p) => p.name === provider.name && p.type === provider.type);
};

/**
 * Finds a provider by name
 */
export const findProviderByName = (providerName: string, providers: Provider[]): Provider | undefined => {
    return providers.find((provider) => provider.name === providerName);
};

/**
 * Sorts providers by name alphabetically
 */
export const sortProviders = (providers: Provider[]): Provider[] => {
    return [...providers].sort((a, b) => a.name.localeCompare(b.name));
};

/**
 * Gets the icon component for a provider based on its type
 */
export const getProviderIcon = (provider: Provider, className = 'h-4 w-4') => {
    if (!provider || !provider.type) {
        return null;
    }

    switch (provider.type) {
        case ProviderType.Anthropic: {
            return (
                <Anthropic
                    aria-label="Anthropic"
                    className={`${className} text-purple-500`}
                />
            );
        }

        case ProviderType.Bedrock: {
            return (
                <Bedrock
                    aria-label="Bedrock"
                    className={`${className} text-blue-500`}
                />
            );
        }

        case ProviderType.Gemini: {
            return (
                <Gemini
                    aria-label="Gemini"
                    className={`${className} text-blue-500`}
                />
            );
        }

        case ProviderType.Ollama: {
            return (
                <Ollama
                    aria-label="Ollama"
                    className={`${className} text-blue-500`}
                />
            );
        }

        case ProviderType.Openai: {
            return (
                <OpenAi
                    aria-label="OpenAI"
                    className={`${className} text-blue-500`}
                />
            );
        }

        default: {
            return (
                <Custom
                    aria-label="Custom provider"
                    className={`${className} text-blue-500`}
                />
            );
        }
    }
};
