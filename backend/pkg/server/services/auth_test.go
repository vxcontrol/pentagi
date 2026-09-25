package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pentagi/pkg/server/auth"
	"pentagi/pkg/server/models"
	"pentagi/pkg/server/oauth"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func countUsers(t *testing.T, db *gorm.DB) int {
	t.Helper()
	var count int
	require.NoError(t, db.Model(&models.User{}).Count(&count).Error)
	return count
}

func TestAuth_AuthLoginCallback_SignsInOneAccountPerVerifiedActiveEmail(t *testing.T) {
	for _, tc := range []struct {
		name            string
		seed            string
		mail            string
		unverified      bool
		hideFirstLookup bool
		wantCode        int
		wantBody        string
		wantUID         any
		wantNewUsers    int
		check           func(t *testing.T, db *gorm.DB, session sessions.Session)
	}{
		{
			name: "an email a local account holds links into that account",
			seed: "INSERT INTO users (id, hash, type, mail, name, status, role_id, provider) " +
				"VALUES (10, '1234567890abcdef1234567890abcdef', 'local', 'victim@corp.com', 'Victim', 'active', 1, NULL)",
			mail:     "victim@corp.com",
			wantCode: http.StatusOK,
			wantUID:  uint64(10),
			check: func(t *testing.T, db *gorm.DB, session sessions.Session) {
				var linked models.User
				require.NoError(t, db.Where("mail = ?", "victim@corp.com").First(&linked).Error)
				assert.Equal(t, uint64(10), linked.ID, "the existing local row is reused")
				assert.Equal(t, models.UserTypeLocal, linked.Type, "linking keeps the password-login capability")
				require.NotNil(t, linked.Provider)
				assert.Equal(t, "github", *linked.Provider, "the provider is backfilled on link")
				assert.Equal(t, uint64(1), session.Get("rid"), "the session role matches the linked account")
				prm, _ := session.Get("prm").([]string)
				assert.Contains(t, prm, "users.create", "the session carries the account role's privileges, not RoleUser's")
			},
		},
		{
			name: "an unverified email is refused",
			seed: "INSERT INTO users (id, hash, type, mail, name, status, role_id, provider) " +
				"VALUES (10, '1234567890abcdef1234567890abcdef', 'local', 'victim@corp.com', 'Victim', 'active', 2, NULL)",
			mail:       "victim@corp.com",
			unverified: true,
			wantCode:   http.StatusInternalServerError,
			wantBody:   `"code":"Auth.InvalidUserData"`,
			check: func(t *testing.T, db *gorm.DB, _ sessions.Session) {
				var victim models.User
				require.NoError(t, db.Where("id = ?", 10).First(&victim).Error)
				assert.Nil(t, victim.Provider, "the account is not linked to the provider")
			},
		},
		{
			name:         "a free email creates an oauth account",
			mail:         "newcomer@corp.com",
			wantCode:     http.StatusOK,
			wantUID:      uint64(3),
			wantNewUsers: 1,
			check: func(t *testing.T, db *gorm.DB, _ sessions.Session) {
				var created models.User
				require.NoError(t, db.Where("mail = ?", "newcomer@corp.com").First(&created).Error)
				assert.Equal(t, models.UserTypeOAuth, created.Type)
				require.NotNil(t, created.Provider)
				assert.Equal(t, "github", *created.Provider)

				var prefCount int
				require.NoError(t, db.Table("user_preferences").Where("user_id = ?", created.ID).Count(&prefCount).Error)
				assert.Equal(t, 1, prefCount, "the preferences row is created alongside the new user")
			},
		},
		{
			name: "a blocked account is refused",
			seed: "INSERT INTO users (id, hash, type, mail, name, status, role_id) " +
				"VALUES (11, '1234567890abcdef1234567890abcdef', 'local', 'blocked@corp.com', 'Blocked', 'blocked', 2)",
			mail:     "blocked@corp.com",
			wantCode: http.StatusForbidden,
			wantBody: `"code":"Auth.InactiveUser"`,
		},
		{
			name: "a row a concurrent first login created is linked, not duplicated",
			seed: "INSERT INTO users (id, hash, type, mail, name, status, role_id) " +
				"VALUES (20, '1234567890abcdef1234567890abcdef', 'oauth', 'race@corp.com', 'Racer', 'active', 2)",
			mail:            "race@corp.com",
			hideFirstLookup: true,
			wantCode:        http.StatusOK,
			wantUID:         uint64(20),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := setupTestDB(t)
			defer db.Close()
			if tc.seed != "" {
				require.NoError(t, db.Exec(tc.seed).Error)
			}
			before := countUsers(t, db)

			if tc.hideFirstLookup {
				hidden := false
				db.Callback().Query().Before("gorm:query").Register("test:hide_first_user_query", func(scope *gorm.Scope) {
					if !hidden && scope.TableName() == "users" {
						hidden = true
						_ = scope.Err(gorm.ErrRecordNotFound)
					}
				})
				defer db.Callback().Query().Remove("test:hide_first_user_query")
			}

			c, w := authOAuthCallback(t, newOAuthServiceVerified(db, tc.mail, !tc.unverified), nil)

			assert.Equal(t, tc.wantCode, w.Code, w.Body.String())
			assert.Contains(t, w.Body.String(), tc.wantBody)
			assert.Equal(t, before+tc.wantNewUsers, countUsers(t, db), "one email is one account")
			assert.Equal(t, tc.wantUID, sessions.Default(c).Get("uid"))
			if tc.check != nil {
				tc.check(t, db, sessions.Default(c))
			}
		})
	}
}

