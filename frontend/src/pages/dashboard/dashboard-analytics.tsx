import { format } from 'date-fns';
import { ChevronRight, Clock, Loader2, Wrench } from 'lucide-react';
import { useState } from 'react';
import { Area, AreaChart, Bar, BarChart, CartesianGrid, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts';

import type { UsageStatsPeriod } from '@/graphql/types';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible';
import {
    useFlowsExecutionStatsByPeriodQuery,
    useFlowsStatsByPeriodQuery,
    useToolcallsStatsByPeriodQuery,
    useUsageStatsByPeriodQuery,
} from '@/graphql/types';
import { formatCost, formatDuration, formatNumber, formatTokenCount } from '@/pages/dashboard/format-utils';

const CHART_COLORS = {
    area1: 'var(--color-chart-1)',
    area2: 'var(--color-chart-2)',
    area3: 'var(--color-chart-3)',
    bar1: 'var(--color-chart-4)',
    bar2: 'var(--color-chart-5)',
};

const formatDateLabel = (dateString: string): string => {
    try {
        return format(new Date(dateString), 'MMM d');
    } catch {
        return dateString;
    }
};

const ChartLoading = () => (
    <div className="flex h-[300px] items-center justify-center">
        <Loader2 className="text-muted-foreground size-6 animate-spin" />
    </div>
);

const CustomTooltip = ({
    active,
    formatter,
    label,
    payload,
}: {
    active?: boolean;
    formatter?: (value: number, name: string) => string;
    label?: string;
    payload?: Array<{ color: string; name: string; value: number }>;
}) => {
    if (!active || !payload?.length) {
        return null;
    }

    return (
        <div className="bg-popover text-popover-foreground rounded-lg border px-3 py-2 shadow-md">
            <p className="text-muted-foreground mb-1 text-xs">{label ? formatDateLabel(label) : ''}</p>
            {payload.map((entry) => (
                <div
                    className="flex items-center gap-2 text-sm"
                    key={entry.name}
                >
                    <span
                        className="size-2 rounded-full"
                        style={{ backgroundColor: entry.color }}
                    />
                    <span className="text-muted-foreground">{entry.name}:</span>
                    <span className="font-medium">
                        {formatter ? formatter(entry.value, entry.name) : formatNumber(entry.value)}
                    </span>
                </div>
            ))}
        </div>
    );
};

export const DashboardAnalytics = ({ period }: { period: UsageStatsPeriod }) => {
    const { data: usageByPeriodData, loading: usageByPeriodLoading } = useUsageStatsByPeriodQuery({
        variables: { period },
    });
    const { data: toolcallsByPeriodData, loading: toolcallsByPeriodLoading } = useToolcallsStatsByPeriodQuery({
        variables: { period },
    });
    const { data: flowsByPeriodData, loading: flowsByPeriodLoading } = useFlowsStatsByPeriodQuery({
        variables: { period },
    });
    const { data: executionStatsData, loading: executionStatsLoading } = useFlowsExecutionStatsByPeriodQuery({
        variables: { period },
    });

    const usageChartData = [...(usageByPeriodData?.usageStatsByPeriod ?? [])].reverse().map((item) => ({
        cacheIn: item.stats.totalUsageCacheIn,
        costIn: item.stats.totalUsageCostIn,
        costOut: item.stats.totalUsageCostOut,
        date: formatDateLabel(item.date),
        tokensIn: item.stats.totalUsageIn,
        tokensOut: item.stats.totalUsageOut,
        totalCost: item.stats.totalUsageCostIn + item.stats.totalUsageCostOut,
    }));

    const toolcallsChartData = [...(toolcallsByPeriodData?.toolcallsStatsByPeriod ?? [])].reverse().map((item) => ({
        count: item.stats.totalCount,
        date: formatDateLabel(item.date),
        duration: item.stats.totalDurationSeconds,
    }));

    const flowsChartData = [...(flowsByPeriodData?.flowsStatsByPeriod ?? [])].reverse().map((item) => ({
        assistants: item.stats.totalAssistantsCount,
        date: formatDateLabel(item.date),
        flows: item.stats.totalFlowsCount,
        subtasks: item.stats.totalSubtasksCount,
        tasks: item.stats.totalTasksCount,
    }));

    const executionStats = executionStatsData?.flowsExecutionStatsByPeriod ?? [];

    return (
        <div className="flex flex-col gap-6">
            <div className="grid gap-6 lg:grid-cols-2">
                <Card>
                    <CardHeader>
                        <CardTitle>Token Usage Over Time</CardTitle>
                        <CardDescription>Input and output tokens processed daily</CardDescription>
                    </CardHeader>
                    <CardContent>
                        {usageByPeriodLoading ? (
                            <ChartLoading />
                        ) : (
                            <ResponsiveContainer
                                height={300}
                                width="100%"
                            >
                                <AreaChart data={usageChartData}>
                                    <CartesianGrid
                                        className="stroke-border"
                                        strokeDasharray="3 3"
                                    />
                                    <XAxis
                                        dataKey="date"
                                        tick={{ fill: 'var(--color-muted-foreground)', fontSize: 12 }}
                                        tickMargin={8}
                                    />
                                    <YAxis
                                        tick={{ fill: 'var(--color-muted-foreground)', fontSize: 12 }}
                                        tickFormatter={formatTokenCount}
                                        tickMargin={8}
                                    />
                                    <Tooltip
                                        content={<CustomTooltip formatter={(value) => formatTokenCount(value)} />}
                                    />
                                    <Area
                                        dataKey="tokensIn"
                                        fill={CHART_COLORS.area1}
                                        fillOpacity={0.3}
                                        name="Tokens In"
                                        stroke={CHART_COLORS.area1}
                                        type="monotone"
                                    />
                                    <Area
                                        dataKey="tokensOut"
                                        fill={CHART_COLORS.area2}
                                        fillOpacity={0.3}
                                        name="Tokens Out"
                                        stroke={CHART_COLORS.area2}
                                        type="monotone"
                                    />
                                </AreaChart>
                            </ResponsiveContainer>
                        )}
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader>
                        <CardTitle>Cost Over Time</CardTitle>
                        <CardDescription>LLM spending per day</CardDescription>
                    </CardHeader>
                    <CardContent>
                        {usageByPeriodLoading ? (
                            <ChartLoading />
                        ) : (
                            <ResponsiveContainer
                                height={300}
                                width="100%"
                            >
                                <AreaChart data={usageChartData}>
                                    <CartesianGrid
                                        className="stroke-border"
                                        strokeDasharray="3 3"
                                    />
                                    <XAxis
                                        dataKey="date"
                                        tick={{ fill: 'var(--color-muted-foreground)', fontSize: 12 }}
                                        tickMargin={8}
                                    />
                                    <YAxis
                                        tick={{ fill: 'var(--color-muted-foreground)', fontSize: 12 }}
                                        tickFormatter={(value) => formatCost(value)}
                                        tickMargin={8}
                                    />
                                    <Tooltip content={<CustomTooltip formatter={(value) => formatCost(value)} />} />
                                    <Area
                                        dataKey="costIn"
                                        fill={CHART_COLORS.area1}
                                        fillOpacity={0.3}
                                        name="Cost In"
                                        stroke={CHART_COLORS.area1}
                                        type="monotone"
                                    />
                                    <Area
                                        dataKey="costOut"
                                        fill={CHART_COLORS.area3}
                                        fillOpacity={0.3}
                                        name="Cost Out"
                                        stroke={CHART_COLORS.area3}
                                        type="monotone"
                                    />
                                </AreaChart>
                            </ResponsiveContainer>
                        )}
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader>
                        <CardTitle>Tool Calls Over Time</CardTitle>
                        <CardDescription>Number of tool executions per day</CardDescription>
                    </CardHeader>
                    <CardContent>
                        {toolcallsByPeriodLoading ? (
                            <ChartLoading />
                        ) : (
                            <ResponsiveContainer
                                height={300}
                                width="100%"
                            >
                                <BarChart data={toolcallsChartData}>
                                    <CartesianGrid
                                        className="stroke-border"
                                        strokeDasharray="3 3"
                                    />
                                    <XAxis
                                        dataKey="date"
                                        tick={{ fill: 'var(--color-muted-foreground)', fontSize: 12 }}
                                        tickMargin={8}
                                    />
                                    <YAxis
                                        tick={{ fill: 'var(--color-muted-foreground)', fontSize: 12 }}
                                        tickMargin={8}
                                    />
                                    <Tooltip
                                        content={<CustomTooltip />}
                                        cursor={{ fill: 'var(--color-muted-foreground)', fillOpacity: 0.1 }}
                                    />
                                    <Bar
                                        dataKey="count"
                                        fill={CHART_COLORS.bar1}
                                        name="Tool Calls"
                                        radius={[4, 4, 0, 0]}
                                    />
                                </BarChart>
                            </ResponsiveContainer>
                        )}
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader>
                        <CardTitle>Flows Activity Over Time</CardTitle>
                        <CardDescription>Flows, tasks, and subtasks created per day</CardDescription>
                    </CardHeader>
                    <CardContent>
                        {flowsByPeriodLoading ? (
                            <ChartLoading />
                        ) : (
                            <ResponsiveContainer
                                height={300}
                                width="100%"
                            >
                                <BarChart data={flowsChartData}>
                                    <CartesianGrid
                                        className="stroke-border"
                                        strokeDasharray="3 3"
                                    />
                                    <XAxis
                                        dataKey="date"
                                        tick={{ fill: 'var(--color-muted-foreground)', fontSize: 12 }}
                                        tickMargin={8}
                                    />
                                    <YAxis
                                        tick={{ fill: 'var(--color-muted-foreground)', fontSize: 12 }}
                                        tickMargin={8}
                                    />
                                    <Tooltip
                                        content={<CustomTooltip />}
                                        cursor={{ fill: 'var(--color-muted-foreground)', fillOpacity: 0.1 }}
                                    />
                                    <Bar
                                        dataKey="flows"
                                        fill={CHART_COLORS.area1}
                                        name="Flows"
                                        radius={[4, 4, 0, 0]}
                                    />
                                    <Bar
                                        dataKey="tasks"
                                        fill={CHART_COLORS.area2}
                                        name="Tasks"
                                        radius={[4, 4, 0, 0]}
                                    />
                                    <Bar
                                        dataKey="subtasks"
                                        fill={CHART_COLORS.area3}
                                        name="Subtasks"
                                        radius={[4, 4, 0, 0]}
                                    />
                                </BarChart>
                            </ResponsiveContainer>
                        )}
                    </CardContent>
                </Card>
            </div>

            <Card>
                <CardHeader>
                    <CardTitle>Flow Execution Details</CardTitle>
                    <CardDescription>Execution time and tool calls breakdown per flow</CardDescription>
                </CardHeader>
                <CardContent>
                    {executionStatsLoading ? (
                        <div className="flex items-center justify-center py-8">
                            <Loader2 className="text-muted-foreground size-6 animate-spin" />
                        </div>
                    ) : !executionStats.length ? (
                        <p className="text-muted-foreground py-8 text-center text-sm">
                            No flow executions in this period
                        </p>
                    ) : (
                        <div className="space-y-1">
                            {executionStats.map((flow) => (
                                <FlowExecutionItem
                                    flow={flow}
                                    key={flow.flowId}
                                />
                            ))}
                        </div>
                    )}
                </CardContent>
            </Card>
        </div>
    );
};

type FlowExecution = {
    flowId: string;
    flowTitle: string;
    tasks: Array<{
        subtasks: Array<{
            subtaskId: string;
            subtaskTitle: string;
            totalDurationSeconds: number;
            totalToolcallsCount: number;
        }>;
        taskId: string;
        taskTitle: string;
        totalDurationSeconds: number;
        totalToolcallsCount: number;
    }>;
    totalAssistantsCount: number;
    totalDurationSeconds: number;
    totalToolcallsCount: number;
};

const FlowExecutionItem = ({ flow }: { flow: FlowExecution }) => {
    const [isOpen, setIsOpen] = useState(false);

    return (
        <Collapsible
            onOpenChange={setIsOpen}
            open={isOpen}
        >
            <CollapsibleTrigger className="hover:bg-muted/50 flex w-full items-center gap-3 rounded-lg px-3 py-2 text-left transition-colors">
                <ChevronRight className={`size-4 shrink-0 transition-transform ${isOpen ? 'rotate-90' : ''}`} />
                <div className="flex-1 truncate font-medium">{flow.flowTitle || `Flow #${flow.flowId}`}</div>
                <div className="text-muted-foreground flex items-center gap-4 text-sm">
                    <span className="flex items-center gap-1">
                        <Clock className="size-3" />
                        {formatDuration(flow.totalDurationSeconds)}
                    </span>
                    <span className="flex items-center gap-1">
                        <Wrench className="size-3" />
                        {flow.totalToolcallsCount}
                    </span>
                </div>
            </CollapsibleTrigger>
            <CollapsibleContent>
                <div className="ml-7 space-y-1 border-l pl-3">
                    {flow.tasks.map((task) => (
                        <TaskExecutionItem
                            key={task.taskId}
                            task={task}
                        />
                    ))}
                </div>
            </CollapsibleContent>
        </Collapsible>
    );
};

const TaskExecutionItem = ({ task }: { task: FlowExecution['tasks'][number] }) => {
    const [isOpen, setIsOpen] = useState(false);
    const hasSubtasks = task.subtasks.length > 0;

    return (
        <Collapsible
            onOpenChange={setIsOpen}
            open={isOpen}
        >
            <CollapsibleTrigger
                className="hover:bg-muted/50 flex w-full items-center gap-3 rounded-lg px-3 py-1.5 text-left text-sm transition-colors"
                disabled={!hasSubtasks}
            >
                {hasSubtasks ? (
                    <ChevronRight className={`size-3 shrink-0 transition-transform ${isOpen ? 'rotate-90' : ''}`} />
                ) : (
                    <span className="size-3 shrink-0" />
                )}
                <div className="text-muted-foreground flex-1 truncate">{task.taskTitle || `Task #${task.taskId}`}</div>
                <div className="text-muted-foreground flex items-center gap-4 text-xs">
                    <span className="flex items-center gap-1">
                        <Clock className="size-3" />
                        {formatDuration(task.totalDurationSeconds)}
                    </span>
                    <span className="flex items-center gap-1">
                        <Wrench className="size-3" />
                        {task.totalToolcallsCount}
                    </span>
                </div>
            </CollapsibleTrigger>
            {hasSubtasks && (
                <CollapsibleContent>
                    <div className="ml-6 space-y-0.5 border-l pl-3">
                        {task.subtasks.map((subtask) => (
                            <div
                                className="text-muted-foreground flex items-center gap-3 px-3 py-1 text-xs"
                                key={subtask.subtaskId}
                            >
                                <div className="flex-1 truncate">
                                    {subtask.subtaskTitle || `Subtask #${subtask.subtaskId}`}
                                </div>
                                <div className="flex items-center gap-4">
                                    <span className="flex items-center gap-1">
                                        <Clock className="size-3" />
                                        {formatDuration(subtask.totalDurationSeconds)}
                                    </span>
                                    <span className="flex items-center gap-1">
                                        <Wrench className="size-3" />
                                        {subtask.totalToolcallsCount}
                                    </span>
                                </div>
                            </div>
                        ))}
                    </div>
                </CollapsibleContent>
            )}
        </Collapsible>
    );
};
