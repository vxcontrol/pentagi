package auth

import (
	"sync"
	"testing"
	"time"

	"pentagi/pkg/server/models"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type usersCacheRead struct {
	info UserInfo
	err  error
}

func usersCacheGet(cache *UserCache, userID uint64) usersCacheRead {
	info, err := cache.GetUser(userID)
	return usersCacheRead{info: info, err: err}
}

// A user blocked and revoked after it was cached, and a missing user created after its absence was.
var usersCacheEntries = []struct {
	name          string
	userID        uint64
	change        string
	before, after usersCacheRead
}{
	{"a stored user", 1, "UPDATE users SET status = 'blocked', session_generation = 2 WHERE id = 1",
		usersCacheRead{info: UserInfo{Hash: "testhash", Status: models.UserStatusActive, Generation: 1}},
		usersCacheRead{info: UserInfo{Hash: "testhash", Status: models.UserStatusBlocked, Generation: 2}}},
	{"a missing user", 9999, "INSERT INTO users (id, hash, mail) VALUES (9999, 'newhash', 'new@example.com')",
		usersCacheRead{err: gorm.ErrRecordNotFound},
		usersCacheRead{info: UserInfo{Hash: "newhash", Status: models.UserStatusActive, Generation: 1}}},
}

func TestUsersCache_GetUser_ServesTheCachedEntryUntilInvalidated(t *testing.T) {
	for _, tc := range usersCacheEntries {
		t.Run(tc.name, func(t *testing.T) {
			db := setupTestDB(t)
			cache := NewUserCache(db)

			require.Equal(t, tc.before, usersCacheGet(cache, tc.userID))
			require.NoError(t, db.Exec(tc.change).Error)
			assert.Equal(t, tc.before, usersCacheGet(cache, tc.userID), "the store changed under a cached entry")

			cache.Invalidate(tc.userID)
			assert.Equal(t, tc.after, usersCacheGet(cache, tc.userID))
		})
	}
}

func TestUsersCache_GetUser_RereadsAnEntryPastItsTTL(t *testing.T) {
	for _, tc := range usersCacheEntries {
		t.Run(tc.name, func(t *testing.T) {
			db := setupTestDB(t)
			cache := NewUserCache(db)
			cache.SetTTL(50 * time.Millisecond)

			require.Equal(t, tc.before, usersCacheGet(cache, tc.userID))
			require.NoError(t, db.Exec(tc.change).Error)
			time.Sleep(100 * time.Millisecond)

			assert.Equal(t, tc.after, usersCacheGet(cache, tc.userID))
		})
	}
}

func TestUsersCache_GetUser_DoesNotCacheAReadOverlappingAnInvalidation(t *testing.T) {
	db := setupTestDB(t)
	cache := NewUserCache(db)

	fired := false
	db.Callback().Query().After("gorm:query").Register("test:revoke_mid_read", func(scope *gorm.Scope) {
		if fired || scope.TableName() != "users" {
			return
		}
		fired = true
		_, err := RevokeSessions(db, cache, 1)
		require.NoError(t, err)
	})

	raced, err := cache.GetUser(1)
	require.NoError(t, err)
	require.True(t, fired, "the revocation has to land between the read and the store")

	after, err := cache.GetUser(1)
	require.NoError(t, err)

	assert.Equal(t, raced.Generation+1, after.Generation,
		"a generation read before a revocation must not be served for the cache's whole TTL")
}

func TestUsersCache_GetUser_ServesReadersDuringInvalidations(t *testing.T) {
	db := setupTestDB(t)
	cache := NewUserCache(db)

	want := map[uint64]usersCacheRead{
		1: {info: UserInfo{Hash: "testhash", Status: models.UserStatusActive, Generation: 1}},
		2: {info: UserInfo{Hash: "testhash2", Status: models.UserStatusActive, Generation: 1}},
	}

	type userRead struct {
		userID uint64
		read   usersCacheRead
	}
	reads := make(chan userRead, 100)
	var wg sync.WaitGroup
	for i := range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			userID := uint64(i%2 + 1)
			for range 10 {
				reads <- userRead{userID, usersCacheGet(cache, userID)}
			}
		}()
	}
	for i := range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 5 {
				cache.Invalidate(uint64(i%2 + 1))
				time.Sleep(10 * time.Millisecond)
			}
		}()
	}
	authWaitOrFail(t, &wg, "concurrent user reads")
	close(reads)

	for got := range reads {
		require.Equal(t, want[got.userID], got.read, "user %d", got.userID)
	}
}

func TestUsersCache_InvalidateAll_DropsEveryEntry(t *testing.T) {
	db := setupTestDB(t)
	cache := NewUserCache(db)

	for _, userID := range []uint64{1, 2} {
		_, err := cache.GetUser(userID)
		require.NoError(t, err)
	}
	require.NoError(t, db.Exec("UPDATE users SET status = 'blocked'").Error)

	cache.InvalidateAll()

	for _, userID := range []uint64{1, 2} {
		info, err := cache.GetUser(userID)
		require.NoError(t, err)
		assert.Equal(t, models.UserStatusBlocked, info.Status, "user %d", userID)
	}
}
