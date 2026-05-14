import { type KeyboardEvent, type ReactNode, useCallback, useEffect, useMemo, useRef, useState } from 'react';

import { ScrollArea } from '@/components/ui/scroll-area';
import { Sheet, SheetContent, SheetHeader, SheetTitle } from '@/components/ui/sheet';
import { cn } from '@/lib/utils';

interface ListNavigationSheetProps<T> {
    currentId: null | string | undefined;
    /**
     * Pre-computed index of `currentId` inside `items` (or `-1` when the
     * current entry doesn't belong to the filtered subset). Supplied by the
     * parent toolbar so the sheet doesn't re-scan `items` for membership.
     */
    currentIndex: number;
    getId: (item: T) => string;
    getLabel: (item: T) => string;
    /** Filtered + sorted items rendered as the sheet's listbox. */
    items: readonly T[];
    onItemSelect: (item: T) => void;
    onOpenChange: (open: boolean) => void;
    open: boolean;
    renderItem?: (item: T, isCurrent: boolean) => ReactNode;
    sheetIcon?: ReactNode;
    sheetTitle: string;
    /** Total count shown in the header — same as `items.length`, named explicitly for clarity. */
    total: number;
}

/**
 * Listbox-style overlay listing the navigable subset.
 *
 * Implements the WAI-ARIA single-select listbox pattern with **roving
 * tabindex**: only the currently-focused option carries `tabIndex={0}`,
 * the rest are `tabIndex={-1}`. Tab takes the user *past* the listbox in
 * one step; arrow keys move focus *within* it.
 *
 * Initial focus on open targets the current entry (if it's part of the
 * filtered subset) so users land oriented inside their own context.
 */
