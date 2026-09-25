import { describe, expect, it } from 'vitest';

import type { AssistantFragmentFragment } from '@/graphql/types';

import { StatusType } from '@/graphql/types';

import { groupAssistantsByState } from './assistant';

const assistant = (id: string, status: StatusType) =>
    ({ id, status, title: `assistant ${id}` }) as AssistantFragmentFragment;

describe('groupAssistantsByState', () => {
    it('lists an assistant that is still being prepared among the active ones', () => {
        const groups = groupAssistantsByState([assistant('1', StatusType.Created)]);

        expect(groups.active.map(({ assistant }) => assistant.id)).toEqual(['1']);
        expect(groups.failed).toHaveLength(0);
        expect(groups.finished).toHaveLength(0);
    });

    it('loses no assistant, whatever state it is in', () => {
        const assistants = [
            assistant('1', StatusType.Created),
            assistant('2', StatusType.Running),
            assistant('3', StatusType.Waiting),
            assistant('4', StatusType.Failed),
            assistant('5', StatusType.Finished),
        ];

        const groups = groupAssistantsByState(assistants);
        const grouped = [...groups.active, ...groups.failed, ...groups.finished];

        expect(grouped).toHaveLength(assistants.length);
        expect(grouped.map(({ assistant }) => assistant.id).sort()).toEqual(['1', '2', '3', '4', '5']);
    });

    it('numbers assistants by their position in the original list', () => {
        const groups = groupAssistantsByState([assistant('1', StatusType.Failed), assistant('2', StatusType.Running)]);

        expect(groups.failed[0]?.index).toBe(1);
        expect(groups.active[0]?.index).toBe(2);
    });
});
