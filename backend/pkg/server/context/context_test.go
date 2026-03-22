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
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(c *gin.Context)
		key     string
		wantVal int64
		wantOK  bool
	}{
		{
			name:    "found",
			setup:   func(c *gin.Context) { c.Set("id", int64(42)) },
			key:     "id",
			wantVal: 42,
			wantOK:  true,
		},
		{
			name:    "missing",
			setup:   func(c *gin.Context) {},
			key:     "id",
			wantVal: 0,
			wantOK:  false,
		},
		{
			name:    "wrong type string",
			setup:   func(c *gin.Context) { c.Set("id", "not-an-int") },
			key:     "id",
			wantVal: 0,
			wantOK:  false,
		},
		{
			name:    "wrong type uint64",
			setup:   func(c *gin.Context) { c.Set("id", uint64(99)) },
			key:     "id",
			wantVal: 0,
			wantOK:  false,
		},
		{
			name:    "zero value",
			setup:   func(c *gin.Context) { c.Set("id", int64(0)) },
			key:     "id",
			wantVal: 0,
			wantOK:  true,
		},
		{
			name:    "negative value",
			setup:   func(c *gin.Context) { c.Set("id", int64(-100)) },
			key:     "id",
			wantVal: -100,
			wantOK:  true,
		},
		{
			name:    "max int64",
			setup:   func(c *gin.Context) { c.Set("id", int64(9223372036854775807)) },
			key:     "id",
			wantVal: 9223372036854775807,
			wantOK:  true,
		},
		{
			name:    "different key",
			setup:   func(c *gin.Context) { c.Set("other", int64(123)) },
			key:     "id",
			wantVal: 0,
			wantOK:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			tt.setup(c)
			val, ok := GetInt64(c, tt.key)
			assert.Equal(t, tt.wantOK, ok)
			assert.Equal(t, tt.wantVal, val)
		})
	}
}

func TestGetUint64(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(c *gin.Context)
		key     string
		wantVal uint64
		wantOK  bool
	}{
		{
			name:    "found",
			setup:   func(c *gin.Context) { c.Set("uid", uint64(99)) },
			key:     "uid",
			wantVal: 99,
			wantOK:  true,
		},
		{
			name:    "missing",
			setup:   func(c *gin.Context) {},
			key:     "uid",
			wantVal: 0,
			wantOK:  false,
		},
		{
			name:    "wrong type int64",
			setup:   func(c *gin.Context) { c.Set("uid", int64(99)) },
			key:     "uid",
			wantVal: 0,
			wantOK:  false,
		},
		{
			name:    "wrong type string",
			setup:   func(c *gin.Context) { c.Set("uid", "99") },
			key:     "uid",
			wantVal: 0,
			wantOK:  false,
		},
		{
			name:    "zero value",
			setup:   func(c *gin.Context) { c.Set("uid", uint64(0)) },
			key:     "uid",
			wantVal: 0,
			wantOK:  true,
		},
		{
			name:    "large value",
			setup:   func(c *gin.Context) { c.Set("uid", uint64(18446744073709551615)) },
			key:     "uid",
			wantVal: 18446744073709551615,
			wantOK:  true,
		},
		{
			name:    "different key",
			setup:   func(c *gin.Context) { c.Set("other", uint64(456)) },
			key:     "uid",
			wantVal: 0,
			wantOK:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			tt.setup(c)
			val, ok := GetUint64(c, tt.key)
			assert.Equal(t, tt.wantOK, ok)
			assert.Equal(t, tt.wantVal, val)
		})
	}
}

func TestGetString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(c *gin.Context)
		key     string
		wantVal string
		wantOK  bool
	}{
		{
			name:    "found",
			setup:   func(c *gin.Context) { c.Set("name", "alice") },
			key:     "name",
			wantVal: "alice",
			wantOK:  true,
		},
		{
			name:    "missing",
			setup:   func(c *gin.Context) {},
			key:     "name",
			wantVal: "",
			wantOK:  false,
		},
		{
			name:    "wrong type int",
			setup:   func(c *gin.Context) { c.Set("name", 123) },
			key:     "name",
			wantVal: "",
			wantOK:  false,
		},
		{
			name:    "wrong type bool",
			setup:   func(c *gin.Context) { c.Set("name", true) },
			key:     "name",
			wantVal: "",
			wantOK:  false,
		},
		{
			name:    "empty string",
			setup:   func(c *gin.Context) { c.Set("name", "") },
			key:     "name",
			wantVal: "",
			wantOK:  true,
		},
		{
			name:    "long string",
			setup:   func(c *gin.Context) { c.Set("name", "very-long-string-with-special-chars-@#$%") },
			key:     "name",
			wantVal: "very-long-string-with-special-chars-@#$%",
			wantOK:  true,
		},
		{
			name:    "different key",
			setup:   func(c *gin.Context) { c.Set("other", "value") },
			key:     "name",
			wantVal: "",
			wantOK:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			tt.setup(c)
			val, ok := GetString(c, tt.key)
			assert.Equal(t, tt.wantOK, ok)
			assert.Equal(t, tt.wantVal, val)
		})
	}
}

