package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

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

// usersCall runs one user handler for the caller the way the router hands it a request.
func usersCall(handler gin.HandlerFunc, uid, rid uint64, privs []string, body string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("uid", uid)
	c.Set("rid", rid)
	c.Set("prm", privs)
	c.Request = httptest.NewRequest(http.MethodPost, "/users/", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler(c)

	return w
}

func TestUsers_CreateUser_StoresTheUserWithItsPreferencesOrNothing(t *testing.T) {
	const newUser = `{"mail":"newuser@test.com","name":"New User","role_id":2,"status":"active","type":"local",` +
		`"password":"SecurePass123!"}`
	create := []string{"users.create"}

	for _, tc := range []struct {
		name      string
		prepare   func(db *gorm.DB)
		privs     []string
		body      string
		mail      string
		wantCode  int
		wantUsers int
		wantName  string
	}{
		{
			name:      "a new user",
			privs:     create,
			body:      newUser,
			mail:      "newuser@test.com",
			wantCode:  http.StatusCreated,
			wantUsers: 1,
			wantName:  "New User",
		},
		{
			name:  "a body that names a session generation",
			privs: create,
			body: `{"mail":"pinned@test.com","name":"Pinned","role_id":2,"status":"active","type":"local",` +
				`"password":"SecurePass123!","session_generation":9223372036854775807}`,
			mail:      "pinned@test.com",
			wantCode:  http.StatusCreated,
			wantUsers: 1,
			wantName:  "Pinned",
		},
		{
			name:     "a preferences insert that fails rolls the user back",
			prepare:  func(db *gorm.DB) { db.Exec("DROP TABLE user_preferences") },
			privs:    create,
			body:     newUser,
			mail:     "newuser@test.com",
			wantCode: http.StatusInternalServerError,
		},
		{
			name:     "a caller without users.create",
			privs:    []string{"flows.view"},
			body:     newUser,
			mail:     "newuser@test.com",
			wantCode: http.StatusForbidden,
		},
		{
			name:     "a malformed body",
			privs:    create,
			body:     "{invalid json",
			wantCode: http.StatusBadRequest,
		},
		{
			name:  "an email another user holds",
			privs: create,
			body: `{"mail":"user1@test.com","name":"Second User","role_id":2,"status":"active","type":"local",` +
				`"password":"AnotherPass456!"}`,
			mail:      "user1@test.com",
			wantCode:  http.StatusInternalServerError,
			wantUsers: 1,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := setupTestDB(t)
			defer db.Close()
			if tc.prepare != nil {
				tc.prepare(db)
			}
			service := NewUserService(db, auth.NewUserCache(db), "/api/v1", 4*60*60)

			w := usersCall(service.CreateUser, 1, 1, tc.privs, tc.body)

			require.Equal(t, tc.wantCode, w.Code, w.Body.String())
			var count int
			require.NoError(t, db.Model(&models.User{}).Where("mail = ?", tc.mail).Count(&count).Error)
			assert.Equal(t, tc.wantUsers, count, "a refused create leaves no row, and a taken email keeps one")
			if tc.wantCode != http.StatusCreated {
				return
			}

			var created models.User
			require.NoError(t, db.Where("mail = ?", tc.mail).First(&created).Error)
			assert.Equal(t, tc.wantName, created.Name)
			assert.Equal(t, uint64(2), created.RoleID)
			assert.Equal(t, uint64(1), created.SessionGeneration,
				"a generation at the top of bigint can never be moved, so the account's sessions could never be revoked")
			assert.NotContains(t, w.Body.String(), "session_generation", "the generation is not part of the API")

			var prefs models.UserPreferences
			require.NoError(t, db.Where("user_id = ?", created.ID).First(&prefs).Error,
				"the preferences row is created with the user")
			assert.NotNil(t, prefs.Preferences.FavoriteFlows)
			assert.Empty(t, prefs.Preferences.FavoriteFlows)
		})
	}
}

