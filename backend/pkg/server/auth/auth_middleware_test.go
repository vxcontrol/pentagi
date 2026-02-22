package auth_test

import (
	"bytes"
	"io"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"pentagi/pkg/server/auth"
	"pentagi/pkg/server/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthTokenProtoRequiredAuthWithCookie(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tokenCache := auth.NewTokenCache(db)
	userCache := auth.NewUserCache(db)
	authMiddleware := auth.NewAuthMiddleware("/base/url", "test", tokenCache, userCache)

	t.Run("test URL", func(t *testing.T) {
		server := newTestServer(t, "/test", db, authMiddleware.AuthTokenRequired)
		defer server.Close()

		assert.False(t, server.CallAndGetStatus(t))

		server.SetSessionCheckFunc(func(t *testing.T, c *gin.Context) {
			t.Helper()
			assert.Equal(t, "", c.GetString("cpt"))
		})

		server.Authorize(t, []string{})
		assert.True(t, server.CallAndGetStatus(t))

		server.Authorize(t, []string{"wrong.permission"})
		assert.True(t, server.CallAndGetStatus(t))

		server.SetSessionCheckFunc(func(t *testing.T, c *gin.Context) {
			t.Helper()
			assert.Equal(t, "automation", c.GetString("cpt"))
		})

		server.Authorize(t, []string{auth.PrivilegeAutomation})
		assert.True(t, server.CallAndGetStatus(t))

		server.Authorize(t, []string{"wrong.permission", auth.PrivilegeAutomation})
		assert.True(t, server.CallAndGetStatus(t))
	})
}

func TestAuthTokenProtoRequiredAuthWithToken(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tokenCache := auth.NewTokenCache(db)
	userCache := auth.NewUserCache(db)
	authMiddleware := auth.NewAuthMiddleware("/base/url", "test", tokenCache, userCache)

	server := newTestServer(t, "/test", db, authMiddleware.AuthTokenRequired)
	defer server.Close()

	server.Authorize(t, []string{auth.PrivilegeAutomation})
	token := server.GetToken(t)
	require.NotEmpty(t, token)

	server.Unauthorize(t)
	assert.False(t, server.CallAndGetStatus(t))

	assert.False(t, server.CallAndGetStatus(t, token))
	assert.False(t, server.CallAndGetStatus(t, "not a bearer "+token))
	assert.False(t, server.CallAndGetStatus(t, "Bearer"+token))
	assert.False(t, server.CallAndGetStatus(t, "Bearer not_a_token"))

	server.SetSessionCheckFunc(func(t *testing.T, c *gin.Context) {
		t.Helper()
		assert.Equal(t, uint64(1), c.GetUint64("uid"))
		assert.Equal(t, uint64(2), c.GetUint64("rid"))
		assert.NotNil(t, c.GetStringSlice("prm"))

		// gtm and exp should now be set for API tokens
		gtm := c.GetInt64("gtm")
		assert.Greater(t, gtm, int64(0), "GTM should be set")

		exp := c.GetInt64("exp")
		assert.Greater(t, exp, gtm, "EXP should be greater than GTM")

		// uuid will be empty for invalid hash (test uses "123" which is not valid MD5)
		assert.NotNil(t, c.GetString("uuid"))

		assert.Equal(t, "automation", c.GetString("cpt"))
		assert.Empty(t, c.GetString("uname"))
	})

	assert.True(t, server.CallAndGetStatus(t, "Bearer "+token))
}

func TestAuthRequiredAuthWithCookie(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tokenCache := auth.NewTokenCache(db)
	userCache := auth.NewUserCache(db)
	authMiddleware := auth.NewAuthMiddleware("/base/url", "test", tokenCache, userCache)

	server := newTestServer(t, "/test", db, authMiddleware.AuthUserRequired)
	defer server.Close()

	server.SetSessionCheckFunc(func(t *testing.T, c *gin.Context) {
		t.Helper()
		assert.Equal(t, uint64(1), c.GetUint64("uid"))
		assert.Equal(t, uint64(2), c.GetUint64("rid"))
		assert.NotNil(t, c.GetStringSlice("prm"))
		assert.NotNil(t, c.GetInt64("gtm"))
		assert.NotNil(t, c.GetInt64("exp"))
		assert.Empty(t, c.GetString("uuid"))
		assert.Equal(t, "User 1", c.GetString("uname"))
	})

	assert.False(t, server.CallAndGetStatus(t))

	server.Authorize(t, []string{"some.permission"})
	assert.True(t, server.CallAndGetStatus(t))
}

type testServer struct {
	testEndpoint     string
	client           *http.Client
	calls            map[string]struct{}
	sessionCheckFunc func(t *testing.T, c *gin.Context)
	db               *gorm.DB
	*httptest.Server
}

