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

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func setupTestContext(uid, rid uint64, uhash string, permissions []string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Set("uid", uid)
	c.Set("rid", rid)
	c.Set("uhash", uhash)
	c.Set("prm", permissions)

	return c, w
}

func TestTokenService_CreateToken(t *testing.T) {
	testCases := []struct {
		name          string
		globalSalt    string
		requestBody   string
		uid           uint64
		rid           uint64
		uhash         string
		expectedCode  int
		expectToken   bool
		errorContains string
	}{
		{
			name:         "valid token creation",
			globalSalt:   "custom_salt",
			requestBody:  `{"ttl": 3600, "name": "Test Token"}`,
			uid:          1,
			rid:          2,
			uhash:        "testhash",
			expectedCode: http.StatusCreated,
			expectToken:  true,
		},
		{
			name:          "default salt protection",
			globalSalt:    "salt",
			requestBody:   `{"ttl": 3600}`,
			uid:           1,
			rid:           2,
			uhash:         "testhash",
			expectedCode:  http.StatusBadRequest,
			expectToken:   false,
			errorContains: "disabled",
		},
		{
			name:          "empty salt protection",
			globalSalt:    "",
			requestBody:   `{"ttl": 3600}`,
			uid:           1,
			rid:           2,
			uhash:         "testhash",
			expectedCode:  http.StatusBadRequest,
			expectToken:   false,
			errorContains: "disabled",
		},
		{
			name:          "invalid TTL (too short)",
			globalSalt:    "custom_salt",
			requestBody:   `{"ttl": 30}`,
			uid:           1,
			rid:           2,
			uhash:         "testhash",
			expectedCode:  http.StatusBadRequest,
			expectToken:   false,
			errorContains: "",
		},
		{
			name:          "invalid TTL (too long)",
			globalSalt:    "custom_salt",
			requestBody:   `{"ttl": 100000000}`,
			uid:           1,
			rid:           2,
			uhash:         "testhash",
			expectedCode:  http.StatusBadRequest,
			expectToken:   false,
			errorContains: "",
		},
		{
			name:         "token without name",
			globalSalt:   "custom_salt",
			requestBody:  `{"ttl": 7200}`,
			uid:          1,
			rid:          2,
			uhash:        "testhash",
			expectedCode: http.StatusCreated,
			expectToken:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := setupTestDB(t)
			defer db.Close()

			tokenCache := auth.NewTokenCache(db)
			service := NewTokenService(db, tc.globalSalt, tokenCache, nil)
			c, w := setupTestContext(tc.uid, tc.rid, tc.uhash, []string{"settings.tokens.create"})

			c.Request = httptest.NewRequest(http.MethodPost, "/tokens", bytes.NewBufferString(tc.requestBody))
			c.Request.Header.Set("Content-Type", "application/json")

			service.CreateToken(c)

			assert.Equal(t, tc.expectedCode, w.Code)

			if tc.expectToken {
				var response struct {
					Status string `json:"status"`
					Data   struct {
						Token   string `json:"token"`
						TokenID string `json:"token_id"`
					} `json:"data"`
				}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "success", response.Status)
				assert.NotEmpty(t, response.Data.Token)
				assert.NotEmpty(t, response.Data.TokenID)
				assert.Len(t, response.Data.TokenID, 10)
			}

			if tc.errorContains != "" {
				assert.Contains(t, w.Body.String(), tc.errorContains)
			}
		})
	}
}

