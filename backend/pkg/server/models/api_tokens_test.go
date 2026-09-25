package models

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func apiTokensToken() APIToken {
	created := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	return APIToken{
		TokenID:   "abcdefghij",
		UserID:    1,
		RoleID:    1,
		TTL:       3600,
		Status:    TokenStatusActive,
		CreatedAt: created,
		UpdatedAt: created,
	}
}

// Rows are keyed by the type whose Valid they call.
func TestAPITokens_Valid_RefusesExactlyTheInvalidField(t *testing.T) {
	t.Parallel()

	unknownStatus := modelsWith(apiTokensToken(), func(at *APIToken) { at.Status = "invalid" })
	name := "my-token"

	modelsCheckValid(t, []modelsValidCase{
		{"an active token status", TokenStatusActive, ""},
		{"a revoked token status", TokenStatusRevoked, ""},
		{"an expired token status", TokenStatusExpired, ""},
		{"an empty token status", TokenStatus(""), "invalid TokenStatus: "},
		{"an unknown token status", TokenStatus("unknown"), "invalid TokenStatus: unknown"},

		{"a complete token", apiTokensToken(), ""},
		{"a token in an unknown status", unknownStatus, "invalid TokenStatus: invalid"},
		{"a token id of the wrong length", modelsWith(apiTokensToken(), func(at *APIToken) { at.TokenID = "short" }),
			"APIToken.TokenID:len"},
		{"a token living under a minute", modelsWith(apiTokensToken(), func(at *APIToken) { at.TTL = 10 }), "APIToken.TTL:min"},
		{"a token living over three years", modelsWith(apiTokensToken(), func(at *APIToken) { at.TTL = 94608001 }),
			"APIToken.TTL:max"},
		{"a token living exactly a minute", modelsWith(apiTokensToken(), func(at *APIToken) { at.TTL = 60 }), ""},
		{"a token living exactly three years", modelsWith(apiTokensToken(), func(at *APIToken) { at.TTL = 94608000 }), ""},

		{"a token request", CreateAPITokenRequest{TTL: 3600}, ""},
		{"a named token request", CreateAPITokenRequest{Name: &name, TTL: 3600}, ""},
		{"a token request under a minute", CreateAPITokenRequest{TTL: 30}, "CreateAPITokenRequest.TTL:min"},
		{"a token request with no lifetime", CreateAPITokenRequest{}, "CreateAPITokenRequest.TTL:required"},

		{"a token update revoking it", UpdateAPITokenRequest{Status: TokenStatusRevoked}, ""},
		{"an empty token update", UpdateAPITokenRequest{}, ""},
		{"a token update to an unknown status", UpdateAPITokenRequest{Status: "invalid"}, "invalid TokenStatus: invalid"},

		{"a token with its secret", APITokenWithSecret{APIToken: apiTokensToken(), Token: modelsJWT}, ""},
		{"an invalid token with its secret", APITokenWithSecret{APIToken: unknownStatus, Token: modelsJWT},
			"invalid TokenStatus: invalid"},
		{"a token whose secret is not a jwt", APITokenWithSecret{APIToken: apiTokensToken(), Token: "not-a-jwt"},
			"APITokenWithSecret.Token:jwt"},

		{"complete token claims", APITokenClaims{TokenID: "abcdefghij", RID: 1, UID: 1, UHASH: "somehash"}, ""},
		{"claims without a token id", APITokenClaims{UHASH: "somehash"}, "APITokenClaims.TokenID:required"},
		{"claims without a user hash", APITokenClaims{TokenID: "abcdefghij"}, "APITokenClaims.UHASH:required"},
		{"claims naming a user id past the range", APITokenClaims{TokenID: "abcdefghij", UID: 10001, UHASH: "somehash"},
			"APITokenClaims.UID:max"},
	})
}

func TestAPITokens_String_SpellsTheStoredValue(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		value fmt.Stringer
		want  string
	}{
		{TokenStatusActive, "active"},
		{TokenStatusRevoked, "revoked"},
		{TokenStatusExpired, "expired"},
	} {
		assert.Equal(t, tc.want, tc.value.String())
	}
}

func TestAPITokens_TableName_NamesTheMappedTable(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "api_tokens", (&APIToken{}).TableName())
}
