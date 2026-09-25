// OpenAI-compatible mock LLM for the Tier-2 e2e stack. The backend talks to it
// through the `custom` provider (LLM_SERVER_URL), so the real langchaingo
// client, streaming, and tool-call plumbing are exercised end-to-end while
// answers stay deterministic: the whole request JSON is matched against
// scenario.mjs rules in order, first hit wins.
import { createServer } from 'node:http';

import { FALLBACK, RULES } from './scenario.mjs';

const PORT = 8080;
// Holds a flow in Running long enough to reach the controls that exist only then (Stop, the
// disabled composer): at the default 0 a whole run settles in about 300 ms.
const DELAY_MS = Number(process.env.MOCK_LLM_DELAY_MS ?? 0);

const readBody = (request) =>
    new Promise((resolve) => {
        let body = '';

        request.on('data', (chunk) => {
            body += chunk;
        });
        request.on('end', () => resolve(body));
    });

const pickAnswer = (payload) => {
    const text = JSON.stringify(payload);

    for (const rule of RULES) {
        if (rule.match.test(text)) {
            return rule;
        }
    }

    return FALLBACK;
};

let completionCounter = 0;

const completionMessage = (rule) => {
    const message = { content: rule.content ?? null, role: 'assistant' };

    if (rule.toolCalls) {
        // Fixed-width ids: the backend samples these to derive its tool-call-ID
        // template, so the shape must stay consistent across calls.
        message.tool_calls = rule.toolCalls.map((call, index) => ({
            function: { arguments: JSON.stringify(call.args ?? {}), name: call.name },
            id: `call_${String(completionCounter * 10 + index).padStart(6, '0')}`,
            type: 'function',
        }));
    }

    return message;
};

const respondJson = (response, status, payload) => {
    response.writeHead(status, { 'content-type': 'application/json' });
    response.end(JSON.stringify(payload));
};

// OpenAI error envelope — langchaingo surfaces `error.message`, a bare string
// under `error` is dropped on the floor.
const respondError = (response, status, message) =>
    respondJson(response, status, { error: { code: null, message, param: null, type: 'invalid_request_error' } });

const respondCompletion = (response, payload, rule) => {
    completionCounter += 1;
    const id = `chatcmpl-e2e-${completionCounter}`;
    const message = completionMessage(rule);
    const finishReason = rule.toolCalls ? 'tool_calls' : 'stop';
    const usage = { completion_tokens: 7, prompt_tokens: 11, total_tokens: 18 };

    if (!payload.stream) {
        respondJson(response, 200, {
            choices: [{ finish_reason: finishReason, index: 0, message }],
            created: 1767225600,
            id,
            model: payload.model ?? 'e2e-mock',
            object: 'chat.completion',
            usage,
        });

        return;
    }

    response.writeHead(200, {
        'cache-control': 'no-cache',
        connection: 'keep-alive',
        'content-type': 'text/event-stream',
    });

    const chunk = (delta, finish = null) =>
        response.write(
            `data: ${JSON.stringify({
                choices: [{ delta, finish_reason: finish, index: 0 }],
                created: 1767225600,
                id,
                model: payload.model ?? 'e2e-mock',
                object: 'chat.completion.chunk',
            })}\n\n`,
        );

    chunk({ role: 'assistant' });

    if (message.tool_calls) {
        chunk({
            tool_calls: message.tool_calls.map((call, index) => ({ index, ...call })),
        });
    } else {
        for (const piece of String(message.content ?? '').match(/.{1,24}/gs) ?? []) {
            chunk({ content: piece });
        }
    }

    chunk({}, finishReason);
    response.write(`data: [DONE]\n\n`);
    response.end();
};

// Same text in, same vector out: a knowledge document's chunks have to land in pgvector with
// stable coordinates, or a semantic search over them answers differently on every run.
const EMBEDDING_DIMENSIONS = 1536;

const embed = (text) => {
    const vector = new Array(EMBEDDING_DIMENSIONS).fill(0);
    let hash = 0;

    for (let index = 0; index < text.length; index += 1) {
        hash = (hash * 31 + text.charCodeAt(index)) % 2147483647;
        vector[index % EMBEDDING_DIMENSIONS] += ((hash % 2000) - 1000) / 1000;
    }

    const length = Math.sqrt(vector.reduce((total, value) => total + value * value, 0)) || 1;

    return vector.map((value) => value / length);
};

createServer(async (request, response) => {
    const { url = '' } = request;

    if (request.method === 'POST' && url.endsWith('/embeddings')) {
        let payload;

        try {
            payload = JSON.parse((await readBody(request)) || '{}');
        } catch (error) {
            respondError(response, 400, `mock-llm: invalid JSON body: ${error}`);

            return;
        }

        const input = Array.isArray(payload?.input) ? payload.input : [payload?.input ?? ''];

        console.log(`[mock-llm] embeddings: ${input.length} input(s)`);
        respondJson(response, 200, {
            data: input.map((text, index) => ({
                embedding: embed(String(text)),
                index,
                object: 'embedding',
            })),
            model: payload?.model ?? 'e2e-mock-embed',
            object: 'list',
            usage: { prompt_tokens: input.length, total_tokens: input.length },
        });

        return;
    }

    if (request.method === 'POST' && url.endsWith('/chat/completions')) {
        // A malformed body must 400, not throw: an uncaught throw in this
        // async callback kills the process, resets completionCounter, and
        // drops every in-flight SSE stream of the run.
        let payload;

        try {
            payload = JSON.parse((await readBody(request)) || '{}');

            // JSON.parse('null')/'42' succeed, but reading payload.tools/.stream below would
            // then throw outside this try and kill the process — reject non-objects here.
            if (typeof payload !== 'object' || payload === null || Array.isArray(payload)) {
                throw new Error('body must be a JSON object');
            }
        } catch (error) {
            respondError(response, 400, `mock-llm: invalid JSON body: ${error}`);

            return;
        }

        const rule = pickAnswer(payload);
        // Defensive end to end because this runs outside the try/catch above, where a throw kills
        // the process: the top-level guard admits {"tools":5} and {"tools":[null]} alike.
        const toolList = Array.isArray(payload.tools) ? payload.tools : [];
        const toolNames = toolList.map((tool) => tool?.function?.name ?? tool?.type ?? '?').join(',');

        console.log(`[mock-llm] ${rule.label}: ${payload.stream ? 'stream' : 'plain'} tools=[${toolNames}]`);

        if (DELAY_MS > 0) {
            await new Promise((resolve) => setTimeout(resolve, DELAY_MS));
        }

        respondCompletion(response, payload, rule);

        return;
    }

    if (request.method === 'GET' && url.endsWith('/models')) {
        respondJson(response, 200, {
            data: [{ id: 'e2e-mock', object: 'model', owned_by: 'e2e' }],
            object: 'list',
        });

        return;
    }

    respondError(response, 404, `mock-llm: unhandled ${request.method} ${url}`);
}).listen(PORT, () => console.log(`[mock-llm] listening on :${PORT}`));
