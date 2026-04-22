# OSINT Integration Scenarios for PentAGI Agents

Issue [#68](https://github.com/vxcontrol/pentagi/issues/68) asks about adding OSINT frameworks such as ransomware.live and Flare. This guide answers the main product question first: what should agents ask these systems, when should they ask, and what should they do with the result?

This is not a built-in integration plan. It is a scenario guide for deciding whether a future external function, MCP server, or provider-specific adapter is worth adding.

## Common Agent Contract

A future OSINT tool should behave like an enrichment tool, not an authorization bypass or automatic scope expansion tool.

Suggested input:

```json
{
  "provider": "ransomware_live|flare",
  "indicator": "example.com",
  "indicator_type": "domain|ip|url|organization|sector|country|ransomware_group",
  "purpose": "triage|exposure_check|report_context|ioc_lookup",
  "flow_id": 123
}
```

Suggested output:

```json
{
  "provider": "ransomware_live",
  "query": "example.com",
  "hits": [],
  "confidence": "low|medium|high",
  "summary": "No matching victim records were found.",
  "recommended_actions": [
    "Keep the finding as informational only.",
    "Do not expand testing scope based on OSINT data alone."
  ],
  "evidence_refs": [],
  "limitations": [
    "External data may be incomplete or delayed."
  ]
}
```

## ransomware.live

As of April 22, 2026, the public ransomware.live API v2 documents these useful endpoint categories:

- `/recentvictims` and `/searchvictims/<keyword>` for recent or keyword-matched victim claims.
- `/groups`, `/group/<group_name>`, and `/groupvictims/<group_name>` for ransomware group context.
- `/recentcyberattacks`, `/countrycyberattacks/<code>`, and `/sectorvictims/<sector>` for country or sector trends.
- `/certs/<country_code>` for national CERT contact context.
- `/yara/<group_name>` for group-associated YARA references.

Relevant docs:

- [ransomware.live API documentation](https://www.ransomware.live/apidocs)
- [ransomware.live API comparison](https://www.ransomware.live/api)

### Scenario: External Recon Risk Context

1. The agent identifies the client's organization name, domain, country, and sector from the authorized target.
2. The agent queries ransomware.live for organization keywords and sector/country context.
3. If a relevant victim or group match exists, the agent summarizes the claim, source, date, and confidence.
4. The agent uses the result to prioritize reporting context and defensive recommendations.
5. The agent does not test newly discovered third-party victims, unrelated domains, or ransomware infrastructure unless those assets are already in scope.

Expected output:

- "Known ransomware exposure context" section in the report.
- A list of matching victim records or "no match found".
- Defensive recommendations such as incident-response readiness, backup review, or exposure monitoring.
- Clear limitations that ransomware.live data is external intelligence, not proof of compromise.

### Scenario: Ransomware Group TTP Enrichment

1. The agent sees a ransomware group name in user-provided context or a confirmed incident artifact.
2. The agent queries group details, group victims, and YARA references.
3. The agent maps relevant context into the report: affected sectors, common targeting pattern, and available detection references.
4. The agent keeps YARA references as defensive context unless the user explicitly asks for detection engineering.

Expected output:

- Group summary.
- Potential detection references.
- Remediation and monitoring suggestions.
- No automatic exploit or campaign emulation.

## Flare

Flare is better suited to customer-owned exposure monitoring because it works with tenants, identifiers, credentials, events, searches, and intelligence feeds. A PentAGI integration should assume the user already has a Flare account, API key, tenant context, and authorization to query exposures for the target organization.

Relevant docs:

- [Flare API getting started](https://api.docs.flare.io/introduction/getting-started)
- [Flare global event search](https://api.docs.flare.io/guides/global-search)
- [Flare tenant events](https://api.docs.flare.io/guides/tenant-events)
- [Flare tenant credentials](https://api.docs.flare.io/guides/tenant-credentials)
- [Flare leaked cookie monitoring](https://api.docs.flare.io/guides/cookie-monitoring)
- [Flare intelligence feeds](https://api.docs.flare.io/guides/ioc-feeds)

### Scenario: Domain Exposure Enrichment

1. The agent discovers or receives an in-scope domain such as `example.com`.
2. The agent queries Flare for matching tenant events, identifier events, global search results, leaked credential summaries, leaked cookie summaries, or IOC feed matches.
3. The agent normalizes results into exposures tied to the in-scope domain.
4. The agent recommends defensive actions such as credential rotation, session invalidation, takedown review, or SOC escalation.
5. The agent does not attempt to log in with leaked credentials, replay cookies, contact exposed users, or test unrelated assets.

Expected output:

- Exposure summary grouped by domain, credential, cookie/session, source, and first/last seen time.
- Recommended defensive action per exposure.
- Evidence references that let the user verify the Flare record in their own tenant.
- A note that exposure intelligence should not expand the active pentest scope without explicit approval.

### Scenario: Threat Intel to Report Context

1. The agent finds an exposed service or technology during an authorized assessment.
2. The agent queries Flare global search or intelligence feeds for matching domains, IPs, URLs, or malware-related IOCs.
3. The agent adds relevant context to the report only when the match is clearly tied to the in-scope organization.
4. The agent marks weak or broad matches as informational and avoids treating them as confirmed compromise.

Expected output:

- "External exposure intelligence" report section.
- Confidence rating for each match.
- Defensive next steps for the customer.
- No autonomous containment, takedown, or credential use.

## Guardrails for Any OSINT Provider

- Query only indicators connected to the authorized target or provided by the user.
- Do not expand testing scope based on third-party OSINT records.
- Do not use leaked credentials, cookies, tokens, or personal data for authentication attempts.
- Respect provider terms, rate limits, tenant boundaries, and API key permissions.
- Cache or summarize results carefully because threat intelligence can include sensitive personal data.
- Mark external intelligence as context unless another in-scope test confirms impact.

## Suggested First Implementation

Start with a provider-neutral external function wrapper:

- `query_osint_intel` accepts provider, indicator, indicator type, purpose, and optional tenant context.
- Provider adapters normalize ransomware.live and Flare responses into the common output shape above.
- Agents receive summaries and evidence references, not raw credential or cookie secrets.
- Results are stored as report context and optional flow evidence, not as automatic new targets.

This keeps the first contribution small while answering whether OSINT enrichment improves PentAGI outcomes before adding a larger native integration.
