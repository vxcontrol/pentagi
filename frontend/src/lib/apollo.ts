import type { Reference, StoreObject } from '@apollo/client';

import { ApolloClient, ApolloLink, gql, HttpLink, InMemoryCache, Observable } from '@apollo/client';
import { CombinedGraphQLErrors, ServerError, ServerParseError } from '@apollo/client/errors';
import { ErrorLink } from '@apollo/client/link/error';
import { GraphQLWsLink } from '@apollo/client/link/subscriptions';
import { getMainDefinition } from '@apollo/client/utilities';
import { createClient } from 'graphql-ws';
import { LRUCache } from 'lru-cache';

import type { AssistantLogFragmentFragment } from '@/graphql/types';
import type { ListOrder } from '@/lib/list-order';

import { fetchWithDeadline } from '@/lib/graphql-deadline';
import { orderOf, placeMissedRows } from '@/lib/list-order';
import { Log } from '@/lib/log';
import { baseUrl } from '@/models/api';

CombinedGraphQLErrors.formatMessage = (errors, { defaultFormatMessage }) =>
    defaultFormatMessage([...new Map(errors.map((error) => [error.message, error])).values()]);

const GRAPHQL_ENDPOINT = `${baseUrl}/graphql`;
const ASSISTANT_LOG_TYPENAME = 'AssistantLog';

const LOCAL_WRITE_CACHE_MAX_ENTRIES = 1000;
const LOCAL_WRITE_CACHE_TTL_MS = 1000 * 60 * 5;

const localWriteAt = new LRUCache<string, number>({
    max: LOCAL_WRITE_CACHE_MAX_ENTRIES,
    ttl: LOCAL_WRITE_CACHE_TTL_MS,
});

/**
 * When a subscription last wrote this row into the cache. A catch-up compares it against the moment
 * it asked: the backend persists on a timer, so anything written locally after the question went
 * out is ahead of the answer coming back.
 */
export const lastLocalWriteAt = (field: string, id: string): number => localWriteAt.get(`${field}:${id}`) ?? 0;
const MAX_RETRY_DELAY_MS = 30_000;
const STREAMING_CACHE_MAX_ENTRIES = 500;
const STREAMING_CACHE_TTL_MS = 1000 * 60 * 5;
const STREAMING_THROTTLE_MS = 50;

type StreamingLogEntry = {
    lastUpdate: number;
    message: null | string;
    result: null | string;
    thinking: null | string;
};

type SubscriptionAction = 'delete' | 'insert';

const EMPTY_LOG_ENTRY: StreamingLogEntry = { lastUpdate: 0, message: null, result: null, thinking: null };

const RESUMED_LOG_FRAGMENT = gql`
    fragment ResumedAssistantLog on AssistantLog {
        message
        result
        thinking
    }
`;

const concatStrings = (existing: null | string | undefined, incoming: null | string | undefined): null | string => {
    if (existing === null || existing === undefined) {
        return incoming ?? null;
    }

    if (incoming === null || incoming === undefined) {
        return existing;
    }

    return `${existing}${incoming}`;
};

const resolveSubscriptionAction = (name: string): SubscriptionAction =>
    name.endsWith('Deleted') ? 'delete' : 'insert';

const isSubscriptionOperation = ({ query }: ApolloLink.Operation): boolean => {
    const definition = getMainDefinition(query);

    return definition.kind === 'OperationDefinition' && definition.operation === 'subscription';
};

const createInterceptLink = (
    transform: (result: ApolloLink.Result, operation: ApolloLink.Operation) => ApolloLink.Result,
): ApolloLink =>
    new ApolloLink(
        (operation: ApolloLink.Operation, forward) =>
            new Observable((observer) => {
                const subscription = forward(operation).subscribe({
                    complete: observer.complete.bind(observer),
                    error: observer.error.bind(observer),
                    next: (result: ApolloLink.Result) => observer.next(transform(result, operation)),
                });

                return () => subscription.unsubscribe();
            }),
    );

