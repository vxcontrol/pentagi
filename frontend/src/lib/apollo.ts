import type { DefaultOptions } from '@apollo/client';
import { ApolloClient, createHttpLink, InMemoryCache, split } from '@apollo/client';
import { GraphQLWsLink } from '@apollo/client/link/subscriptions';
import { getMainDefinition } from '@apollo/client/utilities';
import { createClient } from 'graphql-ws';
import { LRUCache } from 'lru-cache';

import type { AssistantLogFragmentFragment } from '@/graphql/types';
import { AssistantLogFragmentFragmentDoc } from '@/graphql/types';
import { Log } from '@/lib/log';
import { baseUrl } from '@/models/Api';

// Local cache for accumulating assistant log streaming parts during real-time updates.
// We use LRUCache to store each log record by its unique logId, because Apollo cache
// will overwrite fields and does not support partial accumulation for streaming logs.
const streamingAssistantLogs = new LRUCache<string, {
    message: string | null;
    thinking: string | null;
    result: string | null;
}>({
    max: 500, // Maximum number of log records to keep in cache
    ttl: 1000 * 60 * 5, // Each log record lives for 5 minutes (in milliseconds)
});

const httpLink = createHttpLink({
    uri: `${window.location.origin}${baseUrl}/graphql`,
    credentials: 'include',
});

const wsLink = new GraphQLWsLink(
    createClient({
        url: `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}${baseUrl}/graphql`,
        retryAttempts: 5,
        connectionParams: () => {
            return {}; // Cookies are handled automatically
        },
        on: {
            connected: () => Log.debug('GraphQL WebSocket connected'),
            error: (error) => Log.error('GraphQL WebSocket error:', error),
            closed: () => Log.debug('GraphQL WebSocket closed'),
            connecting: () => Log.debug('GraphQL WebSocket connecting...'),
            ping: () => Log.debug('GraphQL WebSocket ping'),
            pong: () => Log.debug('GraphQL WebSocket pong'),
        },
        shouldRetry: () => true,
        retryWait: (retries) => new Promise((resolve) => {
            const timeout = Math.min(1000 * 2 ** retries, 10000);
            setTimeout(() => resolve(), timeout);
        }),
    }),
);

const link = split(
    ({ query }) => {
        const definition = getMainDefinition(query);
        return definition.kind === 'OperationDefinition' && definition.operation === 'subscription';
    },
    wsLink,
    httpLink,
);

// Helper functions
const addIncoming = (existing: any[], incoming: any, cache: any) => {
    const incomingId = cache.identify(incoming);

    if (existing.some((item) => cache.identify(item) === incomingId)) {
        return existing;
    }

    return [...existing, incoming];
};

const addTopIncoming = (existing: any[], incoming: any, cache: any) => {
    const incomingId = cache.identify(incoming);

    if (existing.some((item) => cache.identify(item) === incomingId)) {
        return existing;
    }

    return [incoming, ...existing];
};

const updateIncoming = (existing: any[], incoming: any, cache: any) => {
    const incomingId = cache.identify(incoming);

    return existing.map((item) => (cache.identify(item) === incomingId ? incoming : item));
};

const deleteIncoming = (existing: any[], incoming: any, cache: any) => {
    const incomingId = cache.identify(incoming);

    return existing.filter((item) => cache.identify(item) !== incomingId);
};

const concatStrings = (existing: string | null | undefined, incoming: string | null | undefined) => {
    if (existing && incoming) {
        return `${existing}${incoming}`;
    } else if (existing) {
        return existing;
    } else if (incoming) {
        return incoming;
    }

    return null;
};

