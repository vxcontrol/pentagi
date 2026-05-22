import { Search } from 'lucide-react';
import { motion } from 'motion/react';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';

import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupInput } from '@/components/ui/input-group';
import { Kbd, KbdGroup } from '@/components/ui/kbd';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { useDebouncedValue } from '@/hooks/use-debounced-value';
import { useLatestRef } from '@/hooks/use-latest-ref';
import { cn } from '@/lib/utils';
import { isMac } from '@/lib/utils/platform';

// Burst-typing protection: keystrokes hit `localValue` synchronously for an
// instant UI response; the debounced mirror is what we propagate upstream
// (which may walk the router and re-render the whole page subtree). Mirrors
// `FILTER_DEBOUNCE_MS` in `data-table.tsx` so both search inputs on a page
// share a single "feels responsive" target.
const COMMIT_DEBOUNCE_MS = 150;

interface InputSearchProps {
    /** Accessible label for the trigger button + the input. */
    ariaLabel?: string;
    className?: string;
    /**
     * Lowercase character that, combined with the platform modifier (⌘ on
     * macOS, Ctrl elsewhere), expands and focuses the input from anywhere on
     * the page. Pass `null` to disable the global hotkey entirely — useful on
     * pages that should not steal the browser's own bindings. Default `'f'`.
     */
    hotkey?: null | string;
    /**
     * Width (px or any CSS length) the input grows to once expanded. The
     * collapsed state is always just the icon. Default `140`.
     */
    maxWidth?: number | string;
    onSearchChange: (value: string) => void;
    placeholder?: string;
    searchQuery: string;
}

/**
 * Collapsible search input. Idle state shows only a `<Search>` icon button;
 * clicking it (or pressing the configured hotkey) expands a controlled text
 * input with a smooth width animation. Blurring an empty input collapses it
 * back. Escape clears the value, a second Escape collapses + blurs.
 *
 * The trigger is a real `<button>` (via `InputGroupButton`) — never a `<div>`
 * with `onClick` — so the control is reachable with Tab and announced by
 * screen readers. The input lives outside that button, so we don't nest
 * interactive elements (an `<input>` inside a `<button>` is invalid HTML).
 *
 * The component is controlled — parent owns `searchQuery` and reacts to
 * `onSearchChange`. The only local state is the expand/collapse presentation
 * flag; the initial value mirrors whether the controlled query is non-empty,
 * so deep-linking to `?qs=foo` lands directly in the expanded state.
 */
