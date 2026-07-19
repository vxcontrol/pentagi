import type { Page } from '@playwright/test';

import type { MockWorld } from './world.ts';

type ClientMessage =
    | { id: string; payload: SubscribePayload; type: 'subscribe' }
    | { id: string; type: 'complete' }
    | { payload?: unknown; type: 'connection_init' }
    | { type: 'ping' }
    | { type: 'pong' };

interface SubscribePayload {
    operationName: string;
    query: string;
    variables?: Record<string, unknown>;
}

export const installWsMock = async (page: Page, world: MockWorld): Promise<void> => {
    await page.routeWebSocket('**/api/v1/graphql', (ws) => {
        world.registerSocket(ws);

        const timers = new Set<ReturnType<typeof setTimeout>>();
        let isOpen = true;

        const send = (message: Record<string, unknown>) => {
            if (isOpen) {
                ws.send(JSON.stringify(message));
            }
        };

        const schedule = (delayMs: number, action: () => void) => {
            const timer = setTimeout(() => {
                timers.delete(timer);
                action();
            }, delayMs);

            timers.add(timer);
        };

        ws.onMessage((raw) => {
            const message = JSON.parse(String(raw)) as ClientMessage;

            switch (message.type) {
                case 'connection_init': {
                    // Ack immediately and never send any frame before it: a pre-ack frame
                    // throws inside graphql-ws, and this app's shouldRetry:()=>true turns
                    // that into an infinite reconnect storm.
                    send({ type: 'connection_ack' });
                    break;
                }

                case 'ping': {
                    send({ type: 'pong' });
                    break;
                }

                case 'subscribe': {
                    const { id, payload } = message;
                    const match = world.matchSubscription(payload.operationName, payload.variables);

                    if (!match) {
                        // Unmatched subscriptions get ack'd silence and stay open: a flow page
                        // opens ~15 at once, and `error`/`complete` would poison the client.
                        break;
                    }

                    const { entry, streamKey } = match;
                    const pending = entry.frames.slice(world.frameCursor(streamKey));
                    let elapsed = 0;

                    for (const frame of pending) {
                        elapsed += frame.delayMs ?? 0;
                        schedule(elapsed, () => {
                            world.advanceFrameCursor(streamKey);
                            send({ id, payload: frame.payload, type: 'next' });
                        });
                    }

                    if (entry.complete) {
                        schedule(elapsed, () => send({ id, type: 'complete' }));
                    }

                    break;
                }

                default: {
                    break;
                }
            }
        });

        ws.onClose(() => {
            isOpen = false;

            for (const timer of timers) {
                clearTimeout(timer);
            }

            timers.clear();
            world.unregisterSocket(ws);
        });
    });
};
