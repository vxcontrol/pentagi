export interface Cassette {
    mutations?: Record<string, GraphQLCassetteEntry[]>;
    queries?: Record<string, GraphQLCassetteEntry[]>;
    /** Keyed as `"<METHOD> <pathname>"`, e.g. `"GET /api/v1/info"`. */
    rest?: Record<string, RestCassetteEntry[]>;
    subscriptions?: Record<string, SubscriptionCassetteEntry[]>;
}

export interface CassetteFrame {
    delayMs?: number;
    payload: GraphQLPayload;
    /** Playback pauses until the flag is raised — chains a mutation to its subscription frame. */
    whenFlag?: string;
}

/**
 * Generated types omit __typename (skipTypename), but Apollo's cache
 * normalization and the subscription auto-merge need it at runtime — this adds
 * it while keeping the compile-time field checking of the declared type.
 */
export const entity = <T extends object>(typename: string, value: T): T => ({ __typename: typename, ...value });

/** Section-wise merge: an override replaces the base entry list per operation/path key. */
export const mergeCassettes = (base: Cassette, override: Cassette): Cassette => ({
    mutations: { ...base.mutations, ...override.mutations },
    queries: { ...base.queries, ...override.queries },
    rest: { ...base.rest, ...override.rest },
    subscriptions: { ...base.subscriptions, ...override.subscriptions },
});

export interface GraphQLCassetteEntry extends GraphQLPayload, WorldFlagged {
    variables?: Record<string, unknown>;
}

export interface GraphQLPayload {
    data?: unknown;
    errors?: Array<{ message: string }>;
}

export interface RestCassetteEntry extends WorldFlagged {
    body?: unknown;
    status?: number;
}

export interface SubscriptionCassetteEntry {
    /**
     * Streams stay open by default: the real server never completes live
     * subscriptions, and completing them closes the lazy graphql-ws client's
     * socket, making the reconnect path untestable. Opt in per stream.
     */
    complete?: boolean;
    frames: CassetteFrame[];
    variables?: Record<string, unknown>;
}

/**
 * Minimal world state: an entry can raise a flag when served (`setFlag`) and be
 * eligible only after one is raised (`whenFlag`). Flagged entries outrank
 * unflagged ones, so `GET /info` can flip guest→user after the login mutation
 * without depending on call order.
 */
export interface WorldFlagged {
    setFlag?: string;
    whenFlag?: string;
}
