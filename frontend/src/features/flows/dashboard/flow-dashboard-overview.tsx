import { Activity, CircleDollarSign, Cpu, GitFork, Loader2 } from 'lucide-react';
import { useMemo } from 'react';

import type { UsageStatsFragmentFragment } from '@/graphql/types';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import {
    useFlowStatsByFlowQuery,
    useToolcallsStatsByFlowQuery,
    useToolcallsStatsByFunctionForFlowQuery,
    useUsageStatsByAgentTypeForFlowQuery,
    useUsageStatsByFlowQuery,
} from '@/graphql/types';
import { formatCost, formatDuration, formatNumber, formatTokenCount } from '@/pages/dashboard/format-utils';

const StatCard = ({
    description,
    icon,
    loading,
    title,
    value,
}: {
    description: string;
    icon: React.ReactNode;
    loading: boolean;
    title: string;
    value: string;
}) => (
    <Card>
        <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">{title}</CardTitle>
            {icon}
        </CardHeader>
        <CardContent>
            {loading ? <Skeleton className="h-8 w-24" /> : <div className="text-2xl font-bold">{value}</div>}
            <p className="text-muted-foreground text-xs">{description}</p>
        </CardContent>
    </Card>
);

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

const LoadingTable = () => (
    <div className="flex items-center justify-center py-8">
        <Loader2 className="text-muted-foreground size-6 animate-spin" />
    </div>
);

export const FlowDashboardOverview = ({ flowId }: { flowId: string }) => {
    const { data: usageData, loading: usageLoading } = useUsageStatsByFlowQuery({
        variables: { flowId },
    });
    const { data: usageByAgentData, loading: usageByAgentLoading } = useUsageStatsByAgentTypeForFlowQuery({
        variables: { flowId },
    });
    const { data: toolcallsData, loading: toolcallsLoading } = useToolcallsStatsByFlowQuery({
        variables: { flowId },
    });
    const { data: toolcallsByFunctionData, loading: toolcallsByFunctionLoading } =
        useToolcallsStatsByFunctionForFlowQuery({
            variables: { flowId },
        });
    const { data: flowStatsData, loading: flowStatsLoading } = useFlowStatsByFlowQuery({
        variables: { flowId },
    });

    const usage = usageData?.usageStatsByFlow;
    const toolcalls = toolcallsData?.toolcallsStatsByFlow;
    const flowStats = flowStatsData?.flowStatsByFlow;

    const totalCost = usage ? usage.totalUsageCostIn + usage.totalUsageCostOut : 0;
    const totalTokens = usage ? usage.totalUsageIn + usage.totalUsageOut : 0;

    const agentTypeRows = useMemo(() => {
        const seen = new Set<string>();

        return (usageByAgentData?.usageStatsByAgentTypeForFlow ?? [])
            .filter((item) => {
                if (seen.has(item.agentType)) {
                    return false;
                }

                seen.add(item.agentType);

                return true;
            })
            .map((item) => ({
                label: item.agentType,
                stats: item.stats,
            }));
    }, [usageByAgentData]);

    const toolcallsByFunction = useMemo(() => {
        const seen = new Set<string>();

        return [...(toolcallsByFunctionData?.toolcallsStatsByFunctionForFlow ?? [])]
            .filter((item) => {
                if (seen.has(item.functionName)) {
                    return false;
                }

                seen.add(item.functionName);

                return true;
            })
            .sort((a, b) => b.totalCount - a.totalCount);
    }, [toolcallsByFunctionData]);

    const anyLoading = usageLoading || toolcallsLoading || flowStatsLoading;

    return (
        <div className="space-y-6">
            <div className="grid grid-cols-2 gap-4 md:grid-cols-4">
                <StatCard
                    description="LLM spending for this flow"
                    icon={<CircleDollarSign className="text-muted-foreground size-4" />}
                    loading={anyLoading}
                    title="Cost"
                    value={formatCost(totalCost)}
                />
                <StatCard
                    description="Input + Output tokens"
                    icon={<Cpu className="text-muted-foreground size-4" />}
                    loading={anyLoading}
                    title="Tokens"
                    value={formatTokenCount(totalTokens)}
                />
                <StatCard
                    description={`Duration: ${toolcalls ? formatDuration(toolcalls.totalDurationSeconds) : '—'}`}
                    icon={<Activity className="text-muted-foreground size-4" />}
                    loading={anyLoading}
                    title="Tool Calls"
                    value={toolcalls ? formatNumber(toolcalls.totalCount) : '0'}
                />
                <StatCard
                    description={`Subtasks: ${flowStats?.totalSubtasksCount ?? 0} · Assistants: ${flowStats?.totalAssistantsCount ?? 0}`}
                    icon={<GitFork className="text-muted-foreground size-4" />}
                    loading={anyLoading}
                    title="Tasks"
                    value={flowStats ? formatNumber(flowStats.totalTasksCount) : '0'}
                />
            </div>

            {!!agentTypeRows.length && (
                <Card>
                    <CardHeader>
                        <CardTitle>Usage by Agent Type</CardTitle>
                        <CardDescription>LLM token usage and costs per agent type in this flow</CardDescription>
                    </CardHeader>
                    <CardContent>
                        {usageByAgentLoading ? (
                            <LoadingTable />
                        ) : (
                            <Table>
                                <TableHeader>
                                    <TableRow>
                                        <TableHead>Agent Type</TableHead>
                                        <TableHead className="text-right">Tokens In</TableHead>
                                        <TableHead className="text-right">Tokens Out</TableHead>
                                        <TableHead className="text-right">Cache In</TableHead>
                                        <TableHead className="text-right">Cache Out</TableHead>
                                        <TableHead className="text-right">Cost In</TableHead>
                                        <TableHead className="text-right">Cost Out</TableHead>
                                        <TableHead className="text-right">Total Cost</TableHead>
                                    </TableRow>
                                </TableHeader>
                                <TableBody>
                                    {agentTypeRows.map((row) => (
                                        <UsageStatsRow
                                            key={row.label}
                                            label={row.label}
                                            stats={row.stats}
                                        />
                                    ))}
                                </TableBody>
                            </Table>
                        )}
                    </CardContent>
                </Card>
            )}

            {!!toolcallsByFunction.length && (
                <Card>
                    <CardHeader>
                        <CardTitle>Tool Calls by Function</CardTitle>
                        <CardDescription>Execution statistics per tool function in this flow</CardDescription>
                    </CardHeader>
                    <CardContent>
                        {toolcallsByFunctionLoading ? (
                            <LoadingTable />
                        ) : (
                            <Table>
                                <TableHeader>
                                    <TableRow>
                                        <TableHead>Function</TableHead>
                                        <TableHead>Type</TableHead>
                                        <TableHead className="text-right">Count</TableHead>
                                        <TableHead className="text-right">Total Duration</TableHead>
                                        <TableHead className="text-right">Avg Duration</TableHead>
                                    </TableRow>
                                </TableHeader>
                                <TableBody>
                                    {toolcallsByFunction.map((item) => (
                                        <TableRow key={item.functionName}>
                                            <TableCell className="font-medium">{item.functionName}</TableCell>
                                            <TableCell>
                                                <span
                                                    className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${
                                                        item.isAgent
                                                            ? 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200'
                                                            : 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200'
                                                    }`}
                                                >
                                                    {item.isAgent ? 'Agent' : 'Tool'}
                                                </span>
                                            </TableCell>
                                            <TableCell className="text-right">
                                                {formatNumber(item.totalCount)}
                                            </TableCell>
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
            )}
        </div>
    );
};