func TestGetStringArray(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(c *gin.Context)
		key     string
		wantVal []string
		wantOK  bool
	}{
		{
			name:    "found",
			setup:   func(c *gin.Context) { c.Set("perms", []string{"read", "write"}) },
			key:     "perms",
			wantVal: []string{"read", "write"},
			wantOK:  true,
		},
		{
			name:    "missing",
			setup:   func(c *gin.Context) {},
			key:     "perms",
			wantVal: []string{},
			wantOK:  false,
		},
		{
			name:    "wrong type string",
			setup:   func(c *gin.Context) { c.Set("perms", "not-a-slice") },
			key:     "perms",
			wantVal: []string{},
			wantOK:  false,
		},
		{
			name:    "wrong type int slice",
			setup:   func(c *gin.Context) { c.Set("perms", []int{1, 2, 3}) },
			key:     "perms",
			wantVal: []string{},
			wantOK:  false,
		},
		{
			name:    "empty array",
			setup:   func(c *gin.Context) { c.Set("perms", []string{}) },
			key:     "perms",
			wantVal: []string{},
			wantOK:  true,
		},
		{
			name:    "nil array",
			setup:   func(c *gin.Context) { c.Set("perms", []string(nil)) },
			key:     "perms",
			wantVal: nil,
			wantOK:  true,
		},
		{
			name:    "single element",
			setup:   func(c *gin.Context) { c.Set("perms", []string{"admin"}) },
			key:     "perms",
			wantVal: []string{"admin"},
			wantOK:  true,
		},
		{
			name:    "many elements",
			setup:   func(c *gin.Context) { c.Set("perms", []string{"a", "b", "c", "d", "e"}) },
			key:     "perms",
			wantVal: []string{"a", "b", "c", "d", "e"},
			wantOK:  true,
		},
		{
			name:    "different key",
			setup:   func(c *gin.Context) { c.Set("other", []string{"val"}) },
			key:     "perms",
			wantVal: []string{},
			wantOK:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			tt.setup(c)
			val, ok := GetStringArray(c, tt.key)
			assert.Equal(t, tt.wantOK, ok)
			assert.Equal(t, tt.wantVal, val)
		})
	}
}

func TestGetStringFromSession(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(session sessions.Session)
		key     string
		wantVal string
		wantOK  bool
	}{
		{
			name: "found",
			setup: func(s sessions.Session) {
				s.Set("token", "abc123")
				_ = s.Save()
			},
			key:     "token",
			wantVal: "abc123",
			wantOK:  true,
		},
		{
			name:    "missing",
			setup:   func(s sessions.Session) {},
			key:     "token",
			wantVal: "",
			wantOK:  false,
		},
		{
			name: "wrong type int",
			setup: func(s sessions.Session) {
				s.Set("token", 999)
				_ = s.Save()
			},
			key:     "token",
			wantVal: "",
			wantOK:  false,
		},
		{
			name: "wrong type bool",
			setup: func(s sessions.Session) {
				s.Set("token", true)
				_ = s.Save()
			},
			key:     "token",
			wantVal: "",
			wantOK:  false,
		},
		{
			name: "empty string",
			setup: func(s sessions.Session) {
				s.Set("token", "")
				_ = s.Save()
			},
			key:     "token",
			wantVal: "",
			wantOK:  true,
		},
		{
			name: "different key",
			setup: func(s sessions.Session) {
				s.Set("other", "value")
				_ = s.Save()
			},
			key:     "token",
			wantVal: "",
			wantOK:  false,
		},
		{
			name: "multiple values in session",
			setup: func(s sessions.Session) {
				s.Set("token", "abc123")
				s.Set("user", "alice")
				s.Set("role", "admin")
				_ = s.Save()
			},
			key:     "token",
			wantVal: "abc123",
			wantOK:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			store := cookie.NewStore([]byte("test-secret"))
			router := gin.New()
			router.Use(sessions.Sessions("test", store))

			var val string
			var ok bool

			router.GET("/test", func(c *gin.Context) {
				session := sessions.Default(c)
				tt.setup(session)
				val, ok = GetStringFromSession(c, tt.key)
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantOK, ok)
			assert.Equal(t, tt.wantVal, val)
		})
	}
}

func TestMultipleValuesInContext(t *testing.T) {
	t.Parallel()

	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	c.Set("int64_val", int64(123))
	c.Set("uint64_val", uint64(456))
	c.Set("string_val", "test")
	c.Set("array_val", []string{"a", "b"})

	// Verify all values are independently accessible
	intVal, ok := GetInt64(c, "int64_val")
	assert.True(t, ok)
	assert.Equal(t, int64(123), intVal)

	uintVal, ok := GetUint64(c, "uint64_val")
	assert.True(t, ok)
	assert.Equal(t, uint64(456), uintVal)

	strVal, ok := GetString(c, "string_val")
	assert.True(t, ok)
	assert.Equal(t, "test", strVal)

	arrVal, ok := GetStringArray(c, "array_val")
	assert.True(t, ok)
	assert.Equal(t, []string{"a", "b"}, arrVal)
}

func TestContextOverwrite(t *testing.T) {
	t.Parallel()

	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	// Set initial value
	c.Set("key", "original")
	val, ok := GetString(c, "key")
	assert.True(t, ok)
	assert.Equal(t, "original", val)

	// Overwrite with new value
	c.Set("key", "updated")
	val, ok = GetString(c, "key")
	assert.True(t, ok)
	assert.Equal(t, "updated", val)
}

func TestContextTypeChange(t *testing.T) {
	t.Parallel()

	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	// Set as string
	c.Set("value", "123")
	strVal, ok := GetString(c, "value")
	assert.True(t, ok)
	assert.Equal(t, "123", strVal)

	// Try to get as int64 - should fail
	intVal, ok := GetInt64(c, "value")
	assert.False(t, ok)
	assert.Equal(t, int64(0), intVal)

	// Overwrite with int64
	c.Set("value", int64(123))
	intVal, ok = GetInt64(c, "value")
	assert.True(t, ok)
	assert.Equal(t, int64(123), intVal)
}
