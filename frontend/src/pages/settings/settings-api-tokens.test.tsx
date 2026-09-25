import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { addDays, differenceInCalendarMonths, format, startOfDay } from 'date-fns';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

const queryResult = vi.hoisted(() => ({
    current: {
        data: {
            apiTokens: [
                {
                    createdAt: '2026-01-15T00:00:00Z',
                    id: '1',
                    name: 'Seeded Token',
                    roleId: '0',
                    status: 'active',
                    tokenId: 'tok-1',
                    ttl: 86400,
                    updatedAt: '2026-01-15T00:00:00Z',
                    userId: '0',
                },
            ],
        },
        error: undefined,
        loading: false,
        refetch: () => {},
    },
}));

const mutationSpy = vi.hoisted(() => vi.fn());

vi.mock('@apollo/client/react', () => ({
    useMutation: () => [mutationSpy, {}],
    useQuery: () => queryResult.current,
    useSubscription: () => ({}),
}));

vi.mock('@/hooks/use-table-state', () => ({
    useTableState: () => ({ filter: '', pageIndex: 0, setFilter: vi.fn(), setPage: vi.fn() }),
}));

vi.mock('@/components/layouts/app/app-header', () => {
    const Pass = ({ children }: { children?: React.ReactNode }) => <div>{children}</div>;
    const Action = ({
        icon: _icon,
        label,
        loading: _loading,
        ...props
    }: React.ComponentProps<'button'> & {
        icon?: React.ReactNode;
        label?: React.ReactNode;
        loading?: boolean;
    }) => <button {...props}>{label}</button>;

    return {
        AppHeader: Pass,
        AppHeaderAction: Action,
        AppHeaderActions: Pass,
        AppHeaderContent: Pass,
        AppHeaderTitle: Pass,
    };
});

vi.mock('sonner', () => ({ toast: { error: vi.fn(), success: vi.fn() } }));

import { TokenStatus } from '@/graphql/types';

import SettingsAPITokens, {
    buildUpdateTokenInput,
    calculateTTL,
    getLastSelectableExpiry,
    MAX_TOKEN_TTL_SECONDS,
    MIN_TOKEN_TTL_SECONDS,
    tokenNameSchema,
} from './settings-api-tokens';

describe('tokenNameSchema', () => {
    it('accepts a name at the 100-character boundary', () => {
        expect(tokenNameSchema.safeParse('a'.repeat(100)).success).toBe(true);
    });

    it('rejects a name one character over the 100-character cap', () => {
        expect(tokenNameSchema.safeParse('a'.repeat(101)).success).toBe(false);
    });

    it('counts astral characters once, the way the endpoint does', () => {
        expect(tokenNameSchema.safeParse('\u{1f511}'.repeat(100)).success).toBe(true);
        expect(tokenNameSchema.safeParse('\u{1f511}'.repeat(101)).success).toBe(false);
    });

    it('trims surrounding whitespace before length-checking', () => {
        expect(tokenNameSchema.safeParse(`  ${'a'.repeat(100)}  `)).toMatchObject({
            data: 'a'.repeat(100),
            success: true,
        });
    });

    it('defaults an omitted name to an empty string', () => {
        expect(tokenNameSchema.safeParse(undefined)).toMatchObject({ data: '', success: true });
    });
});

describe('expiry bounds offered by the picker', () => {
    const now = new Date('2026-08-05T12:34:56');

    it('turns the last selectable day into a ttl the endpoint accepts', () => {
        expect(calculateTTL(getLastSelectableExpiry(now), now)).toBeLessThanOrEqual(MAX_TOKEN_TTL_SECONDS);
    });

    it('leaves the next day out of range, so the bound is not off by a day', () => {
        expect(calculateTTL(addDays(getLastSelectableExpiry(now), 1), now)).toBeGreaterThan(MAX_TOKEN_TTL_SECONDS);
    });

    it('floors the ttl at the endpoint minimum when tomorrow is under a minute away', () => {
        const lastMinuteOfTheDay = new Date('2026-08-05T23:59:30');

        expect(calculateTTL(startOfDay(addDays(lastMinuteOfTheDay, 1)), lastMinuteOfTheDay)).toBe(
            MIN_TOKEN_TTL_SECONDS,
        );
    });

    it('sends whole seconds for an ordinary pick', () => {
        expect(calculateTTL(startOfDay(addDays(now, 1)), now)).toBe(41104);
    });
});

