package database

import "time"

type APITokenWithSecret struct {
	ApiToken
	Token string `json:"token"`
}

// IsAPITokenExpired ignores the token status: a revoked token past its TTL is
// still revoked, so callers gate on the status themselves.
func IsAPITokenExpired(createdAt time.Time, ttlSeconds int64) bool {
	return createdAt.Add(time.Duration(ttlSeconds) * time.Second).Before(time.Now())
}
