package auth_test

import (
	"pentagi/pkg/server/auth"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeJWTSigningKey(t *testing.T) {
	salt1 := "test_salt_1"
	salt2 := "test_salt_2"

	// Test that key is generated
	key1 := auth.MakeJWTSigningKey(salt1)
	assert.NotNil(t, key1)
	assert.Len(t, key1, 32, "JWT signing key should be 32 bytes (256 bits)")

	// Test that same salt produces same key (cached)
	key1Again := auth.MakeJWTSigningKey(salt1)
	assert.Equal(t, key1, key1Again, "Same salt should produce same key from cache")

	// Test that different salts produce different keys
	key2 := auth.MakeJWTSigningKey(salt2)
	assert.NotEqual(t, key1, key2, "Different salts should produce different keys")
	assert.Len(t, key2, 32, "JWT signing key should be 32 bytes (256 bits)")

	// Verify consistency for salt2
	key2Again := auth.MakeJWTSigningKey(salt2)
	assert.Equal(t, key2, key2Again, "Same salt should produce same key from cache")
}

func TestMakeCookieStoreKey(t *testing.T) {
	salt := "test_salt"

	// Test that keys are generated
	keys := auth.MakeCookieStoreKey(salt)
	assert.NotNil(t, keys)
	assert.Len(t, keys, 2, "Should return auth and encryption keys")

	// Test that auth key is 64 bytes (SHA512)
	assert.Len(t, keys[0], 64, "Auth key should be 64 bytes")

	// Test that encryption key is 32 bytes (SHA256)
	assert.Len(t, keys[1], 32, "Encryption key should be 32 bytes")

	// Test consistency
	keysAgain := auth.MakeCookieStoreKey(salt)
	assert.Equal(t, keys, keysAgain, "Same salt should produce same keys")
}

func TestMakeJWTSigningKeyDifferentFromCookieKey(t *testing.T) {
	salt := "test_salt"

	jwtKey := auth.MakeJWTSigningKey(salt)
	cookieKeys := auth.MakeCookieStoreKey(salt)

	// JWT signing key should be different from both cookie keys
	assert.NotEqual(t, jwtKey, cookieKeys[0], "JWT key should differ from cookie auth key")
	assert.NotEqual(t, jwtKey, cookieKeys[1], "JWT key should differ from cookie encryption key")
}
