import { describe, expect, it } from 'vitest';

import type { DeafGuardEventFragmentFragment } from '@/graphql/types';

import { countByAction, countByTier, filterEvents, getTierBand, hasActiveFilters, sortNewestFirst } from './lib';

// Factory for building fake events in tests. Keeps each case terse by
// letting you override only the fields the assertion cares about.
const makeEvent = (overrides: Partial<DeafGuardEventFragmentFragment> = {}): DeafGuardEventFragmentFragment => ({
    action: 'log',
    allowed: true,
    category: 'local_utility',
    command: 'ls -la',
    flowId: '42',
    id: '1',
    mode: 'log',
    reason: 'No matching rules — allowed',
    risk: 'none',
    tier: 9,
    timestamp: 1700000000,
    ...overrides,
});

describe('getTierBand', () => {
    it('maps tiers 1-4 to destructive', () => {
        expect(getTierBand(1)).toBe('destructive');
        expect(getTierBand(2)).toBe('destructive');
        expect(getTierBand(3)).toBe('destructive');
        expect(getTierBand(4)).toBe('destructive');
    });

    it('maps tiers 5-7 to warn', () => {
        expect(getTierBand(5)).toBe('warn');
        expect(getTierBand(6)).toBe('warn');
        expect(getTierBand(7)).toBe('warn');
    });

    it('maps tiers 8-9 to neutral', () => {
        expect(getTierBand(8)).toBe('neutral');
        expect(getTierBand(9)).toBe('neutral');
    });

    it('treats out-of-range tiers as neutral', () => {
        expect(getTierBand(0)).toBe('neutral');
        expect(getTierBand(10)).toBe('neutral');
    });
});

describe('hasActiveFilters', () => {
    it('returns false when all filters empty', () => {
        expect(hasActiveFilters({ actions: new Set(), search: '', tiers: new Set() })).toBe(false);
    });

    it('ignores pure-whitespace search', () => {
        expect(hasActiveFilters({ actions: new Set(), search: '   ', tiers: new Set() })).toBe(false);
    });

    it('detects active search', () => {
        expect(hasActiveFilters({ actions: new Set(), search: 'rm', tiers: new Set() })).toBe(true);
    });

    it('detects active tier filter', () => {
        expect(hasActiveFilters({ actions: new Set(), search: '', tiers: new Set([1]) })).toBe(true);
    });

    it('detects active action filter', () => {
        expect(hasActiveFilters({ actions: new Set(['block']), search: '', tiers: new Set() })).toBe(true);
    });
});

describe('filterEvents', () => {
    const events: DeafGuardEventFragmentFragment[] = [
        makeEvent({ action: 'block', command: 'rm -rf /', id: '1', reason: 'Destructive file op', tier: 2 }),
        makeEvent({ action: 'warn', command: 'hydra -l admin', id: '2', reason: 'Brute force tool', tier: 6 }),
        makeEvent({ action: 'log', command: 'ls -la', id: '3', reason: 'No matching rules', tier: 9 }),
        makeEvent({ action: 'block', command: 'mount /dev/sda1 /mnt', id: '4', reason: 'Container escape', tier: 1 }),
    ];

    it('returns everything when all filters are empty', () => {
        const result = filterEvents(events, { actions: new Set(), search: '', tiers: new Set() });
        expect(result).toHaveLength(events.length);
    });

    it('filters by tier (single)', () => {
        const result = filterEvents(events, { actions: new Set(), search: '', tiers: new Set([1]) });
        expect(result.map((e) => e.id)).toEqual(['4']);
    });

    it('filters by tier (multiple — OR)', () => {
        const result = filterEvents(events, { actions: new Set(), search: '', tiers: new Set([1, 2]) });
        expect(result.map((e) => e.id).sort()).toEqual(['1', '4']);
    });

    it('filters by action', () => {
        const result = filterEvents(events, { actions: new Set(['block']), search: '', tiers: new Set() });
        expect(result.map((e) => e.id).sort()).toEqual(['1', '4']);
    });

    it('filters by case-insensitive search across command', () => {
        const result = filterEvents(events, { actions: new Set(), search: 'HYDRA', tiers: new Set() });
        expect(result.map((e) => e.id)).toEqual(['2']);
    });

    it('filters by case-insensitive search across reason', () => {
        const result = filterEvents(events, { actions: new Set(), search: 'container', tiers: new Set() });
        expect(result.map((e) => e.id)).toEqual(['4']);
    });

    it('applies tier + action + search together (AND across filter types)', () => {
        const result = filterEvents(events, {
            actions: new Set(['block']),
            search: 'rm',
            tiers: new Set([1, 2, 3]),
        });
        expect(result.map((e) => e.id)).toEqual(['1']);
    });

    it('returns empty when filters have no overlap', () => {
        const result = filterEvents(events, {
            actions: new Set(['log']),
            search: '',
            tiers: new Set([1]),
        });
        expect(result).toEqual([]);
    });

    it('clearing all filters shows everything (AC requirement)', () => {
        // Simulate "reset" — an empty filter set returns the full list.
        const result = filterEvents(events, { actions: new Set(), search: '', tiers: new Set() });
        expect(result).toHaveLength(events.length);
    });
});

