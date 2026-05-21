import * as PopoverPrimitive from '@radix-ui/react-popover';
import { Command as CommandPrimitive, useCommandState } from 'cmdk';
import * as React from 'react';

import {
    Command,
    CommandEmpty,
    CommandGroup,
    CommandItem,
    CommandList,
    CommandSeparator,
} from '@/components/ui/command';
import { Input } from '@/components/ui/input';
import { Popover, PopoverAnchor } from '@/components/ui/popover';
import { useLatestRef } from '@/hooks/use-latest-ref';
import { cn } from '@/lib/utils';

interface AutocompleteContextValue {
    commit: (value: string) => void;
    inputId: string;
    inputRef: React.RefObject<HTMLInputElement | null>;
    inputValue: string;
    open: boolean;
    openOnFocus: boolean;
    setInputValue: (value: string) => void;
    setOpen: (open: boolean) => void;
}

/**
 * Free-text input with a popover dropdown of substring-filtered suggestions.
 *
 * Built on top of `cmdk` (keyboard navigation) and Radix Popover (positioning,
 * outside-click, escape). The visible field is a real text input — unlike the
 * shadcn `Combobox` pattern, the user can type arbitrary values that are not
 * in the suggestion list.
 *
 * Filtering defaults to a case-insensitive substring match against item
 * `value` and `keywords`. Pass a custom `filter` prop (e.g. cmdk's
 * `command-score` fuzzy matcher) when subsequence matching is desired.
 *
 * @example
 *  <Autocomplete value={path} onValueChange={setPath} onCommit={navigateTo}>
 *      <AutocompleteInput placeholder="/work" />
 *      <AutocompleteContent>
 *          <AutocompleteEmpty>No matches</AutocompleteEmpty>
 *          <AutocompleteGroup heading="Known paths">
 *              {paths.map((p) => (
 *                  <AutocompleteItem key={p} value={p}>{p}</AutocompleteItem>
 *              ))}
 *          </AutocompleteGroup>
 *      </AutocompleteContent>
 *  </Autocomplete>
 */

/**
 * Radix-style controllable state. When `controlled` is `undefined` the hook
 * owns the state; otherwise the parent does and we forward updates via
 * `onChange`. `onChange` always fires so fully-controlled consumers can
 * observe every set (e.g. for logging).
 *
 * Mirrors `@radix-ui/react-use-controllable-state` without pulling in the
 * dependency for a couple of state slots.
 */
function useControllable<T>(controlled: T | undefined, defaultValue: T, onChange?: (value: T) => void) {
    const [internal, setInternal] = React.useState<T>(defaultValue);
    const isControlled = controlled !== undefined;
    const value = isControlled ? (controlled as T) : internal;

    const onChangeRef = useLatestRef(onChange);

    const set = React.useCallback(
        (next: T) => {
            if (!isControlled) {
                setInternal(next);
            }

            onChangeRef.current?.(next);
        },
        [isControlled, onChangeRef],
    );

    return [value, set] as const;
}

const AutocompleteContext = React.createContext<AutocompleteContextValue | null>(null);

function useAutocompleteContext(): AutocompleteContextValue {
    const ctx = React.useContext(AutocompleteContext);

    if (!ctx) {
        throw new Error('Autocomplete sub-components must be rendered inside <Autocomplete>.');
    }

    return ctx;
}

// Case-insensitive substring matcher used as the default. cmdk ships
// `command-score` (fuzzy/subsequence), which is great for command palettes
// but surprising in plain autocomplete usage — e.g. typing `baa` would still
// highlight `bash` because the characters appear in order. Substring matching
// is the conventional autocomplete contract: results contain the typed text
// somewhere, or they don't appear at all. Consumers that need fuzzy semantics
// can opt back in via the `filter` prop.
const substringFilter = (value: string, search: string, keywords?: string[]): number => {
    const needle = search.trim().toLowerCase();

    if (!needle) {
        return 1;
    }

    if (value.toLowerCase().includes(needle)) {
        return 1;
    }

    if (keywords?.some((kw) => kw.toLowerCase().includes(needle))) {
        return 1;
    }

    return 0;
};

