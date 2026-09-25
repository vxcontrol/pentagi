import { useApolloClient, useMutation, useQuery, useSubscription } from '@apollo/client/react';
import { createContext, useCallback, useContext, useEffect, useMemo, useRef, useState } from 'react';
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
import { lastLocalWriteAt } from '@/lib/apollo';
import { isNotFoundError } from '@/lib/errors';
import { placeMissedRows } from '@/lib/list-order';
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
    createAssistant: (values: FlowFormValues) => Promise<boolean>;
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
    submitAssistantMessage: (assistantId: string, values: FlowFormValues) => Promise<boolean>;
    submitAutomationMessage: (values: FlowFormValues) => Promise<boolean>;
}

const FlowContext = createContext<FlowContextValue | undefined>(undefined);

interface FlowProviderProps {
    children: React.ReactNode;
}

const FLOW_CACHE_FIELD = 'flows';

const CATCH_UP_RETRIES = 2;

const textOf = (row: Record<string, unknown>, key: string): string =>
    typeof row[key] === 'string' ? (row[key] as string) : '';

const isStreamAhead = (cachedRow: Record<string, unknown>, row: Record<string, unknown>): boolean =>
    ['result', 'message', 'thinking'].some((key) => {
        const cached = textOf(cachedRow, key);
        const answered = textOf(row, key);

        return cached.length > answered.length && cached.startsWith(answered);
    });

const fractionOf = (moment: string): number => Number((/\.(\d+)/.exec(moment)?.[1] ?? '').padEnd(9, '0') || '0');

const isNewerThan = (cachedRow: Record<string, unknown>, row: Record<string, unknown>): boolean => {
    const cached = textOf(cachedRow, 'updatedAt');
    const answered = textOf(row, 'updatedAt');

    if (!cached || !answered) {
        return false;
    }

    const apart = Date.parse(cached) - Date.parse(answered);

    return apart === 0 ? fractionOf(cached) > fractionOf(answered) : apart > 0;
};

const isCacheAhead = (
    cachedRow: Record<string, unknown>,
    row: Record<string, unknown>,
    isWrittenAfterAsking: boolean,
): boolean => isWrittenAfterAsking || isNewerThan(cachedRow, row) || isStreamAhead(cachedRow, row);

const idOf = (row: unknown): string => String((row as { id?: unknown })?.id ?? '');

const keepRowsWrittenAfter = (
    held: null | Record<string, unknown> | undefined,
    answered: Record<string, unknown>,
    askedAt: number,
): Record<string, unknown> => {
    if (!held) {
        return answered;
    }

    return Object.fromEntries(
        Object.entries(answered).map(([field, value]) => {
            const heldValue = held[field];

            if (Array.isArray(value) && Array.isArray(heldValue)) {
                const fresher = new Map(
                    heldValue
                        .filter((row) => lastLocalWriteAt(field, idOf(row)) > askedAt)
                        .map((row) => [idOf(row), row]),
                );

                return [field, fresher.size ? value.map((row) => fresher.get(idOf(row)) ?? row) : value];
            }

            const isFlowFresher =
                field === 'flow' && !!heldValue && !!value && lastLocalWriteAt(FLOW_CACHE_FIELD, idOf(value)) > askedAt;

            return [field, isFlowFresher ? heldValue : value];
        }),
    );
};

