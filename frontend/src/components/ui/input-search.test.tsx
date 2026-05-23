import type { ReactNode } from 'react';

import { act, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { useState } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { TooltipProvider } from '@/components/ui/tooltip';

import { InputSearch } from './input-search';

// Collapsed state shows a single `<button aria-label="Search docs">` (the
// expand trigger). Expanded state replaces that button with a non-interactive
// `<Search>` icon and exposes the controlled `<input>`. The presence of the
// trigger button is therefore the canonical "is it collapsed" signal — the
// input itself always lives in the DOM (just behind a `width: 0` motion box).
const queryTrigger = () => screen.queryByRole('button', { name: 'Search docs' });
const getInput = () => screen.getByPlaceholderText('Search…') as HTMLInputElement;

// Real-time `sleep` — comfortably past the 150ms debounce.
const sleep = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));
const PAST_DEBOUNCE_MS = 250;

const Wrapper = ({ children }: { children: ReactNode }) => <TooltipProvider>{children}</TooltipProvider>;

beforeEach(() => {
    if (!window.matchMedia) {
        Object.defineProperty(window, 'matchMedia', {
            configurable: true,
            value: (query: string) => ({
                addEventListener: vi.fn(),
                addListener: vi.fn(),
                dispatchEvent: vi.fn(),
                matches: false,
                media: query,
                onchange: null,
                removeEventListener: vi.fn(),
                removeListener: vi.fn(),
            }),
            writable: true,
        });
    }
});

interface SearchHostProps {
    emitted: string[];
    initialQuery?: string;
    onSetExternalRef?: (setter: (next: string) => void) => void;
}

// Mirrors /knowledges: a fresh inline `onSearchChange` per render plus an
// imperative setter for tests that simulate an external source.
const SearchHost = ({ emitted, initialQuery = '', onSetExternalRef }: SearchHostProps) => {
    const [value, setValue] = useState(initialQuery);

    if (onSetExternalRef) {
        onSetExternalRef(setValue);
    }

    return (
        <InputSearch
            ariaLabel="Search docs"
            onSearchChange={(next) => {
                emitted.push(next);
                setValue(next);
            }}
            placeholder="Search…"
            searchQuery={value}
        />
    );
};

describe('InputSearch — collapsed/expanded presentation', () => {
    it('mounts collapsed (trigger present) when initial query is empty', () => {
        render(<SearchHost emitted={[]} />, { wrapper: Wrapper });
        expect(queryTrigger()).toBeInTheDocument();
    });

    it('mounts expanded (trigger absent, input populated) when deep-linked', () => {
        render(
            <SearchHost
                emitted={[]}
                initialQuery="jwt"
            />,
            { wrapper: Wrapper },
        );
        expect(queryTrigger()).not.toBeInTheDocument();
        expect(getInput().value).toBe('jwt');
    });

    it('expands on trigger click and focuses the input on the next animation frame', async () => {
        const user = userEvent.setup();
        render(<SearchHost emitted={[]} />, { wrapper: Wrapper });

        await user.click(queryTrigger()!);

        await waitFor(() => expect(queryTrigger()).not.toBeInTheDocument());
        await waitFor(() => expect(document.activeElement).toBe(getInput()));
    });
});

describe('InputSearch — typing and outbound emit', () => {
    it('coalesces a burst of keystrokes into a single emit with the latest value', async () => {
        const emitted: string[] = [];
        const user = userEvent.setup();

        render(
            <SearchHost
                emitted={emitted}
                initialQuery="x"
            />,
            { wrapper: Wrapper },
        );

        await user.clear(getInput());
        await sleep(PAST_DEBOUNCE_MS);
        emitted.length = 0;

        await user.type(getInput(), 'abc');
        await sleep(PAST_DEBOUNCE_MS);

        expect(emitted).toEqual(['abc']);
    });

    it('drops a pending typed emit when Esc clears mid-debounce', async () => {
        const emitted: string[] = [];
        const user = userEvent.setup();

        render(<SearchHost emitted={emitted} />, { wrapper: Wrapper });
        await user.click(queryTrigger()!);
        await user.type(getInput(), 'ab');
        await user.keyboard('{Escape}');
        await sleep(PAST_DEBOUNCE_MS);

        expect(emitted.some((value) => value.length > 0)).toBe(false);
        expect(getInput().value).toBe('');
    });
});

