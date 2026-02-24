import { AlertTriangle, Loader2, RefreshCw, ShieldAlert, ShieldCheck } from 'lucide-react';

import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { FlowDashboardAnalytics } from '@/features/flows/dashboard/flow-dashboard-analytics';
import { FlowDashboardOverview } from '@/features/flows/dashboard/flow-dashboard-overview';
import { useFlow } from '@/providers/flow-provider';

import type { PentestStatus } from './use-flow-dashboard';

import { useFlowDashboard } from './use-flow-dashboard';

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

const DashboardSkeleton = () => (
    <div className="space-y-6">
        <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
                <Skeleton className="h-7 w-40" />
                <Skeleton className="h-6 w-24 rounded-full" />
            </div>
            <Skeleton className="h-9 w-24" />
        </div>
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
        <Badge className={`gap-1.5 px-2 py-0.5 text-sm ${config.className}`}>
            <Icon className="size-4" />
            {config.label}
        </Badge>
    );
};

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
            <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                    <h2 className="text-lg font-semibold">Flow Dashboard</h2>
                    {dashboard?.status && <StatusBadge status={dashboard.status} />}
                </div>
            </div>

            <Tabs defaultValue="analytics">
                <TabsList>
                    <TabsTrigger value="analytics">Analytics</TabsTrigger>
                    <TabsTrigger value="overview">Overview</TabsTrigger>
                </TabsList>

                <TabsContent value="analytics">
                    <FlowDashboardAnalytics
                        accessChainGraph={accessChainGraph}
                        accessDetails={accessDetails}
                        allCves={allCves}
                        artifacts={artifacts}
                        attackSurface={attackSurface}
                        credentialsStatus={credentialsStatus}
                        dashboard={dashboard}
                        exploitAttempts={exploitAttempts}
                        fullAttackChain={fullAttackChain}
                        hostsWithServices={hostsWithServices}
                        infrastructureGraph={infrastructureGraph}
                        mainAttackChain={mainAttackChain}
                        openPorts={openPorts}
                        shortestPathGraph={shortestPathGraph}
                        toolEffectiveness={toolEffectiveness}
                        toolUsage={toolUsage}
                        vulnerabilitySeverity={vulnerabilitySeverity}
                    />
                </TabsContent>

                <TabsContent value="overview">
                    <FlowDashboardOverview flowId={flowId} />
                </TabsContent>
            </Tabs>

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