export type AutocompleteContentProps = React.ComponentProps<typeof PopoverPrimitive.Content> & {
    /**
     * Class name for the inner cmdk `List`. Use this to override the default
     * max-height (`max-h-[280px]`) per usage site.
     */
    listClassName?: string;
};

export type AutocompleteEmptyProps = React.ComponentProps<typeof CommandEmpty>;

export type AutocompleteGroupProps = React.ComponentProps<typeof CommandGroup>;

export type AutocompleteInputProps = Omit<
    React.ComponentProps<'input'>,
    'defaultValue' | 'onChange' | 'type' | 'value'
>;

export type AutocompleteItemProps = Omit<React.ComponentProps<typeof CommandItem>, 'onSelect' | 'value'> & {
    /**
     * Optional custom select handler. When provided, you become responsible
     * for committing the value (the default behaviour is a no-op so you can
     * decide whether to commit, navigate, etc.). When omitted, the item
     * commits its own `value` via `Autocomplete.onCommit`.
     */
    onSelect?: (value: string) => void;
    /**
     * The value committed when this item is chosen. Also fed into cmdk's
     * filter (so make sure it matches what the user is likely to type).
     */
    value: string;
};

export interface AutocompleteProps {
    children: React.ReactNode;
    defaultOpen?: boolean;
    defaultValue?: string;
    /**
     * Custom matcher forwarded to `cmdk`. Returns `0..1` (1 = best match,
     * 0 = hidden). Defaults to a case-insensitive substring match against
     * the item `value` (and any provided `keywords`); pass cmdk's
     * `command-score` here to opt back into fuzzy/subsequence matching.
     */
    filter?: (value: string, search: string, keywords?: string[]) => number;
    /**
     * Called when the user commits a value:
     *   - selecting a suggestion (commits the item's `value`),
     *   - pressing Enter without an active suggestion (commits the raw input).
     *
     * Selecting always closes the popover; Enter without a match also closes it.
     */
    onCommit?: (value: string) => void;
    onOpenChange?: (open: boolean) => void;
    onValueChange?: (value: string) => void;
    open?: boolean;
    /**
     * Whether to open the popover when the input gains focus. Defaults to `true`.
     * Set to `false` to only open on typing.
     */
    openOnFocus?: boolean;
    /**
     * Disable cmdk's built-in filtering — useful when the consumer pre-filters
     * the items themselves (e.g. async server-side suggestions).
     */
    shouldFilter?: boolean;
    value?: string;
}

export type AutocompleteSeparatorProps = React.ComponentProps<typeof CommandSeparator>;

function Autocomplete({
    children,
    defaultOpen = false,
    defaultValue = '',
    filter,
    onCommit,
    onOpenChange,
    onValueChange,
    open: openProp,
    openOnFocus = true,
    shouldFilter,
    value: valueProp,
}: AutocompleteProps) {
    const [inputValue, setInputValue] = useControllable<string>(valueProp, defaultValue, onValueChange);
    const [open, setOpen] = useControllable<boolean>(openProp, defaultOpen, onOpenChange);
    const inputRef = React.useRef<HTMLInputElement | null>(null);
    const inputId = React.useId();

    const commit = React.useCallback(
        (next: string) => {
            setInputValue(next);
            setOpen(false);
            onCommit?.(next);
        },
        [onCommit, setInputValue, setOpen],
    );

    const ctxValue = React.useMemo<AutocompleteContextValue>(
        () => ({
            commit,
            inputId,
            inputRef,
            inputValue,
            open,
            openOnFocus,
            setInputValue,
            setOpen,
        }),
        [commit, inputId, inputValue, open, openOnFocus, setInputValue, setOpen],
    );

    // cmdk hard-codes a sr-only `<label cmdk-label for={cmdkUseId}>` inside
    // `<Command>` and points its `for` at a useId that cmdk generates for
    // *its own* input. Our visible input gets its id from `FormControl`'s
    // Radix Slot instead, so the two ids never line up and Chrome flags
    // "Incorrect use of <label for=FORM_ELEMENT>". The label can't be
    // disabled from cmdk's API. Re-point the `for` at the real input id
    // after mount so the HTML-level association is valid (cmdk already
    // wires `aria-labelledby` to the same label element through Slot, so
    // screen readers stayed correct — this fix is for DevTools/HTML
    // validators).
    React.useLayoutEffect(() => {
        const input = inputRef.current;
        const cmdkLabel = input?.closest('[cmdk-root]')?.querySelector<HTMLLabelElement>('label[cmdk-label]');

        if (input?.id && cmdkLabel && cmdkLabel.htmlFor !== input.id) {
            cmdkLabel.htmlFor = input.id;
        }
    });

    return (
        <AutocompleteContext.Provider value={ctxValue}>
            <Popover
                onOpenChange={setOpen}
                open={open}
            >
                <Command
                    filter={filter ?? substringFilter}
                    label="Suggestions"
                    shouldFilter={shouldFilter}
                >
                    {children}
                </Command>
            </Popover>
        </AutocompleteContext.Provider>
    );
}