describe('InputSearch — trailing clear button', () => {
    const queryClearButton = () => screen.queryByRole('button', { name: 'Clear search docs' });

    it('does not render the clear button when the field is empty', () => {
        render(<SearchHost emitted={[]} initialQuery="" />, { wrapper: Wrapper });
        expect(queryClearButton()).not.toBeInTheDocument();
    });

    it('renders the clear button on mount when deep-linked with a non-empty query', () => {
        render(<SearchHost emitted={[]} initialQuery="jwt" />, { wrapper: Wrapper });
        expect(queryClearButton()).toBeInTheDocument();
    });

    it('shows the clear button after typing, hides it once the field is empty again', async () => {
        const emitted: string[] = [];
        const user = userEvent.setup();
        render(<SearchHost emitted={emitted} initialQuery="" />, { wrapper: Wrapper });

        await user.click(queryTrigger()!);
        expect(queryClearButton()).not.toBeInTheDocument();

        await user.type(getInput(), 'foo');
        expect(queryClearButton()).toBeInTheDocument();

        await user.clear(getInput());
        expect(queryClearButton()).not.toBeInTheDocument();
    });

    it('clears the value, emits "" upstream, and keeps focus on the input', async () => {
        const emitted: string[] = [];
        const user = userEvent.setup();
        render(<SearchHost emitted={emitted} initialQuery="jwt" />, { wrapper: Wrapper });

        const input = getInput();
        await user.click(queryClearButton()!);
        await sleep(PAST_DEBOUNCE_MS);

        expect(input.value).toBe('');
        expect(emitted.at(-1)).toBe('');
        // Focus returned to the input — the programmatic-clear collapse
        // effect should NOT fire, so the trigger button stays absent.
        expect(document.activeElement).toBe(input);
        expect(queryTrigger()).not.toBeInTheDocument();
    });

    it('drops a pending debounce when clicked mid-typing', async () => {
        const emitted: string[] = [];
        const user = userEvent.setup();
        render(<SearchHost emitted={emitted} />, { wrapper: Wrapper });

        await user.click(queryTrigger()!);
        await user.type(getInput(), 'ab');
        // Clear button appears as soon as a character is typed.
        const clearBtn = queryClearButton();
        expect(clearBtn).toBeInTheDocument();

        await user.click(clearBtn!);
        await sleep(PAST_DEBOUNCE_MS);

        expect(emitted.some((value) => value.length > 0)).toBe(false);
        expect(getInput().value).toBe('');
    });
});

describe('InputSearch — Escape semantics', () => {
    it('first Esc clears a non-empty query but keeps the input expanded', async () => {
        const emitted: string[] = [];
        const user = userEvent.setup();

        render(
            <SearchHost
                emitted={emitted}
                initialQuery="jwt"
            />,
            { wrapper: Wrapper },
        );
        const input = getInput();
        input.focus();

        await user.keyboard('{Escape}');
        await sleep(50);

        expect(input.value).toBe('');
        expect(emitted.at(-1)).toBe('');
        expect(document.activeElement).toBe(input);
        expect(queryTrigger()).not.toBeInTheDocument();
    });

    it('second Esc on an already-empty query collapses the input', async () => {
        const emitted: string[] = [];
        const user = userEvent.setup();

        render(
            <SearchHost
                emitted={emitted}
                initialQuery="jwt"
            />,
            { wrapper: Wrapper },
        );
        const input = getInput();
        input.focus();

        await user.keyboard('{Escape}');
        await sleep(50);
        await user.keyboard('{Escape}');
        await sleep(50);

        expect(queryTrigger()).toBeInTheDocument();
    });
});

describe('InputSearch — external source-of-truth wins', () => {
    it('snaps to an external value mid-debounce and drops the in-flight typed emit', async () => {
        const emitted: string[] = [];
        let setExternal: ((next: string) => void) | undefined;
        const user = userEvent.setup();

        render(
            <SearchHost
                emitted={emitted}
                initialQuery=""
                onSetExternalRef={(setter) => {
                    setExternal = setter;
                }}
            />,
            { wrapper: Wrapper },
        );

        await user.click(queryTrigger()!);
        await user.type(getInput(), 'ab');

        // External source mutates the prop while the debounce is still pending.
        expect(setExternal).toBeDefined();
        act(() => setExternal!('external'));
        await sleep(PAST_DEBOUNCE_MS);

        expect(emitted).not.toContain('ab');
        expect(getInput().value).toBe('external');
    });
});

describe('InputSearch — global hotkey expansion', () => {
    it('Ctrl+F expands the collapsed input', () => {
        render(<SearchHost emitted={[]} />, { wrapper: Wrapper });

        expect(queryTrigger()).toBeInTheDocument();

        act(() => {
            document.dispatchEvent(new KeyboardEvent('keydown', { bubbles: true, ctrlKey: true, key: 'f' }));
        });

        expect(queryTrigger()).not.toBeInTheDocument();
    });

    it('Cmd+F (metaKey) also expands', () => {
        render(<SearchHost emitted={[]} />, { wrapper: Wrapper });

        act(() => {
            document.dispatchEvent(new KeyboardEvent('keydown', { bubbles: true, key: 'f', metaKey: true }));
        });

        expect(queryTrigger()).not.toBeInTheDocument();
    });
});
