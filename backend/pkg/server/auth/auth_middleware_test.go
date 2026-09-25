package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthMiddleware_TryUserCookieAuthentication_AdmitsOnlyALiveSessionOfAnActiveUser(t *testing.T) {
	cases := []struct {
		name       string
		expiresIn  time.Duration // zero sends no session cookie
		privileges []string
		change     string // applied to the store after the login
		wantReason string // empty when the session is admitted
		wantCPT    string
	}{
		{name: "a session without the automation privilege", expiresIn: 5 * time.Minute,
			privileges: []string{"some.permission"}},
		{name: "a session holding the automation privilege among others", expiresIn: 5 * time.Minute,
			privileges: []string{"some.permission", "pentagi.automation"}, wantCPT: "automation"},
		{name: "a session granting no privileges", expiresIn: 5 * time.Minute},
		{name: "no session cookie", wantReason: "cookie claim invalid"},
		{name: "a session past its expiry", expiresIn: -time.Second, privileges: []string{"some.permission"},
			wantReason: "session expired"},
		{name: "a hash changed since the login", expiresIn: 5 * time.Minute, privileges: []string{"some.permission"},
			change:     "UPDATE users SET hash = 'modified_hash' WHERE id = 1",
			wantReason: "user hash mismatch - session invalid for this installation"},
		{name: "a blocked user", expiresIn: 5 * time.Minute, privileges: []string{"some.permission"},
			change: "UPDATE users SET status = 'blocked' WHERE id = 1", wantReason: "user has been blocked"},
		{name: "a user not activated yet", expiresIn: 5 * time.Minute, privileges: []string{"some.permission"},
			change: "UPDATE users SET status = 'created' WHERE id = 1", wantReason: "user is not ready"},
		{name: "a deleted user", expiresIn: 5 * time.Minute, privileges: []string{"some.permission"},
			change: "DELETE FROM users WHERE id = 1", wantReason: "user has been deleted"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := setupTestDB(t)
			server := newTestServer(t, NewAuthMiddleware("/base/url", "test", NewTokenCache(db), NewUserCache(db)).AuthUserRequired)
			if tc.expiresIn != 0 {
				server.Authorize(t, tc.expiresIn, tc.privileges...)
			}
			if tc.change != "" {
				require.NoError(t, db.Exec(tc.change).Error)
			}

			got := server.Call(t, "")

			if tc.wantReason != "" {
				assert.Equal(t, http.StatusForbidden, got.status)
				assert.Equal(t, tc.wantReason, got.reason)
				assert.Nil(t, got.seen)
				return
			}
			assert.Equal(t, http.StatusOK, got.status)
			require.NotNil(t, got.seen)
			assert.Equal(t, authIdentity{
				UID: 1, RID: 2, SGN: 1, UHash: "testhash", TID: "local", CPT: tc.wantCPT, UName: "User 1",
				Prm: tc.privileges,
			}, got.seen.withoutClock())
			assert.Equal(t, int64(300), got.seen.EXP-got.seen.GTM)
		})
	}
}

func TestAuthMiddleware_TryUserCookieAuthentication_RereadsAGenerationNewerThanTheCache(t *testing.T) {
	db := setupTestDB(t)
	userCache := NewUserCache(db)
	server := newTestServer(t, NewAuthMiddleware("/base/url", "test", NewTokenCache(db), userCache).AuthUserRequired)

	require.NoError(t, db.Exec("UPDATE users SET session_generation = 0 WHERE id = 1").Error)
	_, err := userCache.GetUser(1)
	require.NoError(t, err)
	require.NoError(t, db.Exec("UPDATE users SET session_generation = 1 WHERE id = 1").Error)

	server.Authorize(t, 5*time.Minute, "some.permission")

	assert.Equal(t, http.StatusOK, server.Call(t, "").status,
		"a cookie minted from the row after an installer reset must not wait out the cached generation")
}

