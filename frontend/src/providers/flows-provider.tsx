import { NetworkStatus } from '@apollo/client';
import { createContext, useCallback, useContext, useEffect, useMemo } from 'react';
import { toast } from 'sonner';

import type { FlowFormValues } from '@/features/flows/flow-form';
import type { FlowFragmentFragment, FlowsQuery } from '@/graphql/types';

import {
    ResultType,
    useCreateAssistantMutation,
    useCreateFlowMutation,
    useDeleteFlowMutation,
    useFinishFlowMutation,
    useFlowCreatedSubscription,
    useFlowDeletedSubscription,
    useFlowsQuery,
    useFlowUpdatedSubscription,
} from '@/graphql/types';
import { Log } from '@/lib/log';

export type Flow = FlowFragmentFragment;

interface FlowsContextValue {
    createFlow: (values: FlowFormValues) => Promise<null | string>;
    createFlowWithAssistant: (values: FlowFormValues) => Promise<null | string>;
    deleteFlow: (flow: Flow) => Promise<boolean>;
    finishFlow: (flow: Flow) => Promise<boolean>;
    flows: Array<Flow>;
    flowsData: FlowsQuery | undefined;
    flowsError: Error | undefined;
    isLoading: boolean;
}

const FlowsContext = createContext<FlowsContextValue | undefined>(undefined);

interface FlowsProviderProps {
    children: React.ReactNode;
}

export const FlowsProvider = ({ children }: FlowsProviderProps) => {
    // Query for flows list
    const {
        data: flowsData,
        error: flowsError,
        loading,
        networkStatus,
    } = useFlowsQuery({
        notifyOnNetworkStatusChange: true,
    });

    const isLoading = loading && networkStatus === NetworkStatus.loading;
    const flows = useMemo(() => flowsData?.flows ?? [], [flowsData?.flows]);

    useFlowCreatedSubscription();
    useFlowDeletedSubscription();
    useFlowUpdatedSubscription();

    // Show toast notification when flows loading error occurs
    useEffect(() => {
        if (flowsError) {
            toast.error('Error loading flows', {
                description: flowsError.message,
            });
            Log.error('Error loading flows:', flowsError);
        }
    }, [flowsError]);

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
        async (flow: Flow) => {
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
                    optimisticResponse: {
                        deleteFlow: ResultType.Success,
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
        async (flow: Flow) => {
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
                    variables: { flowId },
                });
                // Cache will be automatically updated via mutation policy and flowUpdated subscription

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
            flows,
            flowsData,
            flowsError,
            isLoading,
        }),
        [createFlow, createFlowWithAssistant, deleteFlow, finishFlow, flows, flowsData, flowsError, isLoading],
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