export const subscriptionToCacheFieldMap: Record<string, string> = {
    agentLogAdded: 'agentLogs',
    apiTokenCreated: 'apiTokens',
    apiTokenDeleted: 'apiTokens',
    apiTokenUpdated: 'apiTokens',
    assistantCreated: 'assistants',
    assistantDeleted: 'assistants',
    assistantLogAdded: 'assistantLogs',
    assistantLogUpdated: 'assistantLogs',
    assistantUpdated: 'assistants',
    flowCreated: 'flows',
    flowDeleted: 'flows',
    flowFileAdded: 'flowFiles',
    flowFileDeleted: 'flowFiles',
    flowFileUpdated: 'flowFiles',
    flowTemplateCreated: 'flowTemplates',
    flowTemplateDeleted: 'flowTemplates',
    flowTemplateUpdated: 'flowTemplates',
    flowUpdated: 'flows',
    knowledgeDocumentCreated: 'knowledgeDocuments',
    knowledgeDocumentDeleted: 'knowledgeDocuments',
    knowledgeDocumentUpdated: 'knowledgeDocuments',
    messageLogAdded: 'messageLogs',
    messageLogUpdated: 'messageLogs',
    resourceAdded: 'resources',
    resourceDeleted: 'resources',
    resourceUpdated: 'resources',
    screenshotAdded: 'screenshots',
    searchLogAdded: 'searchLogs',
    settingsUserUpdated: 'settingsUser',
    taskCreated: 'tasks',
    taskUpdated: 'tasks',
    terminalLogAdded: 'terminalLogs',
    vectorStoreLogAdded: 'vectorStoreLogs',
};

const payloadOwnedArgs: Record<string, readonly string[]> = {
    assistantLogs: ['assistantId'],
};

const matchesCacheVariant = (
    storeFieldName: string,
    cacheField: string,
    subscriptionVariables?: Record<string, unknown>,
    item?: Record<string, unknown>,
): boolean => {
    if (!subscriptionVariables || storeFieldName === cacheField) {
        return true;
    }

    const argsStart = storeFieldName.indexOf('{');
    const argsEnd = storeFieldName.lastIndexOf('}');

    if (argsStart === -1 || argsEnd < argsStart) {
        return true;
    }

    const ownedArgs = payloadOwnedArgs[cacheField] ?? [];

    try {
        const storedArgs = JSON.parse(storeFieldName.slice(argsStart, argsEnd + 1)) as Record<string, unknown>;

        return Object.entries(storedArgs).every(([key, value]) => {
            const known = ownedArgs.includes(key) && item && key in item ? item[key] : subscriptionVariables[key];

            if (known === undefined) {
                return true;
            }

            return String(value) === String(known);
        });
    } catch (error) {
        Log.error('Could not parse storeFieldName for subscription cache match; updating all variants', error);

        return true;
    }
};

type CacheActionApplier = (input: {
    existingArray: readonly Reference[];
    filterOutById: () => readonly Reference[];
    itemExists: boolean;
    newRef: Reference;
    order: ListOrder;
}) => readonly Reference[];

const cacheActionStrategies: Record<SubscriptionAction, CacheActionApplier> = {
    delete: ({ existingArray, filterOutById, itemExists }) => (itemExists ? filterOutById() : existingArray),
    insert: ({ existingArray, itemExists, newRef, order }) => {
        if (itemExists) {
            return existingArray;
        }

        return order === 'newestFirst' ? [newRef, ...existingArray] : [...existingArray, newRef];
    },
};

