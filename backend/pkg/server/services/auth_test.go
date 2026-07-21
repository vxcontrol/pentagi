package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"pentagi/pkg/server/models"
	"pentagi/pkg/server/oauth"
)

type fakeOAuthClient struct {
	name     string
	email    string
	verified bool
}

func (f *fakeOAuthClient) ProviderName() string { return f.name }

func (f *fakeOAuthClient) ResolveEmail(context.Context, string, *oauth2.Token) (string, bool, error) {
	return f.email, f.verified, nil
}

func (f *fakeOAuthClient) TokenSource(context.Context, *oauth2.Token) oauth2.TokenSource { return nil }

func (f *fakeOAuthClient) Exchange(context.Context, string, ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	return &oauth2.Token{AccessToken: "test-access-token", Expiry: time.Now().Add(time.Hour)}, nil
}

func (f *fakeOAuthClient) RefreshToken(context.Context, string) (*oauth2.Token, error) {
	return nil, nil
}

func (f *fakeOAuthClient) AuthCodeURL(string, ...oauth2.AuthCodeOption) string { return "" }

func newOAuthService(db *gorm.DB, email string) *AuthService {
	return newOAuthServiceVerified(db, email, true)
}

func newOAuthServiceVerified(db *gorm.DB, email string, verified bool) *AuthService {
	return &AuthService{
		cfg:   AuthServiceConfig{BaseURL: "/", SessionTimeout: 3600},
		db:    db,
		key:   []byte("0123456789abcdef0123456789abcdef"),
		oauth: map[string]oauth.OAuthClient{"github": &fakeOAuthClient{name: "github", email: email, verified: verified}},
	}
}

func newCallbackContext(t *testing.T) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest(http.MethodGet, "/callback", nil)
	req.AddCookie(&http.Cookie{Name: authNonceCookieName, Value: "test-nonce"})
	c.Request = req

	sessions.Sessions("pentagi", cookie.NewStore([]byte("test-secret")))(c)

	return c, w
}

func countUsers(t *testing.T, db *gorm.DB) int {
	t.Helper()
	var count int
	require.NoError(t, db.Model(&models.User{}).Count(&count).Error)
	return count
}

// TestAuthLoginCallback_LinksExistingLocalAccount is the regression guard for the OAuth
// squat DoS: a first OAuth login for an email already held by a LOCAL account must link
// into that account (logging the verified owner in and recording the provider) rather than
// taking the create branch, which would violate users_mail_unique and return 500.
func TestAuthLoginCallback_LinksExistingLocalAccount(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	require.NoError(t, db.Exec(
		"INSERT INTO users (id, hash, type, mail, name, status, role_id, provider) VALUES (10, ?, 'local', 'victim@corp.com', 'Victim', 'active', 2, NULL)",
		"1234567890abcdef1234567890abcdef",
	).Error)

	before := countUsers(t, db)

	svc := newOAuthService(db, "victim@corp.com")
	c, w := newCallbackContext(t)
	svc.authLoginCallback(c, map[string]string{"provider": "github"}, "test-code")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, before, countUsers(t, db), "must link into the existing row, not create a duplicate")

	var linked models.User
	require.NoError(t, db.Where("mail = ?", "victim@corp.com").First(&linked).Error)
	assert.Equal(t, uint64(10), linked.ID, "the existing local row is reused")
	assert.Equal(t, models.UserTypeLocal, linked.Type, "linking keeps the password-login capability")
	require.NotNil(t, linked.Provider)
	assert.Equal(t, "github", *linked.Provider, "the provider is backfilled on link")

	assert.Equal(t, uint64(10), sessions.Default(c).Get("uid"), "session is issued for the linked account")
}

func TestAuthLoginCallback_RejectsUnverifiedEmail(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	require.NoError(t, db.Exec(
		"INSERT INTO users (id, hash, type, mail, name, status, role_id, provider) VALUES (10, ?, 'local', 'victim@corp.com', 'Victim', 'active', 2, NULL)",
		"1234567890abcdef1234567890abcdef",
	).Error)

	before := countUsers(t, db)

	svc := newOAuthServiceVerified(db, "victim@corp.com", false)
	c, w := newCallbackContext(t)
	svc.authLoginCallback(c, map[string]string{"provider": "github"}, "test-code")

	assert.NotEqual(t, http.StatusOK, w.Code, "an unverified email must be rejected, not linked")
	assert.Equal(t, before, countUsers(t, db), "no account is created")

	var victim models.User
	require.NoError(t, db.Where("id = ?", 10).First(&victim).Error)
	assert.Nil(t, victim.Provider, "the victim account is not linked to the provider")
	assert.Nil(t, sessions.Default(c).Get("uid"), "no session is issued")
}

