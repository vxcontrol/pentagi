package context

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestGetInt64(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(c *gin.Context)
		wantVal int64
		wantOK  bool
	}{
		{
			name:    "found",
			setup:   func(c *gin.Context) { c.Set("id", int64(42)) },
			wantVal: 42,
			wantOK:  true,
		},
		{
			name:   "missing",
			setup:  func(c *gin.Context) {},
			wantOK: false,
		},
		{
			name:   "wrong type",
			setup:  func(c *gin.Context) { c.Set("id", "not-an-int") },
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			tt.setup(c)
			val, ok := GetInt64(c, "id")
			assert.Equal(t, tt.wantOK, ok)
			assert.Equal(t, tt.wantVal, val)
		})
	}
}

func TestGetUint64(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(c *gin.Context)
		wantVal uint64
		wantOK  bool
	}{
		{
			name:    "found",
			setup:   func(c *gin.Context) { c.Set("uid", uint64(99)) },
			wantVal: 99,
			wantOK:  true,
		},
		{
			name:   "missing",
			setup:  func(c *gin.Context) {},
			wantOK: false,
		},
		{
			name:   "wrong type",
			setup:  func(c *gin.Context) { c.Set("uid", int64(99)) },
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			tt.setup(c)
			val, ok := GetUint64(c, "uid")
			assert.Equal(t, tt.wantOK, ok)
			assert.Equal(t, tt.wantVal, val)
		})
	}
}

func TestGetString(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(c *gin.Context)
		wantVal string
		wantOK  bool
	}{
		{
			name:    "found",
			setup:   func(c *gin.Context) { c.Set("name", "alice") },
			wantVal: "alice",
			wantOK:  true,
		},
		{
			name:   "missing",
			setup:  func(c *gin.Context) {},
			wantOK: false,
		},
		{
			name:   "wrong type",
			setup:  func(c *gin.Context) { c.Set("name", 123) },
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			tt.setup(c)
			val, ok := GetString(c, "name")
			assert.Equal(t, tt.wantOK, ok)
			assert.Equal(t, tt.wantVal, val)
		})
	}
}

func TestGetStringArray(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(c *gin.Context)
		wantVal []string
		wantOK  bool
	}{
		{
			name:    "found",
			setup:   func(c *gin.Context) { c.Set("perms", []string{"read", "write"}) },
			wantVal: []string{"read", "write"},
			wantOK:  true,
		},
		{
			name:    "missing",
			setup:   func(c *gin.Context) {},
			wantVal: []string{},
			wantOK:  false,
		},
		{
			name:    "wrong type",
			setup:   func(c *gin.Context) { c.Set("perms", "not-a-slice") },
			wantVal: []string{},
			wantOK:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			tt.setup(c)
			val, ok := GetStringArray(c, "perms")
			assert.Equal(t, tt.wantOK, ok)
			assert.Equal(t, tt.wantVal, val)
		})
	}
}

func TestGetStringFromSession(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(session sessions.Session)
		wantVal string
		wantOK  bool
	}{
		{
			name: "found",
			setup: func(s sessions.Session) {
				s.Set("token", "abc123")
				_ = s.Save()
			},
			wantVal: "abc123",
			wantOK:  true,
		},
		{
			name:   "missing",
			setup:  func(s sessions.Session) {},
			wantOK: false,
		},
		{
			name: "wrong type",
			setup: func(s sessions.Session) {
				s.Set("token", 999)
				_ = s.Save()
			},
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := cookie.NewStore([]byte("test-secret"))
			router := gin.New()
			router.Use(sessions.Sessions("test", store))

			var val string
			var ok bool

			router.GET("/test", func(c *gin.Context) {
				session := sessions.Default(c)
				tt.setup(session)
				val, ok = GetStringFromSession(c, "token")
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantOK, ok)
			assert.Equal(t, tt.wantVal, val)
		})
	}
}
