/** @typedef {import('./types.ts').CypherQuery} CypherQuery */
/** @typedef {import('./types.ts').Neo4jNode} Neo4jNode */
/** @typedef {import('./types.ts').Neo4jRelationship} Neo4jRelationship */
/** @typedef {import('./types.ts').Neo4jPath} Neo4jPath */

import { ApolloServer } from '@apollo/server';
import { ApolloServerPluginDrainHttpServer } from '@apollo/server/plugin/drainHttpServer';
import { expressMiddleware } from '@as-integrations/express4';
import cors from 'cors';
import 'dotenv/config';
import express from 'express';
import { GraphQLScalarType, Kind } from 'graphql';
import { PubSub, withFilter } from 'graphql-subscriptions';
import { useServer } from 'graphql-ws/use/ws';
import { makeExecutableSchema } from '@graphql-tools/schema';
import http from 'http';
import neo4j from 'neo4j-driver';
import { WebSocketServer } from 'ws';
import {
    accessChain,
    allCves,
    artifactsProduced,
    attackPathStats,
    attackSurfaceOverview,
    credentialsStatus,
    exploitableVulnerabilities,
    fullAttackChain,
    hostsWithServices,
    infrastructureOnly,
    mainAttackChain,
    openPortsWithServices,
    pentestSummary,
    shortestPath,
    toolEffectiveness,
    toolUsageStats,
    validAccessDetails,
    vulnerabilitySeverity,
} from './queries.js';
import { typeDefs } from './schema.js';

// ---------------------------------------------------------------------------
// Configuration
// ---------------------------------------------------------------------------

const PORT = Number(process.env.PORT) || 4000;
const NEO4J_URI = process.env.NEO4J_URI || 'bolt://localhost:7687';
const NEO4J_USER = process.env.NEO4J_USER || 'neo4j';
const NEO4J_PASSWORD = process.env.NEO4J_PASSWORD || 'neo4j';
const NEO4J_DATABASE = process.env.NEO4J_DATABASE || 'neo4j';
const POLL_INTERVAL_MS = Number(process.env.POLL_INTERVAL_MS) || 5000;

const encryptionInUri = NEO4J_URI.includes('+s://');
const isLocalUri = NEO4J_URI.includes('localhost') || NEO4J_URI.includes('127.0.0.1');
const encrypted =
    process.env.NEO4J_ENCRYPTED === 'true'
        ? 'ENCRYPTION_ON'
        : process.env.NEO4J_ENCRYPTED === 'false'
          ? 'ENCRYPTION_OFF'
          : isLocalUri
            ? 'ENCRYPTION_OFF'
            : 'ENCRYPTION_ON';

const driverConfig = encryptionInUri ? {} : { encrypted };
const driver = neo4j.driver(NEO4J_URI, neo4j.auth.basic(NEO4J_USER, NEO4J_PASSWORD), driverConfig);

// ---------------------------------------------------------------------------
// PubSub for subscriptions
// ---------------------------------------------------------------------------

const pubsub = new PubSub();

const SUBSCRIPTION_EVENTS = {
    DASHBOARD_UPDATED: 'DASHBOARD_UPDATED',
    ATTACK_SURFACE_UPDATED: 'ATTACK_SURFACE_UPDATED',
    MAIN_ATTACK_CHAIN_UPDATED: 'MAIN_ATTACK_CHAIN_UPDATED',
    FULL_ATTACK_CHAIN_UPDATED: 'FULL_ATTACK_CHAIN_UPDATED',
    ATTACK_PATH_STATS_UPDATED: 'ATTACK_PATH_STATS_UPDATED',
    VULNERABILITY_SEVERITY_UPDATED: 'VULNERABILITY_SEVERITY_UPDATED',
};

