import { ApolloClient, createHttpLink, InMemoryCache, split } from '@apollo/client';
import { GraphQLWsLink } from '@apollo/client/link/subscriptions';
import { getMainDefinition } from '@apollo/client/utilities';
import { createClient } from 'graphql-ws';

// Relative URL works through both Vite dev proxy and nginx in production
const NEO4J_GRAPHQL_URL = import.meta.env.VITE_NEO4J_GRAPHQL_URL ?? '/graphql';

const NEO4J_GRAPHQL_WS_URL =
    import.meta.env.VITE_NEO4J_GRAPHQL_WS_URL ??
    `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/graphql`;

const httpLink = createHttpLink({
    uri: NEO4J_GRAPHQL_URL,
});

const wsLink = new GraphQLWsLink(
    createClient({
        on: {
            closed: (event) => console.log('[Neo4j WS] Closed', event),
            connected: () => console.log('[Neo4j WS] Connected to', NEO4J_GRAPHQL_WS_URL),
            error: (error) => console.error('[Neo4j WS] Error', error),
        },
        retryAttempts: 5,
        retryWait: (retries) =>
            new Promise((resolve) => {
                const timeout = Math.min(1000 * 2 ** retries, 10000);
                setTimeout(() => resolve(), timeout);
            }),
        shouldRetry: () => true,
        url: NEO4J_GRAPHQL_WS_URL,
    }),
);

const splitLink = split(
    ({ query }) => {
        const definition = getMainDefinition(query);

        return definition.kind === 'OperationDefinition' && definition.operation === 'subscription';
    },
    wsLink,
    httpLink,
);

export const neo4jClient = new ApolloClient({
    cache: new InMemoryCache(),
    defaultOptions: {
        watchQuery: {
            fetchPolicy: 'network-only',
        },
    },
    link: splitLink,
});
