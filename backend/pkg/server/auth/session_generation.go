package auth

import (
	"pentagi/pkg/server/models"

	"github.com/jinzhu/gorm"
)

// RevokeSessions makes every cookie already issued to a user stop
// authenticating, and returns the generation they now have to carry. A caller
// that wants to keep its own session alive re-stamps its cookie with it.
//
// The cache has to be dropped in the same breath: it holds the generation for
// its whole TTL, so a revocation that leaves it in place takes five minutes to
// reach the middleware.
func RevokeSessions(db *gorm.DB, cache *UserCache, userID uint64) (uint64, error) {
	err := db.Model(&models.User{}).
		Where("id = ?", userID).
		UpdateColumn("session_generation", gorm.Expr("session_generation + 1")).
		Error
	if err != nil {
		return 0, err
	}

	if cache != nil {
		cache.Invalidate(userID)
	}

	var user models.User
	if err := db.Select("session_generation").Where("id = ?", userID).First(&user).Error; err != nil {
		return 0, err
	}

	return user.SessionGeneration, nil
}
