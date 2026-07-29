import { useMutation, useQuery, useSubscription } from '@apollo/client/react';
import { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react';
import { useParams } from 'react-router-dom';
import { toast } from 'sonner';

import type { FlowFormValues } from '@/features/flows/flow-form';
import type { AssistantFragmentFragment, AssistantLogFragmentFragment, FlowQuery } from '@/graphql/types';

import {
    AgentLogAddedDocument,
    AssistantCreatedDocument,
    AssistantDeletedDocument,
    AssistantLogAddedDocument,
    AssistantLogsDocument,
    AssistantLogUpdatedDocument,
    AssistantsDocument,
    AssistantUpdatedDocument,
    CallAssistantDocument,
    CreateAssistantDocument,
    DeleteAssistantDocument,
    FlowDocument,
    FlowUpdatedDocument,
    MessageLogAddedDocument,
    MessageLogUpdatedDocument,
    PutUserInputDocument,
    ResultType,
    ScreenshotAddedDocument,
    SearchLogAddedDocument,
    StatusType,
    StopAssistantDocument,
    StopFlowDocument,
    TaskCreatedDocument,
    TaskUpdatedDocument,
    TerminalLogAddedDocument,
    VectorStoreLogAddedDocument,
} from '@/graphql/types';
import { isNotFoundError } from '@/lib/errors';
import { Log } from '@/lib/log';

/**
 * Under `errorPolicy:'all'` a partial not-found error surfaces alongside a flow that loaded fine, so
 * the not-found disjunct gates on `!flowData?.flow`. Without the gate that partial error redirects
 * the user off a flow that rendered correctly.
 */
export const deriveFlowMissing = (
    flowData: null | undefined | { flow: unknown },
    flowError: undefined | { message: string },
): boolean =>
    Boolean(flowData && !flowData.flow) || Boolean(flowError && !flowData?.flow && isNotFoundError(flowError));

interface FlowContextValue {
    assistantLogs: Array<AssistantLogFragmentFragment>;
    assistants: Array<AssistantFragmentFragment>;
    createAssistant: (values: FlowFormValues) => Promise<void>;
    deleteAssistant: (assistantId: string) => Promise<void>;
    flowData: FlowQuery | undefined;
    flowId: null | string;
    flowLoadError: Error | undefined;
    flowStatus: StatusType | undefined;
    initiateAssistantCreation: () => void;
    isAssistantsLoading: boolean;
    isFlowMissing: boolean;
    isLoading: boolean;
    refetchFlow: () => void;
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
        loading,
        refetch: refetchFlow,
    } = useQuery(FlowDocument, {
        errorPolicy: 'all',
        fetchPolicy: 'cache-first',
        nextFetchPolicy: 'cache-first',
        notifyOnNetworkStatusChange: true,
        skip: !flowId,
        variables: { id: flowId ?? '' },
    });

    // Also gates `subscriptionSkip` below: raising it on a refetch that still holds the flow
    // would tear down 14 live subscriptions mid-flight.
    const isLoading = loading && !flowData?.flow;

    // A real load failure that left nothing to show (cold cache + backend error on a
    // deep link), as opposed to a genuine not-found. The detail page renders this as an
    // in-page ErrorState + Retry instead of silently bouncing to the list.
    const flowLoadError = flowError && !flowData?.flow && !isNotFoundError(flowError) ? flowError : undefined;

    const isFlowMissing = deriveFlowMissing(flowData, flowError);

    const { data: assistantsData, loading: isAssistantsLoading } = useQuery(AssistantsDocument, {
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

    const { data: assistantLogsData } = useQuery(AssistantLogsDocument, {
        fetchPolicy: 'cache-first',
        nextFetchPolicy: 'cache-first',
        skip: !flowId || !selectedAssistantId || selectedAssistantId === '',
        variables: { assistantId: selectedAssistantId ?? '', flowId: flowId ?? '' },
    });

    // Skip subscriptions until the initial flow query has loaded so cache fields exist
    // before subscription deltas arrive.
    const subscriptionVariables = useMemo(() => ({ flowId: flowId || '' }), [flowId]);
    const subscriptionSkip = !flowId || isLoading;

    useSubscription(FlowUpdatedDocument);

    useSubscription(TaskCreatedDocument, { skip: subscriptionSkip, variables: subscriptionVariables });
    useSubscription(TaskUpdatedDocument, { skip: subscriptionSkip, variables: subscriptionVariables });
    useSubscription(ScreenshotAddedDocument, { skip: subscriptionSkip, variables: subscriptionVariables });
    useSubscription(TerminalLogAddedDocument, { skip: subscriptionSkip, variables: subscriptionVariables });
    useSubscription(MessageLogUpdatedDocument, { skip: subscriptionSkip, variables: subscriptionVariables });
    useSubscription(MessageLogAddedDocument, { skip: subscriptionSkip, variables: subscriptionVariables });
    useSubscription(AgentLogAddedDocument, { skip: subscriptionSkip, variables: subscriptionVariables });
    useSubscription(SearchLogAddedDocument, { skip: subscriptionSkip, variables: subscriptionVariables });
    useSubscription(VectorStoreLogAddedDocument, { skip: subscriptionSkip, variables: subscriptionVariables });

    useSubscription(AssistantCreatedDocument, { skip: subscriptionSkip, variables: subscriptionVariables });
    useSubscription(AssistantUpdatedDocument, { skip: subscriptionSkip, variables: subscriptionVariables });
    useSubscription(AssistantDeletedDocument, { skip: subscriptionSkip, variables: subscriptionVariables });
    useSubscription(AssistantLogAddedDocument, { skip: subscriptionSkip, variables: subscriptionVariables });
    useSubscription(AssistantLogUpdatedDocument, { skip: subscriptionSkip, variables: subscriptionVariables });

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

    const [putUserInput] = useMutation(PutUserInputDocument);
    const [stopFlowMutation] = useMutation(StopFlowDocument);
    const [createAssistantMutation] = useMutation(CreateAssistantDocument);
    const [submitAssistantMessageMutation] = useMutation(CallAssistantDocument);
    const [stopAssistantMutation] = useMutation(StopAssistantDocument);
    const [deleteAssistantMutation] = useMutation(DeleteAssistantDocument);

    const flowStatus = useMemo(() => flowData?.flow?.status, [flowData?.flow?.status]);

    // errorPolicy:'all' surfaces a partial error while the flow loaded, so gate on
    // `!flow` or a partial failure toasts over a flow that rendered fine. A real load
    // failure is surfaced in-page (ErrorState via flowLoadError); only the not-found
    // redirect needs a toast to explain the bounce to the list. The stable id keeps the
    // invalid-id "no rows" retries from stacking.
    useEffect(() => {
        if (!flowError || flowData?.flow) {
            return;
        }

        if (isNotFoundError(flowError)) {
            toast.error('Flow not found', { id: 'flow-load-error' });
        }

        Log.error('Error loading flow:', flowError);
    }, [flowError, flowData]);

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
            flowId: flowId ?? null,
            flowLoadError,
            flowStatus,
            initiateAssistantCreation,
            isAssistantsLoading,
            isFlowMissing,
            isLoading,
            refetchFlow,
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
            flowId,
            flowLoadError,
            flowStatus,
            initiateAssistantCreation,
            isAssistantsLoading,
            isFlowMissing,
            isLoading,
            refetchFlow,
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
