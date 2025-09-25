import { Card, CardContent } from '@/components/ui/card';
import { cn } from '@/lib/utils';
import { type ReactNode } from 'react';

interface StatusCardProps {
    icon?: ReactNode;
    title: ReactNode;
    description?: ReactNode;
    action?: ReactNode;
    className?: string;
}

export function StatusCard({ icon, title, description, action, className }: StatusCardProps) {
    return (
        <Card className={cn('', className)}>
            <CardContent className="flex flex-col items-center justify-center py-8 px-4 text-center">
                {icon && <div className="mb-4 flex items-center justify-center">{icon}</div>}
                <h3 className="text-lg font-semibold text-foreground">{title}</h3>
                {description && <div className="mt-2 text-sm text-muted-foreground max-w-sm">{description}</div>}
                {action && <div className="mt-4">{action}</div>}
            </CardContent>
        </Card>
    );
}
