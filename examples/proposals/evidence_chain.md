# RFC: Cryptographic Evidence Chain for PentAGI Operations

Issue [#235](https://github.com/vxcontrol/pentagi/issues/235) proposes signed evidence receipts for PentAGI pentest operations. This RFC captures a possible design direction without committing the project to an implementation.

## Goals

- Make reports verifiable against the tool actions, artifacts, and agent decisions that produced them.
- Detect deleted, modified, reordered, or fabricated evidence after a flow completes.
- Preserve the existing flow, task, subtask, and toolcall model instead of adding a separate audit workflow.
- Keep verification possible outside a running PentAGI instance.

## Non-Goals

- This RFC does not add legal non-repudiation by itself.
- This RFC does not prove that a command was authorized, safe, or in scope.
- This RFC does not require every intermediate LLM token to be signed.
- This RFC does not define a final storage schema or migration.

## Proposed Receipt Chain

Each meaningful operation can emit a signed receipt. Receipts form a hash-linked chain or DAG:

```text
flow_start -> task_created -> subtask_started -> toolcall_finished -> artifact_recorded -> report_created
```

A receipt should include:

- `receipt_id`: stable identifier for the signed receipt.
- `parent_receipt_ids`: prior receipts this event depends on.
- `flow_id`, `task_id`, `subtask_id`, and `toolcall_id` when available.
- `event_type`: for example `toolcall_started`, `toolcall_finished`, `artifact_recorded`, or `report_created`.
- `actor_type`: agent, user, system, or external tool.
- `tool_name` and normalized command metadata when the event is a tool call.
- `input_hash` and `output_hash` for data that should not be embedded directly.
- `artifact_hashes` for files, screenshots, logs, exports, or final reports.
- `created_at` using a server-side timestamp.
- `previous_hash` or a sorted parent hash list.
- `signature` and `public_key_id`.

Example shape:

```json
{
  "receipt_id": "receipt_01h...",
  "parent_receipt_ids": ["receipt_01g..."],
  "flow_id": 42,
  "task_id": 314,
  "subtask_id": 2718,
  "toolcall_id": 1618,
  "event_type": "toolcall_finished",
  "actor_type": "agent",
  "tool_name": "terminal",
  "input_hash": "sha256:...",
  "output_hash": "sha256:...",
  "artifact_hashes": ["sha256:..."],
  "created_at": "2026-04-22T00:00:00Z",
  "parent_hash": "sha256:...",
  "signature": "ed25519:...",
  "public_key_id": "pentagi-instance-2026-04"
}
```

## Signing Model

Ed25519 is a good default for compact signatures and simple verification. The signed payload should be canonicalized before signing so independent verifiers can reproduce the same digest.

Recommended rules:

- Sign canonical JSON or another deterministic encoding.
- Hash large inputs, outputs, and artifacts instead of embedding them in every receipt.
- Include enough metadata to identify the exact flow/task/subtask/toolcall source rows.
- Keep public keys exportable with the evidence package.
- Rotate signing keys by introducing a new `public_key_id`, not by rewriting old receipts.

## Pipeline Integration Points

The first implementation can be incremental:

1. Create a flow-level root receipt when a flow starts or when evidence signing is enabled.
2. Emit toolcall receipts after terminal, browser, search, memory, agent delegation, and result-storage tools finish.
3. Hash artifacts as they are saved to `/work`, object storage, or report export locations.
4. Emit a final report receipt that references the report file hash and the latest receipt parents.
5. Export receipts, public keys, and artifact hashes with the final report bundle.

Tool execution is the highest-value starting point because it binds observed findings to concrete actions without requiring changes to every prompt template.

## Verification Workflow

An external verifier should be able to:

1. Load the receipt bundle and public key material.
2. Recompute each receipt hash from its canonical payload.
3. Verify each Ed25519 signature.
4. Check parent links for missing or reordered receipts.
5. Recompute artifact hashes from exported files.
6. Confirm the final report receipt references the report hash and expected terminal receipts.

The verifier should fail closed when a receipt is missing, a parent link cannot be resolved, a signature is invalid, or an artifact hash does not match.

## Open Questions

- Should receipt signing be always on, opt-in per deployment, or opt-in per flow?
- Where should private keys live: local filesystem, KMS, HSM, Vault, or a user-provided signer?
- Should user approvals and scope changes be signed in the same chain as tool calls?
- How should streaming tool output be chunked and hashed for long-running commands?
- Should receipt bundles use an existing format such as the signed receipt draft referenced in issue #235?
- How much command input should be stored directly versus hashed or redacted?
- Should failed, canceled, and blocked tool calls be signed with the same detail as successful calls?

## Suggested First Milestone

Start with a non-invasive export-only prototype:

- Sign completed toolcall summaries and final report hashes.
- Store receipt payloads as sidecar JSON files in the report export bundle.
- Provide a small verifier command that validates signatures, parent links, and artifact hashes.
- Leave database schema changes and real-time UI visualization for a later milestone after the receipt shape is validated.