// Mapping from snapshot key to the subscription event and payload field name.
// Used by the polling engine to publish only changed datasets.
const SNAPSHOT_PUBLISH_CONFIG = [
    { key: 'dashboard', event: SUBSCRIPTION_EVENTS.DASHBOARD_UPDATED, field: 'dashboardUpdated' },
    { key: 'attackSurface', event: SUBSCRIPTION_EVENTS.ATTACK_SURFACE_UPDATED, field: 'attackSurfaceUpdated' },
    { key: 'mainChain', event: SUBSCRIPTION_EVENTS.MAIN_ATTACK_CHAIN_UPDATED, field: 'mainAttackChainUpdated' },
    { key: 'fullChain', event: SUBSCRIPTION_EVENTS.FULL_ATTACK_CHAIN_UPDATED, field: 'fullAttackChainUpdated' },
    { key: 'pathStats', event: SUBSCRIPTION_EVENTS.ATTACK_PATH_STATS_UPDATED, field: 'attackPathStatsUpdated' },
    {
        key: 'vulnSeverity',
        event: SUBSCRIPTION_EVENTS.VULNERABILITY_SEVERITY_UPDATED,
        field: 'vulnerabilitySeverityUpdated',
    },
];

// ---------------------------------------------------------------------------
// Neo4j helpers
// ---------------------------------------------------------------------------

/**
 * Recursively convert Neo4j driver objects to plain JSON.
 * Handles Node, Relationship, Path, Integer, and nested objects.
 * @param {unknown} value
 * @returns {Neo4jNode | Neo4jRelationship | Neo4jPath | number | unknown}
 */
function toJson(value) {
    if (value == null) return value;
    if (neo4j.isInt(value)) return value.toNumber();
    if (Array.isArray(value)) return value.map(toJson);
    if (value instanceof neo4j.types.Node) {
        return { labels: value.labels, properties: toJson(value.properties) };
    }
    if (value instanceof neo4j.types.Relationship) {
        return { type: value.type, properties: toJson(value.properties) };
    }
    if (value instanceof neo4j.types.Path) {
        return {
            start: toJson(value.start),
            end: toJson(value.end),
            segments: value.segments.map((segment) => ({
                start: toJson(segment.start),
                relationship: toJson(segment.relationship),
                end: toJson(segment.end),
            })),
        };
    }
    if (typeof value === 'object' && !(value instanceof Date)) {
        return Object.fromEntries(
            Object.entries(value)
                .filter(([key]) => !key.endsWith('_embedding'))
                .map(([key, v]) => [key, toJson(v)]),
        );
    }
    return value;
}

/**
 * Run a Cypher query and return records as array of plain objects.
 * @param {CypherQuery} query
 * @returns {Promise<Record<string, unknown>[]>}
 */
async function runQuery({ text, params }) {
    const session = driver.session({ database: NEO4J_DATABASE });
    try {
        const result = await session.run(text, params);
        return result.records.map((record) =>
            Object.fromEntries(record.keys.map((key) => [key, toJson(record.get(key))])),
        );
    } finally {
        await session.close();
    }
}

/**
 * Run a Cypher query and return the raw Neo4j Result object.
 * @param {CypherQuery} query
 * @returns {Promise<import('neo4j-driver').Result>}
 */
async function runRawQuery({ text, params }) {
    const session = driver.session({ database: NEO4J_DATABASE });
    try {
        return await session.run(text, params);
    } finally {
        await session.close();
    }
}

/**
 * Transform raw Neo4j RETURN * result into graph format for visualization.
 * Extracts unique nodes and relationships from all records.
 * @param {import('neo4j-driver').Result} result
 * @returns {{ source: object, target: object, relationType: string }[]}
 */
function transformToGraphData(result) {
    const allValues = result.records.flatMap((record) => record.keys.map((key) => record.get(key)).filter(Boolean));

    const nodesMap = new Map(
        allValues
            .filter((value) => value instanceof neo4j.types.Node)
            .map((node) => [node.identity.toString(), { labels: node.labels, properties: toJson(node.properties) }]),
    );

    const uniqueRelationships = new Map(
        allValues
            .filter((value) => value instanceof neo4j.types.Relationship)
            .map((relationship) => [relationship.identity.toString(), relationship]),
    );

    return [...uniqueRelationships.values()]
        .filter(
            (relationship) => nodesMap.has(relationship.start.toString()) && nodesMap.has(relationship.end.toString()),
        )
        .map((relationship) => ({
            source: nodesMap.get(relationship.start.toString()),
            target: nodesMap.get(relationship.end.toString()),
            relationType: relationship.type,
        }));
}

