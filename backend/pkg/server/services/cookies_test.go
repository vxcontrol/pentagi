package services

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"pentagi/pkg/server/oauth"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCookies_SessionOptions_MarksEverySessionCookieSecureBehindATLSTerminator(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	seedLocalUser(t, db, "known@corp.com")

	gin.SetMode(gin.TestMode)
	authSvc := NewAuthService(AuthServiceConfig{BaseURL: "/", SessionTimeout: 3600}, db, map[string]oauth.OAuthClient{}, nil)
	userSvc := NewUserService(db, nil, "/", 3600)

	engine := gin.New()
	require.NoError(t, engine.SetTrustedProxies(nil))
	engine.Use(sessions.Sessions("pentagi", cookie.NewStore([]byte("test-secret"))))
	engine.POST("/auth/login", authSvc.AuthLogin)
	engine.POST("/auth/logout", authSvc.AuthLogout)
	engine.PUT("/user/password", func(c *gin.Context) { c.Set("uid", sessions.Default(c).Get("uid")) },
		userSvc.ChangePasswordCurrentUser)
	engine.GET("/info", func(c *gin.Context) {
		session := sessions.Default(c)
		for _, key := range []string{"uid", "uhash", "tid", "exp", "prm", "sgn"} {
			c.Set(key, session.Get(key))
		}
		c.Set("gtm", time.Now().Add(-10*time.Minute).Unix())
	}, authSvc.Info)

	var current *http.Cookie
	sessionCookie := func(method, target, body string) *http.Cookie {
		t.Helper()

		req := httptest.NewRequest(method, target, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Forwarded-Proto", "https")
		req.RemoteAddr = "10.0.0.1:54321"
		if current != nil {
			req.AddCookie(current)
		}

		rec := httptest.NewRecorder()
		engine.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

		for _, c := range rec.Result().Cookies() {
			if c.Name == "pentagi" {
				current = c
				return c
			}
		}
		t.Fatalf("%s %s wrote no session cookie", method, target)

		return nil
	}

	login := sessionCookie(http.MethodPost, "/auth/login",
		`{"mail":"known@corp.com","password":"`+knownAccountPassword+`"}`)
	assert.True(t, login.Secure, "the login cookie")

	refreshed := sessionCookie(http.MethodGet, "/info", "")
	assert.True(t, refreshed.Secure, "the cookie /info re-mints once it is five minutes old")

	restamp := sessionCookie(http.MethodPut, "/user/password",
		`{"current_password":"`+knownAccountPassword+`","password":"NextAccountPass2!","confirm_password":"NextAccountPass2!"}`)
	assert.True(t, restamp.Secure, "the cookie re-stamped after a password change")

	logout := sessionCookie(http.MethodPost, "/auth/logout", "")
	assert.True(t, logout.Secure, "the cookie that ends the session")
}

func TestCookies_SessionOptions_MarksTheOAuthSessionCookieSecureBehindATLSTerminator(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	_, w := authOAuthCallback(t, newOAuthServiceVerified(db, "new@corp.com", true), http.Header{"X-Forwarded-Proto": {"https"}})
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	for _, cookie := range w.Result().Cookies() {
		if cookie.Name == "pentagi" {
			assert.True(t, cookie.Secure)
			return
		}
	}
	t.Fatal("the OAuth login wrote no session cookie")
}

func TestCookies_SessionOptions_ReadsTheSchemeFromEveryCommonProxyHeader(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	seedLocalUser(t, db, "known@corp.com")

	gin.SetMode(gin.TestMode)
	authSvc := NewAuthService(AuthServiceConfig{BaseURL: "/", SessionTimeout: 3600}, db, map[string]oauth.OAuthClient{}, nil)

	engine := gin.New()
	require.NoError(t, engine.SetTrustedProxies(nil))
	engine.Use(sessions.Sessions("pentagi", cookie.NewStore([]byte("test-secret"))))
	engine.POST("/auth/login", authSvc.AuthLogin)

	tests := []struct {
		name   string
		header http.Header
		secure bool
	}{
		{"plain http without proxy headers", http.Header{}, false},
		{"x-forwarded-proto https", http.Header{"X-Forwarded-Proto": {"https"}}, true},
		{"x-forwarded-proto upper case", http.Header{"X-Forwarded-Proto": {"HTTPS"}}, true},
		{"x-forwarded-proto chain keeps the client-facing hop", http.Header{"X-Forwarded-Proto": {"https, http"}}, true},
		{"x-forwarded-proto http", http.Header{"X-Forwarded-Proto": {"http"}}, false},
		{"x-forwarded-proto chain starting with http", http.Header{"X-Forwarded-Proto": {"http, https"}}, false},
		{"x-forwarded-ssl on", http.Header{"X-Forwarded-Ssl": {"on"}}, true},
		{"x-forwarded-ssl off", http.Header{"X-Forwarded-Ssl": {"off"}}, false},
		{"forwarded proto https", http.Header{"Forwarded": {"for=192.0.2.1;proto=https;host=example.com"}}, true},
		{"forwarded quoted proto", http.Header{"Forwarded": {`for="[2001:db8::1]";Proto="https"`}}, true},
		{"forwarded proto http", http.Header{"Forwarded": {"for=192.0.2.1;proto=http"}}, false},
		{"forwarded proto only on an inner hop", http.Header{"Forwarded": {"for=192.0.2.1;proto=http, for=10.0.0.1;proto=https"}}, false},
		{"forwarded without proto", http.Header{"Forwarded": {"for=192.0.2.1"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/auth/login",
				strings.NewReader(`{"mail":"known@corp.com","password":"`+knownAccountPassword+`"}`))
			for k, v := range tt.header {
				req.Header[k] = v
			}
			req.Header.Set("Content-Type", "application/json")
			req.RemoteAddr = "10.0.0.1:54321"

			rec := httptest.NewRecorder()
			engine.ServeHTTP(rec, req)
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

			for _, c := range rec.Result().Cookies() {
				if c.Name == "pentagi" {
					assert.Equal(t, tt.secure, c.Secure)
					return
				}
			}
			t.Fatal("the login wrote no session cookie")
		})
	}
}
