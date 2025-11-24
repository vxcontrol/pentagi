import { createContext, useCallback, useContext, useMemo } from 'react';
import { toast } from 'sonner';

import type { FlowFormValues } from '@/features/flows/FlowForm';
import type { FlowOverviewFragmentFragment } from '@/graphql/types';

import {
    useCreateAssistantMutation,
    useCreateFlowMutation,
    useDeleteFlowMutation,
    useFinishFlowMutation,
    useFlowCreatedSubscription,
    useFlowDeletedSubscription,
    useFlowUpdatedSubscription,
} from '@/graphql/types';
import { Log } from '@/lib/log';

interface FlowsContextValue {
    createFlow: (values: FlowFormValues) => Promise<null | string>;
    createFlowWithAssistant: (values: FlowFormValues) => Promise<null | string>;
    deleteFlow: (flow: FlowOverviewFragmentFragment) => Promise<boolean>;
    finishFlow: (flow: FlowOverviewFragmentFragment) => Promise<boolean>;
}

const FlowsContext = createContext<FlowsContextValue | undefined>(undefined);

interface FlowsProviderProps {
    children: React.ReactNode;
}

export const FlowsProvider = ({ children }: FlowsProviderProps) => {
    // Global flow subscriptions - always active regardless of selected flow
    // These subscriptions do not require flowId and listen to ALL flows in the system
    useFlowCreatedSubscription();
    useFlowDeletedSubscription();
    useFlowUpdatedSubscription();

    // Mutations
    const [createFlowMutation] = useCreateFlowMutation();
    const [createAssistantMutation] = useCreateAssistantMutation();
    const [deleteFlowMutation] = useDeleteFlowMutation();
    const [finishFlowMutation] = useFinishFlowMutation();

    const createFlow = useCallback(
        async (values: FlowFormValues) => {
            const { message, providerName } = values;

            const input = message.trim();
            const modelProvider = providerName.trim();

            if (!input || !modelProvider) {
                return null;
            }

            try {
                const { data } = await createFlowMutation({
                    variables: {
                        input,
                        modelProvider,
                    },
                });

                if (data?.createFlow?.id) {
                    return data.createFlow.id;
                }

                return null;
            } catch (error) {
                const description = error instanceof Error ? error.message : 'An error occurred while creating flow';
                toast.error('Failed to create flow', {
                    description,
                });
                Log.error('Error creating flow:', error);

                return null;
            }
        },
        [createFlowMutation],
    );

    const createFlowWithAssistant = useCallback(
        async (values: FlowFormValues) => {
            const { message, providerName, useAgents } = values;

            const input = message.trim();
            const modelProvider = providerName.trim();

            if (!input || !modelProvider) {
                return null;
            }

            try {
                const { data } = await createAssistantMutation({
                    variables: {
                        flowId: '0',
                        input,
                        modelProvider,
                        useAgents,
                    },
                });

                if (data?.createAssistant?.flow?.id) {
                    return data.createAssistant.flow.id;
                }

                return null;
            } catch (error) {
                const description =
                    error instanceof Error ? error.message : 'An error occurred while creating assistant';
                toast.error('Failed to create assistant', {
                    description,
                });
                Log.error('Error creating assistant:', error);

                return null;
            }
        },
        [createAssistantMutation],
    );

    const deleteFlow = useCallback(
        async (flow: FlowOverviewFragmentFragment) => {
            const { id: flowId, title } = flow;

            if (!flowId) {
                return false;
            }

            const flowDescription = `${title || 'Unknown'} (ID: ${flowId})`;

            const loadingToastId = toast.loading('Deleting flow...', {
                description: flowDescription,
            });

            try {
                await deleteFlowMutation({
                    refetchQueries: ['flows'],
                    update: (cache) => {
                        // Remove the flow from Apollo cache
                        cache.evict({ id: `Flow:${flowId}` });
                        cache.gc();
                    },
                    variables: { flowId },
                });

                toast.success('Flow deleted successfully', {
                    description: flowDescription,
                    id: loadingToastId,
                });

                return true;
            } catch (error) {
                const errorMessage = error instanceof Error ? error.message : 'An error occurred while deleting flow';
                toast.error(errorMessage, {
                    description: flowDescription,
                    id: loadingToastId,
                });
                Log.error('Error deleting flow:', error);

                return false;
            }
        },
        [deleteFlowMutation],
    );

    const finishFlow = useCallback(
        async (flow: FlowOverviewFragmentFragment) => {
            const { id: flowId, title } = flow;

            if (!flowId) {
                return false;
            }

            const flowDescription = `${title || 'Unknown'} (ID: ${flowId})`;

            const loadingToastId = toast.loading('Finishing flow...', {
                description: flowDescription,
            });

            try {
                await finishFlowMutation({
                    refetchQueries: ['flows'],
                    variables: { flowId },
                });

                toast.success('Flow finished successfully', {
                    description: flowDescription,
                    id: loadingToastId,
                });

                return true;
            } catch (error) {
                const errorMessage = error instanceof Error ? error.message : 'An error occurred while finishing flow';
                toast.error(errorMessage, {
                    description: flowDescription,
                    id: loadingToastId,
                });
                Log.error('Error finishing flow:', error);

                return false;
            }
        },
        [finishFlowMutation],
    );

    const value = useMemo(
        () => ({
            createFlow,
            createFlowWithAssistant,
            deleteFlow,
            finishFlow,
        }),
        [createFlow, createFlowWithAssistant, deleteFlow, finishFlow],
    );

    return <FlowsContext.Provider value={value}>{children}</FlowsContext.Provider>;
};

export const useFlows = () => {
    const context = useContext(FlowsContext);

    if (context === undefined) {
        throw new Error('useFlows must be used within FlowsProvider');
    }

    return context;
};
