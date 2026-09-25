import { ApolloClient, ApolloLink, Observable } from '@apollo/client';
import { ApolloProvider } from '@apollo/client/react';
import { act, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { toast } from 'sonner';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { createCache, watchQueryDefaults } from '@/lib/apollo';

import { FavoritesProvider, useFavorites } from './favorites-provider';

vi.mock('sonner', () => ({ toast: { error: vi.fn(), info: vi.fn(), success: vi.fn() } }));
vi.mock('@/providers/user-provider', () => ({
    useUser: () => ({ authInfo: { type: 'user' }, isAuthenticated: () => true }),
}));

const SIGNED_IN_USER_ID = '1';

const myFlow = { id: '5', userId: SIGNED_IN_USER_ID };
const theirFlow = { id: '6', userId: '2' };

const userPreferences = (favoriteFlows: string[]) => ({
    settingsUser: { __typename: 'UserPreferences', favoriteFlows, id: SIGNED_IN_USER_ID },
});

const buildClient = ({
    addRefusal = '',
    answeredPreferenceQueries = Number.POSITIVE_INFINITY,
    favoriteFlows = [] as string[],
    isPreferencesAnswering = true,
} = {}) => {
    const preferenceQueries: Array<string | undefined> = [];
    const heldAnswers: Array<() => void> = [];

    const link = new ApolloLink((operation) => {
        if (operation.operationName === 'addFavoriteFlow') {
            return new Observable<{ data: Record<string, unknown> }>((observer) => {
                if (addRefusal) {
                    observer.error(new Error(addRefusal));

                    return;
                }

                observer.next({ data: { addFavoriteFlow: 'Success' } });
                observer.complete();
            });
        }

        if (operation.operationName !== 'settingsUser') {
            return new Observable<{ data: Record<string, unknown> }>(() => undefined);
        }

        preferenceQueries.push(operation.operationName);

        return new Observable<{ data: Record<string, unknown> }>((observer) => {
            const answer = () => {
                observer.next({ data: userPreferences(favoriteFlows) });
                observer.complete();
            };

            if (!isPreferencesAnswering) {
                return;
            }

            if (preferenceQueries.length > answeredPreferenceQueries) {
                heldAnswers.push(answer);

                return;
            }

            answer();
        });
    });

    return {
        client: new ApolloClient({
            cache: createCache(),
            defaultOptions: { watchQuery: watchQueryDefaults },
            link,
        }),
        preferenceQueries,
        releaseHeldAnswers: () => heldAnswers.splice(0).forEach((answer) => answer()),
    };
};

function Probe() {
    const { canToggleFavorite, isLoading } = useFavorites();

    return (
        <>
            <span data-slot="probe-loading">{String(isLoading)}</span>
            <span data-slot="probe-mine">{String(canToggleFavorite(myFlow))}</span>
            <span data-slot="probe-theirs">{String(canToggleFavorite(theirFlow))}</span>
        </>
    );
}

const renderFavorites = (client: ApolloClient) =>
    render(
        <ApolloProvider client={client}>
            <FavoritesProvider>
                <Probe />
            </FavoritesProvider>
        </ApolloProvider>,
    );

describe('the favorite control gated on the signed-in identity', () => {
    it('offers no control at all while the preferences answer is in flight', async () => {
        const { client, preferenceQueries } = buildClient({ isPreferencesAnswering: false });

        renderFavorites(client);

        await waitFor(() => expect(preferenceQueries).toHaveLength(1));

        expect(screen.getByTestId('probe-loading')).toHaveTextContent('true');
        expect(screen.getByTestId('probe-theirs')).toHaveTextContent('false');
        expect(screen.getByTestId('probe-mine')).toHaveTextContent('false');
    });

    it('stays settled while a reconnect refetches the answer it already has', async () => {
        const { client, preferenceQueries, releaseHeldAnswers } = buildClient({ answeredPreferenceQueries: 1 });

        renderFavorites(client);

        await waitFor(() => expect(screen.getByTestId('probe-mine')).toHaveTextContent('true'));

        let refetched: Promise<unknown> = Promise.resolve();

        await act(async () => {
            refetched = client.refetchObservableQueries();
        });

        await waitFor(() => expect(preferenceQueries).toHaveLength(2));
        expect(screen.getByTestId('probe-loading')).toHaveTextContent('false');
        expect(screen.getByTestId('probe-mine')).toHaveTextContent('true');

        await act(async () => {
            releaseHeldAnswers();
            await refetched;
        });
    });

    it('offers the control on your own flow once the answer names you', async () => {
        const { client } = buildClient();

        renderFavorites(client);

        await waitFor(() => expect(screen.getByTestId('probe-mine')).toHaveTextContent('true'));
        expect(screen.getByTestId('probe-theirs')).toHaveTextContent('false');
    });

    it('keeps the control on a foreign flow that is already a favorite', async () => {
        const { client } = buildClient({ favoriteFlows: [theirFlow.id] });

        renderFavorites(client);

        await waitFor(() => expect(screen.getByTestId('probe-theirs')).toHaveTextContent('true'));
    });
});

const starRenders: string[] = [];

function StarProbe({ flowId }: { flowId: string }) {
    const { isFavoriteFlow, toggleFavoriteFlow } = useFavorites();
    const label = isFavoriteFlow(flowId) ? 'starred' : 'not starred';

    if (starRenders.at(-1) !== label) {
        starRenders.push(label);
    }

    return (
        <button
            data-slot="star"
            onClick={() => void toggleFavoriteFlow(flowId)}
        >
            {label}
        </button>
    );
}

const renderStar = (client: ApolloClient) =>
    render(
        <ApolloProvider client={client}>
            <FavoritesProvider>
                <StarProbe flowId={myFlow.id} />
            </FavoritesProvider>
        </ApolloProvider>,
    );

describe('a favorite the server refuses', () => {
    beforeEach(() => {
        vi.mocked(toast.error).mockClear();
        starRenders.length = 0;
    });

    it('lights the star, then puts it back and says the request was refused', async () => {
        const { client } = buildClient({ addRefusal: 'not permitted' });
        const user = userEvent.setup();

        renderStar(client);

        const star = () => screen.getByTestId('star');

        await waitFor(() => expect(star()).toHaveTextContent('not starred'));

        await act(() => user.click(star()));

        await waitFor(() =>
            expect(toast.error).toHaveBeenCalledWith(
                'Failed to add to favorites',
                expect.objectContaining({ description: expect.stringContaining('not permitted') }),
            ),
        );
        await waitFor(() => expect(star(), 'the refused star stayed lit').toHaveTextContent('not starred'));
        expect(starRenders, 'the click never lit the star, so the rollback proved nothing').toEqual([
            'not starred',
            'starred',
            'not starred',
        ]);
    });

    it('announces nothing when the server accepts', async () => {
        const { client } = buildClient();
        const user = userEvent.setup();

        renderStar(client);

        await waitFor(() => expect(screen.getByTestId('star')).toHaveTextContent('not starred'));

        await act(() => user.click(screen.getByTestId('star')));

        await waitFor(() => expect(starRenders).toContain('starred'));
        expect(toast.error).not.toHaveBeenCalled();
    });
});
