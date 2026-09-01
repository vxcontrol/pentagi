import type { ReactNode } from 'react';

import { useQuery } from '@apollo/client/react';
import { createContext, use, useEffect, useMemo } from 'react';

import type { VersionInfoFragmentFragment } from '@/graphql/types';

import { UpdateState, VersionInfoDocument } from '@/graphql/types';
import { useUser } from '@/providers/user-provider';

// The server asks the update service every few hours, so most polls return the
// answer they got last time. Polling exists for two moments: the minutes right
// after a restart, when the verdict is still pending and arrives within two
// minutes, and the once-in-a-while moment a new release lands mid-session.
export const VERSION_POLL_INTERVAL_MS = 15 * 60 * 1000;
export const VERSION_PENDING_POLL_INTERVAL_MS = 30 * 1000;

interface VersionInfoContextType {
    isLoading: boolean;
    versionInfo: null | VersionInfoFragmentFragment;
}

const VersionInfoContext = createContext<undefined | VersionInfoContextType>(undefined);

export function useVersionInfo() {
    const context = use(VersionInfoContext);

    if (context === undefined) {
        throw new Error('useVersionInfo must be used within a VersionInfoProvider');
    }

    return context;
}

export function VersionInfoProvider({ children }: { children: ReactNode }) {
    const { isAuthenticated } = useUser();
    const skip = !isAuthenticated();

    const { data, loading, startPolling, stopPolling } = useQuery(VersionInfoDocument, { skip });
    const versionInfo = data?.versionInfo ?? null;
    const state = versionInfo?.state;

    useEffect(() => {
        if (skip) {
            return undefined;
        }

        startPolling(state === UpdateState.Pending ? VERSION_PENDING_POLL_INTERVAL_MS : VERSION_POLL_INTERVAL_MS);

        return () => stopPolling();
    }, [skip, state, startPolling, stopPolling]);

    const value = useMemo<VersionInfoContextType>(
        () => ({
            isLoading: loading,
            versionInfo,
        }),
        [loading, versionInfo],
    );

    return <VersionInfoContext value={value}>{children}</VersionInfoContext>;
}