// ---------------------------------------------------------------------------
// Query helper functions (used by both resolvers and polling)
// ---------------------------------------------------------------------------

async function fetchDashboard(groupId) {
    const records = await runQuery(pentestSummary(groupId));
    const data = records[0] ?? {};
    return {
        groupId,
        hosts: data.Hosts ?? 0,
        ports: data.Ports ?? 0,
        services: data.Services ?? 0,
        vulnerabilities: data.Vulnerabilities ?? 0,
        accounts: data.Accounts ?? 0,
        validAccess: data.ValidAccess ?? 0,
        status: data.Status ?? 'SECURE',
    };
}

async function fetchAttackSurface(groupId) {
    const records = await runQuery(attackSurfaceOverview(groupId));
    return records
        .filter((record) => record.EntityType != null)
        .map((record) => ({
            entityType: record.EntityType,
            count: record.Count,
        }));
}

async function fetchGraphData(groupId, queryFunction) {
    const result = await runRawQuery(queryFunction(groupId));
    const data = transformToGraphData(result);
    return { groupId, rows: data.length, data };
}

async function fetchAttackPathStats(groupId) {
    const records = await runQuery(attackPathStats(groupId));
    const data = records[0] ?? {};
    return {
        groupId,
        hosts: data.Hosts ?? 0,
        ports: data.Ports ?? 0,
        services: data.Services ?? 0,
        vulnerabilities: data.Vulnerabilities ?? 0,
        validAccess: data.ValidAccess ?? 0,
        accounts: data.Accounts ?? 0,
    };
}

async function fetchVulnerabilitySeverity(groupId) {
    const records = await runQuery(vulnerabilitySeverity(groupId));
    return records.map((record) => ({
        category: record.Category,
        count: record.Count,
        examples: record.Examples ?? [],
    }));
}

// ---------------------------------------------------------------------------
// Custom JSON scalar
// ---------------------------------------------------------------------------

/** @param {import('graphql').ValueNode} ast */
function parseLiteralValue(ast) {
    switch (ast.kind) {
        case Kind.STRING:
            try {
                return JSON.parse(ast.value);
            } catch {
                return ast.value;
            }
        case Kind.INT:
            return parseInt(ast.value, 10);
        case Kind.FLOAT:
            return parseFloat(ast.value);
        case Kind.BOOLEAN:
            return ast.value;
        case Kind.NULL:
            return null;
        case Kind.OBJECT:
            return Object.fromEntries(ast.fields.map((field) => [field.name.value, parseLiteralValue(field.value)]));
        case Kind.LIST:
            return ast.values.map(parseLiteralValue);
        default:
            return undefined;
    }
}

const JSONScalar = new GraphQLScalarType({
    name: 'JSON',
    description: 'Arbitrary JSON value',
    serialize: (value) => value,
    parseValue: (value) => value,
    parseLiteral: parseLiteralValue,
});

// ---------------------------------------------------------------------------
// Resolvers
// ---------------------------------------------------------------------------

