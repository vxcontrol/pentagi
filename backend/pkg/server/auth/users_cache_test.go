package auth_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"pentagi/pkg/server/auth"
	"pentagi/pkg/server/models"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUserTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open("sqlite3", ":memory:")
	require.NoError(t, err)

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

	time.Sleep(200 * time.Millisecond) // wait for database to be ready

	return db
}

func TestUserCache_GetUserHash(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	cache := auth.NewUserCache(db)

	// Insert test user
	user := models.User{
		ID:     1,
		Hash:   "test_hash_123",
		Mail:   "test@example.com",
		Name:   "Test User",
		Status: models.UserStatusActive,
		RoleID: 2,
	}
	err := db.Create(&user).Error
	require.NoError(t, err)

	// Test: Get user hash (should hit database)
	hash, status, err := cache.GetUserHash(1)
	require.NoError(t, err)
	assert.Equal(t, "test_hash_123", hash)
	assert.Equal(t, models.UserStatusActive, status)

	// Test: Get user hash again (should hit cache)
	hash, status, err = cache.GetUserHash(1)
	require.NoError(t, err)
	assert.Equal(t, "test_hash_123", hash)
	assert.Equal(t, models.UserStatusActive, status)

	// Test: Non-existent user
	_, _, err = cache.GetUserHash(999)
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestUserCache_Invalidate(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	cache := auth.NewUserCache(db)

	// Insert test user
	user := models.User{
		ID:     1,
		Hash:   "test_hash_456",
		Mail:   "test2@example.com",
		Name:   "Test User 2",
		Status: models.UserStatusActive,
		RoleID: 2,
	}
	err := db.Create(&user).Error
	require.NoError(t, err)

	// Get hash to populate cache
	hash, status, err := cache.GetUserHash(1)
	require.NoError(t, err)
	assert.Equal(t, "test_hash_456", hash)
	assert.Equal(t, models.UserStatusActive, status)

	// Update user in database
	db.Model(&user).Update("status", models.UserStatusBlocked)

	// Status should still be active (from cache)
	hash, status, err = cache.GetUserHash(1)
	require.NoError(t, err)
	assert.Equal(t, "test_hash_456", hash)
	assert.Equal(t, models.UserStatusActive, status)

	// Invalidate cache
	cache.Invalidate(1)

	// Status should now be blocked (from database)
	hash, status, err = cache.GetUserHash(1)
	require.NoError(t, err)
	assert.Equal(t, "test_hash_456", hash)
	assert.Equal(t, models.UserStatusBlocked, status)
}

func TestUserCache_Expiration(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	// Create cache with very short TTL for testing
	cache := auth.NewUserCache(db)
	cache.SetTTL(300 * time.Millisecond)

	// Insert test user
	user := models.User{
		ID:     1,
		Hash:   "test_hash_789",
		Mail:   "test3@example.com",
		Name:   "Test User 3",
		Status: models.UserStatusActive,
		RoleID: 2,
	}
	err := db.Create(&user).Error
	require.NoError(t, err)

	// Get hash to populate cache
	hash, status, err := cache.GetUserHash(1)
	require.NoError(t, err)
	assert.Equal(t, "test_hash_789", hash)
	assert.Equal(t, models.UserStatusActive, status)

	// Update user in database
	db.Model(&user).Update("status", models.UserStatusBlocked)

	// Wait for cache to expire
	time.Sleep(500 * time.Millisecond)

	// Status should now be blocked (cache expired, reading from DB)
	hash, status, err = cache.GetUserHash(1)
	require.NoError(t, err)
	assert.Equal(t, "test_hash_789", hash)
	assert.Equal(t, models.UserStatusBlocked, status)
}

func TestUserCache_UserStatuses(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	cache := auth.NewUserCache(db)

	testCases := []struct {
		name           string
		userStatus     models.UserStatus
		expectedStatus models.UserStatus
	}{
		{
			name:           "active user",
			userStatus:     models.UserStatusActive,
			expectedStatus: models.UserStatusActive,
		},
		{
			name:           "blocked user",
			userStatus:     models.UserStatusBlocked,
			expectedStatus: models.UserStatusBlocked,
		},
		{
			name:           "created user",
			userStatus:     models.UserStatusCreated,
			expectedStatus: models.UserStatusCreated,
		},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			user := models.User{
				ID:     uint64(i + 1),
				Hash:   "hash_" + tc.name,
				Mail:   tc.name + "@example.com",
				Name:   tc.name,
				Status: tc.userStatus,
				RoleID: 2,
			}
			err := db.Create(&user).Error
			require.NoError(t, err)

			hash, status, err := cache.GetUserHash(user.ID)
			require.NoError(t, err)
			assert.Equal(t, user.Hash, hash)
			assert.Equal(t, tc.expectedStatus, status)
		})
	}
}