type infoBody struct {
	Type string
	User struct {
		ID   uint64
		Mail string
	}
	ExpiresAt time.Time `json:"expires_at"`
}

func infoOf(t *testing.T, rec *httptest.ResponseRecorder) infoBody {
	t.Helper()
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var body struct{ Data infoBody }
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))

	return body.Data
}

func errorCodeOf(t *testing.T, body []byte) string {
	t.Helper()

	var resp struct {
		Code string `json:"code"`
	}
	require.NoError(t, json.Unmarshal(body, &resp))

	return resp.Code
}

func TestAuth_Info_AnswersTheKindOfCaller(t *testing.T) {
	for _, tc := range []struct {
		name     string
		apiToken bool
		wantType string
	}{
		{"a signed-in cookie is a user", false, "user"},
		{"an api token is an api client, with no session generation to compare", true, "api"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := newSessionHarness(t)
			req := httptest.NewRequest(http.MethodGet, "/info", nil)
			var cookies []*http.Cookie
			if tc.apiToken {
				req.Header.Set("Authorization", "Bearer "+h.apiToken(t))
			} else {
				cookies = h.login(t)
			}

			body := infoOf(t, h.send(req, cookies))

			assert.Equal(t, tc.wantType, body.Type)
			assert.Equal(t, h.mail, body.User.Mail)
		})
	}
}

func TestAuth_Info_DoesNotReMintARevokedSession(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	seedLocalUser(t, db, "known@corp.com")
	var user struct {
		ID   uint64
		Hash string
	}
	require.NoError(t, db.Raw("SELECT id, hash FROM users WHERE mail = ?", "known@corp.com").Scan(&user).Error)
	require.NoError(t, db.Exec("UPDATE users SET session_generation = 2 WHERE id = ?", user.ID).Error)

	gin.SetMode(gin.TestMode)
	svc := NewAuthService(AuthServiceConfig{BaseURL: "/", SessionTimeout: 3600}, db, map[string]oauth.OAuthClient{},
		auth.NewUserCache(db))

	engine := gin.New()
	engine.Use(sessions.Sessions("pentagi", cookie.NewStore([]byte("test-secret"))))
	mintedBeforeTheRefreshWindow := time.Now().Add(-10 * time.Minute).Unix()
	engine.GET("/info", func(c *gin.Context) {
		c.Set("uid", user.ID)
		c.Set("uhash", user.Hash)
		c.Set("tid", "local")
		c.Set("exp", time.Now().Add(time.Hour).Unix())
		c.Set("gtm", mintedBeforeTheRefreshWindow)
		c.Set("prm", []string{"flows.view"})
		c.Set("sgn", uint64(1))
	}, svc.Info)

	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/info", nil))

	assert.Equal(t, "guest", infoOf(t, rec).Type, "a revoked session is a guest, as the middleware would have said")
	for _, c := range rec.Result().Cookies() {
		assert.NotEqual(t, "pentagi", c.Name, "a revoked session must not be re-minted at the new generation")
	}
}

