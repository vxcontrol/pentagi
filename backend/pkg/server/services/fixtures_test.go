package services

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"pentagi/pkg/resources"
	"pentagi/pkg/server/models"
	"pentagi/pkg/server/oauth"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	// Create roles table
	db.Exec(`
		CREATE TABLE roles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE
		)
	`)

	// Create privileges table
	db.Exec(`
		CREATE TABLE privileges (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			role_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			UNIQUE(role_id, name)
		)
	`)

	// Create api_tokens table
	db.Exec(`
		CREATE TABLE api_tokens (
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
		)
	`)

	// Insert test roles
	db.Exec("INSERT INTO roles (id, name) VALUES (1, 'Admin'), (2, 'User')")

	// Insert test privileges for Admin role
	db.Exec(`INSERT INTO privileges (role_id, name) VALUES
		(1, 'users.create'),
		(1, 'users.delete'),
		(1, 'users.edit'),
		(1, 'users.view'),
		(1, 'roles.view'),
		(1, 'flows.admin'),
		(1, 'flows.create'),
		(1, 'flows.delete'),
		(1, 'flows.edit'),
		(1, 'flows.view'),
		(1, 'settings.tokens.create'),
		(1, 'settings.tokens.view'),
		(1, 'settings.tokens.edit'),
		(1, 'settings.tokens.delete'),
		(1, 'settings.tokens.admin')`)

	// Insert test privileges for User role
	db.Exec(`INSERT INTO privileges (role_id, name) VALUES
		(2, 'roles.view'),
		(2, 'flows.create'),
		(2, 'flows.delete'),
		(2, 'flows.edit'),
		(2, 'flows.view'),
		(2, 'settings.tokens.create'),
		(2, 'settings.tokens.view'),
		(2, 'settings.tokens.edit'),
		(2, 'settings.tokens.delete')`)

	// Create users table
	db.Exec(`
		CREATE TABLE users (
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
		)
	`)

	// Insert test users
	db.Exec("INSERT INTO users (id, hash, mail, name, status, role_id) VALUES (1, 'testhash1', 'user1@test.com', 'User 1', 'active', 2)")
	db.Exec("INSERT INTO users (id, hash, mail, name, status, role_id) VALUES (2, 'testhash2', 'user2@test.com', 'User 2', 'active', 2)")

	// Create user_preferences table
	db.Exec(`
		CREATE TABLE user_preferences (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL UNIQUE,
			preferences TEXT NOT NULL DEFAULT '{"favoriteFlows": []}',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)
	`)

	// Insert preferences for test users
	db.Exec("INSERT INTO user_preferences (user_id, preferences) VALUES (1, '{\"favoriteFlows\": []}')")
	db.Exec("INSERT INTO user_preferences (user_id, preferences) VALUES (2, '{\"favoriteFlows\": []}')")

	return db
}

const knownAccountPassword = "KnownAccountPass1!"

func seedLocalUser(t *testing.T, db *gorm.DB, mail string) {
	t.Helper()

	seedLocalAccount(t, db, mail, 2)
}

func seedLocalAccount(t *testing.T, db *gorm.DB, mail string, roleID int) {
	t.Helper()

	hash, err := bcrypt.GenerateFromPassword([]byte(knownAccountPassword), bcrypt.DefaultCost)
	require.NoError(t, err)

	require.NoError(t, db.Exec(
		"INSERT INTO users (hash, type, mail, name, status, role_id, password) VALUES (?, 'local', ?, 'Known', 'active', ?, ?)",
		fmt.Sprintf("%032x", len(mail)), mail, roleID, string(hash),
	).Error)
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

func setupFlowFileServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	db.LogMode(false)

	require.NoError(t, db.Exec(`
		CREATE TABLE flows (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			status TEXT NOT NULL DEFAULT 'created',
			title TEXT NOT NULL DEFAULT 'untitled',
			model TEXT NOT NULL DEFAULT '',
			model_provider_name TEXT NOT NULL DEFAULT '',
			model_provider_type TEXT NOT NULL DEFAULT 'openai',
			language TEXT NOT NULL DEFAULT 'english',
			functions TEXT NOT NULL DEFAULT '{}',
			tool_call_id_template TEXT NOT NULL DEFAULT '',
			trace_id TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME
		)
	`).Error)

	require.NoError(t, db.Exec(`
		CREATE TABLE user_resources (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			hash TEXT NOT NULL DEFAULT '',
			name TEXT NOT NULL,
			path TEXT NOT NULL,
			size INTEGER NOT NULL DEFAULT 0,
			is_dir BOOLEAN NOT NULL DEFAULT FALSE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, path)
		)
	`).Error)

	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	return db
}

