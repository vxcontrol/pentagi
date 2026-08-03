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

export const mergeCassettes = (base: Cassette, override: Cassette): Cassette => ({
    mutations: { ...base.mutations, ...override.mutations },
    queries: { ...base.queries, ...override.queries },
    rest: { ...base.rest, ...override.rest },
    subscriptions: { ...base.subscriptions, ...override.subscriptions },
});

export interface GraphQLCassetteEntry extends GraphQLPayload, WorldFlagged {
    /**
     * Served verbatim in place of the `{ data, errors }` envelope. The GraphQL endpoint sits behind
     * the same auth middleware as the REST ones, so a rejected request answers with the REST error
     * body and never reaches gqlgen.
     */
    body?: unknown;
    /** A transport-level status the client reads instead of the GraphQL `errors` array. */
    status?: number;
    variables?: Record<string, unknown>;
}

export interface GraphQLPayload {
    data?: unknown;
    errors?: Array<{ message: string }>;
}

export interface RestCassetteEntry extends WorldFlagged {
    body?: unknown;
    /**
     * Subset match against the request's JSON body: nested plain objects are matched
     * recursively, but an array value is compared by full equality, not element-wise.
     * Without it a path hit answers success no matter what the app sent — pin it on every
     * mutation entry whose payload matters (login credentials, create/update bodies).
     */
    bodySubset?: Record<string, unknown>;
    /**
     * Serves `body` verbatim under this content type instead of encoding it as
     * JSON. Binary endpoints need it: a screenshot `<img>` handed JSON decodes to
     * a broken image while the request still counts as matched.
     */
    contentType?: string;
    /**
     * Subset match against the request's query string, the counterpart to `bodySubset` for
     * endpoints that carry their payload there and send no body — a delete addressed by
     * `?paths[]=…` is otherwise answered success whatever it names. Repeated keys are compared
     * as the full list of values.
     */
    querySubset?: Record<string, string | string[]>;
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
