import { Search, X } from 'lucide-react';
import { motion } from 'motion/react';
import { type ChangeEvent, useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useDebouncedCallback } from 'use-debounce';

import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupInput } from '@/components/ui/input-group';
import { Kbd, KbdGroup } from '@/components/ui/kbd';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { cn } from '@/lib/utils';
import { isMac } from '@/lib/utils/platform';

// Matches `DataTableFilter`'s debounce so both search inputs on a page settle
// at the same pace.
const COMMIT_DEBOUNCE_MS = 150;

interface InputSearchProps {
    /** Accessible label for the trigger and the input. */
    ariaLabel?: string;
    className?: string;
    /** Lowercase key combined with the platform modifier (⌘/Ctrl) to expand from anywhere. `null` disables. */
    hotkey?: null | string;
    /** Expanded input width (default `140`). Collapsed state is always icon-only. */
    maxWidth?: number | string;
    onSearchChange: (value: string) => void;
    placeholder?: string;
    searchQuery: string;
}

/**
 * Collapsible search input. Icon-only when idle; clicking it (or pressing
 * `Mod`+`hotkey`) expands a controlled text input. Clearing has two equivalent
 * paths — the trailing `<X>` button and the first `Escape` press; both route
 * through `handleClear`. A second `Escape` on an empty field collapses + blurs.
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
    const [isExpanded, setIsExpanded] = useState(() => searchQuery.trim().length > 0);
    const inputRef = useRef<HTMLInputElement>(null);
    const [localValue, setLocalValue] = useState(searchQuery);
    const lastEmittedReference = useRef(searchQuery);

    const debouncedEmit = useDebouncedCallback((next: string) => {
        if (next === lastEmittedReference.current) {
            return;
        }

        lastEmittedReference.current = next;
        onSearchChange(next);
    }, COMMIT_DEBOUNCE_MS);

    // External → local sync. Outside source (back button, programmatic clear,
    // sibling control) wins; in-flight typed debounce is dropped.
    useEffect(() => {
        if (searchQuery === lastEmittedReference.current) {
            return;
        }

        debouncedEmit.cancel();
        lastEmittedReference.current = searchQuery;
        setLocalValue(searchQuery);
    }, [searchQuery, debouncedEmit]);

    const expand = useCallback(() => setIsExpanded(true), []);

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

    // `rAF` so motion's transform has started before we focus — otherwise the
    // caret lands on a zero-width box.
    useEffect(() => {
        if (!isExpanded) {
            return;
        }

        const id = requestAnimationFrame(() => inputRef.current?.focus());

        return () => cancelAnimationFrame(id);
    }, [isExpanded]);

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

    // Reads `localValue`, not the prop: a user who typed and blurred before
    // the debounce flushed should keep seeing their text.
    const handleInputBlur = useCallback(() => {
        if (localValue.trim().length === 0) {
            setIsExpanded(false);
        }
    }, [localValue]);

    // Auto-collapse when an external source clears the query while focus is
    // elsewhere; user-initiated blur is handled by `handleInputBlur`.
    useEffect(() => {
        if (searchQuery.trim().length > 0) {
            return;
        }

        if (document.activeElement !== inputRef.current) {
            setIsExpanded(false);
        }
    }, [searchQuery]);

    const handleChange = useCallback(
        (event: ChangeEvent<HTMLInputElement>) => {
            setLocalValue(event.target.value);
            debouncedEmit(event.target.value);
        },
        [debouncedEmit],
    );

    // Focus the input at the end so the auto-collapse effect above doesn't
    // fire — the user just emptied the value and is likely to type again.
    const handleClear = useCallback(() => {
        setLocalValue('');
        debouncedEmit.cancel();

        if (lastEmittedReference.current !== '') {
            lastEmittedReference.current = '';
            onSearchChange('');
        }

        inputRef.current?.focus();
    }, [debouncedEmit, onSearchChange]);

    const handleInputKeyDown = useCallback(
        (event: React.KeyboardEvent<HTMLInputElement>) => {
            if (event.key !== 'Escape') {
                return;
            }

            event.preventDefault();

            if (localValue.length > 0) {
                handleClear();

                return;
            }

            inputRef.current?.blur();
            setIsExpanded(false);
        },
        [handleClear, localValue],
    );

    return (
        <InputGroup
            className={cn(
                'h-8 w-auto shrink-0 transition-[background-color,border-color,box-shadow] duration-100',
                isExpanded
                    ? 'border-input dark:bg-input/30 shadow-xs'
                    : 'border-transparent shadow-none dark:bg-transparent',
                className,
            )}
        >
            <InputGroupAddon
                align="inline-start"
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
                    className="h-8 min-w-0 py-0 pl-2"
                    onBlur={handleInputBlur}
                    onChange={handleChange}
                    onKeyDown={handleInputKeyDown}
                    placeholder={placeholder}
                    ref={inputRef}
                    style={{ width: typeof maxWidth === 'number' ? `${maxWidth}px` : maxWidth }}
                    type="text"
                    value={localValue}
                />
            </motion.div>
            {isExpanded && localValue ? (
                <InputGroupAddon align="inline-end">
                    <InputGroupButton
                        aria-label={`Clear ${ariaLabel.toLowerCase()}`}
                        onClick={handleClear}
                        type="button"
                    >
                        <X />
                    </InputGroupButton>
                </InputGroupAddon>
            ) : null}
        </InputGroup>
    );
}
