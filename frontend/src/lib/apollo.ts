import type { FetchResult, Operation, Reference, StoreObject } from '@apollo/client';

import { ApolloClient, ApolloLink, createHttpLink, InMemoryCache, Observable, split } from '@apollo/client';
import { onError } from '@apollo/client/link/error';
import { GraphQLWsLink } from '@apollo/client/link/subscriptions';
import { getMainDefinition } from '@apollo/client/utilities';
import { createClient } from 'graphql-ws';
import { LRUCache } from 'lru-cache';

import type { AssistantLogFragmentFragment } from '@/graphql/types';

import { Log } from '@/lib/log';
import { baseUrl } from '@/models/api';

// --- Constants ---

const GRAPHQL_ENDPOINT = `${baseUrl}/graphql`;
const ASSISTANT_LOG_TYPENAME = 'AssistantLog';
const MAX_RETRY_DELAY_MS = 30_000;
const STREAMING_CACHE_MAX_ENTRIES = 500;
const STREAMING_CACHE_TTL_MS = 1000 * 60 * 5;

// --- Types ---

type StreamingLogEntry = {
    message: null | string;
    result: null | string;
    thinking: null | string;
};

type SubscriptionAction = 'add' | 'create' | 'delete' | 'update';

// --- Pure utilities ---

const EMPTY_LOG_ENTRY: StreamingLogEntry = { message: null, result: null, thinking: null };

const concatStrings = (existing: null | string | undefined, incoming: null | string | undefined): null | string =>
    existing && incoming ? `${existing}${incoming}` : (existing ?? incoming ?? null);

const resolveSubscriptionAction = (name: string): SubscriptionAction => {
    if (name.endsWith('Deleted')) {
        return 'delete';
    }

    if (name.endsWith('Updated')) {
        return 'update';
    }

    if (name.endsWith('Created')) {
        return 'create';
    }

    return 'add';
};

const isSubscriptionOperation = ({ query }: Operation): boolean => {
    const definition = getMainDefinition(query);

    return definition.kind === 'OperationDefinition' && definition.operation === 'subscription';
};

// --- Link helpers ---

const createInterceptLink = (
    transform: (result: FetchResult, operation: Operation) => FetchResult,
): ApolloLink =>
    new ApolloLink(
        (operation: Operation, forward) =>
            new Observable((observer) => {
                const subscription = forward(operation).subscribe({
                    complete: observer.complete.bind(observer),
                    error: observer.error.bind(observer),
                    next: (result: FetchResult) => observer.next(transform(result, operation)),
                });

                return () => subscription.unsubscribe();
            }),
    );

// --- Subscription → cache configuration ---

const subscriptionToCacheFieldMap: Record<string, string> = {
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
    flowUpdated: 'flows',
    messageLogAdded: 'messageLogs',
    messageLogUpdated: 'messageLogs',
    providerCreated: 'settingsProviders',
    providerDeleted: 'settingsProviders',
    providerUpdated: 'settingsProviders',
    screenshotAdded: 'screenshots',
    searchLogAdded: 'searchLogs',
    settingsUserUpdated: 'settingsUser',
    taskCreated: 'tasks',
    taskUpdated: 'tasks',
    terminalLogAdded: 'terminalLogs',
    vectorStoreLogAdded: 'vectorStoreLogs',
};

// --- Cache variant matching ---

const matchesCacheVariant = (
    storeFieldName: string,
    cacheField: string,
    subscriptionVariables?: Record<string, unknown>,
): boolean => {
    if (!subscriptionVariables || storeFieldName === cacheField) {
        return true;
    }

    const separatorIndex = storeFieldName.indexOf(':');

    if (separatorIndex === -1) {
        return true;
    }

    try {
        const storedArgs = JSON.parse(storeFieldName.slice(separatorIndex + 1)) as Record<string, unknown>;

        return Object.entries(storedArgs).every(([key, value]) => {
            if (!(key in subscriptionVariables)) {
                return true;
            }

            return String(value) === String(subscriptionVariables[key]);
        });
    } catch {
        return true;
    }
};

// --- Cache action strategies ---

type CacheActionApplier = (
    existingArray: readonly Reference[],
    newRef: Reference,
    itemExists: boolean,
    filterOutById: () => readonly Reference[],
) => readonly Reference[];

const cacheActionStrategies: Record<SubscriptionAction, CacheActionApplier> = {
    add: (existingArray, newRef, itemExists) => (itemExists ? existingArray : [...existingArray, newRef]),
    create: (existingArray, newRef, itemExists) => (itemExists ? existingArray : [newRef, ...existingArray]),
    delete: (existingArray, _newRef, itemExists, filterOutById) => (itemExists ? filterOutById() : existingArray),
    update: (existingArray) => existingArray,
};

const updateCacheForSubscription = (
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

                    if (!matchesCacheVariant(storeFieldName, cacheField, subscriptionVariables)) {
                        return existingArray;
                    }

                    const newRef = toReference(newItem as StoreObject, true);

                    if (!newRef) {
                        return existingArray;
                    }

                    const itemExists = existingArray.some((ref) => readField('id', ref) === newItem.id);
                    const action = resolveSubscriptionAction(subscriptionName);

                    Log.debug(
                        `[CacheUpdate] ${subscriptionName} | id: ${newItem.id} | exists: ${itemExists} | count: ${existingArray.length}`,
                    );

                    return cacheActionStrategies[action](existingArray, newRef, itemExists, () =>
                        existingArray.filter((ref) => readField('id', ref) !== newItem.id),
                    );
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
};

// --- Link factories ---

