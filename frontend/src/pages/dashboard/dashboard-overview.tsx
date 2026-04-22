import { Activity, CircleDollarSign, Cpu, GitFork, Loader2 } from 'lucide-react';

import type { UsageStatsFragmentFragment } from '@/graphql/types';

import { MetricCard } from '@/components/dashboard';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import {
    useFlowsStatsTotalQuery,
    useToolcallsStatsByFunctionQuery,
    useToolcallsStatsTotalQuery,
    useUsageStatsByAgentTypeQuery,
    useUsageStatsByModelQuery,
    useUsageStatsByProviderQuery,
    useUsageStatsTotalQuery,
} from '@/graphql/types';
import { formatCost, formatDuration, formatNumber, formatTokenCount } from '@/lib/utils/format';

const UsageStatsRow = ({ label, stats }: { label: string; stats: UsageStatsFragmentFragment }) => (
    <TableRow>
        <TableCell className="font-medium">{label}</TableCell>
        <TableCell className="text-right">{formatTokenCount(stats.totalUsageIn)}</TableCell>
        <TableCell className="text-right">{formatTokenCount(stats.totalUsageOut)}</TableCell>
        <TableCell className="text-right">{formatTokenCount(stats.totalUsageCacheIn)}</TableCell>
        <TableCell className="text-right">{formatTokenCount(stats.totalUsageCacheOut)}</TableCell>
        <TableCell className="text-right">{formatCost(stats.totalUsageCostIn)}</TableCell>
        <TableCell className="text-right">{formatCost(stats.totalUsageCostOut)}</TableCell>
        <TableCell className="text-right font-semibold">
            {formatCost(stats.totalUsageCostIn + stats.totalUsageCostOut)}
        </TableCell>
    </TableRow>
);

const UsageStatsTable = ({ rows }: { rows: Array<{ label: string; stats: UsageStatsFragmentFragment }> }) => (
    <Table>
        <TableHeader>
            <TableRow>
                <TableHead className="whitespace-nowrap">Name</TableHead>
                <TableHead className="whitespace-nowrap text-right">Tokens In</TableHead>
                <TableHead className="whitespace-nowrap text-right">Tokens Out</TableHead>
                <TableHead className="whitespace-nowrap text-right">Cache In</TableHead>
                <TableHead className="whitespace-nowrap text-right">Cache Out</TableHead>
                <TableHead className="whitespace-nowrap text-right">Cost In</TableHead>
                <TableHead className="whitespace-nowrap text-right">Cost Out</TableHead>
                <TableHead className="whitespace-nowrap text-right">Total Cost</TableHead>
            </TableRow>
        </TableHeader>
        <TableBody>
            {rows.map((row) => (
                <UsageStatsRow
                    key={row.label}
                    label={row.label}
                    stats={row.stats}
                />
            ))}
        </TableBody>
    </Table>
);

const LoadingTable = () => (
    <div className="flex items-center justify-center py-8">
        <Loader2 className="text-muted-foreground size-6 animate-spin" />
    </div>
);

