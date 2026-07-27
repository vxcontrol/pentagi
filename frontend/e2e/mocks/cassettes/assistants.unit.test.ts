import { describe, expect, it } from 'vitest';

import { MockWorld } from '../world.ts';
import { assistantsCassette, CREATED_ASSISTANT, RUNNING_ASSISTANT, WAITING_ASSISTANT } from './assistants.ts';

const transcript = (world: MockWorld, assistantId: string): undefined | unknown[] =>
    (
        world.matchGraphQL('assistantLogs', { assistantId, flowId: '5' })?.data as
            | undefined
            | { assistantLogs: unknown[] }
    )?.assistantLogs;

describe('assistants cassette — assistantLogs', () => {
    it('serves an assistant its own transcript on every read, not only the first', () => {
        const world = new MockWorld(assistantsCassette());

        expect(transcript(world, WAITING_ASSISTANT.id)).toHaveLength(1);
        expect(transcript(world, WAITING_ASSISTANT.id)).toHaveLength(1);
        expect(transcript(world, RUNNING_ASSISTANT.id)).toHaveLength(1);
        expect(transcript(world, RUNNING_ASSISTANT.id)).toHaveLength(1);
    });

    it('answers the freshly created assistant with an empty transcript', () => {
        const world = new MockWorld(assistantsCassette());

        expect(transcript(world, CREATED_ASSISTANT.id)).toEqual([]);
    });
});
