/**
 * Cypher queries for pentest metrics.
 * Based on Pentest Metrics Cypher Queries documentation (2026-01-29).
 * Replace $group_id with the flow-id (e.g. 'flow-1571').
 * @typedef {import('./types.ts').CypherQuery} CypherQuery
 */

// ---------------------------------------------------------------------------
// ⭐ PRESET: Main Attack Chain Dashboard
// ---------------------------------------------------------------------------

/** Main Attack Chain — all primary entities and relationships for graph visualization.
 * Uses UNION ALL with intermediate dedup to avoid cartesian product.
 * @param {string} groupId
 * @returns {CypherQuery}
 */
export const mainAttackChain = (groupId) => ({
    text: `
    MATCH (source:Host {group_id: $group_id})-[rel:HAS_PORT]->(target:Port)
    RETURN source, rel, target
    UNION ALL
    MATCH (:Host {group_id: $group_id})-[:HAS_PORT]->(p:Port)
    WITH DISTINCT p
    MATCH (p)-[rel:RUNS_SERVICE|LEADS_TO]->(target:Service)
    RETURN p AS source, rel, target
    UNION ALL
    MATCH (:Host {group_id: $group_id})-[:HAS_PORT]->(:Port)-[:RUNS_SERVICE|LEADS_TO]->(s:Service)
    WITH DISTINCT s
    MATCH (s)-[rel:HAS_VULNERABILITY|DETECTED_VULNERABILITY]->(target:Vulnerability)
    RETURN s AS source, rel, target
    UNION ALL
    MATCH (source:ValidAccess {group_id: $group_id})-[rel:ON_HOST]->(target:Host)
    RETURN source, rel, target
    UNION ALL
    MATCH (source:ValidAccess {group_id: $group_id})-[rel:AS_ACCOUNT]->(target:Account)
    RETURN source, rel, target
  `,
    params: { group_id: groupId },
});

// ---------------------------------------------------------------------------
// 📊 Query 1: Attack Surface Overview
// ---------------------------------------------------------------------------

/** Attack Surface Overview — entity type distribution.
 * @param {string} groupId
 * @returns {CypherQuery}
 */
export const attackSurfaceOverview = (groupId) => ({
    text: `
    MATCH (n:Entity)
    WHERE n.group_id = $group_id
    WITH labels(n)[1] AS EntityType, count(n) AS Count
    ORDER BY Count DESC
    RETURN EntityType, Count
  `,
    params: { group_id: groupId },
});

// ---------------------------------------------------------------------------
// 🔑 Query 2: Credentials & Access Status
// ---------------------------------------------------------------------------

/** 2a: Accounts breakdown — discovered vs compromised.
 * @param {string} groupId
 * @returns {CypherQuery}
 */
export const credentialsStatus = (groupId) => ({
    text: `
    MATCH (a:Account)
    WHERE a.group_id = $group_id
    OPTIONAL MATCH (va:ValidAccess)-[:AS_ACCOUNT]->(a)
    WITH a,
         CASE WHEN va IS NOT NULL THEN 'COMPROMISED' ELSE 'DISCOVERED' END AS Status
    RETURN Status, count(a) AS Count, collect(a.name)[0..5] AS Examples
    ORDER BY Status
  `,
    params: { group_id: groupId },
});

/** 2b: Valid Access details.
 * @param {string} groupId
 * @returns {CypherQuery}
 */
export const validAccessDetails = (groupId) => ({
    text: `
    MATCH (va:ValidAccess)
    WHERE va.group_id = $group_id
    OPTIONAL MATCH (va)-[:AS_ACCOUNT]->(acc:Account)
    OPTIONAL MATCH (va)-[:ON_HOST]->(h:Host)
    OPTIONAL MATCH (va)-[:VIA_SERVICE]->(s:Service)
    RETURN va.name AS Access,
           acc.name AS Account,
           h.name AS Host,
           s.name AS Service,
           va.summary AS Summary
  `,
    params: { group_id: groupId },
});

// ---------------------------------------------------------------------------
// 🖥️ Query 3: Infrastructure Map
// ---------------------------------------------------------------------------

/** 3a: Hosts with ports and services.
 * @param {string} groupId
 * @returns {CypherQuery}
 */
