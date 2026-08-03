import type { ReactNode } from 'react';

import { BarChart2 } from 'lucide-react';
import { ResponsiveContainer } from 'recharts';

import { DashboardError } from '@/components/dashboard/dashboard-error';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Spinner } from '@/components/ui/spinner';

export function ChartCard({
    children,
    className,
    description,
    empty,
    error,
    height = 300,
    loading,
    title,
}: {
    children: ReactNode;
    className?: string;
    description?: ReactNode;
    empty?: boolean;
    error?: boolean;
    height?: number;
    loading?: boolean;
    title: ReactNode;
}) {
    return (
        <Card className={className}>
            <CardHeader>
                <CardTitle>{title}</CardTitle>
                {description && <CardDescription>{description}</CardDescription>}
            </CardHeader>
            <CardContent>
                {loading ? (
                    <div
                        className="flex items-center justify-center"
                        style={{ height }}
                    >
                        <Spinner
                            className="text-muted-foreground size-6"
                            variant="circle"
                        />
                    </div>
                ) : error ? (
                    <DashboardError
                        iconClassName="size-10"
                        style={{ height }}
                    />
                ) : empty ? (
                    <div
                        className="flex flex-col items-center justify-center gap-2"
                        style={{ height }}
                    >
                        <BarChart2 className="text-muted-foreground/30 size-10" />
                        <p className="text-muted-foreground text-sm">No data for this period</p>
                    </div>
                ) : (
                    <ResponsiveContainer
                        height={height}
                        width="100%"
                    >
                        {children}
                    </ResponsiveContainer>
                )}
            </CardContent>
        </Card>
    );
}
