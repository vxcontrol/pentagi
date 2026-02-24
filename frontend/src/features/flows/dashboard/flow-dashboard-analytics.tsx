import {
    Bug,
    FileText,
    Globe,
    KeyRound,
    Network,
    Server,
    Shield,
    ShieldAlert,
    Sparkles,
    Wrench,
    Zap,
} from 'lucide-react';

import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import FlowGraph from '@/features/flows/dashboard/flow-graph';

import type { GraphData } from './use-flow-dashboard';
import type { useFlowDashboard } from './use-flow-dashboard';

type FlowDashboardAnalyticsProps = Pick<
    FlowDashboardData,
    | 'accessChainGraph'
    | 'accessDetails'
    | 'allCves'
    | 'artifacts'
    | 'attackSurface'
    | 'credentialsStatus'
    | 'dashboard'
    | 'exploitAttempts'
    | 'fullAttackChain'
    | 'hostsWithServices'
    | 'infrastructureGraph'
    | 'mainAttackChain'
    | 'openPorts'
    | 'shortestPathGraph'
    | 'toolEffectiveness'
    | 'toolUsage'
    | 'vulnerabilitySeverity'
>;

type FlowDashboardData = ReturnType<typeof useFlowDashboard>;

const METRIC_CARDS = [
    { icon: Server, key: 'hosts' as const, label: 'Hosts' },
    { icon: Network, key: 'ports' as const, label: 'Ports' },
    { icon: Globe, key: 'services' as const, label: 'Services' },
    { icon: Shield, key: 'vulnerabilities' as const, label: 'Vulnerabilities' },
    { icon: KeyRound, key: 'accounts' as const, label: 'Accounts' },
    { icon: ShieldAlert, key: 'validAccess' as const, label: 'Valid Access' },
];

const SEVERITY_COLORS: Record<string, string> = {
    CRITICAL: 'bg-red-500/15 text-red-600 border-red-500/30',
    CVE: 'bg-orange-500/15 text-orange-600 border-orange-500/30',
    EXPLOITED: 'bg-red-500/15 text-red-600 border-red-500/30',
    HIGH: 'bg-orange-500/15 text-orange-600 border-orange-500/30',
    'LOW/INFO': 'bg-slate-500/15 text-slate-600 border-slate-500/30',
    MEDIUM: 'bg-amber-500/15 text-amber-600 border-amber-500/30',
};

const MetricCard = ({
    icon: Icon,
    label,
    value,
}: {
    icon: React.ComponentType<{ className?: string }>;
    label: string;
    value: number;
}) => (
    <Card>
        <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-muted-foreground text-sm font-medium">{label}</CardTitle>
            <Icon className="text-muted-foreground size-4 shrink-0" />
        </CardHeader>
        <CardContent>
            <div className="text-2xl font-bold">{value}</div>
        </CardContent>
    </Card>
);

const hasGraphData = (data: GraphData | null): data is GraphData => !!data?.data?.length;

