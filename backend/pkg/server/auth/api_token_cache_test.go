package auth_test

import (
	"testing"
	"time"

	"pentagi/pkg/server/auth"
	"pentagi/pkg/server/models"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenCache_GetStatus(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	cache := auth.NewTokenCache(db)
	tokenID := "testtoken1"

	// Insert test token
	token := models.APIToken{
		TokenID: tokenID,
		UserID:  1,
		RoleID:  2,
		TTL:     3600,
		Status:  models.TokenStatusActive,
	}
	err := db.Create(&token).Error
	require.NoError(t, err)

	// Test: Get status (should hit database)
	status, privileges, err := cache.GetStatus(tokenID)
	require.NoError(t, err)
	assert.Equal(t, models.TokenStatusActive, status)
	assert.NotEmpty(t, privileges)
	assert.Contains(t, privileges, auth.PrivilegeAutomation)
	assert.Contains(t, privileges, "flows.create")
	assert.Contains(t, privileges, "settings.tokens.view")

	// Test: Get status again (should hit cache)
	status, privileges, err = cache.GetStatus(tokenID)
	require.NoError(t, err)
	assert.Equal(t, models.TokenStatusActive, status)
	assert.NotEmpty(t, privileges)
	assert.Contains(t, privileges, auth.PrivilegeAutomation)

	// Test: Non-existent token
	_, _, err = cache.GetStatus("nonexistent")
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestTokenCache_Invalidate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	cache := auth.NewTokenCache(db)
	tokenID := "testtoken2"

	// Insert test token
	token := models.APIToken{
		TokenID: tokenID,
		UserID:  1,
		RoleID:  2,
		TTL:     3600,
		Status:  models.TokenStatusActive,
	}
	err := db.Create(&token).Error
	require.NoError(t, err)

	// Get status to populate cache
	status, privileges, err := cache.GetStatus(tokenID)
	require.NoError(t, err)
	assert.Equal(t, models.TokenStatusActive, status)
	assert.NotEmpty(t, privileges)

	// Update token in database
	db.Model(&token).Update("status", models.TokenStatusRevoked)

	// Status should still be active (from cache)
	status, privileges, err = cache.GetStatus(tokenID)
	require.NoError(t, err)
	assert.Equal(t, models.TokenStatusActive, status)
	assert.NotEmpty(t, privileges)

	// Invalidate cache
	cache.Invalidate(tokenID)

	// Status should now be revoked (from database)
	status, privileges, err = cache.GetStatus(tokenID)
	require.NoError(t, err)
	assert.Equal(t, models.TokenStatusRevoked, status)
	assert.NotEmpty(t, privileges)
}

func TestTokenCache_InvalidateUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	cache := auth.NewTokenCache(db)
	userID := uint64(1)

	// Insert multiple tokens for user
	tokens := []models.APIToken{
		{
			TokenID: "token1",
			UserID:  userID,
			RoleID:  2,
			TTL:     3600,
			Status:  models.TokenStatusActive,
		},
		{
			TokenID: "token2",
			UserID:  userID,
			RoleID:  2,
			TTL:     3600,
			Status:  models.TokenStatusActive,
		},
	}

	for _, token := range tokens {
		err := db.Create(&token).Error
		require.NoError(t, err)
	}

	// Populate cache
	for _, token := range tokens {
		_, _, err := cache.GetStatus(token.TokenID)
		require.NoError(t, err)
	}

	// Update tokens in database
	db.Model(&models.APIToken{}).Where("user_id = ?", userID).Update("status", models.TokenStatusRevoked)

	// Invalidate all user tokens
	cache.InvalidateUser(userID)

	// All tokens should now show revoked status
	for _, token := range tokens {
		status, privileges, err := cache.GetStatus(token.TokenID)
		require.NoError(t, err)
		assert.Equal(t, models.TokenStatusRevoked, status)
		assert.NotEmpty(t, privileges)
	}
}

func TestTokenCache_Expiration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create cache with very short TTL for testing
	cache := auth.NewTokenCache(db)
	cache.SetTTL(300 * time.Millisecond)

	tokenID := "testtoken3"

	// Insert test token
	token := models.APIToken{
		TokenID: tokenID,
		UserID:  1,
		RoleID:  2,
		TTL:     3600,
		Status:  models.TokenStatusActive,
	}
	err := db.Create(&token).Error
	require.NoError(t, err)

	// Get status to populate cache
	status, privileges, err := cache.GetStatus(tokenID)
	require.NoError(t, err)
	assert.Equal(t, models.TokenStatusActive, status)
	assert.NotEmpty(t, privileges)

	// Update token in database
	db.Model(&token).Update("status", models.TokenStatusRevoked)

	// Wait for cache to expire
	time.Sleep(500 * time.Millisecond)

	// Status should now be revoked (cache expired, reading from DB)
	status, privileges, err = cache.GetStatus(tokenID)
	require.NoError(t, err)
	assert.Equal(t, models.TokenStatusRevoked, status)
	assert.NotEmpty(t, privileges)
}

