import type { DefaultOptions, FetchResult, Operation, Reference } from '@apollo/client';

import { ApolloClient, ApolloLink, createHttpLink, InMemoryCache, Observable, split } from '@apollo/client';
import { GraphQLWsLink } from '@apollo/client/link/subscriptions';
import { getMainDefinition } from '@apollo/client/utilities';
import { createClient } from 'graphql-ws';
import { LRUCache } from 'lru-cache';

import type { AssistantLogFragmentFragment } from '@/graphql/types';

import { Log } from '@/lib/log';
import { baseUrl } from '@/models/api';

// Global cache for accumulating assistant log streaming parts during real-time updates.
// This cache is shared across all components and subscription handlers.
export const streamingAssistantLogs = new LRUCache<
    string,
    {
        message: null | string;
        result: null | string;
        thinking: null | string;
    }
>({
    max: 500, // Maximum number of log records to keep in cache
    ttl: 1000 * 60 * 5, // Each log record lives for 5 minutes (in milliseconds)
});

// Helper to concatenate strings, handling null/undefined values
const concatStrings = (existing: null | string | undefined, incoming: null | string | undefined): null | string => {
    if (existing && incoming) {
        return `${existing}${incoming}`;
    } else if (existing) {
        return existing;
    } else if (incoming) {
        return incoming;
    }

    return null;
};

const httpLink = createHttpLink({
    credentials: 'include',
    uri: `${window.location.origin}${baseUrl}/graphql`,
});

const wsLink = new GraphQLWsLink(
    createClient({
        connectionParams: () => {
            return {}; // Cookies are handled automatically
        },
        on: {
            closed: () => Log.debug('GraphQL WebSocket closed'),
            connected: () => Log.debug('GraphQL WebSocket connected'),
            connecting: () => Log.debug('GraphQL WebSocket connecting...'),
            error: (error) => Log.error('GraphQL WebSocket error:', error),
            ping: () => Log.debug('GraphQL WebSocket ping'),
            pong: () => Log.debug('GraphQL WebSocket pong'),
        },
        retryAttempts: 5,
        retryWait: (retries) =>
            new Promise((resolve) => {
                const timeout = Math.min(1000 * 2 ** retries, 10000);
                setTimeout(() => resolve(), timeout);
            }),
        shouldRetry: () => true,
        url: `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}${baseUrl}/graphql`,
    }),
);

// Link to handle streaming assistant logs (appendPart)
// This link intercepts assistantLogUpdated subscription results and accumulates streaming parts
const streamingLink = new ApolloLink((operation: Operation, forward) => {
    return new Observable((observer) => {
        const subscription = forward(operation).subscribe({
            complete: observer.complete.bind(observer),
            error: observer.error.bind(observer),
            next: (result: FetchResult) => {
                // Check if this is an assistantLogUpdated subscription result
                if (result.data?.assistantLogUpdated) {
                    const logUpdate = result.data.assistantLogUpdated as AssistantLogFragmentFragment;

                    try {
                        // Handle streaming parts (appendPart)
                        if (logUpdate.appendPart && logUpdate.id) {
                            const logRecordKey = `AssistantLog:${logUpdate.id}`;
                            const cachedLog = streamingAssistantLogs.get(logRecordKey) || {
                                message: null,
                                result: null,
                                thinking: null,
                            };

                            // Accumulate the streaming parts
                            const accumulatedLog = {
                                message: concatStrings(cachedLog.message, logUpdate.message),
                                result: concatStrings(cachedLog.result, logUpdate.result),
                                thinking: concatStrings(cachedLog.thinking, logUpdate.thinking),
                            };

                            streamingAssistantLogs.set(logRecordKey, accumulatedLog);

                            // Modify the result to include accumulated data
                            result = {
                                ...result,
                                data: {
                                    ...result.data,
                                    assistantLogUpdated: {
                                        ...logUpdate,
                                        appendPart: false, // Mark as processed
                                        message: accumulatedLog.message || '',
                                        result: accumulatedLog.result || '',
                                        thinking: accumulatedLog.thinking,
                                    },
                                },
                            };
                        } else if (logUpdate.id) {
                            // Final update - clear accumulated data
                            const logRecordKey = `AssistantLog:${logUpdate.id}`;
                            streamingAssistantLogs.delete(logRecordKey);
                        }
                    } catch (error) {
                        Log.error('Error processing streaming assistant log:', error);
                    }
                }

                observer.next(result);
            },
        });

        return () => subscription.unsubscribe();
    });
});

