package auth_test

import (
	"testing"

	"pentagi/pkg/server/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateTokenID(t *testing.T) {
	// Test basic generation
	tokenID, err := auth.GenerateTokenID()
	require.NoError(t, err)
	assert.Len(t, tokenID, auth.TokenIDLength, "Token ID should have correct length")

	// Test that all characters are from base62 charset
	for _, char := range tokenID {
		assert.Contains(t, auth.Base62Chars, string(char), "Token ID should only contain base62 characters")
	}

	// Test uniqueness (generate multiple tokens and check they're different)
	tokens := make(map[string]bool)
	for i := 0; i < 100; i++ {
		token, err := auth.GenerateTokenID()
		require.NoError(t, err)
		assert.Len(t, token, auth.TokenIDLength)
		assert.False(t, tokens[token], "Generated tokens should be unique")
		tokens[token] = true
	}
}

func TestGenerateTokenIDFormat(t *testing.T) {
	// Test that token IDs match expected format
	tokenID, err := auth.GenerateTokenID()
	require.NoError(t, err)

	// Should be exactly 10 characters
	assert.Equal(t, 10, len(tokenID))

	// Should only contain alphanumeric characters
	for _, char := range tokenID {
		isValid := (char >= '0' && char <= '9') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= 'a' && char <= 'z')
		assert.True(t, isValid, "Character %c should be alphanumeric", char)
	}
}
