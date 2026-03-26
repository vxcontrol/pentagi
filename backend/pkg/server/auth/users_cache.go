package auth

import (
	"sync"
	"time"

	"pentagi/pkg/server/models"

	"github.com/jinzhu/gorm"
)

// userCacheEntry represents a cached user status entry
type userCacheEntry struct {
	hash      string
	status    models.UserStatus
	notFound  bool // negative caching
	expiresAt time.Time
}

// UserCache provides caching for user hash lookups
type UserCache struct {
	cache sync.Map
	ttl   time.Duration
	db    *gorm.DB
}

// NewUserCache creates a new user cache instance
func NewUserCache(db *gorm.DB) *UserCache {
	return &UserCache{
		ttl: 5 * time.Minute,
		db:  db,
	}
}

// SetTTL sets the TTL for the user cache
func (uc *UserCache) SetTTL(ttl time.Duration) {
	uc.ttl = ttl
}

// GetUserHash retrieves user hash and status from cache or database
func (uc *UserCache) GetUserHash(userID uint64) (string, models.UserStatus, error) {
	// check cache first
	if entry, ok := uc.cache.Load(userID); ok {
		cached := entry.(userCacheEntry)
		if time.Now().Before(cached.expiresAt) {
			// return cached "not found" error
			if cached.notFound {
				return "", "", gorm.ErrRecordNotFound
			}
			return cached.hash, cached.status, nil
		}
		// cache entry expired, remove it
		uc.cache.Delete(userID)
	}

	// load from database
	var user models.User
	if err := uc.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			// cache negative result (user not found)
			uc.cache.Store(userID, userCacheEntry{
				notFound:  true,
				expiresAt: time.Now().Add(uc.ttl),
			})
			return "", "", gorm.ErrRecordNotFound
		}
		return "", "", err
	}

	// update cache with positive result
	uc.cache.Store(userID, userCacheEntry{
		hash:      user.Hash,
		status:    user.Status,
		notFound:  false,
		expiresAt: time.Now().Add(uc.ttl),
	})

	return user.Hash, user.Status, nil
}

// Invalidate removes a specific user from cache
func (uc *UserCache) Invalidate(userID uint64) {
	uc.cache.Delete(userID)
}

// InvalidateAll clears the entire cache
func (uc *UserCache) InvalidateAll() {
	uc.cache.Range(func(key, value any) bool {
		uc.cache.Delete(key)
		return true
	})
}