const resolvers = {
    JSON: JSONScalar,

    Query: {
        health: () => ({ status: 'ok', neo4j: NEO4J_URI }),

        // Dashboard
        dashboard: (_parent, { groupId }) => fetchDashboard(groupId),

        // Attack Surface (Q1)
        attackSurface: (_parent, { groupId }) => fetchAttackSurface(groupId),

        // Credentials & Access (Q2)
        credentialsStatus: async (_parent, { groupId }) => {
            const records = await runQuery(credentialsStatus(groupId));
            return records.map((record) => ({
                status: record.Status,
                count: record.Count,
                examples: record.Examples ?? [],
            }));
        },

        accessDetails: async (_parent, { groupId }) => {
            const records = await runQuery(validAccessDetails(groupId));
            return records.map((record) => ({
                access: record.Access,
                account: record.Account,
                host: record.Host,
                service: record.Service,
                summary: record.Summary,
            }));
        },

        // Infrastructure (Q3)
        hostsWithServices: async (_parent, { groupId }) => {
            const records = await runQuery(hostsWithServices(groupId));
            return records.map((record) => ({
                host: record.Host,
                ports: record.Ports ?? [],
                services: record.Services ?? [],
            }));
        },

        openPorts: async (_parent, { groupId }) => {
            const records = await runQuery(openPortsWithServices(groupId));
            return records.map((record) => ({
                port: record.Port,
                service: record.Service,
                host: record.Host,
            }));
        },

        // Vulnerabilities (Q4)
        vulnerabilitySeverity: (_parent, { groupId }) => fetchVulnerabilitySeverity(groupId),

        allCves: async (_parent, { groupId }) => {
            const records = await runQuery(allCves(groupId));
            return records.map((record) => ({
                cve: record.CVE,
                foundOn: record.FoundOn,
                source: record.Source,
            }));
        },

        exploitAttempts: async (_parent, { groupId }) => {
            const records = await runQuery(exploitableVulnerabilities(groupId));
            return records.map((record) => ({
                vulnerability: record.Vulnerability,
                attemptCount: record.AttemptCount,
                status: record.Status,
            }));
        },

        // Tools (Q5)
        toolUsage: async (_parent, { groupId }) => {
            const records = await runQuery(toolUsageStats(groupId));
            return records.map((record) => ({
                tool: record.Tool,
                executions: record.Executions,
            }));
        },

        toolEffectiveness: async (_parent, { groupId }) => {
            const records = await runQuery(toolEffectiveness(groupId));
            return records.map((record) => ({
                tool: record.Tool,
                executions: record.Executions,
                discoveries: record.Discoveries,
                discoveryTypes: record.DiscoveryTypes ?? [],
            }));
        },

        artifacts: async (_parent, { groupId }) => {
            const records = await runQuery(artifactsProduced(groupId));
            return records.map((record) => ({
                artifact: record.Artifact,
                producedBy: record.ProducedBy,
                summary: record.Summary,
            }));
        },

        // Graph visualizations (Q6)
        mainAttackChain: (_parent, { groupId }) => fetchGraphData(groupId, mainAttackChain),
        fullAttackChain: (_parent, { groupId }) => fetchGraphData(groupId, fullAttackChain),
        infrastructureGraph: (_parent, { groupId }) => fetchGraphData(groupId, infrastructureOnly),
        accessChainGraph: (_parent, { groupId }) => fetchGraphData(groupId, accessChain),
        shortestPathGraph: (_parent, { groupId }) => fetchGraphData(groupId, shortestPath),

        // Attack Path Stats (Q7)
        attackPathStats: (_parent, { groupId }) => fetchAttackPathStats(groupId),
    },

    Subscription: {
        dashboardUpdated: {
            subscribe: withFilter(
                () => pubsub.asyncIterableIterator(SUBSCRIPTION_EVENTS.DASHBOARD_UPDATED),
                (payload, variables) => payload.groupId === variables.groupId,
            ),
            resolve: (payload) => payload.dashboardUpdated,
        },

        attackSurfaceUpdated: {
            subscribe: withFilter(
                () => pubsub.asyncIterableIterator(SUBSCRIPTION_EVENTS.ATTACK_SURFACE_UPDATED),
                (payload, variables) => payload.groupId === variables.groupId,
            ),
            resolve: (payload) => payload.attackSurfaceUpdated,
        },

        mainAttackChainUpdated: {
            subscribe: withFilter(
                () => pubsub.asyncIterableIterator(SUBSCRIPTION_EVENTS.MAIN_ATTACK_CHAIN_UPDATED),
                (payload, variables) => payload.groupId === variables.groupId,
            ),
            resolve: (payload) => payload.mainAttackChainUpdated,
        },

        fullAttackChainUpdated: {
            subscribe: withFilter(
                () => pubsub.asyncIterableIterator(SUBSCRIPTION_EVENTS.FULL_ATTACK_CHAIN_UPDATED),
                (payload, variables) => payload.groupId === variables.groupId,
            ),
            resolve: (payload) => payload.fullAttackChainUpdated,
        },

        attackPathStatsUpdated: {
            subscribe: withFilter(
                () => pubsub.asyncIterableIterator(SUBSCRIPTION_EVENTS.ATTACK_PATH_STATS_UPDATED),
                (payload, variables) => payload.groupId === variables.groupId,
            ),
            resolve: (payload) => payload.attackPathStatsUpdated,
        },

        vulnerabilitySeverityUpdated: {
            subscribe: withFilter(
                () => pubsub.asyncIterableIterator(SUBSCRIPTION_EVENTS.VULNERABILITY_SEVERITY_UPDATED),
                (payload, variables) => payload.groupId === variables.groupId,
            ),
            resolve: (payload) => payload.vulnerabilitySeverityUpdated,
        },
    },
};

