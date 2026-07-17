import type { ReactNode } from 'react';

import { AlertCircle } from 'lucide-react';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';

export function MetricCard({
    className,
    description,
    error,
    icon,
    loading,
    title,
    value,
}: {
    className?: string;
    description?: ReactNode;
    error?: boolean;
    icon?: ReactNode;
    loading?: boolean;
    title: ReactNode;
    value: ReactNode;
}) {
    return (
        <Card className={className}>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
                <CardTitle className="text-sm font-medium">
                    {loading ? <Skeleton className="h-5 w-24" /> : title}
                </CardTitle>
                {loading ? (
                    <Skeleton className="size-4 shrink-0 rounded" />
                ) : error ? (
                    <AlertCircle className="text-muted-foreground/60 size-4 shrink-0" />
                ) : (
                    icon
                )}
            </CardHeader>
            <CardContent className="flex flex-col gap-1">
                {loading ? (
                    <Skeleton className="h-8 w-24" />
                ) : error ? (
                    <div className="text-muted-foreground/60 text-2xl font-bold">—</div>
                ) : (
                    <div className="text-2xl font-bold">{value}</div>
                )}
                {(description || error) &&
                    (loading ? (
                        <Skeleton className="mt-1 h-3 w-32" />
                    ) : (
                        <p className="text-muted-foreground text-xs">{error ? "Couldn't load" : description}</p>
                    ))}
            </CardContent>
        </Card>
    );
}
