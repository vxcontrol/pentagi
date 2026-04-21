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
    payload,
}: {
    active?: boolean;
    formatter?: (value: number, name: string) => string;
    label?: string;
    labelFormatter?: (label: string) => string;
    payload?: Array<ChartTooltipPayloadEntry>;
}) => {
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