func TestUsers_ChangeEmailCurrentUser_ChangesOnlyToAFreeValidAddressUnderTheRightPassword(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("SecurePass123!"), bcrypt.DefaultCost)
	require.NoError(t, err)
	require.NoError(t, db.Model(&models.User{}).Where("id = 1").Updates(map[string]any{
		"password": string(hashedPassword),
		"hash":     "11111111111111111111111111111111",
	}).Error)

	service := NewUserService(db, auth.NewUserCache(db), "/api/v1", 4*60*60)
	github := "github"

	for _, tc := range []struct {
		name          string
		requestBody   string
		expectedCode  int
		errorContains string
		wantMail      string
		wantProvider  *string
	}{
		{
			name:         "a free address under the right password",
			requestBody:  `{"current_password": "SecurePass123!", "mail": "newemail@test.com"}`,
			expectedCode: http.StatusOK,
			wantMail:     "newemail@test.com",
		},
		{
			name:         "a mixed-case address is stored as entered, as login and OAuth read it",
			requestBody:  `{"current_password": "SecurePass123!", "mail": "Mixed.Case@Example.com"}`,
			expectedCode: http.StatusOK,
			wantMail:     "Mixed.Case@Example.com",
		},
		{
			name:          "a wrong current password",
			requestBody:   `{"current_password": "WrongPassword!", "mail": "another@test.com"}`,
			expectedCode:  http.StatusForbidden,
			errorContains: "invalid current password",
			wantMail:      "user1@test.com",
			wantProvider:  &github,
		},
		{
			name:          "an address another user holds",
			requestBody:   `{"current_password": "SecurePass123!", "mail": "user2@test.com"}`,
			expectedCode:  http.StatusConflict,
			errorContains: "email already exists",
			wantMail:      "user1@test.com",
			wantProvider:  &github,
		},
		{
			name:          "an address that is not an email",
			requestBody:   `{"current_password": "SecurePass123!", "mail": "invalid-email"}`,
			expectedCode:  http.StatusBadRequest,
			errorContains: "failed to validate user email",
			wantMail:      "user1@test.com",
			wantProvider:  &github,
		},
		{
			name:          "a bare uuid, which only the system may use as a mail",
			requestBody:   `{"current_password": "SecurePass123!", "mail": "550e8400-e29b-41d4-a716-446655440000"}`,
			expectedCode:  http.StatusBadRequest,
			errorContains: "failed to validate user email",
			wantMail:      "user1@test.com",
			wantProvider:  &github,
		},
		{
			name:          "the admin sentinel",
			requestBody:   `{"current_password": "SecurePass123!", "mail": "admin"}`,
			expectedCode:  http.StatusBadRequest,
			errorContains: "failed to validate user email",
			wantMail:      "user1@test.com",
			wantProvider:  &github,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, db.Model(&models.User{}).Where("id = 1").Updates(map[string]any{
				"mail":     "user1@test.com",
				"provider": github,
			}).Error)

			w := usersCall(service.ChangeEmailCurrentUser, 1, 2, []string{}, tc.requestBody)

			assert.Equal(t, tc.expectedCode, w.Code)
			assert.Contains(t, w.Body.String(), tc.errorContains)

			var user models.User
			require.NoError(t, db.Where("id = 1").First(&user).Error)
			assert.Equal(t, tc.wantMail, user.Mail)
			assert.Equal(t, tc.wantProvider, user.Provider, "only a completed change clears the now-stale OAuth provider link")
		})
	}
}

func TestUsers_ChangeNameCurrentUser_RenamesAnExistingUserToANonEmptyName(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewUserService(db, auth.NewUserCache(db), "/api/v1", 4*60*60)

	for _, tc := range []struct {
		name          string
		requestBody   string
		uid           uint64
		expectedCode  int
		checkResult   func(t *testing.T)
		errorContains string
	}{
		{
			name:         "a new name for an existing user",
			requestBody:  `{"name": "Renamed User"}`,
			uid:          1,
			expectedCode: http.StatusOK,
			checkResult: func(t *testing.T) {
				var user models.User
				require.NoError(t, db.Where("id = 1").First(&user).Error)
				assert.Equal(t, "Renamed User", user.Name)
			},
		},
		{
			name:          "a user that does not exist",
			requestBody:   `{"name": "Ghost"}`,
			uid:           999,
			expectedCode:  http.StatusNotFound,
			errorContains: "Users.NotFound",
			checkResult: func(t *testing.T) {
				var count int
				require.NoError(t, db.Model(&models.User{}).Where("name = ?", "Ghost").Count(&count).Error)
				assert.Equal(t, 0, count, "no row is created or touched for a missing user")
			},
		},
		{
			name:          "an empty name",
			requestBody:   `{"name": ""}`,
			uid:           1,
			expectedCode:  http.StatusBadRequest,
			errorContains: "failed to validate user name",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := usersCall(service.ChangeNameCurrentUser, tc.uid, 2, []string{}, tc.requestBody)

			assert.Equal(t, tc.expectedCode, w.Code)
			if tc.checkResult != nil {
				tc.checkResult(t)
			}
			assert.Contains(t, w.Body.String(), tc.errorContains)
		})
	}
}

const nextAccountPassword = "NextAccountPass2!"

type passwordHarness struct {
	engine  *gin.Engine
	db      *gorm.DB
	refuse  *bool
	cookies []*http.Cookie
	hash    string
}

