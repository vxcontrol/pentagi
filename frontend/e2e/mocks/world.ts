import type { WebSocketRoute } from '@playwright/test';

import type {
    Cassette,
    CassetteFrame,
    GraphQLCassetteEntry,
    GraphQLPayload,
    RestCassetteEntry,
    SubscriptionCassetteEntry,
    WorldFlagged,
} from './cassette.ts';

const isSubsetMatch = (expected?: Record<string, unknown>, actual?: Record<string, unknown>): boolean =>
    !expected ||
    Object.entries(expected).every(([key, value]) => JSON.stringify(actual?.[key]) === JSON.stringify(value));

const sleep = (delayMs: number) => new Promise<void>((resolve) => setTimeout(resolve, delayMs));

export interface StreamSubscriber {
    complete: () => void;
    next: (payload: GraphQLPayload) => void;
}

interface StreamState {
    cursor: number;
    isDriving: boolean;
    subscribers: Set<StreamSubscriber>;
}

export class MockWorld {
    readonly unmatched: string[] = [];

    private readonly consumed = new Map<string, number>();
    private readonly flags = new Set<string>();
    private readonly flagWaiters = new Map<string, Array<() => void>>();
    private readonly sockets = new Set<WebSocketRoute>();
    private readonly streams = new Map<string, StreamState>();

    constructor(private readonly cassette: Cassette) {}

    dropSockets(code = 1001, reason = 'e2e: forced drop'): void {
        // 1001 is retryable for graphql-ws; 4400+/1011 would kill the client for good.
        for (const socket of this.sockets) {
            socket.close({ code, reason });
        }
    }

    /** Resolves once the flag is raised (immediately if it already is). */
    flagRaised(flag: string): Promise<void> {
        if (this.flags.has(flag)) {
            return Promise.resolve();
        }

        return new Promise((resolve) => {
            const waiters = this.flagWaiters.get(flag) ?? [];

            waiters.push(resolve);
            this.flagWaiters.set(flag, waiters);
        });
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

    /**
     * One driver per stream broadcasts each frame once to every current
     * subscriber: duplicate subscribers (e.g. flowUpdated is opened by both the
     * list and the detail provider) each get the frame, while a re-subscribe
     * after a reconnect joins past the cursor — delta-only, like the real
     * server. Returns an unsubscribe function.
     */
    subscribeStream(streamKey: string, entry: SubscriptionCassetteEntry, subscriber: StreamSubscriber): () => void {
        const state = this.streams.get(streamKey) ?? { cursor: 0, isDriving: false, subscribers: new Set() };

        this.streams.set(streamKey, state);
        state.subscribers.add(subscriber);

        if (!state.isDriving) {
            state.isDriving = true;
            void this.driveStream(state, entry.frames, entry.complete ?? false);
        }

        return () => state.subscribers.delete(subscriber);
    }

    unregisterSocket(socket: WebSocketRoute): void {
        this.sockets.delete(socket);
    }

    private async driveStream(state: StreamState, frames: CassetteFrame[], shouldComplete: boolean): Promise<void> {
        while (state.cursor < frames.length) {
            const frame = frames[state.cursor] as CassetteFrame;

            if (frame.whenFlag) {
                await this.flagRaised(frame.whenFlag);
            }

            if (frame.delayMs) {
                await sleep(frame.delayMs);
            }

            state.cursor += 1;
            state.subscribers.forEach(({ next }) => next(frame.payload));
        }

        if (shouldComplete) {
            state.subscribers.forEach(({ complete }) => complete());
        }
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
            this.raiseFlag(entry.setFlag);
        }

        return entry;
    }

    private raiseFlag(flag: string): void {
        this.flags.add(flag);
        this.flagWaiters.get(flag)?.forEach((resolve) => resolve());
        this.flagWaiters.delete(flag);
    }
}