export const DashboardOverview = () => {
    const { data: usageTotalData, loading: usageTotalLoading } = useUsageStatsTotalQuery();
    const { data: usageByProviderData, loading: usageByProviderLoading } = useUsageStatsByProviderQuery();
    const { data: usageByModelData, loading: usageByModelLoading } = useUsageStatsByModelQuery();
    const { data: usageByAgentTypeData, loading: usageByAgentTypeLoading } = useUsageStatsByAgentTypeQuery();
    const { data: toolcallsTotalData, loading: toolcallsTotalLoading } = useToolcallsStatsTotalQuery();
    const { data: toolcallsByFunctionData, loading: toolcallsByFunctionLoading } = useToolcallsStatsByFunctionQuery();
    const { data: flowsTotalData, loading: flowsTotalLoading } = useFlowsStatsTotalQuery();

    const usageTotal = usageTotalData?.usageStatsTotal;
    const toolcallsTotal = toolcallsTotalData?.toolcallsStatsTotal;
    const flowsTotal = flowsTotalData?.flowsStatsTotal;

    const totalCost = usageTotal ? usageTotal.totalUsageCostIn + usageTotal.totalUsageCostOut : 0;
    const totalTokens = usageTotal ? usageTotal.totalUsageIn + usageTotal.totalUsageOut : 0;

    const providerRows = (usageByProviderData?.usageStatsByProvider ?? []).map((item) => ({
        label: item.provider,
        stats: item.stats,
    }));
    const modelRows = (usageByModelData?.usageStatsByModel ?? []).map((item) => ({
        label: `${item.model} (${item.provider})`,
        stats: item.stats,
    }));
    const agentTypeRows = (usageByAgentTypeData?.usageStatsByAgentType ?? []).map((item) => ({
        label: item.agentType,
        stats: item.stats,
    }));

    const toolcallsByFunction = [...(toolcallsByFunctionData?.toolcallsStatsByFunction ?? [])].sort(
        (a, b) => b.totalCount - a.totalCount,
    );

    return (
        <div className="flex flex-col gap-6">
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
                <MetricCard
                    description={`Tasks: ${flowsTotal?.totalTasksCount ?? 0} · Subtasks: ${flowsTotal?.totalSubtasksCount ?? 0} · Assistants: ${flowsTotal?.totalAssistantsCount ?? 0}`}
                    icon={<GitFork className="text-muted-foreground size-4" />}
                    loading={flowsTotalLoading}
                    title="Total Flows"
                    value={flowsTotal ? formatNumber(flowsTotal.totalFlowsCount) : '0'}
                />
                <MetricCard
                    description={`Total duration: ${toolcallsTotal ? formatDuration(toolcallsTotal.totalDurationSeconds) : '—'}`}
                    icon={<Activity className="text-muted-foreground size-4" />}
                    loading={toolcallsTotalLoading}
                    title="Tool Calls"
                    value={toolcallsTotal ? formatNumber(toolcallsTotal.totalCount) : '0'}
                />
                <MetricCard
                    description="Input + Output tokens processed"
                    icon={<Cpu className="text-muted-foreground size-4" />}
                    loading={usageTotalLoading}
                    title="Total Tokens"
                    value={formatTokenCount(totalTokens)}
                />
                <MetricCard
                    description="Total LLM spending across all providers"
                    icon={<CircleDollarSign className="text-muted-foreground size-4" />}
                    loading={usageTotalLoading}
                    title="Total Cost"
                    value={formatCost(totalCost)}
                />
            </div>

            <Card>
                <CardHeader>
                    <CardTitle>Usage by Provider</CardTitle>
                    <CardDescription>LLM token usage and costs grouped by provider</CardDescription>
                </CardHeader>
                <CardContent>
                    {usageByProviderLoading ? <LoadingTable /> : <UsageStatsTable rows={providerRows} />}
                </CardContent>
            </Card>

            <Card>
                <CardHeader>
                    <CardTitle>Usage by Model</CardTitle>
                    <CardDescription>LLM token usage and costs grouped by model</CardDescription>
                </CardHeader>
                <CardContent>
                    {usageByModelLoading ? <LoadingTable /> : <UsageStatsTable rows={modelRows} />}
                </CardContent>
            </Card>

            <Card>
                <CardHeader>
                    <CardTitle>Usage by Agent Type</CardTitle>
                    <CardDescription>LLM token usage and costs grouped by agent type</CardDescription>
                </CardHeader>
                <CardContent>
                    {usageByAgentTypeLoading ? <LoadingTable /> : <UsageStatsTable rows={agentTypeRows} />}
                </CardContent>
            </Card>

            <Card>
                <CardHeader>
                    <CardTitle>Tool Calls by Function</CardTitle>
                    <CardDescription>Execution statistics for each tool function</CardDescription>
                </CardHeader>
                <CardContent>
                    {toolcallsByFunctionLoading ? (
                        <LoadingTable />
                    ) : (
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead className="whitespace-nowrap">Function</TableHead>
                                    <TableHead className="whitespace-nowrap">Type</TableHead>
                                    <TableHead className="whitespace-nowrap text-right">Count</TableHead>
                                    <TableHead className="whitespace-nowrap text-right">Total Duration</TableHead>
                                    <TableHead className="whitespace-nowrap text-right">Avg Duration</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {toolcallsByFunction.map((item) => (
                                    <TableRow key={item.functionName}>
                                        <TableCell className="font-medium">{item.functionName}</TableCell>
                                        <TableCell>
                                            <Badge variant={item.isAgent ? 'secondary' : 'outline'}>
                                                {item.isAgent ? 'Agent' : 'Tool'}
                                            </Badge>
                                        </TableCell>
                                        <TableCell className="text-right">{formatNumber(item.totalCount)}</TableCell>
                                        <TableCell className="text-right">
                                            {formatDuration(item.totalDurationSeconds)}
                                        </TableCell>
                                        <TableCell className="text-right">
                                            {formatDuration(item.avgDurationSeconds)}
                                        </TableCell>
                                    </TableRow>
                                ))}
                            </TableBody>
                        </Table>
                    )}
                </CardContent>
            </Card>
        </div>
    );
};
