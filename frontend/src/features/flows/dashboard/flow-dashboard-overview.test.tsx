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
            flowStatsByFlow: { totalAssistantsCount: 3, totalSubtasksCount: 2, totalTasksCount: 42 },
            toolcallsStatsByFlow: { totalCount: 7, totalDurationSeconds: 60 },
            toolcallsStatsByFunctionForFlow: [],
            usageStatsByAgentTypeForFlow: [{ agentType: 'pentester', stats }],
            usageStatsByFlow: stats,
            usageStatsByModelAgentsForFlow: [],
        },
        ...query.state,
    }),
}));

import { FlowDashboardOverview } from './flow-dashboard-overview';

const placeholders = () => document.querySelectorAll('[data-slot="skeleton"], .animate-spin').length;

describe('FlowDashboardOverview', () => {
    it('keeps the figures and tables on screen while they refetch', () => {
        query.state = { error: undefined, loading: true };

        render(<FlowDashboardOverview flowId="5" />);

        expect(placeholders()).toBe(0);
        expect(screen.getByText('42')).toBeInTheDocument();
        expect(screen.getByText('pentester')).toBeInTheDocument();
    });

    it('keeps them on screen when a refetch fails', () => {
        query.state = { error: new Error('stats unavailable'), loading: false };

        render(<FlowDashboardOverview flowId="5" />);

        expect(screen.queryByText("Couldn't load")).not.toBeInTheDocument();
        expect(screen.getByText('42')).toBeInTheDocument();
    });
});
