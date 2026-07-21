import type { Page } from '@playwright/test';

import type { MockWorld } from './world.ts';

export interface WsTransport {
    close: (options?: { code?: number; reason?: string }) => unknown;
    send: (data: string) => void;
}

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

export const handleWsConnection = (
    world: MockWorld,
    transport: WsTransport,
): { onClose: () => void; onMessage: (raw: string) => void } => {
    world.registerSocket(transport);

    const unsubscribes = new Map<string, () => void>();
    let isOpen = true;

    const send = (message: Record<string, unknown>) => {
        if (isOpen) {
            transport.send(JSON.stringify(message));
        }
    };

    return {
        onClose: () => {
            isOpen = false;
            unsubscribes.forEach((unsubscribe) => unsubscribe());
            unsubscribes.clear();
            world.unregisterSocket(transport);
        },
        onMessage: (raw) => {
            const message = JSON.parse(raw) as ClientMessage;

            switch (message.type) {
                case 'complete': {
                    unsubscribes.get(message.id)?.();
                    unsubscribes.delete(message.id);
                    break;
                }

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
                        world.reportUnmatchedSubscription(
                            `subscription ${payload.operationName} ${JSON.stringify(payload.variables ?? {})}`,
                        );
                        break;
                    }

                    unsubscribes.set(
                        id,
                        world.subscribeStream(match.streamKey, match.entry, {
                            complete: () => send({ id, type: 'complete' }),
                            next: (framePayload) => send({ id, payload: framePayload, type: 'next' }),
                        }),
                    );
                    break;
                }

                default: {
                    break;
                }
            }
        },
    };
};

export const installWsMock = async (page: Page, world: MockWorld): Promise<void> => {
    await page.routeWebSocket('**/api/v1/graphql', (ws) => {
        const connection = handleWsConnection(world, ws);

        ws.onMessage((raw) => connection.onMessage(String(raw)));
        ws.onClose(() => connection.onClose());
    });
};
