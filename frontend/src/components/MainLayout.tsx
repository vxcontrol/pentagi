import { useEffect, useLayoutEffect, useRef, useState } from 'react';
import { Outlet, useNavigate, useParams } from 'react-router-dom';

import { SidebarInset, SidebarProvider } from '@/components/ui/sidebar';
import ChatSidebar from '@/features/chat/ChatSidebar';
import {
    useFlowCreatedSubscription,
    useFlowDeletedSubscription,
    useFlowUpdatedSubscription,
    useProvidersQuery,
} from '@/graphql/types';
import { findProviderByName, isProviderValid, sortProviders, type Provider } from '@/models/Provider';

const SELECTED_PROVIDER_KEY = 'selectedProvider';

const MainLayout = () => {
    const navigate = useNavigate();
    const { flowId } = useParams();

    const { data: providersData } = useProvidersQuery();

    const [selectedFlowId, setSelectedFlowId] = useState<string | null>(flowId ?? null);
    const [selectedProvider, setSelectedProvider] = useState<Provider | null>(null);

    const needsProviderUpdateRef = useRef(true);
    const previousFlowIdRef = useRef(flowId);

    // Global flow subscriptions - always active regardless of selected flow
    useFlowCreatedSubscription();
    useFlowDeletedSubscription();
    useFlowUpdatedSubscription();

    // Create sorted providers list to ensure consistent order
    const sortedProviders = sortProviders(providersData?.providers || []);

    // Check if provider needs initialization or update when:
    // 1. We have no selected provider yet (first login)
    // 2. Selected provider is not in available providers list
    useEffect(() => {
        // On first load or when providers change
        if (sortedProviders.length > 0) {
            if (!selectedProvider) {
                // Case 1: No provider selected yet (first login)
                needsProviderUpdateRef.current = true;
            } else {
                // Case 2: Check if selected provider is still available
                if (!isProviderValid(selectedProvider, sortedProviders)) {
                    needsProviderUpdateRef.current = true;
                }
            }
        }
    }, [sortedProviders, selectedProvider]);

    // Update provider when needed - separate from detection for ESLint compliance
    useLayoutEffect(() => {
        if (needsProviderUpdateRef.current && sortedProviders.length > 0) {
            needsProviderUpdateRef.current = false;

            // Check if we have a previously saved provider
            const savedProviderName = localStorage.getItem(SELECTED_PROVIDER_KEY);

            // Case 3: If saved provider exists and is valid, use it
            if (savedProviderName) {
                const savedProvider = findProviderByName(savedProviderName, sortedProviders);
                if (savedProvider) {
                    setSelectedProvider(savedProvider);
                    return;
                }
            }

            // If no saved provider or it's invalid, use the first available
            if (sortedProviders.length > 0 && sortedProviders[0]) {
                const firstProvider = sortedProviders[0];
                setSelectedProvider(firstProvider);
                localStorage.setItem(SELECTED_PROVIDER_KEY, firstProvider.name);
            }
        }
    }, [sortedProviders]);

    // Handle provider selection changes
    const handleProviderChange = (provider: Provider) => {
        setSelectedProvider(provider);
        localStorage.setItem(SELECTED_PROVIDER_KEY, provider.name);
    };

    // Check if selected provider is still valid whenever provider list changes
    useEffect(() => {
        // Skip initial render and only check when we have providers and a previously selected provider
        if (sortedProviders.length > 0 && selectedProvider && !isProviderValid(selectedProvider, sortedProviders)) {
            // If selected provider is no longer valid, mark for update
            needsProviderUpdateRef.current = true;
        }
    }, [sortedProviders, selectedProvider]);

    // Sync flowId with selectedFlowId when URL changes
    useLayoutEffect(() => {
        if (flowId && flowId !== previousFlowIdRef.current) {
            previousFlowIdRef.current = flowId;
            if (flowId !== selectedFlowId) {
                setSelectedFlowId(flowId);
            }
        }
    }, [flowId, selectedFlowId]);

    return (
        <SidebarProvider>
            <ChatSidebar />
            <SidebarInset>
                <Outlet
                    context={{
                        selectedProvider,
                        sortedProviders,
                        selectedFlowId,
                        onChangeSelectedProvider: handleProviderChange,
                    }}
                />
            </SidebarInset>
        </SidebarProvider>
    );
};

export default MainLayout;
