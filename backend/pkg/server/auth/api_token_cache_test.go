package auth

import (
	"slices"
	"sync"
	"testing"
	"time"

	"pentagi/pkg/server/models"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPITokenCache_GetStatus_GrantsTheTokenRolePlusAutomation(t *testing.T) {
	cases := []struct {
		name string
		role uint64
		want []string
	}{
		{"an admin token", 1, apiTokenCacheAdminRole},
		{"a user token", 2, apiTokenCacheUserRole},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := setupTestDB(t)
			authStoreToken(t, db, "roletoken1", 1, tc.role)

			status, privileges, err := NewTokenCache(db).GetStatus("roletoken1")

			require.NoError(t, err)
			assert.Equal(t, models.TokenStatusActive, status)
			assert.ElementsMatch(t, tc.want, privileges)
		})
	}
}

// What a token of role 1 and of role 2 is granted, sorted.
var (
	apiTokenCacheAdminRole = []string{
		"flows.admin", "flows.create", "flows.delete", "flows.edit", "flows.view", "pentagi.automation",
		"roles.view", "settings.tokens.admin", "settings.tokens.create", "settings.tokens.delete",
		"settings.tokens.edit", "settings.tokens.view", "users.create", "users.delete", "users.edit", "users.view",
	}
	apiTokenCacheUserRole = []string{
		"flows.create", "flows.delete", "flows.edit", "flows.view", "pentagi.automation", "roles.view",
		"settings.tokens.create", "settings.tokens.delete", "settings.tokens.edit", "settings.tokens.view",
	}
)

type apiTokenCacheRead struct {
	status     models.TokenStatus
	privileges []string
	err        error
}

func apiTokenCacheGet(cache *TokenCache, tokenID string) apiTokenCacheRead {
	status, privileges, err := cache.GetStatus(tokenID)
	return apiTokenCacheRead{status: status, privileges: slices.Sorted(slices.Values(privileges)), err: err}
}

var (
	apiTokenCacheActive   = apiTokenCacheRead{status: models.TokenStatusActive, privileges: apiTokenCacheUserRole}
	apiTokenCacheAdmin    = apiTokenCacheRead{status: models.TokenStatusActive, privileges: apiTokenCacheAdminRole}
	apiTokenCacheRevoked  = apiTokenCacheRead{status: models.TokenStatusRevoked, privileges: apiTokenCacheUserRole}
	apiTokenCacheNotFound = apiTokenCacheRead{err: gorm.ErrRecordNotFound}
)

// A stored token revoked, or moved to another role, after it was cached, and a missing token stored after its absence was.
var apiTokenCacheEntries = []struct {
	name          string
	stored        bool
	change        string
	before, after apiTokenCacheRead
}{
	{"a stored token", true, "UPDATE api_tokens SET status = 'revoked'", apiTokenCacheActive, apiTokenCacheRevoked},
	{"a token moved to another role", true, "UPDATE api_tokens SET role_id = 1", apiTokenCacheActive, apiTokenCacheAdmin},
	{"a missing token", false, "INSERT INTO api_tokens (token_id, user_id, role_id, ttl) VALUES ('cachetoken', 1, 2, 3600)",
		apiTokenCacheNotFound, apiTokenCacheActive},
}

func TestAPITokenCache_GetStatus_ServesTheCachedEntryUntilInvalidated(t *testing.T) {
	for _, tc := range apiTokenCacheEntries {
		t.Run(tc.name, func(t *testing.T) {
			db := setupTestDB(t)
			if tc.stored {
				authStoreToken(t, db, "cachetoken", 1, 2)
			}
			cache := NewTokenCache(db)

			require.Equal(t, tc.before, apiTokenCacheGet(cache, "cachetoken"))
			require.NoError(t, db.Exec(tc.change).Error)
			assert.Equal(t, tc.before, apiTokenCacheGet(cache, "cachetoken"), "the store changed under a cached entry")

			cache.Invalidate("cachetoken")
			assert.Equal(t, tc.after, apiTokenCacheGet(cache, "cachetoken"))
		})
	}
}

func TestAPITokenCache_GetStatus_RereadsAnEntryPastItsTTL(t *testing.T) {
	for _, tc := range apiTokenCacheEntries {
		t.Run(tc.name, func(t *testing.T) {
			db := setupTestDB(t)
			if tc.stored {
				authStoreToken(t, db, "cachetoken", 1, 2)
			}
			cache := NewTokenCache(db)
			cache.SetTTL(50 * time.Millisecond)

			require.Equal(t, tc.before, apiTokenCacheGet(cache, "cachetoken"))
			require.NoError(t, db.Exec(tc.change).Error)
			time.Sleep(100 * time.Millisecond)

			assert.Equal(t, tc.after, apiTokenCacheGet(cache, "cachetoken"))
		})
	}
}

func TestAPITokenCache_GetStatus_ServesConcurrentReaders(t *testing.T) {
	db := setupTestDB(t)
	cache := NewTokenCache(db)

	tokenIDs := make([]string, 10)
	for i := range tokenIDs {
		tokenID, err := GenerateTokenID()
		require.NoError(t, err)
		authStoreToken(t, db, tokenID, 1, 2)
		tokenIDs[i] = tokenID
		require.Equal(t, apiTokenCacheActive, apiTokenCacheGet(cache, tokenID))
	}

	reads := make(chan apiTokenCacheRead, len(tokenIDs)*100)
	var wg sync.WaitGroup
	for _, tokenID := range tokenIDs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 100 {
				reads <- apiTokenCacheGet(cache, tokenID)
			}
		}()
	}
	authWaitOrFail(t, &wg, "concurrent token reads")
	close(reads)

	for read := range reads {
		require.Equal(t, apiTokenCacheActive, read)
	}
}

func TestAPITokenCache_InvalidateUser_DropsEveryTokenOfTheUser(t *testing.T) {
	db := setupTestDB(t)
	cache := NewTokenCache(db)

	for _, tokenID := range []string{"usertoken1", "usertoken2"} {
		authStoreToken(t, db, tokenID, 1, 2)
		require.Equal(t, apiTokenCacheActive, apiTokenCacheGet(cache, tokenID))
	}
	require.NoError(t, db.Exec("UPDATE api_tokens SET status = 'revoked' WHERE user_id = 1").Error)

	cache.InvalidateUser(1)

	for _, tokenID := range []string{"usertoken1", "usertoken2"} {
		assert.Equal(t, apiTokenCacheRevoked, apiTokenCacheGet(cache, tokenID), tokenID)
	}
}
