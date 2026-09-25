import { render, renderHook, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { UpdateState } from '@/graphql/types';

const apollo = vi.hoisted(() => ({
    data: undefined as undefined | { versionInfo: { state: UpdateState } },
    startPolling: vi.fn(),
    stopPolling: vi.fn(),
}));
const user = vi.hoisted(() => ({ authenticated: true }));

vi.mock('@apollo/client/react', () => ({
    useQuery: () => ({
        data: apollo.data,
        loading: false,
        startPolling: apollo.startPolling,
        stopPolling: apollo.stopPolling,
    }),
}));
vi.mock('@/providers/user-provider', () => ({ useUser: () => ({ isAuthenticated: () => user.authenticated }) }));

import {
    useVersionInfo,
    VERSION_PENDING_POLL_INTERVAL_MS,
    VERSION_POLL_INTERVAL_MS,
    VersionInfoProvider,
} from './version-info-provider';

function Probe() {
    const { versionInfo } = useVersionInfo();

    return <span data-slot="state">{versionInfo?.state ?? 'none'}</span>;
}

describe('VersionInfoProvider', () => {
    beforeEach(() => {
        apollo.data = undefined;
        apollo.startPolling.mockClear();
        apollo.stopPolling.mockClear();
        user.authenticated = true;
    });

    it('polls quickly while the first verdict is pending, since it lands within minutes of a restart', () => {
        apollo.data = { versionInfo: { state: UpdateState.Pending } };
        render(
            <VersionInfoProvider>
                <Probe />
            </VersionInfoProvider>,
        );

        expect(screen.getByTestId('state')).toHaveTextContent('pending');
        expect(apollo.startPolling).toHaveBeenLastCalledWith(VERSION_PENDING_POLL_INTERVAL_MS);
    });

    it('slows down once there is a verdict, which changes at most every few hours', () => {
        apollo.data = { versionInfo: { state: UpdateState.UpToDate } };
        render(
            <VersionInfoProvider>
                <Probe />
            </VersionInfoProvider>,
        );

        expect(apollo.startPolling).toHaveBeenLastCalledWith(VERSION_POLL_INTERVAL_MS);
    });

    it('does not poll for a visitor who is not signed in', () => {
        user.authenticated = false;
        render(
            <VersionInfoProvider>
                <Probe />
            </VersionInfoProvider>,
        );

        expect(screen.getByTestId('state')).toHaveTextContent('none');
        expect(apollo.startPolling).not.toHaveBeenCalled();
    });

    it('stops polling when it unmounts', () => {
        apollo.data = { versionInfo: { state: UpdateState.UpToDate } };
        const { unmount } = render(
            <VersionInfoProvider>
                <Probe />
            </VersionInfoProvider>,
        );

        unmount();

        expect(apollo.stopPolling).toHaveBeenCalled();
    });

    it('refuses to be read outside the provider', () => {
        expect(() => renderHook(() => useVersionInfo())).toThrow(
            'useVersionInfo must be used within a VersionInfoProvider',
        );
    });
});
