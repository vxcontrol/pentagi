package auth

import (
	"sync"
	"time"

	"pentagi/pkg/server/models"

	"github.com/jinzhu/gorm"
)

// tokenCacheEntry represents a cached token status entry
type tokenCacheEntry struct {
	status     models.TokenStatus
	privileges []string
	notFound   bool // negative caching
	expiresAt  time.Time
}

// TokenCache provides caching for token status lookups
type TokenCache struct {
	cache sync.Map
	ttl   time.Duration
	db    *gorm.DB
}

// NewTokenCache creates a new token cache instance
func NewTokenCache(db *gorm.DB) *TokenCache {
	return &TokenCache{
		ttl: 5 * time.Minute,
		db:  db,
	}
}

// SetTTL sets the TTL for the token cache
func (tc *TokenCache) SetTTL(ttl time.Duration) {
	tc.ttl = ttl
}

// GetStatus retrieves token status and privileges from cache or database
func (tc *TokenCache) GetStatus(tokenID string) (models.TokenStatus, []string, error) {
	// check cache first
	if entry, ok := tc.cache.Load(tokenID); ok {
		cached := entry.(tokenCacheEntry)
		if time.Now().Before(cached.expiresAt) {
			// return cached "not found" error
			if cached.notFound {
				return "", nil, gorm.ErrRecordNotFound
			}
			return cached.status, cached.privileges, nil
		}
		// cache entry expired, remove it
		tc.cache.Delete(tokenID)
	}

	// load from database with role privileges
	var token models.APIToken
	if err := tc.db.Where("token_id = ? AND deleted_at IS NULL", tokenID).First(&token).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			// cache negative result (token not found)
			tc.cache.Store(tokenID, tokenCacheEntry{
				notFound:  true,
				expiresAt: time.Now().Add(tc.ttl),
			})
			return "", nil, gorm.ErrRecordNotFound
		}
		return "", nil, err
	}

	// load privileges for the token's role
	var privileges []models.Privilege
	if err := tc.db.Where("role_id = ?", token.RoleID).Find(&privileges).Error; err != nil {
		return "", nil, err
	}

	// extract privilege names
	privNames := make([]string, len(privileges))
	for i, priv := range privileges {
		privNames[i] = priv.Name
	}

	// always add automation privilege for API tokens
	privNames = append(privNames, PrivilegeAutomation)

	// update cache with positive result
	tc.cache.Store(tokenID, tokenCacheEntry{
		status:     token.Status,
		privileges: privNames,
		notFound:   false,
		expiresAt:  time.Now().Add(tc.ttl),
	})

	return token.Status, privNames, nil
}

// Invalidate removes a specific token from cache
func (tc *TokenCache) Invalidate(tokenID string) {
	tc.cache.Delete(tokenID)
}

// InvalidateUser removes all tokens for a specific user from cache
func (tc *TokenCache) InvalidateUser(userID uint64) {
	// load all tokens for this user
	var tokens []models.APIToken
	if err := tc.db.Where("user_id = ? AND deleted_at IS NULL", userID).Find(&tokens).Error; err != nil {
		return
	}

	// invalidate each token in cache
	for _, token := range tokens {
		tc.cache.Delete(token.TokenID)
	}
}

// InvalidateAll clears the entire cache
func (tc *TokenCache) InvalidateAll() {
	tc.cache.Range(func(key, value any) bool {
		tc.cache.Delete(key)
		return true
	})
}
