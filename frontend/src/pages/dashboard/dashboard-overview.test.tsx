import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

const query = vi.hoisted(() => ({ state: { error: undefined as Error | undefined, loading: false } }));

const stats = {
    totalUsageCacheIn: 0,
    totalUsageCacheOut: 0,
    totalUsageCostIn: 1,
    totalUsageCostOut: 0.5,
    totalUsageIn: 1000,
    totalUsageOut: 500,
};

vi.mock('@apollo/client/react', () => ({
    useQuery: () => ({
        data: {
            flowsStatsTotal: {
                totalAssistantsCount: 3,
                totalFlowsCount: 42,
                totalSubtasksCount: 2,
                totalTasksCount: 1,
            },
            toolcallsStatsByFunction: [],
            toolcallsStatsTotal: { totalCount: 7, totalDurationSeconds: 60 },
            usageStatsByAgentType: [{ agentType: 'pentester', stats }],
            usageStatsByModel: [{ model: 'gpt-test', provider: 'openai', stats }],
            usageStatsByProvider: [{ provider: 'openai', stats }],
            usageStatsTotal: stats,
        },
        ...query.state,
    }),
}));

import { DashboardOverview } from './dashboard-overview';

const placeholders = () => document.querySelectorAll('[data-slot="skeleton"], .animate-spin').length;

describe('DashboardOverview', () => {
    it('keeps the figures and tables on screen while they refetch', () => {
        query.state = { error: undefined, loading: true };

        render(<DashboardOverview />);

        expect(placeholders()).toBe(0);
        expect(screen.getByText('42')).toBeInTheDocument();
        expect(screen.getByText('openai')).toBeInTheDocument();
    });

    it('keeps them on screen when a refetch fails', () => {
        query.state = { error: new Error('stats unavailable'), loading: false };

        render(<DashboardOverview />);

        expect(screen.queryByText("Couldn't load")).not.toBeInTheDocument();
        expect(screen.getByText('42')).toBeInTheDocument();
        expect(screen.getByText('openai')).toBeInTheDocument();
    });
});
