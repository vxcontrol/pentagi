import { gql } from '@apollo/client';

// ---------------------------------------------------------------------------
// Fragments
// ---------------------------------------------------------------------------

const GRAPH_NODE_FRAGMENT = gql`
    fragment GraphNodeFields on GraphNode {
        labels
        properties
    }
`;

const GRAPH_EDGE_FRAGMENT = gql`
    fragment GraphEdgeFields on GraphEdge {
        source {
            ...GraphNodeFields
        }
        target {
            ...GraphNodeFields
        }
        relationType
    }
    ${GRAPH_NODE_FRAGMENT}
`;

const GRAPH_DATA_FRAGMENT = gql`
    fragment GraphDataFields on GraphData {
        groupId
        rows
        data {
            ...GraphEdgeFields
        }
    }
    ${GRAPH_EDGE_FRAGMENT}
`;

// ---------------------------------------------------------------------------
// Queries
// ---------------------------------------------------------------------------

export const DASHBOARD_QUERY = gql`
    query Dashboard($groupId: String!) {
        dashboard(groupId: $groupId) {
            groupId
            hosts
            ports
            services
            vulnerabilities
            accounts
            validAccess
            status
        }
    }
`;

export const ATTACK_SURFACE_QUERY = gql`
    query AttackSurface($groupId: String!) {
        attackSurface(groupId: $groupId) {
            entityType
            count
        }
    }
`;

export const CREDENTIALS_STATUS_QUERY = gql`
    query CredentialsStatus($groupId: String!) {
        credentialsStatus(groupId: $groupId) {
            status
            count
            examples
        }
    }
`;

export const ACCESS_DETAILS_QUERY = gql`
    query AccessDetails($groupId: String!) {
        accessDetails(groupId: $groupId) {
            access
            account
            host
            service
            summary
        }
    }
`;

export const HOSTS_WITH_SERVICES_QUERY = gql`
    query HostsWithServices($groupId: String!) {
        hostsWithServices(groupId: $groupId) {
            host
            ports
            services
        }
    }
`;

export const OPEN_PORTS_QUERY = gql`
    query OpenPorts($groupId: String!) {
        openPorts(groupId: $groupId) {
            port
            service
            host
        }
    }
`;

export const VULNERABILITY_SEVERITY_QUERY = gql`
    query VulnerabilitySeverity($groupId: String!) {
        vulnerabilitySeverity(groupId: $groupId) {
            category
            count
            examples
        }
    }
`;

export const ALL_CVES_QUERY = gql`
    query AllCves($groupId: String!) {
        allCves(groupId: $groupId) {
            cve
            foundOn
            source
        }
    }
`;

export const EXPLOIT_ATTEMPTS_QUERY = gql`
    query ExploitAttempts($groupId: String!) {
        exploitAttempts(groupId: $groupId) {
            vulnerability
            attemptCount
            status
        }
    }
`;

export const TOOL_USAGE_QUERY = gql`
    query ToolUsage($groupId: String!) {
        toolUsage(groupId: $groupId) {
            tool
            executions
        }
    }
`;

export const TOOL_EFFECTIVENESS_QUERY = gql`
    query ToolEffectiveness($groupId: String!) {
        toolEffectiveness(groupId: $groupId) {
            tool
            executions
            discoveries
            discoveryTypes
        }
    }
`;

export const ARTIFACTS_QUERY = gql`
    query Artifacts($groupId: String!) {
        artifacts(groupId: $groupId) {
            artifact
            producedBy
            summary
        }
    }
`;

export const MAIN_ATTACK_CHAIN_QUERY = gql`
    query MainAttackChain($groupId: String!) {
        mainAttackChain(groupId: $groupId) {
            ...GraphDataFields
        }
    }
    ${GRAPH_DATA_FRAGMENT}
`;

export const FULL_ATTACK_CHAIN_QUERY = gql`
    query FullAttackChain($groupId: String!) {
        fullAttackChain(groupId: $groupId) {
            ...GraphDataFields
        }
    }
    ${GRAPH_DATA_FRAGMENT}
`;

export const INFRASTRUCTURE_GRAPH_QUERY = gql`
    query InfrastructureGraph($groupId: String!) {
        infrastructureGraph(groupId: $groupId) {
            ...GraphDataFields
        }
    }
    ${GRAPH_DATA_FRAGMENT}
`;

