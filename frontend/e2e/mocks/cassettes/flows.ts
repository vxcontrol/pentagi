import type { ResultOf } from '@graphql-typed-document-node/core';

import type {
    AssistantsDocument,
    FlowDocument,
    FlowFilesDocument,
    FlowFragmentFragment,
    FlowReportDocument,
    FlowStatsByFlowDocument,
    MessageLogFragmentFragment,
    TerminalLogFragmentFragment,
    ToolcallsStatsByFlowDocument,
    ToolcallsStatsByFunctionForFlowDocument,
    UsageStatsByAgentTypeForFlowDocument,
    UsageStatsByFlowDocument,
    UsageStatsByModelAgentsForFlowDocument,
    UsageStatsFragmentFragment,
} from '@/graphql/types';

import {
    AgentType,
    MessageLogType,
    ProviderType,
    ResultFormat,
    StatusType,
    TerminalLogType,
    TerminalType,
    VectorStoreAction,
} from '@/graphql/types';

import type { Cassette } from '../cassette.ts';

import { RECONNECTED_FLAG } from '../../helpers/reconnect.ts';
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

export const makeMessage = (
    id: string,
    flowId: string,
    overrides: Partial<MessageLogFragmentFragment> = {},
): MessageLogFragmentFragment =>
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
        ...overrides,
    });

