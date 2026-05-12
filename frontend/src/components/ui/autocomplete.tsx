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
import { cn } from '@/lib/utils';

/**
 * Free-text input with a popover dropdown of fuzzy-filtered suggestions.
 *
 * Built on top of `cmdk` (filtering + arrow/Enter navigation) and Radix
 * Popover (positioning, outside-click, escape). The visible field is a real
 * text input — unlike the shadcn `Combobox` pattern, the user can type
 * arbitrary values that are not in the suggestion list.
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

const AutocompleteContext = React.createContext<AutocompleteContextValue | null>(null);

const useAutocompleteContext = (): AutocompleteContextValue => {
    const ctx = React.useContext(AutocompleteContext);

    if (!ctx) {
        throw new Error('Autocomplete sub-components must be rendered inside <Autocomplete>.');
    }

    return ctx;
};

/**
 * Minimal controllable-state hook so we don't pull in
 * `@radix-ui/react-use-controllable-state` just for two state slots.
 *
 * Mirrors Radix's behaviour: when `controlled` is `undefined` the hook owns
 * the state; otherwise the parent does and we just forward updates via
 * `onChange`. Updates always call `onChange` so consumers can observe even
 * fully-controlled changes (e.g. logging).
 */
const useControllable = <T,>(controlled: T | undefined, defaultValue: T, onChange?: (value: T) => void) => {
    const [internal, setInternal] = React.useState<T>(defaultValue);
    const isControlled = controlled !== undefined;
    const value = isControlled ? (controlled as T) : internal;

    const onChangeRef = React.useRef(onChange);

    React.useEffect(() => {
        onChangeRef.current = onChange;
    }, [onChange]);

    const set = React.useCallback(
        (next: T) => {
            if (!isControlled) {
                setInternal(next);
            }

            onChangeRef.current?.(next);
        },
        [isControlled],
    );

    return [value, set] as const;
};

export interface AutocompleteProps {
    children: React.ReactNode;
    defaultOpen?: boolean;
    defaultValue?: string;
    /**
     * Custom matcher forwarded to `cmdk`. Returns `0..1` (1 = best match,
     * 0 = hidden). Defaults to `command-score` fuzzy matching.
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

const Autocomplete = ({
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
}: AutocompleteProps) => {
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

    return (
        <AutocompleteContext.Provider value={ctxValue}>
            <Popover
                onOpenChange={setOpen}
                open={open}
            >
                <Command
                    filter={filter}
                    label="Suggestions"
                    shouldFilter={shouldFilter}
                >
                    {children}
                </Command>
            </Popover>
        </AutocompleteContext.Provider>
    );
};

Autocomplete.displayName = 'Autocomplete';

export type AutocompleteInputProps = Omit<
    React.InputHTMLAttributes<HTMLInputElement>,
    'defaultValue' | 'onChange' | 'type' | 'value'
>;

const AutocompleteInput = React.forwardRef<HTMLInputElement, AutocompleteInputProps>(
    ({ className, id, onFocus, onKeyDown, onPointerDown, ...props }, forwardedRef) => {
        const { commit, inputId, inputRef, inputValue, openOnFocus, setInputValue, setOpen } = useAutocompleteContext();

        // Read live cmdk store so we know whether Enter would land on a real
        // match. When there are no matches (or no highlighted value), we
        // commit the raw input ourselves; otherwise we let cmdk fire the
        // selected item's `onSelect`, which will commit through the context.
        //
        // Single subscription returning a primitive boolean — `useSyncExternalStore`
        // does referential equality, so collapsing the two slots into one boolean
        // halves the per-keystroke re-renders without breaking the snapshot check
        // (which would happen if we returned a fresh object each time).
        const hasActiveMatch = useCommandState((state) => state.filtered.count > 0 && Boolean(state.value));

        // Skip the very first focus event — it's the `autoFocus` case (or any
        // other focus the parent fires synchronously during mount). Auto-opening
        // a popover before the user did anything is jarring; browser address
        // bars (Chrome, Firefox) behave the same way.
        //
        // No `setTimeout` needed: React's `autoFocus` lives in the commit phase
        // (`commitMount`), and the native `focus` event fires synchronously
        // inside that `el.focus()` call — so our `onFocus` handler runs before
        // any passive effect. The flag is therefore still `true` during the
        // mount-time focus and is flipped to `false` here, in time for every
        // subsequent user-driven focus (click, Tab, re-focus after Escape).
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
                 * `asChild` makes cmdk render through Radix `Slot`, which lets us
                 * delegate the actual DOM `<input>` (and its shadcn styling) to
                 * the project-wide `Input` component. cmdk still owns the value
                 * sync, ARIA combobox attributes and `cmdk-input=""` data hook,
                 * Slot just merges them into `Input` — handlers compose, primitive
                 * props from cmdk fill in whatever the child doesn't define.
                 *
                 * Note: cmdk hard-codes `autoComplete="off"`, `autoCorrect="off"`
                 * and `spellCheck={false}` on its primitive after the props spread,
                 * so we don't repeat them here.
                 */}
                <CommandPrimitive.Input
                    asChild
                    onValueChange={handleValueChange}
                    value={inputValue}
                >
                    <Input
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
    },
);
AutocompleteInput.displayName = 'AutocompleteInput';

