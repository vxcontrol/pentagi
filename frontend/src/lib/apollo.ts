import type { DefaultOptions } from '@apollo/client';
import { ApolloClient, createHttpLink, InMemoryCache, split } from '@apollo/client';
import { GraphQLWsLink } from '@apollo/client/link/subscriptions';
import { getMainDefinition } from '@apollo/client/utilities';
import { createClient } from 'graphql-ws';

import { Log } from '@/lib/log';
import { baseUrl } from '@/models/Api';

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