// ---------------------------------------------------------------------------
// Polling engine — detects changes and publishes subscription events
// ---------------------------------------------------------------------------

const previousSnapshots = new Map();
const activeGroupIds = new Set();

/**
 * Register a groupId for polling. Called when a subscription starts.
 * @param {string} groupId
 */
function registerGroupForPolling(groupId) {
    activeGroupIds.add(groupId);
    console.log(`[Polling] Registered group: ${groupId} (active: ${activeGroupIds.size})`);
}

/**
 * Unregister a groupId from polling. Called when no more subscriptions remain.
 * @param {string} groupId
 */
function unregisterGroupForPolling(groupId) {
    activeGroupIds.delete(groupId);
    previousSnapshots.delete(groupId);
    console.log(`[Polling] Unregistered group: ${groupId} (active: ${activeGroupIds.size})`);
}

/**
 * Poll a single group: fetch all datasets, compare with previous snapshot,
 * and publish events for any that changed.
 * @param {string} groupId
 */
async function pollGroup(groupId) {
    const [dashboard, attackSurface, mainChain, fullChain, pathStats, vulnSeverity] = await Promise.all([
        fetchDashboard(groupId),
        fetchAttackSurface(groupId),
        fetchGraphData(groupId, mainAttackChain),
        fetchGraphData(groupId, fullAttackChain),
        fetchAttackPathStats(groupId),
        fetchVulnerabilitySeverity(groupId),
    ]);

    const fetchedData = { dashboard, attackSurface, mainChain, fullChain, pathStats, vulnSeverity };

    const currentSnapshot = Object.fromEntries(
        Object.entries(fetchedData).map(([key, value]) => [key, JSON.stringify(value)]),
    );

    const previous = previousSnapshots.get(groupId);

    if (previous) {
        SNAPSHOT_PUBLISH_CONFIG.filter(({ key }) => previous[key] !== currentSnapshot[key]).forEach(
            ({ key, event, field }) => {
                pubsub.publish(event, { groupId, [field]: fetchedData[key] });
            },
        );
    }

    previousSnapshots.set(groupId, currentSnapshot);
}

/**
 * Poll Neo4j for each active group in parallel and publish changes.
 */
async function pollForChanges() {
    const groupIds = [...activeGroupIds];
    const results = await Promise.allSettled(groupIds.map(pollGroup));

    results
        .filter((result) => result.status === 'rejected')
        .forEach((result) => {
            console.error('[Polling] Error polling group:', result.reason?.message ?? result.reason);
        });
}

let pollTimeoutId = null;

function startPolling() {
    if (pollTimeoutId !== null) return;
    console.log(`[Polling] Started with interval ${POLL_INTERVAL_MS}ms`);
    schedulePollCycle();
}

function stopPolling() {
    if (pollTimeoutId === null) return;
    clearTimeout(pollTimeoutId);
    pollTimeoutId = null;
    console.log('[Polling] Stopped');
}

/** Schedule the next poll cycle after the current one completes (avoids overlapping polls). */
function schedulePollCycle() {
    pollTimeoutId = setTimeout(async () => {
        try {
            await pollForChanges();
        } catch (error) {
            console.error('[Polling] Unexpected error:', error.message);
        }
        if (pollTimeoutId !== null) {
            schedulePollCycle();
        }
    }, POLL_INTERVAL_MS);
}

// ---------------------------------------------------------------------------
// Per-connection subscription tracking
// ---------------------------------------------------------------------------

// Maps each WebSocket instance to its active subscriptions (subscriptionId → groupId)
const connectionSubscriptions = new WeakMap();

