import { type ReactNode } from 'react';

import { Card, CardContent } from '@/components/ui/card';
import { cn } from '@/lib/utils';

interface StatusCardProps {
    action?: ReactNode;
    className?: string;
    description?: ReactNode;
    icon?: ReactNode;
    title: ReactNode;
}

export function StatusCard({ action, className, description, icon, title }: StatusCardProps) {
    return (
        <Card className={cn('', className)}>
            <CardContent className="flex flex-col items-center justify-center px-4 py-8 text-center">
                {icon && <div className="mb-4 flex items-center justify-center">{icon}</div>}
                <h3 className="text-lg font-semibold text-foreground">{title}</h3>
                {description && <div className="mt-2 max-w-sm text-sm text-muted-foreground">{description}</div>}
                {action && <div className="mt-4">{action}</div>}
            </CardContent>
        </Card>
    );
}
