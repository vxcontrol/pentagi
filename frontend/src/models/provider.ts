import { ProviderType } from '@/graphql/types';

export interface Provider {
    name: string;
    type: ProviderType;
}

export const getProviderDisplayName = (provider: Provider): string => {
    return provider.name;
};

export const isProviderValid = (provider: Provider, providers: Provider[]): boolean => {
    return providers.some((p) => p.name === provider.name && p.type === provider.type);
};

export const isProviderMissing = (provider: Provider, providers: Provider[]): boolean => {
    return providers.length > 0 && !isProviderValid(provider, providers);
};

export const findProviderByName = (providerName: string, providers: Provider[]): Provider | undefined => {
    return providers.find((provider) => provider.name === providerName);
};

export const sortProviders = (providers: Provider[]): Provider[] => {
    return [...providers].sort((a, b) => a.name.localeCompare(b.name));
};
