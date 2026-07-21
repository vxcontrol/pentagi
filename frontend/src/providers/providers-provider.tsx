import { useQuery } from '@apollo/client/react';
import { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react';

import type { Provider } from '@/models/provider';

import { ProvidersDocument } from '@/graphql/types';
import { findProviderByName, sortProviders } from '@/models/provider';
import { useUser } from '@/providers/user-provider';

const SELECTED_PROVIDER_KEY = 'selectedProvider';

interface ProvidersContextValue {
    providers: Provider[];
    selectedProvider: null | Provider;
    setSelectedProvider: (provider: Provider) => void;
}

const ProvidersContext = createContext<ProvidersContextValue | undefined>(undefined);

interface ProvidersProviderProps {
    children: React.ReactNode;
}

export function ProvidersProvider({ children }: ProvidersProviderProps) {
    const { isAuthenticated } = useUser();

    const { data: providersData } = useQuery(ProvidersDocument, {
        skip: !isAuthenticated(),
    });

    const providers = useMemo(() => sortProviders(providersData?.providers || []), [providersData?.providers]);

    const [selectedProviderName, setSelectedProviderName] = useState<null | string>(() => {
        return localStorage.getItem(SELECTED_PROVIDER_KEY);
    });

    const selectedProvider = useMemo(() => {
        if (providers.length === 0) {
            return null;
        }

        if (selectedProviderName) {
            const savedProvider = findProviderByName(selectedProviderName, providers);

            if (savedProvider) {
                return savedProvider;
            }
        }

        return providers[0] ?? null;
    }, [providers, selectedProviderName]);

    useEffect(() => {
        if (selectedProvider) {
            localStorage.setItem(SELECTED_PROVIDER_KEY, selectedProvider.name);
        }
    }, [selectedProvider]);

    const setSelectedProvider = useCallback((provider: Provider) => {
        setSelectedProviderName(provider.name);
    }, []);

    const value = useMemo(
        () => ({
            providers,
            selectedProvider,
            setSelectedProvider,
        }),
        [providers, selectedProvider, setSelectedProvider],
    );

    return <ProvidersContext.Provider value={value}>{children}</ProvidersContext.Provider>;
}

export function useProviders() {
    const context = useContext(ProvidersContext);

    if (context === undefined) {
        throw new Error('useProviders must be used within a ProvidersProvider');
    }

    return context;
}