func newPasswordHarness(t *testing.T, roleID int) *passwordHarness {
	t.Helper()

	db := setupTestDB(t)
	t.Cleanup(func() { db.Close() })
	seedLocalAccount(t, db, "known@corp.com", roleID)

	refuse := false
	gin.SetMode(gin.TestMode)
	userCache := auth.NewUserCache(db)
	authSvc := NewAuthService(AuthServiceConfig{BaseURL: "/", SessionTimeout: 3600}, db, map[string]oauth.OAuthClient{}, userCache)
	userSvc := NewUserService(db, userCache, "/", 3600)
	signedIn := auth.NewAuthMiddleware("/", "test", auth.NewTokenCache(db), userCache).AuthUserRequired

	engine := gin.New()
	require.NoError(t, engine.SetTrustedProxies(nil))
	engine.Use(sessions.Sessions("pentagi", cookie.NewStore([]byte("test-secret"))))
	overfill := func(c *gin.Context) {
		if refuse {
			// Past securecookie's 4096-byte limit, so the handler's Save fails to encode.
			sessions.Default(c).Set("pad", strings.Repeat("x", 5000))
		}
	}
	engine.POST("/auth/login", authSvc.AuthLogin)
	engine.PUT("/user/password", signedIn, overfill, userSvc.ChangePasswordCurrentUser)
	engine.PUT("/users/:hash", signedIn, overfill, userSvc.PatchUser)
	engine.GET("/private", signedIn, func(c *gin.Context) { c.Status(http.StatusOK) })
	engine.GET("/sgn", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"sgn": sessions.Default(c).Get("sgn")}) })

	h := &passwordHarness{engine: engine, db: db, refuse: &refuse}
	h.cookies = h.login(t, "known@corp.com")

	var row struct{ Hash string }
	require.NoError(t, db.Raw("SELECT hash FROM users WHERE mail = ?", "known@corp.com").Scan(&row).Error)
	h.hash = row.Hash

	return h
}

func (h *passwordHarness) login(t *testing.T, mail string) []*http.Cookie {
	t.Helper()

	body, err := json.Marshal(map[string]string{"mail": mail, "password": knownAccountPassword})
	require.NoError(t, err)

	h.cookies = nil
	rec := h.do(t, http.MethodPost, "/auth/login", string(body))
	require.Equal(t, http.StatusOK, rec.Code)

	return rec.Result().Cookies()
}

func (h *passwordHarness) privateStatus(t *testing.T, cookies []*http.Cookie) int {
	t.Helper()

	saved := h.cookies
	defer func() { h.cookies = saved }()
	h.cookies = cookies

	return h.do(t, http.MethodGet, "/private", "").Code
}

func (h *passwordHarness) do(t *testing.T, method, target, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "10.0.0.1:54321"
	for _, c := range h.cookies {
		req.AddCookie(c)
	}

	rec := httptest.NewRecorder()
	h.engine.ServeHTTP(rec, req)

	return rec
}

func (h *passwordHarness) state(t *testing.T) (generation uint64, password string) {
	t.Helper()

	var row struct {
		SessionGeneration uint64
		Password          string
	}
	require.NoError(t, h.db.Raw("SELECT session_generation, password FROM users WHERE mail = ?", "known@corp.com").
		Scan(&row).Error)

	return row.SessionGeneration, row.Password
}

func (h *passwordHarness) cookieGeneration(t *testing.T, fresh []*http.Cookie) uint64 {
	t.Helper()

	byName := map[string]*http.Cookie{}
	for _, c := range append(h.cookies, fresh...) {
		byName[c.Name] = c
	}
	h.cookies = h.cookies[:0]
	for _, c := range byName {
		h.cookies = append(h.cookies, c)
	}
	rec := h.do(t, http.MethodGet, "/sgn", "")
	require.Equal(t, http.StatusOK, rec.Code)

	var body struct{ Sgn uint64 }
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))

	return body.Sgn
}

func (h *passwordHarness) changeThroughTheForm(t *testing.T) *httptest.ResponseRecorder {
	return h.do(t, http.MethodPut, "/user/password",
		`{"current_password":"`+knownAccountPassword+`","password":"`+nextAccountPassword+
			`","confirm_password":"`+nextAccountPassword+`"}`)
}

func (h *passwordHarness) changeThroughTheEditor(t *testing.T) *httptest.ResponseRecorder {
	return h.do(t, http.MethodPut, "/users/"+h.hash,
		`{"hash":"`+h.hash+`","type":"local","mail":"known@corp.com","name":"Known","status":"active",`+
			`"role_id":2,"password":"`+nextAccountPassword+`"}`)
}