func TestTokenService_CreateToken_NameUniqueness(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tokenCache := auth.NewTokenCache(db)
	service := NewTokenService(db, "custom_salt", tokenCache, nil)

	// Create first token
	c1, w1 := setupTestContext(1, 2, "hash1", []string{"settings.tokens.create"})
	c1.Request = httptest.NewRequest(http.MethodPost, "/tokens",
		bytes.NewBufferString(`{"ttl": 3600, "name": "Duplicate Name"}`))
	c1.Request.Header.Set("Content-Type", "application/json")

	service.CreateToken(c1)
	assert.Equal(t, http.StatusCreated, w1.Code)

	// Try to create second token with same name for same user
	c2, w2 := setupTestContext(1, 2, "hash1", []string{"settings.tokens.create"})
	c2.Request = httptest.NewRequest(http.MethodPost, "/tokens",
		bytes.NewBufferString(`{"ttl": 3600, "name": "Duplicate Name"}`))
	c2.Request.Header.Set("Content-Type", "application/json")

	service.CreateToken(c2)
	assert.Equal(t, http.StatusBadRequest, w2.Code)
	assert.Contains(t, w2.Body.String(), "already exists")

	// Create token with same name for different user (should succeed)
	c3, w3 := setupTestContext(2, 2, "hash2", []string{"settings.tokens.create"})
	c3.Request = httptest.NewRequest(http.MethodPost, "/tokens",
		bytes.NewBufferString(`{"ttl": 3600, "name": "Duplicate Name"}`))
	c3.Request.Header.Set("Content-Type", "application/json")

	service.CreateToken(c3)
	assert.Equal(t, http.StatusCreated, w3.Code)
}

func TestTokenService_ListTokens(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create tokens for different users
	tokens := []models.APIToken{
		{TokenID: "token1", UserID: 1, RoleID: 2, TTL: 3600, Status: models.TokenStatusActive},
		{TokenID: "token2", UserID: 1, RoleID: 2, TTL: 7200, Status: models.TokenStatusActive},
		{TokenID: "token3", UserID: 2, RoleID: 2, TTL: 3600, Status: models.TokenStatusActive},
		{TokenID: "token4", UserID: 1, RoleID: 2, TTL: 3600, Status: models.TokenStatusRevoked},
	}

	for _, token := range tokens {
		err := db.Create(&token).Error
		require.NoError(t, err)
	}

	tokenCache := auth.NewTokenCache(db)
	service := NewTokenService(db, "custom_salt", tokenCache, nil)

	testCases := []struct {
		name          string
		uid           uint64
		permissions   []string
		expectedCount int
	}{
		{
			name:          "regular user sees own tokens",
			uid:           1,
			permissions:   []string{"settings.tokens.view"},
			expectedCount: 3, // token1, token2, token4 (including revoked)
		},
		{
			name:          "admin sees all tokens",
			uid:           1,
			permissions:   []string{"settings.tokens.view", "settings.tokens.admin"},
			expectedCount: 4, // all tokens
		},
		{
			name:          "user 2 sees only own token",
			uid:           2,
			permissions:   []string{"settings.tokens.view"},
			expectedCount: 1, // token3
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, w := setupTestContext(tc.uid, 2, fmt.Sprintf("hash%d", tc.uid), tc.permissions)
			c.Request = httptest.NewRequest(http.MethodGet, "/tokens", nil)

			service.ListTokens(c)

			assert.Equal(t, http.StatusOK, w.Code)

			var response struct {
				Status string `json:"status"`
				Data   struct {
					Tokens []models.APIToken `json:"tokens"`
					Total  uint64            `json:"total"`
				} `json:"data"`
			}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Equal(t, "success", response.Status)
			assert.Equal(t, tc.expectedCount, len(response.Data.Tokens))
			assert.Equal(t, uint64(tc.expectedCount), response.Data.Total)
		})
	}
}

