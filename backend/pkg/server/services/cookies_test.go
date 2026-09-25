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