// usersPasswordDoors are the handlers that reach changePassword for the account's own password.
var usersPasswordDoors = []struct {
	name   string
	role   int
	change func(*passwordHarness, *testing.T) *httptest.ResponseRecorder
}{
	{"a user through the password form", 2, (*passwordHarness).changeThroughTheForm},
	{"an administrator through the user editor", 1, (*passwordHarness).changeThroughTheEditor},
}

func TestUsers_ChangePassword_EndsOtherSessionsAndKeepsTheCaller(t *testing.T) {
	for _, door := range usersPasswordDoors {
		t.Run(door.name, func(t *testing.T) {
			h := newPasswordHarness(t, door.role)
			before, _ := h.state(t)
			stolen := slices.Clone(h.cookies)

			rec := door.change(h, t)
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

			after, _ := h.state(t)
			assert.Equal(t, before+1, after, "the other sessions have to end")
			assert.Equal(t, after, h.cookieGeneration(t, rec.Result().Cookies()),
				"the caller's own cookie has to carry the new generation, or the next request is refused")
			assert.NotEqual(t, http.StatusOK, h.privateStatus(t, stolen), "a cookie from before the change is refused at once")
			assert.Equal(t, http.StatusOK, h.privateStatus(t, h.cookies), "the caller stays signed in")
		})
	}
}

func TestUsers_ChangePassword_ThroughTheEditorTakesUsersEdit(t *testing.T) {
	t.Run("an administrator resetting another account stays signed in", func(t *testing.T) {
		h := newPasswordHarness(t, 2)
		seedLocalAccount(t, h.db, "boss.admin@corp.com", 1)
		before, _ := h.state(t)

		h.cookies = h.login(t, "boss.admin@corp.com")
		rec := h.changeThroughTheEditor(t)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

		after, _ := h.state(t)
		assert.Equal(t, before+1, after, "the user's sessions end")
		h.cookieGeneration(t, rec.Result().Cookies())
		assert.Equal(t, http.StatusOK, h.privateStatus(t, h.cookies),
			"the administrator's cookie is not stamped with the other account's generation")
	})

	t.Run("a user cannot set their own password without the current one", func(t *testing.T) {
		h := newPasswordHarness(t, 2)
		beforeGeneration, beforePassword := h.state(t)

		rec := h.changeThroughTheEditor(t)

		assert.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
		assert.Contains(t, rec.Body.String(), "NotPermitted", "the permission gate refuses, not the session check")
		afterGeneration, afterPassword := h.state(t)
		assert.Equal(t, beforeGeneration, afterGeneration)
		assert.Equal(t, beforePassword, afterPassword)
	})
}

func TestUsers_ChangePassword_ChangesNothingWhenAStepFails(t *testing.T) {
	for _, fault := range []struct {
		name   string
		inject func(t *testing.T, h *passwordHarness)
	}{
		{
			name:   "the caller's cookie cannot be re-stamped",
			inject: func(_ *testing.T, h *passwordHarness) { *h.refuse = true },
		},
		{
			name: "the other sessions cannot be ended",
			inject: func(t *testing.T, h *passwordHarness) {
				require.NoError(t, h.db.Exec("CREATE TRIGGER refuse_generation BEFORE UPDATE OF session_generation ON users "+
					"BEGIN SELECT RAISE(FAIL, 'generation locked'); END").Error)
			},
		},
		{
			name: "the commit is refused",
			inject: func(t *testing.T, h *passwordHarness) {
				for _, stmt := range []string{
					"PRAGMA foreign_keys = ON",
					"CREATE TABLE generation_parent (id INTEGER PRIMARY KEY)",
					"CREATE TABLE generation_guard (generation INTEGER REFERENCES generation_parent(id) DEFERRABLE INITIALLY DEFERRED)",
					"CREATE TRIGGER fail_at_commit AFTER UPDATE OF session_generation ON users " +
						"BEGIN INSERT INTO generation_guard (generation) VALUES (NEW.session_generation); END",
				} {
					require.NoError(t, h.db.Exec(stmt).Error, stmt)
				}
			},
		},
	} {
		for _, door := range usersPasswordDoors {
			t.Run(fault.name+", "+door.name, func(t *testing.T) {
				h := newPasswordHarness(t, door.role)
				beforeGeneration, beforePassword := h.state(t)
				fault.inject(t, h)

				rec := door.change(h, t)

				assert.Equal(t, http.StatusInternalServerError, rec.Code)
				afterGeneration, afterPassword := h.state(t)
				assert.Equal(t, beforeGeneration, afterGeneration, "the other sessions are left as they were")
				assert.Equal(t, beforePassword, afterPassword, "a new password without ended sessions leaves a stolen cookie working")
				assert.Equal(t, beforeGeneration, h.cookieGeneration(t, rec.Result().Cookies()),
					"a cookie ahead of the row is refused on the next request")
			})
		}
	}
}
