# RFC: Native Flow Queue and Completion Webhooks

Issue [#298](https://github.com/vxcontrol/pentagi/issues/298) proposes a native queue with a configurable concurrency cap and a completion webhook so external orchestrators can run many flows against PentAGI without spawning all containers at once and without writing custom polling loops. This RFC captures one possible design direction without committing the project to an implementation.

## Context and Constraints

PentAGI today exposes flows through GraphQL and REST. There is no built-in concurrency control: every accepted `createFlow` request starts a Kali container and a flow runner immediately. There is also no native completion notification, so external schedulers poll `GET /api/v1/flows/{id}` until the status changes.

This RFC is shaped by lessons from PR [#268](https://github.com/vxcontrol/pentagi/pull/268), which added an in-memory input queue for running flows and was rejected because the queue was hidden lifecycle state: not persisted, not visible in UI/API/DB, not cancelable, and not durable across restarts.

The same anti-patterns must be avoided here. Any flow queue must be:

- **Persistent** across restart, crash, and stop.
- **Visible** in the database, GraphQL/REST, and the UI.
- **Manageable** by the user (inspect, cancel, reorder).
- **Explicit** about the definition of "finished" and the moment the next flow is promoted.

## Goals

- Let operators cap concurrent active flows with a single configuration knob.
- Make queued and active flows first-class lifecycle states that are visible everywhere a flow is visible today.
- Eliminate the need for external polling by delivering at-least-once completion notifications to a configured endpoint.
- Keep the existing `createFlow` contract usable for callers that do not care about queueing.
- Preserve existing flow, task, subtask, toolcall, and report behavior unchanged for active flows.

## Non-Goals

- This RFC does not introduce any in-memory or hidden background queue.
- This RFC does not propose multi-tenant scheduling, fairness, priorities across users, or SLA semantics.
- This RFC does not propose a generic event bus; webhooks here only fire on terminal flow status transitions.
- This RFC does not redefine what "finished" means for individual tasks, subtasks, or toolcalls.
- This RFC does not commit to a final database schema or migration strategy.

## Design Principles

1. **Persistence first.** A queued flow is a row, not a goroutine. If the backend restarts, queued flows are still queued and active flows are still claimed.
2. **Visibility everywhere.** Every queued flow appears in the same listing, filtering, and detail surfaces as running flows.
3. **Manageability.** A queued flow can be inspected, canceled, or removed by the same actor and the same authorization that already controls flows.
4. **Explicit promotion.** The transition from queued to running is performed by an explicit promoter, not by an implicit timer or input pump. The promoter only runs when a slot frees up.
5. **Clear finished semantics.** A flow is "finished" when its status reaches `finished` or `failed`. Webhooks and queue promotion read from that single source of truth.
6. **At-least-once delivery, not exactly-once.** Webhook receivers must be idempotent. Delivery state is persisted so retries survive restarts.

## Proposed Concurrency Model

### Lifecycle additions

Introduce a new persisted flow status, for example `queued`, that sits before `running`:

```text
created -> queued -> running -> waiting -> finished
                                          \-> failed
```

A flow enters `queued` when a `createFlow` call would otherwise exceed the configured concurrency cap. A queued flow:

- has a row in the existing flows table,
- is associated with the same user, provider, and resources as a running flow,
- has not yet allocated a Kali container,
- is visible everywhere active flows are visible,
- can be canceled by the user, which transitions it directly to `failed` with a documented reason such as `cancelled_in_queue`.

The `created` status is preserved for the brief window before the dispatcher decides whether to promote directly to `running` or to `queued`.

### Configuration knob

A single environment variable, for example `MAX_CONCURRENT_FLOWS`, controls the cap. Default `0` keeps current behavior (no cap). When the cap is set, a global counter on `running` plus `waiting` flows decides whether new flows start immediately or land in `queued`.

### Promotion

A single promoter component, owned by the existing flow controller, handles transitions. It runs in two situations:

- when a flow leaves `running`/`waiting` (transition into `finished` or `failed`), and
- when a new `queued` flow is created while the cap is below the limit.

The promoter selects the oldest `queued` flow whose owner still has capacity (per-user limits are an open question), allocates the container, and transitions it to `running` using the existing flow start path. Promotion is a single transaction that flips the row and records a promoter audit field.

### User and API visibility

Add `queued` to:

- the `FlowStatus` GraphQL enum and the REST flow representation,
- the flow listing, filter, and search surfaces in the UI,
- the existing flow detail page header,
- the assistant view, where the message reads explicitly "queued, waiting for a free slot" rather than treating the flow as running.

A queued flow exposes the same actions as a created flow (inspect, cancel) and explicitly does **not** show actions that imply a running container (terminal, file uploads to `/work`).

### Cancellation and removal

Cancellation of a queued flow goes through the same authorization as deleting or stopping an active flow today and writes a clear terminal status. The flow does not silently disappear from listings.

## Proposed Completion Webhook Model

### Trigger

A webhook fires when a flow row transitions to a terminal status (`finished` or `failed`). Transitions to `queued`, `running`, `waiting` do not fire webhooks in v1; they remain observable via the existing GraphQL subscription.

### Configuration

Two layers, in priority order:

1. **Per-flow webhook**: an optional `webhookUrl` argument on `createFlow` (REST and GraphQL). The value is persisted on the flow row.
2. **Global webhook**: a deployment-level URL set via environment variable (for example `FLOW_WEBHOOK_URL`). Used when the flow has no explicit webhook URL.

If neither is set, webhooks are disabled for that flow and the existing notifier paths are unchanged.

### Payload shape

A small JSON document. Example:

```json
{
  "event": "flow.completed",
  "delivered_at": "2026-04-22T00:00:00Z",
  "flow": {
    "id": 4242,
    "status": "finished",
    "user_id": 7,
    "title": "Assess https://target.example",
    "started_at": "2026-04-21T23:30:00Z",
    "finished_at": "2026-04-22T00:00:00Z",
    "report_available": true
  },
  "delivery_id": "wh_01h..."
}
```

The payload intentionally does not embed the full report or evidence bundle. Receivers fetch the report through the existing API once they know the flow finished.

## Safety and Security

- Sign each delivery with HMAC-SHA256 using a shared secret (for example `FLOW_WEBHOOK_SECRET`) and place the signature in an `X-PentAGI-Signature` header. Verification is the receiver's responsibility.
- Send a stable `delivery_id` so receivers can deduplicate retries.
- Validate outbound URLs at config load. Forbid unspecified, link-local, and loopback addresses by default to limit accidental SSRF surfaces; allow opt-in via an explicit allowlist.
- Reject webhook URLs whose hostnames resolve into reserved or private ranges, evaluated at delivery time, not just at config time, to defend against late-binding DNS rebinding.
- Treat webhook URLs and secrets as sensitive material in logs and audit trails; redact them in error messages and failure surfaces.
- Bound retry attempts and backoff so a misbehaving receiver cannot drive unbounded outbound load from PentAGI.
- Do not retry on 4xx responses other than 408/429; retry on 5xx and on transport errors. This avoids amplifying receiver-side bugs into traffic floods.

### Delivery durability

Persist a row per delivery attempt with `flow_id`, `delivery_id`, `url`, `status_code`, `attempt`, and `created_at`. A worker retries failed deliveries with bounded exponential backoff up to a cap (for example five attempts), then marks the delivery as failed and surfaces it in the flow detail view. Operators must be able to see "this webhook attempted to fire and could not reach the receiver" without reading logs.

## Storage Sketch

This is illustrative; the final shape is up to the implementing PR.

- Extend `flow_status` enum with `queued`.
- Add columns or sidecar tables for `webhook_url`, `webhook_secret_id`, and `cancelled_reason`.
- Add a `flow_webhook_deliveries` table keyed by `delivery_id`.
- Promotion writes to an existing or new audit field documenting why a flow moved.

No new column is added to support hidden background state.

## API Surface Sketch

- `createFlow` accepts an optional `webhookUrl`. The mutation returns a flow whose status is either `running`, `created`, or `queued` depending on the cap.
- A new query, for example `flowQueue`, lists flows in `queued` status with pagination. It is filterable by user when authorized.
- `GET /api/v1/flows` already exposes status and continues to be the primary REST surface; `queued` is just another value.
- Existing GraphQL subscriptions for flow status changes are extended to emit `queued -> running` transitions.

## Open Questions

- Does PentAGI need per-user concurrency limits in addition to a global cap? If yes, where do they live (user table, role, configuration)?
- Should queued flows be canceled automatically on long-running shutdowns, or always preserved?
- Should `createFlow` block synchronously while queued, or always return immediately with the queued ID?
- Should the webhook fire only on terminal transitions, or also on `running -> waiting` for human-in-the-loop checkpoints?
- Should the per-flow webhook URL be plaintext on the flow row, or referenced through a stored credential record?
- Is HMAC-SHA256 sufficient, or should the signature scheme align with the receipt signing direction in [#235](https://github.com/vxcontrol/pentagi/issues/235)?
- Should `queued` flows reserve container CPU/memory budget, or only count against a slot count?
- How should resources, file uploads, and assistant messages be handled if they arrive against a queued flow before promotion?

## Suggested First Milestone

A narrow, end-to-end skeleton without speculative features:

1. Add the `queued` flow status (database, GraphQL enum, REST, UI badge).
2. Add `MAX_CONCURRENT_FLOWS` and the simple global counter; promote queued flows when slots free up.
3. Surface queued flows in the existing flow list, with filter and explicit "Cancel" action.
4. Document the new lifecycle in `README.md` and `backend/docs/flow_execution.md`.
5. Defer webhooks to a follow-up milestone once the queue lifecycle is stable, then add `FLOW_WEBHOOK_URL`, per-flow `webhookUrl`, signed payloads, and a deliveries table in one focused PR.

This staging keeps PR sizes small, lets each lifecycle change land with full visibility, and avoids bundling concurrency control with delivery semantics in a single review.
