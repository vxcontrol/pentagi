import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { type Control, Controller } from 'react-hook-form';
import { createMemoryRouter, Link, RouterProvider } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';

import { KnowledgeAnswerType, KnowledgeDocType } from '@/graphql/types';

import type { FormValues, SubmitResult } from './knowledge-form';

vi.mock('@/hooks/use-breakpoint', () => ({
    useBreakpoint: () => ({ isDesktop: true, isMobile: false }),
}));

vi.mock('@apollo/client/react', () => ({ useMutation: () => [vi.fn(), {}] }));

vi.mock('@/providers/user-provider', () => ({
    useUser: () => ({ authInfo: { privileges: ['anonymize.call'] } }),
}));

vi.mock('sonner', () => ({ toast: { error: vi.fn(), info: vi.fn(), success: vi.fn() } }));

vi.mock('./knowledge-form-layout', () => {
    const Layout = ({ control }: { control: Control<FormValues> }) => (
        <>
            <Controller
                control={control}
                name="content"
                render={({ field }) => (
                    <textarea
                        aria-label="content"
                        {...field}
                    />
                )}
            />
            <Controller
                control={control}
                name="question"
                render={({ field }) => (
                    <input
                        aria-label="question"
                        {...field}
                    />
                )}
            />
        </>
    );

    return { KnowledgeFormLayoutDesktop: Layout, KnowledgeFormLayoutMobile: Layout };
});

vi.mock('./knowledge-header', () => ({
    KnowledgeHeader: ({ saveButton }: { saveButton?: React.ReactNode }) => <div>{saveButton}</div>,
}));

import { KnowledgeForm } from './knowledge-form';

const newValues: FormValues = {
    answerType: KnowledgeAnswerType.Other,
    codeLang: '',
    content: '',
    description: '',
    docType: KnowledgeDocType.Answer,
    guideType: undefined,
    question: '',
};

describe('KnowledgeForm — Save and leave (data router)', () => {
    it('proceeds the blocked navigation without honoring a CREATE redirect', async () => {
        const user = userEvent.setup();
        const onSubmit = vi
            .fn<(values: FormValues, dirty: unknown) => Promise<SubmitResult>>()
            .mockResolvedValue({ redirectTo: '/knowledges/new-id' });

        const router = createMemoryRouter(
            [
                {
                    element: (
                        <>
                            <Link to="/elsewhere">go elsewhere</Link>
                            <KnowledgeForm
                                initialValues={newValues}
                                isNew
                                onSubmit={onSubmit}
                            />
                        </>
                    ),
                    path: '/',
                },
                { element: <div>elsewhere page</div>, path: '/elsewhere' },
                { element: <div>new doc page</div>, path: '/knowledges/new-id' },
            ],
            { initialEntries: ['/'] },
        );

        render(<RouterProvider router={router} />);

        await user.type(screen.getByLabelText('content'), 'hello world');
        await user.type(screen.getByLabelText('question'), 'why');

        await user.click(screen.getByRole('link', { name: 'go elsewhere' }));

        const dialog = await screen.findByRole('dialog');
        await user.click(within(dialog).getByRole('button', { name: 'Save' }));

        await waitFor(() => expect(screen.getByText('elsewhere page')).toBeInTheDocument());
        expect(screen.queryByText('new doc page')).not.toBeInTheDocument();
        expect(onSubmit).toHaveBeenCalledOnce();
    });
});
