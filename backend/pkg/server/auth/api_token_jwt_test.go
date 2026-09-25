package auth

import (
	"testing"
	"time"

	"pentagi/pkg/server/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPITokenJWT_ValidateAPIToken_AcceptsOnlyALiveTokenSignedUnderTheSalt(t *testing.T) {
	const salt = "test_salt"

	sign := func(method jwt.SigningMethod, key any, expiresIn time.Duration) string {
		claims := models.APITokenClaims{
			TokenID: "abc123xyz9",
			RID:     2,
			UID:     1,
			UHASH:   "testhash",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
				IssuedAt:  jwt.NewNumericDate(time.Now().Add(-time.Hour)),
				Subject:   "api_token",
			},
		}
		token, err := jwt.NewWithClaims(method, claims).SignedString(key)
		require.NoError(t, err)
		return token
	}

	minted, err := MakeAPIToken(salt, MakeAPITokenClaims("abc123xyz9", "testhash", 1, 2, 3600))
	require.NoError(t, err)

	cases := []struct {
		name    string
		token   string
		wantErr string
	}{
		{"a token minted by MakeAPIToken", minted, ""},
		{"a token expired a second ago", sign(jwt.SigningMethodHS256, MakeJWTSigningKey(salt), -time.Second),
			"token is either expired or not active yet"},
		{"a token signed under another salt", sign(jwt.SigningMethodHS256, MakeJWTSigningKey("wrong_salt"), time.Hour),
			"token invalid: token signature is invalid"},
		{"a token signed with alg none", sign(jwt.SigningMethodNone, jwt.UnsafeAllowNoneSignatureType, time.Hour),
			"unexpected signing method: none"},
		{"a malformed token", "not.a.valid.jwt.token", "token is malformed"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			claims, err := ValidateAPIToken(tc.token, salt)

			if tc.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
				assert.Nil(t, claims)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, "abc123xyz9", claims.TokenID)
			assert.Equal(t, uint64(2), claims.RID)
			assert.Equal(t, uint64(1), claims.UID)
			assert.Equal(t, "testhash", claims.UHASH)
			assert.Equal(t, "api_token", claims.Subject)
			require.NotNil(t, claims.IssuedAt)
			require.NotNil(t, claims.ExpiresAt)
			assert.InDelta(t, 3600, claims.ExpiresAt.Sub(claims.IssuedAt.Time).Seconds(), 1, "the ttl is in seconds")
		})
	}
}
