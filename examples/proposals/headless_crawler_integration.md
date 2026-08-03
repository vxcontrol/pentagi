# RFC: Headless Crawler Integration for PentAGI

## Summary

Issue [#336](https://github.com/vxcontrol/pentagi/issues/336) asks for
a headless crawler / URL discovery capability. The current web testing
flow can brute-force directories with dictionary tools such as `ffuf`,
but it cannot crawl an application to discover its real routes, links,
forms, parameters, and JavaScript endpoints. Dictionary fuzzing guesses
paths from a wordlist; crawling observes the paths the application
actually exposes. The two are complementary, and PentAGI currently has
no first-class, scoped, structured crawler capability.

This RFC is a design proposal only. It does not add runtime code, a new
tool handler, a crawler client, GraphQL schema, database migrations,
frontend UI, installer logic, provider configuration, Docker image
changes, Docker Compose changes, `.env.example` entries, or generated
files. It proposes how PentAGI could expose crawler-style URL discovery
as an optional capability so maintainers can review scope, safety, and
the artifact model before any implementation lands.

The proposed direction is a tool-agnostic "crawler backend" (or
"discovery tool") abstraction over candidate tools such as `katana`,
`crawlergo`, `rad`, and `jsfinder`, disabled by default, producing
normalized discovery artifacts (URLs, forms, parameters, JavaScript
endpoints, status codes, source page, depth, and scope decision) that
other PentAGI tools, agents, and reports can consume.

This work is deliberately kept separate from the BrowserOS MCP browser
backend proposed in issue
[#342](https://github.com/vxcontrol/pentagi/issues/342). That proposal
targets stateful, interactive browser control (navigation, click,
fill/type, form submission, login-page handling). This RFC targets
URL/route/form/link discovery and passive or semi-passive site mapping,
not interactive session automation. The two can complement each other,
and JavaScript-heavy crawling is the natural overlap point, but their
contracts are different: a crawler returns a discovery artifact, while
an interactive browser drives a session.

## Goals

- Add a reviewable design for an optional crawler / URL discovery
  capability before any runtime work begins.
- Treat the crawler as a tool-agnostic backend with one or more
  selectable discovery tools, not a single hardcoded CLI.
- Produce normalized, structured discovery artifacts that capture URLs,
  HTTP methods, status codes, forms, parameters, JavaScript endpoints,
  the source page, crawl depth, and the in-scope/out-of-scope decision.
- Let PentAGI agents use crawler output to seed `ffuf`/`dirsearch`-style
  content discovery, guide browser checks, reduce repeated manual
  enumeration, and enrich reports.
- Keep the capability secure by default: disabled unless enabled, scoped
  to the authorized target, rate limited, and bounded by depth, page,
  and request limits.
- Reuse PentAGI's existing URL classification and screenshot/evidence
  paths where practical instead of inventing parallel infrastructure.
- Keep crawler discovery clearly separated from the BrowserOS MCP
  interactive browser backend.

## Non-Goals

- This RFC does not implement crawler runtime code, a crawler tool
  handler, or a crawler client.
- This RFC does not add Docker image changes, Docker Compose services,
  installer behavior, `.env.example` entries, new environment variables,
  GraphQL types, REST endpoints, database tables, generated files,
  provider configuration, or frontend settings.
- This RFC does not replace `ffuf`, `dirsearch`, or any existing content
  discovery workflow. Crawling and dictionary fuzzing are complementary.
- This RFC does not provide stateful interactive browser automation such
  as clicking through UI, typing into fields, or driving multi-step
  login flows. That direction belongs to the BrowserOS MCP browser
  backend in issue #342.
- This RFC does not pick a mandatory default crawler tool. `katana`,
  `crawlergo`, `rad`, and `jsfinder` are framed as candidates.
- This RFC does not propose credentialed crawling or automatic active
  form submission as default behavior.
- This RFC does not choose the final storage shape for crawler
  configuration or artifacts.

## Current Browser and Discovery Behavior

PentAGI already has two relevant building blocks, but neither performs
structured crawling.

The browser tool in `backend/pkg/tools/browser.go` is scraper-backed and
request/response oriented. It operates on a single URL at a time and
supports the `markdown`, `html`, and `links` actions, with an optional
screenshot per call through the configured scraper. `SCRAPER_PUBLIC_URL`
and `SCRAPER_PRIVATE_URL` select public or private scraper routing based
on whether the target resolves to a private or public address, and the
tool already classifies local zones and binary URLs to avoid fetching
non-HTML resources. The `links` action returns the links found on one
fetched page, but the browser tool does not recursively follow links,
build a site map, deduplicate routes, extract forms or parameters,
discover endpoints referenced only from JavaScript, or enforce a crawl
scope or budget. It is single-page extraction, not crawling.

The terminal tool lets the pentester agent run arbitrary command-line
tools inside the isolated pentest container. The pentester prompt's web
testing guidance already lists crawler-adjacent tools (for example
`ffuf`, `gobuster`, `dirsearch`, `feroxbuster`, `httpx`, `katana`,
`hakrawler`, `waybackurls`, and `gau`) as examples the agent may use,
with a reminder to verify availability and install missing tools in the
current image. This means a crawler such as `katana` can in principle be
invoked today, but only as an ad-hoc terminal command. Its output is
unstructured stdout, it is not scoped or budgeted by PentAGI, the
results are not normalized into a reusable artifact, and nothing feeds
the discovered routes back into `ffuf`, the browser tool, or the report.

The gap from issue #336 is therefore not "no tool can crawl" but "there
is no first-class, scoped, structured crawler capability." Discovery
today depends on whichever CLI the agent happens to run, returns
free-form text, and cannot be reliably reused across subtasks.

## Proposed Crawler Capability

A future implementation can introduce a crawler / URL discovery
capability that is conceptually parallel to the existing browser tool:
it wraps a discovery backend, accepts one or more seed URLs plus a
scope, and returns a normalized discovery artifact instead of free-form
text.

Proposed shape of the capability:

- Input: one or more seed URLs, an explicit scope (allowed hosts or
  scope entries), a crawl mode, and budget limits (maximum depth, pages,
  requests, duration, and request rate).
- Crawl mode candidates:
  - `passive`: derive URLs from already-collected sources such as link
    extraction, archived URL feeds, or saved responses, without sending
    new requests to the target beyond what is already authorized.
  - `static`: request and parse HTML and linked resources without a
    full browser engine.
  - `headless`: render pages with a headless browser engine so that
    JavaScript-built routes, single-page-application views, and
    dynamically generated links can be discovered.
- Output: a structured discovery artifact (see Artifacts and Reporting)
  that records each discovered URL with its method, status code, source
  page, depth, parameters, forms, JavaScript endpoints, the backend that
  found it, and the scope decision.
- Agent-facing result: a concise text summary for the agent (counts,
  notable routes, limits reached) plus the structured artifact stored
  through PentAGI's existing artifact and memory paths, mirroring how the
  browser tool returns text while saving screenshots.

The capability is observation-focused. It maps what exists; it does not
log in, submit credentials, or perform stateful multi-step interaction.
Where a target genuinely requires JavaScript rendering or authenticated
interaction to reveal routes, that overlaps with the BrowserOS MCP
backend (issue #342); this RFC treats such cases as a handoff boundary,
not as crawler scope creep.

## Candidate Tools

The tools named in issue #336 cover complementary discovery styles. They
are candidates for a pluggable backend, not mandated defaults.

- `katana` (ProjectDiscovery): fast crawler with both a standard and a
  headless mode, configurable depth and scope, form and endpoint
  awareness, and structured (JSON/JSONL) output. It is already named in
  the pentester prompt's web testing list, which makes it a natural
  first candidate for a structured backend.
- `crawlergo`: a Chromium-based dynamic crawler aimed at
  JavaScript-heavy applications. It exercises the DOM to surface routes
  and requests that static crawling misses and emits a request list that
  can seed further testing.
- `rad`: a Chromium-based crawler focused on browser-driven crawling
  with scope controls, useful where rendering is required to enumerate
  application routes.
- `jsfinder`: an endpoint/URL extractor that mines JavaScript files for
  references to API paths and routes. It is semi-passive and complements
  crawlers by recovering endpoints that are present in scripts but never
  linked from rendered pages.

Adjacent tools already referenced by the pentester prompt, such as
`hakrawler`, `waybackurls`, and `gau`, could be modeled as additional
passive discovery sources behind the same abstraction. The RFC does not
require adopting all of them; it requires that the abstraction not be
hardwired to one CLI.

## Tool Selection Model

The capability should expose a `crawler backend` (equivalently a
`discovery tool`) abstraction so that the agent-facing tool surface is
stable while the underlying CLI can vary.

Proposed selection model:

- Default is off. No crawler backend is selected and no crawl runs
  unless an operator explicitly enables at least one backend.
- Operators register which discovery tools are available and enabled in
  the deployment, since tool availability depends on the pentest image.
- Each backend declares the crawl modes it can serve (`passive`,
  `static`, `headless`) and any tool-specific limits.
- The agent (or a future planner step) selects an enabled backend and
  mode appropriate to the subtask. If a requested mode has no enabled
  backend, the capability reports that clearly instead of silently
  falling back to an unsafe or out-of-scope behavior.
- All backends normalize into the same discovery artifact schema so that
  downstream consumers do not depend on which CLI produced the result.

No single tool is promoted to a required default. `katana` is a
reasonable first candidate because it already appears in PentAGI's tool
guidance and supports structured output, but the RFC frames it as one
option among several, not a mandate.

## Agent Workflow

A structured crawler changes how the pentester agent approaches a web
target. Today the agent fuzzes directories and may run an ad-hoc crawler
whose text output is hard to reuse. With a structured capability, a web
subtask can run discovery once, persist the artifact, and let later
subtasks consume it.

Illustrative workflow:

1. Early in a web engagement, the pentester agent runs crawler discovery
   against the authorized target within scope and budget limits.
2. The normalized artifact is stored through PentAGI's artifact and
   long-term memory paths so it is reusable across subtasks rather than
   re-derived.
3. Discovered directories, file extensions, and parameter names seed
   `ffuf`/`dirsearch`-style content discovery, making dictionary fuzzing
   more targeted instead of purely wordlist-driven.
4. Interesting routes (login, upload, admin, API, parameterized
   endpoints) guide focused browser checks, and, where interactive
   rendering or login is required, hand off to the BrowserOS MCP backend
   (issue #342).
5. Parameterized URLs and JavaScript endpoints inform vulnerability
   testing tools the agent already uses (for example `nuclei` or
   `sqlmap`) by giving them concrete inputs.
6. The final report is enriched with a route/endpoint inventory derived
   from the artifact, reducing repeated manual enumeration and giving
   reviewers a clear map of the attack surface.

This fits PentAGI's existing flow/task/subtask model and result tools
(for example `hack_result` and `report_result`) and its RAG memory,
without introducing hidden lifecycle state. The crawl is an explicit
tool call with a visible, inspectable artifact.

## Artifacts and Reporting

Crawler output should become a normalized artifact rather than raw
stdout. The schema below is illustrative, not a final contract.

Per-entry fields:

- `url`: the discovered URL.
- `method`: HTTP method observed or inferred (for example GET or POST).
- `status_code`: response status when the crawler requested the URL.
- `source_page`: the page the URL was discovered from.
- `depth`: crawl depth at which the URL was found.
- `scope_decision`: whether the URL was treated as in scope or out of
  scope, and why, modeled as a `decision` value plus a short `reason`.
- `parameters`: query or body parameter names associated with the URL.
- `forms`: form action, method, and input field names.
- `js_endpoints`: endpoints referenced from JavaScript for this page.
- `discovered_by`: which backend (for example `katana` or `jsfinder`)
  produced the entry.

An illustrative entry:

```json
{
  "url": "https://target.example/admin/login",
  "method": "GET",
  "status_code": 200,
  "source_page": "https://target.example/",
  "depth": 1,
  "scope_decision": { "decision": "in_scope", "reason": "same_origin" },
  "parameters": ["redirect", "lang"],
  "forms": [
    {
      "action": "https://target.example/admin/login",
      "method": "POST",
      "inputs": ["username", "password", "csrf_token"]
    }
  ],
  "js_endpoints": ["https://target.example/api/v1/session"],
  "discovered_by": "katana"
}
```

A run-level summary should accompany the entries: seed URLs, scope
configuration, backend and mode used, counts (pages crawled, URLs,
forms, parameters, JavaScript endpoints), which limits were reached, and
how many entries were dropped as out of scope. Artifacts should be
stored through the same flow-scoped artifact, screenshot, and reporting
paths the browser tool already uses where practical. If the evidence
chain proposal in
[evidence_chain.md](evidence_chain.md) is implemented later, a crawl run
and its artifact are a natural source of toolcall receipts.

## Scope and Safety

Crawling reaches more of a target than single-page scraping, so the
capability must be secure by default and explicitly scoped.

Required safety properties:

- Disabled or explicit by default. No crawl runs unless an operator
  enables a backend and the agent invokes it deliberately.
- Obey flow target scope. Crawling must respect the authorized target
  scope of the flow and must not wander to unrelated hosts.
- Depth, page, and request limits. Every crawl is bounded by maximum
  depth, maximum pages, maximum total requests, and maximum duration.
- Rate limiting. Request rate and concurrency are capped to avoid
  hammering the target or tripping protective controls.
- Same-origin and allowed-host controls. Off-origin and off-allowlist
  links are recorded as out of scope by default rather than followed.
- SSRF and private-network protection. The capability should reuse the
  browser tool's URL classification so link-local, loopback, metadata,
  private, and reserved targets are not crawled unless explicitly
  authorized for the engagement.
- `robots.txt` is treated as an operator policy question, not a hard
  rule. Authorized engagements differ on whether to honor `robots.txt`,
  so the policy (honor, ignore, or record only) should be configurable
  and recorded in the artifact, and the default should be conservative.
  This RFC does not mandate either honoring or ignoring it.
- Avoid crawling outside authorized targets. Out-of-scope discovery is
  recorded for awareness but not actively requested.
- No credentialed crawling by default. The crawler should not assume it
  can use stored credentials, cookies, or authenticated sessions.
- Active form submission requires separate approval or policy. Passive
  discovery of forms is in scope; submitting forms (which mutates target
  state) is a separate, gated capability and is not default behavior.

These properties keep the capability aligned with PentAGI's lawful
pentesting posture: it maps the authorized attack surface, it does not
broaden scope, automate accounts, or take state-changing actions without
an explicit decision.

## Configuration Sketch

This sketch is illustrative only. It is not a proposed `.env.example`
change and does not choose the final storage shape.

```yaml
crawler:
  enabled: false # off by default; operator opt-in
  default_backend: none # default crawl backend; values: none, katana,
                        # crawlergo, rad. Passive extractors such as
                        # jsfinder are not selectable as the default.
  backends:
    katana:
      enabled: false
      modes: [static, headless] # advertised crawl modes
    crawlergo:
      enabled: false
      modes: [headless]
    rad:
      enabled: false
      modes: [headless]
    jsfinder:
      enabled: false
      modes: [passive] # JavaScript endpoint extraction
  scope:
    follow: same_origin # same_origin | allowlist | scope_entries
    allowed_hosts: # used when follow=allowlist
      - target.example
    scope_entries: # used when follow=scope_entries; illustrative flow
      - https://target.example # scope-of-work entries, ideally sourced
      - target.example/app # from the flow target scope if available
    robots_policy: record_only # honor | ignore | record_only
  limits:
    max_depth: 3
    max_pages: 500
    max_requests: 2000
    max_duration_seconds: 600
    requests_per_second: 5
    max_concurrency: 5
  credentials:
    allow_credentialed_crawl: false # no authenticated crawling by default
  active:
    allow_form_submission: false # passive discovery only by default
    require_approval_for_submission: true
```

Implementation notes for a future PR:

- The crawler backend should be selectable so the agent-facing tool
  surface stays stable while the underlying CLI can change.
- Tool identifiers and flags here are illustrative. A real
  implementation must validate the canonical tool invocation and output
  format for each backend before normalizing it.
- Scope and SSRF checks should integrate with PentAGI's existing URL
  classification rather than re-implementing target analysis.
- The final storage mechanism (mounted YAML, database-backed settings,
  existing tool configuration, or another maintainer-approved shape) is
  intentionally left open.

## Failure Modes

A crawler capability should degrade safely and visibly.

Expected failure modes:

- The selected backend is not installed or not available in the current
  pentest image.
- The target is unreachable, blocks the crawler, or rate-limits it.
- A JavaScript-heavy site yields little in `static` mode and needs a
  `headless` backend that is not enabled.
- The crawl hits a trap such as an infinite calendar, faceted search, or
  session-id-in-URL pattern and approaches its budget without progress.
- A link points outside the authorized scope.
- The backend emits a very large result set beyond configured limits.
- The backend returns malformed or unexpected output that cannot be
  normalized.

Recommended behavior:

- Return a clear discovery tool error or partial result to the agent and
  continue the agent loop when possible.
- Record the backend, mode, target, scope decision, and which limit was
  reached.
- Fail closed for scope, allowlist, SSRF, and approval decisions rather
  than retrying with a broader behavior.
- Prefer returning a bounded partial artifact over hanging or crawling
  past configured limits.
- Do not retry in a tight loop; rely on bounded budgets and, if retry is
  added later, on backoff.

## Open Questions

- Should the crawler be exposed as a new first-class tool, as an
  extension of the existing browser tool, or wrapped around terminal
  execution of the chosen CLI?
- Which backend(s) should be vendored into the pentest image by default,
  given that tool availability drives what the agent can actually run?
- Should `headless` crawling reuse a browser engine shared with the
  BrowserOS MCP backend (issue #342), or run as a standalone crawler
  process?
- Where should crawler configuration and artifacts live so operators can
  inspect scope and results without leaking target details?
- How should the discovery artifact integrate with the report,
  long-term memory, and the evidence chain proposal?
- What default rate limits, depth, and page budgets are safe for
  authorized engagements without being so low that discovery is useless?
- Should semi-passive sources (`waybackurls`, `gau`, `hakrawler`) be
  modeled behind the same abstraction or kept as separate passive tools?
- What is the right approval model and UX for the optional active
  form-submission capability?
- How should `robots.txt` policy default be chosen, and should it be set
  per deployment, per flow, or per crawl?

## Incremental Milestones

1. Docs-only RFC.
   - Land this proposal so maintainers can review scope, safety, the
     backend abstraction, and the artifact model before runtime work
     begins.

2. Artifact schema and single-backend prototype.
   - Define the normalized discovery artifact and run summary.
   - Wire one backend (for example `katana` in `static` mode) behind the
     abstraction, scoped and budget-limited, without exposing it to
     agents yet.

3. Scoped discovery tool for the pentester agent.
   - Expose crawler discovery as an explicit, disabled-by-default tool
     with scope, depth, page, request, rate, and duration limits.
   - Persist the artifact through existing artifact and memory paths.

4. Downstream integration.
   - Use the artifact to seed `ffuf`/`dirsearch`, guide browser checks,
     inform parameterized vulnerability testing, and enrich reports.

5. Optional headless and semi-passive backends.
   - Add `headless` backends (for example `crawlergo` or `rad`) and
     JavaScript endpoint extraction (`jsfinder`) for JavaScript-heavy
     targets.
   - Keep `robots.txt` policy explicit, keep credentialed crawling and
     active form submission off by default, and gate any state-changing
     action behind approval.

Refs #336