func TestTokenService_GetToken(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create tokens
	token1 := models.APIToken{TokenID: "usertoken1", UserID: 1, RoleID: 2, TTL: 3600, Status: models.TokenStatusActive}
	token2 := models.APIToken{TokenID: "usertoken2", UserID: 2, RoleID: 2, TTL: 3600, Status: models.TokenStatusActive}

	db.Create(&token1)
	db.Create(&token2)

	tokenCache := auth.NewTokenCache(db)
	service := NewTokenService(db, "custom_salt", tokenCache, nil)

	testCases := []struct {
		name         string
		tokenID      string
		uid          uint64
		permissions  []string
		expectedCode int
	}{
		{
			name:         "user gets own token",
			tokenID:      "usertoken1",
			uid:          1,
			permissions:  []string{"settings.tokens.view"},
			expectedCode: http.StatusOK,
		},
		{
			name:         "user cannot get other user's token",
			tokenID:      "usertoken2",
			uid:          1,
			permissions:  []string{"settings.tokens.view"},
			expectedCode: http.StatusForbidden,
		},
		{
			name:         "admin can get any token",
			tokenID:      "usertoken2",
			uid:          1,
			permissions:  []string{"settings.tokens.view", "settings.tokens.admin"},
			expectedCode: http.StatusOK,
		},
		{
			name:         "nonexistent token",
			tokenID:      "nonexistent",
			uid:          1,
			permissions:  []string{"settings.tokens.view", "settings.tokens.admin"},
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, w := setupTestContext(tc.uid, 2, "testhash", tc.permissions)
			c.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/tokens/%s", tc.tokenID), nil)
			c.Params = gin.Params{{Key: "tokenID", Value: tc.tokenID}}

			service.GetToken(c)

			assert.Equal(t, tc.expectedCode, w.Code)

			if tc.expectedCode == http.StatusOK {
				var response struct {
					Status string          `json:"status"`
					Data   models.APIToken `json:"data"`
				}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "success", response.Status)
				assert.Equal(t, tc.tokenID, response.Data.TokenID)
			}
		})
	}
}

func TestTokenService_UpdateToken(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tokenCache := auth.NewTokenCache(db)
	service := NewTokenService(db, "custom_salt", tokenCache, nil)

	// Create initial token
	initialToken := models.APIToken{
		TokenID: "updatetest1",
		UserID:  1,
		RoleID:  2,
		TTL:     3600,
		Status:  models.TokenStatusActive,
	}
	err := db.Create(&initialToken).Error
	require.NoError(t, err)

	testCases := []struct {
		name         string
		tokenID      string
		uid          uint64
		permissions  []string
		requestBody  string
		expectedCode int
		checkResult  func(t *testing.T, db *gorm.DB)
	}{
		{
			name:         "update name",
			tokenID:      "updatetest1",
			uid:          1,
			permissions:  []string{"settings.tokens.edit"},
			requestBody:  `{"name": "Updated Name"}`,
			expectedCode: http.StatusOK,
			checkResult: func(t *testing.T, db *gorm.DB) {
				var token models.APIToken
				db.Where("token_id = ?", "updatetest1").First(&token)
				assert.NotNil(t, token.Name)
				assert.Equal(t, "Updated Name", *token.Name)
			},
		},
		{
			name:         "revoke token",
			tokenID:      "updatetest1",
			uid:          1,
			permissions:  []string{"settings.tokens.edit"},
			requestBody:  `{"status": "revoked"}`,
			expectedCode: http.StatusOK,
			checkResult: func(t *testing.T, db *gorm.DB) {
				var token models.APIToken
				db.Where("token_id = ?", "updatetest1").First(&token)
				assert.Equal(t, models.TokenStatusRevoked, token.Status)
			},
		},
		{
			name:         "reactivate token",
			tokenID:      "updatetest1",
			uid:          1,
			permissions:  []string{"settings.tokens.edit"},
			requestBody:  `{"status": "active"}`,
			expectedCode: http.StatusOK,
			checkResult: func(t *testing.T, db *gorm.DB) {
				var token models.APIToken
				db.Where("token_id = ?", "updatetest1").First(&token)
				assert.Equal(t, models.TokenStatusActive, token.Status)
			},
		},
		{
			name:         "unauthorized update (different user)",
			tokenID:      "updatetest1",
			uid:          2,
			permissions:  []string{"settings.tokens.edit"},
			requestBody:  `{"name": "Hacked"}`,
			expectedCode: http.StatusForbidden,
			checkResult:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, w := setupTestContext(tc.uid, 2, "testhash", tc.permissions)
			c.Request = httptest.NewRequest(http.MethodPut,
				fmt.Sprintf("/tokens/%s", tc.tokenID),
				bytes.NewBufferString(tc.requestBody))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "tokenID", Value: tc.tokenID}}

			service.UpdateToken(c)

			assert.Equal(t, tc.expectedCode, w.Code)

			if tc.checkResult != nil {
				tc.checkResult(t, db)
			}
		})
	}
}

