import { Search } from 'lucide-react';
import { motion } from 'motion/react';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';

import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupInput } from '@/components/ui/input-group';
import { Kbd, KbdGroup } from '@/components/ui/kbd';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { cn } from '@/lib/utils';
import { isMac } from '@/lib/utils/platform';

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
    // exit. Without this handler, opening the trigger but never typing would
    // leave the input expanded forever once focus moved away (the effect
    // below only re-runs on `searchQuery` changes, which a no-op blur isn't).
    const handleInputBlur = useCallback(() => {
        if (searchQuery.trim().length === 0) {
            setIsExpanded(false);
        }
    }, [searchQuery]);

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
    // This matches the muscle memory of native search controls and keeps the
    // keyboard a viable exit path for users who never reach for the mouse.
    const handleInputKeyDown = useCallback(
        (event: React.KeyboardEvent<HTMLInputElement>) => {
            if (event.key !== 'Escape') {
                return;
            }

            event.preventDefault();

            if (searchQuery.length > 0) {
                onSearchChange('');

                return;
            }

            inputRef.current?.blur();
            setIsExpanded(false);
        },
        [onSearchChange, searchQuery],
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
                    onChange={(event) => onSearchChange(event.target.value)}
                    onKeyDown={handleInputKeyDown}
                    placeholder={placeholder}
                    ref={inputRef}
                    style={{ width: typeof maxWidth === 'number' ? `${maxWidth}px` : maxWidth }}
                    type="text"
                    value={searchQuery}
                />
            </motion.div>
        </InputGroup>
    );
}
