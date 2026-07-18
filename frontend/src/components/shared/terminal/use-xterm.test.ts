import { describe, expect, it } from 'vitest';

import { writeWithFlowControl } from './use-xterm';

const CHUNK = 64 * 1024;

// A terminal mock that defers each chunk's flow-control callback, so a clear() can
// interleave between chunks — the exact window use-xterm's generation token closes.
const deferredTerminal = () => {
    const writes: string[] = [];
    const pending: Array<() => void> = [];

    return {
        drainNext: () => pending.shift()?.(),
        pendingCount: () => pending.length,
        terminal: {
            write: (chunk: string, callback?: () => void) => {
                writes.push(chunk);

                if (callback) {
                    pending.push(callback);
                }
            },
        } as unknown as Parameters<typeof writeWithFlowControl>[0],
        writes,
    };
};

describe('writeWithFlowControl', () => {
    it('stops draining once the write goes stale (a clear cancels in-flight chunks)', () => {
        const mock = deferredTerminal();
        const data = 'a'.repeat(CHUNK * 3 + 100);

        let isStale = false;
        writeWithFlowControl(mock.terminal, data, () => isStale);

        expect(mock.writes.length).toBe(1);
        mock.drainNext();
        expect(mock.writes.length).toBe(2);

        isStale = true;
        const writtenBeforeClear = mock.writes.length;

        while (mock.pendingCount() > 0) {
            mock.drainNext();
        }

        expect(mock.writes.length).toBe(writtenBeforeClear);
    });

    it('drains the whole payload intact when never stale', () => {
        const mock = deferredTerminal();
        const data = 'b'.repeat(CHUNK * 3 + 100);

        writeWithFlowControl(mock.terminal, data, () => false);

        while (mock.pendingCount() > 0) {
            mock.drainNext();
        }

        expect(mock.writes.join('')).toBe(data);
    });
});