export const hostsWithServices = (groupId) => ({
    text: `
    MATCH (h:Host)
    WHERE h.group_id = $group_id
    OPTIONAL MATCH (h)-[:HAS_PORT]->(p:Port)
    OPTIONAL MATCH (p)-[:RUNS_SERVICE|LEADS_TO]->(s:Service)
    WITH h,
         collect(DISTINCT p.name) AS Ports,
         collect(DISTINCT s.name) AS Services
    RETURN h.name AS Host, Ports, Services
  `,
    params: { group_id: groupId },
});

/** 3b: Open ports with services.
 * @param {string} groupId
 * @returns {CypherQuery}
 */
export const openPortsWithServices = (groupId) => ({
    text: `
    MATCH (p:Port)
    WHERE p.group_id = $group_id
    OPTIONAL MATCH (p)-[:RUNS_SERVICE|LEADS_TO]->(s:Service)
    OPTIONAL MATCH (h:Host)-[:HAS_PORT]->(p)
    RETURN p.name AS Port,
           s.name AS Service,
           h.name AS Host
    ORDER BY toInteger(split(p.name, '/')[0])
  `,
    params: { group_id: groupId },
});

// ---------------------------------------------------------------------------
// ⚠️ Query 4: Vulnerability Breakdown
// ---------------------------------------------------------------------------

/** 4a: Severity breakdown.
 * @param {string} groupId
 * @returns {CypherQuery}
 */
export const vulnerabilitySeverity = (groupId) => ({
    text: `
    MATCH (v:Vulnerability)
    WHERE v.group_id = $group_id
    WITH v,
         CASE
             WHEN v.name CONTAINS 'CVE-' THEN 'CVE'
             WHEN v.name =~ '(?i).*(RCE|remote code|command injection).*' THEN 'CRITICAL'
             WHEN v.name =~ '(?i).*(LFI|RFI|SQLi|file inclusion).*' THEN 'HIGH'
             WHEN v.name =~ '(?i).*(XSS|CSRF|disclosure).*' THEN 'MEDIUM'
             ELSE 'LOW/INFO'
         END AS Category
    RETURN Category, count(v) AS Count, collect(v.name)[0..3] AS Examples
    ORDER BY
        CASE Category
            WHEN 'CVE' THEN 1
            WHEN 'CRITICAL' THEN 2
            WHEN 'HIGH' THEN 3
            WHEN 'MEDIUM' THEN 4
            ELSE 5
        END
  `,
    params: { group_id: groupId },
});

/** 4b: All CVEs list.
 * @param {string} groupId
 * @returns {CypherQuery}
 */
export const allCves = (groupId) => ({
    text: `
    MATCH (v:Vulnerability)
    WHERE v.group_id = $group_id
      AND v.name CONTAINS 'CVE-'
    OPTIONAL MATCH (v)<-[:HAS_VULNERABILITY|DETECTED_VULNERABILITY]-(source)
    RETURN v.name AS CVE,
           labels(source)[1] AS FoundOn,
           source.name AS Source
    ORDER BY v.name
  `,
    params: { group_id: groupId },
});

/** 4c: Exploitable vulnerabilities — linked to attempts.
 * @param {string} groupId
 * @returns {CypherQuery}
 */
export const exploitableVulnerabilities = (groupId) => ({
    text: `
    MATCH (a:Attempt)-[:ATTEMPTED_ON]->(v:Vulnerability)
    WHERE v.group_id = $group_id
    OPTIONAL MATCH (va:ValidAccess)-[:VIA_SERVICE]->()-[:HAS_VULNERABILITY]->(v)
    RETURN v.name AS Vulnerability,
           count(a) AS AttemptCount,
           CASE WHEN va IS NOT NULL THEN 'EXPLOITED' ELSE 'ATTEMPTED' END AS Status
    ORDER BY AttemptCount DESC
  `,
    params: { group_id: groupId },
});

// ---------------------------------------------------------------------------
// 🛠️ Query 5: Tool Execution & Effectiveness
// ---------------------------------------------------------------------------

/** 5a: Tool usage statistics.
 * @param {string} groupId
 * @returns {CypherQuery}
 */
export const toolUsageStats = (groupId) => ({
    text: `
    MATCH (t:Tool)
    WHERE t.group_id = $group_id
    OPTIONAL MATCH (te:ToolExecution)-[:EXECUTED_TOOL]->(t)
    WITH t, count(te) AS Executions
    ORDER BY Executions DESC
    RETURN t.name AS Tool, Executions
  `,
    params: { group_id: groupId },
});

