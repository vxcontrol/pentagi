// --- Neo4j primitive types ---

export interface Neo4jDateTime {
    year: number;
    month: number;
    day: number;
    hour: number;
    minute: number;
    second: number;
    nanosecond: number;
    timeZoneOffsetSeconds: number;
    timeZoneId?: string;
}

// --- Neo4j graph objects ---

export interface Neo4jNode {
    labels: string[];
    properties: Record<string, unknown>;
}

export interface Neo4jRelationship {
    type: string;
    properties: Record<string, unknown>;
}

export interface Neo4jPathSegment {
    start: Neo4jNode;
    relationship: Neo4jRelationship;
    end: Neo4jNode;
}

export interface Neo4jPath {
    start: Neo4jNode;
    end: Neo4jNode;
    segments: Neo4jPathSegment[];
}

// --- Cypher query helper ---

export interface CypherQuery {
    text: string;
    params: Record<string, unknown>;
}

// --- GraphQL types (matching schema.js) ---

export type PentestStatus = 'COMPROMISED' | 'VULNERABLE' | 'SECURE';

export interface PentestSummary {
    groupId: string;
    hosts: number;
    ports: number;
    services: number;
    vulnerabilities: number;
    accounts: number;
    validAccess: number;
    status: PentestStatus;
}

export interface AttackSurfaceEntity {
    entityType: string;
    count: number;
}

export interface CredentialStatus {
    status: 'COMPROMISED' | 'DISCOVERED';
    count: number;
    examples: string[];
}

export interface AccessRecord {
    access: string | null;
    account: string | null;
    host: string | null;
    service: string | null;
    summary: string | null;
}

export interface HostWithServices {
    host: string | null;
    ports: string[];
    services: string[];
}

export interface OpenPort {
    port: string | null;
    service: string | null;
    host: string | null;
}

export interface VulnerabilitySeverityRecord {
    category: 'CVE' | 'CRITICAL' | 'HIGH' | 'MEDIUM' | 'LOW/INFO';
    count: number;
    examples: string[];
}

export interface CveRecord {
    cve: string | null;
    foundOn: string | null;
    source: string | null;
}

export interface ExploitAttempt {
    vulnerability: string | null;
    attemptCount: number;
    status: 'EXPLOITED' | 'ATTEMPTED';
}

export interface ToolUsage {
    tool: string | null;
    executions: number;
}

export interface ToolEffectivenessRecord {
    tool: string | null;
    executions: number;
    discoveries: number;
    discoveryTypes: string[];
}

export interface Artifact {
    artifact: string | null;
    producedBy: string | null;
    summary: string | null;
}

export interface GraphEdge {
    source: Neo4jNode;
    target: Neo4jNode;
    relationType: string;
}

export interface GraphData {
    groupId: string;
    rows: number;
    data: GraphEdge[];
}

export interface AttackPathStats {
    groupId: string;
    hosts: number;
    ports: number;
    services: number;
    vulnerabilities: number;
    validAccess: number;
    accounts: number;
}

// --- Relationship property shapes (common in overview paths) ---

export interface EdgeProperties {
    fact: string;
    edge_class: string;
    group_id: string;
    uuid: string;
    version: number;
    weight: number;
    episodes?: string[];
    created_at: Neo4jDateTime;
    valid_at?: Neo4jDateTime;
    invalid_at?: Neo4jDateTime;
    expired_at?: Neo4jDateTime;
    fact_embedding?: number[];
}

// --- Common node property shapes ---

export interface HostProperties {
    name: string;
    hostname: string;
    aliases: string[];
    ip: string;
    group_id: string;
    summary: string;
    uuid: string;
    version: number;
    discovery_order: number;
    created_at: Neo4jDateTime;
    name_embedding: number[];
    labels: string[];
}

export interface MisconfigurationProperties {
    name: string;
    misconfig_name: string;
    misconf_type: string;
    misconf_subtype: string;
    severity: string;
    description: string;
    summary: string;
    target_object: string;
    group_id: string;
    uuid: string;
    version: number;
    discovery_order: number;
    created_at: Neo4jDateTime;
    name_embedding: number[];
    labels: string[];
}

export interface PortProperties {
    name: string;
    number: number;
    proto: string;
    state: string;
    summary: string;
    group_id: string;
    uuid: string;
    version: number;
    discovery_order: number;
    entity_uuid: string;
    created_at: Neo4jDateTime;
    name_embedding: number[];
    labels: string[];
}
