import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

const queryResult = vi.hoisted(() => ({
    current: { data: undefined, error: undefined, loading: false } as Record<string, unknown>,
}));

vi.mock('@apollo/client/react', () => ({
    skipToken: Symbol('skipToken'),
    useQuery: () => queryResult.current,
}));

vi.mock('react-router-dom', () => ({
    useParams: () => ({ flowId: '42' }),
    useSearchParams: () => [new URLSearchParams()],
}));

vi.mock('@/components/shared/markdown', () => ({
    default: ({ children }: { children: string }) => <div data-testid="markdown">{children}</div>,
}));

vi.mock('@/lib/report', () => ({
    generateFileName: () => 'report',
    generatePDFFromMarkdown: vi.fn().mockResolvedValue(undefined),
    generateReport: (tasks: unknown[], flow: { title: string }) => `${flow.title}: ${tasks.length} tasks`,
}));

const { default: FlowReport } = await import('./flow-report');

describe('FlowReport load states', () => {
    it('renders the report when a partial error arrives alongside a usable flow', () => {
        queryResult.current = {
            data: { flow: { id: '42', title: 'Recon' }, tasks: null },
            error: new Error('failed to fetch tasks'),
            loading: false,
        };

        render(<FlowReport />);

        expect(screen.getByTestId('markdown')).toHaveTextContent('Recon: 0 tasks');
        expect(screen.queryByText('Failed to load flow data')).not.toBeInTheDocument();
    });

    it('reports a failure when the flow itself is missing', () => {
        queryResult.current = { data: { flow: null, tasks: null }, error: new Error('boom'), loading: false };

        render(<FlowReport />);

        expect(screen.getByText('Failed to load flow data')).toBeInTheDocument();
    });
});
