# MCP Client Integration RFC

## Summary

Issue [#296](https://github.com/vxcontrol/pentagi/issues/296) proposes that
PentAGI add a generic Model Context Protocol (MCP) client so agents can use
external MCP-compatible security tools as first-class tools, with Burp Suite
Pro as the concrete motivating example. This RFC sketches a possible design
direction for that capability.

This document does not implement runtime behavior. It does not add Go or
JavaScript dependencies, change the tool registry, modify agent
orchestration, change the database schema or generated code, add REST
endpoints, or change Docker Compose or installer behavior. It is a design
surface for maintainers to push back on before any code lands.

The RFC is intentionally staged. MCP is powerful, and the issue's most
direct reading -- auto-discover every MCP tool from every configured
server and expose all of them to every agent -- would land near the
patterns that were pushed back on during PR
[#268](https://github.com/vxcontrol/pentagi/pull/268) review: implicit
lifecycle state, weak operator visibility, and weak operator control.
The proposed v1 below is narrower on purpose so a later implementation
PR can be reviewed in small slices.

## Goals

- Connect configured MCP servers as explicit external tool sources.
- Discover tools from allowlisted MCP servers via the standard MCP
  `tools/list` method.
- Expose only approved MCP tools to agents, alongside native PentAGI
  tools, with a stable namespace.
- Preserve PentAGI's existing tool visibility and auditability
  expectations: operators can see which tools exist, which agent invoked
  which tool, with what inputs and what result.
- Support a staged implementation path that starts with a narrow,
  reviewable milestone (read-only discovery) and grows from there.
- Keep operators in control of network reachability, credentials, and
  tool exposure at every step.

## Non-Goals

- This RFC does not add MCP runtime code, dependencies, registry,
  orchestration, schema, GraphQL, REST, or installer changes.
- This RFC does not propose auto-trusting every tool advertised by a
  configured MCP server. Discovered tools are inert until an operator
  allowlists them.
- This RFC does not propose hidden background execution of MCP tool
  calls. Calls happen inside the same agent loop and are visible the
  same way native tool calls are. This carries forward the explicit
  lessons from PR
  [#268](https://github.com/vxcontrol/pentagi/pull/268) review: no
  hidden in-memory queues, no implicit lifecycle state, no invisible
  background behavior.
- This RFC does not propose a hidden queue, retry buffer, or implicit
  per-tool state machine sitting between agents and MCP servers.
- This RFC does not grant MCP servers arbitrary host or network access
  by default. Reachability is opt-in per server.
- This RFC does not propose a generic plugin marketplace. MCP servers
  are configured by the operator, not browsed.
- This RFC does not guarantee that every MCP transport or server will
  be supported in v1. The proposed v1 is deliberately small.
- This RFC does not change any existing native tool. Native tools keep
  their current behavior.

## Design Principles

- **Explicit configuration over auto-discovery.** Operators name and
  enable each MCP server. Nothing is connected by default.
- **Least privilege and allowlists by default.** Discovered tools are
  inert until allowlisted. The default allowlist is empty.
- **Tool namespace isolation.** MCP tools appear to agents under a fixed
  prefix (proposed: `mcp.<server>.<tool>`) so they cannot collide with
  native tools or with tools from another MCP server.
- **User- and admin-visible tool inventory.** Operators can see every
  configured server, every discovered tool, every allowlisted tool, and
  the last successful discovery time, without reading raw logs.
- **Auditable execution records.** Each MCP tool invocation is logged
  with the same shape as native tool calls: who, when, which flow /
  task / subtask, input, result, and error.
- **Secrets redaction.** Server credentials, headers, and MCP tool
  arguments that look like secrets are never echoed to prompts, logs,
  or UI surfaces.
- **Container and host network boundary clarity.** It is explicit per
  server whether the MCP endpoint sits inside the PentAGI compose
  network, on the host (for example via `host.docker.internal`), or on
  a remote network the operator has chosen to expose.
- **Failure isolation.** A broken or unreachable MCP server surfaces as
  a tool error or a degraded discovery state. It must not crash the
  agent loop, the tool registry, or any native tool.

## Proposed v1 Design

### Configuration model

A future implementation should accept a list of named MCP servers, each
with explicit fields. Names are stable identifiers used in the tool
namespace and the audit trail. Server entries are off by default and
require an explicit `enabled: true` (or equivalent) before any
connection is attempted.

Each entry should describe at minimum:

- a stable server name (used as the namespace prefix segment);
- the transport (see below);
- transport-specific connection details (URL for HTTP/SSE, command and
  argv for stdio);
- optional headers, tokens, or other authentication material, sourced
  from existing secret patterns rather than hard-coded;
- a per-server tool allowlist;
- bounded timeouts and a maximum response size;
- an optional human-readable description shown to operators.

The exact storage shape (env, mounted YAML, DB row, GraphQL settings
mutation) is **not** chosen here. The only constraints this RFC
asserts are that the configuration must be inspectable to operators
and that secrets must not leak into prompts, logs, or UI surfaces.

### Transports

The MCP ecosystem currently uses three primary transports: HTTP, SSE,
and stdio. They have very different operational profiles:

- **HTTP and SSE** are remote-friendly. They fit the existing PentAGI
  deployment shape (containerized backend reaching a service over the
  network) and can target either a host-side endpoint via
  `host.docker.internal` or a remote endpoint chosen by the operator.
  These are the proposed first-class transports for v1.
- **stdio** assumes the MCP client can spawn a child process and pipe
  JSON-RPC over its stdin and stdout. In a containerized PentAGI
  deployment the question of *which* container that child runs in,
  with *which* binaries on PATH, and with *which* network egress, is
  non-trivial. v1 should not promise stdio support beyond the narrow
  case of a server-side managed command shipped or installed by the
  operator. A clean stdio story is a later milestone.

### Tool discovery

Discovery uses the standard MCP `tools/list` method. A successful
discovery records: the set of advertised tool names, their input
schemas, their human-readable descriptions, and a timestamp. A failed
discovery records the error (transport, authentication, schema parse,
timeout) without retrying in a tight loop. Discovery results are
visible to operators independently of agent behavior.

Discovery does not by itself expose any tool to agents. It only updates
the inventory operators can see and allowlist from.

### Tool naming and exposure

Allowlisted tools are presented to agents under a fixed prefix. This
RFC suggests `mcp.<server>.<tool>` -- for example,
`mcp.burp.start_scan` or `mcp.nuclei.run_template`. The prefix is
reserved so it cannot be confused with native tool names, which today
do not start with `mcp.`.

Exposure is allowlist-driven. An MCP tool is callable by agents only
if the operator has both (a) enabled the server and (b) added the
tool to that server's allowlist. The default allowlist is empty.

Per-server and per-tool metadata -- name, description, input schema,
last discovery time, allowlist status, last invocation, last error --
should be visible to operators in whatever surface PentAGI uses for
existing tool and provider configuration. The exact UI surface is
left to the implementation PR; this RFC only asserts that the
information must be visible somewhere operators already look.

### Invocation, audit, and failure surface

An MCP tool call should flow through the same broad permission and
audit mindset as a native tool call. At minimum each MCP invocation
should record the same fields as a native tool call (flow, task,
subtask, agent, tool name, input arguments after redaction, result
summary, error if any, duration).

Failure semantics:

- A transport error, server-side error, or schema mismatch is
  surfaced to the agent as a tool error. The agent loop continues.
- A timeout cancels the in-flight request and is logged as a tool
  error.
- A discovery failure marks the server as degraded and removes its
  previously-allowlisted tools from the agent-facing surface until
  rediscovery succeeds. It does not affect native tools or tools from
  other MCP servers.
- A misbehaving MCP server cannot crash the agent loop or block
  native tools.

### Bounded timeouts and payload limits

Every MCP request must have a per-call timeout and a maximum response
size. Both are configurable per server with conservative defaults. An
overlong response is truncated at the boundary, with a marker, and the
truncation event is recorded in the invocation log so operators can
see that the agent did not receive the full payload.

### Secrets

Server credentials, headers, tokens, and MCP tool arguments that match
known secret patterns must be sourced from existing PentAGI secret
storage where possible and must never appear verbatim in prompts,
agent context, audit logs displayed in the UI, or error messages
surfaced to agents. The exact secret-storage choice (existing env
handling, existing provider-credential patterns, a future dedicated
store) is a question for the implementation PR, not this RFC.

### Schema, API, and UI implications (deferred)

A future implementation may need new database rows for MCP server
records, new GraphQL types for the inventory and allowlist, new audit
log columns, and new UI surfaces. This RFC does **not** choose those
shapes. They are explicitly deferred to the implementation PR(s) so
that the design here can be reviewed without committing to a
migration.

## Burp Suite MCP Example

This section is illustrative. PentAGI does not implement Burp Suite
MCP support today, and shipping this RFC does not change that. The
example exists to make the v1 design concrete against a real MCP
server.

PortSwigger publishes an official Burp MCP server extension at
[`PortSwigger/mcp-server`](https://github.com/PortSwigger/mcp-server),
which the issue body references as the motivating use case.

### Possible operator setup

A pentester running Burp Suite Pro on their own analyst workstation,
or on a controlled internal host, enables the Burp MCP extension. The
extension exposes an MCP endpoint over HTTP on localhost. From a
PentAGI compose deployment running on the same workstation, that
endpoint can be reached at `http://host.docker.internal:<port>` from
the backend container -- but this name is not universally available.
Docker Desktop on macOS and Windows commonly resolves
`host.docker.internal` automatically, while on Linux and other
operator-managed compose stacks the operator typically has to add an
explicit `extra_hosts: - "host.docker.internal:host-gateway"` mapping
to the relevant service, or point the MCP server `url` at another
controlled endpoint (a routed internal IP, a reverse proxy, etc.).
Either way, host reachability for an MCP server should be explicit
and operator-controlled, not implied by the design.

The operator adds a single named server entry (call it `burp`),
explicitly enables it, supplies any required header or token, and
allowlists a small set of read-only tools. Nothing else changes.
Anything beyond read-only -- in particular active scan -- is not part
of this safe initial setup and is described separately below.

### Candidate tool categories

The Burp MCP server, depending on version and configured access, can
expose a range of capabilities. From the issue's framing, useful
read-only categories for PentAGI agents would include:

- reading the current Burp sitemap or list of discovered URLs;
- retrieving structured findings (vulnerability type, severity,
  request and response evidence) for incorporation into the agent's
  reasoning and reports;
- retrieving Collaborator or other out-of-band events, if the MCP
  server exposes them, to detect blind classes of vulnerabilities
  that PentAGI's CLI-only toolset cannot detect today.

Active capabilities -- in particular starting an active scan against
a specific in-scope URL -- are higher risk and intentionally **not**
part of the safe initial example. They belong to a later, explicitly
gated milestone and are discussed under "Scope discipline" below.

### Scope discipline

Active scan tools are not read-only. They send traffic, can be noisy,
and can be destructive against fragile targets. They must be:

- explicitly allowlisted per server;
- only invoked against targets the operator has already declared as
  in scope for the engagement;
- bounded by the same scope and approval mindset PentAGI applies to
  other intrusive tools.

It should not be inferred from this RFC that PentAGI can freely scan
arbitrary targets once Burp MCP is wired in. The agent only ever
invokes a Burp MCP tool that an operator has both configured and
allowlisted, against a target that already exists inside the
engagement's scope.

### Illustrative configuration (non-final pseudocode)

The shape below is **non-final pseudocode** to make the design
tangible. It is not a proposed final environment-variable schema,
YAML schema, or database schema. The implementation PR may use any of
those, or something else.

```text
# illustrative only -- not a final config schema
mcp:
  servers:
    - name: burp
      enabled: true
      transport: http
      url: http://host.docker.internal:1337
      timeout_seconds: 30
      max_response_bytes: 1048576
      auth:
        header: X-API-Key
        secret_ref: burp_mcp_token
      allowlist:
        # safe initial set: read-only tools only.
        # active capabilities like start_active_scan are intentionally
        # excluded here and would be added later, under explicit
        # in-scope target checks and operator approval.
        - get_sitemap
        - get_findings
```

## Security and Safety

MCP support meaningfully expands PentAGI's attack surface and the
agent's reach. The following risks should be addressed by the v1
design rather than left for "later":

- **SSRF and internal network pivot.** A misconfigured `url` value
  can point an enabled MCP server entry at internal infrastructure.
  The implementation should validate URLs at configuration time and
  at request time, and operators should be made aware that an MCP
  server entry effectively grants the agent access to whatever
  network that endpoint sits on.
- **Prompt injection through tool descriptions and results.** MCP
  tool descriptions and results flow into the agent's context. A
  hostile or compromised MCP server can therefore attempt to inject
  instructions the same way a hostile web page can. Tool descriptions
  should be treated as untrusted text, not as instructions the agent
  must follow.
- **Destructive or high-impact MCP tools.** Some MCP tools can modify
  state (start scans, send messages, write to ticketing systems, log
  findings). The allowlist is the primary control here. A future
  per-tool approval step (see Open Questions) may be appropriate for
  the highest-impact tools.
- **Host access via `host.docker.internal`.** This is a real and
  useful capability, but it must be opt-in per server and clearly
  visible to operators. The operator should know that enabling a
  host-reachable MCP server gives the agent reach into the host's
  network namespace.
- **Credential handling and redaction.** Server credentials and any
  tool argument that matches a known secret pattern must be redacted
  in logs, redacted in any UI surface, and never echoed back into
  agent context. This includes tokens passed as headers, query
  parameters, or tool arguments.
- **Scope enforcement for scan targets.** When an MCP tool takes a
  target URL, hostname, or IP, the implementation should be able to
  reuse PentAGI's existing scope notion (target derived from the flow
  or engagement) to refuse out-of-scope inputs, rather than relying
  on the agent to self-police. The exact mechanism is open (see Open
  Questions).
- **Tool output size limits.** The bounded `max_response_bytes` from
  the configuration model also serves as a denial-of-context control:
  an MCP server cannot flood the agent's context with arbitrary
  payloads.
- **Audit trail.** Operators should be able to answer "who enabled
  this server" and "which agent invoked which MCP tool, when, against
  what target" without reading raw logs. Audit records must be tied
  to flow / task / subtask the same way native tool calls are.

Recommended safe defaults:

- All MCP servers disabled by default.
- All allowlists empty by default.
- All timeouts and response-size caps set to conservative values.
- Secrets never echoed into prompts, logs, or UI.
- Discovery never automatically promotes a tool into the allowlist.

## Observability and Auditability

For each configured MCP server, operators should be able to see:

- connection status (enabled / disabled, reachable / unreachable);
- discovery status (last successful discovery time, last error);
- the set of currently advertised tools and their schemas;
- the per-server allowlist and its diff against the advertised set;
- the most recent invocation outcomes, including errors and timeouts.

For each MCP tool invocation, the audit trail should record:

- the flow, task, and subtask the call belongs to;
- the agent that initiated the call;
- the tool name in its namespaced form (`mcp.<server>.<tool>`);
- the redacted input arguments and a redacted summary or hash of the
  output;
- the duration, error, and any truncation events.

MCP tool outputs should map into the existing flow / task / subtask
context the same way native tool outputs do, so that downstream
reporting can treat them uniformly. Future alignment with the evidence
and reporting direction discussed in
[issue #235](https://github.com/vxcontrol/pentagi/issues/235) (see
also the existing `examples/proposals/evidence_chain.md`) is desirable
but should not be a hard dependency for shipping the first MCP
milestone.

## Open Questions

- Which transports should ship first, and what is the minimum
  supported transport set for v1? (The leaning above is HTTP and SSE
  first; stdio later or only as a server-side managed command.)
- Should MCP servers be configured globally for the deployment,
  per-user, or per-flow? Each level has very different implications
  for audit, sharing, and isolation.
- Should MCP tool approval happen entirely in configuration, in the
  UI, or both? A pure-config flow is simpler; a UI flow allows
  per-tool review at first use.
- How should MCP tool output be summarized or truncated when it
  exceeds the response-size cap, both in the agent loop and in the
  audit log?
- How should MCP tool calls fit with PentAGI's existing permission
  and scope controls -- particularly target scope for scan-style
  tools? Should scope checks be enforced by the MCP layer, by the
  tool registry, or by a shared scope service?
- Should MCP tool calls become first-class report and evidence
  artifacts in their own right, beyond the per-call audit log? This
  overlaps with the evidence-chain RFC but does not have to be solved
  in the first milestone.
- How should Burp-specific scope (Burp's own in-scope and
  out-of-scope configuration) be represented to PentAGI without
  hard-coding Burp semantics into core? A generic "scope hint"
  attached to MCP tools is one possibility; a Burp-specific guide is
  another.

## Suggested First Milestone

The first MCP milestone should be deliberately narrow so each
lifecycle change is small enough to review in isolation:

1. **This RFC.** Land the design surface, gather maintainer feedback,
   and freeze the boundaries (namespace, allowlist-by-default, no
   hidden background execution).
2. **Read-only discovery for one configured HTTP or SSE MCP server.**
   The backend can connect to a single named server, run `tools/list`,
   and surface the resulting inventory plus discovery status to
   admins and operators. No agent execution yet, or at most a single
   allowlisted read-only tool whose output is bounded and audited the
   same way native tool calls are.
3. **Allowlisted execution with audit logs and target-scope checks.**
   Once read-only discovery is shipping, extend to allowlisted MCP
   tool invocation by agents, with full audit and with explicit
   handling of target scope for scan-style tools.
4. **Burp-specific operator guide.** After the generic MCP client
   behavior is proven, document the Burp Suite Pro setup as an
   operator guide (similar to
   `examples/proposals/osint-integration-scenarios.md` for OSINT
   providers) -- so Burp lands on top of generic MCP rather than as a
   bespoke integration.

Each milestone is intentionally self-contained: a maintainer can stop
the work after any step without leaving PentAGI in a half-shipped
state.
