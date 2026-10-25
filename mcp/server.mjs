#!/usr/bin/env node
import { McpServer } from '@modelcontextprotocol/sdk/server/mcp.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import { z } from 'zod';

const base = (process.env.PENTAGI_URL ?? 'https://localhost:8443').replace(/\/+$/, '') + '/api/v1';
const token = process.env.PENTAGI_TOKEN;

if (!token) {
  console.error('PENTAGI_TOKEN is required (create an API token in PentAGI settings)');
  process.exit(1);
}

// The stock deployment serves a self-signed certificate on localhost.
if (process.env.PENTAGI_INSECURE_TLS === '1') {
  process.env.NODE_TLS_REJECT_UNAUTHORIZED = '0';
}

async function api(method, path, { query, body } = {}) {
  const url = new URL(base + path);
  for (const [k, v] of Object.entries(query ?? {})) {
    if (v !== undefined) url.searchParams.set(k, String(v));
  }
  const res = await fetch(url, {
    method,
    headers: {
      Authorization: `Bearer ${token}`,
      ...(body ? { 'Content-Type': 'application/json' } : {}),
    },
    body: body ? JSON.stringify(body) : undefined,
    signal: AbortSignal.timeout(30_000),
  });
  const text = await res.text();
  if (!res.ok) throw new Error(`${method} ${path} -> ${res.status}: ${text.slice(0, 500)}`);
  const json = JSON.parse(text);
  return json.data ?? json;
}

const reply = (data) => ({ content: [{ type: 'text', text: JSON.stringify(data, null, 2) }] });
const wrap = (fn) => async (args) => {
  try {
    return reply(await fn(args));
  } catch (err) {
    return { isError: true, content: [{ type: 'text', text: String(err.message ?? err) }] };
  }
};

const flowID = z.number().int().positive().describe('Flow id');
const page = {
  page: z.number().int().min(1).default(1),
  pageSize: z.number().int().min(1).max(200).default(20),
};
const table = ({ page, pageSize }) => ({ type: 'page', page, pageSize });

const server = new McpServer({ name: 'pentagi', version: '0.1.0' });

server.registerTool('pentagi_list_providers', {
  description: 'List configured LLM providers; pass one of them to pentagi_create_flow.',
}, wrap(() => api('GET', '/providers/')));

server.registerTool('pentagi_create_flow', {
  description:
    'Start an automated security test. Only run against targets the user is authorized to test. ' +
    'The input is the task for the first agent, e.g. "Scan 10.0.0.5 for open services and report findings".',
  inputSchema: { input: z.string().min(1), provider: z.string().min(1) },
}, wrap(({ input, provider }) => api('POST', '/flows/', { body: { input, provider } })));

server.registerTool('pentagi_list_flows', {
  description: 'List flows, newest first.',
  inputSchema: page,
}, wrap((a) => api('GET', '/flows/', { query: table(a) })));

server.registerTool('pentagi_get_flow', {
  description: 'Get a flow with its status.',
  inputSchema: { flowID },
}, wrap(({ flowID }) => api('GET', `/flows/${flowID}`)));

server.registerTool('pentagi_get_flow_tasks', {
  description: 'List the tasks of a flow, with status and result.',
  inputSchema: { flowID },
}, wrap(({ flowID }) => api('GET', `/flows/${flowID}/tasks/`, { query: { type: 'init', page: 1, pageSize: -1 } })));

server.registerTool('pentagi_get_flow_subtasks', {
  description: 'List all subtasks of a flow, with status and result.',
  inputSchema: { flowID },
}, wrap(({ flowID }) => api('GET', `/flows/${flowID}/subtasks/`, { query: { type: 'init', page: 1, pageSize: -1 } })));

server.registerTool('pentagi_get_flow_messages', {
  description: 'Read the message log of a flow: agent output, findings, questions waiting for input.',
  inputSchema: { flowID, ...page },
}, wrap(({ flowID, ...a }) => api('GET', `/flows/${flowID}/msglogs/`, { query: table(a) })));

server.registerTool('pentagi_send_input', {
  description: 'Answer a flow that is waiting for user input.',
  inputSchema: { flowID, input: z.string().min(1) },
}, wrap(({ flowID, input }) => api('PUT', `/flows/${flowID}`, { body: { action: 'input', input } })));

server.registerTool('pentagi_stop_flow', {
  description: 'Stop a running flow. The flow and its results are kept.',
  inputSchema: { flowID },
}, wrap(({ flowID }) => api('PUT', `/flows/${flowID}`, { body: { action: 'stop' } })));

server.registerTool('pentagi_get_flow_report', {
  description: 'Generate the report for a flow.',
  inputSchema: { flowID },
}, wrap(({ flowID }) => api('PUT', `/flows/${flowID}`, { body: { action: 'report' } })));

// --- Headless sandboxes: a Docker container with no agent loop and no LLM
// provider behind it. You drive it directly with these tools instead of
// pentagi_create_flow. Needs no provider configured on the PentAGI server.

const sandboxID = z.number().int().positive().describe('Sandbox id (same id space as flows)');

server.registerTool('pentagi_create_sandbox', {
  description:
    'Provision a headless sandbox: a Docker container with no agent and no LLM behind it. ' +
    'Only run against targets the user is authorized to test. Use pentagi_exec_sandbox to drive it yourself.',
  inputSchema: {
    title: z.string().optional(),
    image: z.string().optional().describe('Defaults to the server\'s configured pentest image (e.g. Kali)'),
  },
}, wrap((a) => api('POST', '/sandboxes/', { body: a })));

server.registerTool('pentagi_list_sandboxes', {
  description: 'List your headless sandboxes (running and finished).',
  inputSchema: {},
}, wrap(async () => {
  const flows = await api('GET', '/flows/', { query: { type: 'init', page: 1, pageSize: -1 } });
  const list = Array.isArray(flows) ? flows : flows.flows ?? [];
  return list.filter((f) => f.headless);
}));

server.registerTool('pentagi_exec_sandbox', {
  description:
    'Run a shell command inside a headless sandbox\'s container and get its output back directly ' +
    '(no agent in the loop). Set detach for long-running/interactive commands (servers, listeners, shells).',
  inputSchema: {
    sandboxID,
    input: z.string().min(1),
    cwd: z.string().optional(),
    detach: z.boolean().optional(),
    timeout: z.number().int().min(0).max(10800).optional().describe('Seconds, 0 = server default'),
  },
}, wrap(({ sandboxID, ...body }) => api('POST', `/sandboxes/${sandboxID}/exec`, { body })));

server.registerTool('pentagi_read_sandbox_file', {
  description: 'Read a file from a headless sandbox\'s container.',
  inputSchema: { sandboxID, path: z.string().min(1) },
}, wrap(({ sandboxID, path }) => api('GET', `/sandboxes/${sandboxID}/file`, { query: { path } })));

server.registerTool('pentagi_write_sandbox_file', {
  description: 'Overwrite a file in a headless sandbox\'s container.',
  inputSchema: { sandboxID, path: z.string().min(1), content: z.string() },
}, wrap(({ sandboxID, ...body }) => api('PUT', `/sandboxes/${sandboxID}/file`, { body })));

server.registerTool('pentagi_stop_sandbox', {
  description: 'Stop a headless sandbox and remove its container. The flow record is kept.',
  inputSchema: { sandboxID },
}, wrap(({ sandboxID }) => api('DELETE', `/sandboxes/${sandboxID}`)));

await server.connect(new StdioServerTransport());
