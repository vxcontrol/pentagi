import { format } from 'date-fns';
import { ChevronRight, Clock, Loader2, Wrench } from 'lucide-react';
import { useMemo, useRef, useState } from 'react';
import { Area, AreaChart, Bar, BarChart, CartesianGrid, Tooltip, XAxis, YAxis } from 'recharts';

import type { FlowFragmentFragment, UsageStatsPeriod } from '@/graphql/types';

import { ChartCard, ChartTooltip } from '@/components/dashboard';
import { FlowStatusBadge } from '@/components/icons/flow-status-badge';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible';
import {
    useFlowsExecutionStatsByPeriodQuery,
    useFlowsQuery,
    useFlowsStatsByPeriodQuery,
    useToolcallsStatsByPeriodQuery,
    useUsageStatsByPeriodQuery,
} from '@/graphql/types';
import { formatCost, formatDuration, formatNumber, formatTokenCount } from '@/lib/utils/format';

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

const axisTickStyle = { fill: 'var(--color-muted-foreground)', fontSize: 12 };

// Disables tooltip animation on chart entry and re-enables it only after the
// tooltip has rendered at the correct position for the first time in the session.
// This prevents the tooltip from flying from (0,0) to the cursor on entry,
// while keeping smooth follow animation for all subsequent movements.
const useChartTooltipAnimation = () => {
    const [isAnimationActive, setIsAnimationActive] = useState(false);
    const [sessionKey, setSessionKey] = useState(0);
    const rafRef = useRef<number | undefined>(undefined);

    const onMouseEnter = () => {
        cancelAnimationFrame(rafRef.current!);
        setIsAnimationActive(false);
        setSessionKey((k) => k + 1);
    };

    const onMouseLeave = () => {
        cancelAnimationFrame(rafRef.current!);
        setIsAnimationActive(false);
    };

    // Called by ChartTooltip the first time it becomes visible in this session.
    // One rAF ensures the tooltip has painted at its initial position before
    // animation is re-enabled for subsequent cursor movements.
    const onFirstActive = () => {
        cancelAnimationFrame(rafRef.current!);
        rafRef.current = requestAnimationFrame(() => setIsAnimationActive(true));
    };

    return { isAnimationActive, onFirstActive, onMouseEnter, onMouseLeave, sessionKey };
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
    const { data: flowsData } = useFlowsQuery();

    const flowsTooltip = useChartTooltipAnimation();
    const toolcallsTooltip = useChartTooltipAnimation();
    const tokenUsageTooltip = useChartTooltipAnimation();
    const costTooltip = useChartTooltipAnimation();

    const flowsById = useMemo(() => {
        const map = new Map<string, FlowFragmentFragment>();
        (flowsData?.flows ?? []).forEach((flow) => {
            map.set(flow.id, flow);
        });

        return map;
    }, [flowsData?.flows]);

    const usageChartData = [...(usageByPeriodData?.usageStatsByPeriod ?? [])].reverse().map((item) => ({
        cacheIn: item.stats.totalUsageCacheIn,
        costIn: item.stats.totalUsageCostIn,
        costOut: item.stats.totalUsageCostOut,
        date: item.date,
        tokensIn: item.stats.totalUsageIn,
        tokensOut: item.stats.totalUsageOut,
        totalCost: item.stats.totalUsageCostIn + item.stats.totalUsageCostOut,
    }));

    const toolcallsChartData = [...(toolcallsByPeriodData?.toolcallsStatsByPeriod ?? [])].reverse().map((item) => ({
        count: item.stats.totalCount,
        date: item.date,
        duration: item.stats.totalDurationSeconds,
    }));

    const flowsChartData = [...(flowsByPeriodData?.flowsStatsByPeriod ?? [])].reverse().map((item) => ({
        assistants: item.stats.totalAssistantsCount,
        date: item.date,
        flows: item.stats.totalFlowsCount,
        subtasks: item.stats.totalSubtasksCount,
        tasks: item.stats.totalTasksCount,
    }));

    const executionStats = executionStatsData?.flowsExecutionStatsByPeriod ?? [];

    return (
        <div className="flex flex-col gap-6">
            <ChartCard
                description="Flows, tasks, and subtasks created per day"
                empty={!flowsByPeriodLoading && flowsChartData.length === 0}
                height={320}
                loading={flowsByPeriodLoading}
                title="Flows Activity Over Time"
            >
                <BarChart
                    data={flowsChartData}
                    onMouseEnter={flowsTooltip.onMouseEnter}
                    onMouseLeave={flowsTooltip.onMouseLeave}
                >
                    <CartesianGrid
                        className="stroke-border"
                        strokeDasharray="3 3"
                    />
                    <XAxis
                        dataKey="date"
                        tick={axisTickStyle}
                        tickFormatter={formatDateLabel}
                        tickMargin={8}
                    />
                    <YAxis
                        tick={axisTickStyle}
                        tickMargin={8}
                    />
                    <Tooltip
                        content={
                            <ChartTooltip
                                labelFormatter={formatDateLabel}
                                onFirstActive={flowsTooltip.onFirstActive}
                                sessionKey={flowsTooltip.sessionKey}
                            />
                        }
                        cursor={{ fill: 'var(--color-muted-foreground)', fillOpacity: 0.1 }}
                        isAnimationActive={flowsTooltip.isAnimationActive}
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
            </ChartCard>

            <div className="grid gap-6 lg:grid-cols-2">
                <ChartCard
                    description="Number of tool executions per day"
                    empty={!toolcallsByPeriodLoading && toolcallsChartData.length === 0}
                    loading={toolcallsByPeriodLoading}
                    title="Tool Calls Over Time"
                >
                    <BarChart
                        data={toolcallsChartData}
                        onMouseEnter={toolcallsTooltip.onMouseEnter}
                        onMouseLeave={toolcallsTooltip.onMouseLeave}
                    >
                        <CartesianGrid
                            className="stroke-border"
                            strokeDasharray="3 3"
                        />
                        <XAxis
                            dataKey="date"
                            tick={axisTickStyle}
                            tickFormatter={formatDateLabel}
                            tickMargin={8}
                        />
                        <YAxis
                            tick={axisTickStyle}
                            tickMargin={8}
                        />
                        <Tooltip
                            content={
                                <ChartTooltip
                                    labelFormatter={formatDateLabel}
                                    onFirstActive={toolcallsTooltip.onFirstActive}
                                    sessionKey={toolcallsTooltip.sessionKey}
                                />
                            }
                            cursor={{ fill: 'var(--color-muted-foreground)', fillOpacity: 0.1 }}
                            isAnimationActive={toolcallsTooltip.isAnimationActive}
                        />
                        <Bar
                            dataKey="count"
                            fill={CHART_COLORS.bar1}
                            name="Tool Calls"
                            radius={[4, 4, 0, 0]}
                        />
                    </BarChart>
                </ChartCard>

                <ChartCard
                    description="Input and output tokens processed daily"
                    empty={!usageByPeriodLoading && usageChartData.length === 0}
                    loading={usageByPeriodLoading}
                    title="Token Usage Over Time"
                >
                    <AreaChart
                        data={usageChartData}
                        onMouseEnter={tokenUsageTooltip.onMouseEnter}
                        onMouseLeave={tokenUsageTooltip.onMouseLeave}
                    >
                        <CartesianGrid
                            className="stroke-border"
                            strokeDasharray="3 3"
                        />
                        <XAxis
                            dataKey="date"
                            tick={axisTickStyle}
                            tickFormatter={formatDateLabel}
                            tickMargin={8}
                        />
                        <YAxis
                            tick={axisTickStyle}
                            tickFormatter={formatTokenCount}
                            tickMargin={8}
                        />
                        <Tooltip
                            content={
                                <ChartTooltip
                                    formatter={(value) => formatTokenCount(value)}
                                    labelFormatter={formatDateLabel}
                                    onFirstActive={tokenUsageTooltip.onFirstActive}
                                    sessionKey={tokenUsageTooltip.sessionKey}
                                />
                            }
                            isAnimationActive={tokenUsageTooltip.isAnimationActive}
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
                </ChartCard>
            </div>

            <ChartCard
                description="LLM spending per day. May stay near zero when using local engines — this is expected."
                empty={!usageByPeriodLoading && usageChartData.length === 0}
                height={240}
                loading={usageByPeriodLoading}
                title="Cost Over Time"
            >
                <AreaChart
                    data={usageChartData}
                    onMouseEnter={costTooltip.onMouseEnter}
                    onMouseLeave={costTooltip.onMouseLeave}
                >
                    <CartesianGrid
                        className="stroke-border"
                        strokeDasharray="3 3"
                    />
                    <XAxis
                        dataKey="date"
                        tick={axisTickStyle}
                        tickFormatter={formatDateLabel}
                        tickMargin={8}
                    />
                    <YAxis
                        tick={axisTickStyle}
                        tickFormatter={(value) => formatCost(value)}
                        tickMargin={8}
                    />
                    <Tooltip
                        content={
                            <ChartTooltip
                                formatter={(value) => formatCost(value)}
                                labelFormatter={formatDateLabel}
                                onFirstActive={costTooltip.onFirstActive}
                                sessionKey={costTooltip.sessionKey}
                            />
                        }
                        isAnimationActive={costTooltip.isAnimationActive}
                    />
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
            </ChartCard>

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
                                    flowMeta={flowsById.get(flow.flowId)}
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

const FlowExecutionItem = ({ flow, flowMeta }: { flow: FlowExecution; flowMeta?: FlowFragmentFragment }) => {
    const [isOpen, setIsOpen] = useState(false);
    const taskCount = flow.tasks.length;
    const subtaskCount = flow.tasks.reduce((sum, task) => sum + task.subtasks.length, 0);

    return (
        <Collapsible
            onOpenChange={setIsOpen}
            open={isOpen}
        >
            <CollapsibleTrigger className="hover:bg-muted/50 group flex w-full items-start gap-3 rounded-lg px-3 py-2 text-left transition-colors">
                <ChevronRight className={`mt-1 size-4 shrink-0 transition-transform ${isOpen ? 'rotate-90' : ''}`} />
                <div className="min-w-0 flex-1">
                    <div className="flex flex-wrap items-center gap-2">
                        <span className="truncate font-medium">{flow.flowTitle || `Flow #${flow.flowId}`}</span>
                        {flowMeta?.status && <FlowStatusBadge status={flowMeta.status} />}
                        {flowMeta?.provider?.name && <Badge variant="secondary">{flowMeta.provider.name}</Badge>}
                    </div>
                    <div className="text-muted-foreground mt-0.5 text-xs">
                        {taskCount} {taskCount === 1 ? 'task' : 'tasks'}
                        {subtaskCount > 0 && ` · ${subtaskCount} ${subtaskCount === 1 ? 'subtask' : 'subtasks'}`}
                        {flow.totalAssistantsCount > 0 &&
                            ` · ${flow.totalAssistantsCount} ${flow.totalAssistantsCount === 1 ? 'assistant' : 'assistants'}`}
                    </div>
                </div>
                <div className="text-muted-foreground flex shrink-0 items-center gap-4 pt-1 text-sm">
                    <span className="flex items-center gap-1">
                        <Clock className="size-3" />
                        {formatDuration(flow.totalDurationSeconds)}
                    </span>
                    <span className="flex items-center gap-1">
                        <Wrench className="size-3" />
                        {formatNumber(flow.totalToolcallsCount)}
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
                        {formatNumber(task.totalToolcallsCount)}
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
                                        {formatNumber(subtask.totalToolcallsCount)}
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
