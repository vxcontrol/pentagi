import type { OperationDefinitionNode } from 'graphql';

import { ApolloClient, ApolloLink, Observable } from '@apollo/client';
import { ApolloProvider } from '@apollo/client/react';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes, useNavigate } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';

import { AssistantsDocument, ProviderType, StatusType } from '@/graphql/types';
import { createCache, createSubscriptionCacheLink, watchQueryDefaults } from '@/lib/apollo';
import { Log } from '@/lib/log';

import { FlowProvider, useFlow } from './flow-provider';

vi.mock('sonner', () => ({ toast: { error: vi.fn(), info: vi.fn(), success: vi.fn() } }));

const terminalLog = (id: string) => ({
    __typename: 'TerminalLog',
    createdAt: '2026-01-01T00:00:00Z',
    flowId: '1',
    id,
    subtaskId: null,
    taskId: null,
    terminal: 1,
    text: `row ${id}`,
    type: 'stdout',
});

const flowPayload = (logs: string[]) => ({
    agentLogs: [],
    flow: {
        __typename: 'Flow',
        createdAt: '2026-01-01T00:00:00Z',
        id: '1',
        provider: { __typename: 'Provider', name: 'mock', type: 'custom' },
        status: 'running',
        terminals: [],
        title: 'Flow',
        updatedAt: '2026-01-01T00:00:00Z',
        userId: '1',
    },
    messageLogs: [],
    screenshots: [],
    searchLogs: [],
    tasks: [],
    terminalLogs: logs.map(terminalLog),
    vectorStoreLogs: [],
});

const ANSWERS = [flowPayload(['1']), flowPayload(['1', '2'])];

const assistant = {
    __typename: 'Assistant',
    createdAt: '2026-01-01T00:00:00Z',
    flowId: '1',
    id: '7',
    provider: { __typename: 'Provider', name: 'mock', type: ProviderType.Custom },
    status: StatusType.Waiting,
    title: 'Assistant',
    updatedAt: '2026-01-01T00:00:00Z',
    useAgents: false,
};

const assistantLogRow = (id: string) => ({
    __typename: 'AssistantLog',
    appendPart: false,
    assistantId: '7',
    createdAt: '2026-01-01T00:00:00Z',
    flowId: '1',
    id,
    message: `reply ${id}`,
    result: `reply ${id}`,
    resultFormat: 'plain',
    thinking: null,
    type: 'answer',
});

const otherAssistant = { ...assistant, createdAt: '2026-01-02T00:00:00Z', id: '9' };

