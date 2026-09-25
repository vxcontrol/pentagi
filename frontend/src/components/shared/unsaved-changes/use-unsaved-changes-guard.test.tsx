import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { useState } from 'react';
import { createMemoryRouter, RouterProvider, useLocation, useNavigate } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { useUnsavedChangesGuard } from './use-unsaved-changes-guard';

interface EditorProps {
    isFormValid?: boolean;
    onSave: () => Promise<boolean>;
}

function Editor({ isFormValid = true, onSave }: EditorProps) {
    const [isDirty, setIsDirty] = useState(false);
    const navigate = useNavigate();
    const guard = useUnsavedChangesGuard({ isDirty, isFormValid, onSave });

    return (
        <>
            <span data-slot="where">{useLocation().pathname}</span>
            <span data-slot="dialog-open">{String(guard.isOpen)}</span>
            <button
                data-slot="cancel"
                onClick={guard.handleCancel}
            >
                cancel
            </button>
            <button
                data-slot="discard"
                onClick={guard.handleDiscard}
            >
                discard
            </button>
            <button
                data-slot="save-and-leave"
                onClick={() => void guard.handleSaveAndLeave()}
            >
                save and leave
            </button>
            <button
                data-slot="waive"
                onClick={guard.skipNextBlock}
            >
                waive
            </button>
            <button
                data-slot="dirty"
                onClick={() => setIsDirty(true)}
            >
                type something
            </button>
            <button
                data-slot="clean"
                onClick={() => setIsDirty(false)}
            >
                save
            </button>
            <button
                data-slot="leave"
                onClick={() => void navigate('/elsewhere')}
            >
                leave
            </button>
        </>
    );
}

function Elsewhere() {
    return <span data-slot="where">{useLocation().pathname}</span>;
}

const renderEditor = ({ isFormValid = true, onSave = async () => true } = {}) => {
    const router = createMemoryRouter(
        [
            {
                element: (
                    <Editor
                        isFormValid={isFormValid}
                        onSave={onSave}
                    />
                ),
                path: '/edit',
            },
            { element: <Elsewhere />, path: '/elsewhere' },
        ],
        { initialEntries: ['/edit'] },
    );

    render(<RouterProvider router={router} />);

    return router;
};

const where = () => screen.getByTestId('where').textContent;

const makeDirtyAndLeave = async (user: ReturnType<typeof userEvent.setup>) => {
    await user.click(screen.getByTestId('dirty'));
    await user.click(screen.getByTestId('leave'));
    await waitFor(() => expect(screen.getByTestId('dialog-open')).toHaveTextContent('true'));
};

describe('leaving a form with unsaved changes', () => {
    it('lets a clean form go without asking', async () => {
        const user = userEvent.setup();

        renderEditor();

        await user.click(screen.getByTestId('leave'));

        await waitFor(() => expect(where()).toBe('/elsewhere'));
    });

    it('stops the navigation and asks once the form is dirty', async () => {
        const user = userEvent.setup();

        renderEditor();
        await makeDirtyAndLeave(user);

        expect(where()).toBe('/edit');
    });

    it('stays put on Cancel', async () => {
        const user = userEvent.setup();

        renderEditor();
        await makeDirtyAndLeave(user);

        await user.click(screen.getByTestId('cancel'));

        await waitFor(() => expect(screen.getByTestId('dialog-open')).toHaveTextContent('false'));
        expect(where()).toBe('/edit');
    });

    it('leaves without saving on Discard', async () => {
        const onSave = vi.fn(async () => true);
        const user = userEvent.setup();

        renderEditor({ onSave });
        await makeDirtyAndLeave(user);

        await user.click(screen.getByTestId('discard'));

        await waitFor(() => expect(where()).toBe('/elsewhere'));
        expect(onSave).not.toHaveBeenCalled();
    });

    it('saves first and then leaves on Save', async () => {
        const onSave = vi.fn(async () => true);
        const user = userEvent.setup();

        renderEditor({ onSave });
        await makeDirtyAndLeave(user);

        await user.click(screen.getByTestId('save-and-leave'));

        await waitFor(() => expect(where()).toBe('/elsewhere'));
        expect(onSave).toHaveBeenCalledTimes(1);
    });

    it('holds the reader on the form when the save fails', async () => {
        const onSave = vi.fn(async () => false);
        const user = userEvent.setup();

        renderEditor({ onSave });
        await makeDirtyAndLeave(user);

        await user.click(screen.getByTestId('save-and-leave'));

        expect(onSave).toHaveBeenCalledTimes(1);
        expect(where()).toBe('/edit');
    });

    it('refuses to save from the dialog while the form is invalid', async () => {
        const onSave = vi.fn(async () => true);
        const user = userEvent.setup();

        renderEditor({ isFormValid: false, onSave });
        await makeDirtyAndLeave(user);

        await user.click(screen.getByTestId('save-and-leave'));

        expect(onSave).not.toHaveBeenCalled();
        expect(where()).toBe('/edit');
    });

    it('lets the next navigation through once it is waived', async () => {
        const user = userEvent.setup();

        renderEditor();

        await user.click(screen.getByTestId('dirty'));

        await user.click(screen.getByTestId('waive'));

        await user.click(screen.getByTestId('leave'));

        await waitFor(() => expect(where()).toBe('/elsewhere'));
    });
});

describe('reloading a form with unsaved changes', () => {
    let listeners: string[] = [];
    let removed: string[] = [];

    beforeEach(() => {
        listeners = [];
        removed = [];
        vi.spyOn(window, 'addEventListener').mockImplementation((type: string) => {
            listeners.push(type);
        });
        vi.spyOn(window, 'removeEventListener').mockImplementation((type: string) => {
            removed.push(type);
        });
    });

    afterEach(() => {
        vi.restoreAllMocks();
    });

    it('arms the browser prompt while the form is dirty and drops it once it is saved', async () => {
        const user = userEvent.setup();

        renderEditor();

        expect(listeners, 'a clean form must not arm the prompt').not.toContain('beforeunload');

        await user.click(screen.getByTestId('dirty'));

        await waitFor(() => expect(listeners).toContain('beforeunload'));

        await user.click(screen.getByTestId('clean'));

        await waitFor(() => expect(removed).toContain('beforeunload'));
    });
});
