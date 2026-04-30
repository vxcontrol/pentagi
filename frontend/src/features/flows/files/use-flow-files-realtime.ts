import { useMemo } from 'react';

import {
    useFlowFileAddedSubscription,
    useFlowFileDeletedSubscription,
    useFlowFileUpdatedSubscription,
} from '@/graphql/types';

interface UseFlowFilesRealtimeParams {
    flowId: null | string;
    /**
     * When `true`, every subscription is skipped. The flag is needed to delay subscription
     * delivery until the initial query has populated the Apollo cache so that the cache
     * field exists before the subscription updates arrive.
     */
    isPaused: boolean;
}

/**
 * Wires up the three flow-file subscriptions (added / updated / deleted) under a single
 * call so the parent component does not need to repeat the same `skip`/`variables` pair
 * three times.
 */
export const useFlowFilesRealtime = ({ flowId, isPaused }: UseFlowFilesRealtimeParams): void => {
    const variables = useMemo(() => ({ flowId: flowId ?? '' }), [flowId]);
    const isSkipped = isPaused || !flowId;

    useFlowFileAddedSubscription({ skip: isSkipped, variables });
    useFlowFileUpdatedSubscription({ skip: isSkipped, variables });
    useFlowFileDeletedSubscription({ skip: isSkipped, variables });
};