func TestAuth_Info_SignsOutASessionWhoseGenerationMovedOutsideTheServer(t *testing.T) {
	h := newSessionHarness(t)
	cookies := h.login(t)
	require.Equal(t, http.StatusOK, h.get("/private", cookies).Code, "the account's generation is cached")
	require.NoError(t, h.db.Exec("UPDATE users SET session_generation = session_generation + 1 WHERE mail = ?",
		h.mail).Error)

	body := infoOf(t, h.get("/info", cookies))

	assert.Equal(t, "guest", body.Type)
	assert.Zero(t, body.User.ID, "a guest carries no account")
	assert.True(t, body.ExpiresAt.Before(time.Now()), "the web client reads a future expiry as signed in")
	assert.NotEqual(t, http.StatusOK, h.get("/private", cookies).Code,
		"once /info has seen the moved generation, no route takes the cookie")
}

func TestAuth_Info_KeepsTheSessionThroughAUserStoreOutage(t *testing.T) {
	h := newSessionHarness(t)
	cookies := h.login(t)
	require.NoError(t, h.db.Exec("ALTER TABLE users RENAME TO users_away").Error)

	private := h.get("/private", cookies)
	info := h.get("/info", cookies)
	guest := h.get("/info", nil)

	require.NoError(t, h.db.Exec("ALTER TABLE users_away RENAME TO users").Error)
	assert.Equal(t, http.StatusServiceUnavailable, private.Code, "the browser treats 403 AuthRequired as a lost session")
	assert.Equal(t, "AuthUnavailable", errorCodeOf(t, private.Body.Bytes()))
	assert.Equal(t, http.StatusServiceUnavailable, info.Code, "a guest answer tells the browser it is signed out")
	assert.Equal(t, "AuthUnavailable", errorCodeOf(t, info.Body.Bytes()))
	assert.Equal(t, http.StatusOK, guest.Code, "a browser without a session is still a guest")
	assert.Equal(t, http.StatusOK, h.get("/private", cookies).Code, "the session outlives the outage")
}

func TestAuth_Info_AnswersUnavailableWhenItsOwnReadFailsBehindAWarmCache(t *testing.T) {
	h := newSessionHarness(t)
	cookies := h.login(t)
	require.Equal(t, http.StatusOK, h.get("/private", cookies).Code)
	require.NoError(t, h.db.Exec("ALTER TABLE users RENAME TO users_away").Error)

	info := h.get("/info", cookies)

	require.NoError(t, h.db.Exec("ALTER TABLE users_away RENAME TO users").Error)
	assert.Equal(t, http.StatusServiceUnavailable, info.Code, "a store that cannot answer is not a missing user")
	assert.Equal(t, "AuthUnavailable", errorCodeOf(t, info.Body.Bytes()))
}

type loginCaller struct {
	engine *gin.Engine
	addr   string
	xff    string
}

func newLoginService(t *testing.T, db *gorm.DB, cfg AuthServiceConfig) (*AuthService, func(string) *loginCaller) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	cfg.BaseURL = "/"
	cfg.SessionTimeout = 3600
	svc := NewAuthService(cfg, db, map[string]oauth.OAuthClient{}, nil)

	engine := gin.New()
	require.NoError(t, engine.SetTrustedProxies(nil))
	engine.Use(sessions.Sessions("pentagi", cookie.NewStore([]byte("test-secret"))))
	engine.POST("/auth/login", svc.AuthLogin)

	return svc, func(addr string) *loginCaller {
		return &loginCaller{engine: engine, addr: addr}
	}
}

func (c *loginCaller) forwardedFor(value string) *loginCaller {
	return &loginCaller{engine: c.engine, addr: c.addr, xff: value}
}

func (c *loginCaller) login(t *testing.T, mail, password string) *httptest.ResponseRecorder {
	t.Helper()

	body, err := json.Marshal(map[string]string{"mail": mail, "password": password})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = c.addr + ":54321"
	if c.xff != "" {
		req.Header.Set("X-Forwarded-For", c.xff)
	}

	w := httptest.NewRecorder()
	c.engine.ServeHTTP(w, req)

	return w
}

