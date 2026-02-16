import {
    AlertTriangle,
    Bug,
    FileText,
    Globe,
    KeyRound,
    Loader2,
    Network,
    RefreshCw,
    Server,
    Shield,
    ShieldAlert,
    ShieldCheck,
    Sparkles,
    Wrench,
    Zap,
} from 'lucide-react';

import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import FlowGraph from '@/features/flows/dashboard/flow-graph';
import { useFlow } from '@/providers/flow-provider';

import type { GraphData, PentestStatus } from './use-flow-dashboard';

import { useFlowDashboard } from './use-flow-dashboard';

// --- Status config ---

const STATUS_CONFIG = {
    COMPROMISED: {
        className: 'bg-red-500/15 text-red-600 border-red-500/30',
        icon: AlertTriangle,
        label: 'Compromised',
    },
    SECURE: {
        className: 'bg-emerald-500/15 text-emerald-600 border-emerald-500/30',
        icon: ShieldCheck,
        label: 'Secure',
    },
    VULNERABLE: {
        className: 'bg-amber-500/15 text-amber-600 border-amber-500/30',
        icon: ShieldAlert,
        label: 'Vulnerable',
    },
} as const;

// --- Summary metric cards ---

const METRIC_CARDS = [
    { icon: Server, key: 'hosts' as const, label: 'Hosts' },
    { icon: Network, key: 'ports' as const, label: 'Ports' },
    { icon: Globe, key: 'services' as const, label: 'Services' },
    { icon: Shield, key: 'vulnerabilities' as const, label: 'Vulnerabilities' },
    { icon: KeyRound, key: 'accounts' as const, label: 'Accounts' },
    { icon: ShieldAlert, key: 'validAccess' as const, label: 'Valid Access' },
];

// --- Severity badge colors ---

const SEVERITY_COLORS: Record<string, string> = {
    CRITICAL: 'bg-red-500/15 text-red-600 border-red-500/30',
    CVE: 'bg-orange-500/15 text-orange-600 border-orange-500/30',
    EXPLOITED: 'bg-red-500/15 text-red-600 border-red-500/30',
    HIGH: 'bg-orange-500/15 text-orange-600 border-orange-500/30',
    'LOW/INFO': 'bg-slate-500/15 text-slate-600 border-slate-500/30',
    MEDIUM: 'bg-amber-500/15 text-amber-600 border-amber-500/30',
};

// --- Sub-components ---

const DashboardSkeleton = () => (
    <div className="space-y-6">
        {/* Header: title + badge + refresh button */}
        <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
                <Skeleton className="h-7 w-40" />
                <Skeleton className="h-6 w-24 rounded-full" />
            </div>
            <Skeleton className="h-9 w-24" />
        </div>

        {/* Graph card with tabs */}
        <Card>
            <CardHeader className="items-start">
                <div className="flex flex-wrap gap-2">
                    {Array.from({ length: 4 }).map((_, index) => (
                        <Skeleton
                            className="h-9 w-28 rounded-md"
                            key={index}
                        />
                    ))}
                </div>
            </CardHeader>
            <CardContent>
                <Skeleton className="h-[400px] w-full rounded-md" />
            </CardContent>
        </Card>

        {/* Metric summary cards (same grid as MetricCard) */}
        <div className="grid grid-cols-2 gap-4 md:grid-cols-3">
            {Array.from({ length: 6 }).map((_, index) => (
                <Card key={index}>
                    <CardHeader className="flex flex-row items-center justify-between pb-2">
                        <Skeleton className="h-4 w-24" />
                        <Skeleton className="size-4 shrink-0 rounded" />
                    </CardHeader>
                    <CardContent>
                        <Skeleton className="h-8 w-12" />
                    </CardContent>
                </Card>
            ))}
        </div>

        {/* Placeholder for first content card (e.g. Attack Surface / Credentials) */}
        <Card>
            <CardHeader>
                <Skeleton className="h-5 w-48" />
            </CardHeader>
            <CardContent>
                <div className="space-y-3">
                    {Array.from({ length: 4 }).map((_, index) => (
                        <Skeleton
                            className="h-10 w-full"
                            key={index}
                        />
                    ))}
                </div>
            </CardContent>
        </Card>
    </div>
);

