import { createContext, useContext, useEffect, useMemo, useState } from 'react';

import type { Provider } from '@/models/Provider';

import { useProvidersQuery } from '@/graphql/types';
import { findProviderByName, sortProviders } from '@/models/Provider';

const SELECTED_PROVIDER_KEY = 'selectedProvider';

interface ProvidersContextValue {
    onChangeSelectedProvider: (provider: Provider) => void;
    providers: Provider[];
    selectedProvider: null | Provider;
}

const ProvidersContext = createContext<ProvidersContextValue | undefined>(undefined);

interface ProvidersProviderProps {
    children: React.ReactNode;
}

export const ProvidersProvider = ({ children }: ProvidersProviderProps) => {
    const { data: providersData } = useProvidersQuery();

    // Create sorted providers list to ensure consistent order
    const providers = sortProviders(providersData?.providers || []);

    // Store selected provider name instead of the provider object
    const [selectedProviderName, setSelectedProviderName] = useState<null | string>(() => {
        return localStorage.getItem(SELECTED_PROVIDER_KEY);
    });

    // Compute selected provider from providers list and selected name
    const selectedProvider = useMemo(() => {
        if (providers.length === 0) {
            return null;
        }

        // Try to find saved provider
        if (selectedProviderName) {
            const savedProvider = findProviderByName(selectedProviderName, providers);

            if (savedProvider) {
                return savedProvider;
            }
        }

        // If no saved provider or not found, return first provider
        return providers[0] ?? null;
    }, [providers, selectedProviderName]);

    // Save to localStorage when selected provider changes
    useEffect(() => {
        if (selectedProvider) {
            localStorage.setItem(SELECTED_PROVIDER_KEY, selectedProvider.name);
        }
    }, [selectedProvider]);

    // Handle provider selection changes
    const handleProviderChange = (provider: Provider) => {
        setSelectedProviderName(provider.name);
    };

    const value = {
        onChangeSelectedProvider: handleProviderChange,
        providers,
        selectedProvider,
    };

    return <ProvidersContext.Provider value={value}>{children}</ProvidersContext.Provider>;
};

export const useProviders = () => {
    const context = useContext(ProvidersContext);

    if (context === undefined) {
        throw new Error('useProviders must be used within a ProvidersProvider');
    }

    return context;
};