func newTestServer(t *testing.T, testEndpoint string, db *gorm.DB, middlewares ...gin.HandlerFunc) *testServer {
	t.Helper()

	server := &testServer{
		db: db,
	}

	router := gin.New()
	globalSalt := "test"
	cookieStore := cookie.NewStore(auth.MakeCookieStoreKey(globalSalt)...)
	router.Use(sessions.Sessions("auth", cookieStore))

	server.calls = map[string]struct{}{}

	if testEndpoint == "" {
		testEndpoint = "/test"
	}
	server.testEndpoint = testEndpoint

	router.GET("/auth", func(c *gin.Context) {
		t.Helper()
		privs, _ := c.GetQueryArray("privileges")
		expString, ok := c.GetQuery("expiration")
		assert.True(t, ok)
		exp, err := strconv.Atoi(expString)
		assert.NoError(t, err)
		setTestSession(t, c, privs, exp)
	})

	authRoutes := router.Group("")
	for _, middleware := range middlewares {
		authRoutes.Use(middleware)
	}

	authRoutes.GET(server.testEndpoint, func(c *gin.Context) {
		t.Helper()

		id, _ := c.GetQuery("id")
		require.NotEmpty(t, id)

		if server.sessionCheckFunc != nil {
			server.sessionCheckFunc(t, c)
		}
		server.calls[id] = struct{}{}
	})

	authRoutes.GET("/auth_token", func(c *gin.Context) {
		t.Helper()

		tokenID, err := auth.GenerateTokenID()
		require.NoError(t, err)
		uhash := "testhash"
		uid := uint64(1)
		rid := uint64(2)
		ttl := uint64(3600)
		claims := auth.MakeAPITokenClaims(tokenID, uhash, uid, rid, ttl)
		token, err := auth.MakeAPIToken(globalSalt, claims)
		require.NoError(t, err)

		db.Create(&models.APIToken{
			TokenID: tokenID,
			UserID:  uid,
			RoleID:  rid,
			TTL:     ttl,
			Status:  models.TokenStatusActive,
		})

		c.Writer.Write([]byte(token))
	})

	server.Server = httptest.NewServer(router)
	server.client = server.Client()
	jar, err := cookiejar.New(nil)
	require.NoError(t, err)
	server.client.Jar = jar

	return server
}

func (s *testServer) Authorize(t *testing.T, privileges []string) {
	t.Helper()
	request, err := http.NewRequest(http.MethodGet, s.URL+"/auth", nil)
	require.NoError(t, err)
	query := url.Values{}
	for _, p := range privileges {
		query.Add("privileges", p)
	}
	query.Add("expiration", strconv.Itoa(5*60))
	request.URL.RawQuery = query.Encode()

	resp, err := s.client.Do(request)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func (s *testServer) GetToken(t *testing.T) string {
	t.Helper()
	request, err := http.NewRequest(http.MethodGet, s.URL+"/auth_token", nil)
	require.NoError(t, err)

	resp, err := s.client.Do(request)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	token, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return string(token)
}

func (s *testServer) SetSessionCheckFunc(f func(t *testing.T, c *gin.Context)) {
	s.sessionCheckFunc = f
}

func (s *testServer) Unauthorize(t *testing.T) {
	t.Helper()
	request, err := http.NewRequest(http.MethodGet, s.URL+"/auth", nil)
	require.NoError(t, err)
	query := url.Values{}
	query.Add("expiration", strconv.Itoa(-1))
	request.URL.RawQuery = query.Encode()

	resp, err := s.client.Do(request)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func (s *testServer) TestCall(t *testing.T, token ...string) (string, bool) {
	t.Helper()
	id := strconv.Itoa(rand.Int())

	request, err := http.NewRequest(http.MethodGet, s.URL+s.testEndpoint+"?id="+id, nil)
	require.NoError(t, err)
	if len(token) == 1 {
		request.Header.Add("Authorization", token[0])
	}

	resp, err := s.client.Do(request)
	require.NoError(t, err)

	assert.True(t, resp.StatusCode == http.StatusOK ||
		resp.StatusCode == http.StatusForbidden)

	return id, resp.StatusCode == http.StatusOK
}

func (s *testServer) TestCallWithData(t *testing.T, data string) (string, bool) {
	t.Helper()
	id := strconv.Itoa(rand.Int())

	request, err := http.NewRequest(http.MethodGet, s.URL+s.testEndpoint+"?id="+id, bytes.NewBufferString(data))
	require.NoError(t, err)

	resp, err := s.client.Do(request)
	require.NoError(t, err)

	assert.True(t, resp.StatusCode == http.StatusOK ||
		resp.StatusCode == http.StatusForbidden)

	return id, resp.StatusCode == http.StatusOK
}

func (s *testServer) Called(id string) bool {
	_, ok := s.calls[id]
	return ok
}

func (s *testServer) CallAndGetStatus(t *testing.T, token ...string) bool {
	t.Helper()
	id, ok := s.TestCall(t, token...)
	assert.Equal(t, ok, s.Called(id))
	return ok
}

func setTestSession(t *testing.T, c *gin.Context, privileges []string, expires int) {
	t.Helper()
	session := sessions.Default(c)
	session.Set("uid", uint64(1))
	session.Set("uhash", "testhash")
	session.Set("rid", uint64(2))
	session.Set("tid", models.UserTypeLocal.String())
	session.Set("prm", privileges)
	session.Set("gtm", time.Now().Unix())
	session.Set("exp", time.Now().Add(time.Duration(expires)*time.Second).Unix())
	session.Set("uuid", "uuid1")
	session.Set("uname", "User 1")
	session.Options(sessions.Options{
		HttpOnly: true,
		MaxAge:   expires,
	})
	require.NoError(t, session.Save())
}
