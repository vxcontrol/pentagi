import type { ResultOf } from '@graphql-typed-document-node/core';

import type {
    FlowsExecutionStatsByPeriodDocument,
    FlowsStatsByPeriodDocument,
    FlowsStatsTotalDocument,
    ToolcallsStatsByFunctionDocument,
    ToolcallsStatsByPeriodDocument,
    ToolcallsStatsTotalDocument,
    UsageStatsByAgentTypeDocument,
    UsageStatsByModelDocument,
    UsageStatsByPeriodDocument,
    UsageStatsByProviderDocument,
    UsageStatsFragmentFragment,
    UsageStatsTotalDocument,
} from '@/graphql/types';

import { AgentType, UsageStatsPeriod } from '@/graphql/types';

import type { Cassette } from '../cassette.ts';

import { entity, mergeCassettes } from '../cassette.ts';
import { baseQueries, baseRest } from './base.ts';

const DATES = ['2026-01-15', '2026-01-14', '2026-01-13'];
// The month period returns a distinct date range so the period switch is
// asserted on rendered data (the x-axis labels), not just the refetch firing.
const MONTH_DATES = ['2026-08-15', '2026-08-10', '2026-08-05'];
const FLOW_TITLES = { month: 'E2E Beta', week: 'E2E Alpha' };

const usageStats = (seed: number): UsageStatsFragmentFragment =>
    entity('UsageStats', {
        totalUsageCacheIn: seed * 5,
        totalUsageCacheOut: seed * 3,
        totalUsageCostIn: seed * 0.01,
        totalUsageCostOut: seed * 0.02,
        totalUsageIn: seed * 100,
        totalUsageOut: seed * 40,
    });

const usageStatsByPeriodFor = (dates: string[]): ResultOf<typeof UsageStatsByPeriodDocument> => ({
    usageStatsByPeriod: dates.map((date, index) => entity('DailyUsageStats', { date, stats: usageStats(index + 1) })),
});

const toolcallsStatsByPeriodFor = (dates: string[]): ResultOf<typeof ToolcallsStatsByPeriodDocument> => ({
    toolcallsStatsByPeriod: dates.map((date, index) =>
        entity('DailyToolcallsStats', {
            date,
            stats: entity('ToolcallsStats', { totalCount: index + 2, totalDurationSeconds: (index + 1) * 6 }),
        }),
    ),
});

const flowsStatsByPeriodFor = (dates: string[]): ResultOf<typeof FlowsStatsByPeriodDocument> => ({
    flowsStatsByPeriod: dates.map((date, index) =>
        entity('DailyFlowsStats', {
            date,
            stats: entity('FlowsStats', {
                totalAssistantsCount: index,
                totalFlowsCount: index + 1,
                totalSubtasksCount: (index + 1) * 3,
                totalTasksCount: (index + 1) * 2,
            }),
        }),
    ),
});

const flowsExecutionStatsByPeriodFor = (
    flowId: string,
    flowTitle: string,
): ResultOf<typeof FlowsExecutionStatsByPeriodDocument> => ({
    flowsExecutionStatsByPeriod: [
        entity('FlowExecutionStats', {
            flowId,
            flowTitle,
            tasks: [
                entity('TaskExecutionStats', {
                    subtasks: [
                        entity('SubtaskExecutionStats', {
                            subtaskId: '21',
                            subtaskTitle: 'E2E Subtask',
                            totalDurationSeconds: 30,
                            totalToolcallsCount: 2,
                        }),
                    ],
                    taskId: '11',
                    taskTitle: 'E2E Task',
                    totalDurationSeconds: 60,
                    totalToolcallsCount: 4,
                }),
            ],
            totalAssistantsCount: 1,
            totalDurationSeconds: 120,
            totalToolcallsCount: 7,
        }),
    ],
});

const usageStatsTotal: ResultOf<typeof UsageStatsTotalDocument> = { usageStatsTotal: usageStats(10) };

const usageStatsByProvider: ResultOf<typeof UsageStatsByProviderDocument> = {
    usageStatsByProvider: [entity('ProviderUsageStats', { provider: 'e2e-provider', stats: usageStats(3) })],
};

const usageStatsByModel: ResultOf<typeof UsageStatsByModelDocument> = {
    usageStatsByModel: [
        entity('ModelUsageStats', { model: 'e2e-model', provider: 'e2e-provider', stats: usageStats(2) }),
    ],
};

const usageStatsByAgentType: ResultOf<typeof UsageStatsByAgentTypeDocument> = {
    usageStatsByAgentType: [entity('AgentTypeUsageStats', { agentType: AgentType.Pentester, stats: usageStats(1) })],
};

const toolcallsStatsTotal: ResultOf<typeof ToolcallsStatsTotalDocument> = {
    toolcallsStatsTotal: entity('ToolcallsStats', { totalCount: 42, totalDurationSeconds: 360 }),
};

const toolcallsStatsByFunction: ResultOf<typeof ToolcallsStatsByFunctionDocument> = {
    toolcallsStatsByFunction: [
        entity('FunctionToolcallsStats', {
            avgDurationSeconds: 2,
            functionName: 'e2e_terminal',
            isAgent: false,
            totalCount: 21,
            totalDurationSeconds: 42,
        }),
    ],
};

const flowsStatsTotal: ResultOf<typeof FlowsStatsTotalDocument> = {
    flowsStatsTotal: entity('FlowsStats', {
        totalAssistantsCount: 1,
        totalFlowsCount: 8,
        totalSubtasksCount: 34,
        totalTasksCount: 13,
    }),
};

export const dashboardCassette = (override: Cassette = {}): Cassette =>
    mergeCassettes(
        {
            queries: {
                ...baseQueries(),
                flowsExecutionStatsByPeriod: [
                    {
                        data: flowsExecutionStatsByPeriodFor('5', FLOW_TITLES.week),
                        variables: { period: UsageStatsPeriod.Week },
                    },
                    {
                        data: flowsExecutionStatsByPeriodFor('6', FLOW_TITLES.month),
                        variables: { period: UsageStatsPeriod.Month },
                    },
                ],
                flowsStatsByPeriod: [
                    { data: flowsStatsByPeriodFor(DATES), variables: { period: UsageStatsPeriod.Week } },
                    { data: flowsStatsByPeriodFor(MONTH_DATES), variables: { period: UsageStatsPeriod.Month } },
                ],
                flowsStatsTotal: [{ data: flowsStatsTotal }],
                toolcallsStatsByFunction: [{ data: toolcallsStatsByFunction }],
                toolcallsStatsByPeriod: [
                    { data: toolcallsStatsByPeriodFor(DATES), variables: { period: UsageStatsPeriod.Week } },
                    { data: toolcallsStatsByPeriodFor(MONTH_DATES), variables: { period: UsageStatsPeriod.Month } },
                ],
                toolcallsStatsTotal: [{ data: toolcallsStatsTotal }],
                usageStatsByAgentType: [{ data: usageStatsByAgentType }],
                usageStatsByModel: [{ data: usageStatsByModel }],
                usageStatsByPeriod: [
                    { data: usageStatsByPeriodFor(DATES), variables: { period: UsageStatsPeriod.Week } },
                    { data: usageStatsByPeriodFor(MONTH_DATES), variables: { period: UsageStatsPeriod.Month } },
                ],
                usageStatsByProvider: [{ data: usageStatsByProvider }],
                usageStatsTotal: [{ data: usageStatsTotal }],
            },
            rest: baseRest(),
        },
        override,
    );