const buildClient = ({
    isAssistantCreatedMidFlight = false,
    isAssistantListAnsweringLast = false,
    isCacheFractionShorter = false,
    isCatchUpFailing = false,
    isConversationAnsweringLast = false,
    isDeletedAssistantCached = false,
    isEveryCatchUpFailing = false,
    isFlowStatusArrivingMidFlight = false,
    isListCatchUpFailing = false,
    isLogCatchUpFailing = false,
    isReplyRevisedShorter = false,
    isReplyStreamedBeforeAsking = false,
    isReplyStreaming = false,
    isRowArrivingBetweenTries = false,
    isRowArrivingMidFlight = false,
    isStatusFresherInCache = false,
    isStatusMissedInWindow = false,
} = {}) => {
    let emitStreamedFrame: (() => void) | null = null;
    let emitFlowFrame: (() => void) | null = null;
    const assistantReadsByFlow = new Map<string, number>();

    const flowQueries: Array<{ deduplication: unknown; id: unknown }> = [];
    const order: string[] = [];
    const serverAssistants = [assistant];
    const serverLogRows = ['1'];
    let assistantQueries = 0;
    let logQueries = 0;

    const link = new ApolloLink((operation) => {
        const isSubscription = operation.query.definitions.some(
            (definition): definition is OperationDefinitionNode =>
                definition.kind === 'OperationDefinition' && definition.operation === 'subscription',
        );

        order.push(isSubscription ? `subscription:${operation.operationName}` : `query:${operation.operationName}`);

        if (isSubscription) {
            if (isFlowStatusArrivingMidFlight && operation.operationName === 'flowUpdated') {
                return new Observable<{ data: Record<string, unknown> }>((observer) => {
                    emitFlowFrame = () =>
                        observer.next({
                            data: {
                                flowUpdated: {
                                    ...flowPayload(['1']).flow,
                                    status: 'finished',
                                    updatedAt: '2026-01-02T00:00:00Z',
                                },
                            },
                        });
                });
            }

            if (isRowArrivingBetweenTries && operation.operationName === 'assistantCreated') {
                return new Observable<{ data: Record<string, unknown> }>((observer) => {
                    emitStreamedFrame = () => observer.next({ data: { assistantCreated: otherAssistant } });
                });
            }

            if (isAssistantCreatedMidFlight && operation.operationName === 'assistantCreated') {
                return new Observable<{ data: Record<string, unknown> }>((observer) => {
                    setTimeout(() => observer.next({ data: { assistantCreated: otherAssistant } }), 5);
                });
            }

            if (isReplyRevisedShorter && operation.operationName === 'assistantLogUpdated') {
                return new Observable<{ data: Record<string, unknown> }>((observer) => {
                    emitStreamedFrame = () =>
                        observer.next({
                            data: {
                                assistantLogUpdated: {
                                    ...assistantLogRow('1'),
                                    message: 'a long first draft',
                                    result: 'a long first draft',
                                },
                            },
                        });
                });
            }

            if (isReplyStreamedBeforeAsking && operation.operationName === 'assistantLogUpdated') {
                return new Observable<{ data: Record<string, unknown> }>((observer) => {
                    emitStreamedFrame = () =>
                        observer.next({
                            data: {
                                assistantLogUpdated: {
                                    ...assistantLogRow('1'),
                                    message: 'the whole reply',
                                    result: 'the whole reply',
                                },
                            },
                        });
                });
            }

            if (isReplyStreaming && operation.operationName === 'assistantLogUpdated') {
                return new Observable<{ data: Record<string, unknown> }>((observer) => {
                    setTimeout(
                        () =>
                            observer.next({
                                data: {
                                    assistantLogUpdated: {
                                        ...assistantLogRow('1'),
                                        message: 'the whole reply',
                                        result: 'the whole reply',
                                    },
                                },
                            }),
                        10,
                    );
                });
            }

            if (isCacheFractionShorter && operation.operationName === 'assistantUpdated') {
                return new Observable<{ data: Record<string, unknown> }>((observer) => {
                    emitStreamedFrame = () =>
                        observer.next({
                            data: {
                                assistantUpdated: {
                                    ...assistant,
                                    status: StatusType.Waiting,
                                    updatedAt: '2026-01-01T00:00:00.1Z',
                                },
                            },
                        });
                });
            }

            if (isStatusFresherInCache && operation.operationName === 'assistantUpdated') {
                return new Observable<{ data: Record<string, unknown> }>((observer) => {
                    emitStreamedFrame = () =>
                        observer.next({
                            data: {
                                assistantUpdated: {
                                    ...assistant,
                                    status: StatusType.Finished,
                                    updatedAt: '2026-01-02T00:00:00Z',
                                },
                            },
                        });
                });
            }

            if (isAssistantListAnsweringLast && operation.operationName === 'assistantCreated') {
                return new Observable<{ data: Record<string, unknown> }>((observer) => {
                    setTimeout(() => {
                        serverAssistants.push(otherAssistant);
                        observer.next({ data: { assistantCreated: otherAssistant } });
                    }, 15);
                });
            }

            if (isConversationAnsweringLast && operation.operationName === 'assistantLogAdded') {
                return new Observable<{ data: Record<string, unknown> }>((observer) => {
                    setTimeout(() => {
                        serverLogRows.push('2');
                        observer.next({ data: { assistantLogAdded: assistantLogRow('2') } });
                    }, 15);
                });
            }

            if (isRowArrivingMidFlight && operation.operationName === 'assistantLogAdded') {
                // The frame lands while the catch-up below is still in flight, which is the whole
                // point: its answer was taken before this row existed.
                return new Observable<{ data: Record<string, unknown> }>((observer) => {
                    setTimeout(() => observer.next({ data: { assistantLogAdded: assistantLogRow('2') } }), 0);
                });
            }

            return new Observable(() => undefined);
        }

        if (operation.operationName === 'assistants') {
            if (isAssistantCreatedMidFlight) {
                const isCatchUp = (assistantQueries += 1) > 1;

                return new Observable<{ data: Record<string, unknown> }>((observer) => {
                    setTimeout(
                        () => {
                            observer.next({ data: { assistants: [assistant] } });
                            observer.complete();
                        },
                        isCatchUp ? 25 : 0,
                    );
                });
            }

            if (isRowArrivingBetweenTries) {
                const attempt = (assistantQueries += 1);

                return new Observable<{ data: Record<string, unknown> }>((observer) => {
                    if (attempt === 2) {
                        emitStreamedFrame?.();
                        observer.error(new Error('the list went away'));

                        return;
                    }

                    observer.next({ data: { assistants: [assistant] } });
                    observer.complete();
                });
            }

            if (isEveryCatchUpFailing) {
                const flow = String(operation.variables.flowId);
                const reads = (assistantReadsByFlow.get(flow) ?? 0) + 1;

                assistantReadsByFlow.set(flow, reads);

                return new Observable<{ data: Record<string, unknown> }>((observer) => {
                    if (reads > 1) {
                        observer.error(new Error('the list went away'));

                        return;
                    }

                    observer.next({ data: { assistants: [assistant] } });
                    observer.complete();
                });
            }

            if (isListCatchUpFailing) {
                const isCatchUp = (assistantQueries += 1) > 1;

                return new Observable<{ data: Record<string, unknown> }>((observer) => {
                    if (isCatchUp && assistantQueries < 3) {
                        observer.error(new Error('the list went away'));

                        return;
                    }

                    observer.next({ data: { assistants: [assistant] } });
                    observer.complete();
                });
            }

            if (isCacheFractionShorter) {
                const isCatchUp = (assistantQueries += 1) > 1;

                return new Observable<{ data: Record<string, unknown> }>((observer) => {
                    setTimeout(
                        () => {
                            observer.next({
                                data: {
                                    assistants: [
                                        isCatchUp
                                            ? {
                                                  ...assistant,
                                                  status: StatusType.Finished,
                                                  updatedAt: '2026-01-01T00:00:00.123Z',
                                              }
                                            : assistant,
                                    ],
                                },
                            });
                            observer.complete();

                            if (!isCatchUp) {
                                emitStreamedFrame?.();
                            }
                        },
                        isCatchUp ? 20 : 0,
                    );
                });
            }

            if (isStatusFresherInCache) {
                const isCatchUp = (assistantQueries += 1) > 1;

                return new Observable<{ data: Record<string, unknown> }>((observer) => {
                    setTimeout(
                        () => {
                            observer.next({
                                data: { assistants: isCatchUp ? [otherAssistant, assistant] : [assistant] },
                            });
                            observer.complete();

                            if (!isCatchUp) {
                                emitStreamedFrame?.();
                            }
                        },
                        isCatchUp ? 20 : 0,
                    );
                });
            }

            if (isStatusMissedInWindow) {
                // The watched query saw it waiting; the frame that turned it finished fell in the
                // gap before the subscriptions attached, so only the catch-up can still bring it.
                const isCatchUp = (assistantQueries += 1) > 1;

                return new Observable<{ data: Record<string, unknown> }>((observer) => {
                    observer.next({
                        data: {
                            assistants: [isCatchUp ? { ...assistant, status: StatusType.Finished } : assistant],
                        },
                    });
                    observer.complete();
                });
            }

            if (isAssistantListAnsweringLast) {
                const asked = [...serverAssistants];
                const isWatched = (assistantQueries += 1) === 1;

                return new Observable<{ data: Record<string, unknown> }>((observer) => {
                    setTimeout(
                        () => {
                            observer.next({ data: { assistants: asked } });
                            observer.complete();
                        },
                        isWatched ? 40 : 0,
                    );
                });
            }

            return new Observable<{ data: Record<string, unknown> }>((observer) => {
                // The list query hides a deleted assistant, so the answer simply omits it.
                observer.next({ data: { assistants: isDeletedAssistantCached ? [otherAssistant] : [assistant] } });
                observer.complete();
            });
        }

        if (operation.operationName === 'assistantLogs') {
            if (isLogCatchUpFailing) {
                const isCatchUp = (logQueries += 1) > 1;

                return new Observable<{ data: Record<string, unknown> }>((observer) => {
                    if (isCatchUp && logQueries < 3) {
                        observer.error(new Error('the conversation went away'));

                        return;
                    }

                    observer.next({ data: { assistantLogs: [assistantLogRow('1'), assistantLogRow('2')] } });
                    observer.complete();
                });
            }

            if (isReplyRevisedShorter) {
                const revised = { ...assistantLogRow('1'), message: 'rewritten', result: 'rewritten' };
                const isCatchUp = (logQueries += 1) > 1;

                return new Observable<{ data: Record<string, unknown> }>((observer) => {
                    setTimeout(
                        () => {
                            observer.next({
                                data: { assistantLogs: isCatchUp ? [revised, assistantLogRow('2')] : [revised] },
                            });
                            observer.complete();

                            if (!isCatchUp) {
                                emitStreamedFrame?.();
                            }
                        },
                        isCatchUp ? 20 : 0,
                    );
                });
            }

            if (isReplyStreamedBeforeAsking) {
                const truncated = { ...assistantLogRow('1'), message: 'the whole', result: 'the whole' };
                const isCatchUp = (logQueries += 1) > 1;

                return new Observable<{ data: Record<string, unknown> }>((observer) => {
                    setTimeout(
                        () => {
                            observer.next({
                                data: {
                                    assistantLogs: isCatchUp ? [truncated, assistantLogRow('2')] : [truncated],
                                },
                            });
                            observer.complete();

                            if (!isCatchUp) {
                                emitStreamedFrame?.();
                            }
                        },
                        isCatchUp ? 20 : 0,
                    );
                });
            }

            if (isReplyStreaming) {
                const truncated = { ...assistantLogRow('1'), message: 'the whole', result: 'the whole' };
                const isCatchUp = (logQueries += 1) > 1;

                return new Observable<{ data: Record<string, unknown> }>((observer) => {
                    setTimeout(
                        () => {
                            observer.next({ data: { assistantLogs: [truncated] } });
                            observer.complete();
                        },
                        isCatchUp ? 30 : 0,
                    );
                });
            }

            if (isConversationAnsweringLast) {
                // A snapshot is taken when the question is asked and delivered later, which is what
                // makes the earlier request carry the older answer.
                const asked = serverLogRows.map(assistantLogRow);
                const isWatched = (logQueries += 1) === 1;

                return new Observable<{ data: Record<string, unknown> }>((observer) => {
                    setTimeout(
                        () => {
                            observer.next({ data: { assistantLogs: asked } });
                            observer.complete();
                        },
                        isWatched ? 40 : 0,
                    );
                });
            }

            const rows = isRowArrivingMidFlight ? [assistantLogRow('1')] : [];
            // The catch-up is the second read of this list. Answer it late, so the subscription
            // frame lands first and the answer carries the snapshot taken before that row existed.
            const isCatchUp = isRowArrivingMidFlight && (logQueries += 1) > 1;

            return new Observable<{ data: Record<string, unknown> }>((observer) => {
                const deliver = () => {
                    observer.next({ data: { assistantLogs: rows } });
                    observer.complete();
                };

                if (isCatchUp) {
                    setTimeout(deliver, 20);
                } else {
                    deliver();
                }
            });
        }

        if (operation.operationName !== 'flow') {
            return new Observable<{ data: Record<string, unknown> }>((observer) => {
                observer.next({ data: {} });
                observer.complete();
            });
        }

        const answer = ANSWERS[Math.min(flowQueries.length, ANSWERS.length - 1)] ?? ANSWERS[0];
        const isCatchUp = flowQueries.length > 0;
        const isRepeatReadOfThisFlow = flowQueries.some((query) => query.id === operation.variables.id);

        flowQueries.push({
            deduplication: operation.getContext().queryDeduplication,
            id: operation.variables.id,
        });

        if (isEveryCatchUpFailing && isRepeatReadOfThisFlow) {
            return new Observable<{ data: Record<string, unknown> }>((observer) => {
                observer.error(new Error('the network went away'));
            });
        }

        if (isCatchUpFailing && isCatchUp && flowQueries.length < 3) {
            return new Observable<{ data: Record<string, unknown> }>((observer) => {
                observer.error(new Error('the network went away'));
            });
        }

        if (isFlowStatusArrivingMidFlight && isCatchUp) {
            return new Observable<{ data: Record<string, unknown> }>((observer) => {
                setTimeout(() => emitFlowFrame?.(), 5);
                setTimeout(() => {
                    observer.next({ data: answer as Record<string, unknown> });
                    observer.complete();
                }, 20);
            });
        }

        return new Observable<{ data: Record<string, unknown> }>((observer) => {
            observer.next({ data: answer as Record<string, unknown> });
            observer.complete();
        });
    });

    const cache = createCache();

    return {
        client: new ApolloClient({
            cache,
            defaultOptions: { watchQuery: watchQueryDefaults },
            // The app puts subscription frames into the cache through this link; without it a
            // frame is delivered to the hook and forgotten, which is not the app's behaviour.
            link: ApolloLink.from([createSubscriptionCacheLink(cache), link]),
        }),
        flowQueries,
        order,
    };
};