export function ListNavigationSheet<T>({
    currentId,
    currentIndex,
    getId,
    getLabel,
    items,
    onItemSelect,
    onOpenChange,
    open,
    renderItem,
    sheetIcon,
    sheetTitle,
    total,
}: ListNavigationSheetProps<T>) {
    const listRef = useRef<HTMLUListElement>(null);
    const buttonRefs = useRef(new Map<string, HTMLButtonElement>());
    const [focusedId, setFocusedId] = useState<null | string>(null);

    const hasEntries = items.length > 0;

    // Build an `id → index` map once per `items`/`getId` change so both the
    // per-render membership check below and the keyboard handler's lookup
    // stay O(1) instead of O(n). The IIFE that adjusts focus during render
    // previously called `items.some(...)` on every commit; the keyboard
    // handler ran `items.findIndex(...)` on every keystroke. Sharing one
    // structure between the two also makes the contract explicit: an entry
    // is "in the filtered subset" iff `indexById.has(id)`.
    const indexById = useMemo(() => {
        const map = new Map<string, number>();

        items.forEach((item, index) => {
            map.set(String(getId(item)), index);
        });

        return map;
    }, [items, getId]);

    // Single render-phase focus reconciliation. React's "adjust state when a
    // prop changes" idiom — see https://react.dev/reference/react/useState#storing-information-from-previous-renders
    // — collapsed into one comparison so the next desired focus is decided
    // once per render and committed in the same pass that prompted it (no
    // flash of stale focus, no double-setState ping-pong on edge cases like
    // "items change while the sheet was reopening with no current item").
    //
    // Priorities, top-down:
    //   1. open→close / close→open transition: re-pin to `currentId` (or the
    //      first entry when no current exists) on open, and clear on close.
    //   2. While the sheet stays open, if the focused entry left the
    //      filtered subset (list page narrowed the filter behind it), fall
    //      back to the first survivor — otherwise the keyboard model would
    //      stall on a row that's no longer rendered.
    //   3. Otherwise hold whatever focus the user chose via arrow keys.
    //
    // `lastOpen` starts at `false` so an initial `open=true` still trips
    // the open transition on the very first render.
    const [lastOpen, setLastOpen] = useState(false);

    const desiredFocusId = (() => {
        if (lastOpen !== open) {
            if (!open) {
                return null;
            }

            const firstItem = items[0];

            if (!firstItem) {
                return null;
            }

            // `currentId != null` narrows to `string`; the parent toolbar
            // already verified `currentId` belongs to the filtered subset
            // when it computed `currentIndex`, so no re-scan needed.
            return currentId != null && currentIndex >= 0 ? String(currentId) : String(getId(firstItem));
        }

        if (open && focusedId !== null && hasEntries && !indexById.has(focusedId)) {
            const fallbackItem = items[0];

            return fallbackItem ? String(getId(fallbackItem)) : null;
        }

        return focusedId;
    })();

    if (lastOpen !== open) {
        setLastOpen(open);
    }

    if (desiredFocusId !== focusedId) {
        setFocusedId(desiredFocusId);
    }

    // After roving focus moves, push the focus into the DOM. `rAF` defers past
    // Radix's own focus management so we don't fight its open-time focus trap.
    useEffect(() => {
        if (!open || focusedId === null) {
            return;
        }

        const id = requestAnimationFrame(() => {
            const node = buttonRefs.current.get(focusedId);

            if (!node) {
                return;
            }

            node.focus();

            if (focusedId === String(currentId ?? '')) {
                node.scrollIntoView({ block: 'center' });
            }
        });

        return () => cancelAnimationFrame(id);
    }, [open, focusedId, currentId]);

    // Translate arrow / Home / End into roving moves over `items`. Using the
    // array index instead of `querySelectorAll` keeps the keyboard model in
    // step with the React tree even if the sheet ever virtualises the list.
    // O(1) lookup via the shared `indexById` map — `findIndex` would scan on
    // every keystroke for nothing.
    const handleListKeyDown = useCallback(
        (event: KeyboardEvent<HTMLUListElement>) => {
            if (!hasEntries || focusedId === null) {
                return;
            }

            const focusedIndex = indexById.get(focusedId);

            if (focusedIndex === undefined) {
                return;
            }

            const moveTo = (index: number) => {
                event.preventDefault();
                const target = items[index];

                if (target) {
                    setFocusedId(String(getId(target)));
                }
            };

            if (event.key === 'ArrowDown') {
                moveTo(Math.min(focusedIndex + 1, items.length - 1));

                return;
            }

            if (event.key === 'ArrowUp') {
                moveTo(Math.max(focusedIndex - 1, 0));

                return;
            }

            if (event.key === 'Home') {
                moveTo(0);

                return;
            }

            if (event.key === 'End') {
                moveTo(items.length - 1);
            }
        },
        [focusedId, getId, hasEntries, indexById, items],
    );

    const handleItemClick = useCallback(
        (item: T) => {
            onItemSelect(item);
        },
        [onItemSelect],
    );

    // One stable callback ref reused for every button. The previous shape —
    // `setButtonRef(id) => (node) => …` — manufactured a new closure per id
    // on every render, which made React re-attach refs (a `delete` + `set`
    // round-trip on the `Map`) on every list re-render. Reading the id from
    // `data-item-id` keeps the closure identity-stable and the React-19
    // cleanup return value handles unmount without leaks.
    const setButtonRef = useCallback((node: HTMLButtonElement | null) => {
        if (!node) {
            return;
        }

        const id = node.dataset.itemId;

        if (!id) {
            return;
        }

        buttonRefs.current.set(id, node);

        return () => {
            buttonRefs.current.delete(id);
        };
    }, []);

    return (
        <Sheet
            onOpenChange={onOpenChange}
            open={open}
        >
            <SheetContent
                className="flex w-full max-w-sm flex-col gap-0 p-0 sm:max-w-sm"
                side="right"
            >
                <SheetHeader className="border-b p-4">
                    <SheetTitle className="flex items-center gap-2 text-base">
                        {sheetIcon}
                        <span>{sheetTitle}</span>
                        <span className="text-muted-foreground ml-auto text-sm font-normal tabular-nums">{total}</span>
                    </SheetTitle>
                </SheetHeader>
                {hasEntries ? (
                    <ScrollArea className="flex-1">
                        <ul
                            aria-label={sheetTitle}
                            className="flex flex-col gap-0.5 p-2"
                            onKeyDown={handleListKeyDown}
                            ref={listRef}
                            role="listbox"
                        >
                            {items.map((item) => {
                                const id = String(getId(item));
                                const isCurrent = currentId != null && id === String(currentId);
                                const isFocused = id === focusedId;

                                return (
                                    <li
                                        key={id}
                                        role="presentation"
                                    >
                                        <button
                                            aria-selected={isCurrent}
                                            className={cn(
                                                'hover:bg-muted/50 focus-visible:ring-ring flex w-full items-center gap-2 rounded-md px-3 py-2 text-left text-sm focus-visible:ring-2 focus-visible:outline-hidden',
                                                isCurrent && 'bg-muted text-foreground font-medium',
                                            )}
                                            data-item-id={id}
                                            onClick={() => handleItemClick(item)}
                                            onFocus={() => setFocusedId(id)}
                                            ref={setButtonRef}
                                            role="option"
                                            tabIndex={isFocused ? 0 : -1}
                                            type="button"
                                        >
                                            {renderItem ? (
                                                renderItem(item, isCurrent)
                                            ) : (
                                                <span className="truncate">{getLabel(item)}</span>
                                            )}
                                        </button>
                                    </li>
                                );
                            })}
                        </ul>
                    </ScrollArea>
                ) : (
                    <div className="text-muted-foreground flex flex-1 items-center justify-center px-4 text-center text-sm">
                        No items match the current filter.
                    </div>
                )}
            </SheetContent>
        </Sheet>
    );
}
