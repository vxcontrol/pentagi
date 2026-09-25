package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPermissions_PrivilegesRequired_DemandsEveryListedPrivilege(t *testing.T) {
	db := setupTestDB(t)
	middleware := NewAuthMiddleware("/base/url", "test", NewTokenCache(db), NewUserCache(db))
	server := newTestServer(t, middleware.AuthTokenRequired, PrivilegesRequired("priv1", "priv2"))

	cases := []struct {
		name       string
		privileges []string
		wantReason string
	}{
		{"none of the listed privileges", []string{"some.permission"}, "'priv1' is not set"},
		{"one of the two", []string{"priv1"}, "'priv2' is not set"},
		{"both of them", []string{"priv1", "priv2"}, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server.Authorize(t, 5*time.Minute, tc.privileges...)

			got := server.Call(t, "")

			if tc.wantReason != "" {
				assert.Equal(t, http.StatusForbidden, got.status)
				assert.Equal(t, tc.wantReason, got.reason)
				assert.Nil(t, got.seen)
				return
			}
			assert.Equal(t, http.StatusOK, got.status)
			require.NotNil(t, got.seen)
			assert.Equal(t, uint64(1), got.seen.UID)
		})
	}
}
