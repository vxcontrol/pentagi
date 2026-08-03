import { useQuery, useSubscription } from '@apollo/client/react';
import { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react';

import type { Provider } from '@/models/provider';

import {
    ProviderCreatedDocument,
    ProviderDeletedDocument,
    ProvidersDocument,
    ProviderUpdatedDocument,
} from '@/graphql/types';
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

    const { data: providersData, refetch: refetchProviders } = useQuery(ProvidersDocument, {
        skip: !isAuthenticated(),
    });

    // The providers list is a separate root field from the settings page's own
    // query, so editing a provider there leaves this copy stale — and a stale
    // copy means the composer offers a provider name the backend no longer
    // resolves. Refetch on every provider mutation instead of waiting for a
    // page reload.
    const refetchOnProviderEvent = useCallback(() => {
        if (!isAuthenticated()) {
            return;
        }

        void refetchProviders();
    }, [isAuthenticated, refetchProviders]);

    const subscriptionSkip = !isAuthenticated();
    useSubscription(ProviderCreatedDocument, { onData: refetchOnProviderEvent, skip: subscriptionSkip });
    useSubscription(ProviderUpdatedDocument, { onData: refetchOnProviderEvent, skip: subscriptionSkip });
    useSubscription(ProviderDeletedDocument, { onData: refetchOnProviderEvent, skip: subscriptionSkip });

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
