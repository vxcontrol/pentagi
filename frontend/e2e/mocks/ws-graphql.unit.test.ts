import type { Client } from 'graphql-ws';

import { createClient } from 'graphql-ws';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { WebSocket, WebSocketServer } from 'ws';

import type { Cassette } from './cassette.ts';

import { MockWorld } from './world.ts';
import { handleWsConnection } from './ws-graphql.ts';

/**
 * Protocol-contract suite: a real graphql-ws client (configured like the app's
 * in src/lib/apollo.ts — infinite retries, shouldRetry always) driven against
 * a real ws server backed by the exact `handleWsConnection` the Playwright
 * mock uses. Violations of these invariants otherwise only surface as
 * reconnect storms inside flaky browser specs.
 */

interface Harness {
    client: Client;
    world: MockWorld;
}

const disposers: Array<() => Promise<void> | void> = [];

afterEach(async () => {
    while (disposers.length) {
        await disposers.pop()?.();
    }
});

const start = async (cassette: Cassette, onConnected?: (wasRetry: boolean) => void): Promise<Harness> => {
    const world = new MockWorld(cassette);
    const server = new WebSocketServer({
        handleProtocols: (protocols) => (protocols.has('graphql-transport-ws') ? 'graphql-transport-ws' : false),
        port: 0,
    });

    server.on('connection', (socket) => {
        const connection = handleWsConnection(world, {
            close: ({ code = 1000, reason = '' } = {}) => socket.close(code, reason),
            send: (data) => socket.send(data),
        });

        socket.on('message', (raw) => connection.onMessage(String(raw)));
        socket.on('close', () => connection.onClose());
    });

    await new Promise<void>((resolve) => server.on('listening', resolve));

    const address = server.address();
    const port = typeof address === 'object' && address ? address.port : 0;

    const client = createClient({
        lazy: true,
        on: { connected: (_socket, _payload, wasRetry) => onConnected?.(wasRetry) },
        retryAttempts: Infinity,
        retryWait: () => new Promise((resolve) => setTimeout(resolve, 10)),
        shouldRetry: () => true,
        url: `ws://127.0.0.1:${port}/api/v1/graphql`,
        webSocketImpl: WebSocket,
    });

    disposers.push(async () => {
        await Promise.resolve(client.dispose()).catch(() => undefined);
        await new Promise<void>((resolve) => server.close(() => resolve()));
    });

    return { client, world };
};

const collect = (client: Client, operationName: string, variables?: Record<string, unknown>) => {
    const received: unknown[] = [];
    const errors: unknown[] = [];
    let isComplete = false;

    const unsubscribe = client.subscribe(
        { operationName, query: `subscription ${operationName} { stub }`, variables },
        {
            complete: () => {
                isComplete = true;
            },
            error: (error) => errors.push(error),
            next: (value) => received.push(value),
        },
    );

    return { errors, isComplete: () => isComplete, received, unsubscribe };
};

describe('graphql-ws mock protocol contract', () => {
    it('acks the handshake and delivers a matched subscription frame', async () => {
        const { client } = await start({
            subscriptions: { demo: [{ frames: [{ payload: { data: { demo: 1 } } }] }] },
        });
        const sink = collect(client, 'demo');

        await vi.waitFor(() => expect(sink.received).toEqual([{ data: { demo: 1 } }]));
        expect(sink.errors).toEqual([]);
    });

    it('answers an unmatched subscription with recorded silence — no error, no complete', async () => {
        const { client, world } = await start({});
        const sink = collect(client, 'unknownOp', { id: '9' });

        await vi.waitFor(() => expect(world.unmatchedSubscriptions).toEqual(['subscription unknownOp {"id":"9"}']));
        await new Promise((resolve) => setTimeout(resolve, 50));
        expect(sink.received).toEqual([]);
        expect(sink.errors).toEqual([]);
        expect(sink.isComplete()).toBe(false);
    });

    it('completes the sink only for streams that opt into complete', async () => {
        const { client } = await start({
            subscriptions: {
                finite: [{ complete: true, frames: [{ payload: { data: { finite: 1 } } }] }],
                open: [{ frames: [{ payload: { data: { open: 1 } } }] }],
            },
        });
        const finite = collect(client, 'finite');
        const open = collect(client, 'open');

        await vi.waitFor(() => expect(finite.isComplete()).toBe(true));
        expect(finite.received).toEqual([{ data: { finite: 1 } }]);

        // A stream without `complete` delivers its frames but must stay open — the real server
        // never completes live subscriptions, so flipping the default to complete would fail here.
        await vi.waitFor(() => expect(open.received).toEqual([{ data: { open: 1 } }]));
        await new Promise((resolve) => setTimeout(resolve, 50));
        expect(open.isComplete()).toBe(false);
    });

    it('recovers from a 1001 drop and never replays already-delivered frames', async () => {
        const retries: boolean[] = [];
        const { client, world } = await start(
            {
                queries: { raise: [{ data: {}, setFlag: 'later' }] },
                subscriptions: {
                    stream: [
                        {
                            frames: [
                                { payload: { data: { seq: 1 } } },
                                { payload: { data: { seq: 2 } }, whenFlag: 'later' },
                            ],
                        },
                    ],
                },
            },
            (wasRetry) => retries.push(wasRetry),
        );
        const sink = collect(client, 'stream');

        await vi.waitFor(() => expect(sink.received).toEqual([{ data: { seq: 1 } }]));

        world.dropSockets();

        // graphql-ws resubscribes active sinks in a microtask AFTER the retry connect, so gating on
        // the reconnect alone races the resubscribe: raising the flag before the sink re-registers
        // delivers seq:2 to an empty subscriber set and the no-replay contract loses it. Gate on the
        // server-side resubscription instead.
        await vi.waitFor(() => expect(retries).toEqual([false, true]));
        await vi.waitFor(() => expect(world.subscriberCount('sub:stream:{}')).toBe(1));

        world.matchGraphQL('raise');

        await vi.waitFor(() => expect(sink.received).toEqual([{ data: { seq: 1 } }, { data: { seq: 2 } }]));
        expect(sink.errors).toEqual([]);
    });
});
