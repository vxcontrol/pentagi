type ChartRow = Record<string, number | string>;

export const isSeriesEmpty = (rows: readonly ChartRow[]): boolean =>
    rows.every((row) => Object.entries(row).every(([key, value]) => key === 'date' || value === 0));
