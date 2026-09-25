package auth

import (
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync"
	"testing"
	"time"

	"pentagi/pkg/server/models"
	"pentagi/pkg/version"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/stretchr/testify/require"
)

// setupTestDB seeds users 1 ("testhash") and 2 ("testhash2") and the privileges of roles 1 and 2, not 3.
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	// every pooled connection to ":memory:" opens its own empty database
	db.DB().SetMaxOpenConns(1)

	for _, stmt := range []string{
		`CREATE TABLE privileges (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			role_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			UNIQUE(role_id, name)
		)`,
		`CREATE TABLE api_tokens (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			token_id TEXT NOT NULL UNIQUE,
			user_id INTEGER NOT NULL,
			role_id INTEGER NOT NULL,
			name TEXT,
			ttl INTEGER NOT NULL,
			status TEXT NOT NULL DEFAULT 'active',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME
		)`,
		`CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			hash TEXT NOT NULL UNIQUE,
			type TEXT NOT NULL DEFAULT 'local',
			mail TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'active',
			role_id INTEGER NOT NULL DEFAULT 2,
			session_generation INTEGER NOT NULL DEFAULT 1,
			password TEXT,
			password_change_required BOOLEAN NOT NULL DEFAULT false,
			provider TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME
		)`,
		`INSERT INTO privileges (role_id, name) VALUES
			(1, 'users.create'), (1, 'users.delete'), (1, 'users.edit'), (1, 'users.view'),
			(1, 'roles.view'), (1, 'flows.admin'), (1, 'flows.create'), (1, 'flows.delete'),
			(1, 'flows.edit'), (1, 'flows.view'), (1, 'settings.tokens.create'),
			(1, 'settings.tokens.view'), (1, 'settings.tokens.edit'), (1, 'settings.tokens.delete'),
			(1, 'settings.tokens.admin'),
			(2, 'roles.view'), (2, 'flows.create'), (2, 'flows.delete'), (2, 'flows.edit'),
			(2, 'flows.view'), (2, 'settings.tokens.create'), (2, 'settings.tokens.view'),
			(2, 'settings.tokens.edit'), (2, 'settings.tokens.delete')`,
		`INSERT INTO users (id, hash, mail, name, status, role_id) VALUES
			(1, 'testhash', 'user1@test.com', 'User 1', 'active', 2),
			(2, 'testhash2', 'user2@test.com', 'User 2', 'active', 2)`,
	} {
		require.NoError(t, db.Exec(stmt).Error)
	}

	return db
}

func authStoreToken(t *testing.T, db *gorm.DB, tokenID string, uid, roleID uint64) {
	t.Helper()

	require.NoError(t, db.Create(&models.APIToken{
		TokenID: tokenID,
		UserID:  uid,
		RoleID:  roleID,
		TTL:     3600,
		Status:  models.TokenStatusActive,
	}).Error)
}

// authIssueToken stores an active role-3 token for the user and signs it under salt, claiming role 2 for an hour.
func authIssueToken(t *testing.T, db *gorm.DB, salt string, uid uint64, uhash string) string {
	t.Helper()

	tokenID, err := GenerateTokenID()
	require.NoError(t, err)
	authStoreToken(t, db, tokenID, uid, 3)

	token, err := MakeAPIToken(salt, MakeAPITokenClaims(tokenID, uhash, uid, 2, 3600))
	require.NoError(t, err)

	return token
}

func authWaitOrFail(t *testing.T, wg *sync.WaitGroup, what string) {
	t.Helper()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(30 * time.Second):
		t.Fatalf("%s did not finish within 30s", what)
	}
}

// authIdentity is what the protected handler found in the request context.
type authIdentity struct {
	UID, RID, SGN                uint64
	UHash, TID, CPT, UName, UUID string
	Prm                          []string
	GTM, EXP                     int64
}

func (id authIdentity) withoutClock() authIdentity {
	id.GTM, id.EXP = 0, 0
	return id
}

type authCall struct {
	status int
	// reason is the refused credential's error, which newTestServer's develop mode puts in the body
	reason string
	seen   *authIdentity
}

// testServer serves /protected behind the middlewares and /auth, which issues a local login's session cookie.
type testServer struct {
	*httptest.Server
	client *http.Client
	seen   *authIdentity
}

func newTestServer(t *testing.T, middlewares ...gin.HandlerFunc) *testServer {
	t.Helper()

	// authCall.reason is read from a body field that only a develop-mode build fills
	packageVer := version.PackageVer
	version.PackageVer = ""
	t.Cleanup(func() { version.PackageVer = packageVer })

	server := &testServer{}

	router := gin.New()
	router.Use(sessions.Sessions("auth", cookie.NewStore(MakeCookieStoreKey("test")...)))

	router.GET("/auth", func(c *gin.Context) {
		privileges, _ := c.GetQueryArray("privileges")
		expiresIn, err := strconv.Atoi(c.Query("expires_in"))
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		now := time.Now()
		session := sessions.Default(c)
		session.Set("uid", uint64(1))
		session.Set("uhash", "testhash")
		session.Set("rid", uint64(2))
		session.Set("tid", models.UserTypeLocal.String())
		session.Set("prm", privileges)
		session.Set("gtm", now.Unix())
		session.Set("exp", now.Add(time.Duration(expiresIn)*time.Second).Unix())
		session.Set("uname", "User 1")
		session.Set("sgn", uint64(1))
		if err := session.Save(); err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
		}
	})

	protected := router.Group("", middlewares...)
	protected.GET("/protected", func(c *gin.Context) {
		server.seen = &authIdentity{
			UID:   c.GetUint64("uid"),
			RID:   c.GetUint64("rid"),
			SGN:   c.GetUint64("sgn"),
			UHash: c.GetString("uhash"),
			TID:   c.GetString("tid"),
			CPT:   c.GetString("cpt"),
			UName: c.GetString("uname"),
			UUID:  c.GetString("uuid"),
			Prm:   c.GetStringSlice("prm"),
			GTM:   c.GetInt64("gtm"),
			EXP:   c.GetInt64("exp"),
		}
	})

	server.Server = httptest.NewServer(router)
	t.Cleanup(server.Close)

	server.client = server.Client()
	server.client.Timeout = 30 * time.Second
	jar, err := cookiejar.New(nil)
	require.NoError(t, err)
	server.client.Jar = jar

	return server
}

// Authorize stores a session cookie for user 1 (hash "testhash", generation 1) in the client.
func (s *testServer) Authorize(t *testing.T, expiresIn time.Duration, privileges ...string) {
	t.Helper()

	query := url.Values{"expires_in": {strconv.Itoa(int(expiresIn / time.Second))}}
	for _, p := range privileges {
		query.Add("privileges", p)
	}

	resp, err := s.client.Get(s.URL + "/auth?" + query.Encode())
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func (s *testServer) Call(t *testing.T, authorization string) authCall {
	t.Helper()

	s.seen = nil
	request, err := http.NewRequest(http.MethodGet, s.URL+"/protected", nil)
	require.NoError(t, err)
	if authorization != "" {
		request.Header.Set("Authorization", authorization)
	}

	resp, err := s.client.Do(request)
	require.NoError(t, err)
	defer resp.Body.Close()

	call := authCall{status: resp.StatusCode, seen: s.seen}
	if resp.StatusCode != http.StatusOK {
		var body struct {
			Error string `json:"error"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
		call.reason = body.Error
	}

	return call
}