// A cookie is always sent: an uncheckable header falls through to it, a refused token does not.
func TestAuthMiddleware_TryProtoTokenAuthentication_AdmitsOnlyALiveTokenOfAnUnblockedUser(t *testing.T) {
	admitted := map[string]struct {
		identity authIdentity
		lifetime int64
	}{
		"token": {authIdentity{
			UID: 1, RID: 2, UHash: "testhash", TID: "api", CPT: "automation", Prm: []string{"pentagi.automation"},
		}, 3600},
		"cookie": {authIdentity{
			UID: 1, RID: 2, SGN: 1, UHash: "testhash", TID: "local", UName: "User 1", Prm: []string{"some.permission"},
		}, 300},
	}

	cases := []struct {
		name       string
		salt       string // the middleware's, and the one the token is signed under
		header     string // {token} stands for the issued token
		change     string
		admittedAs string
		wantReason string
	}{
		{name: "a live token", salt: "test", header: "Bearer {token}", admittedAs: "token"},
		{name: "no authorization header", salt: "test", header: "", admittedAs: "cookie"},
		{name: "a token without the bearer scheme", salt: "test", header: "{token}", admittedAs: "cookie"},
		{name: "the bearer scheme without its space", salt: "test", header: "Bearer{token}", admittedAs: "cookie"},
		{name: "a token under the default salt", salt: "salt", header: "Bearer {token}", admittedAs: "cookie"},
		{name: "a token under an empty salt", salt: "", header: "Bearer {token}", admittedAs: "cookie"},
		{name: "a token that fails validation", salt: "test", header: "Bearer not_a_token",
			wantReason: "token is invalid"},
		{name: "a revoked token", salt: "test", header: "Bearer {token}",
			change: "UPDATE api_tokens SET status = 'revoked'", wantReason: "token has been revoked"},
		{name: "a deleted token", salt: "test", header: "Bearer {token}",
			change: "UPDATE api_tokens SET deleted_at = CURRENT_TIMESTAMP", wantReason: "token not found in database"},
		{name: "a hash changed since the token was issued", salt: "test", header: "Bearer {token}",
			change:     "UPDATE users SET hash = 'different_hash' WHERE id = 1",
			wantReason: "user hash mismatch - token invalid for this installation"},
		{name: "a blocked user", salt: "test", header: "Bearer {token}",
			change: "UPDATE users SET status = 'blocked' WHERE id = 1", wantReason: "user has been blocked"},
		{name: "a deleted user", salt: "test", header: "Bearer {token}",
			change: "DELETE FROM users WHERE id = 1", wantReason: "user has been deleted"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := setupTestDB(t)
			token := authIssueToken(t, db, tc.salt, 1, "testhash")
			server := newTestServer(t, NewAuthMiddleware("/base/url", tc.salt, NewTokenCache(db), NewUserCache(db)).AuthTokenRequired)
			server.Authorize(t, 5*time.Minute, "some.permission")
			if tc.change != "" {
				require.NoError(t, db.Exec(tc.change).Error)
			}

			got := server.Call(t, strings.ReplaceAll(tc.header, "{token}", token))

			if tc.wantReason != "" {
				assert.Equal(t, http.StatusForbidden, got.status)
				assert.Equal(t, tc.wantReason, got.reason)
				assert.Nil(t, got.seen)
				return
			}
			want := admitted[tc.admittedAs]
			assert.Equal(t, http.StatusOK, got.status)
			require.NotNil(t, got.seen)
			assert.Equal(t, want.identity, got.seen.withoutClock())
			assert.InDelta(t, want.lifetime, got.seen.EXP-got.seen.GTM, 60)
		})
	}
}

func TestAuthMiddleware_AuthUserRequired_RefusesAnAPIToken(t *testing.T) {
	db := setupTestDB(t)
	token := authIssueToken(t, db, "test", 1, "testhash")
	server := newTestServer(t, NewAuthMiddleware("/base/url", "test", NewTokenCache(db), NewUserCache(db)).AuthUserRequired)

	got := server.Call(t, "Bearer "+token)

	assert.Equal(t, http.StatusForbidden, got.status, "a leaked API token must not reach the password and e-mail change")
	assert.Equal(t, "cookie claim invalid", got.reason)
}

func TestAuthMiddleware_LevelForAuthFailure_BlamesTheServerOnlyForABackendFailure(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want logrus.Level
	}{
		{"a refused credential", errors.New("user has been blocked"), logrus.WarnLevel},
		{"a session cookie missing claims", errCookieClaimInvalid, logrus.WarnLevel},
		{"a session past its expiry", errSessionExpired, logrus.WarnLevel},
		{"a hash left over from a password change", fmt.Errorf("%w - session invalid", errUserHashMismatch), logrus.WarnLevel},
		{"a store that cannot answer", fmt.Errorf("%w: checking user status: %w", errAuthBackend, errors.New("dial tcp: refused")), logrus.ErrorLevel},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, levelForAuthFailure(tc.err))
		})
	}
}
