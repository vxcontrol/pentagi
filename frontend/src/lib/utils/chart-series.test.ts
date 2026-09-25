import { describe, expect, it } from 'vitest';

import { isSeriesEmpty } from './chart-series';

describe('isSeriesEmpty', () => {
    it('calls a dense week of zeros empty', () => {
        const week = Array.from({ length: 8 }, (_, day) => ({
            date: `2026-09-0${day + 1}`,
            flows: 0,
            subtasks: 0,
            tasks: 0,
        }));

        expect(isSeriesEmpty(week)).toBe(true);
    });

    it('calls a week with one busy day not empty', () => {
        const week = Array.from({ length: 8 }, (_, day) => ({
            date: `2026-09-0${day + 1}`,
            flows: day === 3 ? 1 : 0,
            subtasks: 0,
            tasks: 0,
        }));

        expect(isSeriesEmpty(week)).toBe(false);
    });

    it('reads every measure, not just the first', () => {
        expect(isSeriesEmpty([{ costIn: 0, costOut: 0.004, date: '2026-09-01', tokensIn: 0 }])).toBe(false);
    });

    it('still calls a series with no rows at all empty', () => {
        expect(isSeriesEmpty([])).toBe(true);
    });

    it('does not let the date decide', () => {
        expect(isSeriesEmpty([{ count: 0, date: '2026-09-01', duration: 0 }])).toBe(true);
    });
});
