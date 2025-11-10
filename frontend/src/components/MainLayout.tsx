import { useCallback, useEffect, useLayoutEffect, useRef, useState } from 'react';
import { Outlet, useLocation, useNavigate, useParams } from 'react-router-dom';

import { SidebarInset, SidebarProvider } from '@/components/ui/sidebar';
import ChatSidebar from '@/features/chat/ChatSidebar';
import {
    useDeleteFlowMutation,
    useFinishFlowMutation,
    useFlowCreatedSubscription,
    useFlowDeletedSubscription,
    useFlowsQuery,
    useFlowUpdatedSubscription,
    useProvidersQuery,
} from '@/graphql/types';
import { Log } from '@/lib/log';
import { findProviderByName, isProviderValid, sortProviders, type Provider } from '@/models/Provider';
import { useUser } from '@/providers/UserProvider';

const SELECTED_PROVIDER_KEY = 'selectedProvider';

const MainLayout = () => {
    const navigate = useNavigate();
    const location = useLocation();
    const { flowId } = useParams();

    const { data: flowsData, refetch: refetchFlows } = useFlowsQuery();
    const { data: providersData } = useProvidersQuery();

    const [selectedFlowId, setSelectedFlowId] = useState<string | null>(flowId ?? null);
    const [selectedProvider, setSelectedProvider] = useState<Provider | null>(null);

    const { isAuthenticated } = useUser();

    const needsProviderUpdateRef = useRef(true);
    const previousFlowIdRef = useRef(flowId);
    const userInitialized = useRef(false);

    // Store for API mutations
    const [deleteFlow] = useDeleteFlowMutation();
    const [finishFlow] = useFinishFlowMutation();

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

    // Check authentication and redirect if needed
    useEffect(() => {
        if (!userInitialized.current) {
            userInitialized.current = true;

            if (!isAuthenticated()) {
                // Save current path for redirect after login
                const currentPath = window.location.pathname;
                // Only save if it's not the default route
                const returnParam = currentPath !== '/flows/new' ? `?returnUrl=${encodeURIComponent(currentPath)}` : '';
                navigate(`/login${returnParam}`);
            }
        }
    }, [navigate, isAuthenticated]);

    // Handle provider selection changes
    const handleProviderChange = useCallback((provider: Provider) => {
        setSelectedProvider(provider);
        localStorage.setItem(SELECTED_PROVIDER_KEY, provider.name);
    }, []);

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

    const handleChangeSelectedFlowId = useCallback(
        (id: string) => {
            setSelectedFlowId(id);
            navigate(`/flows/${id}`);
        },
        [navigate],
    );

    const handleDeleteFlow = useCallback(
        async (id: string) => {
            try {
                // Store current state before deletion
                const wasCurrentFlow = String(selectedFlowId) === String(id);

                const result = await deleteFlow({
                    variables: { flowId: id },
                    refetchQueries: ['flows'],
                    update: (cache) => {
                        // Remove the flow from Apollo cache
                        cache.evict({ id: `Flow:${id}` });
                        cache.gc();
                    },
                });

                if (result.data?.deleteFlow === 'success') {
                    // Force refresh the flows list
                    await refetchFlows().catch((error) => {
                        Log.error('Failed to refetch flows:', error);
                    });

                    // If we deleted the currently selected flow, redirect to new flow page
                    if (wasCurrentFlow) {
                        setSelectedFlowId('new');
                        navigate('/flows/new');
                    }
                }
            } catch (error) {
                Log.error('Error deleting flow:', error);
            }
        },
        [selectedFlowId, deleteFlow, refetchFlows, navigate],
    );

    const handleFinishFlow = useCallback(
        async (id: string) => {
            try {
                await finishFlow({ variables: { flowId: id } });
            } catch {
                // ignore
            }
        },
        [finishFlow],
    );

    return (
        <SidebarProvider>
            <ChatSidebar
                providers={sortedProviders}
                selectedProvider={selectedProvider}
                onChangeSelectedProvider={handleProviderChange}
            />
            <SidebarInset>
                <Outlet
                    context={{
                        selectedProvider,
                        sortedProviders,
                        selectedFlowId,
                    }}
                />
            </SidebarInset>
        </SidebarProvider>
    );
};

export default MainLayout;
