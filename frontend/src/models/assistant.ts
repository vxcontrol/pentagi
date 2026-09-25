import type { AssistantFragmentFragment } from '@/graphql/types';

import { StatusType } from '@/graphql/types';

export interface AssistantGroups {
    active: AssistantItem[];
    failed: AssistantItem[];
    finished: AssistantItem[];
}

export interface AssistantItem {
    assistant: AssistantFragmentFragment;
    index: number;
}

const ACTIVE_STATUSES: StatusType[] = [StatusType.Created, StatusType.Running, StatusType.Waiting];

export const groupAssistantsByState = (assistants: AssistantFragmentFragment[]): AssistantGroups =>
    assistants.reduce<AssistantGroups>(
        (groups, assistant, position) => {
            const item = { assistant, index: position + 1 };

            if (ACTIVE_STATUSES.includes(assistant.status)) {
                groups.active.push(item);
            } else if (assistant.status === StatusType.Failed) {
                groups.failed.push(item);
            } else if (assistant.status === StatusType.Finished) {
                groups.finished.push(item);
            }

            return groups;
        },
        { active: [], failed: [], finished: [] },
    );
