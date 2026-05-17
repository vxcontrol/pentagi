import type { ReactNode } from 'react';

import { Button, type ButtonProps } from '@/components/ui/button';
import { cn } from '@/lib/utils';

interface HeaderButtonProps extends Omit<ButtonProps, 'children'> {
    endIcon?: ReactNode;
    icon: ReactNode;
    label: ReactNode;
}

// Action button rendered inside a page header. Collapses to an icon-only square
// on viewports narrower than the `md` breakpoint (matches `useBreakpoint`'s
// `mobile` threshold of 768px) and expands to icon + label (and optional
// trailing icon, e.g. a dropdown chevron) on wider screens. `aria-label` is
// auto-derived from `label` when it's a plain string so the icon-only mobile
// state stays accessible without the caller having to remember it.
export function HeaderButton({
    'aria-label': ariaLabel,
    className,
    endIcon,
    icon,
    label,
    size = 'sm',
    ...props
}: HeaderButtonProps) {
    const accessibleLabel = ariaLabel ?? (typeof label === 'string' ? label : undefined);

    return (
        <Button
            aria-label={accessibleLabel}
            className={cn('w-8 px-0 md:w-auto md:px-3', className)}
            size={size}
            {...props}
        >
            {icon}
            <span className="hidden md:inline">{label}</span>
            {endIcon ? <span className="hidden md:inline-flex">{endIcon}</span> : null}
        </Button>
    );
}