/** 5b: Tool effectiveness — discoveries per tool.
 * @param {string} groupId
 * @returns {CypherQuery}
 */
export const toolEffectiveness = (groupId) => ({
    text: `
    MATCH (te:ToolExecution)-[:EXECUTED_TOOL]->(t:Tool)
    WHERE t.group_id = $group_id
    OPTIONAL MATCH (te)-[:TOOL_DISCOVERED]->(discovered)
    WITH t.name AS Tool,
         count(DISTINCT te) AS Executions,
         count(DISTINCT discovered) AS Discoveries,
         collect(DISTINCT labels(discovered)[1]) AS DiscoveryTypes
    RETURN Tool, Executions, Discoveries, DiscoveryTypes
    ORDER BY Discoveries DESC
  `,
    params: { group_id: groupId },
});

/** 5c: Artifacts produced.
 * @param {string} groupId
 * @returns {CypherQuery}
 */
export const artifactsProduced = (groupId) => ({
    text: `
    MATCH (a:Artifact)
    WHERE a.group_id = $group_id
    OPTIONAL MATCH (te:ToolExecution)-[:PRODUCED]->(a)
    RETURN a.name AS Artifact,
           te.name AS ProducedBy,
           a.summary AS Summary
    ORDER BY a.created_at DESC
    LIMIT 20
  `,
    params: { group_id: groupId },
});

// ---------------------------------------------------------------------------
// 📈 Bonus: Pentest Summary Dashboard
// ---------------------------------------------------------------------------

/** Pentest summary — one-row summary with status.
 * @param {string} groupId
 * @returns {CypherQuery}
 */
export const pentestSummary = (groupId) => ({
    text: `
    MATCH (n:Entity)
    WHERE n.group_id = $group_id
    WITH labels(n)[1] AS Type, count(n) AS C
    WITH collect({type: Type, count: C}) AS stats
    WITH stats,
         [x IN stats WHERE x.type = 'Host'][0].count AS Hosts,
         [x IN stats WHERE x.type = 'Vulnerability'][0].count AS Vulns,
         [x IN stats WHERE x.type = 'Account'][0].count AS Accounts,
         [x IN stats WHERE x.type = 'ValidAccess'][0].count AS Access,
         [x IN stats WHERE x.type = 'Port'][0].count AS Ports,
         [x IN stats WHERE x.type = 'Service'][0].count AS Services
    RETURN
         coalesce(Hosts, 0) AS Hosts,
         coalesce(Ports, 0) AS Ports,
         coalesce(Services, 0) AS Services,
         coalesce(Vulns, 0) AS Vulnerabilities,
         coalesce(Accounts, 0) AS Accounts,
         coalesce(Access, 0) AS ValidAccess,
         CASE
             WHEN coalesce(Access, 0) > 0 THEN 'COMPROMISED'
             WHEN coalesce(Vulns, 0) > 5 THEN 'VULNERABLE'
             ELSE 'SECURE'
         END AS Status
  `,
    params: { group_id: groupId },
});

// ---------------------------------------------------------------------------
// 🔗 Query 6: Graph Visualization
// ---------------------------------------------------------------------------

/** 6a: Full Attack Chain — Host → Port → Service → Vulnerability + ValidAccess.
 * Uses UNION ALL with intermediate dedup to avoid cartesian product.
 * @param {string} groupId
 * @returns {CypherQuery}
 */
export const fullAttackChain = (groupId) => ({
    text: `
    MATCH (source:Host {group_id: $group_id})-[rel:HAS_PORT]->(target:Port)
    RETURN source, rel, target
    UNION ALL
    MATCH (:Host {group_id: $group_id})-[:HAS_PORT]->(p:Port)
    WITH DISTINCT p
    MATCH (p)-[rel:RUNS_SERVICE|LEADS_TO]->(target:Service)
    RETURN p AS source, rel, target
    UNION ALL
    MATCH (:Host {group_id: $group_id})-[:HAS_PORT]->(:Port)-[:RUNS_SERVICE|LEADS_TO]->(s:Service)
    WITH DISTINCT s
    MATCH (s)-[rel:HAS_VULNERABILITY|DETECTED_VULNERABILITY]->(target:Vulnerability)
    RETURN s AS source, rel, target
    UNION ALL
    MATCH (source:ValidAccess {group_id: $group_id})-[rel:VIA_SERVICE]->(target:Service)
    RETURN source, rel, target
    UNION ALL
    MATCH (source:ValidAccess {group_id: $group_id})-[rel:ON_HOST]->(target:Host)
    RETURN source, rel, target
    UNION ALL
    MATCH (source:ValidAccess {group_id: $group_id})-[rel:AS_ACCOUNT]->(target:Account)
    RETURN source, rel, target
  `,
    params: { group_id: groupId },
});

