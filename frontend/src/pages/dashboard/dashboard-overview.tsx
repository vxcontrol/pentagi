import { useQuery } from '@apollo/client/react';
import { Activity, CircleDollarSign, Cpu, GitFork } from 'lucide-react';

import type { UsageStatsFragmentFragment } from '@/graphql/types';

import { MetricCard } from '@/components/dashboard';
import { DashboardError } from '@/components/dashboard/dashboard-error';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Spinner } from '@/components/ui/spinner';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import {
    FlowsStatsTotalDocument,
    ToolcallsStatsByFunctionDocument,
    ToolcallsStatsTotalDocument,
    UsageStatsByAgentTypeDocument,
    UsageStatsByModelDocument,
    UsageStatsByProviderDocument,
    UsageStatsTotalDocument,
} from '@/graphql/types';
import { useI18n } from '@/hooks/use-i18n';
import { formatCost, formatDuration, formatNumber, formatTokenCount } from '@/lib/utils/format';

export function DashboardOverview() {
    const { t } = useI18n();
    const {
        data: usageTotalData,
        error: usageTotalError,
        loading: usageTotalLoading,
    } = useQuery(UsageStatsTotalDocument);
    const {
        data: usageByProviderData,
        error: usageByProviderError,
        loading: usageByProviderLoading,
    } = useQuery(UsageStatsByProviderDocument);
    const {
        data: usageByModelData,
        error: usageByModelError,
        loading: usageByModelLoading,
    } = useQuery(UsageStatsByModelDocument);
    const {
        data: usageByAgentTypeData,
        error: usageByAgentTypeError,
        loading: usageByAgentTypeLoading,
    } = useQuery(UsageStatsByAgentTypeDocument);
    const {
        data: toolcallsTotalData,
        error: toolcallsTotalError,
        loading: toolcallsTotalLoading,
    } = useQuery(ToolcallsStatsTotalDocument);
    const {
        data: toolcallsByFunctionData,
        error: toolcallsByFunctionError,
        loading: toolcallsByFunctionLoading,
    } = useQuery(ToolcallsStatsByFunctionDocument);
    const {
        data: flowsTotalData,
        error: flowsTotalError,
        loading: flowsTotalLoading,
    } = useQuery(FlowsStatsTotalDocument);

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
                    description={t('dashboard.flowsSummary', {
                        assistants: flowsTotal?.totalAssistantsCount ?? 0,
                        subtasks: flowsTotal?.totalSubtasksCount ?? 0,
                        tasks: flowsTotal?.totalTasksCount ?? 0,
                    })}
                    error={!!flowsTotalError}
                    icon={<GitFork className="text-muted-foreground size-4" />}
                    loading={flowsTotalLoading}
                    title={t('dashboard.totalFlows')}
                    value={flowsTotal ? formatNumber(flowsTotal.totalFlowsCount) : '0'}
                />
                <MetricCard
                    description={t('dashboard.totalDuration', {
                        duration: toolcallsTotal ? formatDuration(toolcallsTotal.totalDurationSeconds) : '—',
                    })}
                    error={!!toolcallsTotalError}
                    icon={<Activity className="text-muted-foreground size-4" />}
                    loading={toolcallsTotalLoading}
                    title={t('dashboard.toolCalls')}
                    value={toolcallsTotal ? formatNumber(toolcallsTotal.totalCount) : '0'}
                />
                <MetricCard
                    description={t('dashboard.inputOutputTokens')}
                    error={!!usageTotalError}
                    icon={<Cpu className="text-muted-foreground size-4" />}
                    loading={usageTotalLoading}
                    title={t('dashboard.totalTokens')}
                    value={formatTokenCount(totalTokens)}
                />
                <MetricCard
                    description={t('dashboard.totalLlmSpending')}
                    error={!!usageTotalError}
                    icon={<CircleDollarSign className="text-muted-foreground size-4" />}
                    loading={usageTotalLoading}
                    title={t('dashboard.totalCost')}
                    value={formatCost(totalCost)}
                />
            </div>

            <Card>
                <CardHeader>
                    <CardTitle>{t('dashboard.usageByProvider')}</CardTitle>
                    <CardDescription>{t('dashboard.usageByProviderDesc')}</CardDescription>
                </CardHeader>
                <CardContent>
                    {usageByProviderLoading ? (
                        <LoadingTable />
                    ) : usageByProviderError ? (
                        <ErrorTable />
                    ) : (
                        <UsageStatsTable rows={providerRows} />
                    )}
                </CardContent>
            </Card>

            <Card>
                <CardHeader>
                    <CardTitle>{t('dashboard.usageByModel')}</CardTitle>
                    <CardDescription>{t('dashboard.usageByModelDesc')}</CardDescription>
                </CardHeader>
                <CardContent>
                    {usageByModelLoading ? (
                        <LoadingTable />
                    ) : usageByModelError ? (
                        <ErrorTable />
                    ) : (
                        <UsageStatsTable rows={modelRows} />
                    )}
                </CardContent>
            </Card>

            <Card>
                <CardHeader>
                    <CardTitle>{t('dashboard.usageByAgentType')}</CardTitle>
                    <CardDescription>{t('dashboard.usageByAgentTypeDesc')}</CardDescription>
                </CardHeader>
                <CardContent>
                    {usageByAgentTypeLoading ? (
                        <LoadingTable />
                    ) : usageByAgentTypeError ? (
                        <ErrorTable />
                    ) : (
                        <UsageStatsTable rows={agentTypeRows} />
                    )}
                </CardContent>
            </Card>

            <Card>
                <CardHeader>
                    <CardTitle>{t('dashboard.toolCallsByFunction')}</CardTitle>
                    <CardDescription>{t('dashboard.toolCallsByFunctionDesc')}</CardDescription>
                </CardHeader>
                <CardContent>
                    {toolcallsByFunctionLoading ? (
                        <LoadingTable />
                    ) : toolcallsByFunctionError ? (
                        <ErrorTable />
                    ) : (
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead className="whitespace-nowrap">{t('dashboard.function')}</TableHead>
                                    <TableHead className="whitespace-nowrap">{t('dashboard.type')}</TableHead>
                                    <TableHead className="text-right whitespace-nowrap">
                                        {t('dashboard.count')}
                                    </TableHead>
                                    <TableHead className="text-right whitespace-nowrap">
                                        {t('dashboard.totalDurationCol')}
                                    </TableHead>
                                    <TableHead className="text-right whitespace-nowrap">
                                        {t('dashboard.avgDuration')}
                                    </TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {toolcallsByFunction.map((item) => (
                                    <TableRow key={item.functionName}>
                                        <TableCell className="font-medium">{item.functionName}</TableCell>
                                        <TableCell>
                                            <Badge variant={item.isAgent ? 'secondary' : 'outline'}>
                                                {item.isAgent ? t('dashboard.agentBadge') : t('dashboard.toolBadge')}
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
}

function ErrorTable() {
    return <DashboardError className="py-8" />;
}

function LoadingTable() {
    return (
        <div className="flex items-center justify-center py-8">
            <Spinner
                className="text-muted-foreground size-6"
                variant="circle"
            />
        </div>
    );
}

function UsageStatsRow({ label, stats }: { label: string; stats: UsageStatsFragmentFragment }) {
    return (
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
}

function UsageStatsTable({ rows }: { rows: Array<{ label: string; stats: UsageStatsFragmentFragment }> }) {
    const { t } = useI18n();

    return (
        <Table>
            <TableHeader>
                <TableRow>
                    <TableHead className="whitespace-nowrap">{t('dashboard.name')}</TableHead>
                    <TableHead className="text-right whitespace-nowrap">{t('dashboard.tokensIn')}</TableHead>
                    <TableHead className="text-right whitespace-nowrap">{t('dashboard.tokensOut')}</TableHead>
                    <TableHead className="text-right whitespace-nowrap">{t('dashboard.cacheIn')}</TableHead>
                    <TableHead className="text-right whitespace-nowrap">{t('dashboard.cacheOut')}</TableHead>
                    <TableHead className="text-right whitespace-nowrap">{t('dashboard.costIn')}</TableHead>
                    <TableHead className="text-right whitespace-nowrap">{t('dashboard.costOut')}</TableHead>
                    <TableHead className="text-right whitespace-nowrap">{t('dashboard.totalCostCol')}</TableHead>
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
}
