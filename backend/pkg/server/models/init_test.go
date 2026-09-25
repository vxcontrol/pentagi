package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Rows are keyed by the tag init registers; a refused value names that tag.
func TestInit_GetValidator_AppliesTheRegisteredTags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		tag   string
		value string
		valid bool
	}{
		{"a strong password of eight composed characters", "stpass", "Pass1!ab", true},
		{"a strong password mixing latin and cyrillic", "stpass", "аБвгZq1!", true},
		{"fifteen cyrillic characters are too few for a strong password", "stpass", "абвгдеёжзийклмн", false},
		{"cyrillic letters do not stand for the latin classes", "stpass", "аБв1!где", false},

		{"a mail address", "vmail", "test@example.com", true},
		{"the admin login in place of a mail", "vmail", "admin", true},
		{"a mail too short to be an address", "vmail", "a@b", false},
		{"a mail without an at sign", "vmail", "testexample.com", false},

		{"a real address", "realemail", "test@example.com", true},
		{"a real address on a subdomain", "realemail", "user@mail.example.com", true},
		{"a real address in mixed case", "realemail", "John.Doe@Example.com", true},
		{"a real address under a long tld", "realemail", "user@example.cloud", true},
		{"the admin login is not a real address", "realemail", "admin", false},
		{"a uuid is not a real address", "realemail", "550e8400-e29b-41d4-a716-446655440000", false},
		{"an address too short", "realemail", "a@b", false},
		{"no address", "realemail", "", false},
		{"an address without an at sign", "realemail", "testexample.com", false},
		{"an address without a domain", "realemail", "test@", false},

		{"the openid and email scopes", "oauth_min_scope", "openid email", true},
		{"the minimum scopes among others", "oauth_min_scope", "openid email profile", true},
		{"the minimum scopes in another case", "oauth_min_scope", "OpenID Email", true},
		{"scopes without openid", "oauth_min_scope", "email profile", false},
		{"scopes without email", "oauth_min_scope", "openid profile", false},
		{"no scopes", "oauth_min_scope", "", false},

		{"a lowercase word", "solid", "hello", true},
		{"letters and digits", "solid", "abc123", true},
		{"an underscore", "solid", "my_name", true},
		{"a dash", "solid", "my-name", true},
		{"an uppercase letter", "solid", "Hello", false},
		{"a space", "solid", "hello world", false},
		{"another symbol", "solid", "hello@world", false},
		{"an empty word", "solid", "", false},

		{"a two-part version", "semver", "1.0", true},
		{"a three-part version", "semver", "1.2.3", true},
		{"a zero version", "semver", "0.0.0", true},
		{"a version with a v prefix", "semver", "v1.0.0", false},
		{"a one-part version", "semver", "1", false},
		{"text for a version", "semver", "abc", false},
		{"no version", "semver", "", false},

		{"a two-part extended version", "semverex", "1.0", true},
		{"a three-part extended version", "semverex", "1.2.3", true},
		{"an extended version with a v prefix", "semverex", "v1.0.0", true},
		{"an extended version with a prerelease", "semverex", "1.2.3-beta", true},
		{"a four-part extended version", "semverex", "1.2.3.4", true},
		{"text for an extended version", "semverex", "abc", false},
		{"no extended version", "semverex", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			want := ":" + tt.tag
			if tt.valid {
				want = ""
			}

			assert.Equal(t, want, modelsRefusal(GetValidator().Var(tt.value, tt.tag)))
		})
	}
}

func TestInit_ScanFromJSON_DecodesAStringOrBytes(t *testing.T) {
	t.Parallel()

	type scanned struct {
		Name string `json:"name"`
	}

	for _, tc := range []struct {
		input any
		want  string
	}{
		{`{"name":"test"}`, "test"},
		{[]byte(`{"name":"hello"}`), "hello"},
	} {
		var out scanned
		require.NoError(t, scanFromJSON(tc.input, &out), "%T", tc.input)
		assert.Equal(t, tc.want, out.Name, "%T", tc.input)
	}

	for _, input := range []any{"not-json", []byte("not-json")} {
		var syntax *json.SyntaxError
		assert.ErrorAs(t, scanFromJSON(input, &scanned{}), &syntax, "%T", input)
	}

	assert.EqualError(t, scanFromJSON(12345, &scanned{}), "unsupported type of input value to scan")
}