func TestAuth_AuthLogin_AnswersAnUnknownAccountLikeAWrongPassword(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	seedLocalUser(t, db, "known@corp.com")
	svc, from := newLoginService(t, db, AuthServiceConfig{})

	calibration := time.Now()
	_ = bcrypt.CompareHashAndPassword(svc.dummyHash, []byte("calibration"))
	floor := time.Since(calibration) / 2

	start := time.Now()
	miss := from("10.0.0.1").login(t, "nobody@corp.com", "WrongPass1!")
	elapsed := time.Since(start)
	wrong := from("10.0.0.2").login(t, "known@corp.com", "WrongPass1!")

	require.Equal(t, http.StatusUnauthorized, miss.Code)
	require.Equal(t, http.StatusUnauthorized, wrong.Code)
	assert.JSONEq(t, miss.Body.String(), wrong.Body.String(),
		"the answer must not tell an unknown account from a wrong password")
	assert.GreaterOrEqual(t, elapsed, floor,
		"a miss has to cost the same hashing work as a wrong password, or the delay names the account")
}

func TestAuth_AuthLogin_LocksOutTheAttackerNotTheOwner(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	seedLocalUser(t, db, "victim@corp.com")
	_, from := newLoginService(t, db, AuthServiceConfig{
		LoginPairAttemptLimit: 3,
		LoginAddrAttemptLimit: 100,
		LoginFailureWindow:    time.Minute,
		LoginLockout:          time.Hour,
	})

	attacker := from("10.0.0.9")
	for i := 1; i <= 3; i++ {
		require.Equal(t, http.StatusUnauthorized, attacker.login(t, "victim@corp.com", "WrongPass1!").Code,
			"attempt %d is within the limit", i)
	}

	locked := attacker.login(t, "victim@corp.com", knownAccountPassword)
	assert.Equal(t, http.StatusTooManyRequests, locked.Code,
		"past the limit the attacker is refused even with the right password")
	assert.NotEmpty(t, locked.Header().Get("Retry-After"))

	owner := from("10.0.0.2").login(t, "victim@corp.com", knownAccountPassword)
	assert.Equal(t, http.StatusOK, owner.Code,
		"the owner elsewhere has to get in while the attacker is locked out")
}

func TestAuth_AuthLogin_IgnoresAForwardedForFromAnUntrustedPeer(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	seedLocalUser(t, db, "target@corp.com")
	_, from := newLoginService(t, db, AuthServiceConfig{
		LoginPairAttemptLimit: 100,
		LoginAddrAttemptLimit: 3,
		LoginFailureWindow:    time.Minute,
		LoginLockout:          time.Hour,
	})

	attacker := from("10.0.0.9")
	for i := 0; i < 4; i++ {
		attacker.forwardedFor(fmt.Sprintf("203.0.113.%d", i)).login(t, "target@corp.com", "WrongPass1!")
	}

	refused := attacker.forwardedFor("203.0.113.250").login(t, "target@corp.com", "WrongPass1!")
	assert.Equal(t, http.StatusTooManyRequests, refused.Code,
		"a rotating X-Forwarded-For must not buy a fresh address budget")

	bystander := from("198.51.100.4").login(t, "target@corp.com", knownAccountPassword)
	assert.Equal(t, http.StatusOK, bystander.Code,
		"and a forged address must not lock out whoever really holds it")
}

func TestAuth_AuthLogin_LetsTheOwnerBackInAfterTheLockout(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	seedLocalUser(t, db, "owner@corp.com")
	_, from := newLoginService(t, db, AuthServiceConfig{
		LoginPairAttemptLimit: 2,
		LoginAddrAttemptLimit: 100,
		LoginFailureWindow:    time.Minute,
		LoginLockout:          time.Nanosecond,
	})

	owner := from("10.0.0.3")
	for i := 0; i < 3; i++ {
		owner.login(t, "owner@corp.com", "WrongPass1!")
	}

	assert.Equal(t, http.StatusOK, owner.login(t, "owner@corp.com", knownAccountPassword).Code,
		"the lockout has to end on its own")
}

