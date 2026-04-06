package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTokenStatusValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		status  TokenStatus
		wantErr bool
	}{
		{"valid active", TokenStatusActive, false},
		{"valid revoked", TokenStatusRevoked, false},
		{"valid expired", TokenStatusExpired, false},
		{"invalid empty", TokenStatus(""), true},
		{"invalid unknown", TokenStatus("unknown"), true},
		{"invalid deleted", TokenStatus("deleted"), true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.status.Valid()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid TokenStatus")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTokenStatusString(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "active", TokenStatusActive.String())
	assert.Equal(t, "revoked", TokenStatusRevoked.String())
	assert.Equal(t, "expired", TokenStatusExpired.String())
}

func TestAPITokenValid(t *testing.T) {
	t.Parallel()

	now := time.Now()
	validToken := APIToken{
		TokenID:   "abcdefghij",
		UserID:    1,
		RoleID:    1,
		TTL:       3600,
		Status:    TokenStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}

	t.Run("valid token", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, validToken.Valid())
	})

	t.Run("invalid status", func(t *testing.T) {
		t.Parallel()
		token := validToken
		token.Status = TokenStatus("invalid")
		assert.Error(t, token.Valid())
	})

	t.Run("token id wrong length", func(t *testing.T) {
		t.Parallel()
		token := validToken
		token.TokenID = "short"
		assert.Error(t, token.Valid())
	})

	t.Run("ttl too small", func(t *testing.T) {
		t.Parallel()
		token := validToken
		token.TTL = 10
		assert.Error(t, token.Valid())
	})

	t.Run("ttl too large", func(t *testing.T) {
		t.Parallel()
		token := validToken
		token.TTL = 94608001
		assert.Error(t, token.Valid())
	})

	t.Run("ttl minimum boundary", func(t *testing.T) {
		t.Parallel()
		token := validToken
		token.TTL = 60
		assert.NoError(t, token.Valid())
	})

	t.Run("ttl maximum boundary", func(t *testing.T) {
		t.Parallel()
		token := validToken
		token.TTL = 94608000
		assert.NoError(t, token.Valid())
	})
}

func TestAPITokenTableName(t *testing.T) {
	t.Parallel()
	at := &APIToken{}
	assert.Equal(t, "api_tokens", at.TableName())
}

func TestCreateAPITokenRequestValid(t *testing.T) {
	t.Parallel()

	t.Run("valid request", func(t *testing.T) {
		t.Parallel()
		req := CreateAPITokenRequest{TTL: 3600}
		assert.NoError(t, req.Valid())
	})

	t.Run("valid request with name", func(t *testing.T) {
		t.Parallel()
		name := "my-token"
		req := CreateAPITokenRequest{Name: &name, TTL: 3600}
		assert.NoError(t, req.Valid())
	})

	t.Run("ttl too small", func(t *testing.T) {
		t.Parallel()
		req := CreateAPITokenRequest{TTL: 30}
		assert.Error(t, req.Valid())
	})

	t.Run("zero ttl", func(t *testing.T) {
		t.Parallel()
		req := CreateAPITokenRequest{TTL: 0}
		assert.Error(t, req.Valid())
	})
}

func TestUpdateAPITokenRequestValid(t *testing.T) {
	t.Parallel()

	t.Run("valid update with status", func(t *testing.T) {
		t.Parallel()
		req := UpdateAPITokenRequest{Status: TokenStatusRevoked}
		assert.NoError(t, req.Valid())
	})

	t.Run("valid empty update", func(t *testing.T) {
		t.Parallel()
		req := UpdateAPITokenRequest{}
		assert.NoError(t, req.Valid())
	})

	t.Run("invalid status", func(t *testing.T) {
		t.Parallel()
		req := UpdateAPITokenRequest{Status: TokenStatus("invalid")}
		assert.Error(t, req.Valid())
	})
}

func TestAPITokenWithSecretValid(t *testing.T) {
	t.Parallel()

	now := time.Now()
	validJWT := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.POstGetfAytaZS82wHcjoTyoqhMyxXiWdR7Nn7A29DNSl0EiXLdwJ6xC6AfgZWF1bOsS_TuYI3OG85AmiExREkrS6tDfTQ2B3WXlrr-wp5AokiRbz3_oB4OxG-W9KcEEbDRcZc0nH3L7LzYptiy1PtAylQGxHTWZXtGz4ht0bAecBgmpdgXMguEIcoqPJ1n3pIWk_dUZegpqx0Lka21H6XxUTxiy8OcaarA8zdnPUnV6AmNP3ecFawIFYdvJB_cm-GvpCSbr8G8y_Mllj8f4x9nBH8pQux89_6gUY618iYv7tuPWBFfEbLxtF2pZS6YC1aSfLQxaOoaBSTNRg"

	baseToken := APIToken{
		TokenID:   "abcdefghij",
		UserID:    1,
		RoleID:    1,
		TTL:       3600,
		Status:    TokenStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}

	t.Run("valid token with secret", func(t *testing.T) {
		t.Parallel()
		ats := APITokenWithSecret{
			APIToken: baseToken,
			Token:    validJWT,
		}
		assert.NoError(t, ats.Valid())
	})

	t.Run("invalid embedded api token", func(t *testing.T) {
		t.Parallel()
		badToken := baseToken
		badToken.Status = TokenStatus("invalid")
		ats := APITokenWithSecret{
			APIToken: badToken,
			Token:    validJWT,
		}
		assert.Error(t, ats.Valid())
	})

	t.Run("invalid jwt token string", func(t *testing.T) {
		t.Parallel()
		ats := APITokenWithSecret{
			APIToken: baseToken,
			Token:    "not-a-jwt",
		}
		assert.Error(t, ats.Valid())
	})
}

func TestAPITokenClaimsValid(t *testing.T) {
	t.Parallel()

	t.Run("valid claims", func(t *testing.T) {
		t.Parallel()
		claims := APITokenClaims{
			TokenID: "abcdefghij",
			RID:     1,
			UID:     1,
			UHASH:   "somehash",
		}
		assert.NoError(t, claims.Valid())
	})

	t.Run("missing token id", func(t *testing.T) {
		t.Parallel()
		claims := APITokenClaims{
			TokenID: "",
			UHASH:   "somehash",
		}
		assert.Error(t, claims.Valid())
	})

	t.Run("missing uhash", func(t *testing.T) {
		t.Parallel()
		claims := APITokenClaims{
			TokenID: "abcdefghij",
			UHASH:   "",
		}
		assert.Error(t, claims.Valid())
	})

	t.Run("uid too large", func(t *testing.T) {
		t.Parallel()
		claims := APITokenClaims{
			TokenID: "abcdefghij",
			UID:     10001,
			UHASH:   "somehash",
		}
		assert.Error(t, claims.Valid())
	})
}
