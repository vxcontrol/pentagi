import { describe, expect, it, vi } from 'vitest';

import type { SubscriptionCassetteEntry } from './cassette.ts';
import type { StreamSubscriber } from './world.ts';

import { MockWorld } from './world.ts';

const collect = (into: unknown[]): StreamSubscriber => ({ complete: () => {}, next: (payload) => into.push(payload) });

describe('MockWorld matching', () => {
    it('matches variables regardless of property order inside arrays', () => {
        const world = new MockWorld({
            mutations: { save: [{ data: { ok: true }, variables: { items: [{ a: 1, b: 2 }] } }] },
        });

        // Built key-by-key: a reversed object literal would be re-sorted by the linter,
        // hiding that stableStringify must normalise key order inside array values.
        const item: Record<string, number> = {};
        item.b = 2;
        item.a = 1;

        expect(world.matchGraphQL('save', { items: [item] })?.data).toEqual({ ok: true });
    });

    it('treats a pinned object as a subset — extra siblings match, a wrong pinned value does not', () => {
        const world = new MockWorld({
            mutations: { create: [{ data: { ok: true }, variables: { input: { name: 'x' } } }] },
        });

        expect(world.matchGraphQL('create', { input: { name: 'x', ttl: 99 } })?.data).toEqual({ ok: true });
        expect(world.matchGraphQL('create', { input: { name: 'y', ttl: 99 } })).toBeUndefined();
    });

    it('serves same-key entries in order and repeats the last', () => {
        const world = new MockWorld({
            queries: {
                flow: [
                    { data: { n: 1 }, variables: { id: '5' } },
                    { data: { n: 2 }, variables: { id: '5' } },
                ],
            },
        });

        expect(world.matchGraphQL('flow', { id: '5' })?.data).toEqual({ n: 1 });
        expect(world.matchGraphQL('flow', { id: '5' })?.data).toEqual({ n: 2 });
        expect(world.matchGraphQL('flow', { id: '5' })?.data).toEqual({ n: 2 });
    });

    it('leaves a REST path hit with a mismatched body unmatched instead of answering success', () => {
        const world = new MockWorld({
            rest: { 'POST /api/v1/resources/mkdir': [{ body: { ok: true }, bodySubset: { path: 'new-folder' } }] },
        });

        expect(world.matchRest('POST', '/api/v1/resources/mkdir', { path: 'new-folder' })).toBeDefined();
        expect(world.matchRest('POST', '/api/v1/resources/mkdir', { path: 'wrong' })).toBeUndefined();
        expect(world.matchRest('POST', '/api/v1/resources/mkdir/', { path: 'new-folder' })).toBeDefined();
    });

    it('hides a flag-gated entry until the flag is raised, then lets it outrank the unflagged one', () => {
        const world = new MockWorld({
            mutations: { login: [{ data: { ok: true }, setFlag: 'logged-in' }] },
            queries: { info: [{ data: { role: 'guest' } }, { data: { role: 'user' }, whenFlag: 'logged-in' }] },
        });

        expect(world.matchGraphQL('info', {})?.data).toEqual({ role: 'guest' });
        world.matchGraphQL('login', {});
        expect(world.matchGraphQL('info', {})?.data).toEqual({ role: 'user' });
    });
});

describe('MockWorld streams', () => {
    it('broadcasts each frame once to every current subscriber; a later subscriber joins delta-only', async () => {
        const world = new MockWorld({ mutations: { trigger: [{ data: {}, setFlag: 'go' }] } });
        const entry: SubscriptionCassetteEntry = {
            frames: [{ payload: { data: { n: 1 } } }, { payload: { data: { n: 2 } }, whenFlag: 'go' }],
        };
        const first: unknown[] = [];
        const late: unknown[] = [];

        world.subscribeStream('k', entry, collect(first));
        await vi.waitFor(() => expect(first).toHaveLength(1));

        world.subscribeStream('k', entry, collect(late));
        world.matchGraphQL('trigger', {});
        await vi.waitFor(() => expect(first).toHaveLength(2));

        expect(late).toEqual([{ data: { n: 2 } }]);
    });
});
