import { render } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

const xterm = vi.hoisted(() => {
    const state = {
        clear() {
            state.clearCount += 1;
            state.visible = '';
        },
        clearCount: 0,
        scrollToBottom() {},
        visible: '',
        write(chunk: string) {
            state.visible += chunk;
        },
    };

    return state;
});

vi.mock('./use-xterm', () => ({
    useXterm: () => ({
        clear: xterm.clear,
        containerRef: { current: null },
        isReady: true,
        scrollToBottom: xterm.scrollToBottom,
        searchAddon: null,
        write: xterm.write,
    }),
}));

vi.mock('./use-terminal-search', () => ({
    useTerminalSearch: () => ({ findNext: () => {}, findPrevious: () => {} }),
}));

vi.mock('@/hooks/use-theme', () => ({ useTheme: () => ({ theme: 'dark' }) }));

vi.mock('./terminal-sanitizer', () => ({ processLog: (line: string) => line }));

import Terminal from './terminal';

const visibleLines = () => {
    const lines = xterm.visible.split('\r\n');

    // Keep interior blank lines (garbling can surface as a stray \r\n); drop only the trailing terminator.
    if (lines.at(-1) === '') {
        lines.pop();
    }

    return lines;
};

const reset = () => {
    xterm.visible = '';
    xterm.clearCount = 0;
};

describe('Terminal incremental writer', () => {
    it('restores the full log set after a filter narrows then clears', () => {
        reset();
        const full = ['alpha-0', 'bravo-1', 'alpha-2', 'bravo-3', 'alpha-4'];
        const filtered = ['bravo-1', 'bravo-3'];

        const { rerender } = render(<Terminal logs={full} />);
        expect(visibleLines()).toEqual(full);

        rerender(<Terminal logs={filtered} />);
        expect(visibleLines()).toEqual(filtered);

        rerender(<Terminal logs={full} />);
        expect(visibleLines()).toEqual(full);
    });

    it('appends streamed lines incrementally without clearing', () => {
        reset();
        const { rerender } = render(<Terminal logs={['line-0']} />);

        rerender(<Terminal logs={['line-0', 'line-1']} />);
        rerender(<Terminal logs={['line-0', 'line-1', 'line-2']} />);

        expect(visibleLines()).toEqual(['line-0', 'line-1', 'line-2']);
        expect(xterm.clearCount).toBe(0);
    });

    it('rewrites when the filter changes to a different, non-prefix set', () => {
        reset();
        const full = ['keep-0', 'drop-1', 'keep-2', 'drop-3'];

        const { rerender } = render(<Terminal logs={full} />);
        rerender(<Terminal logs={['keep-0', 'keep-2']} />);
        rerender(<Terminal logs={['drop-1', 'drop-3']} />);

        expect(visibleLines()).toEqual(['drop-1', 'drop-3']);
    });
});
