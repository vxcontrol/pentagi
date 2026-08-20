# Deaf Guard — Operator Runbook

Management, configuration, and debugging reference for the Deaf Guard command classification engine.

- Source: `backend/pkg/tools/deafguard/`
- Integration point: `backend/pkg/tools/executor.go`

## 1. What Deaf Guard is

Deaf Guard is a pre-execution command classification layer that sits between the AI agent's tool-call request and the sandboxed Kali container. Every `terminal` tool call is read, classified against a 9-tier rule table (~50 regex rules), and one of three actions is returned: LOG (safe), WARN (aggressive but defensible), or BLOCK (dangerous). The current enforcement mode determines whether a classification just gets recorded, gets returned to the agent as a warning, or hard-blocks execution.

Non-terminal tools (`search`, `browser`, `file`, `memorist`, and others) short-circuit with an allow-all result and are not classified.

## 2. Where Deaf Guard is configured

There are three places Deaf Guard reads configuration from, in the following order of precedence:

1. **Flow snapshot.** When a flow is created, the current Deaf Guard config is captured into the flow's tool context. Runtime changes via the API or UI do not affect already-running flows — only new flows pick up the new config.
2. **REST API / Settings UI.** `GET /api/v1/deafguard/config` and `PUT /api/v1/deafguard/config` expose `{enabled, mode}`. Settings → Security wraps this API.
3. **Environment variables.** Read from `.env` at backend startup. These are the fallback defaults if nothing has been overridden via the API.

## 3. Environment variables

| Variable | Type | Default | Purpose |
|---|---|---|---|
| `DEAF_GUARD_ENABLED` | bool | `true` | Master switch. When `false`, every tool call returns allow-all without classification. |
| `DEAF_GUARD_MODE` | string | `log` | Enforcement posture. Valid values: `log`, `warn`, `enforce`. |
| `DEBUG` | bool | `false` | When `true`, logrus log level drops to Debug, which reveals LOG-tier Deaf Guard classifications (the safe majority). Also affects all other backend logging. |

Defined in `backend/pkg/config/config.go`. Mode validation is re-applied inside `deafguard.New()` — any value outside `log/warn/enforce` silently falls back to `log`.

## 4. Enforcement modes

| Mode | BLOCK-tier commands | WARN-tier commands | LOG-tier commands | Typical use |
|---|---|---|---|---|
| `log` | Allowed, classified | Allowed, classified | Allowed, classified | Baseline observation. Collect classification data from real flows before tightening. |
| `warn` | Blocked; agent receives a warning | Allowed, classified | Allowed, classified | Transitional. Agent can see BLOCK hits and adjust; operator still reviews WARNs out of band. |
| `enforce` | Hard-blocked | Hard-blocked | Allowed, classified | Restrictive posture. Only LOG-tier commands execute. |

Promotion path is **log → warn → enforce**. Do not skip modes; the transition from `log` to `warn` surfaces false positives and tuning gaps that are much harder to diagnose after enforce.

## 5. The 9 tiers at a glance

| Tier | Category | Default action | Examples |
|---|---|---|---|
| 1 | Container Escape | BLOCK | Docker socket, `nsenter`, `chroot`, host `/proc/1/*`, Docker/kubectl CLI |
| 2 | Destructive File Ops | BLOCK | Recursive deletion of system paths, `shred`, writing to block devices, `mkfs` |
| 3 | Network DoS | BLOCK | Flood tools, fork bombs, `stress-ng`, flood ping, nmap DoS scripts |
| 4 | Persistence / Implants | BLOCK | Account creation, SSH key injection, cron, systemd unit paths |
| 5 | Reverse Shells / Exfil | WARN | Reverse-shell patterns, `scp`, `rsync`, file uploads |
| 6 | Credential Abuse | WARN | Brute-force / credential tools — verify target is in scope |
| 7 | Aggressive Pentest Flags | WARN | High-risk sqlmap flags, nmap exploit/brute scripts, very high thread counts |
| 8 | Standard Pentesting | LOG | Default-flag recon and testing tools |
| 9 | Local Utility | LOG | Container-internal utilities and the no-rule-matched fallback |

