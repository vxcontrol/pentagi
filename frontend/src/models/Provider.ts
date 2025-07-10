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
  if (provider.name === provider.type) {
    return provider.name;
  }
  return `${provider.name} - ${provider.type}`;
};

/**
 * Generates a tooltip for a provider
 * If the name matches the type, only the name is returned
 * Otherwise, returns "name - type"
 */
export const getProviderTooltip = (provider: Provider): string => {
  return getProviderDisplayName(provider);
};

/**
 * Checks if a provider exists in the list of providers
 */
export const isProviderValid = (provider: Provider, providers: Provider[]): boolean => {
  return providers.some(p => p.name === provider.name && p.type === provider.type);
};

/**
 * Finds a provider by name and type
 */
export const findProvider = (provider: Provider, providers: Provider[]): Provider | undefined => {
  return providers.find(p => p.name === provider.name && p.type === provider.type);
};

/**
 * Finds a provider by name
 */
export const findProviderByName = (providerName: string, providers: Provider[]): Provider | undefined => {
  return providers.find(provider => provider.name === providerName);
};

/**
 * Sorts providers by name alphabetically
 */
export const sortProviders = (providers: Provider[]): Provider[] => {
  return [...providers].sort((a, b) => a.name.localeCompare(b.name));
};