export const updateCacheForSubscription = (
    cache: InMemoryCache,
    subscriptionName: string,
    cacheField: string,
    newItem: { id: number | string },
    subscriptionVariables?: Record<string, unknown>,
): void => {
    if (!newItem?.id) {
        return;
    }

    if (subscriptionName === 'settingsUserUpdated') {
        try {
            cache.modify({
                fields: {
                    [cacheField]: () => newItem,
                },
            });
        } catch (error) {
            Log.error(`Error updating cache for ${subscriptionName}:`, {
                cacheField,
                error,
                itemId: newItem.id,
                subscriptionName,
            });
        }

        return;
    }

    try {
        cache.modify({
            fields: {
                [cacheField](existing, { readField, storeFieldName, toReference }) {
                    const existingArray = (existing ?? []) as readonly Reference[];

                    if (!matchesCacheVariant(storeFieldName, cacheField, subscriptionVariables, newItem)) {
                        return existingArray;
                    }

                    // ID equality must be type-tolerant: REST hydration writes some
                    // entries with `id` as a string (GraphQL `ID!` convention)
                    const targetId = String(newItem.id);
                    const idMatches = (ref: Reference) => String(readField('id', ref)) === targetId;

                    const itemExists = existingArray.some(idMatches);
                    const newRef = toReference(newItem as StoreObject, true);

                    if (!newRef) {
                        return existingArray;
                    }

                    const action = resolveSubscriptionAction(subscriptionName);

                    return cacheActionStrategies[action]({
                        existingArray,
                        filterOutById: () => existingArray.filter((ref) => !idMatches(ref)),
                        itemExists,
                        newRef,
                        order: orderOf(cacheField),
                    });
                },
            },
        });
    } catch (error) {
        Log.error(`Error updating cache for ${subscriptionName}:`, {
            cacheField,
            error,
            itemId: newItem.id,
        });
    }

    if (subscriptionName === 'flowDeleted') {
        cache.evict({ id: cache.identify({ __typename: 'Flow', id: newItem.id }) });
    }
};

export const createStreamingLink = (cache: InMemoryCache): ApolloLink => {
    const streamingLogs = new LRUCache<string, StreamingLogEntry>({
        max: STREAMING_CACHE_MAX_ENTRIES,
        ttl: STREAMING_CACHE_TTL_MS,
    });
    const reseedFromCache = new Set<string>();

    const baseFor = (cacheKey: string): StreamingLogEntry => {
        const entry = streamingLogs.get(cacheKey);

        if (entry) {
            return entry;
        }

        if (!reseedFromCache.delete(cacheKey)) {
            return EMPTY_LOG_ENTRY;
        }

        const row = cache.readFragment<Pick<AssistantLogFragmentFragment, 'message' | 'result' | 'thinking'>>({
            fragment: RESUMED_LOG_FRAGMENT,
            id: cacheKey,
        });

        return row
            ? {
                  lastUpdate: 0,
                  message: row.message ?? null,
                  result: row.result ?? null,
                  thinking: row.thinking ?? null,
              }
            : EMPTY_LOG_ENTRY;
    };

    const accumulateStreamingLog = (logUpdate: AssistantLogFragmentFragment): StreamingLogEntry => {
        const cacheKey = `${ASSISTANT_LOG_TYPENAME}:${logUpdate.id}`;
        const cachedLog = baseFor(cacheKey);

        const accumulatedLog: StreamingLogEntry = {
            lastUpdate: cachedLog.lastUpdate,
            message: concatStrings(cachedLog.message, logUpdate.message),
            result: concatStrings(cachedLog.result, logUpdate.result),
            thinking: concatStrings(cachedLog.thinking, logUpdate.thinking),
        };

        streamingLogs.set(cacheKey, accumulatedLog);

        return accumulatedLog;
    };

    const shouldEmitUpdate = (logId: string): boolean => {
        const entry = streamingLogs.get(`${ASSISTANT_LOG_TYPENAME}:${logId}`);

        if (!entry) {
            return true;
        }

        const now = Date.now();

        if (now - entry.lastUpdate >= STREAMING_THROTTLE_MS) {
            entry.lastUpdate = now;

            return true;
        }

        return false;
    };

    window.addEventListener('ws:reconnected', () => {
        for (const cacheKey of streamingLogs.keys()) {
            reseedFromCache.add(cacheKey);
        }

        streamingLogs.clear();
        localWriteAt.clear();
    });

    return new ApolloLink((operation, forward) => {
        return new Observable((observer) => {
            const accumulating = new Set<string>();
            const subscription = forward(operation).subscribe({
                complete: observer.complete.bind(observer),
                error: observer.error.bind(observer),
                next: (result) => {
                    const logUpdate = result.data?.assistantLogUpdated as AssistantLogFragmentFragment | undefined;

                    if (!logUpdate) {
                        observer.next(result);

                        return;
                    }

                    try {
                        if (logUpdate.appendPart && logUpdate.id) {
                            const accumulatedLog = accumulateStreamingLog(logUpdate);

                            accumulating.add(`${ASSISTANT_LOG_TYPENAME}:${logUpdate.id}`);

                            if (!shouldEmitUpdate(logUpdate.id)) {
                                return;
                            }

                            observer.next({
                                ...result,
                                data: {
                                    ...result.data,
                                    assistantLogUpdated: {
                                        ...logUpdate,
                                        appendPart: false,
                                        message: accumulatedLog.message ?? '',
                                        result: accumulatedLog.result ?? '',
                                        thinking: accumulatedLog.thinking,
                                    },
                                },
                            });

                            return;
                        }

                        if (logUpdate.id) {
                            accumulating.delete(`${ASSISTANT_LOG_TYPENAME}:${logUpdate.id}`);
                            streamingLogs.delete(`${ASSISTANT_LOG_TYPENAME}:${logUpdate.id}`);
                        }
                    } catch (error) {
                        Log.error('Error processing streaming assistant log:', error);
                    }

                    observer.next(result);
                },
            });

            return () => {
                for (const cacheKey of accumulating) {
                    streamingLogs.delete(cacheKey);
                }

                subscription.unsubscribe();
            };
        });
    });
};

