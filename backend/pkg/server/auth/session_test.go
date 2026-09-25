package auth

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Literals from an independent PBKDF2: a changed derivation signs out every issued token and cookie.
func TestSession_MakeJWTSigningKey_DerivesAStableKeyPerSaltApartFromTheCookieKeys(t *testing.T) {
	for range 2 {
		assert.Equal(t, "30c8728cb80e76b834425f8ef8afc62557dd94f7906a28c06c317fed9a8b4de5",
			hex.EncodeToString(MakeJWTSigningKey("kat_salt")), "derived first, then served from the cache")
	}

	other := MakeJWTSigningKey("another_salt")
	assert.Len(t, other, 32)
	assert.NotEqual(t, MakeJWTSigningKey("kat_salt"), other)

	cookieKeys := MakeCookieStoreKey("kat_salt")
	assert.NotEqual(t, MakeJWTSigningKey("kat_salt"), cookieKeys[0])
	assert.NotEqual(t, MakeJWTSigningKey("kat_salt"), cookieKeys[1])
}

func TestSession_MakeCookieStoreKey_DerivesAStableAuthAndEncryptionKey(t *testing.T) {
	for range 2 {
		keys := MakeCookieStoreKey("kat_salt")
		require.Len(t, keys, 2)
		require.Len(t, keys[0], 64)
		require.Len(t, keys[1], 32)
		assert.Equal(t, "82db383af22e5e9744541f50404e1b77", hex.EncodeToString(keys[0][:16]))
		assert.Equal(t, "3d8abedc0cb244d2e3ad59e7d5f125d8", hex.EncodeToString(keys[1][:16]))
	}
}