/** 6b: Infrastructure Only — Host → Port → Service → Vulnerability (no ValidAccess).
 * Uses UNION ALL with intermediate dedup to avoid cartesian product.
 * @param {string} groupId
 * @returns {CypherQuery}
 */
export const infrastructureOnly = (groupId) => ({
    text: `
    MATCH (source:Host {group_id: $group_id})-[rel:HAS_PORT]->(target:Port)
    RETURN source, rel, target
    UNION ALL
    MATCH (:Host {group_id: $group_id})-[:HAS_PORT]->(p:Port)
    WITH DISTINCT p
    MATCH (p)-[rel:RUNS_SERVICE|LEADS_TO]->(target:Service)
    RETURN p AS source, rel, target
    UNION ALL
    MATCH (:Host {group_id: $group_id})-[:HAS_PORT]->(:Port)-[:RUNS_SERVICE|LEADS_TO]->(s:Service)
    WITH DISTINCT s
    MATCH (s)-[rel:HAS_VULNERABILITY|DETECTED_VULNERABILITY]->(target:Vulnerability)
    RETURN s AS source, rel, target
  `,
    params: { group_id: groupId },
});

/** 6c: Access Chain — ValidAccess → Host + Service + Account.
 * @param {string} groupId
 * @returns {CypherQuery}
 */
export const accessChain = (groupId) => ({
    text: `
    MATCH (va:ValidAccess {group_id: $group_id})-[:ON_HOST]->(h:Host)
    MATCH (va)-[:AS_ACCOUNT]->(a:Account)
    OPTIONAL MATCH (va)-[:VIA_SERVICE]->(s:Service)
    RETURN *
  `,
    params: { group_id: groupId },
});

/** 6d: Shortest Path — Host ← ValidAccess → Account.
 * @param {string} groupId
 * @returns {CypherQuery}
 */
export const shortestPath = (groupId) => ({
    text: `
    MATCH (h:Host)<-[:ON_HOST]-(va:ValidAccess)-[:AS_ACCOUNT]->(a:Account)
    WHERE h.group_id = $group_id
    RETURN *
  `,
    params: { group_id: groupId },
});

// ---------------------------------------------------------------------------
// 📊 Query 7: Attack Path Stats (Summary)
// ---------------------------------------------------------------------------

/** Attack Path Stats — counts along the attack chain.
 * Uses CALL subqueries to avoid cartesian product when counting.
 * @param {string} groupId
 * @returns {CypherQuery}
 */
export const attackPathStats = (groupId) => ({
    text: `
    CALL { MATCH (h:Host {group_id: $group_id}) RETURN count(h) AS Hosts }
    CALL { MATCH (:Host {group_id: $group_id})-[:HAS_PORT]->(p:Port) RETURN count(DISTINCT p) AS Ports }
    CALL { MATCH (:Host {group_id: $group_id})-[:HAS_PORT]->(:Port)-[:RUNS_SERVICE|LEADS_TO]->(s:Service) RETURN count(DISTINCT s) AS Services }
    CALL { MATCH (:Host {group_id: $group_id})-[:HAS_PORT]->(:Port)-[:RUNS_SERVICE|LEADS_TO]->(:Service)-[:HAS_VULNERABILITY|DETECTED_VULNERABILITY]->(v:Vulnerability) RETURN count(DISTINCT v) AS Vulnerabilities }
    CALL { MATCH (va:ValidAccess {group_id: $group_id}) RETURN count(va) AS ValidAccess }
    CALL { MATCH (va:ValidAccess {group_id: $group_id})-[:AS_ACCOUNT]->(a:Account) RETURN count(DISTINCT a) AS Accounts }
    RETURN Hosts, Ports, Services, Vulnerabilities, ValidAccess, Accounts
  `,
    params: { group_id: groupId },
});