describe('the create row date picker', () => {
    // A start date whose ceiling lands on a month end: the two boundary days then
    // sit in different grids.
    beforeEach(() => {
        vi.useFakeTimers({ shouldAdvanceTime: true });
        vi.setSystemTime(new Date('2026-09-01T12:00:00'));
    });

    afterEach(() => {
        vi.useRealTimers();
    });

    const openPicker = () => {
        shownMonth = new Date();
        render(
            <MemoryRouter>
                <SettingsAPITokens />
            </MemoryRouter>,
        );

        fireEvent.click(screen.getByRole('button', { name: 'Create Token' }));
        fireEvent.click(screen.getByRole('button', { name: 'Pick date' }));
    };

    let shownMonth = new Date();

    const dayButton = (date: Date) => {
        const step = differenceInCalendarMonths(date, shownMonth);
        const label = step < 0 ? 'Go to the Previous Month' : 'Go to the Next Month';

        for (let months = Math.abs(step); months > 0; months--) {
            fireEvent.click(screen.getByRole('button', { name: label }));
        }

        shownMonth = date;

        return within(screen.getByRole('grid')).getByRole('button', { name: format(date, 'PPPP') });
    };

    const submitButton = () => screen.getByRole('button', { name: 'Submit' });

    it('offers the last day the endpoint accepts and no day past it', async () => {
        const lastAllowed = getLastSelectableExpiry();
        const firstRefused = addDays(lastAllowed, 1);

        openPicker();

        const refused = dayButton(firstRefused);

        expect(refused).toBeDisabled();

        fireEvent.click(refused);
        expect(screen.getByRole('button', { name: 'Pick date' })).toBeInTheDocument();

        fireEvent.click(dayButton(lastAllowed));
        expect(screen.getByRole('button', { name: format(lastAllowed, 'd MMM yyyy') })).toBeInTheDocument();
        await waitFor(() => expect(submitButton()).toBeEnabled());
    }, 30_000);
});

describe('buildUpdateTokenInput', () => {
    it('omits a status the user did not touch', () => {
        expect(
            buildUpdateTokenInput({ status: TokenStatus.Expired }, { name: 'renamed', status: TokenStatus.Expired }),
        ).toEqual({
            name: 'renamed',
        });
    });

    it('sends the status once the user picked another one', () => {
        expect(
            buildUpdateTokenInput({ status: TokenStatus.Active }, { name: 'renamed', status: TokenStatus.Revoked }),
        ).toEqual({
            name: 'renamed',
            status: TokenStatus.Revoked,
        });
    });

    it('sends an empty name so clearing it reaches the resolver', () => {
        expect(
            buildUpdateTokenInput({ status: TokenStatus.Active }, { name: '   ', status: TokenStatus.Active }),
        ).toEqual({
            name: '',
        });
    });
});

describe('a revoked token that also outlived its ttl', () => {
    beforeEach(() => {
        queryResult.current = {
            ...queryResult.current,
            data: {
                apiTokens: [
                    {
                        createdAt: '2026-01-15T00:00:00Z',
                        id: '1',
                        name: 'Seeded Token',
                        roleId: '0',
                        status: 'revoked',
                        tokenId: 'tok-1',
                        ttl: 60,
                        updatedAt: '2026-01-15T00:00:00Z',
                        userId: '0',
                    },
                ],
            },
        };
    });

    it('reads as revoked, the way the server reports it', () => {
        render(
            <MemoryRouter>
                <SettingsAPITokens />
            </MemoryRouter>,
        );

        expect(screen.getByText('revoked')).toBeInTheDocument();
        expect(screen.queryByText('expired')).not.toBeInTheDocument();
    });
});

