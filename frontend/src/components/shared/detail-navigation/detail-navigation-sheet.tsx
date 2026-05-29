import { Search, X } from 'lucide-react';
import { type KeyboardEvent, type ReactNode, useCallback, useEffect, useMemo, useRef, useState } from 'react';

import { Badge } from '@/components/ui/badge';
import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupInput } from '@/components/ui/input-group';
import { Sheet, SheetContent, SheetHeader, SheetTitle } from '@/components/ui/sheet';
import { cn } from '@/lib/utils';

import type { DetailNavigationController } from './use-detail-navigation';

interface DetailNavigationSheetProps<T extends { id: string }> {
    controller: DetailNavigationController<T>;
    /**
     * Render a free-text search input in the sheet header that drives
     * `controller.setSearchQuery`. Defaults to `true` — pass `false` to opt
     * out for the rare consumer that wants the sheet to stay URL-filter-only.
     */
    hasSearch?: boolean;
    renderItem?: (item: T, isCurrent: boolean) => ReactNode;
    /** Placeholder for the in-sheet search input. Defaults to "Search…". */
    searchPlaceholder?: string;
    sheetIcon?: ReactNode;
    sheetTitle: string;
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
export function DetailNavigationSheet<T extends { id: string }>({
    controller,
    hasSearch = true,
    renderItem,
    searchPlaceholder = 'Search…',
    sheetIcon,
    sheetTitle,
}: DetailNavigationSheetProps<T>) {
    // Destructure at the top so existing `useMemo` / `useEffect` deps below
    // read individual fields rather than the controller object — keeps the
    // identity story the same as before the refactor.
    const {
        clearSearchQuery,
        currentId,
        currentIndex,
        filteredItems: items,
        getId,
        getLabel,
        handleItemSelect: onItemSelect,
        isSheetOpen: open,
        searchQuery,
        setSearchQuery,
        setSheetOpen: onOpenChange,
        total,
    } = controller;

    const listRef = useRef<HTMLUListElement>(null);
    const searchInputRef = useRef<HTMLInputElement>(null);
    const buttonRefs = useRef(new Map<string, HTMLButtonElement>());
    const [focusedId, setFocusedId] = useState<null | string>(null);

    const hasEntries = items.length > 0;
    const trimmedQuery = searchQuery.trim();
    const hasClearButton = hasSearch && trimmedQuery.length > 0;

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

            // `currentId != null` narrows to `string`; the controller has
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
    //
    // Gate: only auto-focus an option when the user's focus is already
    // somewhere that *expects* roving (a list option) or has not yet landed
    // (`document.activeElement === document.body` right after open, before
    // Radix moves focus into the dialog). Crucially, we must NOT steal focus
    // when the user is typing in the search input: every keystroke that
    // flushes the debounce can re-target `focusedId` (when `currentId` falls
    // out of the filtered subset, render-phase reconciliation snaps focus to
    // the first survivor). Without this gate, focus jumps out of the input
    // mid-type and the user can only enter a few characters before losing
    // their place — exact repro: type "aes" with current flow 837 visible,
    // focus moves to button 836 at the 150 ms debounce boundary.
    //
    // The fix is at the cause, not the symptom: a local input mirror (the
    // `InputSearch` / `DataTableFilter` pattern) would keep keystrokes from
    // being dropped during a state round-trip, but it would not stop this
    // effect from yanking focus away. The two patterns solve different bugs.
    useEffect(() => {
        if (!open || focusedId === null) {
            return;
        }

        const activeEl = document.activeElement;
        const focusIsOnSearchInput = activeEl !== null && activeEl === searchInputRef.current;

        if (focusIsOnSearchInput) {
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

    // ArrowDown from the input jumps focus into the listbox so keyboard
    // users can flow `type → arrow down → enter` without hunting for the
    // list. We move focus *here* synchronously (not in the auto-focus
    // effect) because that effect now skips when focus is on the search
    // input — otherwise it would yank focus mid-type. Escape is *not*
    // handled here — Radix listens for it on a native document handler
    // that React-level `stopPropagation` cannot reach. See
    // `handleEscapeKeyDown` below, wired through `onEscapeKeyDown`.
    const handleSearchKeyDown = useCallback(
        (event: React.KeyboardEvent<HTMLInputElement>) => {
            if (event.key === 'ArrowDown' && hasEntries) {
                event.preventDefault();
                const first = items[0];

                if (!first) {
                    return;
                }

                const firstId = String(getId(first));
                setFocusedId(firstId);
                // Button refs for the current items are already mounted —
                // this handler fires during a real user keystroke, so the
                // listbox commit that wired them up has already happened.
                buttonRefs.current.get(firstId)?.focus();
            }
        },
        [getId, hasEntries, items],
    );

    // Intercept Esc at the Radix-Content level so we can clear a non-empty
    // search before the dialog's built-in "close on Esc" fires. Once the
    // query is empty, we let Radix close as usual — matches the two-step
    // Esc affordance of `InputSearch` (clear, then dismiss).
    const handleEscapeKeyDown = useCallback(
        (event: KeyboardEvent) => {
            if (hasSearch && trimmedQuery.length > 0) {
                event.preventDefault();
                clearSearchQuery();
            }
        },
        [clearSearchQuery, hasSearch, trimmedQuery.length],
    );

    const handleSearchChange = useCallback(
        (event: React.ChangeEvent<HTMLInputElement>) => {
            setSearchQuery(event.target.value);
        },
        [setSearchQuery],
    );

    const handleSearchClear = useCallback(() => {
        clearSearchQuery();
        searchInputRef.current?.focus();
    }, [clearSearchQuery]);

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
                // Radix expects either a `<Description>` or an explicit
                // `aria-describedby={undefined}` opt-out. The sheet is just a
                // listbox of items, the `SheetTitle` already describes it.
                aria-describedby={undefined}
                className="flex w-full max-w-sm flex-col gap-0 p-0 sm:max-w-sm"
                onEscapeKeyDown={handleEscapeKeyDown}
                side="right"
            >
                <SheetHeader className="gap-3 border-b p-4">
                    <SheetTitle className="flex items-center gap-2 pr-8 text-base">
                        {sheetIcon}
                        <span>{sheetTitle}</span>
                        <Badge
                            className="ml-auto font-normal tabular-nums"
                            variant="secondary"
                        >
                            {total}
                        </Badge>
                    </SheetTitle>
                    {hasSearch ? (
                        <InputGroup className="h-9">
                            <InputGroupAddon align="inline-start">
                                <Search
                                    aria-hidden="true"
                                    className="text-muted-foreground"
                                />
                            </InputGroupAddon>
                            <InputGroupInput
                                aria-label={searchPlaceholder}
                                className="h-9 py-0"
                                onChange={handleSearchChange}
                                onKeyDown={handleSearchKeyDown}
                                placeholder={searchPlaceholder}
                                ref={searchInputRef}
                                type="text"
                                value={searchQuery}
                            />
                            {hasClearButton ? (
                                <InputGroupAddon align="inline-end">
                                    <InputGroupButton
                                        aria-label="Clear search"
                                        onClick={handleSearchClear}
                                        size="icon-sm"
                                        type="button"
                                        variant="ghost"
                                    >
                                        <X aria-hidden="true" />
                                    </InputGroupButton>
                                </InputGroupAddon>
                            ) : null}
                        </InputGroup>
                    ) : null}
                </SheetHeader>
                {hasEntries ? (
                    <div className="min-w-0 flex-1 overflow-y-auto">
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
                                        className="min-w-0"
                                        key={id}
                                        role="presentation"
                                    >
                                        <button
                                            aria-selected={isCurrent}
                                            className={cn(
                                                'hover:bg-muted/50 focus-visible:ring-ring flex w-full min-w-0 items-center gap-2 rounded-md px-3 py-2 text-left text-sm focus-visible:ring-2 focus-visible:outline-hidden',
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
                                                <span className="min-w-0 flex-1 truncate">{getLabel(item)}</span>
                                            )}
                                        </button>
                                    </li>
                                );
                            })}
                        </ul>
                    </div>
                ) : (
                    <div className="text-muted-foreground flex flex-1 items-center justify-center px-4 text-center text-sm">
                        {trimmedQuery.length > 0
                            ? `No items match "${trimmedQuery}".`
                            : 'No items match the current filter.'}
                    </div>
                )}
            </SheetContent>
        </Sheet>
    );
}
