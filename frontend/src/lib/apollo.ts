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

// Add provider to sorted list in lexicographic order by name
const addProviderSorted = (existing: any[], incoming: any, cache: any) => {
    const incomingName = incoming.name;
    const incomingType = incoming.type;

    // Check if provider already exists
    if (existing.some((item) => item.name === incomingName && item.type === incomingType)) {
        return existing;
    }

    // Find the correct position to insert (sorted by name in lexicographic order)
    const insertIndex = existing.findIndex((item) => item.name > incomingName);
    
    if (insertIndex === -1) {
        return [...existing, incoming];
    } else {
        // Insert at the correct position
        return [
            ...existing.slice(0, insertIndex),
            incoming,
            ...existing.slice(insertIndex)
        ];
    }
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

// Helper to convert ProviderConfig to Provider
const providerConfigToProvider = (providerConfig: any) => {
    return {
        __typename: 'Provider',
        name: providerConfig.name,
        type: providerConfig.type,
    };
};

const cache = new InMemoryCache({
    typePolicies: {
        ProviderConfig: {
            keyFields: (object) => {
                // don't normalize default providers with id: 0
                // they should be stored inline within DefaultProvidersConfig
                if (object.id === 0 || object.id === '0') {
                    return false;
                }
                // normalize user-defined providers by id
                return ['id'];
            },
        },
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
                settingsProviders: {
                    merge(_existing, incoming) {
                        return incoming; // Always use latest providers settings
                    },
                },
                settingsPrompts: {
                    merge(_existing, incoming) {
                        return incoming; // Always use latest prompts settings
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
                createProvider: {
                    merge(_, incoming, { cache }) {
                        // Add to providers list (short format, sorted by name)
                        const provider = providerConfigToProvider(incoming);
                        cache.modify({
                            fields: {
                                providers: (existing = []) => addProviderSorted(existing, provider, cache),
                            },
                        });

                        // Update settingsProviders userDefined list (add to end, sorted by ID)
                        cache.modify({
                            fields: {
                                settingsProviders: (existing) => {
                                    if (existing && existing.userDefined) {
                                        return {
                                            ...existing,
                                            userDefined: addIncoming(existing.userDefined, incoming, cache),
                                        };
                                    }
                                    return existing;
                                },
                            },
                        });
                    },
                },
                updateProvider: {
                    merge(_, incoming, { cache }) {
                        const provider = providerConfigToProvider(incoming);
                        cache.modify({
                            fields: {
                                providers: (existing = []) => {
                                    // Check if provider exists in the list
                                    const existingProvider = existing.find((item: any) => 
                                        item.name === provider.name && item.type === provider.type
                                    );
                                    
                                    if (!existingProvider) {
                                        // Provider not found, invalidate cache to refetch
                                        cache.evict({ fieldName: 'providers' });
                                        return existing;
                                    }
                                    
                                    // Update existing provider (only name can change, type cannot)
                                    return existing.map((item: any) => 
                                        item.name === provider.name && item.type === provider.type 
                                            ? provider 
                                            : item
                                    );
                                },
                            },
                        });

                        cache.modify({
                            fields: {
                                settingsProviders: (existing) => {
                                    if (existing && existing.userDefined) {
                                        return {
                                            ...existing,
                                            userDefined: updateIncoming(existing.userDefined, incoming, cache),
                                        };
                                    }
                                    return existing;
                                },
                            },
                        });
                    },
                },
                createPrompt: {
                    merge(_, incoming, { cache }) {
                        cache.modify({
                            fields: {
                                settingsPrompts: (existing) => {
                                    if (existing && existing.userDefined) {
                                        return {
                                            ...existing,
                                            userDefined: addTopIncoming(existing.userDefined, incoming, cache),
                                        };
                                    }
                                    return existing;
                                },
                            },
                        });
                    },
                },
                updatePrompt: {
                    merge(_, incoming, { cache }) {
                        cache.modify({
                            fields: {
                                settingsPrompts: (existing) => {
                                    if (existing && existing.userDefined) {
                                        return {
                                            ...existing,
                                            userDefined: updateIncoming(existing.userDefined, incoming, cache),
                                        };
                                    }
                                    return existing;
                                },
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
                providerCreated: {
                    merge(_, incoming, { cache }) {
                        // Add to providers list (short format, sorted by name)
                        const provider = providerConfigToProvider(incoming);
                        cache.modify({
                            fields: {
                                providers: (existing = []) => addProviderSorted(existing, provider, cache),
                            },
                        });

                        // Update settingsProviders userDefined list (add to end, sorted by ID)
                        cache.modify({
                            fields: {
                                settingsProviders: (existing) => {
                                    if (existing && existing.userDefined) {
                                        return {
                                            ...existing,
                                            userDefined: addIncoming(existing.userDefined, incoming, cache),
                                        };
                                    }
                                    return existing;
                                },
                            },
                        });
                    },
                },
                providerUpdated: {
                    merge(_, incoming, { cache }) {
                        // Update providers list (short format)
                        const provider = providerConfigToProvider(incoming);
                        cache.modify({
                            fields: {
                                providers: (existing = []) => {
                                    // Check if provider exists in the list
                                    const existingProvider = existing.find((item: any) => 
                                        item.name === provider.name && item.type === provider.type
                                    );
                                    
                                    if (!existingProvider) {
                                        // Provider not found, invalidate cache to refetch
                                        cache.evict({ fieldName: 'providers' });
                                        return existing;
                                    }
                                    
                                    // Update existing provider (only name can change, type cannot)
                                    return existing.map((item: any) => 
                                        item.name === provider.name && item.type === provider.type 
                                            ? provider 
                                            : item
                                    );
                                },
                            },
                        });

                        // Update settingsProviders userDefined list
                        cache.modify({
                            fields: {
                                settingsProviders: (existing) => {
                                    if (existing && existing.userDefined) {
                                        return {
                                            ...existing,
                                            userDefined: updateIncoming(existing.userDefined, incoming, cache),
                                        };
                                    }
                                    return existing;
                                },
                            },
                        });
                    },
                },
                providerDeleted: {
                    merge(_, incoming, { cache }) {
                        // Remove from providers list (short format)
                        const provider = providerConfigToProvider(incoming);
                        cache.modify({
                            fields: {
                                providers: (existing = []) => {
                                    // Filter out by name+type since Provider doesn't have id
                                    return existing.filter((item: any) => 
                                        !(item.name === provider.name && item.type === provider.type)
                                    );
                                },
                            },
                        });

                        // Remove from settingsProviders userDefined list
                        cache.modify({
                            fields: {
                                settingsProviders: (existing) => {
                                    if (existing && existing.userDefined) {
                                        return {
                                            ...existing,
                                            userDefined: deleteIncoming(existing.userDefined, incoming, cache),
                                        };
                                    }
                                    return existing;
                                },
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
