import { createContext, useCallback, useContext, useMemo, useState } from 'react';
import { useParams } from 'react-router-dom';
import { toast } from 'sonner';

import type { FlowFormValues } from '@/features/flows/FlowForm';
import type { AssistantFragmentFragment, AssistantLogFragmentFragment, FlowQuery } from '@/graphql/types';

import {
    StatusType,
    useAssistantLogsQuery,
    useAssistantsQuery,
    useCallAssistantMutation,
    useCreateAssistantMutation,
    useDeleteAssistantMutation,
    useFlowQuery,
    usePutUserInputMutation,
    useStopAssistantMutation,
    useStopFlowMutation,
} from '@/graphql/types';
import { Log } from '@/lib/log';
import { useProviders } from '@/providers/ProvidersProvider';

interface FlowContextValue {
    assistantLogs: Array<AssistantLogFragmentFragment>;
    assistants: Array<AssistantFragmentFragment>;
    callAssistant: (assistantId: string, message: string, useAgents: boolean) => Promise<void>;
    createAssistant: (message: string, useAgents: boolean) => Promise<void>;
    deleteAssistant: (assistantId: string) => Promise<void>;
    flowData: FlowQuery | undefined;
    flowId: null | string;
    initiateAssistantCreation: () => void;
    isAssistantsLoading: boolean;
    isLoading: boolean;
    refetchAssistantLogs: () => void;
    selectAssistant: (assistantId: null | string) => void;
    selectedAssistantId: null | string;
    stopAssistant: (assistantId: string) => Promise<void>;
    stopAutomationFlow: (flowIdToStop: string) => Promise<void>;
    submitAutomationMessage: (values: FlowFormValues) => Promise<void>;
}

const FlowContext = createContext<FlowContextValue | undefined>(undefined);

interface FlowProviderProps {
    children: React.ReactNode;
}

