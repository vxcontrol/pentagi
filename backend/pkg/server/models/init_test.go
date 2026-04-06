package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetValidator(t *testing.T) {
	t.Parallel()

	v := GetValidator()
	require.NotNil(t, v)
}

func TestStrongPasswordValidator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		pw      string
		wantErr bool
	}{
		{"long password over 15 chars", "abcdefghijklmnop", false},
		{"exactly 16 chars", "abcdefghijklmnop", false},
		{"8 chars with all requirements", "Pass1!ab", false},
		{"short no special", "pass1", true},
		{"7 chars with requirements", "Pa1!abc", true},
		{"8 chars no number", "Pass!abc", true},
		{"8 chars no uppercase", "pass1!ab", true},
		{"8 chars no lowercase", "PASS1!AB", true},
		{"8 chars no special", "Pass1abc", true},
		{"empty password", "", true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := GetValidator().Var(tt.pw, "stpass")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEmailValidator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"valid email", "test@example.com", false},
		{"admin special case", "admin", false},
		{"email with subdomain", "user@mail.example.com", false},
		{"email too short", "a@b", true},
		{"empty email", "", true},
		{"no at sign", "testexample.com", true},
		{"no domain", "test@", true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := GetValidator().Var(tt.email, "vmail")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOAuthMinScopeValidator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		scope   string
		wantErr bool
	}{
		{"valid openid email", "openid email", false},
		{"valid with extra scopes", "openid email profile", false},
		{"valid case insensitive", "OpenID Email", false},
		{"missing openid", "email profile", true},
		{"missing email", "openid profile", true},
		{"empty scope", "", true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := GetValidator().Var(tt.scope, "oauth_min_scope")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSolidValidator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid lowercase", "hello", false},
		{"valid with numbers", "abc123", false},
		{"valid with underscore", "my_name", false},
		{"valid with dash", "my-name", false},
		{"invalid uppercase", "Hello", true},
		{"invalid spaces", "hello world", true},
		{"invalid special chars", "hello@world", true},
		{"empty string", "", true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := GetValidator().Var(tt.input, "solid")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSemverValidator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid two part", "1.0", false},
		{"valid three part", "1.2.3", false},
		{"valid zero", "0.0.0", false},
		{"invalid prefix v", "v1.0.0", true},
		{"invalid single", "1", true},
		{"invalid text", "abc", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := GetValidator().Var(tt.input, "semver")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSemverExValidator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid two part", "1.0", false},
		{"valid three part", "1.2.3", false},
		{"valid with v prefix", "v1.0.0", false},
		{"valid with prerelease", "1.2.3-beta", false},
		{"valid four part", "1.2.3.4", false},
		{"invalid text only", "abc", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := GetValidator().Var(tt.input, "semverex")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestScanFromJSON(t *testing.T) {
	t.Parallel()

	type testStruct struct {
		Name string `json:"name"`
	}

	t.Run("scan from string", func(t *testing.T) {
		t.Parallel()
		var out testStruct
		err := scanFromJSON(`{"name":"test"}`, &out)
		require.NoError(t, err)
		assert.Equal(t, "test", out.Name)
	})

	t.Run("scan from bytes", func(t *testing.T) {
		t.Parallel()
		var out testStruct
		err := scanFromJSON([]byte(`{"name":"hello"}`), &out)
		require.NoError(t, err)
		assert.Equal(t, "hello", out.Name)
	})

	t.Run("scan unsupported type", func(t *testing.T) {
		t.Parallel()
		var out testStruct
		err := scanFromJSON(12345, &out)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported type")
	})

	t.Run("scan invalid json string", func(t *testing.T) {
		t.Parallel()
		var out testStruct
		err := scanFromJSON("not-json", &out)
		assert.Error(t, err)
	})

	t.Run("scan invalid json bytes", func(t *testing.T) {
		t.Parallel()
		var out testStruct
		err := scanFromJSON([]byte("not-json"), &out)
		assert.Error(t, err)
	})
}
