package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A cookie captured before the revocation is what an attacker holds.
func TestSessionGeneration_RevokeSessions_RefusesACookieMintedEarlier(t *testing.T) {
	db := setupTestDB(t)
	userCache := NewUserCache(db)
	server := newTestServer(t, NewAuthMiddleware("/base/url", "test", NewTokenCache(db), userCache).AuthUserRequired)

	server.Authorize(t, 5*time.Minute, "some.permission")
	require.Equal(t, http.StatusOK, server.Call(t, "").status, "the freshly issued cookie has to authenticate")

	generation, err := RevokeSessions(db, userCache, 1)
	require.NoError(t, err)
	assert.Equal(t, uint64(2), generation, "the caller re-stamps its own cookie with the generation it gets back")

	got := server.Call(t, "")
	assert.Equal(t, http.StatusForbidden, got.status)
	assert.Equal(t, "session revoked", got.reason, "the same cookie must stop authenticating once the generation moves")
}

func TestSessionGeneration_RevokeSessions_LeavesAPITokensAlone(t *testing.T) {
	db := setupTestDB(t)
	token := authIssueToken(t, db, "test", 1, "testhash")
	userCache := NewUserCache(db)
	server := newTestServer(t, NewAuthMiddleware("/base/url", "test", NewTokenCache(db), userCache).AuthTokenRequired)

	require.Equal(t, http.StatusOK, server.Call(t, "Bearer "+token).status)

	_, err := RevokeSessions(db, userCache, 1)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, server.Call(t, "Bearer "+token).status,
		"an API token is not a browser session and must outlive the revocation")
}
