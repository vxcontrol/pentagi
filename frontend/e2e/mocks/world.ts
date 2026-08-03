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

const isPlainObject = (value: unknown): value is Record<string, unknown> => {
    if (typeof value !== 'object' || value === null || Array.isArray(value)) {
        return false;
    }

    // Only true plain objects recurse in isSubsetMatch. A Date/Map/Set/class instance has no own
    // enumerable entries, so `every` over [] would be vacuously true and the pin would match any object.
    const proto = Object.getPrototypeOf(value) as unknown;

    return proto === Object.prototype || proto === null;
};

const isSubsetMatch = (expected?: Record<string, unknown>, actual?: Record<string, unknown>): boolean =>
    !expected ||
    Object.entries(expected).every(([key, value]) => {
        const found = actual?.[key];

        return isPlainObject(value) && isPlainObject(found)
            ? isSubsetMatch(value, found)
            : stableStringify(found) === stableStringify(value);
    });

const sleep = (delayMs: number) => new Promise<void>((resolve) => setTimeout(resolve, delayMs));

/** Repeated keys (`paths[]=a&paths[]=b`) collapse to the full list, so a subset match sees both. */
const queryRecord = (search?: URLSearchParams): Record<string, string | string[]> =>
    Object.fromEntries(
        [...(search?.keys() ?? [])]
            .map((key) => [key, search!.getAll(key)])
            .map(([key, values]) => [key, (values as string[]).length > 1 ? values : (values as string[])[0]]),
    );

export const subscriptionStreamKey = (operationName: string, variables?: Record<string, unknown>): string =>
    `sub:${operationName}:${stableStringify(variables ?? {})}`;

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

    private readonly consumed = new Map<string, Set<number>>();
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
                // Same cursor rule as matchRest below: key on the selected entries' pins, so an
                // operation whose variables carry a per-call volatile field still advances its
                // sequence instead of re-serving step one.
                const signature = matching.map((entry) => stableStringify(entry.variables ?? null)).join('|');

                return this.nextEntry(`gql:${operationName}:${signature}`, candidates, matching);
            }
        }

        return undefined;
    }

    matchRest(
        method: string,
        pathname: string,
        body?: Record<string, unknown>,
        search?: URLSearchParams,
    ): RestCassetteEntry | undefined {
        const normalized = pathname.replace(/\/+$/, '');
        const key = Object.keys(this.cassette.rest ?? {}).find((candidate) => {
            const [candidateMethod, candidatePath = ''] = candidate.split(' ');

            return candidateMethod === method && candidatePath.replace(/\/+$/, '') === normalized;
        });
        const query = queryRecord(search);
        const matching = (key ? this.cassette.rest?.[key] : undefined)?.filter(
            (entry) => isSubsetMatch(entry.bodySubset, body) && isSubsetMatch(entry.querySubset, query),
        );
        const candidates = this.eligible(matching);

        // Key the cursor on which entries the request selected — their pins — not the raw body/query.
        // Two payload-differentiated sequences still get distinct keys (their entries pin different
        // subsets), but a request carrying a per-call volatile field (an idempotency key, a timestamp)
        // no longer fragments its own cursor into a fresh sequence on every call.
        const signature = matching
            ?.map((entry) => stableStringify([entry.bodySubset ?? null, entry.querySubset ?? null]))
            .join('|');

        return matching && candidates?.length
            ? this.nextEntry(`rest:${method}:${normalized}:${signature}`, candidates, matching)
            : undefined;
    }

    matchSubscription(
        operationName: string,
        variables?: Record<string, unknown>,
    ): undefined | { entry: SubscriptionCassetteEntry; streamKey: string } {
        const entry = this.cassette.subscriptions?.[operationName]?.find((candidate) =>
            isSubsetMatch(candidate.variables, variables),
        );

        // Key on the request's variables, not the entry's: an entry written without variables would
        // otherwise merge every flow's subscribers onto one stream, and the second would join delta-only.
        return entry ? { entry, streamKey: subscriptionStreamKey(operationName, variables) } : undefined;
    }

    raiseFlag(flag: string): void {
        this.flags.add(flag);
        this.flagWaiters.get(flag)?.forEach((resolve) => resolve());
        this.flagWaiters.delete(flag);
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

    subscriberCount(streamKey: string): number {
        return this.streams.get(streamKey)?.subscribers.size ?? 0;
    }

    subscribeStream(streamKey: string, entry: SubscriptionCassetteEntry, subscriber: StreamSubscriber): () => void {
        const state = this.streams.get(streamKey) ?? { cursor: 0, isDriving: false, subscribers: new Set() };

        this.streams.set(streamKey, state);
        state.subscribers.add(subscriber);

        if (!state.isDriving) {
            state.isDriving = true;
            void this.driveStream(state, entry.frames, entry.complete ?? false);
        } else if (state.cursor >= entry.frames.length && (entry.complete ?? false)) {
            // Late subscriber to an already-exhausted stream: no frame replay (the no-replay-after-
            // reconnect contract), but a complete:true stream must still signal completion.
            subscriber.complete();
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
     * Sequenced responses: entries for one match key are served in order and the last repeats.
     * Consumption is tracked by entry identity (its index in the unfiltered `all` list), not by a
     * cursor into the current eligible subset — so a raised flag that grows the subset serves the
     * newly-enabled entry next, instead of restarting at 0 and replaying an already-served one.
     */
    private nextEntry<T extends WorldFlagged>(key: string, candidates: T[], all: T[]): T {
        const served = this.consumed.get(key) ?? new Set<number>();

        this.consumed.set(key, served);

        const entry = candidates.find((candidate) => !served.has(all.indexOf(candidate))) ?? candidates.at(-1)!;

        served.add(all.indexOf(entry));

        if (entry.setFlag) {
            this.raiseFlag(entry.setFlag);
        }

        return entry;
    }
}
