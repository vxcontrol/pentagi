# RFC: Configurable Tool And Model Fallbacks

Issue [#341](https://github.com/vxcontrol/pentagi/issues/341) proposes a configurable fallback layer for repeated tool failures, missing CLI tools, and model or tool-call failures. This RFC sketches a safe design direction before any runtime behavior is implemented.

This document does not add Go or JavaScript code, dependencies, database migrations, GraphQL schema, generated files, Docker changes, installer behavior, or new environment variables. It is a design proposal for maintainers to review before the project commits to a fallback implementation.

PentAGI already has several recovery mechanisms: retries, tool-call argument fixing, repeated-call detection, reflector recovery, execution monitoring, planner support, Installer Agent concepts, and configurable tool-call limits. The proposed fallback layer should build on those mechanisms without hiding extra automation behind the flow lifecycle.

## Goals

- Define a conservative fallback design for failed tools, missing CLI commands, and model or tool-call failures.
- Keep fallback disabled or policy-limited by default.
- Bound every fallback path so a bad tool, bad model, or bad install suggestion cannot create an infinite loop.
- Prefer already configured and already available tools before considering installation.
- Require explicit operator approval before installing tools by default.
- Make each fallback decision visible in flow output, logs, and future audit surfaces.
- Preserve PentAGI's existing flow, task, subtask, agent, and tool-call model.
- Provide incremental milestones that can be implemented and reviewed in small PRs.

## Non-Goals

- This RFC does not implement runtime fallback behavior.
- This RFC does not add a fallback manager, background retry buffer, hidden queue, or implicit lifecycle state.
- This RFC does not add database migrations, GraphQL schema, REST API, generated code, UI, Docker, installer, or `.env.example` changes.
- This RFC does not propose automatic installation by default.
- This RFC does not allow random `curl | sh`, untrusted package repositories, or prompt-discovered install commands.
- This RFC does not silently switch model providers or send a flow's context to a different provider without traceability and policy approval.
- This RFC does not change the behavior of existing recovery paths such as reflector, mentor, planner, repeated-call detection, or chain restoration.

## Current Recovery Behavior

PentAGI already has multiple recovery layers that should remain the first line of defense:

- Agent and provider calls retry before returning failure.
- Tool execution retries and can invoke the tool-call argument fixer when arguments are invalid or rejected.
- Repeated tool-call detection warns the agent to try another tool and eventually aborts repeated loops.
- Reflector recovery handles unstructured responses or repeated failure to produce valid tool calls.
- Caller-reflector recovery gives the original agent guidance after model/tool-call generation retries are exhausted.
- Execution monitoring can invoke the Adviser as a mentor when tool usage appears stuck or inefficient.
- Specialist agents such as Pentester, Coder, and Installer can receive planner-generated steps before execution.
- Configurable tool-call limits cap general and limited agents so runaway chains terminate.
- Chain AST repair can add fallback responses for pending tool calls and sanitize invalid tool-call arguments so interrupted chains remain replayable.

Those behaviors are recovery and consistency mechanisms. A future fallback layer should not replace them. It should run only after the relevant existing recovery path has failed or reached its configured limit, and it should explain why it is trying a fallback.

## Problem Cases

Fallback is useful when the current recovery layers cannot complete the same step safely:

- A search provider fails repeatedly, but another configured search tool can answer the same question.
- A CLI command fails with a missing executable error such as `command not found` or `executable file not found`.
- A requested CLI tool is not installed, but an already installed alternative can satisfy the subtask.
- A model repeatedly emits invalid tool calls even after normal retries and reflector guidance.
- A local model is available but unreliable for tool calling, while an operator has configured an approved fallback provider for that agent type.
- An installation suggestion exists, but the source, command, package name, or scope is not trusted enough to run automatically.

## Proposed v1 Design

A future v1 implementation should treat fallback as an explicit policy decision that happens inside the existing agent execution path, not as a separate background worker.

The top-level fallback policy should use one mode field, not overlapping `enabled` and `mode` switches:

- `disabled`: do not evaluate fallback candidates; keep the existing failure path.
- `observe`: classify the failure and log the fallback candidate that would have been considered, but do not execute fallback actions, install tools, or switch models.
- `enforce`: execute configured fallback actions within policy, approval, and attempt limits.

The fallback decision should include:

- the failed action, tool, command, or model;
- the failure category, such as tool unavailable, tool execution failed, missing CLI, repeated tool call, or invalid tool-call generation;
- the existing recovery steps already attempted;
- the candidate fallback action;
- the policy rule that allowed or denied the fallback;
- the attempt number and remaining fallback budget.

If fallback is disabled, the agent should receive the existing failure behavior. If fallback is enabled but no safe candidate exists, the system should fail visibly with a clear reason instead of inventing a path.

The design should avoid the hidden lifecycle risk learned from previous automation queue work. No fallback attempt should sit in an invisible in-memory queue, silently wait for later execution, or blur when a task or flow is actually complete. If a fallback decision ever needs to survive a turn, restart, or approval wait, that state must be first-class and inspectable before implementation.

## Tool Fallback

Tool fallback should try another already available tool only when the alternative has a compatible capability and is allowed for the current agent and flow.

Examples:

- If `tavily` fails, the fallback candidates might be `duckduckgo`, `google`, `searxng`, `perplexity`, or `browser`, depending on what is configured and available.
- If `sploitus` fails, the fallback could be a general search tool plus a visible note that the result is not from the same exploit-specific source.
- If `browser` fails because a target is unreachable, switching to a search API may be safe for public information gathering but should not be treated as equivalent to live target validation.

Tool fallback rules:

- Prefer native, configured, already available tools before any install path.
- Respect the current agent's tool access. Fallback must not grant a tool that the agent could not otherwise use.
- Preserve target scope. Fallback must not expand a pentest target or move from passive to active behavior without policy approval.
- Surface the selected fallback tool and reason to the agent and operator.
- Record the original failure and fallback result separately so reports do not hide the failed primary tool.
- Do not chain indefinitely across similar tools. A failed fallback consumes budget and may end the path.

## Missing CLI Tool Handling

Missing CLI handling should distinguish between "the requested command is absent" and "installation is safe." Those are separate decisions.

When a terminal command fails because a binary is missing, a future implementation should:

1. Classify the error as a missing CLI only when the signal is clear.
2. Check whether an existing native tool or installed CLI can satisfy the same goal.
3. Ask the Installer Agent or guide memory for an install plan only if existing tools are insufficient.
4. Present the operator with the package, source, exact command, expected effect, and risk before installation.
5. Run installation only after explicit approval unless a future trusted auto-install policy has been deliberately enabled.
6. Log the approval decision, install command, result, and retry outcome.

Safe install policy should reject:

- commands copied from arbitrary web pages;
- `curl | sh`, `wget | sh`, and similar pipe-to-shell patterns by default;
- untrusted package repositories;
- unsigned or checksum-less binaries when a signed/checksummed source is available;
- install commands that modify host or container state outside the allowed worker environment;
- package names or commands invented by the model without a trusted source.

The deny-pattern labels used in the configuration sketch are canonical policy keys. For example, `curl_pipe_shell` means a command that downloads content with `curl` and pipes it directly into `sh`, `bash`, or another shell; `wget_pipe_shell` is the same pattern with `wget`; `model_generated_unverified_command` means the command came from model output without a trusted package source or verification step.

Trusted install sources, if supported later, should be allowlisted by operators. Examples might include distro packages, pinned official release artifacts with checksum verification, or explicit language package coordinates. The exact allowlist format is a later implementation question, not part of this docs-only PR.

## Model Fallback

Model fallback should be more conservative than tool fallback because it can change cost, latency, data residency, and provider trust boundaries.

A future implementation should only attempt model fallback when:

- the primary model repeatedly fails to produce usable tool calls after normal retries and reflector guidance;
- the operator has configured an allowed fallback model or provider for the same agent type;
- the fallback provider is allowed to receive the current flow context;
- the switch is visible in logs and flow output before or when it happens;
- the fallback attempt is bounded and does not recursively trigger more model switches.

Model fallback must not silently change provider behavior. Operators should be able to answer which provider handled each agent step, why the switch happened, what failure triggered it, and whether any provider-specific reasoning or tool-call format had to be normalized.

If a fallback provider is cloud-hosted and the primary model is local, policy should treat that as a material trust-boundary change. The default should be to deny or require explicit approval unless the operator has already configured that fallback path.

## Operator Approval and Policy

Fallback policy should be explicit and easy to disable.

Recommended defaults:

- global fallback `mode: disabled`, with `mode: observe` available for logging-only rollout;
- tool fallback `mode: disabled` until configured, or limited to already configured tools in `observe` before `enforce`;
- install fallback `mode: disabled` by default;
- installation requires operator approval by default;
- model fallback `mode: disabled` until an ordered fallback list is configured;
- no trusted auto-install unless the operator creates a narrow allowlist.

Approval records should include the requested action, approving actor, timestamp, flow/task/subtask, package or provider involved, exact command when applicable, and the policy rule that allowed it.

If `ASK_USER` or another human-in-the-loop surface is unavailable, install fallback should fail closed. The system should not reinterpret a missing approval path as permission to install.

## Visibility and Auditability

Every fallback decision should be visible without reading raw source code.

At minimum, a future implementation should emit a clear flow-visible message such as:

```text
Fallback considered: tavily failed after normal retries. Trying configured search fallback duckduckgo. Attempt 1 of 2.
```

For each fallback decision, the audit surface should record:

- flow, task, subtask, agent, and message chain;
- original tool, command, or model;
- failure category and short error summary;
- existing recovery steps already attempted;
- selected fallback candidate or reason no candidate was allowed;
- approval requirement and approval result;
- attempt count and remaining budget;
- final fallback result.

Fallback output should not overwrite the original tool result. The final report should be able to state that a primary tool failed and an approved fallback was used.

## Configuration Sketch

The following is illustrative pseudocode only. It is not a proposed final environment-variable schema, config file, database shape, GraphQL type, or installer setting.

```yaml
fallback:
  # One of: disabled, observe, enforce.
  # disabled: existing behavior only.
  # observe: classify and log fallback candidates, but do not execute them.
  # enforce: execute configured fallback actions within policy and bounds.
  mode: disabled
  max_total_attempts_per_step: 2

  tool_fallback:
    mode: disabled
    prefer_existing_tools: true
    max_attempts: 2
    allowed_capabilities:
      search:
        - duckduckgo
        - google
        - searxng
        - perplexity
        - browser

  missing_cli:
    mode: disabled
    prefer_existing_tools: true
    ask_installer_for_plan: true

  install_fallback:
    mode: disabled
    require_operator_approval: true
    allow_auto_install: false
    trusted_sources_only: true
    denied_patterns:
      - curl_pipe_shell
      - wget_pipe_shell
      - untrusted_repository
      - model_generated_unverified_command

  model_fallback:
    mode: disabled
    require_policy_match: true
    max_attempts: 1
    fallback_order:
      primary_agent:
        - configured-openai-tool-model
        - configured-anthropic-tool-model

  audit:
    log_reason: true
    log_selected_alternative: true
    log_approval_decision: true
    log_install_commands: true
```

The exact storage and naming should be chosen in an implementation PR. The important v1 constraints are conservative defaults, explicit policy, bounded attempts, and visible decisions.

## Failure Bounds

A fallback implementation should have hard stop conditions:

- maximum fallback attempts per failed step;
- maximum install attempts per missing CLI;
- maximum model fallback attempts per agent call;
- no recursive fallback from fallback-generated failures unless explicitly budgeted;
- no retrying the same fallback candidate against the same failure signature;
- no hidden queue or delayed background retry;
- terminal error once the fallback budget is exhausted.

Fallback should also inherit existing timeouts, output-size limits, agent tool-call limits, and cancellation behavior. A fallback path is not a way around those bounds.

## Security Considerations

Fallback changes can expand what an autonomous agent can do, so the safe default should be denial.

Key risks:

- **Supply chain risk.** Installing tools can run untrusted code. Only trusted, allowlisted sources should be considered, and operator approval should be required by default.
- **Prompt-injected install instructions.** Search results, web pages, and tool output are untrusted. The agent should not turn arbitrary text into an install command.
- **Scope expansion.** A fallback tool must not move from passive lookup to active scanning or expand target scope without explicit policy approval.
- **Provider trust boundary.** Model fallback can send context to a different provider. That decision must be configured and visible.
- **Secret exposure.** Fallback logs should redact credentials, tokens, headers, and sensitive target material.
- **Audit gaps.** If a fallback hides the original failure, operators cannot debug or report accurately. Original failure and fallback result must both remain visible.
- **Resource exhaustion.** Repeated fallback attempts can amplify cost and runtime. Attempts, timeouts, output size, and install retries must be bounded.

## Open Questions

- Where should fallback policy live: static config, database settings, admin UI, per-flow options, or a combination?
- How should PentAGI define capability equivalence between tools without hard-coding too much policy into prompts?
- Should tool fallback be configured globally, per user, per agent type, or per flow?
- What is the approval surface when a flow is running but `ASK_USER` is disabled?
- Which install sources are trusted enough for a first implementation?
- Should trusted auto-install exist at all, or should v1 always require approval?
- How should model fallback handle provider-specific reasoning signatures and tool-call ID formats?
- How should fallback attempts be represented in UI/API/DB if they need to survive restarts or approval waits?
- Should fallback events become first-class report or evidence artifacts?

## Incremental Milestones

1. **Land this RFC.** Agree on the safety model: conservative defaults, visible decisions, bounded attempts, explicit approval, and no hidden queues.
2. **Add fallback event vocabulary.** Define how fallback reasons and decisions are logged, without changing behavior.
3. **Implement existing-tool fallback only.** Allow a configured tool fallback map for already available tools, with visible flow output and strict attempt caps.
4. **Add missing CLI classification.** Detect clear missing-command failures and surface safe alternatives or an install plan, but do not install automatically.
5. **Add approval-gated install fallback.** Execute only allowlisted, trusted install plans after explicit operator approval.
6. **Add model fallback.** Retry with a configured fallback provider only after current retries and reflector recovery fail, with clear provider traceability and strict bounds.

This staging keeps the first implementation small and reviewable. It also prevents the fallback feature from becoming an invisible automation layer that is harder to audit than the original failure.

## Recommendation and Requested Decision

For v1, this RFC recommends the smallest safe slice (milestones 1-3) with these defaults, chosen to match the safety model above:

- **Scope:** existing-tool fallback only, via an operator-defined fallback map. No prompt-inferred capability equivalence in v1.
- **Policy location:** static configuration with an optional per-flow override; no admin UI in v1.
- **Auto-install:** disabled. Missing-CLI handling classifies the failure and surfaces an install plan but never installs automatically; trusted auto-install is deferred to a later, approval-gated milestone.
- **Approval when `ASK_USER` is disabled:** do not fall back to install or any unapproved action; surface the original failure instead of acting without consent.
- **Model fallback:** deferred to a later milestone and triggered only after existing retries and reflector recovery have failed.
- **Visibility:** every fallback decision and the original failure stay visible in flow output and audit, with bounded attempts, timeouts, and output size.

Decision requested from the maintainer: approve this v1 scope and safety model (milestones 1-3 with the defaults above), request changes, or decline. Approval unblocks an implementation PR limited to existing-tool fallback with strict bounds.
