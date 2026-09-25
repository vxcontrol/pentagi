import { renderHook } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

const viewer = { id: 0 };

vi.mock('@/providers/user-provider', () => ({
    useUser: () => ({ authInfo: viewer.id ? { user: { id: viewer.id } } : null }),
}));

import { useIsOwnFlow } from './use-is-own-flow';

const isOwn = (flow: null | { userId: number | string }) => renderHook(() => useIsOwnFlow()).result.current(flow);

describe('useIsOwnFlow', () => {
    it('owns the flow whose userId matches the viewer', () => {
        viewer.id = 11;

        expect(isOwn({ userId: '11' })).toBe(true);
    });

    it('owns the flow when its userId arrives as a number, the way the API serializes an ID', () => {
        viewer.id = 11;

        expect(isOwn({ userId: 11 })).toBe(true);
        expect(isOwn({ userId: 1 })).toBe(false);
    });

    it('does not own another account flow', () => {
        viewer.id = 11;

        expect(isOwn({ userId: '1' })).toBe(false);
    });

    it('owns nothing while no one is signed in', () => {
        viewer.id = 0;

        expect(isOwn({ userId: '1' })).toBe(false);
    });

    it('owns nothing when there is no flow', () => {
        viewer.id = 11;

        expect(isOwn(null)).toBe(false);
    });
});