export const FlowDashboardAnalytics = ({
    accessChainGraph,
    accessDetails,
    allCves,
    artifacts,
    attackSurface,
    credentialsStatus,
    dashboard,
    exploitAttempts,
    fullAttackChain,
    hostsWithServices,
    infrastructureGraph,
    mainAttackChain,
    openPorts,
    shortestPathGraph,
    toolEffectiveness,
    toolUsage,
    vulnerabilitySeverity,
}: FlowDashboardAnalyticsProps) => {
    const graphViews = [
        { data: mainAttackChain, label: 'Main Attack Chain', value: 'overview' },
        { data: fullAttackChain, label: 'Full Attack Chain', value: 'full' },
        { data: infrastructureGraph, label: 'Infrastructure', value: 'infrastructure' },
        { data: accessChainGraph, label: 'Access Chain', value: 'access' },
        { data: shortestPathGraph, label: 'Shortest Path', value: 'shortest-path' },
    ].filter((view) => hasGraphData(view.data));

    const defaultGraphTab = graphViews[0]?.value;

    return (
        <div className="space-y-6">
            {!!graphViews.length && defaultGraphTab && (
                <Tabs defaultValue={defaultGraphTab}>
                    <Card>
                        <CardHeader className="items-start">
                            <TabsList className="bg-background flex-wrap">
                                {graphViews.map((view) => (
                                    <TabsTrigger
                                        className="data-[state=active]:bg-card"
                                        key={view.value}
                                        value={view.value}
                                    >
                                        {view.label}
                                    </TabsTrigger>
                                ))}
                            </TabsList>
                        </CardHeader>
                        <CardContent>
                            {graphViews.map((view) => (
                                <TabsContent
                                    className="mt-0"
                                    key={view.value}
                                    value={view.value}
                                >
                                    <FlowGraph data={view.data!} />
                                </TabsContent>
                            ))}
                        </CardContent>
                    </Card>
                </Tabs>
            )}

            {dashboard && (
                <div className="grid grid-cols-2 gap-4 md:grid-cols-3">
                    {METRIC_CARDS.map(({ icon, key, label }) => (
                        <MetricCard
                            icon={icon}
                            key={key}
                            label={label}
                            value={dashboard[key]}
                        />
                    ))}
                </div>
            )}

            {!!attackSurface.length && (
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2 text-base">
                            <Sparkles className="size-4" />
                            Attack Surface Overview
                        </CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="flex flex-wrap gap-3">
                            {attackSurface.map((record) => (
                                <div
                                    className="bg-muted/50 flex items-center gap-2 rounded-lg border px-3 py-2"
                                    key={record.entityType}
                                >
                                    <span className="text-muted-foreground text-sm">{record.entityType}</span>
                                    <Badge variant="secondary">{record.count}</Badge>
                                </div>
                            ))}
                        </div>
                    </CardContent>
                </Card>
            )}

            {!!credentialsStatus.length && (
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2 text-base">
                            <KeyRound className="size-4" />
                            Credentials Status
                        </CardTitle>
                    </CardHeader>
                    <CardContent>
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead>Status</TableHead>
                                    <TableHead className="text-right">Count</TableHead>
                                    <TableHead>Examples</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {credentialsStatus.map((record) => (
                                    <TableRow key={record.status}>
                                        <TableCell>
                                            <Badge
                                                className={
                                                    record.status === 'COMPROMISED'
                                                        ? 'border-red-500/30 bg-red-500/15 text-red-600'
                                                        : 'border-blue-500/30 bg-blue-500/15 text-blue-600'
                                                }
                                            >
                                                {record.status}
                                            </Badge>
                                        </TableCell>
                                        <TableCell className="text-right font-bold">{record.count}</TableCell>
                                        <TableCell>
                                            <div className="flex flex-wrap gap-1">
                                                {record.examples.filter(Boolean).map((example) => (
                                                    <Badge
                                                        key={example}
                                                        variant="outline"
                                                    >
                                                        {example}
                                                    </Badge>
                                                ))}
                                            </div>
                                        </TableCell>
                                    </TableRow>
                                ))}
                            </TableBody>
                        </Table>
                    </CardContent>
                </Card>
            )}

            {!!accessDetails.length && (
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2 text-base">
                            <KeyRound className="size-4" />
                            Valid Access
                        </CardTitle>
                    </CardHeader>
                    <CardContent>
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead>Access</TableHead>
                                    <TableHead>Account</TableHead>
                                    <TableHead>Host</TableHead>
                                    <TableHead>Service</TableHead>
                                    <TableHead>Summary</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {accessDetails.map((record, index) => (
                                    <TableRow key={index}>
                                        <TableCell className="font-medium">{record.access ?? '—'}</TableCell>
                                        <TableCell className="font-mono text-sm">{record.account ?? '—'}</TableCell>
                                        <TableCell className="font-mono text-sm">{record.host ?? '—'}</TableCell>
                                        <TableCell>{record.service ?? '—'}</TableCell>
                                        <TableCell className="max-w-xs truncate text-sm">
                                            {record.summary ?? '—'}
                                        </TableCell>
                                    </TableRow>
                                ))}
                            </TableBody>
                        </Table>
                    </CardContent>
                </Card>
            )}

            {!!hostsWithServices.length && (
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2 text-base">
                            <Globe className="size-4" />
                            Infrastructure Map
                        </CardTitle>
                    </CardHeader>
                    <CardContent>
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead>Host</TableHead>
                                    <TableHead>Ports</TableHead>
                                    <TableHead>Services</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {hostsWithServices.map((record, index) => (
                                    <TableRow key={index}>
                                        <TableCell className="font-mono font-medium">{record.host ?? '—'}</TableCell>
                                        <TableCell>
                                            <div className="flex flex-wrap gap-1">
                                                {record.ports.filter(Boolean).map((port) => (
                                                    <Badge
                                                        key={port}
                                                        variant="outline"
                                                    >
                                                        {port}
                                                    </Badge>
                                                ))}
                                                {!record.ports.filter(Boolean).length && (
                                                    <span className="text-muted-foreground">—</span>
                                                )}
                                            </div>
                                        </TableCell>
                                        <TableCell>
                                            <div className="flex flex-wrap gap-1">
                                                {record.services.filter(Boolean).map((service) => (
                                                    <Badge
                                                        key={service}
                                                        variant="secondary"
                                                    >
                                                        {service}
                                                    </Badge>
                                                ))}
                                                {!record.services.filter(Boolean).length && (
                                                    <span className="text-muted-foreground">—</span>
                                                )}
                                            </div>
                                        </TableCell>
                                    </TableRow>
                                ))}
                            </TableBody>
                        </Table>
                    </CardContent>
                </Card>
            )}

            {!!openPorts.length && (
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2 text-base">
                            <Network className="size-4" />
                            Open Ports
                        </CardTitle>
                    </CardHeader>
                    <CardContent>
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead>Port</TableHead>
                                    <TableHead>Service</TableHead>
                                    <TableHead>Host</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {openPorts.map((record, index) => (
                                    <TableRow key={index}>
                                        <TableCell className="font-mono font-medium">{record.port ?? '—'}</TableCell>
                                        <TableCell>
                                            {record.service ? (
                                                <Badge variant="secondary">{record.service}</Badge>
                                            ) : (
                                                <span className="text-muted-foreground">—</span>
                                            )}
                                        </TableCell>
                                        <TableCell className="font-mono text-sm">{record.host ?? '—'}</TableCell>
                                    </TableRow>
                                ))}
                            </TableBody>
                        </Table>
                    </CardContent>
                </Card>
            )}

            {!!vulnerabilitySeverity.length && (
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2 text-base">
                            <ShieldAlert className="size-4" />
                            Vulnerability Breakdown
                        </CardTitle>
                    </CardHeader>
                    <CardContent>
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead>Category</TableHead>
                                    <TableHead className="text-right">Count</TableHead>
                                    <TableHead>Examples</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {vulnerabilitySeverity.map((record) => (
                                    <TableRow key={record.category}>
                                        <TableCell>
                                            <Badge className={SEVERITY_COLORS[record.category] ?? ''}>
                                                {record.category}
                                            </Badge>
                                        </TableCell>
                                        <TableCell className="text-right font-bold">{record.count}</TableCell>
                                        <TableCell>
                                            <div className="flex flex-wrap gap-1">
                                                {record.examples.filter(Boolean).map((example) => (
                                                    <Badge
                                                        className="max-w-[200px] truncate"
                                                        key={example}
                                                        variant="outline"
                                                    >
                                                        {example}
                                                    </Badge>
                                                ))}
                                            </div>
                                        </TableCell>
                                    </TableRow>
                                ))}
                            </TableBody>
                        </Table>
                    </CardContent>
                </Card>
            )}

            {!!allCves.length && (
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2 text-base">
                            <Shield className="size-4" />
                            Detected CVEs
                        </CardTitle>
                    </CardHeader>
                    <CardContent>
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead>CVE</TableHead>
                                    <TableHead>Found On</TableHead>
                                    <TableHead>Source</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {allCves.map((record, index) => (
                                    <TableRow key={index}>
                                        <TableCell className="font-mono text-sm font-medium">
                                            {record.cve ?? '—'}
                                        </TableCell>
                                        <TableCell>
                                            {record.foundOn ? (
                                                <Badge variant="outline">{record.foundOn}</Badge>
                                            ) : (
                                                <span className="text-muted-foreground">—</span>
                                            )}
                                        </TableCell>
                                        <TableCell className="font-mono text-sm">{record.source ?? '—'}</TableCell>
                                    </TableRow>
                                ))}
                            </TableBody>
                        </Table>
                    </CardContent>
                </Card>
            )}

            {!!exploitAttempts.length && (
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2 text-base">
                            <Bug className="size-4" />
                            Exploit Attempts
                        </CardTitle>
                    </CardHeader>
                    <CardContent>
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead>Vulnerability</TableHead>
                                    <TableHead className="text-right">Attempts</TableHead>
                                    <TableHead>Status</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {exploitAttempts.map((record, index) => (
                                    <TableRow key={index}>
                                        <TableCell className="font-mono text-sm font-medium">
                                            {record.vulnerability ?? '—'}
                                        </TableCell>
                                        <TableCell className="text-right">
                                            <Badge variant="secondary">{record.attemptCount}</Badge>
                                        </TableCell>
                                        <TableCell>
                                            <Badge className={SEVERITY_COLORS[record.status] ?? ''}>
                                                {record.status}
                                            </Badge>
                                        </TableCell>
                                    </TableRow>
                                ))}
                            </TableBody>
                        </Table>
                    </CardContent>
                </Card>
            )}

            {!!toolUsage.length && (
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2 text-base">
                            <Wrench className="size-4" />
                            Tool Usage
                        </CardTitle>
                    </CardHeader>
                    <CardContent>
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead>Tool</TableHead>
                                    <TableHead className="text-right">Executions</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {toolUsage.map((record, index) => (
                                    <TableRow key={index}>
                                        <TableCell className="font-medium">{record.tool ?? '—'}</TableCell>
                                        <TableCell className="text-right">
                                            <Badge variant="secondary">{record.executions}</Badge>
                                        </TableCell>
                                    </TableRow>
                                ))}
                            </TableBody>
                        </Table>
                    </CardContent>
                </Card>
            )}

            {!!toolEffectiveness.length && (
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2 text-base">
                            <Zap className="size-4" />
                            Tool Effectiveness
                        </CardTitle>
                    </CardHeader>
                    <CardContent>
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead>Tool</TableHead>
                                    <TableHead className="text-right">Executions</TableHead>
                                    <TableHead className="text-right">Discoveries</TableHead>
                                    <TableHead>Discovery Types</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {toolEffectiveness.map((record, index) => (
                                    <TableRow key={index}>
                                        <TableCell className="font-medium">{record.tool ?? '—'}</TableCell>
                                        <TableCell className="text-right">
                                            <Badge variant="secondary">{record.executions}</Badge>
                                        </TableCell>
                                        <TableCell className="text-right">
                                            <Badge variant="secondary">{record.discoveries}</Badge>
                                        </TableCell>
                                        <TableCell>
                                            <div className="flex flex-wrap gap-1">
                                                {record.discoveryTypes.filter(Boolean).map((discoveryType) => (
                                                    <Badge
                                                        key={discoveryType}
                                                        variant="outline"
                                                    >
                                                        {discoveryType}
                                                    </Badge>
                                                ))}
                                            </div>
                                        </TableCell>
                                    </TableRow>
                                ))}
                            </TableBody>
                        </Table>
                    </CardContent>
                </Card>
            )}

            {!!artifacts.length && (
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2 text-base">
                            <FileText className="size-4" />
                            Artifacts
                        </CardTitle>
                    </CardHeader>
                    <CardContent>
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead>Artifact</TableHead>
                                    <TableHead>Produced By</TableHead>
                                    <TableHead>Summary</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {artifacts.map((record, index) => (
                                    <TableRow key={index}>
                                        <TableCell className="font-medium">{record.artifact ?? '—'}</TableCell>
                                        <TableCell className="font-mono text-sm">{record.producedBy ?? '—'}</TableCell>
                                        <TableCell className="max-w-xs truncate text-sm">
                                            {record.summary ?? '—'}
                                        </TableCell>
                                    </TableRow>
                                ))}
                            </TableBody>
                        </Table>
                    </CardContent>
                </Card>
            )}
        </div>
    );
};
