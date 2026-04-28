import { LayoutDashboard } from 'lucide-react';
import { useMemo, useState } from 'react';

import { Breadcrumb, BreadcrumbItem, BreadcrumbList, BreadcrumbPage } from '@/components/ui/breadcrumb';
import { Separator } from '@/components/ui/separator';
import { SidebarTrigger } from '@/components/ui/sidebar';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { UsageStatsPeriod } from '@/graphql/types';
import { getPeriodStorageKey } from '@/lib/storage-keys';
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

const Dashboard = () => {
    const periodStorageKey = useMemo(() => getPeriodStorageKey(), []);
    const [activeTab, setActiveTab] = useState('analytics');
    const [period, setPeriod] = useState<UsageStatsPeriod>(() => loadPeriod(periodStorageKey));

    const handlePeriodChange = (value: string) => {
        const next = value as UsageStatsPeriod;
        setPeriod(next);
        savePeriod(periodStorageKey, next);
    };

    return (
        <>
            <header className="bg-background sticky top-0 z-10 flex h-12 w-full shrink-0 items-center gap-2 border-b transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-12">
                <div className="flex items-center gap-2 px-4">
                    <SidebarTrigger className="-ml-1" />
                    <Separator
                        className="h-4"
                        orientation="vertical"
                    />
                    <Breadcrumb>
                        <BreadcrumbList>
                            <BreadcrumbItem>
                                <LayoutDashboard className="size-4" />
                                <BreadcrumbPage>Dashboard</BreadcrumbPage>
                            </BreadcrumbItem>
                        </BreadcrumbList>
                    </Breadcrumb>
                </div>
            </header>

            <div className="flex flex-col gap-6 p-4">
                <Tabs
                    className="w-full"
                    onValueChange={setActiveTab}
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
                                            key={value}
                                            value={value}
                                        >
                                            {label}
                                        </TabsTrigger>
                                    ))}
                                </TabsList>
                            </Tabs>
                        )}
                    </div>

                    <TabsContent value="analytics">
                        <DashboardAnalytics period={period} />
                    </TabsContent>

                    <TabsContent value="overview">
                        <DashboardOverview />
                    </TabsContent>
                </Tabs>
            </div>
        </>
    );
};

export default Dashboard;