export function FlowProvider({ children }: FlowProviderProps) {
    const { flowId } = useParams();
    const client = useApolloClient();

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

    const { data: assistantLogsData, loading: isAssistantLogsLoading } = useQuery(AssistantLogsDocument, {
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

    const caughtUpForRef = useRef<null | string>(null);
    const caughtUpLogsForRef = useRef<null | string>(null);
    const catchUpFailuresRef = useRef(0);
    const catchUpFailuresForRef = useRef<null | string>(null);
    const catchUpBaselinesRef = useRef(new Map<string, Set<string>>());
    const catchUpLogFailuresRef = useRef(0);
    const catchUpLogFailuresForRef = useRef<null | string>(null);
    const [catchUpRetry, setCatchUpRetry] = useState(0);
    const [catchUpLogRetry, setCatchUpLogRetry] = useState(0);

    const catchUp = useCallback(
        async (query: Parameters<typeof client.query>[0]['query'], variables: Record<string, unknown>) => {
            const askedAt = Date.now();
            const { data, error } = await client.query<Record<string, unknown>>({
                context: { queryDeduplication: false },
                // The watched query tolerates a partial error; without the same policy here a
                // partial response throws inside Apollo and the catch-up writes nothing.
                errorPolicy: 'all',
                fetchPolicy: 'no-cache',
                query,
                variables,
            });

            if (data) {
                client.cache.updateQuery<Record<string, unknown>>({ query, variables }, (held) =>
                    keepRowsWrittenAfter(held, data, askedAt),
                );
            }

            return { error };
        },
        [client],
    );

    // Unlike the flow document's union-merged lists, these two replace theirs on write, so the
    // answer is kept out of the cache and merged in by hand.
    const catchUpList = useCallback(
        async (
            query: Parameters<typeof client.query>[0]['query'],
            variables: Record<string, unknown>,
            field: 'assistantLogs' | 'assistants',
        ) => {
            // Only rows that were not here when the question was asked may be kept: the list query
            // hides a deleted assistant by omitting it, exactly as it omits one that arrived late.
            const target = `${field}:${JSON.stringify(variables)}`;
            const beforeRows = client.cache.readQuery<Record<string, unknown>>({ query, variables })?.[field];
            const before =
                catchUpBaselinesRef.current.get(target) ??
                new Set(Array.isArray(beforeRows) ? beforeRows.map((row: { id: string }) => String(row.id)) : []);

            catchUpBaselinesRef.current.set(target, before);

            const askedAt = Date.now();
            const { data, error } = await client.query<Record<string, unknown>>({
                context: { queryDeduplication: false },
                errorPolicy: 'all',
                fetchPolicy: 'no-cache',
                query,
                variables,
            });

            const answered = data?.[field];

            if (error || !Array.isArray(answered)) {
                return { error: error ?? new Error(`the ${field} catch-up answered nothing`) };
            }

            const cached = client.cache.readQuery<Record<string, unknown>>({ query, variables });
            const held = cached?.[field];
            const arrived = new Set(answered.map((row: { id: string }) => String(row.id)));
            const heldRows = Array.isArray(held) ? held : [];
            const missed = heldRows.filter(
                (row: { id: string }) => !arrived.has(String(row.id)) && !before.has(String(row.id)),
            );

            const heldById = new Map(
                heldRows.map((row: { id: string }) => [String(row.id), row as Record<string, unknown>]),
            );

            // The answer is the newer read except where a subscription wrote the row while it was in
            // flight. Keeping the cache unconditionally would drop the very updates this request
            // exists to recover: the watched query is cache-first and never asks again.
            const settled = answered.flatMap((row: { id: string }) => {
                const id = String(row.id);
                const cachedRow = heldById.get(id);

                if (!cachedRow) {
                    // Gone from the cache but known before the question: deleted while it was in
                    // flight, and the answer predates that.
                    return before.has(id) ? [] : [row];
                }

                return [isCacheAhead(cachedRow, row, lastLocalWriteAt(field, id) > askedAt) ? cachedRow : row];
            });

            client.writeQuery({
                data: {
                    [field]: placeMissedRows(field, settled, missed),
                },
                query,
                variables,
            });

            catchUpBaselinesRef.current.delete(target);

            return { error: undefined };
        },
        [client],
    );

    useEffect(() => {
        if (subscriptionSkip || !flowId || isAssistantsLoading) {
            caughtUpForRef.current = null;
            catchUpFailuresRef.current = 0;
            catchUpFailuresForRef.current = null;
            catchUpBaselinesRef.current.clear();

            return;
        }

        if (catchUpFailuresForRef.current !== flowId) {
            catchUpFailuresForRef.current = flowId;
            catchUpFailuresRef.current = 0;
        }

        if (caughtUpForRef.current === flowId) {
            return;
        }

        caughtUpForRef.current = flowId;

        // Under `errorPolicy:'all'` Apollo resolves even a network failure into a result carrying
        // the error, so a `.catch` here never runs.
        let isWaveCounted = false;

        const giveUp = (error: unknown) => {
            if (caughtUpForRef.current === flowId) {
                caughtUpForRef.current = null;
            }

            Log.error('Flow catch-up after subscribing failed:', error);

            if (isWaveCounted) {
                return;
            }

            isWaveCounted = true;

            if (catchUpFailuresRef.current < CATCH_UP_RETRIES) {
                catchUpFailuresRef.current += 1;
                setCatchUpRetry((attempt) => attempt + 1);
            }
        };

        const settle = ({ error }: { error?: unknown }) => {
            if (error) {
                giveUp(error);
            }
        };

        catchUp(FlowDocument, { id: flowId }).then(settle, giveUp);
        // An assistant created in the same window is missing from the list until a reload: the
        // list is read cache-first and `assistantCreated` only inserts into what is already there.
        catchUpList(AssistantsDocument, { flowId }, 'assistants').then(settle, giveUp);
    }, [catchUp, catchUpList, catchUpRetry, flowId, isAssistantsLoading, subscriptionSkip]);

    // Assistant log rows are append-only for the cache too, so one missed `assistantLogAdded`
    // leaves a hole in the conversation that no later frame fills.
    useEffect(() => {
        if (subscriptionSkip || !flowId || !selectedAssistantId || isAssistantLogsLoading) {
            caughtUpLogsForRef.current = null;
            catchUpLogFailuresRef.current = 0;
            catchUpLogFailuresForRef.current = null;

            return;
        }

        const conversation = `${flowId}:${selectedAssistantId}`;

        if (catchUpLogFailuresForRef.current !== conversation) {
            catchUpLogFailuresForRef.current = conversation;
            catchUpLogFailuresRef.current = 0;
        }

        if (caughtUpLogsForRef.current === conversation) {
            return;
        }

        caughtUpLogsForRef.current = conversation;

        const retryLogs = (error: unknown) => {
            if (caughtUpLogsForRef.current === conversation) {
                caughtUpLogsForRef.current = null;
            }

            Log.error('Assistant log catch-up after subscribing failed:', error);

            if (catchUpLogFailuresRef.current < CATCH_UP_RETRIES) {
                catchUpLogFailuresRef.current += 1;
                setCatchUpLogRetry((attempt) => attempt + 1);
            }
        };

        catchUpList(AssistantLogsDocument, { assistantId: selectedAssistantId, flowId }, 'assistantLogs').then(
            ({ error }) => {
                if (error) {
                    retryLogs(error);
                }
            },
            (error: unknown) => {
                retryLogs(error);
            },
        );
    }, [catchUpList, catchUpLogRetry, flowId, isAssistantLogsLoading, selectedAssistantId, subscriptionSkip]);

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
                return false;
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

                return true;
            } catch (error) {
                const description =
                    error instanceof Error ? error.message : 'An error occurred while submitting message';
                toast.error('Failed to submit message', {
                    description,
                });
                Log.error('Error submitting message:', error);

                return false;
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
                return false;
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

                return true;
            } catch (error) {
                const description =
                    error instanceof Error ? error.message : 'An error occurred while creating assistant';
                toast.error('Failed to create assistant', {
                    description,
                });
                Log.error('Error creating assistant:', error);

                return false;
            }
        },
        [flowId, createAssistantMutation, selectAssistant],
    );

    const submitAssistantMessage = useCallback(
        async (assistantId: string, values: FlowFormValues) => {
            const { message, resourceIds, useAgents } = values;

            const input = message.trim();

            if (!flowId || !assistantId || !input) {
                return false;
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

                return true;
            } catch (error) {
                const description =
                    error instanceof Error ? error.message : 'An error occurred while calling assistant';
                toast.error('Failed to call assistant', {
                    description,
                });
                Log.error('Error calling assistant:', error);

                return false;
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