func TestAuthLoginCallback_CreatesUserWhenEmailFree(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	before := countUsers(t, db)

	svc := newOAuthService(db, "newcomer@corp.com")
	c, w := newCallbackContext(t)
	svc.authLoginCallback(c, map[string]string{"provider": "github"}, "test-code")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, before+1, countUsers(t, db))

	var created models.User
	require.NoError(t, db.Where("mail = ?", "newcomer@corp.com").First(&created).Error)
	assert.Equal(t, models.UserTypeOAuth, created.Type)
	require.NotNil(t, created.Provider)
	assert.Equal(t, "github", *created.Provider)

	var prefCount int
	require.NoError(t, db.Table("user_preferences").Where("user_id = ?", created.ID).Count(&prefCount).Error)
	assert.Equal(t, 1, prefCount, "preferences row is created alongside the new user")
}

// TestAuthLoginCallback_RejectsBlockedAccount guards the status gate: linking must not log in a
// non-active account. Before the lookup was broadened, a blocked LOCAL row was invisible to the
// callback (type filter) and a fresh active OAuth row was created instead — the gate now applies.
func TestAuthLoginCallback_RejectsBlockedAccount(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	require.NoError(t, db.Exec(
		"INSERT INTO users (id, hash, type, mail, name, status, role_id) VALUES (11, ?, 'local', 'blocked@corp.com', 'Blocked', 'blocked', 2)",
		"1234567890abcdef1234567890abcdef",
	).Error)
	before := countUsers(t, db)

	svc := newOAuthService(db, "blocked@corp.com")
	c, w := newCallbackContext(t)
	svc.authLoginCallback(c, map[string]string{"provider": "github"}, "test-code")

	assert.Equal(t, http.StatusForbidden, w.Code, "a blocked account must not be logged in via OAuth")
	assert.Nil(t, sessions.Default(c).Get("uid"), "no session is issued for a blocked account")
	assert.Equal(t, before, countUsers(t, db), "no shadow account is created for a blocked email")
}

// TestAuthLoginCallback_LinkInheritsAccountRole pins that a linked session carries the privileges of
// the account's actual role, not a hardcoded RoleUser set — rid and prm must agree. setupTestDB seeds
// Admin (role 1) with users.create and User (role 2) without it.
func TestAuthLoginCallback_LinkInheritsAccountRole(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	require.NoError(t, db.Exec(
		"INSERT INTO users (id, hash, type, mail, name, status, role_id) VALUES (12, ?, 'local', 'admin2@corp.com', 'Admin Two', 'active', 1)",
		"1234567890abcdef1234567890abcdef",
	).Error)

	svc := newOAuthService(db, "admin2@corp.com")
	c, w := newCallbackContext(t)
	svc.authLoginCallback(c, map[string]string{"provider": "github"}, "test-code")

	require.Equal(t, http.StatusOK, w.Code)
	sess := sessions.Default(c)
	assert.Equal(t, uint64(1), sess.Get("rid"), "session role matches the linked account")
	prm, _ := sess.Get("prm").([]string)
	assert.Contains(t, prm, "users.create", "linked session carries the account role's privileges, not RoleUser")
}

// TestAuthLoginCallback_LinksOnConcurrentCreateConflict simulates the TOCTOU race where a second
// first-login for the same new email commits between this request's lookup and its insert. The
// one-shot callback hides the row from the first lookup, so the handler enters the create branch and
// trips users_mail_unique; it must then re-fetch and link rather than return a 500.
func TestAuthLoginCallback_LinksOnConcurrentCreateConflict(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	require.NoError(t, db.Exec(
		"INSERT INTO users (id, hash, type, mail, name, status, role_id) VALUES (20, ?, 'oauth', 'race@corp.com', 'Racer', 'active', 2)",
		"1234567890abcdef1234567890abcdef",
	).Error)
	before := countUsers(t, db)

	hiddenOnce := false
	db.Callback().Query().Before("gorm:query").Register("test:hide_first_user_query", func(scope *gorm.Scope) {
		if !hiddenOnce && scope.TableName() == "users" {
			hiddenOnce = true
			scope.Err(gorm.ErrRecordNotFound)
		}
	})
	defer db.Callback().Query().Remove("test:hide_first_user_query")

	svc := newOAuthService(db, "race@corp.com")
	c, w := newCallbackContext(t)
	svc.authLoginCallback(c, map[string]string{"provider": "github"}, "test-code")

	assert.Equal(t, http.StatusOK, w.Code, "a create conflict must relink, not 500")
	assert.Equal(t, before, countUsers(t, db), "no duplicate is created on the conflict")
	assert.Equal(t, uint64(20), sessions.Default(c).Get("uid"), "session is issued for the existing row")
}
