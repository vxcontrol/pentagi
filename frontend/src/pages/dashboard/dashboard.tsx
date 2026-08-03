import { LayoutDashboard } from 'lucide-react';
import { useState, useTransition } from 'react';

import { AppHeader, AppHeaderContent, AppHeaderTitle } from '@/components/layouts/app/app-header';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { UsageStatsPeriod } from '@/graphql/types';
import { usePageStorageKeys } from '@/hooks/use-page-storage-keys';
import { cn } from '@/lib/utils';
import { DashboardAnalytics } from '@/pages/dashboard/dashboard-analytics';
import { DashboardOverview } from '@/pages/dashboard/dashboard-overview';

const periodOptions: { label: string; value: UsageStatsPeriod }[] = [
    { label: 'Week', value: UsageStatsPeriod.Week },
    { label: 'Month', value: UsageStatsPeriod.Month },
    { label: 'Quarter', value: UsageStatsPeriod.Quarter },
];

const VALID_PERIODS = new Set<string>(Object.values(UsageStatsPeriod));

const loadPeriod = (storageKey: string): UsageStatsPeriod => {
    try {
        const stored = localStorage.getItem(storageKey);

        if (stored && VALID_PERIODS.has(stored)) {
            return stored as UsageStatsPeriod;
        }
    } catch {
        /* localStorage may be unavailable */
    }

    return UsageStatsPeriod.Week;
};

const savePeriod = (storageKey: string, value: UsageStatsPeriod): void => {
    try {
        localStorage.setItem(storageKey, value);
    } catch {
        /* localStorage may be unavailable */
    }
};

function Dashboard() {
    const { period: periodStorageKey } = usePageStorageKeys();
    const [activeTab, setActiveTab] = useState('analytics');
    const [period, setPeriod] = useState<UsageStatsPeriod>(() => loadPeriod(periodStorageKey));
    // Both transitions wrap heavy re-renders: switching activeTab swaps the
    // entire Analytics/Overview subtree (Analytics alone pulls a ~386 kB chunk
    // with four Recharts views), and switching period invalidates the analytics
    // query and re-paints every chart. Without a transition the Tab click feels
    // stuck because React commits the heavy work on the urgent track; with one,
    // the trigger button itself updates instantly and the new content streams
    // in. We surface the pending state as a subtle opacity dim on the
    // transitioning region so the user sees that work is in flight.
    const [isTabPending, startTabTransition] = useTransition();
    const [isPeriodPending, startPeriodTransition] = useTransition();

    const handleTabChange = (value: string) => {
        startTabTransition(() => {
            setActiveTab(value);
        });
    };

    const handlePeriodChange = (value: string) => {
        const next = value as UsageStatsPeriod;
        savePeriod(periodStorageKey, next);
        startPeriodTransition(() => {
            setPeriod(next);
        });
    };

    const isPending = isTabPending || isPeriodPending;

    return (
        <>
            <AppHeader>
                <AppHeaderContent>
                    <AppHeaderTitle icon={<LayoutDashboard className="size-4 shrink-0" />}>Dashboard</AppHeaderTitle>
                </AppHeaderContent>
            </AppHeader>

            <div className="flex flex-col gap-6 p-4">
                <Tabs
                    className="w-full"
                    onValueChange={handleTabChange}
                    value={activeTab}
                >
                    <div className="flex items-center justify-between">
                        <TabsList>
                            <TabsTrigger value="analytics">Analytics</TabsTrigger>
                            <TabsTrigger value="overview">Overview</TabsTrigger>
                        </TabsList>

                        {activeTab === 'analytics' && (
                            <Tabs
                                onValueChange={handlePeriodChange}
                                value={period}
                            >
                                <TabsList>
                                    {periodOptions.map(({ label, value }) => (
                                        <TabsTrigger
                                            aria-label={label}
                                            className="size-7 px-0 sm:size-auto sm:px-3"
                                            key={value}
                                            value={value}
                                        >
                                            <span
                                                aria-hidden="true"
                                                className="sm:hidden"
                                            >
                                                {label[0]}
                                            </span>
                                            <span className="hidden sm:inline">{label}</span>
                                        </TabsTrigger>
                                    ))}
                                </TabsList>
                            </Tabs>
                        )}
                    </div>

                    <div
                        aria-busy={isPending}
                        className={cn('transition-opacity', isPending && 'opacity-60')}
                    >
                        <TabsContent value="analytics">
                            <DashboardAnalytics period={period} />
                        </TabsContent>

                        <TabsContent value="overview">
                            <DashboardOverview />
                        </TabsContent>
                    </div>
                </Tabs>
            </div>
        </>
    );
}

export default Dashboard;
