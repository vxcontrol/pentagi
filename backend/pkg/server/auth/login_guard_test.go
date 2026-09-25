package auth

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestGuard(limit int, window, lockout time.Duration) (*LoginGuard, func(time.Duration)) {
	clock := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	guard := NewLoginGuard(limit, window, lockout)
	guard.now = func() time.Time { return clock }

	return guard, func(d time.Duration) { clock = clock.Add(d) }
}

func takeN(t *testing.T, guard *LoginGuard, key string, n int) {
	t.Helper()

	for i := 0; i < n; i++ {
		guard.Take(key)
	}
}

func TestLoginGuard_Take_CountsAttemptsPerKeyWithinTheWindow(t *testing.T) {
	cases := []struct {
		name             string
		limit            int
		window, lockout  time.Duration
		before           func(t *testing.T, guard *LoginGuard, advance func(time.Duration))
		key              string
		wantOK           bool
		wantRetryAfter   time.Duration
		wantTrackedCount int
	}{
		{
			name: "the attempt past the limit is locked out", limit: 3, window: time.Minute, lockout: 10 * time.Minute,
			before: func(t *testing.T, guard *LoginGuard, _ func(time.Duration)) {
				for i := 1; i <= 3; i++ {
					_, ok := guard.Take("victim@example.com|10.0.0.1")
					require.True(t, ok, "attempt %d is within the limit", i)
				}
			},
			key: "victim@example.com|10.0.0.1", wantOK: false, wantRetryAfter: 10 * time.Minute, wantTrackedCount: 1,
		},
		{
			name: "a lockout lifts on its own", limit: 3, window: time.Minute, lockout: 10 * time.Minute,
			before: func(t *testing.T, guard *LoginGuard, advance func(time.Duration)) {
				takeN(t, guard, "victim@example.com|10.0.0.1", 4)
				advance(10 * time.Minute)
			},
			key: "victim@example.com|10.0.0.1", wantOK: true, wantTrackedCount: 1,
		},
		{
			name: "the same account from another address stays free", limit: 2, window: time.Minute, lockout: time.Minute,
			before: func(t *testing.T, guard *LoginGuard, _ func(time.Duration)) {
				takeN(t, guard, "locked@example.com|10.0.0.1", 3)
				_, ok := guard.Take("locked@example.com|10.0.0.1")
				require.False(t, ok)
			},
			key: "locked@example.com|10.0.0.2", wantOK: true, wantTrackedCount: 2,
		},
		{
			name: "attempts spread past the window do not add up", limit: 3, window: time.Minute, lockout: time.Hour,
			before: func(t *testing.T, guard *LoginGuard, advance func(time.Duration)) {
				takeN(t, guard, "slow@example.com", 2)
				advance(2 * time.Minute)
				takeN(t, guard, "slow@example.com", 2)
			},
			key: "slow@example.com", wantOK: true, wantTrackedCount: 1,
		},
		{
			name: "an empty key is never counted", limit: 1, window: time.Minute, lockout: time.Hour,
			before: func(*testing.T, *LoginGuard, func(time.Duration)) {},
			key:    "", wantOK: true, wantTrackedCount: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			guard, advance := newTestGuard(tc.limit, tc.window, tc.lockout)
			tc.before(t, guard, advance)

			retryAfter, ok := guard.Take(tc.key)

			assert.Equal(t, tc.wantOK, ok)
			assert.Equal(t, tc.wantRetryAfter, retryAfter)
			assert.Len(t, guard.entries, tc.wantTrackedCount)
		})
	}
}

func TestLoginGuard_Take_CountsConcurrentAttemptsOnce(t *testing.T) {
	guard, _ := newTestGuard(10, time.Minute, time.Hour)

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		allowed int
	)
	for range 200 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, ok := guard.Take("burst@example.com|10.0.0.1"); ok {
				mu.Lock()
				allowed++
				mu.Unlock()
			}
		}()
	}
	authWaitOrFail(t, &wg, "the burst")

	assert.Equal(t, 10, allowed,
		"a burst has to be counted as it arrives, not after each password check returns")
}

