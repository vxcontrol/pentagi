package auth_test

import (
	"sync"
	"testing"
	"time"

	"pentagi/pkg/server/auth"
	"pentagi/pkg/server/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEndToEndAPITokenFlow tests complete flow from creation to usage
func TestEndToEndAPITokenFlow(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tokenCache := auth.NewTokenCache(db)
	userCache := auth.NewUserCache(db)
	authMiddleware := auth.NewAuthMiddleware("/base/url", "test_salt", tokenCache, userCache)

	testCases := []struct {
		name          string
		tokenID       string
		status        models.TokenStatus
		shouldPass    bool
		errorContains string
	}{
		{
			name:       "active token authenticates successfully",
			tokenID:    "active123",
			status:     models.TokenStatusActive,
			shouldPass: true,
		},
		{
			name:          "revoked token is rejected",
			tokenID:       "revoked456",
			status:        models.TokenStatusRevoked,
			shouldPass:    false,
			errorContains: "revoked",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create token in database
			apiToken := models.APIToken{
				TokenID: tc.tokenID,
				UserID:  1,
				RoleID:  2,
				TTL:     3600,
				Status:  tc.status,
			}
			err := db.Create(&apiToken).Error
			require.NoError(t, err)

			// Create JWT token
			claims := models.APITokenClaims{
				TokenID: tc.tokenID,
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
			tokenString, err := token.SignedString(auth.MakeJWTSigningKey("test_salt"))
			require.NoError(t, err)

			// Test authentication
			server := newTestServer(t, "/protected", db, authMiddleware.AuthTokenRequired)
			defer server.Close()

			success := server.CallAndGetStatus(t, "Bearer "+tokenString)
			assert.Equal(t, tc.shouldPass, success)
		})
	}
}

// TestAPIToken_RoleIsolation verifies that token inherits creator's role
func TestAPIToken_RoleIsolation(t *testing.T) {
	testCases := []struct {
		name        string
		creatorRole uint64
		tokenRole   uint64
		expectMatch bool
	}{
		{
			name:        "user creates token with user role",
			creatorRole: 2,
			tokenRole:   2,
			expectMatch: true,
		},
		{
			name:        "admin creates token with admin role",
			creatorRole: 1,
			tokenRole:   1,
			expectMatch: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokenID, err := auth.GenerateTokenID()
			require.NoError(t, err)

			// Create JWT with specific role
			claims := models.APITokenClaims{
				TokenID: tokenID,
				RID:     tc.tokenRole,
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

			// Validate and check role
			validated, err := auth.ValidateAPIToken(tokenString, "test")
			require.NoError(t, err)

			if tc.expectMatch {
				assert.Equal(t, tc.tokenRole, validated.RID)
			}
		})
	}
}

// TestAPIToken_SignatureVerification tests various signature attacks
func TestAPIToken_SignatureVerification(t *testing.T) {
	correctSalt := "correct_salt"
	wrongSalt := "wrong_salt"

	testCases := []struct {
		name          string
		signSalt      string
		verifySalt    string
		expectValid   bool
		errorContains string
	}{
		{
			name:        "matching salt - valid",
			signSalt:    correctSalt,
			verifySalt:  correctSalt,
			expectValid: true,
		},
		{
			name:          "mismatched salt - invalid",
			signSalt:      correctSalt,
			verifySalt:    wrongSalt,
			expectValid:   false,
			errorContains: "invalid",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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
			tokenString, err := token.SignedString(auth.MakeJWTSigningKey(tc.signSalt))
			require.NoError(t, err)

			validated, err := auth.ValidateAPIToken(tokenString, tc.verifySalt)

			if tc.expectValid {
				assert.NoError(t, err)
				assert.NotNil(t, validated)
			} else {
				assert.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			}
		})
	}
}