func TestUserCache_DeletedUser(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	cache := auth.NewUserCache(db)

	// Insert test user
	user := models.User{
		ID:     1,
		Hash:   "deleted_hash",
		Mail:   "deleted@example.com",
		Name:   "Deleted User",
		Status: models.UserStatusActive,
		RoleID: 2,
	}
	err := db.Create(&user).Error
	require.NoError(t, err)

	// Get hash to populate cache
	hash, status, err := cache.GetUserHash(1)
	require.NoError(t, err)
	assert.Equal(t, "deleted_hash", hash)
	assert.Equal(t, models.UserStatusActive, status)

	// Soft delete user
	db.Delete(&user)

	// Invalidate cache
	cache.Invalidate(1)

	// Should return error for deleted user
	_, _, err = cache.GetUserHash(1)
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestUserCache_ConcurrentAccess(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	cache := auth.NewUserCache(db)

	// Insert test users
	for i := 1; i <= 10; i++ {
		user := models.User{
			ID:     uint64(i),
			Hash:   fmt.Sprintf("concurrent_hash_%d", i),
			Mail:   fmt.Sprintf("concurrent%d@example.com", i),
			Name:   "Concurrent User",
			Status: models.UserStatusActive,
			RoleID: 2,
		}
		err := db.Create(&user).Error
		require.NoError(t, err)
	}

	// warm up cache
	for i := range 10 {
		_, _, err := cache.GetUserHash(uint64(i%10 + 1))
		require.NoError(t, err)
	}

	var wg sync.WaitGroup
	errors := make(chan error, 100)

	// Concurrent reads
	for i := range 10 {
		wg.Add(1)
		go func(userID uint64) {
			defer wg.Done()
			for range 10 {
				_, _, err := cache.GetUserHash(userID)
				if err != nil {
					errors <- err
				}
			}
		}(uint64(i%10 + 1))
	}

	// Concurrent invalidations
	for i := range 5 {
		wg.Add(1)
		go func(userID uint64) {
			defer wg.Done()
			for range 5 {
				cache.Invalidate(userID)
				time.Sleep(10 * time.Millisecond)
			}
		}(uint64(i%10 + 1))
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent access error: %v", err)
	}
}

func TestUserCache_InvalidateAll(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	cache := auth.NewUserCache(db)

	// Insert multiple users
	for i := 1; i <= 5; i++ {
		user := models.User{
			ID:     uint64(i),
			Hash:   fmt.Sprintf("invalidate_all_%d", i),
			Mail:   fmt.Sprintf("all%d@example.com", i),
			Name:   fmt.Sprintf("User %d", i),
			Status: models.UserStatusActive,
			RoleID: 2,
		}
		err := db.Create(&user).Error
		require.NoError(t, err)
	}

	// Populate cache
	for i := 1; i <= 5; i++ {
		_, _, err := cache.GetUserHash(uint64(i))
		require.NoError(t, err)
	}

	// Update all users in database
	db.Model(&models.User{}).Where("id > 0").Update("status", models.UserStatusBlocked)

	// Invalidate all
	cache.InvalidateAll()

	// All users should now show blocked status
	for i := 1; i <= 5; i++ {
		_, status, err := cache.GetUserHash(uint64(i))
		require.NoError(t, err)
		assert.Equal(t, models.UserStatusBlocked, status)
	}
}

func TestUserCache_NegativeCaching(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	cache := auth.NewUserCache(db)
	nonExistentUserID := uint64(9999)

	// First call - should hit database and cache the "not found"
	_, _, err := cache.GetUserHash(nonExistentUserID)
	require.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)

	// Second call - should return from cache without hitting DB
	_, _, err = cache.GetUserHash(nonExistentUserID)
	require.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err, "Should return cached not found error")

	// Now create the user in DB
	user := models.User{
		ID:     nonExistentUserID,
		Hash:   "new_user_hash",
		Mail:   "new@example.com",
		Name:   "New User",
		Status: models.UserStatusActive,
		RoleID: 2,
	}
	err = db.Create(&user).Error
	require.NoError(t, err)

	// Should still return cached "not found" until invalidated
	_, _, err = cache.GetUserHash(nonExistentUserID)
	require.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err, "Should still return cached not found")

	// Invalidate cache
	cache.Invalidate(nonExistentUserID)

	// Now should find the user
	hash, status, err := cache.GetUserHash(nonExistentUserID)
	require.NoError(t, err)
	assert.Equal(t, "new_user_hash", hash)
	assert.Equal(t, models.UserStatusActive, status)
}

func TestUserCache_NegativeCachingExpiration(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	cache := auth.NewUserCache(db)
	cache.SetTTL(300 * time.Millisecond)

	nonExistentUserID := uint64(8888)

	// First call - cache the "not found"
	_, _, err := cache.GetUserHash(nonExistentUserID)
	require.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)

	// Create user in DB
	user := models.User{
		ID:     nonExistentUserID,
		Hash:   "temp_user_hash",
		Mail:   "temp@example.com",
		Name:   "Temp User",
		Status: models.UserStatusActive,
		RoleID: 2,
	}
	err = db.Create(&user).Error
	require.NoError(t, err)

	// Wait for cache to expire
	time.Sleep(500 * time.Millisecond)

	// Now should find the user (cache expired)
	hash, status, err := cache.GetUserHash(nonExistentUserID)
	require.NoError(t, err)
	assert.Equal(t, "temp_user_hash", hash)
	assert.Equal(t, models.UserStatusActive, status)
}
