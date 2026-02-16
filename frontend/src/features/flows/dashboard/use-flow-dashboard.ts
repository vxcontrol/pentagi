import { useQuery, useSubscription } from '@apollo/client';
import { useCallback, useEffect, useMemo } from 'react';

import { neo4jClient } from './neo4j-client';
import {
    ATTACK_PATH_STATS_UPDATED_SUBSCRIPTION,
    ATTACK_SURFACE_UPDATED_SUBSCRIPTION,
    DASHBOARD_UPDATED_SUBSCRIPTION,
    FULL_DASHBOARD_QUERY,
    MAIN_ATTACK_CHAIN_UPDATED_SUBSCRIPTION,
    VULNERABILITY_SEVERITY_UPDATED_SUBSCRIPTION,
} from './neo4j-queries';

// ---------------------------------------------------------------------------
// Types (matching GraphQL schema)
// ---------------------------------------------------------------------------

export interface AccessRecord {
    access: null | string;
    account: null | string;
    host: null | string;
    service: null | string;
    summary: null | string;
}

export interface ArtifactRecord {
    artifact: null | string;
    producedBy: null | string;
    summary: null | string;
}

export interface AttackPathStats {
    accounts: number;
    groupId: string;
    hosts: number;
    ports: number;
    services: number;
    validAccess: number;
    vulnerabilities: number;
}

export interface AttackSurfaceEntity {
    count: number;
    entityType: string;
}

export interface CredentialStatus {
    count: number;
    examples: string[];
    status: string;
}

export interface CveRecord {
    cve: null | string;
    foundOn: null | string;
    source: null | string;
}

export interface ExploitAttempt {
    attemptCount: number;
    status: string;
    vulnerability: null | string;
}

export interface FlowDashboardResult {
    accessChainGraph: GraphData | null;
    accessDetails: AccessRecord[];
    allCves: CveRecord[];
    artifacts: ArtifactRecord[];
    attackPathStats: AttackPathStats | null;
    attackSurface: AttackSurfaceEntity[];
    credentialsStatus: CredentialStatus[];
    dashboard: null | PentestSummary;
    error: null | string;
    exploitAttempts: ExploitAttempt[];
    fullAttackChain: GraphData | null;
    hostsWithServices: HostWithServices[];
    infrastructureGraph: GraphData | null;
    isLoading: boolean;
    mainAttackChain: GraphData | null;
    openPorts: OpenPort[];
    refetch: () => void;
    shortestPathGraph: GraphData | null;
    toolEffectiveness: ToolEffectivenessRecord[];
    toolUsage: ToolUsageRecord[];
    vulnerabilitySeverity: VulnerabilitySeverityRecord[];
}

export interface GraphData {
    data: GraphEdge[];
    groupId: string;
    rows: number;
}

export interface GraphEdge {
    relationType: string;
    source: GraphNode;
    target: GraphNode;
}

export interface GraphNode {
    labels: string[];
    properties: Record<string, unknown>;
}

export interface HostWithServices {
    host: null | string;
    ports: string[];
    services: string[];
}

export interface OpenPort {
    host: null | string;
    port: null | string;
    service: null | string;
}

export type PentestStatus = 'COMPROMISED' | 'SECURE' | 'VULNERABLE';

export interface PentestSummary {
    accounts: number;
    groupId: string;
    hosts: number;
    ports: number;
    services: number;
    status: PentestStatus;
    validAccess: number;
    vulnerabilities: number;
}

export interface ToolEffectivenessRecord {
    discoveries: number;
    discoveryTypes: string[];
    executions: number;
    tool: null | string;
}

export interface ToolUsageRecord {
    executions: number;
    tool: null | string;
}

export interface VulnerabilitySeverityRecord {
    category: string;
    count: number;
    examples: string[];
}

// Full query response
interface FullDashboardData {
    accessChainGraph: GraphData;
    accessDetails: AccessRecord[];
    allCves: CveRecord[];
    artifacts: ArtifactRecord[];
    attackPathStats: AttackPathStats;
    attackSurface: AttackSurfaceEntity[];
    credentialsStatus: CredentialStatus[];
    dashboard: PentestSummary;
    exploitAttempts: ExploitAttempt[];
    fullAttackChain: GraphData;
    hostsWithServices: HostWithServices[];
    infrastructureGraph: GraphData;
    mainAttackChain: GraphData;
    openPorts: OpenPort[];
    shortestPathGraph: GraphData;
    toolEffectiveness: ToolEffectivenessRecord[];
    toolUsage: ToolUsageRecord[];
    vulnerabilitySeverity: VulnerabilitySeverityRecord[];
}

// ---------------------------------------------------------------------------
// Hook
// ---------------------------------------------------------------------------