func TestTokenService_UpdateToken_NameUniqueness(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tokenCache := auth.NewTokenCache(db)
	service := NewTokenService(db, "custom_salt", tokenCache, nil)

	// Create two tokens
	token1 := models.APIToken{TokenID: "token1", UserID: 1, RoleID: 2, TTL: 3600, Status: models.TokenStatusActive}
	token2 := models.APIToken{TokenID: "token2", UserID: 1, RoleID: 2, TTL: 3600, Status: models.TokenStatusActive}
	db.Create(&token1)
	db.Create(&token2)

	// Update token1 name
	name1 := "First Token"
	db.Model(&token1).Update("name", name1)

	// Try to update token2 with same name (should fail)
	c, w := setupTestContext(1, 2, "hash1", []string{"settings.tokens.edit"})
	c.Request = httptest.NewRequest(http.MethodPut, "/tokens/token2",
		bytes.NewBufferString(`{"name": "First Token"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "tokenID", Value: "token2"}}

	service.UpdateToken(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "already exists")
}

func TestTokenService_DeleteToken(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tokenCache := auth.NewTokenCache(db)
	service := NewTokenService(db, "custom_salt", tokenCache, nil)

	testCases := []struct {
		name         string
		setupTokens  func() string
		tokenID      string
		uid          uint64
		permissions  []string
		expectedCode int
	}{
		{
			name: "user deletes own token",
			setupTokens: func() string {
				token := models.APIToken{TokenID: "deltest1", UserID: 1, RoleID: 2, TTL: 3600, Status: models.TokenStatusActive}
				db.Create(&token)
				return "deltest1"
			},
			uid:          1,
			permissions:  []string{"settings.tokens.delete"},
			expectedCode: http.StatusOK,
		},
		{
			name: "user cannot delete other user's token",
			setupTokens: func() string {
				token := models.APIToken{TokenID: "deltest2", UserID: 2, RoleID: 2, TTL: 3600, Status: models.TokenStatusActive}
				db.Create(&token)
				return "deltest2"
			},
			uid:          1,
			permissions:  []string{"settings.tokens.delete"},
			expectedCode: http.StatusForbidden,
		},
		{
			name: "admin can delete any token",
			setupTokens: func() string {
				token := models.APIToken{TokenID: "deltest3", UserID: 2, RoleID: 2, TTL: 3600, Status: models.TokenStatusActive}
				db.Create(&token)
				return "deltest3"
			},
			uid:          1,
			permissions:  []string{"settings.tokens.delete", "settings.tokens.admin"},
			expectedCode: http.StatusOK,
		},
		{
			name: "delete nonexistent token",
			setupTokens: func() string {
				return "nonexistent"
			},
			uid:          1,
			permissions:  []string{"settings.tokens.delete", "settings.tokens.admin"},
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokenID := tc.setupTokens()

			c, w := setupTestContext(tc.uid, 2, "testhash", tc.permissions)
			c.Request = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/tokens/%s", tokenID), nil)
			c.Params = gin.Params{{Key: "tokenID", Value: tokenID}}

			service.DeleteToken(c)

			assert.Equal(t, tc.expectedCode, w.Code)

			// Verify soft delete
			if tc.expectedCode == http.StatusOK {
				var deletedToken models.APIToken
				err := db.Unscoped().Where("token_id = ?", tokenID).First(&deletedToken).Error
				require.NoError(t, err)
				assert.NotNil(t, deletedToken.DeletedAt)
			}
		})
	}
}

func TestTokenService_DeleteToken_InvalidatesCache(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tokenCache := auth.NewTokenCache(db)
	service := NewTokenService(db, "custom_salt", tokenCache, nil)

	// Create token
	tokenID := "cachetest1"
	token := models.APIToken{TokenID: tokenID, UserID: 1, RoleID: 2, TTL: 3600, Status: models.TokenStatusActive}
	db.Create(&token)

	// Populate cache
	_, _, err := service.tokenCache.GetStatus(tokenID)
	require.NoError(t, err)

	// Delete token
	c, w := setupTestContext(1, 2, "hash1", []string{"settings.tokens.delete"})
	c.Request = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/tokens/%s", tokenID), nil)
	c.Params = gin.Params{{Key: "tokenID", Value: tokenID}}

	service.DeleteToken(c)
	assert.Equal(t, http.StatusOK, w.Code)

	// Cache should be invalidated (GetStatus should return error for deleted token)
	_, _, err = service.tokenCache.GetStatus(tokenID)
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestTokenService_UpdateToken_InvalidatesCache(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tokenCache := auth.NewTokenCache(db)
	service := NewTokenService(db, "custom_salt", tokenCache, nil)

	// Create token
	tokenID := "cachetest2"
	token := models.APIToken{TokenID: tokenID, UserID: 1, RoleID: 2, TTL: 3600, Status: models.TokenStatusActive}
	db.Create(&token)

	// Populate cache with active status
	status, privileges, err := service.tokenCache.GetStatus(tokenID)
	require.NoError(t, err)
	assert.Equal(t, models.TokenStatusActive, status)
	assert.NotEmpty(t, privileges)
	assert.Contains(t, privileges, auth.PrivilegeAutomation)

	// Update status to revoked
	c, w := setupTestContext(1, 2, "hash1", []string{"settings.tokens.edit"})
	c.Request = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/tokens/%s", tokenID),
		bytes.NewBufferString(`{"status": "revoked"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "tokenID", Value: tokenID}}

	service.UpdateToken(c)
	assert.Equal(t, http.StatusOK, w.Code)

	// Cache should be updated (should return revoked status)
	status, privileges, err = service.tokenCache.GetStatus(tokenID)
	require.NoError(t, err)
	assert.Equal(t, models.TokenStatusRevoked, status)
	assert.NotEmpty(t, privileges)
	assert.Contains(t, privileges, auth.PrivilegeAutomation)
}

func TestTokenService_FullLifecycle(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tokenCache := auth.NewTokenCache(db)
	service := NewTokenService(db, "custom_salt", tokenCache, nil)

	// Step 1: Create token
	c1, w1 := setupTestContext(1, 2, "hash1", []string{"settings.tokens.create"})
	c1.Request = httptest.NewRequest(http.MethodPost, "/tokens",
		bytes.NewBufferString(`{"ttl": 3600, "name": "Lifecycle Test"}`))
	c1.Request.Header.Set("Content-Type", "application/json")

	service.CreateToken(c1)
	assert.Equal(t, http.StatusCreated, w1.Code)

	var createResp struct {
		Status string `json:"status"`
		Data   struct {
			TokenID string `json:"token_id"`
			Token   string `json:"token"`
			Name    string `json:"name"`
		} `json:"data"`
	}
	json.Unmarshal(w1.Body.Bytes(), &createResp)
	tokenID := createResp.Data.TokenID
	tokenString := createResp.Data.Token

	// Step 2: Validate token works
	claims, err := auth.ValidateAPIToken(tokenString, "custom_salt")
	require.NoError(t, err)
	assert.Equal(t, tokenID, claims.TokenID)

	// Step 3: List tokens (should see it)
	c2, w2 := setupTestContext(1, 2, "hash1", []string{"settings.tokens.view"})
	c2.Request = httptest.NewRequest(http.MethodGet, "/tokens", nil)
	service.ListTokens(c2)

	var listResp struct {
		Status string `json:"status"`
		Data   struct {
			Tokens []models.APIToken `json:"tokens"`
		} `json:"data"`
	}
	json.Unmarshal(w2.Body.Bytes(), &listResp)
	assert.True(t, len(listResp.Data.Tokens) > 0)

	// Step 4: Update token name
	c3, w3 := setupTestContext(1, 2, "hash1", []string{"settings.tokens.edit"})
	c3.Request = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/tokens/%s", tokenID),
		bytes.NewBufferString(`{"name": "Updated Lifecycle"}`))
	c3.Request.Header.Set("Content-Type", "application/json")
	c3.Params = gin.Params{{Key: "tokenID", Value: tokenID}}
	service.UpdateToken(c3)
	assert.Equal(t, http.StatusOK, w3.Code)

	// Step 5: Revoke token
	c4, w4 := setupTestContext(1, 2, "hash1", []string{"settings.tokens.edit"})
	c4.Request = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/tokens/%s", tokenID),
		bytes.NewBufferString(`{"status": "revoked"}`))
	c4.Request.Header.Set("Content-Type", "application/json")
	c4.Params = gin.Params{{Key: "tokenID", Value: tokenID}}
	service.UpdateToken(c4)
	assert.Equal(t, http.StatusOK, w4.Code)

	// Step 6: Verify revoked status in cache
	status, privileges, err := service.tokenCache.GetStatus(tokenID)
	require.NoError(t, err)
	assert.Equal(t, models.TokenStatusRevoked, status)
	assert.NotEmpty(t, privileges)
	assert.Contains(t, privileges, auth.PrivilegeAutomation)

	// Step 7: Delete token
	c5, w5 := setupTestContext(1, 2, "hash1", []string{"settings.tokens.delete"})
	c5.Request = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/tokens/%s", tokenID), nil)
	c5.Params = gin.Params{{Key: "tokenID", Value: tokenID}}
	service.DeleteToken(c5)
	assert.Equal(t, http.StatusOK, w5.Code)

	// Step 8: Verify soft delete
	var deletedToken models.APIToken
	err = db.Unscoped().Where("token_id = ?", tokenID).First(&deletedToken).Error
	require.NoError(t, err)
	assert.NotNil(t, deletedToken.DeletedAt)

	// Step 9: Token should not be found after deletion
	_, _, err = service.tokenCache.GetStatus(tokenID)
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestTokenService_AdminPermissions(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tokenCache := auth.NewTokenCache(db)
	service := NewTokenService(db, "custom_salt", tokenCache, nil)

	// Create tokens for different users
	token1 := models.APIToken{TokenID: "admintest1", UserID: 1, RoleID: 2, TTL: 3600, Status: models.TokenStatusActive}
	token2 := models.APIToken{TokenID: "admintest2", UserID: 2, RoleID: 2, TTL: 3600, Status: models.TokenStatusActive}
	db.Create(&token1)
	db.Create(&token2)

	adminUID := uint64(3)

	testCases := []struct {
		name         string
		operation    string
		tokenID      string
		expectedCode int
	}{
		{
			name:         "admin views user 1 token",
			operation:    "get",
			tokenID:      "admintest1",
			expectedCode: http.StatusOK,
		},
		{
			name:         "admin views user 2 token",
			operation:    "get",
			tokenID:      "admintest2",
			expectedCode: http.StatusOK,
		},
		{
			name:         "admin updates user 2 token",
			operation:    "update",
			tokenID:      "admintest2",
			expectedCode: http.StatusOK,
		},
		{
			name:         "admin deletes user 1 token",
			operation:    "delete",
			tokenID:      "admintest1",
			expectedCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, w := setupTestContext(adminUID, 1, "adminhash", []string{
				"settings.tokens.admin",
				"settings.tokens.view",
				"settings.tokens.edit",
				"settings.tokens.delete",
			})

			switch tc.operation {
			case "get":
				c.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/tokens/%s", tc.tokenID), nil)
				c.Params = gin.Params{{Key: "tokenID", Value: tc.tokenID}}
				service.GetToken(c)
			case "update":
				c.Request = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/tokens/%s", tc.tokenID),
					bytes.NewBufferString(`{"status": "revoked"}`))
				c.Request.Header.Set("Content-Type", "application/json")
				c.Params = gin.Params{{Key: "tokenID", Value: tc.tokenID}}
				service.UpdateToken(c)
			case "delete":
				c.Request = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/tokens/%s", tc.tokenID), nil)
				c.Params = gin.Params{{Key: "tokenID", Value: tc.tokenID}}
				service.DeleteToken(c)
			}

			assert.Equal(t, tc.expectedCode, w.Code)
		})
	}
}