func TestTokenCache_PrivilegesByRole(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	cache := auth.NewTokenCache(db)

	// Test Admin token (role_id = 1)
	adminTokenID := "admin_token"
	adminToken := models.APIToken{
		TokenID: adminTokenID,
		UserID:  1,
		RoleID:  1,
		TTL:     3600,
		Status:  models.TokenStatusActive,
	}
	err := db.Create(&adminToken).Error
	require.NoError(t, err)

	status, adminPrivs, err := cache.GetStatus(adminTokenID)
	require.NoError(t, err)
	assert.Equal(t, models.TokenStatusActive, status)
	assert.NotEmpty(t, adminPrivs)
	assert.Contains(t, adminPrivs, auth.PrivilegeAutomation)
	assert.Contains(t, adminPrivs, "users.create")
	assert.Contains(t, adminPrivs, "users.delete")
	assert.Contains(t, adminPrivs, "settings.tokens.admin")

	// Test User token (role_id = 2)
	userTokenID := "user_token"
	userToken := models.APIToken{
		TokenID: userTokenID,
		UserID:  2,
		RoleID:  2,
		TTL:     3600,
		Status:  models.TokenStatusActive,
	}
	err = db.Create(&userToken).Error
	require.NoError(t, err)

	status, userPrivs, err := cache.GetStatus(userTokenID)
	require.NoError(t, err)
	assert.Equal(t, models.TokenStatusActive, status)
	assert.NotEmpty(t, userPrivs)
	assert.Contains(t, userPrivs, auth.PrivilegeAutomation)
	assert.Contains(t, userPrivs, "flows.create")
	assert.Contains(t, userPrivs, "settings.tokens.view")

	// User should NOT have admin privileges
	assert.NotContains(t, userPrivs, "users.create")
	assert.NotContains(t, userPrivs, "users.delete")
	assert.NotContains(t, userPrivs, "settings.tokens.admin")

	// Admin should have more privileges than User
	assert.Greater(t, len(adminPrivs), len(userPrivs))
}

func TestTokenCache_NegativeCaching(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	cache := auth.NewTokenCache(db)
	nonExistentTokenID := "nonexistent"

	// First call - should hit database and cache the "not found"
	_, _, err := cache.GetStatus(nonExistentTokenID)
	require.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)

	// Second call - should return from cache without hitting DB
	// We can verify this by checking error is still the same
	_, _, err = cache.GetStatus(nonExistentTokenID)
	require.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err, "Should return cached not found error")

	// Now create the token in DB
	token := models.APIToken{
		TokenID: nonExistentTokenID,
		UserID:  1,
		RoleID:  2,
		TTL:     3600,
		Status:  models.TokenStatusActive,
	}
	err = db.Create(&token).Error
	require.NoError(t, err)

	// Should still return cached "not found" until invalidated
	_, _, err = cache.GetStatus(nonExistentTokenID)
	require.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err, "Should still return cached not found")

	// Invalidate cache
	cache.Invalidate(nonExistentTokenID)

	// Now should find the token
	status, privileges, err := cache.GetStatus(nonExistentTokenID)
	require.NoError(t, err)
	assert.Equal(t, models.TokenStatusActive, status)
	assert.NotEmpty(t, privileges)
}

func TestTokenCache_NegativeCachingExpiration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	cache := auth.NewTokenCache(db)
	cache.SetTTL(300 * time.Millisecond)

	nonExistentTokenID := "temp_nonexistent"

	// First call - cache the "not found"
	_, _, err := cache.GetStatus(nonExistentTokenID)
	require.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)

	// Create token in DB
	token := models.APIToken{
		TokenID: nonExistentTokenID,
		UserID:  1,
		RoleID:  2,
		TTL:     3600,
		Status:  models.TokenStatusActive,
	}
	err = db.Create(&token).Error
	require.NoError(t, err)

	// Wait for cache to expire
	time.Sleep(500 * time.Millisecond)

	// Now should find the token (cache expired)
	status, privileges, err := cache.GetStatus(nonExistentTokenID)
	require.NoError(t, err)
	assert.Equal(t, models.TokenStatusActive, status)
	assert.NotEmpty(t, privileges)
}