// Mapping of subscription names to their corresponding cache field names
const subscriptionToCacheFieldMap: Record<string, string> = {
    agentLogAdded: 'agentLogs',
    assistantLogAdded: 'assistantLogs',
    assistantLogUpdated: 'assistantLogs',
    messageLogAdded: 'messageLogs',
    messageLogUpdated: 'messageLogs',
    screenshotAdded: 'screenshots',
    searchLogAdded: 'searchLogs',
    taskCreated: 'tasks',
    terminalLogAdded: 'terminalLogs',
    vectorStoreLogAdded: 'vectorStoreLogs',
};

// Helper function to find item index in cache array
const findItemIndex = (
    array: readonly Reference[],
    itemId: number | string,
    readField: (fieldName: string, ref: Reference) => unknown,
): number => array.findIndex((ref) => readField('id', ref as Reference) === itemId);

// Helper function to handle cache field update for a single subscription
const handleCacheFieldUpdate = (
    subscriptionName: string,
    existing: unknown,
    newItem: { id: number | string },
    readField: (fieldName: string, ref: Reference) => unknown,
    toReference: unknown,
): readonly Reference[] => {
    const existingArray = (existing ?? []) as readonly Reference[];
    const toReferenceFunc = toReference as (item: unknown) => Reference | undefined;
    const newRef = toReferenceFunc(newItem);

    if (!newRef) {
        return existingArray;
    }

    const existingItemIndex = findItemIndex(existingArray, newItem.id, readField);
    const itemExists = existingItemIndex !== -1;

    // If item exists, keep existing array (Apollo auto-updates normalized entities)
    if (itemExists) {
        return existingArray;
    }

    // Item doesn't exist - add it to the end
    return [...existingArray, newRef];
};

// Helper function to process a single subscription update
const processSubscriptionUpdate = (
    cache: InMemoryCache,
    subscriptionName: string,
    cacheField: string,
    newItem: undefined | { id: number | string },
): void => {
    if (!newItem?.id) {
        return;
    }

    try {
        cache.modify({
            fields: {
                [cacheField]: (existing, { readField, toReference }) =>
                    handleCacheFieldUpdate(subscriptionName, existing, newItem, readField, toReference),
            },
        });
    } catch (error) {
        Log.error(`Error updating cache for ${subscriptionName}:`, error);
    }
};

// Link to automatically update cache when "Added" subscriptions fire
// This link intercepts subscription results and adds new items to the corresponding cache arrays
const createSubscriptionCacheLink = (cache: InMemoryCache) => {
    return new ApolloLink((operation: Operation, forward) => {
        return new Observable((observer) => {
            const subscription = forward(operation).subscribe({
                complete: observer.complete.bind(observer),
                error: observer.error.bind(observer),
                next: (result: FetchResult) => {
                    // Process each subscription type and update cache
                    if (result.data) {
                        Object.entries(subscriptionToCacheFieldMap)
                            .map(([subscriptionName, cacheField]) => ({
                                cacheField,
                                newItem: result.data?.[subscriptionName],
                                subscriptionName,
                            }))
                            .filter(({ newItem }) => newItem?.id)
                            .forEach(({ cacheField, newItem, subscriptionName }) => {
                                processSubscriptionUpdate(cache, subscriptionName, cacheField, newItem);
                            });
                    }

                    // Process the result after cache update
                    observer.next(result);
                },
            });

            return () => subscription.unsubscribe();
        });
    });
};

// Split traffic between WebSocket (subscriptions) and HTTP (queries/mutations)
const transportLink = split(
    ({ query }) => {
        const definition = getMainDefinition(query);

        return definition.kind === 'OperationDefinition' && definition.operation === 'subscription';
    },
    wsLink,
    httpLink,
);