export const createSocketPresence = () => {
    let liveSubscriptions = 0;
    let isReportedDown = false;

    return {
        closed: () => {
            if (liveSubscriptions > 0 && !isReportedDown) {
                isReportedDown = true;
                window.dispatchEvent(new Event('ws:disconnected'));
            }
        },
        connected: () => {
            if (isReportedDown) {
                isReportedDown = false;
                window.dispatchEvent(new Event('ws:connected'));
            }
        },
        link: new ApolloLink(
            (operation, forward) =>
                new Observable((observer) => {
                    liveSubscriptions += 1;

                    const subscription = forward(operation).subscribe({
                        complete: () => observer.complete(),
                        error: (error: unknown) => observer.error(error),
                        next: (result) => observer.next(result),
                    });

                    return () => {
                        liveSubscriptions = Math.max(0, liveSubscriptions - 1);

                        if (liveSubscriptions === 0 && isReportedDown) {
                            isReportedDown = false;
                            window.dispatchEvent(new Event('ws:connected'));
                        }

                        subscription.unsubscribe();
                    };
                }),
        ),
    };
};

export const createSubscriptionCacheLink = (cacheInstance: InMemoryCache): ApolloLink =>
    createInterceptLink((result, operation) => {
        if (result.data) {
            const variables = operation.variables as Record<string, unknown> | undefined;

            try {
                Object.entries(result.data)
                    .map(([key, value]) => ({ cacheField: subscriptionToCacheFieldMap[key], key, value }))
                    .filter(
                        (entry): entry is { cacheField: string; key: string; value: { id: number | string } } =>
                            !!entry.cacheField && !!entry.value?.id,
                    )
                    .forEach(({ cacheField, key, value }) => {
                        updateCacheForSubscription(cacheInstance, key, cacheField, value, variables);
                        localWriteAt.set(`${cacheField}:${String(value.id)}`, Date.now());
                    });
            } catch (error) {
                Log.error('Error processing subscription cache update:', error);
            }
        }

        return result;
    });

const replaceWithIncoming = {
    merge: (_existing: unknown, incoming: unknown) => incoming,
};

export const APPEND_ONLY_LIST_FIELDS = new Set([
    'agentLogs',
    'messageLogs',
    'screenshots',
    'searchLogs',
    'tasks',
    'terminalLogs',
    'vectorStoreLogs',
]);