func TestAuth_AuthLogin_ClearsTheCountOnSuccess(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	seedLocalUser(t, db, "clears@corp.com")
	_, from := newLoginService(t, db, AuthServiceConfig{
		LoginPairAttemptLimit: 3,
		LoginAddrAttemptLimit: 100,
		LoginFailureWindow:    time.Minute,
		LoginLockout:          time.Hour,
	})

	caller := from("10.0.0.4")
	require.Equal(t, http.StatusUnauthorized, caller.login(t, "clears@corp.com", "WrongPass1!").Code)
	require.Equal(t, http.StatusUnauthorized, caller.login(t, "clears@corp.com", "WrongPass1!").Code)
	require.Equal(t, http.StatusOK, caller.login(t, "clears@corp.com", knownAccountPassword).Code)

	require.Equal(t, http.StatusUnauthorized, caller.login(t, "clears@corp.com", "WrongPass1!").Code)
	require.Equal(t, http.StatusUnauthorized, caller.login(t, "clears@corp.com", "WrongPass1!").Code)

	assert.Equal(t, http.StatusOK, caller.login(t, "clears@corp.com", knownAccountPassword).Code,
		"a successful login has to clear the attempts before it")
}

func TestAuth_AuthLogin_SpendsTheAttemptOfAnInactiveAccount(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	seedLocalUser(t, db, "blocked@corp.com")
	require.NoError(t, db.Exec("UPDATE users SET status = 'blocked' WHERE mail = ?", "blocked@corp.com").Error)

	_, from := newLoginService(t, db, AuthServiceConfig{
		LoginPairAttemptLimit: 2,
		LoginAddrAttemptLimit: 100,
		LoginFailureWindow:    time.Minute,
		LoginLockout:          time.Hour,
	})

	caller := from("10.0.0.9")
	require.Equal(t, http.StatusForbidden, caller.login(t, "blocked@corp.com", knownAccountPassword).Code)
	require.Equal(t, http.StatusForbidden, caller.login(t, "blocked@corp.com", knownAccountPassword).Code)

	assert.Equal(t, http.StatusTooManyRequests, caller.login(t, "blocked@corp.com", knownAccountPassword).Code,
		"a password that cannot log in must not hand the attempt back")
}

func TestAuth_AuthLogin_ChargesNothingToTheAccountForAnAddressRefusal(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	seedLocalUser(t, db, "owner2@corp.com")
	seedLocalUser(t, db, "noise@corp.com")
	svc, from := newLoginService(t, db, AuthServiceConfig{
		LoginPairAttemptLimit: 3,
		LoginAddrAttemptLimit: 2,
		LoginFailureWindow:    time.Minute,
		LoginLockout:          time.Hour,
	})

	caller := from("10.0.0.12")
	for i := 1; i <= 3; i++ {
		caller.login(t, "noise@corp.com", "WrongPass1!")
	}
	require.Equal(t, http.StatusTooManyRequests, caller.login(t, "owner2@corp.com", knownAccountPassword).Code,
		"the address has to be locked for the case to mean anything")

	for i := 1; i <= 4; i++ {
		require.Equal(t, http.StatusTooManyRequests, caller.login(t, "owner2@corp.com", knownAccountPassword).Code)
	}

	svc.addrGuard.Reset("10.0.0.12")

	assert.Equal(t, http.StatusOK, caller.login(t, "owner2@corp.com", knownAccountPassword).Code,
		"attempts the address refused must not be charged to the account")
}

func TestAuth_AuthLogin_ChargesNothingToTheAddressForAnAccountRefusal(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	seedLocalUser(t, db, "victim@corp.com")
	seedLocalUser(t, db, "colleague@corp.com")
	_, from := newLoginService(t, db, AuthServiceConfig{
		LoginPairAttemptLimit: 2,
		LoginAddrAttemptLimit: 6,
		LoginFailureWindow:    time.Minute,
		LoginLockout:          time.Hour,
	})

	office := from("10.0.0.1")
	for i := 1; i <= 2; i++ {
		require.Equal(t, http.StatusUnauthorized, office.login(t, "victim@corp.com", "WrongPass1!").Code)
	}
	for i := 1; i <= 10; i++ {
		require.Equal(t, http.StatusTooManyRequests, office.login(t, "victim@corp.com", "WrongPass1!").Code,
			"attempt %d is over the account limit", i)
	}

	colleague := office.login(t, "colleague@corp.com", knownAccountPassword)
	assert.Equal(t, http.StatusOK, colleague.Code,
		"attempts the account limit already refused must not use up the address for everyone else")
}

