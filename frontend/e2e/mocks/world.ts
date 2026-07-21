import type {
    Cassette,
    CassetteFrame,
    GraphQLCassetteEntry,
    GraphQLPayload,
    RestCassetteEntry,
    SubscriptionCassetteEntry,
    WorldFlagged,
} from './cassette.ts';

const stableStringify = (value: unknown): string =>
    JSON.stringify(value, (_key, node: unknown) =>
        node && typeof node === 'object' && !Array.isArray(node)
            ? Object.fromEntries(Object.entries(node).sort(([a], [b]) => a.localeCompare(b)))
            : node,
    ) ?? 'undefined';

const isPlainObject = (value: unknown): value is Record<string, unknown> =>
    typeof value === 'object' && value !== null && !Array.isArray(value);

const isSubsetMatch = (expected?: Record<string, unknown>, actual?: Record<string, unknown>): boolean =>
    !expected ||
    Object.entries(expected).every(([key, value]) => {
        const found = actual?.[key];

        return isPlainObject(value) && isPlainObject(found)
            ? isSubsetMatch(value, found)
            : stableStringify(found) === stableStringify(value);
    });

const sleep = (delayMs: number) => new Promise<void>((resolve) => setTimeout(resolve, delayMs));

export interface MockSocket {
    close: (options?: { code?: number; reason?: string }) => unknown;
}

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
    /**
     * Unmatched subscriptions are expected (a flow page opens ~15 and cassettes
     * mock only the relevant ones), so unlike `unmatched` this is not a teardown
     * gate — it is attached as a diagnostic when a test fails, so a typo'd
     * subscription key points at the cassette instead of a UI timeout.
     */
    readonly unmatchedSubscriptions: string[] = [];

    private readonly consumed = new Map<string, { index: number; signature: string }>();
    private readonly flags = new Set<string>();
    private readonly flagWaiters = new Map<string, Array<() => void>>();
    private readonly sockets = new Set<MockSocket>();
    private readonly streams = new Map<string, StreamState>();

    constructor(private readonly cassette: Cassette) {}

    dropSockets(code = 1001, reason = 'e2e: forced drop'): void {
        // 1001 is retryable for graphql-ws; 4400+/1011 would kill the client for good.
        for (const socket of this.sockets) {
            socket.close({ code, reason });
        }
    }

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
            const matching = section?.[operationName]?.filter((entry) => isSubsetMatch(entry.variables, variables));
            const candidates = this.eligible(matching);

            if (matching && candidates?.length) {
                return this.nextEntry(`gql:${operationName}:${stableStringify(variables ?? {})}`, candidates, matching);
            }
        }

        return undefined;
    }

    matchRest(method: string, pathname: string, body?: Record<string, unknown>): RestCassetteEntry | undefined {
        const normalized = pathname.replace(/\/+$/, '');
        const key = Object.keys(this.cassette.rest ?? {}).find((candidate) => {
            const [candidateMethod, candidatePath = ''] = candidate.split(' ');

            return candidateMethod === method && candidatePath.replace(/\/+$/, '') === normalized;
        });
        const matching = (key ? this.cassette.rest?.[key] : undefined)?.filter((entry) =>
            isSubsetMatch(entry.bodySubset, body),
        );
        const candidates = this.eligible(matching);

        return matching && candidates?.length
            ? this.nextEntry(`rest:${method}:${normalized}`, candidates, matching)
            : undefined;
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

    registerSocket(socket: MockSocket): void {
        this.sockets.add(socket);
    }

    reportUnmatched(description: string): void {
        this.unmatched.push(description);
    }

    reportUnmatchedSubscription(description: string): void {
        this.unmatchedSubscriptions.push(description);
    }

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

    unregisterSocket(socket: MockSocket): void {
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

    private eligible<T extends WorldFlagged>(candidates?: T[]): T[] | undefined {
        const open = candidates?.filter((entry) => !entry.whenFlag || this.flags.has(entry.whenFlag));

        return open?.some((entry) => entry.whenFlag) ? open.filter((entry) => entry.whenFlag) : open;
    }

    /**
     * Sequenced responses: entries for the same match key are consumed in
     * order; the last one repeats. The cursor is scoped to the current eligible
     * subset — when a raised flag changes which entries are visible, carrying a
     * pre-flag cursor over would skip the new subset's leading entries.
     */
    private nextEntry<T extends WorldFlagged>(key: string, candidates: T[], all: T[]): T {
        const signature = candidates.map((candidate) => all.indexOf(candidate)).join(',');
        const state = this.consumed.get(key);
        const index = state?.signature === signature ? state.index : 0;

        this.consumed.set(key, { index: index + 1, signature });

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
