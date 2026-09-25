package auth

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAPITokenID_GenerateTokenID_ReturnsDistinctBase62IDsOfTenCharacters(t *testing.T) {
	seen := make(map[string]bool, 10000)

	for range 10000 {
		tokenID, err := GenerateTokenID()
		require.NoError(t, err)
		require.Len(t, tokenID, 10)

		for _, char := range tokenID {
			isBase62 := (char >= '0' && char <= '9') || (char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z')
			require.True(t, isBase62, "%q is not a base62 character", char)
		}

		require.False(t, seen[tokenID], "duplicate token ID %s", tokenID)
		seen[tokenID] = true
	}
}
