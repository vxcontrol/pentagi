import type { DeafGuardEventFragmentFragment } from '@/graphql/types';

// ---------- Domain constants ----------

// Actions surfaced by the Deaf Guard classifier. String values match the Go
// backend's Action enum in pkg/tools/deafguard/rules.go.
export const DEAF_GUARD_ACTIONS = ['block', 'warn', 'log'] as const;
export type DeafGuardAction = (typeof DEAF_GUARD_ACTIONS)[number];

// Tiers 1-9 from the classification taxonomy. Lower numbers = higher severity.
// Tier 1 is container-escape territory; tier 9 is benign local-utility.
export const DEAF_GUARD_TIERS = [1, 2, 3, 4, 5, 6, 7, 8, 9] as const;
export type DeafGuardTier = (typeof DEAF_GUARD_TIERS)[number];

// ---------- Severity banding ----------

/**
 * Maps a tier to a severity band for row coloring.
 *
 * Mapping comes directly from SEC-4744: tiers 1-4 surface as destructive
 * (container-escape through persistence — things that should never run),
 * 5-7 as warn (reverse-shell/credential-abuse/aggressive-flags — scope-
 * dependent), 8-9 as neutral (standard pentest + local utilities).
 */
export type SeverityBand = 'destructive' | 'neutral' | 'warn';

export const getTierBand = (tier: number): SeverityBand => {
    if (tier >= 1 && tier <= 4) {
        return 'destructive';
    }

    if (tier >= 5 && tier <= 7) {
        return 'warn';
    }

    return 'neutral';
};

// ---------- Filtering ----------

export interface DeafGuardFilters {
    // Empty set = no action filter applied (show all actions).
    actions: ReadonlySet<string>;
    // Debounced lowercase search term, matched against command + reason.
    search: string;
    // Empty set = no tier filter applied (show all tiers).
    tiers: ReadonlySet<number>;
}

export const hasActiveFilters = (filters: DeafGuardFilters): boolean =>
    filters.tiers.size > 0 || filters.actions.size > 0 || filters.search.trim().length > 0;

/**
 * Applies tier, action, and search filters together. Empty filter sets are
 * no-ops so "clear all filters shows everything" is the default.
 * Case-insensitive substring match across command + reason.
 */
export const filterEvents = (
    events: readonly DeafGuardEventFragmentFragment[],
    filters: DeafGuardFilters,
): DeafGuardEventFragmentFragment[] => {
    const search = filters.search.trim().toLowerCase();
    const hasSearch = search.length > 0;
    const hasTierFilter = filters.tiers.size > 0;
    const hasActionFilter = filters.actions.size > 0;

    if (!hasSearch && !hasTierFilter && !hasActionFilter) {
        return [...events];
    }

    return events.filter((event) => {
        if (hasTierFilter && !filters.tiers.has(event.tier)) {
            return false;
        }

        if (hasActionFilter && !filters.actions.has(event.action)) {
            return false;
        }

        if (hasSearch) {
            const command = event.command.toLowerCase();
            const reason = event.reason.toLowerCase();

            if (!command.includes(search) && !reason.includes(search)) {
                return false;
            }
        }

        return true;
    });
};

// ---------- Counters ----------

/**
 * Counts events by action. Always returns a shape with every known action
 * present, so downstream code can render pill badges without null checks.
 */
export const countByAction = (events: readonly DeafGuardEventFragmentFragment[]): Record<DeafGuardAction, number> => {
    const counts: Record<DeafGuardAction, number> = { block: 0, log: 0, warn: 0 };

    for (const event of events) {
        const action = event.action as DeafGuardAction;

        if (action in counts) {
            counts[action] += 1;
        }
    }

    return counts;
};

/**
 * Counts events by tier 1-9. Tiers with zero events are still represented
 * with a count of 0 so the tier pill strip renders consistently.
 */
export const countByTier = (events: readonly DeafGuardEventFragmentFragment[]): Record<number, number> => {
    const counts: Record<number, number> = {};

    for (const tier of DEAF_GUARD_TIERS) {
        counts[tier] = 0;
    }

    for (const event of events) {
        const current = counts[event.tier];
        if (current !== undefined) {
            counts[event.tier] = current + 1;
        }
    }

    return counts;
};

// ---------- Sorting ----------

/**
 * Returns events newest first (highest timestamp at index 0). When two events
 * share a timestamp — possible when a flow emits many classifications in the
 * same second — the higher ID wins (monotonic per process, so ID is a correct
 * tiebreaker for arrival order).
 */
export const sortNewestFirst = (
    events: readonly DeafGuardEventFragmentFragment[],
): DeafGuardEventFragmentFragment[] => {
    return [...events].sort((a, b) => {
        if (b.timestamp !== a.timestamp) {
            return b.timestamp - a.timestamp;
        }

        // Treat ids as BigInt-safe via Number — IDs come from a process-local
        // atomic counter so they stay well under 2^53 in any reasonable runtime.
        return Number(b.id) - Number(a.id);
    });
};
