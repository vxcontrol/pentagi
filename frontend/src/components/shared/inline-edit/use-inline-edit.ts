import type React from 'react';

import { useCallback, useEffect, useRef, useState } from 'react';

interface UseInlineEditOptions {
    /**
     * When this value changes, the edit session is reset (closes any open
     * input). Use it for the entity id that owns the editor — navigating
     * between items should not carry a stale draft over.
     */
    resetKey?: null | string | undefined;
}

interface UseInlineEditResult<TElement extends HTMLElement = HTMLInputElement> {
    /**
     * Spread onto a Radix `<DropdownMenuContent>` (or any component with the
     * same `onCloseAutoFocus` semantics) when the dropdown contains a button
     * that toggles the inline editor. Prevents Radix's default focus-restore
     * from racing the editor's `requestAnimationFrame`-driven focus below.
     */
    handleDropdownCloseAutoFocus: (event: Event) => void;
    /** Ref to wire to the inline `<input>` element. */
    inputRef: React.RefObject<null | TElement>;
    isEditing: boolean;
    /** Begin an inline edit — the input mounts and receives focus on next frame. */
    startEdit: () => void;
    /** End the edit session without committing. */
    stopEdit: () => void;
}

/**
 * Shared state machine for inline-edit surfaces (double-click to rename,
 * quick-add, in-place note edits).
 *
 * Combines four micro-responsibilities that every editable surface needs:
 *   - `isEditing` boolean + start/stop helpers,
 *   - a ref for the inline input,
 *   - deferred focus + select-all on the next animation frame, so the focus
 *     lands *after* Radix's dropdown close-focus-restore completes (otherwise
 *     Radix wins the race and the input never receives focus),
 *   - an `onCloseAutoFocus` handler that opts out of Radix's restore while
 *     editing — used together with the deferred focus above.
 *
 * Pass the entity id as `resetKey` so navigation between items closes any
 * stale editor automatically.
 */
export function useInlineEdit<TElement extends HTMLElement = HTMLInputElement>({
    resetKey,
}: UseInlineEditOptions = {}): UseInlineEditResult<TElement> {
    const [isEditing, setIsEditing] = useState(false);
    const inputRef = useRef<null | TElement>(null);

    // Reset the edit session when `resetKey` changes. React docs call out
    // "adjust state when a prop changes" as the canonical render-phase
    // `setState` pattern — it's preferred over `useEffect` because the new
    // state lands in the same commit (no flash of stale "still editing" UI
    // on the new item) and it doesn't trigger the "cascading renders"
    // lint complaint that an effect-based reset would.
    const [lastResetKey, setLastResetKey] = useState(resetKey);

    if (lastResetKey !== resetKey) {
        setLastResetKey(resetKey);

        if (isEditing) {
            setIsEditing(false);
        }
    }

    useEffect(() => {
        if (!isEditing) {
            return;
        }

        const id = requestAnimationFrame(() => {
            const input = inputRef.current;

            if (!input) {
                return;
            }

            input.focus();

            // `<input>` and `<textarea>` both expose `select()`, but typing
            // `TElement extends HTMLElement` is wider than that — guard so
            // a future caller with `HTMLDivElement` doesn't crash here.
            if ('select' in input && typeof (input as { select: unknown }).select === 'function') {
                (input as unknown as HTMLInputElement).select();
            }
        });

        return () => cancelAnimationFrame(id);
    }, [isEditing]);

    const startEdit = useCallback(() => setIsEditing(true), []);
    const stopEdit = useCallback(() => setIsEditing(false), []);

    // Closure over `isEditing` directly. The callback identity flips on each
    // edit-mode toggle, but that's fine: Radix's `<DropdownMenuContent>` is
    // not memoized today and `onCloseAutoFocus` fires inside the same commit
    // that closes the dropdown — a ref-based stable callback would risk
    // reading a stale value through `useLatestRef`'s `useEffect` lag.
    const handleDropdownCloseAutoFocus = useCallback(
        (event: Event) => {
            if (isEditing) {
                event.preventDefault();
            }
        },
        [isEditing],
    );

    return { handleDropdownCloseAutoFocus, inputRef, isEditing, startEdit, stopEdit };
}
