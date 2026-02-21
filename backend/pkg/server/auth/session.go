package auth

import (
	"crypto/sha512"
	"strings"
	"sync"

	"golang.org/x/crypto/pbkdf2"
)

var (
	cookieStoreKeys sync.Map // cache of cookie keys per salt
	jwtSigningKeys  sync.Map // cache of JWT signing keys per salt
)

const (
	pbkdf2Iterations = 210000 // OWASP 2023 recommendation
	jwtKeyLength     = 32     // 256 bits for HS256
	authKeyLength    = 64     // 512 bits for cookie auth key
	encKeyLength     = 32     // 256 bits for cookie encryption key
)

// MakeCookieStoreKey is function to generate auth and encryption keys for cookie store
func MakeCookieStoreKey(globalSalt string) [][]byte {
	// Check cache for existing keys
	if cached, ok := cookieStoreKeys.Load(globalSalt); ok {
		return cached.([][]byte)
	}

	// Generate new keys for this salt using PBKDF2
	password := []byte(strings.Join([]string{
		"a8d0abae36f749588f4393e6fc292690",
		globalSalt,
		"7c9be62adec5076970fa946e78f256e2",
	}, "|"))

	// Auth key (64 bytes) - using salt variant 1
	authSalt := []byte("pentagi.cookie.auth|" + globalSalt)
	authKey := pbkdf2.Key(password, authSalt, pbkdf2Iterations, authKeyLength, sha512.New)

	// Encryption key (32 bytes) - using salt variant 2
	encSalt := []byte("pentagi.cookie.enc|" + globalSalt)
	encKey := pbkdf2.Key(password, encSalt, pbkdf2Iterations, encKeyLength, sha512.New)

	newKeys := [][]byte{authKey, encKey}

	// Store in cache (LoadOrStore handles concurrent access)
	actual, _ := cookieStoreKeys.LoadOrStore(globalSalt, newKeys)
	return actual.([][]byte)
}

// MakeJWTSigningKey is function to generate signing key for JWT tokens
func MakeJWTSigningKey(globalSalt string) []byte {
	// Check cache for existing key
	if cached, ok := jwtSigningKeys.Load(globalSalt); ok {
		return cached.([]byte)
	}

	// Generate new key for this salt using PBKDF2
	password := []byte(strings.Join([]string{
		"4c1e9cb77df7f9a58fcc5f52d40af685",
		globalSalt,
		"09784e190148d13d48885aa47cf8a297",
	}, "|"))
	salt := []byte("pentagi.jwt.signing|" + globalSalt)
	newKey := pbkdf2.Key(password, salt, pbkdf2Iterations, jwtKeyLength, sha512.New)

	// Store in cache (LoadOrStore handles concurrent access)
	actual, _ := jwtSigningKeys.LoadOrStore(globalSalt, newKey)
	return actual.([]byte)
}
