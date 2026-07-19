import type { WebSocketRoute } from '@playwright/test';

import type {
    Cassette,
    GraphQLCassetteEntry,
    RestCassetteEntry,
    SubscriptionCassetteEntry,
    WorldFlagged,
} from './cassette.ts';

const isSubsetMatch = (expected?: Record<string, unknown>, actual?: Record<string, unknown>): boolean =>
    !expected ||
    Object.entries(expected).every(([key, value]) => JSON.stringify(actual?.[key]) === JSON.stringify(value));

export class MockWorld {
    readonly unmatched: string[] = [];

    private readonly consumed = new Map<string, number>();
    private readonly flags = new Set<string>();
    private readonly frameCursors = new Map<string, number>();
    private readonly sockets = new Set<WebSocketRoute>();

    constructor(private readonly cassette: Cassette) {}

    advanceFrameCursor(streamKey: string): void {
        this.frameCursors.set(streamKey, this.frameCursor(streamKey) + 1);
    }

    dropSockets(code = 1001, reason = 'e2e: forced drop'): void {
        // 1001 is retryable for graphql-ws; 4400+/1011 would kill the client for good.
        for (const socket of this.sockets) {
            socket.close({ code, reason });
        }
    }

    /** Frames already delivered for a stream survive reconnects — replay is delta-only, matching the real server. */
    frameCursor(streamKey: string): number {
        return this.frameCursors.get(streamKey) ?? 0;
    }

    matchGraphQL(operationName: string, variables?: Record<string, unknown>): GraphQLCassetteEntry | undefined {
        const sections = [this.cassette.queries, this.cassette.mutations];

        for (const section of sections) {
            const candidates = this.eligible(
                section?.[operationName]?.filter((entry) => isSubsetMatch(entry.variables, variables)),
            );

            if (candidates?.length) {
                return this.nextEntry(`gql:${operationName}:${JSON.stringify(variables ?? {})}`, candidates);
            }
        }

        return undefined;
    }

    matchRest(method: string, pathname: string): RestCassetteEntry | undefined {
        const normalized = pathname.replace(/\/+$/, '');
        const key = Object.keys(this.cassette.rest ?? {}).find((candidate) => {
            const [candidateMethod, candidatePath = ''] = candidate.split(' ');

            return candidateMethod === method && candidatePath.replace(/\/+$/, '') === normalized;
        });
        const candidates = this.eligible(key ? this.cassette.rest?.[key] : undefined);

        return candidates?.length ? this.nextEntry(`rest:${method}:${normalized}`, candidates) : undefined;
    }

    matchSubscription(
        operationName: string,
        variables?: Record<string, unknown>,
    ): undefined | { entry: SubscriptionCassetteEntry; streamKey: string } {
        const entry = this.cassette.subscriptions?.[operationName]?.find((candidate) =>
            isSubsetMatch(candidate.variables, variables),
        );

        return entry
            ? { entry, streamKey: `sub:${operationName}:${JSON.stringify(entry.variables ?? {})}` }
            : undefined;
    }

    registerSocket(socket: WebSocketRoute): void {
        this.sockets.add(socket);
    }

    reportUnmatched(description: string): void {
        this.unmatched.push(description);
    }

    unregisterSocket(socket: WebSocketRoute): void {
        this.sockets.delete(socket);
    }

    /** Entries gated on an unraised flag are invisible; raised-flag entries outrank unflagged ones. */
    private eligible<T extends WorldFlagged>(candidates?: T[]): T[] | undefined {
        const open = candidates?.filter((entry) => !entry.whenFlag || this.flags.has(entry.whenFlag));

        return open?.some((entry) => entry.whenFlag) ? open.filter((entry) => entry.whenFlag) : open;
    }

    /** Sequenced responses: entries for the same match key are consumed in order; the last one repeats. */
    private nextEntry<T extends WorldFlagged>(key: string, candidates: T[]): T {
        const index = this.consumed.get(key) ?? 0;

        this.consumed.set(key, index + 1);

        const entry = candidates[Math.min(index, candidates.length - 1)] as T;

        if (entry.setFlag) {
            this.flags.add(entry.setFlag);
        }

        return entry;
    }
}