function AutocompleteContent({
    align = 'start',
    children,
    className,
    listClassName,
    onFocusOutside,
    onInteractOutside,
    onOpenAutoFocus,
    ref,
    sideOffset = 4,
    ...props
}: AutocompleteContentProps) {
    const { inputRef } = useAutocompleteContext();
    const contentRef = React.useRef<null | React.ComponentRef<typeof PopoverPrimitive.Content>>(null);
    const composedRef = React.useCallback(
        (node: null | React.ComponentRef<typeof PopoverPrimitive.Content>) => {
            contentRef.current = node;

            if (typeof ref === 'function') {
                ref(node);
            } else if (ref) {
                (ref as React.MutableRefObject<typeof node>).current = node;
            }
        },
        [ref],
    );

    return (
        /*
         * Render inline — NO `PopoverPrimitive.Portal`. When the autocomplete
         * lives inside a Radix `Dialog`, the dialog locks scroll on `body`
         * (`body { pointer-events: none }` + a wheel/touch interceptor on
         * the document). A portal'd popover ends up on `body`, outside the
         * dialog DOM, and inherits the lock — wheel events on the suggestion
         * list never fire and the user can't scroll. Rendering inline keeps
         * the popover inside the dialog content, where the scroll lock
         * doesn't apply, while Radix Popper still positions it correctly
         * relative to the anchor (the visible input).
         *
         * Styles are copied verbatim from `components/ui/popover.tsx` so
         * appearance and animations stay consistent with shadcn defaults.
         */
        <PopoverPrimitive.Content
            align={align}
            className={cn(
                'bg-popover text-popover-foreground data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2 z-50 origin-(--radix-popover-content-transform-origin) rounded-md border shadow-md outline-hidden',
                // Width matches the input (anchor); max-h keeps the popover
                // inside the viewport when the suggestion list is large.
                'max-h-(--radix-popover-content-available-height) w-(--radix-popover-trigger-width) min-w-(--radix-popover-trigger-width) overflow-hidden p-0',
                className,
            )}
            onFocusOutside={(event) => {
                onFocusOutside?.(event);

                if (event.defaultPrevented) {
                    return;
                }

                // Radix `DismissableLayer` treats the anchor input as
                // "outside" (it isn't a Trigger), and its React-tree flag
                // races on Content's own focusin when our `CommandList`
                // synchronously bounces focus back to the input — both
                // produce spurious `focusOutside` events that would
                // dismiss the popover. Treat focus on the anchor input
                // or anywhere inside Content as "inside".
                //
                // Radix dispatches this event directly on the focused
                // element (`bubbles: false`), so `event.target` IS it.
                const focusTarget = event.target as Node | null;
                const focusInsideInput = !!(focusTarget && inputRef.current?.contains(focusTarget));
                const focusInsideContent = !!(focusTarget && contentRef.current?.contains(focusTarget));

                if (focusInsideInput || focusInsideContent) {
                    event.preventDefault();
                }
            }}
            onInteractOutside={(event) => {
                onInteractOutside?.(event);

                if (event.defaultPrevented) {
                    return;
                }

                // Clicking the input itself shouldn't dismiss the popover —
                // Radix treats the anchor as "outside" since it isn't a
                // Trigger. The real pointer target is the focused
                // element via `event.target` (custom event is
                // dispatched on the target, `bubbles: false`).
                const interactTarget = event.target as Node | null;

                if (interactTarget && inputRef.current?.contains(interactTarget)) {
                    event.preventDefault();
                }
            }}
            onOpenAutoFocus={(event) => {
                onOpenAutoFocus?.(event);

                if (event.defaultPrevented) {
                    return;
                }

                // Keep focus on the input so typing keeps working immediately.
                event.preventDefault();
            }}
            ref={composedRef}
            sideOffset={sideOffset}
            {...props}
        >
            {/*
             * Explicit max-height + scroll on the cmdk list. shadcn's
             * `CommandList` defaults to `max-h-[300px] overflow-y-auto`, but
             * re-applying it here lets consumers override the cap via
             * `listClassName` without losing scroll behaviour, and pins the
             * list so a long suggestion set scrolls inside the popover
             * instead of pushing the popover past the viewport.
             */}
            <CommandList
                className={cn('max-h-[280px] overflow-x-hidden overflow-y-auto', listClassName)}
                onFocus={(event) => {
                    // cmdk re-focuses the input by id when its value
                    // changes (`document.getElementById(inputId).focus()`),
                    // falling back to the list when the lookup fails — and
                    // it fails as soon as a wrapper like `FormControl`
                    // overrides the input's id via Radix Slot. The listbox
                    // then steals focus and typing breaks. Per the
                    // aria-activedescendant pattern the listbox shouldn't
                    // hold real focus (tabindex=-1), so bounce it back.
                    // `target === currentTarget` skips legitimate child focus.
                    if (event.target === event.currentTarget) {
                        inputRef.current?.focus();
                    }
                }}
            >
                {children}
            </CommandList>
        </PopoverPrimitive.Content>
    );
}

