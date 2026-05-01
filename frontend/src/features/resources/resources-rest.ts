import type { UserResourceFragmentFragment } from '@/graphql/types';

/**
 * Canonicalisation utilities for `UserResource` entries flowing through the
 * Apollo cache.
 *
 * SOURCE OF TRUTH — DB:
 *   `user_resources.id` is `BIGINT` (int64). The Go layer (GORM/REST and
 *   gqlgen) preserves this as `int64` / `uint64`.
 *
 * WIRE FORMAT (uniform across both transports in this project):
 *   Both REST JSON (`models.ResourceEntry`) and GraphQL subscription payloads
 *   (gqlgen `ID!` is bound to `graphql.Int64` in `gqlgen.yml`) emit IDs as
 *   JSON **numbers**. The frontend `UserResourceFragmentFragment.id: string`
 *   typing is an unfortunate codegen default for the `ID` scalar — it lies
 *   about the runtime shape until the codegen mapping is fixed (Phase C).
 *
 * CANONICAL CACHE SHAPE:
 *   To keep Apollo's normalised store internally consistent with what the
 *   GraphQL subscription auto-writer emits, REST hydration MUST insert the
 *   IDs in the same numeric form. Otherwise array-field dedup that compares
 *   `id` strictly (`===`) would treat `"28"` (REST-coerced) and `28`
 *   (subscription) as different items and append a duplicate ref to the
 *   `resources` query field.
 *
 * Downstream consumers that compare cache `id` against form-state strings
 * (e.g. `FlowForm.resourceIds: z.array(z.string())`) MUST coerce defensively
 * via `String(resource.id)` at the boundary. These local coercions disappear
 * once the codegen mapping is updated to `id: number`.
 */

/** Wire shape of `models.ResourceEntry` (REST JSON, snake_case). */
export interface RestResourceEntry {
    created_at: string;
    id: number;
    is_dir: boolean;
    name: string;
    path: string;
    size: number;
    updated_at: string;
    user_id: number;
}

/** Wire shape of `models.ResourceList` (REST JSON). */
export interface RestResourceList {
    items?: null | RestResourceEntry[];
    total?: number;
}

/**
 * Convert a REST `RestResourceEntry` into the canonical camelCase shape that
 * Apollo stores under `UserResource:<id>`. Numeric `id` / `user_id` are kept
 * as `number` (matching gqlgen subscription wire), with a `as unknown as
 * string` cast to satisfy the codegen `UserResourceFragmentFragment.id:
 * string` contract until that mapping is corrected upstream.
 */
export const restResourceEntryToFragment = (rest: RestResourceEntry): UserResourceFragmentFragment => ({
    createdAt: rest.created_at,
    id: rest.id as unknown as string,
    isDir: rest.is_dir,
    name: rest.name,
    path: rest.path,
    size: rest.size,
    updatedAt: rest.updated_at,
    userId: rest.user_id as unknown as string,
});

/**
 * Convert a single canonical resource ID (as it lives in the cache) into the
 * REST numeric wire format. Accepts both the codegen-typed `string` and the
 * actual runtime `number` that the cache stores, so callers can pass values
 * read straight from `UserResource.id` without first stringifying them.
 *
 * Throws on any value that is not a positive `Number.isSafeInteger`. For
 * string inputs an additional round-trip check (`String(parsed) === id`)
 * rejects leading zeros, whitespace, signs and other malformed encodings —
 * such input would signal a contract violation upstream, not user error.
 */
export const resourceIdToWire = (id: number | string): number => {
    if (typeof id === 'number') {
        if (!Number.isSafeInteger(id) || id <= 0) {
            throw new Error(`Invalid resource ID for REST wire: ${JSON.stringify(id)}`);
        }

        return id;
    }

    const parsed = Number(id);

    if (!Number.isSafeInteger(parsed) || parsed <= 0 || String(parsed) !== id) {
        throw new Error(`Invalid resource ID for REST wire: ${JSON.stringify(id)}`);
    }

    return parsed;
};

/** Bulk variant of {@link resourceIdToWire}. Throws on the first invalid ID. */
export const resourceIdsToWire = (ids: ReadonlyArray<number | string>): number[] =>
    ids.map((id) => resourceIdToWire(id));