export function InputSearch({
    ariaLabel = 'Search',
    className,
    hotkey = 'f',
    maxWidth = 140,
    onSearchChange,
    placeholder = 'Search...',
    searchQuery,
}: InputSearchProps) {
    // Initialise expanded if the parent already has a non-empty query: a
    // shared link with `?qs=foo` should land expanded, not behind the icon.
    const [isExpanded, setIsExpanded] = useState(() => searchQuery.trim().length > 0);
    const inputRef = useRef<HTMLInputElement>(null);

    // Local mirror of the controlled `searchQuery`. Keystrokes set this
    // synchronously so the caret never lags; the debounced shadow below is
    // what we hand back to the parent. Mirrors `DataTableFilter`'s approach
    // — without it, parents that funnel `onSearchChange` through the router
    // (URL → re-render → controlled value snaps back) can drop characters
    // when typing faster than the round-trip.
    const [localValue, setLocalValue] = useState(searchQuery);
    const debouncedLocalValue = useDebouncedValue(localValue, COMMIT_DEBOUNCE_MS);
    // Tracks the last value we either emitted upstream or accepted from the
    // parent, so the two sync effects below can distinguish "our own
    // round-trip arrived" from "the parent reset me from elsewhere" without
    // a race.
    const lastEmittedReference = useRef(searchQuery);
    // Mirror onto a ref so the emit effect can depend on `debouncedLocalValue`
    // alone. Parents commonly pass an inline arrow; listing it in deps would
    // re-run the emit effect on every parent render with a stale
    // `debouncedLocalValue` (still holding the pre-clear value), letting a
    // synchronous Esc / programmatic clear race the debounce timer and
    // resurrect the previous query. See same fix in `DataTableFilter`.
    const onSearchChangeReference = useLatestRef(onSearchChange);

    // External → local sync. The parent owns the source of truth (URL,
    // upstream state, etc.); when *they* change it (Esc clear, programmatic
    // wipe of `?qs=`, back button), reset the local mirror. Skip when the
    // incoming value matches what we last emitted — that's our own commit
    // coming home and overwriting in-flight typing would lose characters.
    useEffect(() => {
        if (searchQuery !== lastEmittedReference.current) {
            setLocalValue(searchQuery);
            lastEmittedReference.current = searchQuery;
        }
    }, [searchQuery]);

    // Local → external emit, debounced. We only call `onSearchChange` when
    // the debounced value differs from what's already upstream — otherwise
    // every external reset would round-trip back through this effect.
    useEffect(() => {
        if (debouncedLocalValue !== lastEmittedReference.current) {
            lastEmittedReference.current = debouncedLocalValue;
            onSearchChangeReference.current(debouncedLocalValue);
        }
    }, [debouncedLocalValue, onSearchChangeReference]);

    const expand = useCallback(() => setIsExpanded(true), []);

    // Hint shown over the collapsed trigger. The hotkey badge mirrors the
    // platform's modifier glyph (⌘ on Apple, Ctrl elsewhere) so the tooltip
    // is actionable — the user sees the exact keys they can press without
    // having to discover them by trial. `<Kbd>` already self-themes for
    // `[data-slot=tooltip-content]` containers (see `components/ui/kbd.tsx`),
    // so no extra styling is needed to make the chip readable on the dark
    // tooltip surface.
    const tooltipContent = useMemo(() => {
        if (!hotkey) {
            return placeholder;
        }

        const modifier = isMac() ? '⌘' : 'Ctrl';

        return (
            <>
                {placeholder}{' '}
                <KbdGroup>
                    <Kbd>{modifier}</Kbd>
                    <Kbd>{hotkey.toUpperCase()}</Kbd>
                </KbdGroup>
            </>
        );
    }, [hotkey, placeholder]);

    // Focus the input the frame after it appears. `rAF` defers past the same
    // commit so motion's transform has started, otherwise focus can land on
    // a still-zero-width box and the user's caret blinks invisibly until the
    // animation completes.
    useEffect(() => {
        if (!isExpanded) {
            return;
        }

        const id = requestAnimationFrame(() => inputRef.current?.focus());

        return () => cancelAnimationFrame(id);
    }, [isExpanded]);

    // Global hotkey: Mod+<hotkey> from anywhere on the page. We hijack the
    // browser default (e.g. Ctrl+F's "find in page") only for the exact
    // configured key — pass `null` to opt out entirely.
    useEffect(() => {
        if (!hotkey) {
            return;
        }

        const normalized = hotkey.toLowerCase();

        const handleKeyDown = (event: KeyboardEvent) => {
            if (!(event.metaKey || event.ctrlKey)) {
                return;
            }

            if (event.key.toLowerCase() !== normalized) {
                return;
            }

            event.preventDefault();
            expand();
        };

        document.addEventListener('keydown', handleKeyDown);

        return () => document.removeEventListener('keydown', handleKeyDown);
    }, [expand, hotkey]);

    // Blur collapses when the value is empty — the classic "lost interest"
    // exit. Reads `localValue`, not the prop: a user who typed then blurred
    // before the debounce flushed still has visible text in the field, and
    // collapsing that out from under them would feel like a data-loss bug.
    const handleInputBlur = useCallback(() => {
        if (localValue.trim().length === 0) {
            setIsExpanded(false);
        }
    }, [localValue]);

    // Belt-and-braces for programmatic clears: if the controlled query drops
    // to empty while focus is somewhere else entirely (e.g. another effect
    // wipes `?qs=` from the URL), this still collapses. The user-blur path
    // above handles the common case.
    useEffect(() => {
        if (searchQuery.trim().length > 0) {
            return;
        }

        if (document.activeElement !== inputRef.current) {
            setIsExpanded(false);
        }
    }, [searchQuery]);

    // Escape: first press clears the value, second press collapses + blurs.
    // Decision and emit both run against `localValue` so the keypress feels
    // instant: clear is unconditional and side-steps the debounce, going
    // straight to the parent without waiting on the next tick.
    const handleInputKeyDown = useCallback(
        (event: React.KeyboardEvent<HTMLInputElement>) => {
            if (event.key !== 'Escape') {
                return;
            }

            event.preventDefault();

            if (localValue.length > 0) {
                setLocalValue('');
                lastEmittedReference.current = '';
                onSearchChangeReference.current('');

                return;
            }

            inputRef.current?.blur();
            setIsExpanded(false);
        },
        [localValue, onSearchChangeReference],
    );

    return (
        <InputGroup
            className={cn(
                // Mirror the `size="sm"` button geometry used by `HeaderButton`
                // so the search control lines up vertically with sibling
                // actions in any page header (h-8 / rounded-md / px-3).
                'h-8 w-auto shrink-0 transition-[background-color,border-color,box-shadow] duration-100',
                isExpanded
                    ? 'border-input dark:bg-input/30 shadow-xs'
                    : 'border-transparent shadow-none dark:bg-transparent',
                className,
            )}
        >
            <InputGroupAddon
                align="inline-start"
                // Collapsed trigger sits flush with the addon edge — the
                // addon's default `pl-3` would otherwise add a 12 px gutter
                // that wastes header real estate when only the icon is on
                // screen. `has-[>button]:ml-[-0.45rem]` already handles the
                // analogous adjustment when expanded.
                className={cn(!isExpanded && '-mx-1.5')}
            >
                {isExpanded ? (
                    <Search
                        aria-hidden="true"
                        className="text-muted-foreground"
                    />
                ) : (
                    <Tooltip>
                        <TooltipTrigger asChild>
                            <InputGroupButton
                                aria-label={ariaLabel}
                                onClick={expand}
                                size="icon-sm"
                                type="button"
                                variant="ghost"
                            >
                                <Search aria-hidden="true" />
                            </InputGroupButton>
                        </TooltipTrigger>
                        <TooltipContent>{tooltipContent}</TooltipContent>
                    </Tooltip>
                )}
            </InputGroupAddon>
            <motion.div
                animate={{
                    opacity: isExpanded ? 1 : 0,
                    width: isExpanded ? maxWidth : 0,
                }}
                className="overflow-hidden"
                initial={false}
                transition={{ duration: 0.1, ease: 'easeOut' }}
            >
                <InputGroupInput
                    aria-label={ariaLabel}
                    // `Input` defaults to `h-9` — that's 4 px taller than our
                    // `h-8` group, which would let the input edge poke past
                    // the bordered group container. Pin it to the group's
                    // height; the `min-w-0` is to keep flexbox from
                    // bottoming out on the implicit `min-content` width.
                    className="h-8 min-w-0 py-0 pl-2"
                    onBlur={handleInputBlur}
                    onChange={(event) => setLocalValue(event.target.value)}
                    onKeyDown={handleInputKeyDown}
                    placeholder={placeholder}
                    ref={inputRef}
                    style={{ width: typeof maxWidth === 'number' ? `${maxWidth}px` : maxWidth }}
                    type="text"
                    value={localValue}
                />
            </motion.div>
        </InputGroup>
    );
}