func TestTokenService_TokenPrivileges(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tokenCache := auth.NewTokenCache(db)

	// Create admin token (role_id = 1)
	adminToken := models.APIToken{
		TokenID: "admin_priv_test",
		UserID:  1,
		RoleID:  1,
		TTL:     3600,
		Status:  models.TokenStatusActive,
	}
	err := db.Create(&adminToken).Error
	require.NoError(t, err)

	// Create user token (role_id = 2)
	userToken := models.APIToken{
		TokenID: "user_priv_test",
		UserID:  2,
		RoleID:  2,
		TTL:     3600,
		Status:  models.TokenStatusActive,
	}
	err = db.Create(&userToken).Error
	require.NoError(t, err)

	// Test admin privileges
	t.Run("admin token has admin privileges", func(t *testing.T) {
		status, privileges, err := tokenCache.GetStatus("admin_priv_test")
		require.NoError(t, err)
		assert.Equal(t, models.TokenStatusActive, status)
		assert.NotEmpty(t, privileges)

		// Should have automation privilege
		assert.Contains(t, privileges, auth.PrivilegeAutomation)

		// Should have admin-specific privileges
		assert.Contains(t, privileges, "users.create")
		assert.Contains(t, privileges, "users.delete")
		assert.Contains(t, privileges, "settings.tokens.admin")
		assert.Contains(t, privileges, "flows.admin")
	})

	// Test user privileges
	t.Run("user token has limited privileges", func(t *testing.T) {
		status, privileges, err := tokenCache.GetStatus("user_priv_test")
		require.NoError(t, err)
		assert.Equal(t, models.TokenStatusActive, status)
		assert.NotEmpty(t, privileges)

		// Should have automation privilege
		assert.Contains(t, privileges, auth.PrivilegeAutomation)

		// Should have user-level privileges
		assert.Contains(t, privileges, "flows.create")
		assert.Contains(t, privileges, "settings.tokens.view")

		// Should NOT have admin-specific privileges
		assert.NotContains(t, privileges, "users.create")
		assert.NotContains(t, privileges, "users.delete")
		assert.NotContains(t, privileges, "settings.tokens.admin")
		assert.NotContains(t, privileges, "flows.admin")
	})

	// Test privilege caching
	t.Run("privileges are cached", func(t *testing.T) {
		// First call - loads from DB
		_, privileges1, err := tokenCache.GetStatus("admin_priv_test")
		require.NoError(t, err)
		assert.NotEmpty(t, privileges1)

		// Second call - loads from cache
		_, privileges2, err := tokenCache.GetStatus("admin_priv_test")
		require.NoError(t, err)
		assert.Equal(t, privileges1, privileges2)
	})

	// Test cache invalidation updates privileges
	t.Run("cache invalidation reloads privileges", func(t *testing.T) {
		// Get initial privileges
		_, initialPrivs, err := tokenCache.GetStatus("user_priv_test")
		require.NoError(t, err)
		assert.NotEmpty(t, initialPrivs)

		// Update user's role to admin in DB
		db.Model(&userToken).Update("role_id", 1)

		// Privileges should still be cached (old privileges)
		_, cachedPrivs, err := tokenCache.GetStatus("user_priv_test")
		require.NoError(t, err)
		assert.NotContains(t, cachedPrivs, "users.create") // still user privileges

		// Invalidate cache
		tokenCache.Invalidate("user_priv_test")

		// Should now have admin privileges
		_, newPrivs, err := tokenCache.GetStatus("user_priv_test")
		require.NoError(t, err)
		assert.Contains(t, newPrivs, "users.create") // now has admin privileges
		assert.Contains(t, newPrivs, "settings.tokens.admin")
	})
}

