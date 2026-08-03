import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { routes } from '@/lib/routes';

const { navigate } = vi.hoisted(() => ({ navigate: vi.fn() }));
const { toastError } = vi.hoisted(() => ({ toastError: vi.fn() }));
const queryResult = vi.hoisted(() => ({
    current: { data: undefined, error: undefined, loading: false, refetch: vi.fn() } as Record<string, unknown>,
}));

vi.mock('@apollo/client/react', () => ({
    skipToken: Symbol('skipToken'),
    useQuery: () => queryResult.current,
}));

vi.mock('react-router-dom', () => ({
    useNavigate: () => navigate,
    useParams: () => ({ knowledgeId: '7' }),
}));

vi.mock('sonner', () => ({ toast: { error: toastError } }));

vi.mock('@/providers/knowledges-provider', () => ({
    useKnowledges: () => ({ createKnowledge: vi.fn(), updateKnowledge: vi.fn() }),
}));

// The layout shell and the form drag in the header/editor tree; the branching under test sits
// above them in Knowledge itself, so stub them to a marker.
vi.mock('@/features/knowledges/knowledge-layout', () => ({
    KnowledgeLayout: ({ children }: { children: React.ReactNode }) => <div data-testid="layout">{children}</div>,
}));
vi.mock('@/features/knowledges/knowledge-form', async (importOriginal) => ({
    ...(await importOriginal<Record<string, unknown>>()),
    KnowledgeForm: () => <div data-testid="form" />,
}));

const { default: Knowledge } = await import('./knowledge');

describe('Knowledge detail load states', () => {
    beforeEach(() => {
        navigate.mockClear();
        toastError.mockClear();
    });

    it('renders an in-page error with Retry on a real load failure, without redirecting', () => {
        queryResult.current = { data: undefined, error: new Error('network down'), loading: false, refetch: vi.fn() };

        render(<Knowledge />);

        expect(screen.getByText('Error loading knowledge document')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /Try again/ })).toBeInTheDocument();
        expect(navigate).not.toHaveBeenCalled();
    });

    it('redirects to the list on a genuine not-found error', () => {
        queryResult.current = {
            data: undefined,
            error: new Error('no rows in result set'),
            loading: false,
            refetch: vi.fn(),
        };

        render(<Knowledge />);

        expect(screen.queryByText('Error loading knowledge document')).not.toBeInTheDocument();
        expect(toastError).toHaveBeenCalled();
        expect(navigate).toHaveBeenCalledWith(routes.knowledges, { replace: true });
    });

    it('redirects to the list when the query resolves with no document', () => {
        queryResult.current = { data: { knowledgeDocument: null }, error: undefined, loading: false, refetch: vi.fn() };

        render(<Knowledge />);

        expect(navigate).toHaveBeenCalledWith(routes.knowledges, { replace: true });
    });
});