// Create cache first
const cache = new InMemoryCache({
    typePolicies: {
        // Type-specific policies for normalization and identification (in alphabetical order)
        AgentLog: {
            keyFields: ['id'],
        },
        Assistant: {
            keyFields: ['id'],
        },
        AssistantLog: {
            keyFields: ['id'],
        },
        Flow: {
            keyFields: ['id'],
        },
        MessageLog: {
            keyFields: ['id'],
        },
        ProviderConfig: {
            keyFields: (object) => {
                // Don't normalize default providers with id: 0
                // They should be stored inline within DefaultProvidersConfig
                if (object.id === 0 || object.id === '0') {
                    return false;
                }

                // Normalize user-defined providers by id
                return ['id'];
            },
        },
        // Query field policies
        Query: {
            fields: {
                // Agent logs - cache by flowId argument
                agentLogs: {
                    keyArgs: ['flowId'],
                    merge(_existing, incoming) {
                        return incoming;
                    },
                },
                // Assistant logs - cache by flowId and assistantId arguments
                assistantLogs: {
                    keyArgs: ['flowId', 'assistantId'],
                    merge(_existing, incoming) {
                        return incoming;
                    },
                },
                // Assistants list - cache by flowId argument
                assistants: {
                    keyArgs: ['flowId'],
                    merge(_existing, incoming) {
                        return incoming;
                    },
                },
                // Flow query with arguments - cache by flowId
                flow: {
                    read(existing, { args, toReference }) {
                        if (!args?.flowId) {
                            return existing;
                        }

                        return (
                            existing ||
                            toReference({
                                __typename: 'Flow',
                                id: args.flowId,
                            })
                        );
                    },
                },
                // Flows list - always use latest data from network
                flows: {
                    merge(_existing, incoming) {
                        return incoming;
                    },
                },
                // Message logs - cache by flowId argument
                messageLogs: {
                    keyArgs: ['flowId'],
                    merge(_existing, incoming) {
                        return incoming;
                    },
                },
                // Providers list - always use latest
                providers: {
                    merge(_existing, incoming) {
                        return incoming;
                    },
                },
                // Screenshots - cache by flowId argument
                screenshots: {
                    keyArgs: ['flowId'],
                    merge(_existing, incoming) {
                        return incoming;
                    },
                },
                // Search logs - cache by flowId argument
                searchLogs: {
                    keyArgs: ['flowId'],
                    merge(_existing, incoming) {
                        return incoming;
                    },
                },
                // Settings - always use latest
                settingsPrompts: {
                    merge(_existing, incoming) {
                        return incoming;
                    },
                },
                settingsProviders: {
                    merge(_existing, incoming) {
                        return incoming;
                    },
                },
                // Tasks - cache by flowId argument
                tasks: {
                    keyArgs: ['flowId'],
                    merge(_existing, incoming) {
                        return incoming;
                    },
                },
                // Terminal logs - cache by flowId argument
                terminalLogs: {
                    keyArgs: ['flowId'],
                    merge(_existing, incoming) {
                        return incoming;
                    },
                },
                // Vector store logs - cache by flowId argument
                vectorStoreLogs: {
                    keyArgs: ['flowId'],
                    merge(_existing, incoming) {
                        return incoming;
                    },
                },
            },
        },
        Screenshot: {
            keyFields: ['id'],
        },
        SearchLog: {
            keyFields: ['id'],
        },
        Subtask: {
            keyFields: ['id'],
        },
        Task: {
            keyFields: ['id'],
        },
        TerminalLog: {
            keyFields: ['id'],
        },
        UserPrompt: {
            keyFields: ['id'],
        },
        VectorStoreLog: {
            keyFields: ['id'],
        },
    },
});

// Create subscription cache link with reference to cache instance
const subscriptionCacheLink = createSubscriptionCacheLink(cache);

// Final link chain: subscriptionCacheLink -> streamingLink -> transportLink
// Order matters: cache updates happen after streaming processing, before result is sent to components
const link = ApolloLink.from([subscriptionCacheLink, streamingLink, transportLink]);

const defaultOptions: DefaultOptions = {
    watchQuery: {
        fetchPolicy: 'cache-and-network',
        nextFetchPolicy: 'cache-first',
        notifyOnNetworkStatusChange: true,
    },
};

export const client = new ApolloClient({
    cache,
    defaultOptions,
    link,
});

export default client;