func seedFlow(t *testing.T, db *gorm.DB, id, userID uint64) {
	t.Helper()

	require.NoError(t, db.Exec(
		`INSERT INTO flows (id, user_id, model, model_provider_name, tool_call_id_template, trace_id) VALUES (?, ?, 'gpt', 'openai', 'tcid', '')`,
		id, userID,
	).Error)
}

// newFlowFileTestContext creates a gin test context with the path param,
// uid, prm and request body pre-populated.
func newFlowFileTestContext(
	method, target string,
	body io.Reader,
	privs []string,
	uid, flowID uint64,
) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("uid", uid)
	c.Set("prm", privs)
	c.Params = gin.Params{
		{Key: "flowID", Value: strconv.FormatUint(flowID, 10)},
	}
	c.Request = httptest.NewRequest(method, target, body)
	return c, w
}

func decodeResourceListResponse(t *testing.T, w *httptest.ResponseRecorder) models.ResourceList {
	t.Helper()

	var resp struct {
		Status string              `json:"status"`
		Data   models.ResourceList `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, "success", resp.Status)
	return resp.Data
}

func seedResource(t *testing.T, db *gorm.DB, rec models.UserResource) models.UserResource {
	t.Helper()

	require.NoError(t, db.Create(&rec).Error)
	return rec
}

type uploadTestFile struct {
	name    string
	content string
}

func multipartUploadBodyWithField(
	t *testing.T,
	files []uploadTestFile,
	fieldName string,
) (*bytes.Buffer, string) {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for _, file := range files {
		part, err := writer.CreateFormFile(fieldName, file.name)
		require.NoError(t, err)
		_, err = part.Write([]byte(file.content))
		require.NoError(t, err)
	}
	require.NoError(t, writer.Close())
	return &body, writer.FormDataContentType()
}

// rawMultipartUpload writes the file name into the header unescaped, as a
// hand-rolled client would; mime/multipart percent-encodes CR and LF.
func rawMultipartUpload(fieldName, fileName, content string) (*bytes.Buffer, string) {
	const boundary = "PentagiRawBoundary"

	var body bytes.Buffer
	fmt.Fprintf(&body, "--%s\r\n", boundary)
	fmt.Fprintf(&body, "Content-Disposition: form-data; name=%q; filename=\"%s\"\r\n", fieldName, fileName)
	fmt.Fprint(&body, "Content-Type: application/octet-stream\r\n\r\n")
	fmt.Fprintf(&body, "%s\r\n", content)
	fmt.Fprintf(&body, "--%s--\r\n", boundary)

	return &body, "multipart/form-data; boundary=" + boundary
}

func writeResourceBlob(t *testing.T, dataDir, hash, content string) {
	t.Helper()

	blobPath := resources.BlobPath(dataDir, hash)
	require.NoError(t, os.MkdirAll(filepath.Dir(blobPath), 0755))
	require.NoError(t, os.WriteFile(blobPath, []byte(content), 0644))
}

func md5HexForService(content string) string {
	sum := md5.Sum([]byte(content))
	return hex.EncodeToString(sum[:])
}

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

func newOAuthServiceVerified(db *gorm.DB, email string, verified bool) *AuthService {
	return &AuthService{
		cfg:   AuthServiceConfig{BaseURL: "/", SessionTimeout: 3600},
		db:    db,
		key:   []byte("0123456789abcdef0123456789abcdef"),
		oauth: map[string]oauth.OAuthClient{"github": &fakeOAuthClient{name: "github", email: email, verified: verified}},
	}
}

// authOAuthCallback brings the state AuthAuthorize signed back to AuthLoginGetCallback, as the provider's redirect does.
func authOAuthCallback(t *testing.T, svc *AuthService, header http.Header) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()

	authorized := httptest.NewRecorder()
	authorize, _ := gin.CreateTestContext(authorized)
	authorize.Request = httptest.NewRequest(http.MethodGet, "/auth/authorize?provider=github", nil)
	svc.AuthAuthorize(authorize)
	require.Equal(t, http.StatusTemporaryRedirect, authorized.Code, authorized.Body.String())

	var state string
	for _, c := range authorized.Result().Cookies() {
		if c.Name == authStateCookieName {
			state = c.Value
		}
	}
	require.NotEmpty(t, state, "the authorize step set no state cookie")

	c, w := newCallbackContext(t)
	c.Request.URL.RawQuery = url.Values{"code": {"test-code"}, "state": {state}}.Encode()
	c.Request.AddCookie(&http.Cookie{Name: authStateCookieName, Value: state})
	for name, values := range header {
		c.Request.Header[name] = values
	}

	svc.AuthLoginGetCallback(c)

	return c, w
}
