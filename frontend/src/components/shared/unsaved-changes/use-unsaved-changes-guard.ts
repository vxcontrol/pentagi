import { useCallback, useEffect, useRef, useState } from 'react';
import { type BlockerFunction, useBlocker } from 'react-router-dom';

export interface UnsavedChangesGuard {
    handleCancel: () => void;
    handleDiscard: () => void;
    handleOpenChange: (open: boolean) => void;
    handleSaveAndLeave: () => Promise<void>;
    isOpen: boolean;
    isSavingFromDialog: boolean;
    /**
     * Allows the next router navigation to bypass the blocker. Use this after
     * a successful save when the consumer needs to navigate to a fresh URL
     * (e.g. the new document page) without showing the dialog.
     */
    skipNextBlock: () => void;
}

export interface UseUnsavedChangesGuardArgs {
    isDirty: boolean;
    isFormValid: boolean;
    onSave: () => Promise<boolean>;
}

/**
 * Generic "are you sure you want to leave?" guard for any editing form
 * inside React Router. Combines two mechanisms:
 *
 *   1. `useBlocker` — intercepts in-app router navigations and exposes a
 *      blocked state that the consumer can render as a dialog.
 *   2. `beforeunload` — covers full page reloads and tab close attempts;
 *      the browser shows its native prompt while there are dirty changes.
 *
 * Designed to be UI-agnostic: pair with `<UnsavedChangesDialog>` (or any
 * custom dialog) by wiring the returned handlers.
 */
export function useUnsavedChangesGuard({
    isDirty,
    isFormValid,
    onSave,
}: UseUnsavedChangesGuardArgs): UnsavedChangesGuard {
    const allowNextRef = useRef(false);
    const isDirtyRef = useRef(isDirty);

    useEffect(() => {
        isDirtyRef.current = isDirty;
    }, [isDirty]);

    const blockerFn = useCallback<BlockerFunction>(({ currentLocation, nextLocation }) => {
        if (allowNextRef.current) {
            allowNextRef.current = false;

            return false;
        }

        if (currentLocation.pathname === nextLocation.pathname && currentLocation.search === nextLocation.search) {
            return false;
        }

        return isDirtyRef.current;
    }, []);

    const blocker = useBlocker(blockerFn);
    const [isSavingFromDialog, setIsSavingFromDialog] = useState(false);

    useEffect(() => {
        if (!isDirty) {
            return;
        }

        const handleBeforeUnload = (event: BeforeUnloadEvent) => {
            event.preventDefault();
            event.returnValue = '';
        };

        window.addEventListener('beforeunload', handleBeforeUnload);

        return () => {
            window.removeEventListener('beforeunload', handleBeforeUnload);
        };
    }, [isDirty]);

    const isOpen = blocker.state === 'blocked';

    const handleCancel = useCallback(() => {
        if (blocker.state === 'blocked') {
            blocker.reset();
        }
    }, [blocker]);

    const handleDiscard = useCallback(() => {
        if (blocker.state === 'blocked') {
            blocker.proceed();
        }
    }, [blocker]);

    const handleSaveAndLeave = useCallback(async () => {
        if (isSavingFromDialog || !isFormValid) {
            return;
        }

        setIsSavingFromDialog(true);

        try {
            const success = await onSave();

            if (success && blocker.state === 'blocked') {
                blocker.proceed();
            }
        } finally {
            setIsSavingFromDialog(false);
        }
    }, [blocker, isFormValid, isSavingFromDialog, onSave]);

    const handleOpenChange = useCallback(
        (open: boolean) => {
            if (isSavingFromDialog) {
                return;
            }

            if (!open && blocker.state === 'blocked') {
                blocker.reset();
            }
        },
        [blocker, isSavingFromDialog],
    );

    const skipNextBlock = useCallback(() => {
        allowNextRef.current = true;
    }, []);

    return {
        handleCancel,
        handleDiscard,
        handleOpenChange,
        handleSaveAndLeave,
        isOpen,
        isSavingFromDialog,
        skipNextBlock,
    };
}