function AutocompleteEmpty(props: AutocompleteEmptyProps) {
    return <CommandEmpty {...props} />;
}

function AutocompleteGroup(props: AutocompleteGroupProps) {
    return <CommandGroup {...props} />;
}

function AutocompleteInput({
    className,
    id,
    onFocus,
    onKeyDown,
    onPointerDown,
    ref: forwardedRef,
    ...props
}: AutocompleteInputProps & { ref?: React.Ref<HTMLInputElement> }) {
    const { commit, inputId, inputRef, inputValue, open, openOnFocus, setInputValue, setOpen } =
        useAutocompleteContext();

    // True when Enter would land on a real match. Selecting a suggestion
    // commits through cmdk's `onSelect`; Enter without a match commits the
    // raw input ourselves. Returning a primitive (not an object) lets
    // `useSyncExternalStore`'s referential check halve per-keystroke renders.
    const hasActiveMatch = useCommandState((state) => state.filtered.count > 0 && Boolean(state.value));

    // Suppress the autoFocus-fired `focus` on mount: opening the popover
    // before the user did anything is jarring (browser address bars behave
    // the same). React's autoFocus fires the native event synchronously
    // during `commitMount`, before any effect — so the flag is still `true`
    // for that first focus and flipped to `false` here for every
    // user-driven focus (click, Tab, re-focus after Escape).
    const isMountFocusRef = React.useRef(true);

    React.useEffect(() => {
        isMountFocusRef.current = false;
    }, []);

    const setInputRef = React.useCallback(
        (node: HTMLInputElement | null) => {
            inputRef.current = node;

            if (typeof forwardedRef === 'function') {
                forwardedRef(node);
            } else if (forwardedRef) {
                (forwardedRef as React.MutableRefObject<HTMLInputElement | null>).current = node;
            }
        },
        [forwardedRef, inputRef],
    );

    const handleFocus = (event: React.FocusEvent<HTMLInputElement>) => {
        onFocus?.(event);

        if (event.defaultPrevented) {
            return;
        }

        if (openOnFocus && !isMountFocusRef.current) {
            setOpen(true);
        }
    };

    const handlePointerDown = (event: React.PointerEvent<HTMLInputElement>) => {
        onPointerDown?.(event);

        if (event.defaultPrevented) {
            return;
        }

        // Without this, clicking an already-focused input (e.g. when the
        // dialog opened with autoFocus on the input and the user then
        // clicks it) won't fire `focus` again — and the popover would stay
        // closed despite a clear user intent. PointerDown is the cleanest
        // signal of "user is interacting with the field right now".
        if (openOnFocus) {
            setOpen(true);
        }
    };

    const handleKeyDown = (event: React.KeyboardEvent<HTMLInputElement>) => {
        onKeyDown?.(event);

        if (event.defaultPrevented) {
            return;
        }

        if (event.key === 'Escape') {
            setOpen(false);

            return;
        }

        if (event.key === 'ArrowDown') {
            // Make the dropdown reachable from the keyboard even when it
            // didn't auto-open on focus. cmdk's own ArrowDown handler will
            // then move the highlight into the list.
            setOpen(true);

            return;
        }

        if (event.key === 'Enter' && !hasActiveMatch) {
            // Stop cmdk from also handling Enter (which would no-op here
            // but keeps the popover open).
            event.preventDefault();
            event.stopPropagation();
            commit(inputValue);
        }
    };

    const handleValueChange = (next: string) => {
        setInputValue(next);
        // Re-open on typing — handles the case where the user closed the
        // popover with Escape but then resumes editing, and is the main
        // entry point now that focus alone no longer auto-opens on mount.
        setOpen(true);
    };

    return (
        <PopoverAnchor asChild>
            {/*
             * `asChild` delegates the DOM input (and shadcn styling) to the
             * project's `Input` while cmdk keeps value sync, ARIA combobox
             * attributes and `cmdk-input=""`. cmdk also hard-codes
             * `autoComplete/autoCorrect="off"` and `spellCheck={false}` —
             * don't repeat them.
             */}
            <CommandPrimitive.Input
                asChild
                onValueChange={handleValueChange}
                value={inputValue}
            >
                {/*
                 * `aria-expanded` is hard-coded to `true` inside cmdk —
                 * the attribute lies whenever the popover is closed
                 * (after Enter, Escape, outside click). Radix Slot
                 * merges child props on top of slot props for non-handler
                 * attributes, so passing `aria-expanded={open}` here
                 * wins over cmdk's hard-coded value and keeps it in sync
                 * with the real popover state for screen readers.
                 */}
                <Input
                    aria-expanded={open}
                    className={className}
                    id={id ?? inputId}
                    onFocus={handleFocus}
                    onKeyDown={handleKeyDown}
                    onPointerDown={handlePointerDown}
                    ref={setInputRef}
                    {...props}
                />
            </CommandPrimitive.Input>
        </PopoverAnchor>
    );
}

function AutocompleteItem({ onSelect, value, ...props }: AutocompleteItemProps) {
    const { commit } = useAutocompleteContext();

    return (
        <CommandItem
            onSelect={(selected) => {
                if (onSelect) {
                    onSelect(selected);

                    return;
                }

                commit(selected);
            }}
            value={value}
            {...props}
        />
    );
}

function AutocompleteSeparator(props: AutocompleteSeparatorProps) {
    return <CommandSeparator {...props} />;
}

export {
    Autocomplete,
    AutocompleteContent,
    AutocompleteEmpty,
    AutocompleteGroup,
    AutocompleteInput,
    AutocompleteItem,
    AutocompleteSeparator,
};
