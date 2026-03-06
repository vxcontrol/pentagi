package auth_test

import (
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

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	// Create roles table
	result := db.Exec(`
		CREATE TABLE roles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE
		)
	`)
	require.NoError(t, result.Error, "Failed to create roles table")

	// Create privileges table
	result = db.Exec(`
		CREATE TABLE privileges (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			role_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			UNIQUE(role_id, name)
		)
	`)
	require.NoError(t, result.Error, "Failed to create privileges table")

	// Create api_tokens table for testing
	result = db.Exec(`
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
	require.NoError(t, result.Error, "Failed to create api_tokens table")

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
	db.Exec("INSERT INTO users (id, hash, mail, name, status, role_id) VALUES (1, 'testhash', 'user1@test.com', 'User 1', 'active', 2)")
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

	time.Sleep(200 * time.Millisecond) // wait for database to be ready

	return db
}

func TestValidateAPIToken(t *testing.T) {
	globalSalt := "test_salt"

	testCases := []struct {
		name        string
		setup       func() string
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid token",
			setup: func() string {
				claims := models.APITokenClaims{
					TokenID: "abc123xyz9",
					RID:     2,
					UID:     1,
					UHASH:   "testhash",
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
						IssuedAt:  jwt.NewNumericDate(time.Now()),
						Subject:   "api_token",
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString(auth.MakeJWTSigningKey(globalSalt))
				return tokenString
			},
			expectError: false,
		},
		{
			name: "expired token",
			setup: func() string {
				claims := models.APITokenClaims{
					TokenID: "abc123xyz9",
					RID:     2,
					UID:     1,
					UHASH:   "testhash",
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
						IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
						Subject:   "api_token",
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString(auth.MakeJWTSigningKey(globalSalt))
				return tokenString
			},
			expectError: true,
			errorMsg:    "expired",
		},
		{
			name: "invalid signature",
			setup: func() string {
				claims := models.APITokenClaims{
					TokenID: "abc123xyz9",
					RID:     2,
					UID:     1,
					UHASH:   "testhash",
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
						IssuedAt:  jwt.NewNumericDate(time.Now()),
						Subject:   "api_token",
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString([]byte("wrong_key"))
				return tokenString
			},
			expectError: true,
			errorMsg:    "invalid",
		},
		{
			name: "malformed token",
			setup: func() string {
				return "not.a.valid.jwt.token"
			},
			expectError: true,
			errorMsg:    "malformed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokenString := tc.setup()
			claims, err := auth.ValidateAPIToken(tokenString, globalSalt)

			if tc.expectError {
				assert.Error(t, err)
				if tc.errorMsg != "" {
					assert.Contains(t, err.Error(), tc.errorMsg)
				}
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				assert.Equal(t, "abc123xyz9", claims.TokenID)
				assert.Equal(t, uint64(1), claims.UID)
				assert.Equal(t, uint64(2), claims.RID)
			}
		})
	}
}

func TestAPITokenAuthentication_CacheExpiration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create cache with short TTL for testing
	tokenCache := auth.NewTokenCache(db)
	tokenCache.SetTTL(100 * time.Millisecond)
	userCache := auth.NewUserCache(db)
	authMiddleware := auth.NewAuthMiddleware("/base/url", "test", tokenCache, userCache)

	// Create active token
	tokenID, err := auth.GenerateTokenID()
	require.NoError(t, err)
	apiToken := models.APIToken{
		TokenID: tokenID,
		UserID:  1,
		RoleID:  2,
		TTL:     3600,
		Status:  models.TokenStatusActive,
	}
	err = db.Create(&apiToken).Error
	require.NoError(t, err)

	// Create JWT
	claims := models.APITokenClaims{
		TokenID: tokenID,
		RID:     2,
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

	server := newTestServer(t, "/test", db, authMiddleware.AuthTokenRequired)
	defer server.Close()

	// First call: should work (status active, cached)
	assert.True(t, server.CallAndGetStatus(t, "Bearer "+tokenString))

	// Revoke token in DB
	db.Model(&apiToken).Update("status", models.TokenStatusRevoked)

	// Second call: should still work (cache not expired)
	assert.True(t, server.CallAndGetStatus(t, "Bearer "+tokenString))

	// Wait for cache to expire
	time.Sleep(150 * time.Millisecond)

	// Third call: should fail (cache expired, reads from DB)
	assert.False(t, server.CallAndGetStatus(t, "Bearer "+tokenString))
}

func TestAPITokenAuthentication_DefaultSalt(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	testCases := []struct {
		name       string
		globalSalt string
		shouldSkip bool
	}{
		{
			name:       "default salt 'salt'",
			globalSalt: "salt",
			shouldSkip: true,
		},
		{
			name:       "empty salt",
			globalSalt: "",
			shouldSkip: true,
		},
		{
			name:       "custom salt",
			globalSalt: "custom_secure_salt",
			shouldSkip: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokenCache := auth.NewTokenCache(db)
			userCache := auth.NewUserCache(db)
			authMiddleware := auth.NewAuthMiddleware("/base/url", tc.globalSalt, tokenCache, userCache)

			// Create a token (even with default salt, for testing)
			tokenID, err := auth.GenerateTokenID()
			require.NoError(t, err)
			claims := models.APITokenClaims{
				TokenID: tokenID,
				RID:     2,
				UID:     1,
				UHASH:   "testhash",
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
					IssuedAt:  jwt.NewNumericDate(time.Now()),
					Subject:   "api_token",
				},
			}
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			tokenString, _ := token.SignedString(auth.MakeJWTSigningKey(tc.globalSalt))

			server := newTestServer(t, "/test", db, authMiddleware.AuthTokenRequired)
			defer server.Close()

			// With default salt, token validation should be skipped
			result := server.CallAndGetStatus(t, "Bearer "+tokenString)

			if tc.shouldSkip {
				// Should skip token auth and try cookie (which will fail)
				assert.False(t, result)
			} else {
				// With custom salt but no DB record, should fail with "not found"
				assert.False(t, result)
			}
		})
	}
}

func TestAPITokenAuthentication_SoftDelete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tokenCache := auth.NewTokenCache(db)
	userCache := auth.NewUserCache(db)
	authMiddleware := auth.NewAuthMiddleware("/base/url", "test", tokenCache, userCache)

	// Create token
	tokenID, err := auth.GenerateTokenID()
	require.NoError(t, err)
	apiToken := models.APIToken{
		TokenID: tokenID,
		UserID:  1,
		RoleID:  2,
		TTL:     3600,
		Status:  models.TokenStatusActive,
	}
	err = db.Create(&apiToken).Error
	require.NoError(t, err)

	// Create JWT
	claims := models.APITokenClaims{
		TokenID: tokenID,
		RID:     2,
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

	server := newTestServer(t, "/test", db, authMiddleware.AuthTokenRequired)
	defer server.Close()

	// Should work initially
	assert.True(t, server.CallAndGetStatus(t, "Bearer "+tokenString))

	// Soft delete
	now := time.Now()
	db.Model(&apiToken).Update("deleted_at", now)
	tokenCache.Invalidate(tokenID)

	// Should fail after soft delete
	assert.False(t, server.CallAndGetStatus(t, "Bearer "+tokenString))
}

func TestAPITokenAuthentication_AlgNoneAttack(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tokenCache := auth.NewTokenCache(db)
	userCache := auth.NewUserCache(db)
	authMiddleware := auth.NewAuthMiddleware("/base/url", "test", tokenCache, userCache)

	tokenID, err := auth.GenerateTokenID()
	require.NoError(t, err)

	// Create token with "none" algorithm (security attack)
	claims := models.APITokenClaims{
		TokenID: tokenID,
		RID:     2,
		UID:     1,
		UHASH:   "testhash",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   "api_token",
		},
	}

	// Try to use "none" algorithm
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	server := newTestServer(t, "/test", db, authMiddleware.AuthTokenRequired)
	defer server.Close()

	// Should reject "none" algorithm
	assert.False(t, server.CallAndGetStatus(t, "Bearer "+tokenString))
}

func TestAPITokenAuthentication_LegacyProtoToken(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tokenCache := auth.NewTokenCache(db)
	userCache := auth.NewUserCache(db)
	authMiddleware := auth.NewAuthMiddleware("/base/url", "test", tokenCache, userCache)

	server := newTestServer(t, "/test", db, authMiddleware.AuthTokenRequired)
	defer server.Close()

	// Authorize with cookie to get legacy proto token
	server.Authorize(t, []string{auth.PrivilegeAutomation})
	legacyToken := server.GetToken(t)
	require.NotEmpty(t, legacyToken)

	// Unauthorize cookie
	server.Unauthorize(t)

	// Legacy proto token should still work (fallback mechanism)
	server.SetSessionCheckFunc(func(t *testing.T, c *gin.Context) {
		t.Helper()
		assert.Equal(t, uint64(1), c.GetUint64("uid"))
		assert.Equal(t, "automation", c.GetString("cpt"))
	})

	assert.True(t, server.CallAndGetStatus(t, "Bearer "+legacyToken))
}