function Probe() {
    const { assistantLogs, assistants, flowData, flowId, selectedAssistantId } = useFlow();

    return (
        <>
            <span data-slot="probe-rows">{flowData?.terminalLogs?.length ?? 0}</span>
            <span data-slot="probe-flow-id">{flowId}</span>
            <span data-slot="probe-logs">{assistantLogs.map(({ id }) => id).join(',')}</span>
            <span data-slot="probe-log-text">{assistantLogs.map(({ message }) => message).join('|')}</span>
            <span data-slot="probe-assistants">{assistants.map(({ id }) => id).join(',')}</span>
            <span data-slot="probe-selected">{selectedAssistantId}</span>
            <span data-slot="probe-assistant-status">{assistants.map(({ status }) => status).join(',')}</span>
            <span data-slot="probe-flow-status">{flowData?.flow?.status}</span>
        </>
    );
}

function Switcher() {
    const navigate = useNavigate();

    return (
        <button
            onClick={() => navigate('/flows/2')}
            type="button"
        >
            open the other flow
        </button>
    );
}

const renderFlowPage = (client: ApolloClient) =>
    render(
        <ApolloProvider client={client}>
            <MemoryRouter initialEntries={['/flows/1']}>
                <Switcher />
                <Routes>
                    <Route
                        element={
                            <FlowProvider>
                                <Probe />
                            </FlowProvider>
                        }
                        path="/flows/:flowId"
                    />
                </Routes>
            </MemoryRouter>
        </ApolloProvider>,
    );

