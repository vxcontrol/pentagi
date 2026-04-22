import type { ReactNode } from 'react';

import { BarChart2, Loader2 } from 'lucide-react';
import { ResponsiveContainer } from 'recharts';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

export const ChartCard = ({
    children,
    className,
    description,
    empty,
    height = 300,
    loading,
    title,
}: {
    children: ReactNode;
    className?: string;
    description?: ReactNode;
    empty?: boolean;
    height?: number;
    loading?: boolean;
    title: ReactNode;
}) => (
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
                    <Loader2 className="text-muted-foreground size-6 animate-spin" />
                </div>
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
