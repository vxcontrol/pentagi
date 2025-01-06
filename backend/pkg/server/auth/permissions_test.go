package auth

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestPrivilegesRequired(t *testing.T) {
	authMiddleware := NewAuthMiddleware("/base/url", "test")
	server := newTestServer(t, "/test", authMiddleware.AuthRequired, PrivilegesRequired("priv1", "priv2"))
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