const unionWithIncoming = (field: string) => ({
    merge: (existing: readonly Reference[] | undefined, incoming: null | readonly Reference[]) => {
        if (!incoming || !existing?.length) {
            return incoming ?? existing ?? null;
        }

        const arrived = new Set(incoming.map((reference) => reference.__ref));
        const missed = existing.filter((reference) => !arrived.has(reference.__ref));

        return missed.length ? placeMissedRows(field, incoming, missed) : incoming;
    },
});

export const queryKeyArgs = {
    agentLogs: ['flowId'],
    assistantLogs: ['flowId', 'assistantId'],
    assistants: ['flowId'],
    flowFiles: ['flowId'],
    knowledgeDocuments: ['filter', 'withContent'],
    messageLogs: ['flowId'],
    resources: ['path', 'recursive'],
    screenshots: ['flowId'],
    searchLogs: ['flowId'],
    tasks: ['flowId'],
    terminalLogs: ['flowId'],
    vectorStoreLogs: ['flowId'],
} as const satisfies Record<string, readonly string[]>;

const keyedListFields = Object.fromEntries(
    Object.entries(queryKeyArgs).map(([field, keyArgs]) => [
        field,
        { keyArgs, ...(APPEND_ONLY_LIST_FIELDS.has(field) ? unionWithIncoming(field) : replaceWithIncoming) },
    ]),
);

const keepLongerPrefix = (existing: null | string | undefined, incoming: null | string | undefined) =>
    typeof existing === 'string' &&
    typeof incoming === 'string' &&
    existing.length > incoming.length &&
    existing.startsWith(incoming)
        ? existing
        : incoming;

const preserveDeliveredResult = {
    message: {
        merge: keepLongerPrefix,
    },
    result: {
        merge: (existing: string | undefined, incoming: string) => incoming || existing || '',
    },
    resultFormat: {
        merge: (existing: string | undefined, incoming: string) =>
            incoming === 'plain' && existing ? existing : incoming,
    },
    thinking: {
        merge: keepLongerPrefix,
    },
};

export const createCache = () =>
    new InMemoryCache({
        typePolicies: {
            APIToken: {
                keyFields: ['tokenId'],
            },
            AssistantLog: {
                fields: { ...preserveDeliveredResult },
            },
            FlowFile: {
                keyFields: ['flowId', 'id'],
            },
            KnowledgeDocument: {
                fields: {
                    // `content` arrives empty from the list query (withContent:false,
                    // to save bandwidth) but full from the detail query — both
                    // normalize to this shared entity. Never let an empty incoming
                    // blank out a body the detail already loaded.
                    content: {
                        merge: (existing: string | undefined, incoming: string) => incoming || existing || '',
                    },
                },
            },
            MessageLog: {
                fields: { ...preserveDeliveredResult },
            },
            ProviderConfig: {
                keyFields: (object) => {
                    if (object.id === 0 || object.id === '0') {
                        return false;
                    }

                    return ['id'];
                },
            },
            Query: {
                fields: {
                    ...keyedListFields,
                    apiTokens: { ...replaceWithIncoming },
                    flow: {
                        read(existing, { args, toReference }) {
                            if (!args?.flowId) {
                                return existing;
                            }

                            return existing ?? toReference({ __typename: 'Flow', id: args.flowId });
                        },
                    },
                    flows: { ...replaceWithIncoming },
                    flowTemplates: { ...replaceWithIncoming },
                    providers: { ...replaceWithIncoming },
                    settingsPrompts: { ...replaceWithIncoming },
                    settingsProviders: { ...replaceWithIncoming },
                    settingsUser: { ...replaceWithIncoming },
                },
            },
        },
    });

export const watchQueryDefaults = {
    fetchPolicy: 'cache-and-network',
    nextFetchPolicy: 'cache-first',
    notifyOnNetworkStatusChange: true,
    refetchWritePolicy: 'merge',
} as const;

const WS_KEEPALIVE_MS = 20_000;