const createStreamingLink = (): ApolloLink => {
    const streamingLogs = new LRUCache<string, StreamingLogEntry>({
        max: STREAMING_CACHE_MAX_ENTRIES,
        ttl: STREAMING_CACHE_TTL_MS,
    });

    const accumulateStreamingLog = (logUpdate: AssistantLogFragmentFragment): StreamingLogEntry => {
        const cacheKey = `${ASSISTANT_LOG_TYPENAME}:${logUpdate.id}`;
        const cachedLog = streamingLogs.get(cacheKey) ?? EMPTY_LOG_ENTRY;

        const accumulatedLog: StreamingLogEntry = {
            message: concatStrings(cachedLog.message, logUpdate.message),
            result: concatStrings(cachedLog.result, logUpdate.result),
            thinking: concatStrings(cachedLog.thinking, logUpdate.thinking),
        };

        streamingLogs.set(cacheKey, accumulatedLog);

        return accumulatedLog;
    };

    return createInterceptLink((result, _operation) => {
        const logUpdate = result.data?.assistantLogUpdated as AssistantLogFragmentFragment | undefined;

        if (!logUpdate) {
            return result;
        }

        try {
            if (logUpdate.appendPart && logUpdate.id) {
                const accumulatedLog = accumulateStreamingLog(logUpdate);

                return {
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
                };
            }

            if (logUpdate.id) {
                streamingLogs.delete(`${ASSISTANT_LOG_TYPENAME}:${logUpdate.id}`);
            }
        } catch (error) {
            Log.error('Error processing streaming assistant log:', error);
        }

        return result;
    });
};

const createSubscriptionCacheLink = (cacheInstance: InMemoryCache): ApolloLink =>
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
                    });
            } catch (error) {
                Log.error('Error processing subscription cache update:', error);
            }
        }

        return result;
    });

// --- Cache merge policy ---

const replaceWithIncoming = {
    merge: (_existing: unknown, incoming: unknown) => incoming,
};

// --- Client factory ---

const createApolloClient = () => {
    const httpLink = createHttpLink({
        credentials: 'include',
        uri: `${window.location.origin}${GRAPHQL_ENDPOINT}`,
    });

    const wsLink = new GraphQLWsLink(
        createClient({
            lazy: true,
            on: {
                closed: () => Log.debug('GraphQL WebSocket closed'),
                connected: () => Log.debug('GraphQL WebSocket connected'),
                connecting: () => Log.debug('GraphQL WebSocket connecting...'),
                error: (error) => Log.error('GraphQL WebSocket error:', error),
                ping: () => Log.debug('GraphQL WebSocket ping'),
                pong: () => Log.debug('GraphQL WebSocket pong'),
            },
            retryAttempts: Infinity,
            retryWait: (retries) =>
                new Promise((resolve) => {
                    setTimeout(resolve, Math.min(1000 * 2 ** retries, MAX_RETRY_DELAY_MS));
                }),
            shouldRetry: () => true,
            url: `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}${GRAPHQL_ENDPOINT}`,
        }),
    );

    const transportLink = split(isSubscriptionOperation, wsLink, httpLink);

    const errorLink = onError(({ graphQLErrors, networkError, operation }) => {
        if (graphQLErrors) {
            for (const { locations, message, path } of graphQLErrors) {
                Log.error(`[GraphQL Error] ${message}`, {
                    locations,
                    operation: operation.operationName,
                    path,
                });
            }
        }

        if (networkError) {
            Log.error(`[Network Error] ${networkError.message}`, networkError);
        }
    });

    const cache = new InMemoryCache({
        typePolicies: {
            APIToken: {
                keyFields: ['tokenId'],
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
                    agentLogs: { keyArgs: ['flowId'], ...replaceWithIncoming },
                    apiTokens: { ...replaceWithIncoming },
                    assistantLogs: { keyArgs: ['flowId', 'assistantId'], ...replaceWithIncoming },
                    assistants: { keyArgs: ['flowId'], ...replaceWithIncoming },
                    flow: {
                        read(existing, { args, toReference }) {
                            if (!args?.flowId) {
                                return existing;
                            }

                            return existing ?? toReference({ __typename: 'Flow', id: args.flowId });
                        },
                    },
                    flows: { ...replaceWithIncoming },
                    messageLogs: { keyArgs: ['flowId'], ...replaceWithIncoming },
                    providers: { ...replaceWithIncoming },
                    screenshots: { keyArgs: ['flowId'], ...replaceWithIncoming },
                    searchLogs: { keyArgs: ['flowId'], ...replaceWithIncoming },
                    settingsPrompts: { ...replaceWithIncoming },
                    settingsProviders: { ...replaceWithIncoming },
                    settingsUser: { ...replaceWithIncoming },
                    tasks: { keyArgs: ['flowId'], ...replaceWithIncoming },
                    terminalLogs: { keyArgs: ['flowId'], ...replaceWithIncoming },
                    vectorStoreLogs: { keyArgs: ['flowId'], ...replaceWithIncoming },
                },
            },
        },
    });

    const streamingLink = createStreamingLink();
    const subscriptionCacheLink = createSubscriptionCacheLink(cache);

    const link = ApolloLink.from([errorLink, subscriptionCacheLink, streamingLink, transportLink]);

    return new ApolloClient({
        cache,
        defaultOptions: {
            watchQuery: {
                fetchPolicy: 'cache-and-network',
                nextFetchPolicy: 'cache-first',
                notifyOnNetworkStatusChange: true,
            },
        },
        link,
    });
};

export const client = createApolloClient();

export default client;
