---
name: pentagi
description: Run security testing commands and automated pentest flows on the user's PentAGI instance through the pentagi MCP tools. Use when the user asks to scan, pentest, exec commands in, or assess a target they are authorized to test, or to check a PentAGI flow or sandbox.
---

# PentAGI

Drive the user's PentAGI server with the `pentagi_*` MCP tools. Confirm the user is authorized to test the target before doing anything; if not stated, ask once.

## Headless sandbox (default — no LLM provider needed on the PentAGI server)

You are the agent; PentAGI just gives you a Docker container (Kali by default) to run commands in.

1. `pentagi_create_sandbox` (optionally `image`, e.g. `vxcontrol/kali-linux` or `debian:latest`).
2. `pentagi_exec_sandbox` to run commands yourself, `pentagi_read_sandbox_file` / `pentagi_write_sandbox_file` for files. Use `detach: true` for long-running/interactive commands (listeners, servers, shells).
3. `pentagi_list_sandboxes` to see what's open. `pentagi_stop_sandbox` when done — it removes the container but keeps the record.

A sandbox is just a container: nothing plans or re-plans for you. You decide the next command from the previous one's output.

## Agent-driven flow (needs an LLM provider configured on the server)

PentAGI's own agents (planner, pentester, coder, ...) run the whole engagement.

1. `pentagi_list_providers`, then `pentagi_create_flow` with a concrete task and a provider that is configured.
2. Poll `pentagi_get_flow` until status leaves `running`; read `pentagi_get_flow_messages` and `pentagi_get_flow_subtasks` for progress.
3. If the flow waits for input, answer with `pentagi_send_input`.
4. Summarize findings; `pentagi_get_flow_report` produces the full report. `pentagi_stop_flow` stops a run; nothing here deletes data.