// TestAPIToken_CacheInvalidation verifies cache invalidation scenarios
func TestAPIToken_CacheInvalidation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tokenCache := auth.NewTokenCache(db)

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

	// Load into cache
	status1, _, err := tokenCache.GetStatus(tokenID)
	require.NoError(t, err)
	assert.Equal(t, models.TokenStatusActive, status1)

	// Update in DB
	db.Model(&apiToken).Update("status", models.TokenStatusRevoked)

	// Should still return active from cache
	status2, _, err := tokenCache.GetStatus(tokenID)
	require.NoError(t, err)
	assert.Equal(t, models.TokenStatusActive, status2, "Cache should return stale value")

	// Invalidate cache
	tokenCache.Invalidate(tokenID)

	// Should now return revoked from DB
	status3, _, err := tokenCache.GetStatus(tokenID)
	require.NoError(t, err)
	assert.Equal(t, models.TokenStatusRevoked, status3, "Cache should be refreshed from DB")
}

// TestAPIToken_ConcurrentAccess tests thread-safety of cache
func TestAPIToken_ConcurrentAccess(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tokenCache := auth.NewTokenCache(db)

	// Create multiple tokens
	tokenIDs := make([]string, 10)
	for i := range 10 {
		tokenID, err := auth.GenerateTokenID()
		require.NoError(t, err)
		tokenIDs[i] = tokenID
		apiToken := models.APIToken{
			TokenID: tokenID,
			UserID:  1,
			RoleID:  2,
			TTL:     3600,
			Status:  models.TokenStatusActive,
		}
		err = db.Create(&apiToken).Error
		require.NoError(t, err)
	}

	// Verify tokens were created
	var count int
	db.Model(&models.APIToken{}).Where("deleted_at IS NULL").Count(&count)
	require.Equal(t, 10, count)

	// Warm up cache
	for i := range 10 {
		status, _, err := tokenCache.GetStatus(tokenIDs[i])
		require.NoError(t, err)
		assert.Equal(t, models.TokenStatusActive, status)
	}

	// Concurrent cache access using channels for error reporting
	type testResult struct {
		success bool
		err     error
	}
	results := make(chan testResult, 10)

	var wg sync.WaitGroup
	wg.Add(10)
	for i := range 10 {
		go func(tokenID string) {
			defer wg.Done()
			for range 100 {
				status, _, err := tokenCache.GetStatus(tokenID)
				if err != nil {
					results <- testResult{success: false, err: err}
					return
				}
				if status != models.TokenStatusActive {
					results <- testResult{success: false, err: assert.AnError}
					return
				}
			}
			results <- testResult{success: true, err: nil}
		}(tokenIDs[i])
	}

	wg.Wait()
	close(results)

	// Wait and check all results
	for result := range results {
		assert.NoError(t, result.err)
		assert.True(t, result.success, "Goroutine should complete successfully")
	}
}

// TestAPIToken_JSONStructure verifies JWT payload structure
func TestAPIToken_JSONStructure(t *testing.T) {
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
	tokenString, err := token.SignedString(auth.MakeJWTSigningKey("test"))
	require.NoError(t, err)

	// Parse and verify all fields
	parsed, err := auth.ValidateAPIToken(tokenString, "test")
	require.NoError(t, err)

	assert.Equal(t, tokenID, parsed.TokenID, "TokenID should match")
	assert.Equal(t, uint64(2), parsed.RID, "RID should match")
	assert.Equal(t, uint64(1), parsed.UID, "UID should match")
	assert.Equal(t, "testhash", parsed.UHASH, "UHASH should match")
	assert.Equal(t, "api_token", parsed.Subject, "Subject should match")
	assert.NotNil(t, parsed.ExpiresAt, "ExpiresAt should be set")
	assert.NotNil(t, parsed.IssuedAt, "IssuedAt should be set")
}