export type AutocompleteContentProps = React.ComponentPropsWithoutRef<typeof PopoverPrimitive.Content> & {
    /**
     * Class name for the inner cmdk `List`. Use this to override the default
     * max-height (`max-h-[280px]`) per usage site.
     */
    listClassName?: string;
};

const AutocompleteContent = React.forwardRef<
    React.ElementRef<typeof PopoverPrimitive.Content>,
    AutocompleteContentProps
>(
    (
        {
            align = 'start',
            children,
            className,
            listClassName,
            onInteractOutside,
            onOpenAutoFocus,
            sideOffset = 4,
            ...props
        },
        ref,
    ) => {
        const { inputRef } = useAutocompleteContext();

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
                onInteractOutside={(event) => {
                    onInteractOutside?.(event);

                    if (event.defaultPrevented) {
                        return;
                    }

                    // Clicking the input itself shouldn't dismiss the popover —
                    // Radix treats the anchor as "outside" since it isn't a Trigger.
                    if (inputRef.current?.contains(event.target as Node)) {
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
                ref={ref}
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
                <CommandList className={cn('max-h-[280px] overflow-x-hidden overflow-y-auto', listClassName)}>
                    {children}
                </CommandList>
            </PopoverPrimitive.Content>
        );
    },
);
AutocompleteContent.displayName = 'AutocompleteContent';

export type AutocompleteEmptyProps = React.ComponentPropsWithoutRef<typeof CommandEmpty>;

const AutocompleteEmpty = React.forwardRef<React.ElementRef<typeof CommandEmpty>, AutocompleteEmptyProps>(
    (props, ref) => (
        <CommandEmpty
            ref={ref}
            {...props}
        />
    ),
);
AutocompleteEmpty.displayName = 'AutocompleteEmpty';

export type AutocompleteGroupProps = React.ComponentPropsWithoutRef<typeof CommandGroup>;

const AutocompleteGroup = React.forwardRef<React.ElementRef<typeof CommandGroup>, AutocompleteGroupProps>(
    (props, ref) => (
        <CommandGroup
            ref={ref}
            {...props}
        />
    ),
);
AutocompleteGroup.displayName = 'AutocompleteGroup';

export type AutocompleteItemProps = Omit<React.ComponentPropsWithoutRef<typeof CommandItem>, 'onSelect' | 'value'> & {
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

const AutocompleteItem = React.forwardRef<React.ElementRef<typeof CommandItem>, AutocompleteItemProps>(
    ({ onSelect, value, ...props }, ref) => {
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
                ref={ref}
                value={value}
                {...props}
            />
        );
    },
);
AutocompleteItem.displayName = 'AutocompleteItem';

export type AutocompleteSeparatorProps = React.ComponentPropsWithoutRef<typeof CommandSeparator>;

const AutocompleteSeparator = React.forwardRef<React.ElementRef<typeof CommandSeparator>, AutocompleteSeparatorProps>(
    (props, ref) => (
        <CommandSeparator
            ref={ref}
            {...props}
        />
    ),
);
AutocompleteSeparator.displayName = 'AutocompleteSeparator';

export {
    Autocomplete,
    AutocompleteContent,
    AutocompleteEmpty,
    AutocompleteGroup,
    AutocompleteInput,
    AutocompleteItem,
    AutocompleteSeparator,
};
