package auth_test

import (
	"pentagi/pkg/server/auth"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestPrivilegesRequired(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tokenCache := auth.NewTokenCache(db)
	userCache := auth.NewUserCache(db)
	authMiddleware := auth.NewAuthMiddleware("/base/url", "test", tokenCache, userCache)
	server := newTestServer(t, "/test", db, authMiddleware.AuthTokenRequired, auth.PrivilegesRequired("priv1", "priv2"))
	defer server.Close()

	server.SetSessionCheckFunc(func(t *testing.T, c *gin.Context) {
		t.Helper()
		assert.Equal(t, uint64(1), c.GetUint64("uid"))
	})

	assert.False(t, server.CallAndGetStatus(t))

	server.Authorize(t, []string{"some.permission"})
	assert.False(t, server.CallAndGetStatus(t))

	server.Authorize(t, []string{"priv1"})
	assert.False(t, server.CallAndGetStatus(t))

	server.Authorize(t, []string{"priv1", "priv2"})
	assert.True(t, server.CallAndGetStatus(t))
}