// TestAPIToken_Expiration verifies TTL enforcement
func TestAPIToken_Expiration(t *testing.T) {
	testCases := []struct {
		name        string
		ttl         time.Duration
		expectValid bool
	}{
		{
			name:        "future expiration - valid",
			ttl:         1 * time.Hour,
			expectValid: true,
		},
		{
			name:        "past expiration - invalid",
			ttl:         -1 * time.Hour,
			expectValid: false,
		},
		{
			name:        "just expired - invalid",
			ttl:         -1 * time.Second,
			expectValid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokenID, err := auth.GenerateTokenID()
			require.NoError(t, err)

			claims := models.APITokenClaims{
				TokenID: tokenID,
				RID:     2,
				UID:     1,
				UHASH:   "testhash",
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(tc.ttl)),
					IssuedAt:  jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
					Subject:   "api_token",
				},
			}

			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			tokenString, err := token.SignedString(auth.MakeJWTSigningKey("test"))
			require.NoError(t, err)

			validated, err := auth.ValidateAPIToken(tokenString, "test")

			if tc.expectValid {
				assert.NoError(t, err)
				assert.NotNil(t, validated)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "expired")
			}
		})
	}
}

// TestDualAuthentication verifies both cookie and token auth work together
func TestDualAuthentication(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tokenCache := auth.NewTokenCache(db)
	userCache := auth.NewUserCache(db)
	authMiddleware := auth.NewAuthMiddleware("/base/url", "test", tokenCache, userCache)

	server := newTestServer(t, "/test", db, authMiddleware.AuthTokenRequired)
	defer server.Close()

	// Test 1: Cookie authentication
	server.Authorize(t, []string{auth.PrivilegeAutomation})
	assert.True(t, server.CallAndGetStatus(t), "Cookie auth should work")

	// Test 2: Create and use API token
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
	tokenString, _ := token.SignedString(auth.MakeJWTSigningKey("test"))

	// Unauthorize cookie
	server.Unauthorize(t)

	// Test 3: Token authentication should work
	assert.True(t, server.CallAndGetStatus(t, "Bearer "+tokenString), "Token auth should work")

	// Test 4: Both should work simultaneously
	server.Authorize(t, []string{auth.PrivilegeAutomation})
	assert.True(t, server.CallAndGetStatus(t, "Bearer "+tokenString), "Both auth methods should work")
}

