# RFC: BrowserOS MCP Browser Backend for PentAGI Agents

## Summary

Issue [#342](https://github.com/vxcontrol/pentagi/issues/342)
proposes optional BrowserOS MCP support so PentAGI agents can use a
stateful browser session for pages that need interaction, JavaScript
rendering, form input, login-page handling, screenshots, and visual UI
inspection.

This RFC is a design proposal only. It does not add runtime code,
BrowserOS client code, MCP client code, GraphQL schema, database
migrations, frontend UI, Docker Compose changes, `.env.example`
entries, or generated files. The current scraper-backed browser tool
remains the default behavior.

The proposed direction is to treat BrowserOS as a specialized browser
backend that builds on the generic
[MCP client integration RFC](mcp_client_integration.md), not as a
separate MCP stack. BrowserOS is an open-source Chromium-based browser
with a documented HTTP MCP endpoint and browser automation tools such as
navigation, page observation, screenshots, clicks, typing/filling, and
dialogs:

- BrowserOS repository: <https://github.com/browseros-ai/BrowserOS>
- BrowserOS MCP client docs:
  <https://docs.browseros.com/features/use-with-claude-code>

## Goals

- Add a reviewable design for a future optional BrowserOS MCP browser
  backend.
- Preserve the existing scraper backend as the default browser
  behavior.
- Support interactive browser workflows in a future implementation:
  open URLs, click elements, fill inputs, submit forms, wait for page
  changes, capture screenshots, and extract visible page state.
- Reuse the generic MCP client direction for server configuration, tool
  discovery, allowlisting, invocation, timeouts, payload limits, and
  auditability.
- Keep operator control explicit for network reachability, target
  scope, sensitive actions, and credential-adjacent behavior.
- Make browser sessions isolated per flow so cookies, local storage,
  tabs, downloads, and screenshots do not leak between flows.
- Ensure every interactive action is visible in logs and observability
  records with flow, task, and subtask context.

## Non-Goals

- This RFC does not implement BrowserOS, MCP, or browser backend runtime
  code.
- This RFC does not replace the current scraper backend.
- This RFC does not expose every BrowserOS MCP tool to agents. BrowserOS
  also documents bookmarks, history, window management, and external app
  integrations; those should stay out of scope by default.
- This RFC does not propose unrestricted browsing, credential
  automation, or automatic login handling.
- This RFC does not add new environment variables, Docker Compose
  services, installer behavior, GraphQL types, REST endpoints, database
  tables, or frontend settings.
- This RFC does not choose the final storage shape for browser backend
  configuration.
- This RFC does not require `host.docker.internal` reachability. Any
  host-side BrowserOS endpoint must be explicitly configured and tested
  by the operator.
- This RFC does not change terminal, search, memory, reporting, or
  evidence-chain behavior.

## Current Browser Tool Behavior

PentAGI's current browser tool is implemented in
`backend/pkg/tools/browser.go`. It is request/response oriented and uses
configured scraper URLs rather than a long-lived browser session.

Current behavior:

- The tool accepts the native `browser` tool name.
- Supported actions are `markdown`, `html`, and `links`.
- Content requests call scraper endpoints such as `/markdown`, `/html`,
  and `/links`.
- Screenshot capture calls the scraper `/screenshot` endpoint and saves
  screenshots under the flow screenshot directory.
- Content extraction and screenshot capture run in parallel for the
  supported actions.
- `SCRAPER_PUBLIC_URL` and `SCRAPER_PRIVATE_URL` select public or
  private scraper routing based on target URL classification.
- Browser errors are wrapped into text for the agent instead of
  hard-failing the entire chain.
- The tool is available only when a public or private scraper URL is
  configured.

This works well for simple extraction and screenshots. It is less suited
for stateful web apps that require clicking through UI, filling forms,
waiting for dynamic changes, observing post-navigation state, or using a
real browser session across several steps.

## Proposed BrowserOS MCP Backend

A future implementation can introduce BrowserOS as an optional backend
behind PentAGI's browser capability. BrowserOS MCP should be configured
as an MCP server and discovered through the same client layer proposed in
the MCP client integration RFC.

The BrowserOS backend should expose a narrow browser-focused tool
surface to PentAGI agents. Candidate action groups:

- Navigation: open a URL, reload, go back, go forward, create or close a
  session-scoped page.
- Observation: get visible text, get page markdown/content, inspect the
  accessibility snapshot, capture screenshot, inspect current URL and
  title.
- Interaction: click, focus, fill/type, clear, select options, press a
  key, scroll, and handle dialogs.
- Page timing: wait for navigation, wait for visible text, wait for a
  selector-like target, or wait for network/DOM quiescence when the MCP
  server supports it.
- Controlled submission: submit forms only when the action is in target
  scope and any configured approval policy is satisfied.

The backend should not expose BrowserOS's broader MCP catalog by
default. In particular, app integrations, bookmarks, browser history,
cloud sync, stored passwords, personal profiles, and cross-application
actions should be blocked unless maintainers later design a separate
opt-in capability with its own review.

The agent-facing result should preserve PentAGI's current expectations:
return a concise text result to the agent, attach screenshots through
the normal screenshot path when available, and surface errors as browser
tool errors that the agent can reason about.

## Backend Selection Model

The default backend remains `scraper`. BrowserOS MCP is disabled unless
an operator explicitly enables it.

Candidate backend modes for a later implementation:

- `scraper`: current behavior. This remains the default and should be
  used when no BrowserOS MCP backend is configured.
- `browseros_mcp`: use the allowlisted BrowserOS MCP browser tool
  surface as the browser backend.
- `hybrid`: use scraper for simple extraction and BrowserOS MCP only
  for explicitly interactive steps or documented scraper failure modes.

Hybrid mode should not silently escalate from static extraction to
interactive browsing without an audit record. A fallback from scraper to
BrowserOS should log the reason, the target URL, the selected backend,
and the flow/task/subtask context. If the failure is caused by target
scope policy, target allowlists, approval requirements, or SSRF
protection, fallback must fail closed rather than trying another backend.

Backend selection should be visible to operators in logs and future UI
surfaces. Agents should receive enough result text to understand which
backend handled the request, but not secrets or raw connection details.

## Session and Flow Isolation

BrowserOS MCP adds state that the scraper backend does not have:
cookies, local storage, tabs, windows, downloads, DOM state, and
possibly authenticated sessions. A future implementation must make this
state explicit and isolated.

Recommended isolation model:

- Each PentAGI flow gets its own BrowserOS session or profile namespace.
- Task and subtask records include the session identifier in audit logs,
  but the raw BrowserOS endpoint and credentials are redacted.
- Browser pages created for one flow are not reused by another flow.
- Cookies, local storage, downloads, and screenshots from one flow are
  not available to another flow.
- Session cleanup is explicit when a flow finishes, fails, is canceled,
  or is deleted.
- If cleanup fails, the failure is logged for the operator and does not
  silently make the session reusable.
- The system should define restart behavior before implementation:
  whether sessions are restored, marked stale, or forcibly closed.

This follows the lesson from earlier flow lifecycle feedback: do not add
hidden state that operators cannot inspect, control, or reason about.
If BrowserOS session state becomes durable, it should be reflected in a
durable and operator-visible design, not only in memory.

## Security and Safety

Browser automation is more sensitive than static page scraping. The
future backend should be secure-by-default and explicitly scoped.

Required safety properties:

- Disabled by default. No BrowserOS connection is attempted unless an
  operator enables the backend.
- Target allowlists. Operators can restrict BrowserOS navigation to
  approved hosts, domains, CIDR ranges, or flow scope entries.
- SSRF protection. BrowserOS navigation must reject link-local,
  metadata-service, loopback, private, and otherwise reserved targets
  unless explicitly allowed for the pentest scope.
- DNS rebinding defense. Hostname allowlist checks should consider
  resolved IP addresses at navigation time, not only string matching at
  configuration time.
- Cross-flow isolation. Cookies, storage, tabs, screenshots, downloads,
  and session handles do not leak between flows.
- Sensitive-action approval. Credential entry, form submission,
  destructive UI actions, file upload, checkout/payment paths, and
  OAuth-style consent screens can require operator approval.
- Secret redaction. Typed values, credentials, tokens, cookies,
  authorization headers, and BrowserOS MCP connection details are
  redacted from prompts, logs, UI, Langfuse metadata, and error text.
- No credential automation by default. The backend should not assume it
  can use stored BrowserOS passwords, personal browser profiles, OAuth
  app integrations, or existing user sessions.
- Bounded execution. Each MCP call has a timeout, maximum response size,
  and cancellation path tied to the agent/tool call context.
- Operator-controlled reachability. Connecting from a PentAGI container
  to a host BrowserOS MCP endpoint must be an explicit deployment
  choice. The design must not assume `host.docker.internal`, host
  network mode, or access to a user's desktop browser.

The backend should preserve PentAGI's lawful pentesting posture:
interactive browsing is useful for authorized targets, but it must not
be framed as a way to bypass site controls, automate personal accounts,
or broaden scope without approval.

## Observability and Audit Logs

Interactive browser actions should be first-class tool events. Operators
must be able to reconstruct what the agent did without replaying the
entire browser session.

Each BrowserOS-backed tool invocation should record:

- flow ID, task ID, subtask ID, and agent/tool identity;
- selected backend (`browseros_mcp` or `hybrid`);
- BrowserOS session/profile identifier after redaction;
- action name, such as navigate, click, fill, submit, wait, extract, or
  screenshot;
- target URL, normalized host, and allowlist decision;
- selector, element ID, or accessibility target when applicable;
- redacted input metadata for fill/type actions, never raw secrets;
- whether an approval gate was required and whether it was approved,
  denied, expired, or skipped by policy;
- result summary, screenshot artifact name, duration, timeout, and
  error category.

Actions such as click, type/fill, submit, navigation, dialog handling,
and file upload should be visible in logs because they mutate browser
state or interact with the target. Screenshots and extracted content
should be stored through the same evidence/reporting paths used by the
current browser tool where practical.

If the evidence-chain RFC is implemented later, BrowserOS actions are a
natural source of toolcall receipts: action input/output hashes,
screenshots, page snapshots, and approval decisions can be linked to the
flow evidence chain.

## Configuration Sketch

This sketch is illustrative only. It is not a proposed `.env.example`
change and does not choose the final storage shape.

The local HTTP URL below mirrors the BrowserOS documentation example,
not a production exposure recommendation. A future implementation should
prefer a localhost-bound BrowserOS MCP listener for same-host use. If an
operator exposes BrowserOS MCP across a network boundary, the endpoint
should require explicit authentication and transport protection, such as
an auth token over TLS or mTLS, and should not be exposed on untrusted
networks.

```yaml
browser:
  backend: scraper # candidate values: scraper, browseros_mcp, hybrid

  browseros_mcp:
    enabled: false
    mcp_server: browseros
    transport: http
    url: http://127.0.0.1:9239/mcp # local docs example; operator supplied
    tool_allowlist: # illustrative names; validate via MCP tools/list
      - navigate_page
      - get_page_content
      - get_page_links
      - take_snapshot
      - take_screenshot
      - click
      - fill
      - press_key
      - scroll
      - handle_dialog
    target_allowlist:
      - https://target.example
    require_approval_for:
      - credential_entry
      - form_submit
      - file_upload
      - oauth_consent
      - destructive_action
    session_scope: flow
    timeout_seconds: 30 # 30s per MCP action
    max_response_bytes: 1048576 # 1 MiB page snapshot/result cap
```

Implementation notes for a future PR:

- The BrowserOS MCP server entry should use the generic MCP client
  configuration model where possible.
- BrowserOS tool identifiers in this sketch are illustrative. The final
  implementation must validate canonical names and schemas through MCP
  discovery before exposing or mapping them to PentAGI actions.
- The tool allowlist should be narrow by default and explicitly exclude
  app integrations, history, bookmarks, and profile-management tools.
- The target allowlist should integrate with PentAGI's flow scope model
  if such a scope source is available at implementation time.
- Any credentials or headers needed to reach BrowserOS MCP should use
  existing secret-handling patterns and must be redacted everywhere.
- The final storage mechanism could be mounted YAML, database-backed
  settings, existing provider/tool configuration, or another
  maintainer-approved shape. This RFC does not choose one.

## Failure Modes and Fallbacks

BrowserOS MCP failures should degrade safely and visibly.

Expected failure modes:

- BrowserOS MCP endpoint is unreachable from the PentAGI backend.
- The endpoint is configured with a host-only address that containers
  cannot reach.
- The MCP server is reachable but tool discovery fails.
- A required browser tool is missing or not allowlisted.
- Navigation is denied by target policy or SSRF protection.
- A page requires login, MFA, CAPTCHA, or operator interaction.
- An action waits for an element that never appears.
- BrowserOS returns a large page snapshot or screenshot beyond configured
  limits.
- Session cleanup fails after a flow completes.

Recommended behavior:

- Return a clear browser tool error to the agent and continue the agent
  loop when possible.
- Log the backend, action, target, flow/task/subtask context, and error
  category.
- Fail closed for policy, allowlist, approval, and SSRF decisions.
- Fall back to scraper only when the requested action can be represented
  safely by scraper behavior, such as markdown/html/links/screenshot
  extraction.
- Do not fall back from BrowserOS to scraper for interactive actions
  such as click, fill, submit, dialog handling, or login-page workflows.
- Do not retry in a tight loop. Use bounded timeouts and backoff if
  retry is added later.

## Open Questions

- Should BrowserOS be exposed as a replacement implementation of the
  existing `browser` tool, as a separate `browseros` tool, or as
  MCP-namespaced tools only?
- Should `hybrid` mode exist in v1, or should maintainers first review a
  strict `scraper` versus `browseros_mcp` selection model?
- Where should the browser backend configuration live so operators can
  inspect it without leaking secrets?
- How should BrowserOS sessions map to PentAGI flows, tasks, and
  subtasks if a flow runs multiple parallel browser tasks?
- What is the right approval UX for credential entry, form submission,
  OAuth consent, file upload, and destructive UI actions?
- Should BrowserOS screenshots be stored exactly like scraper
  screenshots, or should page snapshots and DOM extracts become separate
  artifacts?
- How should target allowlists integrate with future scope-of-work,
  evidence-chain, or policy features?
- Should BrowserOS MCP support be HTTP-only at first, or inherit every
  transport supported by the generic MCP client implementation?
- Should a future implementation support headless/hidden BrowserOS pages,
  or require visible browser state for operator trust?
- How should cleanup behave after backend crashes, PentAGI restarts, or
  BrowserOS is closed externally?

## Incremental Milestones

1. Docs-only RFC.
   - Land this proposal so maintainers can review scope, safety, and
     backend-selection direction before runtime work begins.

2. Discovery-only prototype.
   - Configure a BrowserOS MCP server through the generic MCP client
     path.
   - Discover BrowserOS tools and show which browser tools would be
     allowlisted.
   - Do not expose tools to agents yet.

3. Read-only BrowserOS backend.
   - Support navigation, visible page extraction, link extraction,
     snapshots, and screenshots.
   - Keep scraper as the default backend.
   - Store screenshots through the existing screenshot artifact path.

4. Interactive actions with approvals.
   - Add click, fill/type, key press, scroll, dialog handling, and
     submit-style actions.
   - Require policy checks and approval gates for sensitive actions.
   - Log every state-mutating browser action.

5. Hybrid mode evaluation.
   - Add explicit, audited fallback rules only after the standalone
     BrowserOS backend behavior is understood.
   - Keep policy failures fail-closed and do not use fallback to bypass
     scope or approval decisions.

## Recommendation and Requested Decision

This RFC recommends adopting BrowserOS MCP as an optional, disabled-by-default browser backend while the scraper backend remains the default, implemented incrementally (milestones 1-3 first):

- **Default:** the scraper stays the default backend; BrowserOS MCP is opt-in via configuration, so existing deployments are unaffected.
- **Transport and security:** bind the MCP endpoint to localhost by default and require an auth token (or mTLS) for any remote endpoint; treat the tool identifiers in this RFC as illustrative until validated against a real BrowserOS MCP server.
- **Scope for the first PR:** read-only browsing (navigation, page and link extraction, snapshots, and screenshots through the existing artifact path). State-mutating interactive actions (click, fill, submit) are deferred to a later, approval-gated milestone.
- **Isolation and audit:** per-flow session isolation, fail-closed policy checks, and audit logging of every state-mutating action.

Decision requested from the maintainer: approve BrowserOS MCP as an optional, default-off backend with this read-only-first scope, request changes, or decline. Approval unblocks an implementation PR for the read-only backend behind a disabled-by-default flag.

Refs #342