export const VARIED_MESSAGES = [
    makeMessage('501', '5', { message: 'Planning the run', thinking: 'internal reasoning about the plan' }),
    makeMessage('502', '5', {
        message: 'Task finished',
        result: '# Report\n\nCommand executed successfully.',
        resultFormat: ResultFormat.Markdown,
        type: MessageLogType.Report,
    }),
    // Terminal, not Report: `getMessageType` pairs the terminal tool's output with the terminal
    // type, so a Report carrying a Terminal body is a state the backend cannot emit.
    makeMessage('503', '5', {
        message: '',
        result: 'e2e-terminal-marker\nexit 0',
        resultFormat: ResultFormat.Terminal,
        type: MessageLogType.Terminal,
    }),
    makeMessage('504', '5', { message: 'run the smoke command', type: MessageLogType.Input }),
];

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
export const FLOW_A_REPLAY_SENTINEL_ID = '107';
/** Gates the resubscribe replay; the spec raises it once the reconnect has landed. */
export const REPLAY_FLAG = 'flow-a-replay';
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
                        whenFlag: RECONNECTED_FLAG,
                    },
                    { data: flowQueryData(FLOW_B, messagesFor('6', FLOW_B_INITIAL_IDS)), variables: { id: '6' } },
                ],
                flows: [{ data: { flows: [FLOW_A, FLOW_B] } }],
            },
            rest: baseRest(),
            subscriptions: {
                messageLogAdded: [
                    {
                        frames: [
                            ...FLOW_A_STREAMED_IDS.map((id) => addedFrame(makeMessage(id, '5'), 80)),
                            // A resubscribe replays what the refetch already delivered; the sentinel
                            // after it proves the replay was received.
                            {
                                ...addedFrame(makeMessage(FLOW_A_RECONNECT_ID, '5'), 0),
                                whenFlag: REPLAY_FLAG,
                            },
                            addedFrame(makeMessage(FLOW_A_REPLAY_SENTINEL_ID, '5'), 0),
                        ],
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

export const FLOW_B_SENTINEL_ID = '205';
export const PAGER_SWITCH_FLAG = 'pager-switched';

/** Holds flow B a frame back so the spec can prove the new stream still delivers after the switch. */
export const pagerStreamsCassette = (): Cassette =>
    flowsCassette({
        subscriptions: {
            messageLogAdded: [
                {
                    frames: FLOW_A_STREAMED_IDS.map((id) => addedFrame(makeMessage(id, '5'), 80)),
                    variables: { flowId: '5' },
                },
                {
                    frames: [
                        ...FLOW_B_STREAMED_IDS.map((id) => addedFrame(makeMessage(id, '6'), 80)),
                        { ...addedFrame(makeMessage(FLOW_B_SENTINEL_ID, '6'), 0), whenFlag: PAGER_SWITCH_FLAG },
                    ],
                    variables: { flowId: '6' },
                },
            ],
        },
    });

export const variedMessagesCassette = (override: Cassette = {}): Cassette =>
    mergeCassettes(
        flowsCassette({
            queries: { flow: [{ data: flowQueryData(FLOW_A, VARIED_MESSAGES), variables: { id: '5' } }] },
            subscriptions: { messageLogAdded: [{ frames: [], variables: { flowId: '5' } }] },
        }),
        override,
    );

const TABS_TASK = entity('Task', {
    createdAt: T,
    flowId: '5',
    id: '11',
    input: 'Enumerate the target',
    result: 'Enumeration complete',
    status: StatusType.Finished,
    subtasks: [
        entity('Subtask', {
            createdAt: T,
            description: 'Run an nmap sweep',
            id: '21',
            result: 'Ports 22, 80 open',
            status: StatusType.Finished,
            taskId: '11',
            title: 'E2E Subtask Scan',
            updatedAt: T,
        }),
    ],
    title: 'E2E Task Alpha',
    updatedAt: T,
});

const TABS_AGENT_LOG = entity('AgentLog', {
    createdAt: T,
    executor: AgentType.Pentester,
    flowId: '5',
    id: '41',
    initiator: AgentType.Adviser,
    result: 'Reconnaissance summary ready',
    subtaskId: null,
    task: 'E2E agent reconnaissance',
    taskId: '11',
});

const TABS_SEARCH_LOG = entity('SearchLog', {
    createdAt: T,
    engine: 'duckduckgo',
    executor: AgentType.Searcher,
    flowId: '5',
    id: '51',
    initiator: AgentType.Pentester,
    query: 'E2E search for the CVE',
    result: 'Found relevant advisories',
    subtaskId: null,
    taskId: '11',
});

const TABS_VECTOR_LOG = entity('VectorStoreLog', {
    action: VectorStoreAction.Retrieve,
    createdAt: T,
    executor: AgentType.Memorist,
    filter: '{}',
    flowId: '5',
    id: '61',
    initiator: AgentType.Pentester,
    query: 'E2E recall prior findings',
    result: 'Recalled 3 memories',
    subtaskId: null,
    taskId: '11',
});

export const TABS_SCREENSHOT_ID = '71';
export const TABS_SCREENSHOT_NAME = 'login-form';
export const TABS_SCREENSHOT_URL = 'https://e2e.invalid/login-form';

const TABS_SCREENSHOT = entity('Screenshot', {
    createdAt: T,
    flowId: '5',
    id: TABS_SCREENSHOT_ID,
    name: TABS_SCREENSHOT_NAME,
    subtaskId: null,
    taskId: '11',
    url: TABS_SCREENSHOT_URL,
});

// The file manager groups by path prefix (uploads/resources/container), so a file
// seeded outside those roots renders nowhere.
export const TABS_FILE_NAME = 'scan-report.txt';

const flowFiles: ResultOf<typeof FlowFilesDocument> = {
    flowFiles: [
        entity('FlowFile', {
            id: '1',
            isDir: false,
            modifiedAt: T,
            name: TABS_FILE_NAME,
            path: `uploads/${TABS_FILE_NAME}`,
            size: 2048,
        }),
        entity('FlowFile', {
            id: '2',
            isDir: false,
            modifiedAt: T,
            name: 'wordlist.txt',
            path: 'resources/wordlist.txt',
            size: 512,
        }),
    ],
};

// 1×1 transparent PNG: the Screenshots tab renders an <img>, and the hermetic gate
// records every unmatched call, so the blob needs an entry even though no assertion
// looks at its pixels.
const PNG_1X1 = Buffer.from(
    'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==',
    'base64',
);

const flowTabsData: ResultOf<typeof FlowDocument> = {
    agentLogs: [TABS_AGENT_LOG],
    flow: FLOW_A,
    messageLogs: [],
    screenshots: [TABS_SCREENSHOT],
    searchLogs: [TABS_SEARCH_LOG],
    tasks: [TABS_TASK],
    terminalLogs: [],
    vectorStoreLogs: [TABS_VECTOR_LOG],
};

const flowUsage: UsageStatsFragmentFragment = entity('UsageStats', {
    totalUsageCacheIn: 5,
    totalUsageCacheOut: 3,
    totalUsageCostIn: 0.01,
    totalUsageCostOut: 0.02,
    totalUsageIn: 100,
    totalUsageOut: 40,
});

const usageStatsByFlow: ResultOf<typeof UsageStatsByFlowDocument> = { usageStatsByFlow: flowUsage };

const usageStatsByAgentTypeForFlow: ResultOf<typeof UsageStatsByAgentTypeForFlowDocument> = {
    usageStatsByAgentTypeForFlow: [entity('AgentTypeUsageStats', { agentType: AgentType.Pentester, stats: flowUsage })],
};

const usageStatsByModelAgentsForFlow: ResultOf<typeof UsageStatsByModelAgentsForFlowDocument> = {
    usageStatsByModelAgentsForFlow: [
        entity('ModelAgentsUsageStats', {
            agentTypes: [AgentType.Pentester],
            model: 'e2e-model',
            provider: 'e2e-provider',
            stats: flowUsage,
        }),
    ],
};

const toolcallsStatsByFlow: ResultOf<typeof ToolcallsStatsByFlowDocument> = {
    toolcallsStatsByFlow: entity('ToolcallsStats', { totalCount: 12, totalDurationSeconds: 34 }),
};

const toolcallsStatsByFunctionForFlow: ResultOf<typeof ToolcallsStatsByFunctionForFlowDocument> = {
    toolcallsStatsByFunctionForFlow: [
        entity('FunctionToolcallsStats', {
            avgDurationSeconds: 2.5,
            functionName: 'terminal',
            isAgent: false,
            totalCount: 8,
            totalDurationSeconds: 20,
        }),
    ],
};

const flowStatsByFlow: ResultOf<typeof FlowStatsByFlowDocument> = {
    flowStatsByFlow: entity('FlowStats', { totalAssistantsCount: 1, totalSubtasksCount: 3, totalTasksCount: 1 }),
};

/** Text that exists only in a streamed frame, so a panel that ignores its subscription stays blank. */
export const STREAMED = {
    agent: 'E2E streamed agent step',
    search: 'E2E streamed search',
    task: 'E2E Streamed Task',
    taskRetitled: 'E2E Task Alpha (revised)',
    terminal: 'e2e-streamed-terminal-line',
    vector: 'E2E streamed recall',
};

const streamedFrame = (payload: Record<string, unknown>) => ({ delayMs: 60, payload: { data: payload } });

export const livePanelsCassette = (): Cassette =>
    mergeCassettes(flowTabsCassette(), {
        subscriptions: {
            agentLogAdded: [
                {
                    frames: [
                        streamedFrame({
                            agentLogAdded: entity('AgentLog', {
                                createdAt: T,
                                executor: AgentType.Pentester,
                                flowId: '5',
                                id: '42',
                                initiator: AgentType.Adviser,
                                result: 'Streamed step complete',
                                subtaskId: null,
                                task: STREAMED.agent,
                                taskId: '11',
                            }),
                        }),
                    ],
                    variables: { flowId: '5' },
                },
            ],
            searchLogAdded: [
                {
                    frames: [
                        streamedFrame({
                            searchLogAdded: entity('SearchLog', {
                                createdAt: T,
                                engine: 'duckduckgo',
                                executor: AgentType.Searcher,
                                flowId: '5',
                                id: '52',
                                initiator: AgentType.Pentester,
                                query: STREAMED.search,
                                result: 'Streamed advisories',
                                subtaskId: null,
                                taskId: '11',
                            }),
                        }),
                    ],
                    variables: { flowId: '5' },
                },
            ],
            taskCreated: [
                {
                    frames: [
                        streamedFrame({
                            taskCreated: entity('Task', {
                                createdAt: T,
                                flowId: '5',
                                id: '12',
                                input: 'Exploit the open port',
                                result: '',
                                status: StatusType.Running,
                                subtasks: [],
                                title: STREAMED.task,
                                updatedAt: T,
                            }),
                        }),
                    ],
                    variables: { flowId: '5' },
                },
            ],
            taskUpdated: [
                {
                    frames: [
                        streamedFrame({
                            taskUpdated: { ...TABS_TASK, status: StatusType.Running, title: STREAMED.taskRetitled },
                        }),
                    ],
                    variables: { flowId: '5' },
                },
            ],
            terminalLogAdded: [
                {
                    frames: [
                        streamedFrame({
                            terminalLogAdded: terminalLog('303', TerminalLogType.Stdout, STREAMED.terminal),
                        }),
                    ],
                    variables: { flowId: '5' },
                },
            ],
            vectorStoreLogAdded: [
                {
                    frames: [
                        streamedFrame({
                            vectorStoreLogAdded: entity('VectorStoreLog', {
                                action: VectorStoreAction.Retrieve,
                                createdAt: T,
                                executor: AgentType.Memorist,
                                filter: '{}',
                                flowId: '5',
                                id: '62',
                                initiator: AgentType.Pentester,
                                query: STREAMED.vector,
                                result: 'Streamed recall complete',
                                subtaskId: null,
                                taskId: '11',
                            }),
                        }),
                    ],
                    variables: { flowId: '5' },
                },
            ],
        },
    });

const flowReportData: ResultOf<typeof FlowReportDocument> = { flow: FLOW_A, tasks: [TABS_TASK] };

/** The Report menu only appears when the flow query returns tasks, so this overrides `flow` too. */
export const flowReportCassette = (): Cassette =>
    flowsCassette({
        queries: {
            flow: [{ data: flowTabsData, variables: { id: '5' } }],
            flowReport: [{ data: flowReportData, variables: { id: '5' } }],
        },
        subscriptions: { messageLogAdded: [{ frames: [], variables: { flowId: '5' } }] },
    });

export const flowTabsCassette = (): Cassette =>
    flowsCassette({
        queries: {
            flow: [{ data: flowTabsData, variables: { id: '5' } }],
            flowFiles: [{ data: flowFiles, variables: { flowId: '5' } }],
            flowStatsByFlow: [{ data: flowStatsByFlow, variables: { flowId: '5' } }],
            toolcallsStatsByFlow: [{ data: toolcallsStatsByFlow, variables: { flowId: '5' } }],
            toolcallsStatsByFunctionForFlow: [{ data: toolcallsStatsByFunctionForFlow, variables: { flowId: '5' } }],
            usageStatsByAgentTypeForFlow: [{ data: usageStatsByAgentTypeForFlow, variables: { flowId: '5' } }],
            usageStatsByFlow: [{ data: usageStatsByFlow, variables: { flowId: '5' } }],
            usageStatsByModelAgentsForFlow: [{ data: usageStatsByModelAgentsForFlow, variables: { flowId: '5' } }],
        },
        rest: {
            [`GET /api/v1/flows/5/screenshots/${TABS_SCREENSHOT_ID}/file`]: [
                { body: PNG_1X1, contentType: 'image/png' },
            ],
        },
        subscriptions: { messageLogAdded: [{ frames: [], variables: { flowId: '5' } }] },
    });
