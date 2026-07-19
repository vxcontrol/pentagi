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
}

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