func TestLoginGuard_Reset_ClearsTheCount(t *testing.T) {
	guard, _ := newTestGuard(3, time.Minute, time.Hour)

	takeN(t, guard, "user@example.com", 2)
	guard.Reset("user@example.com")
	takeN(t, guard, "user@example.com", 2)

	_, ok := guard.Take("user@example.com")
	assert.True(t, ok, "a successful login has to clear what came before it")
}

func TestLoginGuard_Relax_HalvesTheCountButKeepsALockout(t *testing.T) {
	cases := []struct {
		name         string
		limit        int
		lockout      time.Duration
		takenBefore  int
		allowedAfter int // attempts that must still fit after Relax; the next one must not
	}{
		{"eight halved to four leave six attempts under a limit of ten", 10, time.Hour, 8, 6},
		{"a lockout already in force survives", 3, 10 * time.Minute, 4, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			guard, _ := newTestGuard(tc.limit, time.Minute, tc.lockout)
			takeN(t, guard, "10.0.0.1", tc.takenBefore)

			guard.Relax("10.0.0.1")

			for i := 1; i <= tc.allowedAfter; i++ {
				_, ok := guard.Take("10.0.0.1")
				require.True(t, ok, "attempt %d after Relax", i)
			}
			_, ok := guard.Take("10.0.0.1")
			assert.False(t, ok)
		})
	}
}

func TestLoginGuard_EntryLocked_KeepsCountingWhenTheTableIsFull(t *testing.T) {
	guard, _ := newTestGuard(3, time.Minute, time.Hour)

	for i := 0; i < maxGuardEntries*2; i++ {
		guard.Take(fmt.Sprintf("flood-%d@example.com", i))
	}

	require.LessOrEqual(t, len(guard.entries), maxGuardEntries,
		"an attacker inventing keys must not grow the table without bound")

	takeN(t, guard, "victim@example.com|10.0.0.1", 3)
	_, ok := guard.Take("victim@example.com|10.0.0.1")

	assert.False(t, ok, "filling the table must not switch the guard off for the account being attacked")
}

func TestLoginGuard_EvictCheapestLocked_KeepsACounterAtTheLimitThroughAFlood(t *testing.T) {
	guard, advance := newTestGuard(3, time.Hour, time.Hour)

	takeN(t, guard, "victim@example.com|10.0.0.1", 3)
	advance(time.Minute)
	for i := 0; i < maxGuardEntries*2; i++ {
		guard.Take(fmt.Sprintf("flood-%d@example.com|10.0.0.9", i))
	}

	retryAfter, ok := guard.Take("victim@example.com|10.0.0.1")
	assert.False(t, ok, "a flood of fresh keys must not hand the attacker a new budget")
	assert.Equal(t, time.Hour, retryAfter)
}

func TestLoginGuard_EvictCheapestLocked_KeepsTheCostlierKeyWhateverTheMapOrder(t *testing.T) {
	cases := []struct {
		name  string
		limit int
		setup func(guard *LoginGuard, advance func(time.Duration))
		kept  string
	}{
		{"more attempts outlast fewer however old", 5, func(guard *LoginGuard, advance func(time.Duration)) {
			for range 3 {
				guard.Take("counting@example.com|10.0.0.1")
			}
			advance(time.Minute)
			guard.Take("fresh@example.com|10.0.0.9")
		}, "counting@example.com|10.0.0.1"},
		{"a lockout outlasts any count", 1, func(guard *LoginGuard, _ func(time.Duration)) {
			guard.Take("locked@example.com|10.0.0.1")
			guard.Take("locked@example.com|10.0.0.1")
			guard.Take("counting@example.com|10.0.0.9")
		}, "locked@example.com|10.0.0.1"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for range 64 {
				guard, advance := newTestGuard(tc.limit, time.Hour, time.Hour)
				tc.setup(guard, advance)
				require.Len(t, guard.entries, 2)

				guard.mu.Lock()
				guard.evictCheapestLocked()
				_, isKept := guard.entries[tc.kept]
				guard.mu.Unlock()

				require.True(t, isKept, "the eviction must not depend on which key the map yields first")
			}
		})
	}
}
