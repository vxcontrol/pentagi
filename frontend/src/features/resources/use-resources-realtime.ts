import {
    useResourceAddedSubscription,
    useResourceDeletedSubscription,
    useResourceUpdatedSubscription,
} from '@/graphql/types';

interface UseResourcesRealtimeParams {
    /**
     * When `true`, every subscription is skipped. The flag is needed to delay
     * subscription delivery until the initial query has populated the Apollo cache
     * so that the cache field exists before the subscription updates arrive.
     */
    isPaused: boolean;
}

/**
 * Wires up the three user-resource subscriptions (added / updated / deleted).
 * The Apollo subscription-cache link in `lib/apollo.ts` writes incoming entries
 * straight into the `resources` query cache, so callers don't need extra refetch.
 */
export const useResourcesRealtime = ({ isPaused }: UseResourcesRealtimeParams): void => {
    useResourceAddedSubscription({ skip: isPaused });
    useResourceUpdatedSubscription({ skip: isPaused });
    useResourceDeletedSubscription({ skip: isPaused });
};