describe('a token created in an earlier year', () => {
    beforeEach(() => {
        queryResult.current = {
            ...queryResult.current,
            data: {
                apiTokens: [
                    {
                        createdAt: '2025-06-15T12:00:00Z',
                        id: '1',
                        name: 'Seeded Token',
                        roleId: '0',
                        status: 'active',
                        tokenId: 'tok-1',
                        ttl: 86400,
                        updatedAt: '2025-06-15T12:00:00Z',
                        userId: '0',
                    },
                ],
            },
        };
    });

    it('dates it without the time', () => {
        render(
            <MemoryRouter>
                <SettingsAPITokens />
            </MemoryRouter>,
        );

        expect(screen.getByText(format(new Date('2025-06-15T12:00:00Z'), 'd MMM yyyy'))).toBeInTheDocument();
        expect(screen.getByText(format(new Date('2025-06-16T12:00:00Z'), 'd MMM yyyy'))).toBeInTheDocument();
    });
});

describe('creating the very first token', () => {
    beforeEach(() => {
        vi.useFakeTimers({ shouldAdvanceTime: true });
        vi.setSystemTime(new Date('2026-09-01T12:00:00'));
        mutationSpy.mockReset();
        mutationSpy.mockResolvedValue({ data: { createAPIToken: { token: 'pk-secret-value' } } });
        queryResult.current = { ...queryResult.current, data: { apiTokens: [] } };
    });

    afterEach(() => {
        vi.useRealTimers();
    });

    it('reveals the one-time secret even though the list is still empty', async () => {
        render(
            <MemoryRouter>
                <SettingsAPITokens />
            </MemoryRouter>,
        );

        fireEvent.click(screen.getAllByRole('button', { name: 'Create Token' })[0]!);
        fireEvent.click(screen.getByRole('button', { name: 'Pick date' }));
        fireEvent.click(
            within(screen.getByRole('grid')).getByRole('button', {
                name: format(addDays(new Date('2026-09-01T12:00:00'), 7), 'PPPP'),
            }),
        );

        const submit = screen.getByRole('button', { name: 'Submit' });

        await waitFor(() => expect(submit).toBeEnabled());
        fireEvent.click(submit);

        await waitFor(() => expect(mutationSpy).toHaveBeenCalled());
        expect(await screen.findByText('API Token Created')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'Copy Token' })).toBeInTheDocument();
    }, 30_000);
});

describe('renaming an expired token', () => {
    beforeEach(() => {
        mutationSpy.mockClear();
        queryResult.current = {
            ...queryResult.current,
            data: {
                apiTokens: [
                    {
                        createdAt: '2026-01-15T00:00:00Z',
                        id: '1',
                        name: 'Seeded Token',
                        roleId: '0',
                        status: 'expired',
                        tokenId: 'tok-1',
                        ttl: 60,
                        updatedAt: '2026-01-15T00:00:00Z',
                        userId: '0',
                    },
                ],
            },
        };
    });

    it('sends the new name without echoing the derived status back', async () => {
        render(
            <MemoryRouter>
                <SettingsAPITokens />
            </MemoryRouter>,
        );

        fireEvent.pointerDown(
            screen.getByRole('button', { name: 'Open menu' }),
            new MouseEvent('pointerdown', { bubbles: true, cancelable: true }),
        );
        fireEvent.click(await screen.findByRole('menuitem', { name: 'Edit' }));

        fireEvent.change(screen.getByPlaceholderText('Token name (optional)'), { target: { value: 'renamed' } });

        const submit = screen.getByRole('button', { name: 'Submit' });

        await waitFor(() => expect(submit).toBeEnabled());
        fireEvent.click(submit);

        await waitFor(() => expect(mutationSpy).toHaveBeenCalled());

        expect(mutationSpy.mock.calls[0]![0].variables).toEqual({
            input: { name: 'renamed' },
            tokenId: 'tok-1',
        });
    });
});