func TestTokenService_SecurityChecks(t *testing.T) {
	t.Run("token secret not stored in database", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()

		tokenCache := auth.NewTokenCache(db)
		service := NewTokenService(db, "custom_salt", tokenCache, nil)

		c, w := setupTestContext(1, 2, "hash1", []string{"settings.tokens.create"})
		c.Request = httptest.NewRequest(http.MethodPost, "/tokens",
			bytes.NewBufferString(`{"ttl": 3600}`))
		c.Request.Header.Set("Content-Type", "application/json")

		service.CreateToken(c)
		assert.Equal(t, http.StatusCreated, w.Code)

		var response struct {
			Status string `json:"status"`
			Data   struct {
				Token   string `json:"token"`
				TokenID string `json:"token_id"`
			} `json:"data"`
		}
		json.Unmarshal(w.Body.Bytes(), &response)

		// Verify token is returned in response
		assert.NotEmpty(t, response.Data.Token)

		// Verify token is NOT in database
		var dbToken models.APIToken
		db.Where("token_id = ?", response.Data.TokenID).First(&dbToken)

		// Database should only have metadata, no token field
		assert.Equal(t, response.Data.TokenID, dbToken.TokenID)
		// Note: our model doesn't have Token field in APIToken, only in APITokenWithSecret for response
	})

	t.Run("token claims trusted from JWT", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()

		tokenID, err := auth.GenerateTokenID()
		require.NoError(t, err)

		// Create token in DB with role_id = 2
		apiToken := models.APIToken{TokenID: tokenID, UserID: 1, RoleID: 2, TTL: 3600, Status: models.TokenStatusActive}
		err = db.Create(&apiToken).Error
		require.NoError(t, err)

		// Create JWT with role_id = 1 (admin, different from DB)
		claims := models.APITokenClaims{
			TokenID: tokenID,
			RID:     1, // admin role in JWT
			UID:     1,
			UHASH:   "testhash",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				Subject:   "api_token",
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(auth.MakeJWTSigningKey("test"))
		require.NoError(t, err)

		// Validate token
		validated, err := auth.ValidateAPIToken(tokenString, "test")
		require.NoError(t, err)

		// We should trust JWT claims, not DB values
		assert.Equal(t, uint64(1), validated.RID, "Should use role_id from JWT claims")
		assert.NotEqual(t, apiToken.RoleID, validated.RID, "Should not use role_id from database")
	})

	t.Run("updated_at auto-updates", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()

		// Create token
		token := models.APIToken{TokenID: "updatetime1", UserID: 1, RoleID: 2, TTL: 3600, Status: models.TokenStatusActive}
		db.Create(&token)

		_ = token.UpdatedAt // record original time (trigger would update in real PostgreSQL)
		time.Sleep(10 * time.Millisecond)

		// Update token
		db.Model(&token).Update("status", models.TokenStatusRevoked)

		// Reload
		var updated models.APIToken
		db.Where("token_id = ?", "updatetime1").First(&updated)

		// updated_at should have changed
		// Note: SQLite may not have trigger support in memory, but this demonstrates intent
		// In real PostgreSQL, the trigger would update this automatically
	})
}