func TestAuth_NewAuthService_TakesTheDefaultsForAZeroLoginConfig(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	svc := NewAuthService(AuthServiceConfig{}, db, map[string]oauth.OAuthClient{}, nil)

	assert.Equal(t, 10, svc.cfg.LoginPairAttemptLimit)
	assert.Equal(t, 30, svc.cfg.LoginAddrAttemptLimit)
	assert.Equal(t, 5*time.Minute, svc.cfg.LoginFailureWindow)
	assert.Equal(t, 15*time.Minute, svc.cfg.LoginLockout,
		"a zero field means the default, not an unlimited login endpoint")
}

var authLogoutDoors = []struct {
	name   string
	target string
	okCode int
}{
	{"through the logout", "/auth/logout", http.StatusOK},
	{"through the provider's logout callback", "/auth/logout-callback", http.StatusSeeOther},
}

func clearsTheSession(rec *httptest.ResponseRecorder) bool {
	for _, c := range rec.Result().Cookies() {
		if c.Name == "pentagi" && c.MaxAge < 0 {
			return true
		}
	}

	return false
}

func TestAuth_RevokeCallerSessions_EndsEverySessionOfTheUser(t *testing.T) {
	for _, door := range authLogoutDoors {
		t.Run(door.name, func(t *testing.T) {
			h := newSessionHarness(t)
			cookies := h.login(t)
			before := h.generation(t)

			loggedOut := h.post(door.target, cookies)

			assert.Equal(t, door.okCode, loggedOut.Code)
			assert.Equal(t, before+1, h.generation(t), "logging out has to end the sessions the user holds elsewhere")
		})
	}
}

func TestAuth_RevokeCallerSessions_FailsWhenTheSessionsCannotBeEnded(t *testing.T) {
	for _, door := range authLogoutDoors {
		t.Run(door.name, func(t *testing.T) {
			h := newSessionHarness(t)
			cookies := h.login(t)
			require.NoError(t, h.db.Exec("CREATE TRIGGER refuse_generation BEFORE UPDATE OF session_generation ON users "+
				"BEGIN SELECT RAISE(FAIL, 'generation locked'); END").Error)

			loggedOut := h.post(door.target, cookies)

			assert.Equal(t, http.StatusInternalServerError, loggedOut.Code,
				"a logout that could not end the other sessions must not report success")
			assert.True(t, clearsTheSession(loggedOut), "this browser is signed out even so")
		})
	}
}

func TestAuth_RevokeCallerSessions_EndsNoOtherSessionForARevokedCookie(t *testing.T) {
	for _, door := range authLogoutDoors {
		t.Run(door.name, func(t *testing.T) {
			h := newSessionHarness(t)
			stale := h.login(t)
			require.NoError(t, h.db.Exec("UPDATE users SET session_generation = session_generation + 1 WHERE mail = ?",
				h.mail).Error)
			before := h.generation(t)

			h.post(door.target, stale)

			assert.Equal(t, before, h.generation(t), "a cookie that no longer authenticates must not end the live sessions")
		})
	}
}

func TestAuth_RevokeCallerSessions_SignsOutTheBrowserOfADeletedAccount(t *testing.T) {
	h := newSessionHarness(t)
	cookies := h.login(t)
	require.Equal(t, http.StatusOK, h.get("/private", cookies).Code, "the cookie is accepted and the account cached")
	require.NoError(t, h.db.Exec("DELETE FROM users WHERE mail = ?", h.mail).Error)

	loggedOut := h.post("/auth/logout", cookies)

	assert.Equal(t, http.StatusOK, loggedOut.Code, "an account that is gone has no sessions left to end")
	assert.True(t, clearsTheSession(loggedOut))
}

func TestAuth_RevokeCallerSessions_FailsWhenTheCookieCannotBeChecked(t *testing.T) {
	h := newSessionHarness(t)
	cookies := h.login(t)
	require.NoError(t, h.db.Exec("ALTER TABLE users RENAME TO users_away").Error)

	loggedOut := h.post("/auth/logout", cookies)

	require.NoError(t, h.db.Exec("ALTER TABLE users_away RENAME TO users").Error)
	assert.Equal(t, http.StatusInternalServerError, loggedOut.Code,
		"a logout that could not check whose sessions to end must not report them ended")
	assert.True(t, clearsTheSession(loggedOut), "this browser is signed out even so")
}