describe('countByAction', () => {
    it('returns zero counts when no events', () => {
        expect(countByAction([])).toEqual({ block: 0, log: 0, warn: 0 });
    });

    it('counts each action independently', () => {
        const events = [
            makeEvent({ action: 'block', id: '1' }),
            makeEvent({ action: 'block', id: '2' }),
            makeEvent({ action: 'warn', id: '3' }),
            makeEvent({ action: 'log', id: '4' }),
            makeEvent({ action: 'log', id: '5' }),
            makeEvent({ action: 'log', id: '6' }),
        ];
        expect(countByAction(events)).toEqual({ block: 2, log: 3, warn: 1 });
    });

    it('ignores unknown action values gracefully', () => {
        const events = [makeEvent({ action: 'mystery' as never, id: '1' }), makeEvent({ action: 'block', id: '2' })];
        expect(countByAction(events)).toEqual({ block: 1, log: 0, warn: 0 });
    });
});

describe('countByTier', () => {
    it('always returns all 9 tiers, even when empty', () => {
        const result = countByTier([]);
        expect(
            Object.keys(result)
                .map(Number)
                .sort((a, b) => a - b),
        ).toEqual([1, 2, 3, 4, 5, 6, 7, 8, 9]);
        expect(Object.values(result).every((c) => c === 0)).toBe(true);
    });

    it('counts each tier independently', () => {
        const events = [
            makeEvent({ id: '1', tier: 1 }),
            makeEvent({ id: '2', tier: 2 }),
            makeEvent({ id: '3', tier: 2 }),
            makeEvent({ id: '4', tier: 9 }),
        ];
        const result = countByTier(events);
        expect(result[1]).toBe(1);
        expect(result[2]).toBe(2);
        expect(result[9]).toBe(1);
        expect(result[5]).toBe(0);
    });

    it('ignores out-of-range tier values (defensive)', () => {
        const events = [makeEvent({ id: '1', tier: 0 }), makeEvent({ id: '2', tier: 1 })];
        expect(countByTier(events)[1]).toBe(1);
    });
});

describe('sortNewestFirst', () => {
    it('sorts by timestamp descending', () => {
        const events = [
            makeEvent({ id: '1', timestamp: 1000 }),
            makeEvent({ id: '2', timestamp: 3000 }),
            makeEvent({ id: '3', timestamp: 2000 }),
        ];
        expect(sortNewestFirst(events).map((e) => e.id)).toEqual(['2', '3', '1']);
    });

    it('breaks timestamp ties by id descending (arrival order)', () => {
        const events = [
            makeEvent({ id: '1', timestamp: 1000 }),
            makeEvent({ id: '2', timestamp: 1000 }),
            makeEvent({ id: '3', timestamp: 1000 }),
        ];
        expect(sortNewestFirst(events).map((e) => e.id)).toEqual(['3', '2', '1']);
    });

    it('returns a new array (does not mutate input)', () => {
        const input = [makeEvent({ id: '1', timestamp: 1000 }), makeEvent({ id: '2', timestamp: 2000 })];
        const snapshot = input.map((e) => e.id);
        sortNewestFirst(input);
        expect(input.map((e) => e.id)).toEqual(snapshot);
    });
});