describe('flow page catch-up after the subscriptions attach', () => {
    it('pulls the rows written between the first snapshot and the subscriptions', async () => {
        const { client, flowQueries } = buildClient();

        renderFlowPage(client);

        await waitFor(() => expect(screen.getByTestId('probe-rows')).toHaveTextContent('2'));
        expect(flowQueries).toHaveLength(2);
    });

    // Apollo answers an identical query issued this soon after the first with a replay of it, so
    // without the flag the catch-up never reaches the link at all.
    it('issues the catch-up outside the deduplication of the query it follows', async () => {
        const { client, flowQueries } = buildClient();

        renderFlowPage(client);

        await waitFor(() => expect(flowQueries).toHaveLength(2));
        expect(flowQueries[1]?.deduplication).toBe(false);
    });

    it('runs after the subscriptions, which is the gap it exists to close', async () => {
        const { client, flowQueries, order } = buildClient();

        renderFlowPage(client);

        await waitFor(() => expect(flowQueries).toHaveLength(2));

        const lastSubscription = order.findLastIndex((operation) => operation.startsWith('subscription:'));
        const catchUp = order.findLastIndex((operation) => operation === 'query:flow');

        expect(catchUp).toBeGreaterThan(lastSubscription);
    });

    // The assistant list and the open conversation are read cache-first and their subscriptions
    // only insert into what is already cached, so a frame missed in this window never comes back.
    it('also catches up the assistant list and the open conversation', async () => {
        const { client, order } = buildClient();

        renderFlowPage(client);

        await waitFor(() => expect(order.filter((operation) => operation === 'query:assistants')).toHaveLength(2));
        await waitFor(() => expect(order.filter((operation) => operation === 'query:assistantLogs')).toHaveLength(2));

        const lastSubscription = order.findLastIndex((operation) => operation.startsWith('subscription:'));

        expect(order.findLastIndex((operation) => operation === 'query:assistants')).toBeGreaterThan(lastSubscription);
        expect(order.findLastIndex((operation) => operation === 'query:assistantLogs')).toBeGreaterThan(
            lastSubscription,
        );
    });

    it('keeps a reply that arrived while the conversation was being caught up', async () => {
        const { client, order } = buildClient({ isRowArrivingMidFlight: true });

        renderFlowPage(client);

        // The reply lands first and the catch-up answers 20 ms later with the snapshot taken
        // before it, so the assertion has to wait for that answer — asserting earlier passes on
        // the very code this guards.
        await waitFor(() => expect(order.filter((operation) => operation === 'query:assistantLogs')).toHaveLength(2));
        await waitFor(() => expect(screen.getByTestId('probe-logs')).toHaveTextContent('1,2'), { timeout: 2000 });
        await new Promise((resolve) => {
            setTimeout(resolve, 100);
        });

        // Chronological: the reply the catch-up did not know about belongs after the one it did.
        expect(screen.getByTestId('probe-logs')).toHaveTextContent('1,2');
    });

    it('keeps a reply the watched query had not seen when it answers after the catch-up', async () => {
        const { client, order } = buildClient({ isConversationAnsweringLast: true });

        renderFlowPage(client);

        await waitFor(() => expect(order.filter((operation) => operation === 'query:assistantLogs')).toHaveLength(2));
        await waitFor(() => expect(screen.getByTestId('probe-logs')).toHaveTextContent('1,2'), { timeout: 2000 });
        await new Promise((resolve) => {
            setTimeout(resolve, 100);
        });

        expect(screen.getByTestId('probe-logs')).toHaveTextContent('1,2');
    });

    it('keeps an assistant the watched list had not seen when it answers after the catch-up', async () => {
        const { client, order } = buildClient({ isAssistantListAnsweringLast: true });

        renderFlowPage(client);

        await waitFor(() => expect(order.filter((operation) => operation === 'query:assistants')).toHaveLength(2));
        await waitFor(() =>
            expect(screen.getByTestId('probe-assistants').textContent?.split(',').sort()).toEqual(['7', '9']),
        );
        await new Promise((resolve) => {
            setTimeout(resolve, 100);
        });

        expect(screen.getByTestId('probe-assistants').textContent?.split(',').sort()).toEqual(['7', '9']);
    });

    it('keeps the status a frame delivered before the catch-up asked for the same assistant', async () => {
        const { client, order } = buildClient({ isStatusFresherInCache: true });

        renderFlowPage(client);

        await waitFor(() => expect(order.filter((operation) => operation === 'query:assistants')).toHaveLength(2));
        await waitFor(() => expect(screen.getByTestId('probe-assistants')).toHaveTextContent('9,7'));

        expect(screen.getByTestId('probe-assistant-status')).toHaveTextContent(StatusType.Finished);
    });

    it('takes the answer whose fraction is longer, and later, than the frame the cache holds', async () => {
        const { client, order } = buildClient({ isCacheFractionShorter: true });

        renderFlowPage(client);

        await waitFor(() => expect(order.filter((operation) => operation === 'query:assistants')).toHaveLength(2));
        await waitFor(() =>
            expect(screen.getByTestId('probe-assistant-status')).toHaveTextContent(StatusType.Finished),
        );
    });

    it('takes the answer when the body was rewritten rather than extended', async () => {
        const { client, order } = buildClient({ isReplyRevisedShorter: true });

        renderFlowPage(client);

        await waitFor(() => expect(order.filter((operation) => operation === 'query:assistantLogs')).toHaveLength(2));
        await waitFor(() => expect(screen.getByTestId('probe-logs')).toHaveTextContent('1,2'));

        expect(screen.getByTestId('probe-log-text')).toHaveTextContent('rewritten');
    });

    it('keeps the streamed body when the frame landed before the catch-up asked', async () => {
        const { client, order } = buildClient({ isReplyStreamedBeforeAsking: true });

        renderFlowPage(client);

        await waitFor(() => expect(order.filter((operation) => operation === 'query:assistantLogs')).toHaveLength(2));
        // Row 2 arrives with the catch-up answer alone, so waiting for it proves the merge landed.
        await waitFor(() => expect(screen.getByTestId('probe-logs')).toHaveTextContent('1,2'));

        expect(screen.getByTestId('probe-log-text')).toHaveTextContent('the whole reply');
    });

    it('does not write a stale reply body over the one the stream accumulated', async () => {
        const { client, order } = buildClient({ isReplyStreaming: true });

        renderFlowPage(client);

        await waitFor(() => expect(order.filter((operation) => operation === 'query:assistantLogs')).toHaveLength(2));
        await waitFor(() => expect(screen.getByTestId('probe-log-text')).toHaveTextContent('the whole reply'));
        await new Promise((resolve) => {
            setTimeout(resolve, 100);
        });

        expect(screen.getByTestId('probe-log-text')).toHaveTextContent('the whole reply');
    });

    it('brings back a status change the subscription window swallowed', async () => {
        const { client, order } = buildClient({ isStatusMissedInWindow: true });

        renderFlowPage(client);

        await waitFor(() => expect(order.filter((operation) => operation === 'query:assistants')).toHaveLength(2));
        await waitFor(() => expect(screen.getByTestId('probe-assistant-status')).toHaveTextContent('finished'));
    });

    it('puts an assistant created mid-flight at the front, where newest-first wants it', async () => {
        const { client, order } = buildClient({ isAssistantCreatedMidFlight: true });

        renderFlowPage(client);

        await waitFor(() => expect(order.filter((operation) => operation === 'query:assistants')).toHaveLength(2));
        await waitFor(() => expect(screen.getByTestId('probe-assistants')).toHaveTextContent('9,7'));
        await new Promise((resolve) => {
            setTimeout(resolve, 50);
        });

        expect(screen.getByTestId('probe-assistants').textContent).toBe('9,7');
    });

    it('tries the assistant list again after its catch-up failed, without losing the cache', async () => {
        const reported = vi.spyOn(Log, 'error').mockImplementation(() => undefined);
        const { client, order } = buildClient({ isListCatchUpFailing: true });

        renderFlowPage(client);

        const watchedCatchUpAndRetry = 3;

        await waitFor(
            () =>
                expect(order.filter((operation) => operation === 'query:assistants')).toHaveLength(
                    watchedCatchUpAndRetry,
                ),
            { timeout: 3000 },
        );
        expect(screen.getByTestId('probe-assistants')).toHaveTextContent('7');

        reported.mockRestore();
    });

    it('tries the conversation again after its catch-up failed', async () => {
        const reported = vi.spyOn(Log, 'error').mockImplementation(() => undefined);
        const { client, order } = buildClient({ isLogCatchUpFailing: true });

        renderFlowPage(client);

        const watchedCatchUpAndRetry = 3;

        await waitFor(
            () =>
                expect(order.filter((operation) => operation === 'query:assistantLogs')).toHaveLength(
                    watchedCatchUpAndRetry,
                ),
            { timeout: 3000 },
        );
        await waitFor(() => expect(screen.getByTestId('probe-logs')).toHaveTextContent('1,2'));
        expect(reported).toHaveBeenCalledWith(
            'Assistant log catch-up after subscribing failed:',
            expect.objectContaining({ message: expect.stringContaining('the conversation went away') }),
        );

        reported.mockRestore();
    });

    it('keeps a row the stream delivered between a failed catch-up and its retry', async () => {
        const { client, order } = buildClient({ isRowArrivingBetweenTries: true });

        renderFlowPage(client);

        await waitFor(() => expect(order.filter((operation) => operation === 'query:assistants')).toHaveLength(3));
        await waitFor(() => expect(screen.getByTestId('probe-assistants')).toHaveTextContent('9,7'));
    });

    it('does not put back an assistant the server has dropped', async () => {
        const { client } = buildClient({ isDeletedAssistantCached: true });

        // The cache still holds the assistant that was deleted from another tab: nothing here saw
        // the `assistantDeleted` frame, because this page was not mounted when it arrived.
        client.cache.writeQuery({
            data: { assistants: [otherAssistant, assistant] },
            query: AssistantsDocument,
            variables: { flowId: '1' },
        });

        renderFlowPage(client);

        await waitFor(() => expect(screen.getByTestId('probe-assistants')).toHaveTextContent('9'));
        expect(screen.getByTestId('probe-assistants').textContent).toBe('9');
        expect(screen.getByTestId('probe-selected')).toHaveTextContent('9');
    });

    it('keeps the status a frame delivered while the flow catch-up was in flight', async () => {
        const { client, flowQueries } = buildClient({ isFlowStatusArrivingMidFlight: true });

        renderFlowPage(client);

        await waitFor(() => expect(flowQueries).toHaveLength(2));
        await waitFor(() => expect(screen.getByTestId('probe-flow-status')).toHaveTextContent('finished'));
        await new Promise((resolve) => {
            setTimeout(resolve, 60);
        });

        expect(screen.getByTestId('probe-flow-status')).toHaveTextContent('finished');
    });

    it('reports a catch-up that failed', async () => {
        const reported = vi.spyOn(Log, 'error').mockImplementation(() => undefined);
        const { client, flowQueries } = buildClient({ isCatchUpFailing: true });

        renderFlowPage(client);

        await waitFor(() => expect(flowQueries).toHaveLength(2));
        await waitFor(() =>
            expect(reported).toHaveBeenCalledWith(
                'Flow catch-up after subscribing failed:',
                expect.objectContaining({ message: 'the network went away' }),
            ),
        );

        reported.mockRestore();
    });

    it('tries the catch-up again after a failure instead of leaving the gap open', async () => {
        const reported = vi.spyOn(Log, 'error').mockImplementation(() => undefined);
        const { client, flowQueries } = buildClient({ isCatchUpFailing: true });

        renderFlowPage(client);

        const watchedCatchUpAndRetry = 3;

        await waitFor(() => expect(flowQueries).toHaveLength(watchedCatchUpAndRetry), { timeout: 3000 });
        await waitFor(() => expect(screen.getByTestId('probe-rows')).toHaveTextContent('2'));

        reported.mockRestore();
    });

    it('catches up again on a flow whose snapshot the cache already held', async () => {
        const { client, flowQueries } = buildClient();

        renderFlowPage(client);

        await waitFor(() => expect(flowQueries).toHaveLength(2));

        await userEvent.click(screen.getByRole('button', { name: 'open the other flow' }));
        await waitFor(() => expect(screen.getByTestId('probe-flow-id')).toHaveTextContent('2'));
        await waitFor(() => expect(flowQueries.filter((query) => query.id === '2')).toHaveLength(2));
    });

    it('spends both promised retries when the whole wave fails at once', async () => {
        const reported = vi.spyOn(Log, 'error').mockImplementation(() => undefined);
        const { client, flowQueries } = buildClient({ isEveryCatchUpFailing: true });

        renderFlowPage(client);

        const readsOfFlowOne = () => flowQueries.filter((query) => query.id === '1');
        const initialAndThreeAttempts = 4;

        await waitFor(() => expect(readsOfFlowOne()).toHaveLength(initialAndThreeAttempts), { timeout: 3000 });

        // The budget is a ceiling too: a wave counted twice stops early, one never counted never stops.
        await new Promise((settle) => {
            setTimeout(settle, 200);
        });

        expect(readsOfFlowOne()).toHaveLength(initialAndThreeAttempts);

        reported.mockRestore();
    });

    it('gives a flow opened later its own retry budget', async () => {
        const reported = vi.spyOn(Log, 'error').mockImplementation(() => undefined);
        const { client, flowQueries } = buildClient({ isEveryCatchUpFailing: true });

        renderFlowPage(client);

        const initialAndThreeAttempts = 4;

        await waitFor(
            () => expect(flowQueries.filter((query) => query.id === '1')).toHaveLength(initialAndThreeAttempts),
            {
                timeout: 3000,
            },
        );

        await userEvent.click(screen.getByRole('button', { name: 'open the other flow' }));
        await waitFor(() => expect(screen.getByTestId('probe-flow-id')).toHaveTextContent('2'));

        await waitFor(
            () => expect(flowQueries.filter((query) => query.id === '2')).toHaveLength(initialAndThreeAttempts),
            {
                timeout: 3000,
            },
        );

        reported.mockRestore();
    });
});