const cache = new InMemoryCache({
    typePolicies: {
        Query: {
            fields: {
                // Ensure tasks field is properly merged with incoming data
                tasks: {
                    merge(_existing = [], incoming) {
                        return incoming; // Always use latest task data
                    },
                },
                assistants: {
                    merge(_existing = [], incoming) {
                        return incoming; // Always use latest assistants data
                    },
                },
                assistantLogs: {
                    merge(_existing = [], incoming) {
                        return incoming; // Always use latest assistantLogs data
                    },
                },
            },
        },
        Mutation: {
            fields: {
                createFlow: {
                    merge(_, incoming, { cache }) {
                        cache.modify({
                            fields: {
                                flows: (existing = []) => addTopIncoming(existing, incoming, cache),
                            },
                        });
                    },
                },
                createAssistant: {
                    merge(_, incoming, { cache }) {
                        // Update the flow
                        if (incoming?.flow) {
                            cache.modify({
                                fields: {
                                    flows: (existing = []) => updateIncoming(existing, incoming.flow, cache),
                                },
                            });
                        }

                        // Add the assistant to the list
                        if (incoming?.assistant) {
                            cache.modify({
                                fields: {
                                    assistants: (existing = []) => addTopIncoming(existing, incoming.assistant, cache),
                                },
                            });
                        }
                    },
                },
                stopAssistant: {
                    merge(_, incoming, { cache }) {
                        cache.modify({
                            fields: {
                                assistants: (existing = []) => updateIncoming(existing, incoming, cache),
                            },
                        });
                    },
                },
            },
        },
        Subscription: {
            fields: {
                messageLogAdded: {
                    merge(_, incoming, { cache }) {
                        cache.modify({
                            fields: {
                                messageLogs: (existing = []) => addIncoming(existing, incoming, cache),
                            },
                        });
                    },
                },
                messageLogUpdated: {
                    merge(_, incoming, { cache }) {
                        cache.modify({
                            fields: {
                                messageLogs: (existing = []) => updateIncoming(existing, incoming, cache),
                            },
                        });
                    },
                },
                screenshotAdded: {
                    merge(_, incoming, { cache }) {
                        cache.modify({
                            fields: {
                                screenshots: (existing = []) => addIncoming(existing, incoming, cache),
                            },
                        });
                    },
                },
                terminalLogAdded: {
                    merge(_, incoming, { cache }) {
                        cache.modify({
                            fields: {
                                terminalLogs: (existing = []) => addIncoming(existing, incoming, cache),
                            },
                        });
                    },
                },
                taskCreated: {
                    merge(_, incoming, { cache }) {
                        cache.modify({
                            fields: {
                                tasks: (existing = []) => {
                                    // Add the new task to the top of the list
                                    return addTopIncoming(existing, incoming, cache);
                                },
                            },
                        });
                        // Force refresh any related queries
                        cache.gc();
                    },
                },
                taskUpdated: {
                    merge(_, incoming, { cache }) {
                        cache.modify({
                            fields: {
                                tasks: (existing = []) => updateIncoming(existing, incoming, cache),
                            },
                        });
                    },
                },
                flowCreated: {
                    merge(_, incoming, { cache }) {
                        cache.modify({
                            fields: {
                                flows: (existing = []) => addTopIncoming(existing, incoming, cache),
                            },
                        });
                    },
                },
                flowUpdated: {
                    merge(_, incoming, { cache }) {
                        cache.modify({
                            fields: {
                                flows: (existing = []) => updateIncoming(existing, incoming, cache),
                            },
                        });
                    },
                },
                flowDeleted: {
                    merge(_, incoming, { cache }) {
                        cache.modify({
                            fields: {
                                flows: (existing = []) => deleteIncoming(existing, incoming, cache),
                            },
                        });
                    },
                },
                agentLogAdded: {
                    merge(_, incoming, { cache }) {
                        cache.modify({
                            fields: {
                                agentLogs: (existing = []) => addIncoming(existing, incoming, cache),
                            },
                        });
                    },
                },
                searchLogAdded: {
                    merge(_, incoming, { cache }) {
                        cache.modify({
                            fields: {
                                searchLogs: (existing = []) => addIncoming(existing, incoming, cache),
                            },
                        });
                    },
                },
                vectorStoreLogAdded: {
                    merge(_, incoming, { cache }) {
                        cache.modify({
                            fields: {
                                vectorStoreLogs: (existing = []) => addIncoming(existing, incoming, cache),
                            },
                        });
                    },
                },
                assistantLogAdded: {
                    merge(_, incoming, { cache }) {
                        cache.modify({
                            fields: {
                                assistantLogs: (existing = []) => addIncoming(existing, incoming, cache),
                            },
                        });
                    },
                },
                assistantLogUpdated: {
                    merge(_, incoming, { cache, toReference }) {
                        cache.modify({
                            fields: {
                                assistantLogs: (existing = []) => {
                                    // Extract actual log record object from Apollo cache reference
                                    const incomingId = cache.identify(incoming);
                                    const logRecord = cache.readFragment({
                                        id: incomingId,
                                        fragment: AssistantLogFragmentFragmentDoc,
                                    }) as AssistantLogFragmentFragment;
                                    if (!logRecord) {
                                        return addIncoming(existing, incoming, cache);
                                    }

                                    // Initiate streaming for new assistant log record
                                    const logRecordKey = incomingId || `${logRecord.id}`;
                                    const existingIndex = existing.findIndex((item: Record<string, any>) => cache.identify(item) === incomingId);
                                    if (existingIndex === -1) {
                                        streamingAssistantLogs.set(logRecordKey, {
                                            message: logRecord.message,
                                            thinking: logRecord.thinking || null,
                                            result: logRecord.result,
                                        });
                                        return addIncoming(existing, incoming, cache);
                                    }

                                    if (logRecord.appendPart === true) {
                                        // Handle streaming message parts - accumulate locally to prevent Apollo cache overwrites
                                        const emptyLogRecord = { message: null, thinking: null, result: null };
                                        const cachedLogRecord = streamingAssistantLogs.get(logRecordKey) || emptyLogRecord;
                                        const accumulatedLogRecord = {
                                            message: concatStrings(cachedLogRecord.message, logRecord.message),
                                            thinking: concatStrings(cachedLogRecord.thinking, logRecord.thinking),
                                            result: concatStrings(cachedLogRecord.result, logRecord.result),
                                        };
                                        streamingAssistantLogs.set(logRecordKey, accumulatedLogRecord);

                                        const updatedLogRecord = toReference({
                                            ...logRecord,
                                            appendPart: false, // prevent infinite loop on updating the log record
                                            message: accumulatedLogRecord.message || '',
                                            thinking: accumulatedLogRecord.thinking,
                                            result: accumulatedLogRecord.result || '',
                                        }, true);

                                        return updateIncoming(existing, updatedLogRecord, cache);
                                    }

                                    return updateIncoming(existing, incoming, cache);
                                },
                            },
                        });
                    },
                },
                assistantCreated: {
                    merge(_, incoming, { cache }) {
                        cache.modify({
                            fields: {
                                assistants: (existing = []) => addTopIncoming(existing, incoming, cache),
                            },
                        });
                    },
                },
                assistantUpdated: {
                    merge(_, incoming, { cache }) {
                        cache.modify({
                            fields: {
                                assistants: (existing = []) => updateIncoming(existing, incoming, cache),
                            },
                        });
                    },
                },
                assistantDeleted: {
                    merge(_, incoming, { cache }) {
                        cache.modify({
                            fields: {
                                assistants: (existing = []) => deleteIncoming(existing, incoming, cache),
                            },
                        });
                    },
                },
            },
        },
    },
});

const defaultOptions: DefaultOptions = {
    watchQuery: {
        fetchPolicy: 'cache-and-network',
        nextFetchPolicy: 'cache-first',
        notifyOnNetworkStatusChange: true,
    },
};

export const client = new ApolloClient({
    link,
    cache,
    defaultOptions,
});

export default client;
