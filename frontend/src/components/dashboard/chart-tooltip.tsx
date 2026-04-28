import { useEffect, useRef } from 'react';

import { formatNumber } from '@/lib/utils/format';

export type ChartTooltipPayloadEntry = {
    color: string;
    name: string;
    value: number;
};

export const ChartTooltip = ({
    active,
    formatter,
    label,
    labelFormatter,
    onFirstActive,
    payload,
    sessionKey,
}: {
    active?: boolean;
    formatter?: (value: number, name: string) => string;
    label?: string;
    labelFormatter?: (label: string) => string;
    onFirstActive?: () => void;
    payload?: Array<ChartTooltipPayloadEntry>;
    sessionKey?: number;
}) => {
    // Store in ref to avoid stale closures in effects
    const onFirstActiveRef = useRef(onFirstActive);
    useEffect(() => {
        onFirstActiveRef.current = onFirstActive;
    }, [onFirstActive]);

    const shownInSessionRef = useRef(false);

    // New mouse-entry session: reset "already shown" tracking
    useEffect(() => {
        shownInSessionRef.current = false;
    }, [sessionKey]);

    // Notify parent the moment the tooltip first becomes visible in this session
    useEffect(() => {
        if (active && !shownInSessionRef.current) {
            shownInSessionRef.current = true;
            onFirstActiveRef.current?.();
        }
    }, [active]);

    if (!active || !payload?.length) {
        return null;
    }

    const renderedLabel = label ? (labelFormatter ? labelFormatter(label) : label) : '';

    return (
        <div className="bg-popover text-popover-foreground rounded-lg border px-3 py-2 shadow-md">
            {renderedLabel && <p className="text-muted-foreground mb-1 text-xs">{renderedLabel}</p>}
            {payload.map((entry) => (
                <div
                    className="flex items-center gap-2 text-sm"
                    key={entry.name}
                >
                    <span
                        className="size-2 rounded-full"
                        style={{ backgroundColor: entry.color }}
                    />
                    <span className="text-muted-foreground">{entry.name}:</span>
                    <span className="font-medium">
                        {formatter ? formatter(entry.value, entry.name) : formatNumber(entry.value)}
                    </span>
                </div>
            ))}
        </div>
    );
};
