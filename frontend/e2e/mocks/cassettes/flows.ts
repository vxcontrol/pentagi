import type { ResultOf } from '@graphql-typed-document-node/core';

import type {
    AssistantsDocument,
    FlowDocument,
    FlowFragmentFragment,
    MessageLogFragmentFragment,
    TerminalLogFragmentFragment,
} from '@/graphql/types';

import { MessageLogType, ProviderType, ResultFormat, StatusType, TerminalLogType, TerminalType } from '@/graphql/types';

import type { Cassette } from '../cassette.ts';

import { entity, mergeCassettes } from '../cassette.ts';
import { baseQueries, baseRest } from './base.ts';

const T = '2026-01-15T11:30:00Z';

export const PROVIDER = entity('Provider', { name: 'E2E Provider', type: ProviderType.Custom });

const terminal = (id: string) =>
    entity('Terminal', {
        connected: true,
        createdAt: T,
        id,
        image: 'debian:stable-slim',
        name: `terminal-${id}`,
        type: TerminalType.Primary,
    });

export const makeFlow = (id: string, title: string, status: StatusType = StatusType.Running): FlowFragmentFragment =>
    entity('Flow', {
        createdAt: T,
        id,
        provider: PROVIDER,
        status,
        terminals: [terminal(id)],
        title,
        updatedAt: T,
    });

export const makeMessage = (id: string, flowId: string): MessageLogFragmentFragment =>
    entity('MessageLog', {
        createdAt: T,
        flowId,
        id,
        message: `Message ${id} for flow ${flowId}`,
        result: '',
        resultFormat: ResultFormat.Plain,
        subtaskId: null,
        taskId: null,
        thinking: null,
        type: MessageLogType.Answer,
    });

export const FLOW_A = makeFlow('5', 'E2E Alpha');
export const FLOW_B = makeFlow('6', 'E2E Beta');

export const TERMINAL_OUTPUT = 'Linux e2e-sandbox 6.1.0 x86_64 GNU/Linux';

const terminalLog = (id: string, type: TerminalLogType, text: string): TerminalLogFragmentFragment =>
    entity('TerminalLog', {
        createdAt: T,
        flowId: '5',
        id,
        subtaskId: null,
        taskId: null,
        terminal: '5',
        text,
        type,
    });

const FLOW_A_TERMINAL_LOGS = [
    terminalLog('301', TerminalLogType.Stdin, 'uname -a'),
    terminalLog('302', TerminalLogType.Stdout, TERMINAL_OUTPUT),
];

export const FLOW_A_INITIAL_IDS = ['101', '102', '103'];
export const FLOW_A_STREAMED_IDS = ['104', '105'];
export const FLOW_A_RECONNECT_ID = '106';
export const FLOW_B_INITIAL_IDS = ['201', '202'];
export const FLOW_B_STREAMED_IDS = ['203', '204'];

const messagesFor = (flowId: string, ids: string[]) => ids.map((id) => makeMessage(id, flowId));

export const flowQueryData = (
    flow: FlowFragmentFragment,
    messageLogs: MessageLogFragmentFragment[],
    terminalLogs: TerminalLogFragmentFragment[] = [],
): ResultOf<typeof FlowDocument> => ({
    agentLogs: [],
    flow,
    messageLogs,
    screenshots: [],
    searchLogs: [],
    tasks: [],
    terminalLogs,
    vectorStoreLogs: [],
});

const noAssistants: ResultOf<typeof AssistantsDocument> = { assistants: [] };

const addedFrame = (message: MessageLogFragmentFragment, delayMs: number) => ({
    delayMs,
    payload: { data: { messageLogAdded: message } },
});

/**
 * Two concurrent flows with disjoint message streams. The second `flow` entry
 * for flow 5 is the post-reconnect reconcile response: everything delivered so
 * far plus the message "missed" while the socket was down.
 */
export const flowsCassette = (override: Cassette = {}): Cassette =>
    mergeCassettes(
        {
            queries: {
                ...baseQueries(),
                assistants: [{ data: noAssistants }],
                flow: [
                    {
                        data: flowQueryData(FLOW_A, messagesFor('5', FLOW_A_INITIAL_IDS), FLOW_A_TERMINAL_LOGS),
                        variables: { id: '5' },
                    },
                    {
                        data: flowQueryData(
                            FLOW_A,
                            messagesFor('5', [...FLOW_A_INITIAL_IDS, ...FLOW_A_STREAMED_IDS, FLOW_A_RECONNECT_ID]),
                            FLOW_A_TERMINAL_LOGS,
                        ),
                        variables: { id: '5' },
                    },
                    { data: flowQueryData(FLOW_B, messagesFor('6', FLOW_B_INITIAL_IDS)), variables: { id: '6' } },
                ],
                flows: [{ data: { flows: [FLOW_A, FLOW_B] } }],
            },
            rest: baseRest(),
            subscriptions: {
                messageLogAdded: [
                    {
                        frames: FLOW_A_STREAMED_IDS.map((id) => addedFrame(makeMessage(id, '5'), 80)),
                        variables: { flowId: '5' },
                    },
                    {
                        frames: FLOW_B_STREAMED_IDS.map((id) => addedFrame(makeMessage(id, '6'), 80)),
                        variables: { flowId: '6' },
                    },
                ],
            },
        },
        override,
    );
