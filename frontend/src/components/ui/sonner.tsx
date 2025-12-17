'use client';

import { CircleCheckIcon, InfoIcon, Loader2Icon, OctagonXIcon, TriangleAlertIcon } from 'lucide-react';
import { Toaster as Sonner, type ToasterProps } from 'sonner';

import { useTheme } from '@/hooks/use-theme';

const Toaster = ({ ...props }: ToasterProps) => {
    const { theme } = useTheme();

    return (
        <Sonner
            className="toaster group"
            icons={{
                error: <OctagonXIcon className="size-4" />,
                info: <InfoIcon className="size-4" />,
                loading: <Loader2Icon className="size-4 animate-spin" />,
                success: <CircleCheckIcon className="size-4" />,
                warning: <TriangleAlertIcon className="size-4" />,
            }}
            style={
                {
                    '--border-radius': 'var(--radius)',
                    '--normal-bg': 'hsl(var(--popover))',
                    '--normal-border': 'hsl(var(--border))',
                    '--normal-text': 'hsl(var(--popover-foreground))',
                    '--toast-icon-margin-end': '0px',
                    '--toast-icon-margin-start': '0px',
                    '--width': '320px',
                } as React.CSSProperties
            }
            theme={theme as ToasterProps['theme']}
            {...props}
        />
    );
};

export { Toaster };
