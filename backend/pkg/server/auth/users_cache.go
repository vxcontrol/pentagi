package auth

import (
	"sync"
	"time"

	"pentagi/pkg/server/models"

	"github.com/jinzhu/gorm"
)

// userCacheEntry represents a cached user status entry
type userCacheEntry struct {
	hash       string
	status     models.UserStatus
	generation uint64
	notFound   bool // negative caching
	expiresAt  time.Time
}

// UserInfo is what an authenticating request needs to know about a user.
type UserInfo struct {
	Hash       string
	Status     models.UserStatus
	Generation uint64
}

// UserCache provides caching for user hash lookups
type UserCache struct {
	cache sync.Map
	ttl   time.Duration
	db    *gorm.DB

	storeMu sync.Mutex
	epoch   uint64
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
	info, err := uc.GetUser(userID)
	if err != nil {
		return "", "", err
	}
	return info.Hash, info.Status, nil
}

// GetUser retrieves the user record an authenticating request checks against,
// from cache or database.
func (uc *UserCache) GetUser(userID uint64) (UserInfo, error) {
	if entry, ok := uc.cache.Load(userID); ok {
		cached := entry.(userCacheEntry)
		if time.Now().Before(cached.expiresAt) {
			if cached.notFound {
				return UserInfo{}, gorm.ErrRecordNotFound
			}
			return UserInfo{Hash: cached.hash, Status: cached.status, Generation: cached.generation}, nil
		}
		uc.cache.Delete(userID)
	}

	uc.storeMu.Lock()
	epoch := uc.epoch
	uc.storeMu.Unlock()

	var user models.User
	if err := uc.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			uc.storeIfCurrent(epoch, userID, userCacheEntry{
				notFound:  true,
				expiresAt: time.Now().Add(uc.ttl),
			})
			return UserInfo{}, gorm.ErrRecordNotFound
		}
		return UserInfo{}, err
	}

	uc.storeIfCurrent(epoch, userID, userCacheEntry{
		hash:       user.Hash,
		status:     user.Status,
		generation: user.SessionGeneration,
		notFound:   false,
		expiresAt:  time.Now().Add(uc.ttl),
	})

	return UserInfo{Hash: user.Hash, Status: user.Status, Generation: user.SessionGeneration}, nil
}

func (uc *UserCache) storeIfCurrent(epoch, userID uint64, entry userCacheEntry) {
	uc.storeMu.Lock()
	defer uc.storeMu.Unlock()

	if uc.epoch == epoch {
		uc.cache.Store(userID, entry)
	}
}

// Invalidate removes a specific user from cache
func (uc *UserCache) Invalidate(userID uint64) {
	uc.storeMu.Lock()
	defer uc.storeMu.Unlock()

	uc.epoch++
	uc.cache.Delete(userID)
}

// InvalidateAll clears the entire cache
func (uc *UserCache) InvalidateAll() {
	uc.storeMu.Lock()
	defer uc.storeMu.Unlock()

	uc.epoch++
	uc.cache.Range(func(key, value any) bool {
		uc.cache.Delete(key)
		return true
	})
}
