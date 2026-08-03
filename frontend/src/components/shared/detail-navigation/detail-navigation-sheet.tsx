import { Search, X } from 'lucide-react';
import { type ReactNode, startTransition, useCallback, useMemo, useRef, useState } from 'react';

import { Badge } from '@/components/ui/badge';
import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupInput } from '@/components/ui/input-group';
import { Sheet, SheetContent, SheetHeader, SheetTitle } from '@/components/ui/sheet';
import { useElementVirtualList } from '@/hooks/use-element-virtual-list';
import { cn } from '@/lib/utils';

import type { DetailNavigationController } from './use-detail-navigation';

const VIRTUALIZE_THRESHOLD = 100;
const ROW_HEIGHT = 36;
const estimateRowSize = () => ROW_HEIGHT;

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
    const [scrollEl, setScrollEl] = useState<HTMLDivElement | null>(null);
    const buttonRefs = useRef(new Map<string, HTMLButtonElement>());
    const pendingFocusIdRef = useRef<null | string>(null);
    const [focusedId, setFocusedId] = useState<null | string>(null);

    const [localQuery, setLocalQuery] = useState(searchQuery);

    const hasEntries = items.length > 0;
    const trimmedQuery = localQuery.trim();
    const hasClearButton = hasSearch && trimmedQuery.length > 0;
    const shouldVirtualize = items.length > VIRTUALIZE_THRESHOLD;

    const { scrollToIndex, totalSize, virtualItems } = useElementVirtualList({
        count: items.length,
        enabled: shouldVirtualize,
        estimateSize: estimateRowSize,
        gap: 2,
        getItemKey: (index) => {
            const item = items[index];

            return item ? String(getId(item)) : index;
        },
        getScrollElement: () => scrollEl,
        overscan: 10,
        padding: 8,
    });

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
    // once per render and committed in the same pass that prompted it.
    //
    // Priorities, top-down:
    //   1. open→close / close→open transition: re-pin to `currentId` (or the
    //      first entry when no current exists) on open, and clear on close.
    //   2. While the sheet stays open, if the focused entry left the filtered
    //      subset, fall back to the first survivor.
    //   3. Otherwise hold whatever focus the user chose via arrow keys.
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

        // Re-sync the mirror on open only: the sheet is modal, so a controlled
        // searchQuery can change externally only while closed. A searchQuery-watching
        // effect would instead clobber in-progress typing.
        if (open) {
            setLocalQuery(searchQuery);
        }
    }

    if (desiredFocusId !== focusedId) {
        setFocusedId(desiredFocusId);
    }

    const focusOption = useCallback(
        (id: string, index: number) => {
            setFocusedId(id);

            if (shouldVirtualize) {
                scrollToIndex(index, { align: 'auto' });
            }

            const node = buttonRefs.current.get(id);

            if (node) {
                node.focus();
            } else {
                pendingFocusIdRef.current = id;
            }
        },
        [scrollToIndex, shouldVirtualize],
    );

    const handleOpenAutoFocus = useCallback(
        (event: Event) => {
            const index = focusedId === null ? undefined : indexById.get(focusedId);

            if (focusedId === null || index === undefined) {
                if (searchInputRef.current) {
                    event.preventDefault();
                    searchInputRef.current.focus();
                }

                return;
            }

            event.preventDefault();
            focusOption(focusedId, index);
        },
        [focusedId, focusOption, indexById],
    );

    const handleListKeyDown = useCallback(
        (event: React.KeyboardEvent<HTMLUListElement>) => {
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
                    focusOption(String(getId(target)), index);
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
        [focusedId, focusOption, getId, hasEntries, indexById, items],
    );

    const handleItemClick = useCallback(
        (item: T) => {
            onItemSelect(item);
        },
        [onItemSelect],
    );

    const handleSearchKeyDown = useCallback(
        (event: React.KeyboardEvent<HTMLInputElement>) => {
            if (event.key === 'ArrowDown' && hasEntries) {
                event.preventDefault();
                const first = items[0];

                if (first) {
                    focusOption(String(getId(first)), 0);
                }
            }
        },
        [focusOption, getId, hasEntries, items],
    );

    const clearSearch = useCallback(() => {
        setLocalQuery('');
        startTransition(() => clearSearchQuery());
    }, [clearSearchQuery]);

    // Intercept Esc at the Radix-Content level so we can clear a non-empty
    // search before the dialog's built-in "close on Esc" fires. Once the query
    // is empty, we let Radix close as usual — the two-step Esc affordance.
    const handleEscapeKeyDown = useCallback(
        (event: KeyboardEvent) => {
            if (hasSearch && trimmedQuery.length > 0) {
                event.preventDefault();
                clearSearch();
            }
        },
        [clearSearch, hasSearch, trimmedQuery.length],
    );

    const handleSearchChange = useCallback(
        (event: React.ChangeEvent<HTMLInputElement>) => {
            const { value } = event.target;
            setLocalQuery(value);
            startTransition(() => setSearchQuery(value));
        },
        [setSearchQuery],
    );

    const handleSearchClear = useCallback(() => {
        clearSearch();
        searchInputRef.current?.focus();
    }, [clearSearch]);

    const setButtonRef = useCallback((node: HTMLButtonElement | null) => {
        if (!node) {
            return;
        }

        const id = node.dataset.itemId;

        if (!id) {
            return;
        }

        buttonRefs.current.set(id, node);

        if (pendingFocusIdRef.current === id) {
            node.focus();
            pendingFocusIdRef.current = null;
        }

        return () => {
            buttonRefs.current.delete(id);
        };
    }, []);

    const renderOptionButton = useCallback(
        (item: T, index: number) => {
            const id = String(getId(item));
            const isCurrent = currentId != null && id === String(currentId);

            return (
                <button
                    aria-posinset={index + 1}
                    aria-selected={isCurrent}
                    aria-setsize={total}
                    className={cn(
                        'hover:bg-muted/50 focus-visible:ring-ring flex w-full min-w-0 items-center gap-2 rounded-md px-3 py-2 text-left text-sm focus-visible:ring-2 focus-visible:outline-hidden',
                        isCurrent && 'bg-muted text-foreground font-medium',
                    )}
                    data-item-id={id}
                    onClick={() => handleItemClick(item)}
                    onFocus={() => setFocusedId(id)}
                    ref={setButtonRef}
                    role="option"
                    tabIndex={id === focusedId ? 0 : -1}
                    type="button"
                >
                    {renderItem ? (
                        renderItem(item, isCurrent)
                    ) : (
                        <span className="min-w-0 flex-1 truncate">{getLabel(item)}</span>
                    )}
                </button>
            );
        },
        [currentId, focusedId, getId, getLabel, handleItemClick, renderItem, setButtonRef, total],
    );

    const listItems = useMemo(
        () =>
            items.map((item, index) => (
                <li
                    className="min-w-0"
                    key={String(getId(item))}
                    role="presentation"
                >
                    {renderOptionButton(item, index)}
                </li>
            )),
        [getId, items, renderOptionButton],
    );

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
                onOpenAutoFocus={handleOpenAutoFocus}
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
                                value={localQuery}
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
                    <div
                        className="min-w-0 flex-1 overflow-y-auto"
                        ref={setScrollEl}
                    >
                        {shouldVirtualize ? (
                            <ul
                                aria-label={sheetTitle}
                                className="relative"
                                onKeyDown={handleListKeyDown}
                                ref={listRef}
                                role="listbox"
                                style={{ height: totalSize }}
                            >
                                {virtualItems.map((virtualRow) => {
                                    const item = items[virtualRow.index];

                                    if (!item) {
                                        return null;
                                    }

                                    return (
                                        <li
                                            className="absolute inset-x-0 min-w-0 px-2"
                                            key={String(getId(item))}
                                            role="presentation"
                                            style={{
                                                height: ROW_HEIGHT,
                                                transform: `translateY(${virtualRow.start}px)`,
                                            }}
                                        >
                                            {renderOptionButton(item, virtualRow.index)}
                                        </li>
                                    );
                                })}
                            </ul>
                        ) : (
                            <ul
                                aria-label={sheetTitle}
                                className="flex flex-col gap-0.5 p-2"
                                onKeyDown={handleListKeyDown}
                                ref={listRef}
                                role="listbox"
                            >
                                {listItems}
                            </ul>
                        )}
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