// TestSecurityAudit_ClaimsInJWT verifies all security-critical data is in JWT
func TestSecurityAudit_ClaimsInJWT(t *testing.T) {
	// Create token in DB with certain values
	tokenID, err := auth.GenerateTokenID()
	require.NoError(t, err)
	dbToken := models.APIToken{
		TokenID: tokenID,
		UserID:  1,
		RoleID:  2, // User role in DB
	}

	// Create JWT with different role (simulating compromise scenario)
	jwtClaims := models.APITokenClaims{
		TokenID: tokenID,
		RID:     1, // Admin role in JWT (different from DB!)
		UID:     1,
		UHASH:   "testhash",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   "api_token",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
	tokenString, _ := token.SignedString(auth.MakeJWTSigningKey("test"))

	// Validate token
	validated, err := auth.ValidateAPIToken(tokenString, "test")
	require.NoError(t, err)

	// We trust JWT claims, not DB values
	assert.Equal(t, uint64(1), validated.RID, "Should use role from JWT, not DB")
	assert.NotEqual(t, dbToken.RoleID, validated.RID, "JWT role differs from DB role")
	assert.Equal(t, dbToken.UserID, validated.UID)
	assert.Equal(t, dbToken.TokenID, validated.TokenID)

	// This is CORRECT behavior: DB only stores metadata for management
	// Actual authorization data comes from signed JWT
}

// TestSecurityAudit_TokenIDUniqueness verifies token ID collision resistance
func TestSecurityAudit_TokenIDUniqueness(t *testing.T) {
	iterations := 10000
	tokens := make(map[string]bool, iterations)

	for i := 0; i < iterations; i++ {
		tokenID, err := auth.GenerateTokenID()
		require.NoError(t, err)

		// Check format
		assert.Len(t, tokenID, 10)

		// Check uniqueness
		if tokens[tokenID] {
			t.Fatalf("Duplicate token ID generated: %s", tokenID)
		}
		tokens[tokenID] = true
	}

	t.Logf("Generated %d unique token IDs without collision", iterations)
}

// TestSecurityAudit_SaltIsolation verifies JWT and Cookie keys are different
func TestSecurityAudit_SaltIsolation(t *testing.T) {
	salts := []string{"salt1", "salt2", "production_salt"}

	for _, salt := range salts {
		t.Run("salt="+salt, func(t *testing.T) {
			jwtKey := auth.MakeJWTSigningKey(salt)
			cookieKeys := auth.MakeCookieStoreKey(salt)

			// JWT key must be different from both cookie keys
			assert.NotEqual(t, jwtKey, cookieKeys[0], "JWT key must differ from cookie auth key")
			assert.NotEqual(t, jwtKey, cookieKeys[1], "JWT key must differ from cookie encryption key")

			// Verify key lengths
			assert.Len(t, jwtKey, 32, "JWT key must be 32 bytes")
			assert.Len(t, cookieKeys[0], 64, "Cookie auth key must be 64 bytes")
			assert.Len(t, cookieKeys[1], 32, "Cookie encryption key must be 32 bytes")
		})
	}
}

// TestAPIToken_ContextSetup verifies correct context values are set
func TestAPIToken_ContextSetup(t *testing.T) {
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
		UserID:  5,
		RoleID:  3,
		TTL:     3600,
		Status:  models.TokenStatusActive,
	}
	err = db.Create(&apiToken).Error
	require.NoError(t, err)

	user := models.User{
		ID:     5,
		Hash:   "user5hash",
		Mail:   "user5@example.com",
		Name:   "User 5",
		Status: models.UserStatusActive,
		RoleID: 2,
	}
	err = db.Create(&user).Error
	require.NoError(t, err)

	claims := models.APITokenClaims{
		TokenID: tokenID,
		RID:     2,
		UID:     5,
		UHASH:   "user5hash",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   "api_token",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(auth.MakeJWTSigningKey("test"))

	server := newTestServer(t, "/test", db, authMiddleware.AuthTokenRequired)
	defer server.Close()

	server.SetSessionCheckFunc(func(t *testing.T, c *gin.Context) {
		t.Helper()

		// Verify all context values are set correctly
		assert.Equal(t, uint64(5), c.GetUint64("uid"), "UID from JWT")
		assert.Equal(t, uint64(2), c.GetUint64("rid"), "RID from JWT")
		assert.Equal(t, "user5hash", c.GetString("uhash"), "UHASH from JWT")
		assert.Equal(t, "automation", c.GetString("cpt"), "CPT from JWT")
		assert.Equal(t, "api", c.GetString("tid"), "TID should be 'api' for API tokens")

		prms := c.GetStringSlice("prm")
		assert.Contains(t, prms, auth.PrivilegeAutomation, "Should have automation privilege")

		// Verify session timing fields
		gtm := c.GetInt64("gtm")
		assert.Greater(t, gtm, int64(0), "GTM (generation time) should be set")

		exp := c.GetInt64("exp")
		assert.Greater(t, exp, gtm, "EXP (expiration time) should be greater than GTM")

		// UUID might be empty if hash is invalid (which is expected in tests)
		uuid := c.GetString("uuid")
		assert.NotNil(t, uuid, "UUID should be set (even if empty)")
	})

	assert.True(t, server.CallAndGetStatus(t, "Bearer "+tokenString))
}

// TestUserHashValidation_CookieAuth tests uhash validation with cookie authentication
func TestUserHashValidation_CookieAuth(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tokenCache := auth.NewTokenCache(db)
	userCache := auth.NewUserCache(db)
	authMiddleware := auth.NewAuthMiddleware("/base/url", "test", tokenCache, userCache)

	server := newTestServer(t, "/test", db, authMiddleware.AuthUserRequired)
	defer server.Close()

	// Create test user with ID=1 and hash="123" to match session
	var count int
	db.Model(&models.User{}).Where("id = ?", 1).Count(&count)
	testUser := models.User{
		ID:     1,
		Hash:   "123",
		Mail:   "test_user@example.com",
		Name:   "Test User",
		Status: models.UserStatusActive,
		RoleID: 2,
	}
	if count == 0 {
		err := db.Create(&testUser).Error
		require.NoError(t, err)
	} else {
		db.First(&testUser, 1)
	}

	t.Run("correct uhash succeeds", func(t *testing.T) {
		server.Authorize(t, []string{"test.permission"})
		assert.True(t, server.CallAndGetStatus(t))
	})

	t.Run("modified uhash in database fails", func(t *testing.T) {
		// Update user hash in database
		db.Model(&testUser).Where("id = ?", 1).Update("hash", "modified_hash")
		userCache.Invalidate(1)

		// Try to authenticate with old session (has hash="123")
		assert.False(t, server.CallAndGetStatus(t))
	})

	t.Run("blocked user fails", func(t *testing.T) {
		// Restore original hash
		db.Model(&testUser).Where("id = ?", 1).Update("hash", "123")
		// Block user
		db.Model(&testUser).Where("id = ?", 1).Update("status", models.UserStatusBlocked)
		userCache.Invalidate(1)

		assert.False(t, server.CallAndGetStatus(t))
	})

	t.Run("deleted user fails", func(t *testing.T) {
		// Undelete and unblock first
		db.Model(&models.User{}).Unscoped().Where("id = ?", 1).Update("deleted_at", nil)
		db.Model(&testUser).Where("id = ?", 1).Update("status", models.UserStatusActive)

		// Delete user
		db.Delete(&testUser, 1)
		userCache.Invalidate(1)

		assert.False(t, server.CallAndGetStatus(t))
	})
}

// TestUserHashValidation_TokenAuth tests uhash validation with token authentication
func TestUserHashValidation_TokenAuth(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tokenCache := auth.NewTokenCache(db)
	userCache := auth.NewUserCache(db)
	authMiddleware := auth.NewAuthMiddleware("/base/url", "test_salt", tokenCache, userCache)

	server := newTestServer(t, "/protected", db, authMiddleware.AuthTokenRequired)
	defer server.Close()

	// Create test user
	testUser := models.User{
		ID:     200,
		Hash:   "token_test_hash",
		Mail:   "token_user@example.com",
		Name:   "Token Test User",
		Status: models.UserStatusActive,
		RoleID: 2,
	}
	err := db.Create(&testUser).Error
	require.NoError(t, err)

	// Create API token
	tokenID, err := auth.GenerateTokenID()
	require.NoError(t, err)
	apiToken := models.APIToken{
		TokenID: tokenID,
		UserID:  200,
		RoleID:  2,
		TTL:     3600,
		Status:  models.TokenStatusActive,
	}
	err = db.Create(&apiToken).Error
	require.NoError(t, err)

	// Create JWT token with correct hash
	claims := models.APITokenClaims{
		TokenID: tokenID,
		RID:     2,
		UID:     200,
		UHASH:   "token_test_hash",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   "api_token",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(auth.MakeJWTSigningKey("test_salt"))
	require.NoError(t, err)

	t.Run("correct uhash succeeds", func(t *testing.T) {
		success := server.CallAndGetStatus(t, "Bearer "+tokenString)
		assert.True(t, success)
	})

	t.Run("modified uhash in database fails", func(t *testing.T) {
		// Update user hash in database
		db.Model(&testUser).Update("hash", "different_hash")
		userCache.Invalidate(200)

		// Try to authenticate with token (has original hash)
		success := server.CallAndGetStatus(t, "Bearer "+tokenString)
		assert.False(t, success)
	})

	t.Run("blocked user fails", func(t *testing.T) {
		// Restore original hash
		db.Model(&testUser).Update("hash", "token_test_hash")
		// Block user
		db.Model(&testUser).Update("status", models.UserStatusBlocked)
		userCache.Invalidate(200)

		success := server.CallAndGetStatus(t, "Bearer "+tokenString)
		assert.False(t, success)
	})

	t.Run("deleted user fails", func(t *testing.T) {
		// Unblock and restore for clean state
		db.Model(&models.User{}).Unscoped().Where("id = ?", 200).Update("deleted_at", nil)
		db.Model(&testUser).Update("status", models.UserStatusActive)

		// Delete user
		db.Delete(&testUser)
		userCache.Invalidate(200)

		success := server.CallAndGetStatus(t, "Bearer "+tokenString)
		assert.False(t, success)
	})
}

// TestUserHashValidation_CrossInstallation simulates different installations
func TestUserHashValidation_CrossInstallation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tokenCache := auth.NewTokenCache(db)
	userCache := auth.NewUserCache(db)
	authMiddleware := auth.NewAuthMiddleware("/base/url", "test_salt", tokenCache, userCache)

	server := newTestServer(t, "/protected", db, authMiddleware.AuthTokenRequired)
	defer server.Close()

	// Simulate Installation A
	userInstallationA := models.User{
		ID:     300,
		Hash:   "installation_a_hash",
		Mail:   "cross@example.com",
		Name:   "Cross Installation User",
		Status: models.UserStatusActive,
		RoleID: 2,
	}
	err := db.Create(&userInstallationA).Error
	require.NoError(t, err)

	// Create API token for Installation A
	tokenID, err := auth.GenerateTokenID()
	require.NoError(t, err)
	apiToken := models.APIToken{
		TokenID: tokenID,
		UserID:  300,
		RoleID:  2,
		TTL:     3600,
		Status:  models.TokenStatusActive,
	}
	err = db.Create(&apiToken).Error
	require.NoError(t, err)

	// Create JWT token with Installation A hash
	claimsA := models.APITokenClaims{
		TokenID: tokenID,
		RID:     2,
		UID:     300,
		UHASH:   "installation_a_hash",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   "api_token",
		},
	}
	tokenA := jwt.NewWithClaims(jwt.SigningMethodHS256, claimsA)
	tokenStringA, err := tokenA.SignedString(auth.MakeJWTSigningKey("test_salt"))
	require.NoError(t, err)

	t.Run("token works on Installation A", func(t *testing.T) {
		success := server.CallAndGetStatus(t, "Bearer "+tokenStringA)
		assert.True(t, success)
	})

	t.Run("token from Installation A fails on Installation B", func(t *testing.T) {
		// Simulate Installation B - user has different hash
		db.Model(&userInstallationA).Update("hash", "installation_b_hash")
		userCache.Invalidate(300)

		// Try to use token from Installation A (has installation_a_hash)
		success := server.CallAndGetStatus(t, "Bearer "+tokenStringA)
		assert.False(t, success, "Token from Installation A should not work on Installation B")
	})

	t.Run("new token from Installation B works", func(t *testing.T) {
		// Create new token for Installation B
		tokenIDB, err := auth.GenerateTokenID()
		require.NoError(t, err)
		apiTokenB := models.APIToken{
			TokenID: tokenIDB,
			UserID:  300,
			RoleID:  2,
			TTL:     3600,
			Status:  models.TokenStatusActive,
		}
		err = db.Create(&apiTokenB).Error
		require.NoError(t, err)

		// Create JWT token with Installation B hash
		claimsB := models.APITokenClaims{
			TokenID: tokenIDB,
			RID:     2,
			UID:     300,
			UHASH:   "installation_b_hash", // correct hash for Installation B
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				Subject:   "api_token",
			},
		}
		tokenB := jwt.NewWithClaims(jwt.SigningMethodHS256, claimsB)
		tokenStringB, err := tokenB.SignedString(auth.MakeJWTSigningKey("test_salt"))
		require.NoError(t, err)

		// Token from Installation B should work
		success := server.CallAndGetStatus(t, "Bearer "+tokenStringB)
		assert.True(t, success, "New token from Installation B should work")
	})
}