// Global count of subscriptions per groupId (across all connections)
const groupSubscriptionCounts = new Map();

/**
 * Get or create the subscriptions map for a given connection context.
 * @param {object} context - graphql-ws connection context (ctx.extra is the raw WebSocket)
 * @returns {Map<string, string>} subscriptionId → groupId
 */
function getConnectionSubscriptions(context) {
    let subscriptions = connectionSubscriptions.get(context.extra);
    if (!subscriptions) {
        subscriptions = new Map();
        connectionSubscriptions.set(context.extra, subscriptions);
    }
    return subscriptions;
}

/**
 * Decrement the global subscription count for a groupId.
 * Unregisters the group from polling when count reaches zero.
 * @param {string} groupId
 */
function decrementGroupSubscription(groupId) {
    const count = groupSubscriptionCounts.get(groupId) ?? 0;
    if (count <= 1) {
        groupSubscriptionCounts.delete(groupId);
        unregisterGroupForPolling(groupId);
    } else {
        groupSubscriptionCounts.set(groupId, count - 1);
    }
    if (!groupSubscriptionCounts.size) stopPolling();
}

// ---------------------------------------------------------------------------
// Server setup
// ---------------------------------------------------------------------------

const app = express();
const httpServer = http.createServer(app);

const schema = makeExecutableSchema({ typeDefs, resolvers });

const wsServer = new WebSocketServer({
    server: httpServer,
    path: '/graphql',
});

const serverCleanup = useServer(
    {
        schema,
        onSubscribe: (context, message) => {
            const groupId = message.payload?.variables?.groupId;
            if (!groupId) return;

            const subscriptions = getConnectionSubscriptions(context);
            subscriptions.set(message.id, groupId);

            const count = (groupSubscriptionCounts.get(groupId) ?? 0) + 1;
            groupSubscriptionCounts.set(groupId, count);
            registerGroupForPolling(groupId);
            if (pollTimeoutId === null) startPolling();
        },

        onComplete: (context, message) => {
            const subscriptions = connectionSubscriptions.get(context.extra);
            if (!subscriptions) return;

            const groupId = subscriptions.get(message.id);
            if (!groupId) return;

            subscriptions.delete(message.id);
            decrementGroupSubscription(groupId);
        },

        onDisconnect: (context) => {
            const subscriptions = connectionSubscriptions.get(context.extra);
            if (!subscriptions) return;

            for (const [, groupId] of subscriptions) {
                decrementGroupSubscription(groupId);
            }
            subscriptions.clear();
        },
    },
    wsServer,
);

const server = new ApolloServer({
    schema,
    plugins: [
        ApolloServerPluginDrainHttpServer({ httpServer }),
        {
            async serverWillStart() {
                return {
                    async drainServer() {
                        stopPolling();
                        await serverCleanup.dispose();
                    },
                };
            },
        },
    ],
});

await server.start();

app.use('/graphql', cors(), express.json(), expressMiddleware(server));

// Health check endpoint (kept for compatibility)
app.get('/health', (_request, response) => {
    response.json({ status: 'ok', neo4j: NEO4J_URI });
});

httpServer.listen(PORT, () => {
    console.log(`\nGraphQL server ready at http://localhost:${PORT}/graphql`);
    console.log(`WebSocket subscriptions at ws://localhost:${PORT}/graphql`);
    console.log(`Neo4j: ${NEO4J_URI} (database: ${NEO4J_DATABASE}, encrypted: ${encryptionInUri ? 'from URI' : encrypted})`);
    console.log(`Poll interval: ${POLL_INTERVAL_MS}ms`);
    console.log(`\nExplorer: http://localhost:${PORT}/graphql`);
});

// ---------------------------------------------------------------------------
// Graceful shutdown
// ---------------------------------------------------------------------------

async function gracefulShutdown(signal) {
    console.log(`\n[${signal}] Shutting down gracefully...`);
    stopPolling();
    await server.stop();
    await driver.close();
    process.exit(0);
}

process.on('SIGINT', () => gracefulShutdown('SIGINT'));
process.on('SIGTERM', () => gracefulShutdown('SIGTERM'));