Full regex set: `backend/pkg/tools/deafguard/rules.go`. Human-readable descriptions live in `CategoryInfo` in the same file and feed the Settings UI.

## 6. The Deaf Guard tab — primary operator surface

Classifications stream live into the **Deaf Guard** tab in the flow detail view, sitting between Agents and Searches. Open any running flow and click the tab — events appear as they fire, with no restart or refresh required.

**What the tab shows:**

- **Counter pills** at the top break down all events by action (`block` / `warn` / `log`) and by tier (T1–T9). These pills count the full event set for the flow, regardless of any filters applied below.
- **Filter bar** supports multi-select tier, multi-select action, and a 500 ms debounced search that matches against both the `command` and `reason` fields. Filters compose (tier AND action AND search); clearing all filters shows everything.
- **Row list** renders newest first. Each row shows timestamp, tier badge, action badge, outcome (allowed/blocked), and a truncated command. Rows are left-bordered by severity band — tiers 1–4 destructive (red), 5–7 warn (amber), 8–9 neutral (grey). Click any row to expand and see the full reason, category, risk, mode, allowed flag, and the full command.
- **Empty state** differentiates between "no events yet" (flow hasn't classified anything) and "no events match your filters" (offers a reset button).

**Scope caveats:**

- Events are **in-memory only** in the browser session. Closing the tab mid-flow and reopening does not replay prior events. Persistence is not implemented; the GraphQL `deafGuardEvents` query always returns an empty list so Apollo can append live `deafGuardEventAdded` events.
- Events are scoped per flow. Switching flows via the sidebar resets the tab state.
- Under extreme classification bursts a slow subscriber may drop events — the backend channel uses non-blocking sends. If you need guaranteed capture for audit or evidence, fall back to the container log approach in §9.

## 7. Debug mode — seeing every classification

By default, only **BLOCK** and **WARN** classifications are visible in `docker compose logs pentagi`. LOG-tier classifications (the safe majority) are written at Debug level and suppressed at the default Info log level. The **Deaf Guard tab** surfaces all classifications regardless of log level — use the tab first; enable DEBUG below only when you need correlated backend log lines.

To see every classification including LOG-tier:

```bash
# Flip DEBUG in .env (add it if the line doesn't exist)
grep -q '^DEBUG=' .env && sed -i '' 's/^DEBUG=.*/DEBUG=true/' .env || echo 'DEBUG=true' >> .env

# Restart the backend — config is read once at startup
docker compose restart pentagi

# Tail classifications live
docker compose logs -f pentagi | grep 'deaf guard:'
```

To revert:

```bash
sed -i '' 's/^DEBUG=.*/DEBUG=false/' .env
docker compose restart pentagi
```

Restarting the `pentagi` container terminates any in-flight flow. Finish or stop the current flow before enabling debug, unless you are intentionally starting fresh.

## 8. Changing enforcement mode at runtime

Three ways, listed in order of safety:

### 8a. Settings → Security UI

1. Open Settings → Security.
2. Toggle Enable Deaf Guard and/or Enforcement Mode.

Changes apply to all **new** flows. Running flows retain their original snapshot.

Requires the `settings.view` privilege.

### 8b. REST API

Successful responses use the standard PentAGI envelope `{ "status": "success", "data": { "enabled": ..., "mode": ... } }`.

```bash
# Inspect current config
curl -sk https://localhost:8443/api/v1/deafguard/config \
  -H "Authorization: Bearer $PENTAGI_API_TOKEN"

# Switch to warn mode
curl -sk -X PUT https://localhost:8443/api/v1/deafguard/config \
  -H "Authorization: Bearer $PENTAGI_API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"mode": "warn"}'
```

API tokens are generated in Settings → API Tokens. Session cookies also work from a logged-in browser session.

### 8c. Environment variable + restart

Edit `.env`, set `DEAF_GUARD_MODE=warn` or `enforce`, then `docker compose restart pentagi`. This also terminates running flows. Use this only for a persistent baseline change — runtime adjustments should go through the UI or API.

## 9. Reading classification logs

For live operator visibility, prefer the Deaf Guard tab. This section covers the container-log fallback for scripted analysis, evidence capture, or review after the browser session closed.

Every classification writes a single structured logrus line with these fields:

| Field | Meaning |
|---|---|
| `component` | Always `deaf_guard` — filter marker for grep/Loki queries |
| `command` | Truncated to 200 runes; original command the agent requested |
| `category` | One of the 9 categories |
| `tier` | 1 (worst) through 9 (safe) |
| `risk` | `none` / `low` / `medium` / `high` / `critical` |
| `action` | `block` / `warn` / `log` — what the classifier decided |
| `allowed` | What actually happened after mode was applied |
| `mode` | `log` / `warn` / `enforce` at the time of classification |
| `reason` | Which rule matched |

### Useful log queries

```bash
# Tail live
docker compose logs -f pentagi | grep 'deaf guard:'

# Pull everything from the current run to a file
LOGFILE=dg-run-$(date +%Y%m%d-%H%M).log
docker compose logs pentagi 2>&1 | grep 'deaf guard:' > "$LOGFILE"

# By-tier distribution
grep 'deaf guard:' "$LOGFILE" | grep -oE 'tier=[0-9]+' | sort | uniq -c | sort -rn

# Only commands that were actually blocked
grep 'deaf guard:' "$LOGFILE" | grep 'allowed=false'
```

## 10. Known limitations

- **Interpreter-nested commands.** Nested interpreter payloads can bypass regex matching of the inner command.
- **Encoding evasion.** Encoding, substitution, escape sequences, and PATH-qualified binaries can slip past pattern matching.
- **`file` tool is unclassified.** The agent can write via the `file` tool without Deaf Guard inspection.
- **No scope awareness.** The classifier does not know which hosts are in-scope for the flow.
- **Per-flow snapshot.** API / UI mode changes do not retroactively apply to running flows.
- **Shell parser is intentionally simple.** `splitCommand` splits on `; && || |` but does not handle heredocs, escaped quotes inside strings, or process substitution.

## 11. Troubleshooting

### No `deaf guard:` lines in logs at all

- Confirm `DEAF_GUARD_ENABLED=true` in `.env` (or in the Settings UI runtime config).
- Confirm the agent is actually calling the `terminal` tool. Non-terminal tools do not emit classifications.
- Confirm the backend is running: `docker compose ps pentagi`.

### Only BLOCK/WARN lines, no LOG lines

Expected behavior at the default log level. Enable `DEBUG=true` to see LOG-tier. See §7.

### A command I expected to block went through

1. Open the Deaf Guard tab and search for the command. The row shows the action/tier and whether mode permitted it.
2. If the row shows `block` action but `allowed` outcome, the mode is `log`. Promote to `warn` or `enforce` per §8.
3. If the command does not appear in the tab or fell through to tier 9, the command evaded regex matching. Check for interpreter-nesting or encoding evasion (§10). Consider adding a rule in `rules.go` and a test in `deafguard_test.go`.

### Changed mode in the UI but existing flow isn't respecting it

Expected. Flow config is snapshotted at flow creation. Create a new flow to pick up the change.

### Debug mode is very noisy

`DEBUG=true` lowers the log level globally, not just for Deaf Guard. Filter with `grep 'deaf guard:'` during analysis. Flip `DEBUG=false` and restart when done.

## 12. Related docs

- [config.md](./config.md) — application configuration reference
- [flow_execution.md](./flow_execution.md) — how flows execute tool calls