export const ACCESS_CHAIN_GRAPH_QUERY = gql`
    query AccessChainGraph($groupId: String!) {
        accessChainGraph(groupId: $groupId) {
            ...GraphDataFields
        }
    }
    ${GRAPH_DATA_FRAGMENT}
`;

export const SHORTEST_PATH_GRAPH_QUERY = gql`
    query ShortestPathGraph($groupId: String!) {
        shortestPathGraph(groupId: $groupId) {
            ...GraphDataFields
        }
    }
    ${GRAPH_DATA_FRAGMENT}
`;

export const ATTACK_PATH_STATS_QUERY = gql`
    query AttackPathStats($groupId: String!) {
        attackPathStats(groupId: $groupId) {
            groupId
            hosts
            ports
            services
            vulnerabilities
            validAccess
            accounts
        }
    }
`;

// ---------------------------------------------------------------------------
// Combined query — fetches all dashboard data in a single request
// ---------------------------------------------------------------------------

export const FULL_DASHBOARD_QUERY = gql`
    query FullDashboard($groupId: String!) {
        dashboard(groupId: $groupId) {
            groupId
            hosts
            ports
            services
            vulnerabilities
            accounts
            validAccess
            status
        }
        attackSurface(groupId: $groupId) {
            entityType
            count
        }
        credentialsStatus(groupId: $groupId) {
            status
            count
            examples
        }
        accessDetails(groupId: $groupId) {
            access
            account
            host
            service
            summary
        }
        hostsWithServices(groupId: $groupId) {
            host
            ports
            services
        }
        openPorts(groupId: $groupId) {
            port
            service
            host
        }
        vulnerabilitySeverity(groupId: $groupId) {
            category
            count
            examples
        }
        allCves(groupId: $groupId) {
            cve
            foundOn
            source
        }
        exploitAttempts(groupId: $groupId) {
            vulnerability
            attemptCount
            status
        }
        toolUsage(groupId: $groupId) {
            tool
            executions
        }
        toolEffectiveness(groupId: $groupId) {
            tool
            executions
            discoveries
            discoveryTypes
        }
        artifacts(groupId: $groupId) {
            artifact
            producedBy
            summary
        }
        mainAttackChain(groupId: $groupId) {
            ...GraphDataFields
        }
        fullAttackChain(groupId: $groupId) {
            ...GraphDataFields
        }
        infrastructureGraph(groupId: $groupId) {
            ...GraphDataFields
        }
        accessChainGraph(groupId: $groupId) {
            ...GraphDataFields
        }
        shortestPathGraph(groupId: $groupId) {
            ...GraphDataFields
        }
        attackPathStats(groupId: $groupId) {
            groupId
            hosts
            ports
            services
            vulnerabilities
            validAccess
            accounts
        }
    }
    ${GRAPH_DATA_FRAGMENT}
`;

// ---------------------------------------------------------------------------
// Subscriptions
// ---------------------------------------------------------------------------

export const DASHBOARD_UPDATED_SUBSCRIPTION = gql`
    subscription DashboardUpdated($groupId: String!) {
        dashboardUpdated(groupId: $groupId) {
            groupId
            hosts
            ports
            services
            vulnerabilities
            accounts
            validAccess
            status
        }
    }
`;

export const ATTACK_SURFACE_UPDATED_SUBSCRIPTION = gql`
    subscription AttackSurfaceUpdated($groupId: String!) {
        attackSurfaceUpdated(groupId: $groupId) {
            entityType
            count
        }
    }
`;

export const MAIN_ATTACK_CHAIN_UPDATED_SUBSCRIPTION = gql`
    subscription MainAttackChainUpdated($groupId: String!) {
        mainAttackChainUpdated(groupId: $groupId) {
            ...GraphDataFields
        }
    }
    ${GRAPH_DATA_FRAGMENT}
`;

export const ATTACK_PATH_STATS_UPDATED_SUBSCRIPTION = gql`
    subscription AttackPathStatsUpdated($groupId: String!) {
        attackPathStatsUpdated(groupId: $groupId) {
            groupId
            hosts
            ports
            services
            vulnerabilities
            validAccess
            accounts
        }
    }
`;

export const VULNERABILITY_SEVERITY_UPDATED_SUBSCRIPTION = gql`
    subscription VulnerabilitySeverityUpdated($groupId: String!) {
        vulnerabilitySeverityUpdated(groupId: $groupId) {
            category
            count
            examples
        }
    }
`;
