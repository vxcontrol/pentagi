import type { ResultOf } from '@graphql-typed-document-node/core';

import type { AssistantFragmentFragment, AssistantLogFragmentFragment, AssistantsDocument } from '@/graphql/types';

import { MessageLogType, ResultFormat, StatusType } from '@/graphql/types';

import type { Cassette } from '../cassette.ts';

import { entity, mergeCassettes } from '../cassette.ts';
import { flowsCassette, PROVIDER } from './flows.ts';

const T = '2026-01-15T11:30:00Z';

export const makeAssistant = (
    id: string,
    title: string,
    status: StatusType = StatusType.Waiting,
): AssistantFragmentFragment =>
    entity('Assistant', {
        createdAt: T,
        flowId: '5',
        id,
        provider: PROVIDER,
        status,
        title,
        updatedAt: T,
        useAgents: false,
    });

export const WAITING_ASSISTANT = makeAssistant('11', 'E2E Assistant Waiting');
export const RUNNING_ASSISTANT = makeAssistant('12', 'E2E Assistant Running', StatusType.Running);
export const CREATED_ASSISTANT = makeAssistant('13', 'E2E Assistant Created');

export const ASSISTANT_ANSWER = 'The scan found an open port on 8080.';

const assistantLog = (id: string, assistantId: string, message: string): AssistantLogFragmentFragment =>
    entity('AssistantLog', {
        appendPart: false,
        assistantId,
        createdAt: T,
        flowId: '5',
        id,
        message,
        result: ASSISTANT_ANSWER,
        resultFormat: ResultFormat.Markdown,
        thinking: null,
        type: MessageLogType.Answer,
    });

const seeded: ResultOf<typeof AssistantsDocument> = { assistants: [WAITING_ASSISTANT, RUNNING_ASSISTANT] };

export const assistantsCassette = (override: Cassette = {}): Cassette =>
    mergeCassettes(
        flowsCassette({
            queries: {
                assistantLogs: [
                    {
                        data: { assistantLogs: [assistantLog('41', '11', 'what did the scan find?')] },
                        variables: { assistantId: '11', flowId: '5' },
                    },
                    {
                        data: { assistantLogs: [assistantLog('42', '12', 'keep going')] },
                        variables: { assistantId: '12', flowId: '5' },
                    },
                    { data: { assistantLogs: [] }, variables: { assistantId: '13', flowId: '5' } },
                ],
                assistants: [{ data: seeded, variables: { flowId: '5' } }],
                providers: [{ data: { providers: [PROVIDER] } }],
            },
        }),
        override,
    );