func TestAuth_RevokeCallerSessions_EndsNoBrowserSessionForAnAPIToken(t *testing.T) {
	h := newSessionHarness(t)
	cookies := h.login(t)
	before := h.generation(t)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+h.apiToken(t))
	h.send(req, nil)

	assert.Equal(t, before, h.generation(t), "an API token is not a browser session and ends none")
	assert.Equal(t, http.StatusOK, h.get("/private", cookies).Code)
}

func newSessionHarness(t *testing.T) *sessionHarness {
	t.Helper()

	db := setupTestDB(t)
	t.Cleanup(func() { db.Close() })
	seedLocalUser(t, db, "known@corp.com")

	gin.SetMode(gin.TestMode)
	userCache := auth.NewUserCache(db)
	svc := NewAuthService(AuthServiceConfig{BaseURL: "/", SessionTimeout: 3600}, db, map[string]oauth.OAuthClient{}, userCache)
	middleware := auth.NewAuthMiddleware("/", "test", auth.NewTokenCache(db), userCache)
	engine := gin.New()
	require.NoError(t, engine.SetTrustedProxies(nil))
	engine.Use(sessions.Sessions("pentagi", cookie.NewStore([]byte("test-secret"))))
	engine.Use(middleware.TryAuth)
	engine.POST("/auth/login", svc.AuthLogin)
	engine.POST("/auth/logout", svc.AuthLogout)
	engine.POST("/auth/logout-callback", svc.AuthLogoutCallback)
	engine.GET("/info", svc.Info)
	engine.GET("/private", middleware.AuthUserRequired, func(c *gin.Context) { c.Status(http.StatusOK) })

	return &sessionHarness{engine: engine, db: db, mail: "known@corp.com"}
}

type sessionHarness struct {
	engine *gin.Engine
	db     *gorm.DB
	mail   string
}

func (h *sessionHarness) login(t *testing.T) []*http.Cookie {
	t.Helper()

	body, err := json.Marshal(map[string]string{"mail": h.mail, "password": knownAccountPassword})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "10.0.0.1:54321"

	rec := httptest.NewRecorder()
	h.engine.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	return rec.Result().Cookies()
}

func (h *sessionHarness) post(target string, cookies []*http.Cookie) *httptest.ResponseRecorder {
	return h.send(httptest.NewRequest(http.MethodPost, target, nil), cookies)
}

func (h *sessionHarness) get(target string, cookies []*http.Cookie) *httptest.ResponseRecorder {
	return h.send(httptest.NewRequest(http.MethodGet, target, nil), cookies)
}

func (h *sessionHarness) send(req *http.Request, cookies []*http.Cookie) *httptest.ResponseRecorder {
	for _, c := range cookies {
		req.AddCookie(c)
	}

	rec := httptest.NewRecorder()
	h.engine.ServeHTTP(rec, req)

	return rec
}

func (h *sessionHarness) apiToken(t *testing.T) string {
	t.Helper()

	var user struct {
		ID     uint64
		Hash   string
		RoleID uint64
	}
	require.NoError(t, h.db.Raw("SELECT id, hash, role_id FROM users WHERE mail = ?", h.mail).Scan(&user).Error)

	tokenID, err := auth.GenerateTokenID()
	require.NoError(t, err)
	require.NoError(t, h.db.Create(&models.APIToken{
		TokenID: tokenID, UserID: user.ID, RoleID: user.RoleID, TTL: 3600, Status: models.TokenStatusActive,
	}).Error)

	token, err := auth.MakeAPIToken("test", auth.MakeAPITokenClaims(tokenID, user.Hash, user.ID, user.RoleID, 3600))
	require.NoError(t, err)

	return token
}

func (h *sessionHarness) generation(t *testing.T) uint64 {
	t.Helper()

	var row struct{ SessionGeneration uint64 }
	require.NoError(t, h.db.Raw("SELECT session_generation FROM users WHERE mail = ?", h.mail).Scan(&row).Error)

	return row.SessionGeneration
}
