import gql from 'graphql-tag';

export const typeDefs = gql`
    # ---------------------------------------------------------------------------
    # Scalar & shared types
    # ---------------------------------------------------------------------------

    type GraphNode {
        labels: [String!]!
        properties: JSON
    }

    type GraphEdge {
        source: GraphNode!
        target: GraphNode!
        relationType: String!
    }

    # ---------------------------------------------------------------------------
    # Dashboard / Summary
    # ---------------------------------------------------------------------------

    type PentestSummary {
        groupId: String!
        hosts: Int!
        ports: Int!
        services: Int!
        vulnerabilities: Int!
        accounts: Int!
        validAccess: Int!
        status: PentestStatus!
    }

    enum PentestStatus {
        COMPROMISED
        VULNERABLE
        SECURE
    }

    # ---------------------------------------------------------------------------
    # Attack Surface (Q1)
    # ---------------------------------------------------------------------------

    type AttackSurfaceEntity {
        entityType: String
        count: Int!
    }

    # ---------------------------------------------------------------------------
    # Credentials & Access (Q2)
    # ---------------------------------------------------------------------------

    type CredentialStatus {
        status: String!
        count: Int!
        examples: [String!]!
    }

    type AccessRecord {
        access: String
        account: String
        host: String
        service: String
        summary: String
    }

    # ---------------------------------------------------------------------------
    # Infrastructure (Q3)
    # ---------------------------------------------------------------------------

    type HostWithServices {
        host: String
        ports: [String!]!
        services: [String!]!
    }

    type OpenPort {
        port: String
        service: String
        host: String
    }

    # ---------------------------------------------------------------------------
    # Vulnerabilities (Q4)
    # ---------------------------------------------------------------------------

    type VulnerabilitySeverity {
        category: String!
        count: Int!
        examples: [String!]!
    }

    type CveRecord {
        cve: String
        foundOn: String
        source: String
    }

    type ExploitAttempt {
        vulnerability: String
        attemptCount: Int!
        status: String!
    }

    # ---------------------------------------------------------------------------
    # Tools (Q5)
    # ---------------------------------------------------------------------------

    type ToolUsage {
        tool: String
        executions: Int!
    }

    type ToolEffectiveness {
        tool: String
        executions: Int!
        discoveries: Int!
        discoveryTypes: [String!]!
    }

    type Artifact {
        artifact: String
        producedBy: String
        summary: String
    }

    # ---------------------------------------------------------------------------
    # Graph data (Q6)
    # ---------------------------------------------------------------------------

    type GraphData {
        groupId: String!
        rows: Int!
        data: [GraphEdge!]!
    }

    # ---------------------------------------------------------------------------
    # Attack Path Stats (Q7)
    # ---------------------------------------------------------------------------

    type AttackPathStats {
        groupId: String!
        hosts: Int!
        ports: Int!
        services: Int!
        vulnerabilities: Int!
        validAccess: Int!
        accounts: Int!
    }

    # ---------------------------------------------------------------------------
    # Queries
    # ---------------------------------------------------------------------------

    type Query {
        health: HealthStatus!

        # Dashboard
        dashboard(groupId: String!): PentestSummary!

        # Attack Surface (Q1)
        attackSurface(groupId: String!): [AttackSurfaceEntity!]!

        # Credentials & Access (Q2)
        credentialsStatus(groupId: String!): [CredentialStatus!]!
        accessDetails(groupId: String!): [AccessRecord!]!

        # Infrastructure (Q3)
        hostsWithServices(groupId: String!): [HostWithServices!]!
        openPorts(groupId: String!): [OpenPort!]!

        # Vulnerabilities (Q4)
        vulnerabilitySeverity(groupId: String!): [VulnerabilitySeverity!]!
        allCves(groupId: String!): [CveRecord!]!
        exploitAttempts(groupId: String!): [ExploitAttempt!]!

        # Tools (Q5)
        toolUsage(groupId: String!): [ToolUsage!]!
        toolEffectiveness(groupId: String!): [ToolEffectiveness!]!
        artifacts(groupId: String!): [Artifact!]!

        # Graph visualizations (Q6)
        mainAttackChain(groupId: String!): GraphData!
        fullAttackChain(groupId: String!): GraphData!
        infrastructureGraph(groupId: String!): GraphData!
        accessChainGraph(groupId: String!): GraphData!
        shortestPathGraph(groupId: String!): GraphData!

        # Attack Path Stats (Q7)
        attackPathStats(groupId: String!): AttackPathStats!
    }

    type HealthStatus {
        status: String!
        neo4j: String!
    }

    # ---------------------------------------------------------------------------
    # Subscriptions — real-time updates via WebSocket
    # ---------------------------------------------------------------------------

    type Subscription {
        # Dashboard updates
        dashboardUpdated(groupId: String!): PentestSummary!

        # Attack surface changes
        attackSurfaceUpdated(groupId: String!): [AttackSurfaceEntity!]!

        # Graph data updates
        mainAttackChainUpdated(groupId: String!): GraphData!
        fullAttackChainUpdated(groupId: String!): GraphData!

        # Attack path stats changes
        attackPathStatsUpdated(groupId: String!): AttackPathStats!

        # Vulnerability updates
        vulnerabilitySeverityUpdated(groupId: String!): [VulnerabilitySeverity!]!
    }

    # ---------------------------------------------------------------------------
    # Custom scalar for arbitrary JSON (node properties)
    # ---------------------------------------------------------------------------

    scalar JSON
`;
