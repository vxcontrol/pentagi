import { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react';
import { useParams } from 'react-router-dom';
import { toast } from 'sonner';

import type { FlowFormValues } from '@/features/flows/flow-form';
import type { AssistantFragmentFragment, AssistantLogFragmentFragment, FlowQuery } from '@/graphql/types';

import {
    ResultType,
    StatusType,
    useAgentLogAddedSubscription,
    useAssistantCreatedSubscription,
    useAssistantDeletedSubscription,
    useAssistantLogAddedSubscription,
    useAssistantLogsQuery,
    useAssistantLogUpdatedSubscription,
    useAssistantsQuery,
    useAssistantUpdatedSubscription,
    useCallAssistantMutation,
    useCreateAssistantMutation,
    useDeleteAssistantMutation,
    useFlowQuery,
    useFlowUpdatedSubscription,
    useMessageLogAddedSubscription,
    useMessageLogUpdatedSubscription,
    usePutUserInputMutation,
    useScreenshotAddedSubscription,
    useSearchLogAddedSubscription,
    useStopAssistantMutation,
    useStopFlowMutation,
    useTaskCreatedSubscription,
    useTaskUpdatedSubscription,
    useTerminalLogAddedSubscription,
    useVectorStoreLogAddedSubscription,
} from '@/graphql/types';
import { Log } from '@/lib/log';

interface FlowContextValue {
    assistantLogs: Array<AssistantLogFragmentFragment>;
    assistants: Array<AssistantFragmentFragment>;
    createAssistant: (values: FlowFormValues) => Promise<void>;
    deleteAssistant: (assistantId: string) => Promise<void>;
    flowData: FlowQuery | undefined;
    flowError: Error | undefined;
    flowId: null | string;
    flowStatus: StatusType | undefined;
    initiateAssistantCreation: () => void;
    isAssistantsLoading: boolean;
    isLoading: boolean;
    selectAssistant: (assistantId: null | string) => void;
    selectedAssistantId: null | string;
    stopAssistant: (assistantId: string) => Promise<void>;
    stopAutomation: () => Promise<void>;
    submitAssistantMessage: (assistantId: string, values: FlowFormValues) => Promise<void>;
    submitAutomationMessage: (values: FlowFormValues) => Promise<void>;
}

const FlowContext = createContext<FlowContextValue | undefined>(undefined);

interface FlowProviderProps {
    children: React.ReactNode;
}

export function FlowProvider({ children }: FlowProviderProps) {
    const { flowId } = useParams();

    const [selectedAssistantIds, setSelectedAssistantIds] = useState<Record<string, null | string>>({});

    const {
        data: flowData,
        error: flowError,
        loading: isLoading,
    } = useFlowQuery({
        errorPolicy: 'all',
        fetchPolicy: 'cache-first',
        nextFetchPolicy: 'cache-first',
        notifyOnNetworkStatusChange: true,
        skip: !flowId,
        variables: { id: flowId ?? '' },
    });

    const { data: assistantsData, loading: isAssistantsLoading } = useAssistantsQuery({
        fetchPolicy: 'cache-first',
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

        if (explicitSelection !== undefined) {
            if (explicitSelection === null) {
                return null;
            }

            if (assistants.some((assistant) => assistant.id === explicitSelection)) {
                return explicitSelection;
            }
        }

        return assistants?.[0]?.id ?? null;
    }, [flowId, selectedAssistantIds, assistants]);

    const { data: assistantLogsData } = useAssistantLogsQuery({
        fetchPolicy: 'cache-first',
        nextFetchPolicy: 'cache-first',
        skip: !flowId || !selectedAssistantId || selectedAssistantId === '',
        variables: { assistantId: selectedAssistantId ?? '', flowId: flowId ?? '' },
    });

    // Skip subscriptions until the initial flow query has loaded so cache fields exist
    // before subscription deltas arrive.
    const subscriptionVariables = useMemo(() => ({ flowId: flowId || '' }), [flowId]);
    const subscriptionSkip = !flowId || isLoading;

    useFlowUpdatedSubscription();

    useTaskCreatedSubscription({ skip: subscriptionSkip, variables: subscriptionVariables });
    useTaskUpdatedSubscription({ skip: subscriptionSkip, variables: subscriptionVariables });
    useScreenshotAddedSubscription({ skip: subscriptionSkip, variables: subscriptionVariables });
    useTerminalLogAddedSubscription({ skip: subscriptionSkip, variables: subscriptionVariables });
    useMessageLogUpdatedSubscription({ skip: subscriptionSkip, variables: subscriptionVariables });
    useMessageLogAddedSubscription({ skip: subscriptionSkip, variables: subscriptionVariables });
    useAgentLogAddedSubscription({ skip: subscriptionSkip, variables: subscriptionVariables });
    useSearchLogAddedSubscription({ skip: subscriptionSkip, variables: subscriptionVariables });
    useVectorStoreLogAddedSubscription({ skip: subscriptionSkip, variables: subscriptionVariables });

    useAssistantCreatedSubscription({ skip: subscriptionSkip, variables: subscriptionVariables });
    useAssistantUpdatedSubscription({ skip: subscriptionSkip, variables: subscriptionVariables });
    useAssistantDeletedSubscription({ skip: subscriptionSkip, variables: subscriptionVariables });
    useAssistantLogAddedSubscription({ skip: subscriptionSkip, variables: subscriptionVariables });
    useAssistantLogUpdatedSubscription({ skip: subscriptionSkip, variables: subscriptionVariables });

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

    const [putUserInput] = usePutUserInputMutation();
    const [stopFlowMutation] = useStopFlowMutation();
    const [createAssistantMutation] = useCreateAssistantMutation();
    const [submitAssistantMessageMutation] = useCallAssistantMutation();
    const [stopAssistantMutation] = useStopAssistantMutation();
    const [deleteAssistantMutation] = useDeleteAssistantMutation();

    const flowStatus = useMemo(() => flowData?.flow?.status, [flowData?.flow?.status]);

    // A single Postgres "no rows in result set" surfaces here every time a sibling
    // query/subscription retries against an invalid flow id; without a stable
    // toast id Sonner would stack 8 copies of the same message before the page
    // redirects. Surface a friendly message and drop the raw SQL detail entirely.
    useEffect(() => {
        if (flowError) {
            const raw = flowError.message ?? '';
            const isNotFound = /no rows in result set|not found/i.test(raw);
            toast.error(isNotFound ? 'Flow not found' : 'Failed to load flow', {
                description: isNotFound ? undefined : raw || undefined,
                id: 'flow-load-error',
            });
            Log.error('Error loading flow:', flowError);
        }
    }, [flowError]);

    const submitAutomationMessage = useCallback(
        async (values: FlowFormValues) => {
            if (!flowId || flowStatus === StatusType.Finished) {
                return;
            }

            const { message: input, providerName, resourceIds } = values;

            try {
                await putUserInput({
                    variables: {
                        flowId,
                        input,
                        modelProvider: providerName || undefined,
                        resourceIds: resourceIds?.length ? resourceIds : undefined,
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

    const stopAutomation = useCallback(async () => {
        if (!flowId) {
            return;
        }

        try {
            await stopFlowMutation({
                variables: {
                    flowId,
                },
            });
        } catch (error) {
            const description = error instanceof Error ? error.message : 'An error occurred while stopping flow';
            toast.error('Failed to stop flow', {
                description,
            });
            Log.error('Error stopping flow:', error);
        }
    }, [flowId, stopFlowMutation]);

    const createAssistant = useCallback(
        async (values: FlowFormValues) => {
            const { message, providerName, resourceIds, useAgents } = values;

            const input = message.trim();
            const modelProvider = providerName.trim();

            if (!input || !modelProvider || !flowId) {
                return;
            }

            try {
                const { data } = await createAssistantMutation({
                    variables: {
                        flowId,
                        input,
                        modelProvider,
                        resourceIds: resourceIds?.length ? resourceIds : undefined,
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
        [flowId, createAssistantMutation, selectAssistant],
    );

    const submitAssistantMessage = useCallback(
        async (assistantId: string, values: FlowFormValues) => {
            const { message, resourceIds, useAgents } = values;

            const input = message.trim();

            if (!flowId || !assistantId || !input) {
                return;
            }

            try {
                await submitAssistantMessageMutation({
                    variables: {
                        assistantId,
                        flowId,
                        input,
                        resourceIds: resourceIds?.length ? resourceIds : undefined,
                        useAgents,
                    },
                });
            } catch (error) {
                const description =
                    error instanceof Error ? error.message : 'An error occurred while calling assistant';
                toast.error('Failed to call assistant', {
                    description,
                });
                Log.error('Error calling assistant:', error);
            }
        },
        [flowId, submitAssistantMessageMutation],
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
            } catch (error) {
                const description =
                    error instanceof Error ? error.message : 'An error occurred while stopping assistant';
                toast.error('Failed to stop assistant', {
                    description,
                });
                Log.error('Error stopping assistant:', error);
            }
        },
        [flowId, stopAssistantMutation],
    );

    const deleteAssistant = useCallback(
        async (assistantId: string) => {
            if (!flowId || !assistantId) {
                return;
            }

            try {
                const wasSelected = selectedAssistantId === assistantId;

                await deleteAssistantMutation({
                    optimisticResponse: {
                        deleteAssistant: ResultType.Success,
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
            createAssistant,
            deleteAssistant,
            flowData,
            flowError,
            flowId: flowId ?? null,
            flowStatus,
            initiateAssistantCreation,
            isAssistantsLoading,
            isLoading,
            selectAssistant,
            selectedAssistantId,
            stopAssistant,
            stopAutomation,
            submitAssistantMessage,
            submitAutomationMessage,
        }),
        [
            assistantLogsData?.assistantLogs,
            assistants,
            createAssistant,
            deleteAssistant,
            flowData,
            flowError,
            flowId,
            flowStatus,
            initiateAssistantCreation,
            isAssistantsLoading,
            isLoading,
            selectAssistant,
            selectedAssistantId,
            stopAssistant,
            stopAutomation,
            submitAssistantMessage,
            submitAutomationMessage,
        ],
    );

    return <FlowContext.Provider value={value}>{children}</FlowContext.Provider>;
}

export function useFlow() {
    const context = useContext(FlowContext);

    if (context === undefined) {
        throw new Error('useFlow must be used within FlowProvider');
    }

    return context;
}
