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
    } = useQuery(FlowDocument, {
        errorPolicy: 'all',
        fetchPolicy: 'cache-first',
        nextFetchPolicy: 'cache-first',
        notifyOnNetworkStatusChange: true,
        skip: !flowId,
        variables: { id: flowId ?? '' },
    });

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