export function useFlowDashboard(groupId: null | string): FlowDashboardResult {
    const skip = !groupId;
    const variables = useMemo(() => ({ groupId: groupId ?? '' }), [groupId]);

    // Main query — fetches all dashboard data at once
    const { data, error, loading, refetch } = useQuery<FullDashboardData>(
        FULL_DASHBOARD_QUERY,
        {
            client: neo4jClient,
            skip,
            variables,
        },
    );

    // Subscriptions — update data in real time when Neo4j data changes
    const subscriptionOptions = useMemo(
        () => ({ client: neo4jClient, skip, variables }),
        [skip, variables],
    );

    const { data: dashboardUpdate } = useSubscription(
        DASHBOARD_UPDATED_SUBSCRIPTION,
        {
            ...subscriptionOptions,
            onData: ({ data }) => console.log('[Sub] dashboardUpdated', data.data),
            onError: (error) => console.error('[Sub] dashboardUpdated error', error),
        },
    );

    const { data: attackSurfaceUpdate } = useSubscription(
        ATTACK_SURFACE_UPDATED_SUBSCRIPTION,
        {
            ...subscriptionOptions,
            onData: ({ data }) => console.log('[Sub] attackSurfaceUpdated', data.data),
            onError: (error) => console.error('[Sub] attackSurfaceUpdated error', error),
        },
    );

    const { data: mainAttackChainUpdate } = useSubscription(
        MAIN_ATTACK_CHAIN_UPDATED_SUBSCRIPTION,
        {
            ...subscriptionOptions,
            onData: ({ data }) => console.log('[Sub] mainAttackChainUpdated', data.data),
            onError: (error) => console.error('[Sub] mainAttackChainUpdated error', error),
        },
    );

    const { data: attackPathStatsUpdate } = useSubscription(
        ATTACK_PATH_STATS_UPDATED_SUBSCRIPTION,
        {
            ...subscriptionOptions,
            onData: ({ data }) => console.log('[Sub] attackPathStatsUpdated', data.data),
            onError: (error) => console.error('[Sub] attackPathStatsUpdated error', error),
        },
    );

    const { data: vulnerabilitySeverityUpdate } = useSubscription(
        VULNERABILITY_SEVERITY_UPDATED_SUBSCRIPTION,
        {
            ...subscriptionOptions,
            onData: ({ data }) => console.log('[Sub] vulnerabilitySeverityUpdated', data.data),
            onError: (error) => console.error('[Sub] vulnerabilitySeverityUpdated error', error),
        },
    );

    // Refetch all data on demand (triggers full query reload)
    const handleRefetch = useCallback(() => {
        if (!skip) {
            refetch();
        }
    }, [refetch, skip]);

    // Refetch on initial mount when groupId becomes available
    useEffect(() => {
        if (!skip) {
            refetch();
        }
    }, [skip, refetch]);

    // Merge query data with subscription updates (subscriptions take priority)
    const dashboard: null | PentestSummary =
        dashboardUpdate?.dashboardUpdated ?? data?.dashboard ?? null;

    const attackSurface: AttackSurfaceEntity[] =
        attackSurfaceUpdate?.attackSurfaceUpdated ?? data?.attackSurface ?? [];

    const mainAttackChain: GraphData | null =
        mainAttackChainUpdate?.mainAttackChainUpdated ?? data?.mainAttackChain ?? null;

    const attackPathStats: AttackPathStats | null =
        attackPathStatsUpdate?.attackPathStatsUpdated ?? data?.attackPathStats ?? null;

    const vulnerabilitySeverity: VulnerabilitySeverityRecord[] =
        vulnerabilitySeverityUpdate?.vulnerabilitySeverityUpdated ??
        data?.vulnerabilitySeverity ?? [];

    return {
        accessChainGraph: data?.accessChainGraph ?? null,
        accessDetails: data?.accessDetails ?? [],
        allCves: data?.allCves ?? [],
        artifacts: data?.artifacts ?? [],
        attackPathStats,
        attackSurface,
        credentialsStatus: data?.credentialsStatus ?? [],
        dashboard,
        error: error ? error.message : null,
        exploitAttempts: data?.exploitAttempts ?? [],
        fullAttackChain: data?.fullAttackChain ?? null,
        hostsWithServices: data?.hostsWithServices ?? [],
        infrastructureGraph: data?.infrastructureGraph ?? null,
        isLoading: loading,
        mainAttackChain,
        openPorts: data?.openPorts ?? [],
        refetch: handleRefetch,
        shortestPathGraph: data?.shortestPathGraph ?? null,
        toolEffectiveness: data?.toolEffectiveness ?? [],
        toolUsage: data?.toolUsage ?? [],
        vulnerabilitySeverity,
    };
}
