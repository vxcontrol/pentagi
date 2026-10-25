# PentAGI MCP server

Stdio MCP server that lets Claude Code drive a PentAGI instance: headless Docker sandboxes (no LLM provider needed on the server — you drive the container), and the agent-driven flow API (needs a provider configured).

## Setup

```bash
cd mcp && npm install
```

Create an API token (`POST /api/v1/tokens/` with a session cookie, or Settings → Tokens in the web UI), then register the server:

```bash
claude mcp add pentagi -s user \
  -e PENTAGI_URL=https://localhost:8443 \
  -e PENTAGI_TOKEN=<token> \
  -e PENTAGI_INSECURE_TLS=1 \
  -- node /path/to/pentagi/mcp/server.mjs
```

Drop `PENTAGI_INSECURE_TLS` if the server has a real certificate, and use `http://` if it's running without TLS (`SERVER_USE_SSL` unset/no cert, e.g. a local dev build).

Install the skill: `mkdir -p ~/.claude/skills/pentagi && cp skill/SKILL.md ~/.claude/skills/pentagi/`.

## Tools

**Headless sandboxes** — `pentagi_create_sandbox`, `pentagi_list_sandboxes`, `pentagi_exec_sandbox`, `pentagi_read_sandbox_file`, `pentagi_write_sandbox_file`, `pentagi_stop_sandbox`. Backed by the `backend/pkg/server/services/sandboxes.go` REST endpoints added for this: a flow row with `headless = true` and a real Docker container, no agent loop, no LLM provider. You are the one deciding what command to run next.

**Agent-driven flows** — `pentagi_list_providers`, `pentagi_create_flow`, `pentagi_list_flows`, `pentagi_get_flow`, `pentagi_get_flow_tasks`, `pentagi_get_flow_subtasks`, `pentagi_get_flow_messages`, `pentagi_send_input`, `pentagi_stop_flow`, `pentagi_get_flow_report`. These need a provider configured on the server (`OLLAMA_SERVER_URL` for a free local model, or any provider key you already have) — PentAGI's own agents plan and run the engagement.