const StatusBadge = ({ status }: { status: PentestStatus }) => {
    const config = STATUS_CONFIG[status];
    const Icon = config.icon;

    return (
        <Badge className={`gap-1.5 px-3 py-1.5 text-sm ${config.className}`}>
            <Icon className="size-4" />
            {config.label}
        </Badge>
    );
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

// --- Main component ---

const hasGraphData = (data: GraphData | null): data is GraphData => !!data?.data?.length;

const FlowDashboard = () => {
    const { flowId } = useFlow();
    const groupId = flowId ? `flow-${flowId}` : null;
    const {
        accessChainGraph,
        accessDetails,
        allCves,
        artifacts,
        attackSurface,
        credentialsStatus,
        dashboard,
        error,
        exploitAttempts,
        fullAttackChain,
        hostsWithServices,
        infrastructureGraph,
        isLoading,
        mainAttackChain,
        openPorts,
        refetch,
        shortestPathGraph,
        toolEffectiveness,
        toolUsage,
        vulnerabilitySeverity,
    } = useFlowDashboard(groupId);

    const graphViews = [
        { data: mainAttackChain, label: 'Main Attack Chain', value: 'overview' },
        { data: fullAttackChain, label: 'Full Attack Chain', value: 'full' },
        { data: infrastructureGraph, label: 'Infrastructure', value: 'infrastructure' },
        { data: accessChainGraph, label: 'Access Chain', value: 'access' },
        { data: shortestPathGraph, label: 'Shortest Path', value: 'shortest-path' },
    ].filter((view) => hasGraphData(view.data));

    const defaultGraphTab = graphViews[0]?.value;

    if (!flowId) {
        return (
            <div className="text-muted-foreground flex items-center justify-center py-12">
                Select a flow to view the dashboard
            </div>
        );
    }

    if (isLoading && !dashboard) {
        return <DashboardSkeleton />;
    }

    if (error && !dashboard) {
        return (
            <div className="flex flex-col items-center justify-center gap-4 py-12">
                <p className="text-destructive">{error}</p>
                <Button
                    onClick={refetch}
                    size="sm"
                    variant="outline"
                >
                    <RefreshCw className="mr-2 size-4" />
                    Retry
                </Button>
            </div>
        );
    }

    return (
        <div className="space-y-6">
            {/* Header with status and refresh */}
            <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                    <h2 className="text-lg font-semibold">Flow Dashboard</h2>
                    {dashboard?.status && <StatusBadge status={dashboard.status} />}
                </div>
                <Button
                    disabled={isLoading}
                    onClick={refetch}
                    size="sm"
                    variant="ghost"
                >
                    <RefreshCw className={`mr-2 size-4 ${isLoading ? 'animate-spin' : ''}`} />
                    Refresh
                </Button>
            </div>

            {/* Graph views */}
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

            {/* Metric summary cards */}
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

            {/* Attack Surface Overview (Q1) */}
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

            {/* Credentials Status (Q2a) */}
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

            {/* Valid access (Q2b) */}
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

            {/* Discovered services (Q3a) */}
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

            {/* Open ports with services (Q3b) */}
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

            {/* Vulnerability severity breakdown (Q4a) */}
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

            {/* Detected vulnerabilities / CVEs (Q4b) */}
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

            {/* Exploitable vulnerabilities (Q4c) */}
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

            {/* Executed tools (Q5a) */}
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

            {/* Tool effectiveness (Q5b) */}
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

            {/* Artifacts produced (Q5c) */}
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

            {/* Empty state when no data */}
            {!dashboard && !isLoading && (
                <div className="flex flex-col items-center justify-center gap-2 py-12">
                    <Loader2 className="text-muted-foreground size-8" />
                    <p className="text-muted-foreground">No dashboard data available for this flow</p>
                </div>
            )}
        </div>
    );
};

export default FlowDashboard;