const createApolloClient = () => {
    // Holds the client for the ws `connected` handler, which is defined before the
    // client exists; `lazy: true` means the socket only opens on the first
    // subscription, after `.current` is set below.
    const clientRef: { current?: ApolloClient } = {};

    const httpLink = new HttpLink({
        credentials: 'include',
        fetch: fetchWithDeadline,
        uri: `${window.location.origin}${GRAPHQL_ENDPOINT}`,
    });

    const presence = createSocketPresence();

    const wsLink = new GraphQLWsLink(
        createClient({
            keepAlive: WS_KEEPALIVE_MS,
            lazy: true,
            on: {
                closed: () => {
                    Log.debug('GraphQL WebSocket closed');
                    presence.closed();
                },
                connected: (_socket, _payload, wasRetry) => {
                    Log.debug('GraphQL WebSocket connected');

                    presence.connected();

                    // Subscriptions are delta-only — the server never replays events
                    // published while we were disconnected — so on a reconnect refetch
                    // active queries to reconcile the cache with the backend.
                    if (wasRetry) {
                        // Unlike per-query refetch(), the aggregate promise isn't wrapped
                        // in preventUnhandledRejection — a failed reconcile (e.g. a transient
                        // 502 during the reconnect) would otherwise surface as an
                        // unhandledrejection.
                        void clientRef.current?.refetchObservableQueries().catch((error: unknown) => {
                            Log.error('Reconnect cache reconcile failed:', error);
                        });

                        // refetchObservableQueries skips cache-only queries, so the REST-hydrated
                        // resources slot isn't reconciled above — its provider re-fetches on this.
                        window.dispatchEvent(new Event('ws:reconnected'));
                    }
                },
                connecting: () => Log.debug('GraphQL WebSocket connecting...'),
                error: (error) => {
                    Log.error('GraphQL WebSocket error:', error);

                    // A WebSocket error event doesn't expose the handshake HTTP status, so a
                    // 403 can't be detected here — let /info classify it via auth:refresh.
                    window.dispatchEvent(new Event('auth:refresh'));
                },
                ping: () => Log.debug('GraphQL WebSocket ping'),
                pong: () => Log.debug('GraphQL WebSocket pong'),
            },
            retryAttempts: Infinity,
            retryWait: (retries) =>
                new Promise((resolve) => {
                    // Jitter so a mass reconnect (e.g. backend restart) doesn't thundering-herd the server.
                    setTimeout(resolve, Math.min(1000 * 2 ** retries, MAX_RETRY_DELAY_MS) + Math.random() * 3000);
                }),
            shouldRetry: () => true,
            url: `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}${GRAPHQL_ENDPOINT}`,
        }),
    );

    const transportLink = ApolloLink.split(isSubscriptionOperation, ApolloLink.from([presence.link, wsLink]), httpLink);

    const errorLink = new ErrorLink(({ error, operation }) => {
        if (CombinedGraphQLErrors.is(error)) {
            for (const { extensions, locations, message, path } of error.errors) {
                Log.error(`[GraphQL Error] ${message}`, {
                    locations,
                    operation: operation.operationName,
                    path,
                });

                const errorCode = extensions?.code as string | undefined;

                if (errorCode === 'UNAUTHENTICATED' || errorCode === 'FORBIDDEN') {
                    Log.warn('GraphQL authorization error detected, refreshing auth info');
                    window.dispatchEvent(new Event('auth:refresh'));
                }
            }
        } else if (error) {
            Log.error(`[Network Error] ${error.message}`, error);

            const statusCode = ServerError.is(error) || ServerParseError.is(error) ? error.statusCode : undefined;

            if (statusCode === 401 || statusCode === 403) {
                Log.warn('Network authorization error detected, refreshing auth info');
                window.dispatchEvent(new Event('auth:refresh'));
            }
        }
    });

    const cache = createCache();

    const streamingLink = createStreamingLink(cache);
    const subscriptionCacheLink = createSubscriptionCacheLink(cache);

    const link = ApolloLink.from([errorLink, subscriptionCacheLink, streamingLink, transportLink]);

    const apolloClient = new ApolloClient({
        cache,
        defaultOptions: { watchQuery: watchQueryDefaults },
        link,
    });

    clientRef.current = apolloClient;

    return apolloClient;
};

export const client = createApolloClient();

export default client;
