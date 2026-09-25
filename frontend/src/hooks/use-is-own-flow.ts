import { useCallback } from 'react';

import { useUser } from '@/providers/user-provider';

// The flows query answers globally for an administrator while every provider
// query answers per user, so the two lists are not about the same person.
export function useIsOwnFlow() {
    const { authInfo } = useUser();
    const viewerId = authInfo?.user?.id;

    return useCallback(
        (flow?: null | { userId: number | string }) =>
            flow != null && viewerId != null && String(viewerId) === String(flow.userId),
        [viewerId],
    );
}