export const FlowProvider = ({ children }: FlowProviderProps) => {
    const { flowId } = useParams();
    const { selectedProvider } = useProviders();

    const [selectedAssistantIds, setSelectedAssistantIds] = useState<Record<string, null | string>>({});

    const { data: flowData, loading: isLoading } = useFlowQuery({
        errorPolicy: 'all',
        fetchPolicy: 'cache-and-network',
        nextFetchPolicy: 'cache-first',
        notifyOnNetworkStatusChange: false,
        skip: !flowId,
        variables: { id: flowId ?? '' },
    });

    const { data: assistantsData, loading: isAssistantsLoading } = useAssistantsQuery({
        fetchPolicy: 'cache-and-network',
        nextFetchPolicy: 'cache-first',
        skip: !flowId,
        variables: { flowId: flowId ?? '' },
    });

    const assistants = useMemo(() => assistantsData?.assistants ?? [], [assistantsData?.assistants]);

    const selectedAssistantId = useMemo(() => {
        if (!flowId) {
            return null;
        }

        const explicitSelection = selectedAssistantIds[flowId];

        // If there's an explicit selection (including null for "no selection")
        if (explicitSelection !== undefined) {
            // If explicitly set to null, return null
            if (explicitSelection === null) {
                return null;
            }

            // If the selected assistant still exists in the list, return it
            if (assistants.some((assistant) => assistant.id === explicitSelection)) {
                return explicitSelection;
            }
        }

        // Otherwise, auto-select the first assistant
        return assistants?.[0]?.id ?? null;
    }, [flowId, selectedAssistantIds, assistants]);

    const { data: assistantLogsData, refetch: refetchAssistantLogs } = useAssistantLogsQuery({
        fetchPolicy: 'cache-and-network',
        nextFetchPolicy: 'cache-first',
        skip: !flowId || !selectedAssistantId || selectedAssistantId === '',
        variables: { assistantId: selectedAssistantId ?? '', flowId: flowId ?? '' },
    });

    const selectAssistant = useCallback(
        (assistantId: null | string) => {
            if (!flowId) {
                return;
            }

            setSelectedAssistantIds((prev) => ({
                ...prev,
                [flowId]: assistantId,
            }));
        },
        [flowId],
    );

    const initiateAssistantCreation = useCallback(() => {
        if (!flowId) {
            return;
        }

        selectAssistant(null);
    }, [flowId, selectAssistant]);

    // Mutations
    const [putUserInput] = usePutUserInputMutation();
    const [stopFlowMutation] = useStopFlowMutation();
    const [createAssistantMutation] = useCreateAssistantMutation();
    const [callAssistantMutation] = useCallAssistantMutation();
    const [stopAssistantMutation] = useStopAssistantMutation();
    const [deleteAssistantMutation] = useDeleteAssistantMutation();

    const flowStatus = useMemo(() => flowData?.flow?.status, [flowData?.flow?.status]);
    const selectedProviderName = useMemo(() => selectedProvider?.name ?? null, [selectedProvider?.name]);

    // Helper function to refetch assistant logs if needed
    const handleRefetchAssistantLogs = useCallback(() => {
        if (selectedAssistantId && refetchAssistantLogs) {
            refetchAssistantLogs();
        }
    }, [selectedAssistantId, refetchAssistantLogs]);

    const submitAutomationMessage = useCallback(
        async (values: FlowFormValues) => {
            if (!flowId || flowStatus === StatusType.Finished) {
                return;
            }

            const { message } = values;

            try {
                await putUserInput({
                    variables: {
                        flowId,
                        input: message,
                    },
                });
            } catch (error) {
                const description =
                    error instanceof Error ? error.message : 'An error occurred while submitting message';
                toast.error('Failed to submit message', {
                    description,
                });
                Log.error('Error submitting message:', error);
            }
        },
        [flowId, flowStatus, putUserInput],
    );

    const stopAutomationFlow = useCallback(
        async (flowIdToStop: string) => {
            try {
                await stopFlowMutation({
                    variables: {
                        flowId: flowIdToStop,
                    },
                });
            } catch (error) {
                const description = error instanceof Error ? error.message : 'An error occurred while stopping flow';
                toast.error('Failed to stop flow', {
                    description,
                });
                Log.error('Error stopping flow:', error);
            }
        },
        [stopFlowMutation],
    );

    const createAssistant = useCallback(
        async (message: string, useAgents: boolean) => {
            if (!message.trim() || !selectedProviderName || !flowId) {
                return;
            }

            try {
                const { data } = await createAssistantMutation({
                    variables: {
                        flowId,
                        input: message.trim(),
                        modelProvider: selectedProviderName,
                        useAgents,
                    },
                });

                if (data?.createAssistant) {
                    const { assistant } = data.createAssistant;

                    if (assistant?.id) {
                        selectAssistant(assistant.id);
                    }
                }
            } catch (error) {
                const description =
                    error instanceof Error ? error.message : 'An error occurred while creating assistant';
                toast.error('Failed to create assistant', {
                    description,
                });
                Log.error('Error creating assistant:', error);
            }
        },
        [flowId, selectedProviderName, createAssistantMutation, selectAssistant],
    );

    const callAssistant = useCallback(
        async (assistantId: string, message: string, useAgents: boolean) => {
            if (!flowId || !assistantId || !message.trim()) {
                return;
            }

            try {
                await callAssistantMutation({
                    variables: {
                        assistantId,
                        flowId,
                        input: message.trim(),
                        useAgents,
                    },
                });

                handleRefetchAssistantLogs();
            } catch (error) {
                const description =
                    error instanceof Error ? error.message : 'An error occurred while calling assistant';
                toast.error('Failed to call assistant', {
                    description,
                });
                Log.error('Error calling assistant:', error);
            }
        },
        [flowId, callAssistantMutation, handleRefetchAssistantLogs],
    );

    const stopAssistant = useCallback(
        async (assistantId: string) => {
            if (!flowId || !assistantId) {
                return;
            }

            try {
                await stopAssistantMutation({
                    variables: {
                        assistantId,
                        flowId,
                    },
                });

                handleRefetchAssistantLogs();
            } catch (error) {
                const description =
                    error instanceof Error ? error.message : 'An error occurred while stopping assistant';
                toast.error('Failed to stop assistant', {
                    description,
                });
                Log.error('Error stopping assistant:', error);
            }
        },
        [flowId, stopAssistantMutation, handleRefetchAssistantLogs],
    );

    const deleteAssistant = useCallback(
        async (assistantId: string) => {
            if (!flowId || !assistantId) {
                return;
            }

            try {
                const wasSelected = selectedAssistantId === assistantId;

                await deleteAssistantMutation({
                    update: (cache) => {
                        cache.evict({ id: `Assistant:${assistantId}` });
                        cache.gc();
                    },
                    variables: {
                        assistantId,
                        flowId,
                    },
                });

                if (wasSelected) {
                    selectAssistant(null);
                }
            } catch (error) {
                const description =
                    error instanceof Error ? error.message : 'An error occurred while deleting assistant';
                toast.error('Failed to delete assistant', {
                    description,
                });
                Log.error('Error deleting assistant:', error);
            }
        },
        [flowId, selectedAssistantId, deleteAssistantMutation, selectAssistant],
    );

    const value = useMemo(
        () => ({
            assistantLogs: assistantLogsData?.assistantLogs ?? [],
            assistants,
            callAssistant,
            createAssistant,
            deleteAssistant,
            flowData,
            flowId: flowId ?? null,
            initiateAssistantCreation,
            isAssistantsLoading,
            isLoading,
            refetchAssistantLogs,
            selectAssistant,
            selectedAssistantId,
            stopAssistant,
            stopAutomationFlow,
            submitAutomationMessage,
        }),
        [
            assistantLogsData?.assistantLogs,
            assistants,
            callAssistant,
            createAssistant,
            deleteAssistant,
            flowData,
            flowId,
            initiateAssistantCreation,
            isAssistantsLoading,
            isLoading,
            refetchAssistantLogs,
            selectAssistant,
            selectedAssistantId,
            stopAssistant,
            stopAutomationFlow,
            submitAutomationMessage,
        ],
    );

    return <FlowContext.Provider value={value}>{children}</FlowContext.Provider>;
};

export const useFlow = () => {
    const context = useContext(FlowContext);

    if (context === undefined) {
        throw new Error('useFlow must be used within FlowProvider');
    }

    return context;
};
