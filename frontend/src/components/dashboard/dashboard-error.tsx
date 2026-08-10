import type { CSSProperties } from 'react';

import { AlertCircle } from 'lucide-react';

import { useI18n } from '@/hooks/use-i18n';
import { cn } from '@/lib/utils';

export function DashboardError({
    className,
    iconClassName,
    style,
}: {
    className?: string;
    iconClassName?: string;
    style?: CSSProperties;
}) {
    const { t } = useI18n();

    return (
        <div
            className={cn('text-muted-foreground flex flex-col items-center justify-center gap-2', className)}
            style={style}
        >
            <AlertCircle className={cn('text-muted-foreground/40 size-6', iconClassName)} />
            <p className="text-sm">{t('errors.couldntLoad')}</p>
        </div>
    );
}
