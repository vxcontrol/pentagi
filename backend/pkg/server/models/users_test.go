package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserStatusValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		status  UserStatus
		wantErr bool
	}{
		{"valid created", UserStatusCreated, false},
		{"valid active", UserStatusActive, false},
		{"valid blocked", UserStatusBlocked, false},
		{"invalid empty", UserStatus(""), true},
		{"invalid unknown", UserStatus("unknown"), true},
		{"invalid suspended", UserStatus("suspended"), true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.status.Valid()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid UserStatus")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserStatusString(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "created", UserStatusCreated.String())
	assert.Equal(t, "active", UserStatusActive.String())
	assert.Equal(t, "blocked", UserStatusBlocked.String())
}

func TestUserTypeValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		userType UserType
		wantErr  bool
	}{
		{"valid local", UserTypeLocal, false},
		{"valid oauth", UserTypeOAuth, false},
		{"valid api", UserTypeAPI, false},
		{"invalid empty", UserType(""), true},
		{"invalid unknown", UserType("unknown"), true},
		{"invalid saml", UserType("saml"), true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.userType.Valid()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid UserType")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserTypeString(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "local", UserTypeLocal.String())
	assert.Equal(t, "oauth", UserTypeOAuth.String())
	assert.Equal(t, "api", UserTypeAPI.String())
}

func TestLoginValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		login   Login
		wantErr bool
	}{
		{
			name:    "valid login",
			login:   Login{Mail: "test@example.com", Password: "password123"},
			wantErr: false,
		},
		{
			name:    "valid admin mail",
			login:   Login{Mail: "admin", Password: "password123"},
			wantErr: false,
		},
		{
			name:    "empty mail",
			login:   Login{Mail: "", Password: "password123"},
			wantErr: true,
		},
		{
			name:    "empty password",
			login:   Login{Mail: "test@example.com", Password: ""},
			wantErr: true,
		},
		{
			name:    "password too short",
			login:   Login{Mail: "test@example.com", Password: "ab"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.login.Valid()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPasswordValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		pw      Password
		wantErr bool
	}{
		{
			name: "valid strong password with special chars",
			pw: Password{
				CurrentPassword: "OldPass1!abc",
				Password:        "NewPass1!abc",
				ConfirmPassword: "NewPass1!abc",
			},
			wantErr: false,
		},
		{
			name: "valid long password over 15 chars",
			pw: Password{
				CurrentPassword: "oldpasswordvalue",
				Password:        "newpasswordvalue1",
				ConfirmPassword: "newpasswordvalue1",
			},
			wantErr: false,
		},
		{
			name: "confirm password mismatch",
			pw: Password{
				CurrentPassword: "OldPass1!abc",
				Password:        "NewPass1!abc",
				ConfirmPassword: "DifferentPass1!",
			},
			wantErr: true,
		},
		{
			name: "current equals new password",
			pw: Password{
				CurrentPassword: "SamePass1!abc",
				Password:        "SamePass1!abc",
				ConfirmPassword: "SamePass1!abc",
			},
			wantErr: true,
		},
		{
			name: "weak password no special chars",
			pw: Password{
				CurrentPassword: "OldPass1!abc",
				Password:        "newpass1",
				ConfirmPassword: "newpass1",
			},
			wantErr: true,
		},
		{
			name: "empty current password",
			pw: Password{
				CurrentPassword: "",
				Password:        "NewPass1!abc",
				ConfirmPassword: "NewPass1!abc",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.pw.Valid()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserValid(t *testing.T) {
	t.Parallel()

	validUser := User{
		ID:     1,
		Hash:   "abcdef1234567890abcdef1234567890",
		Type:   UserTypeLocal,
		Mail:   "test@example.com",
		Status: UserStatusActive,
		RoleID: RoleUser,
	}

	t.Run("valid user", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, validUser.Valid())
	})

	t.Run("missing mail", func(t *testing.T) {
		t.Parallel()
		u := validUser
		u.Mail = ""
		assert.Error(t, u.Valid())
	})

	t.Run("invalid user type", func(t *testing.T) {
		t.Parallel()
		u := validUser
		u.Type = UserType("invalid")
		assert.Error(t, u.Valid())
	})

	t.Run("invalid user status", func(t *testing.T) {
		t.Parallel()
		u := validUser
		u.Status = UserStatus("invalid")
		assert.Error(t, u.Valid())
	})

	t.Run("invalid hash length", func(t *testing.T) {
		t.Parallel()
		u := validUser
		u.Hash = "tooshort"
		assert.Error(t, u.Valid())
	})
}

func TestUserTableName(t *testing.T) {
	t.Parallel()
	u := &User{}
	assert.Equal(t, "users", u.TableName())
}

func TestUserPasswordValid(t *testing.T) {
	t.Parallel()

	t.Run("valid user password", func(t *testing.T) {
		t.Parallel()
		up := UserPassword{
			Password: "somepassword",
			User: User{
				ID:     1,
				Hash:   "abcdef1234567890abcdef1234567890",
				Type:   UserTypeLocal,
				Mail:   "test@example.com",
				Status: UserStatusActive,
				RoleID: RoleUser,
			},
		}
		assert.NoError(t, up.Valid())
	})

	t.Run("empty password", func(t *testing.T) {
		t.Parallel()
		up := UserPassword{
			Password: "",
			User: User{
				ID:     1,
				Hash:   "abcdef1234567890abcdef1234567890",
				Type:   UserTypeLocal,
				Mail:   "test@example.com",
				Status: UserStatusActive,
				RoleID: RoleUser,
			},
		}
		assert.Error(t, up.Valid())
	})

	t.Run("invalid user in user password", func(t *testing.T) {
		t.Parallel()
		up := UserPassword{
			Password: "somepassword",
			User: User{
				Mail: "",
			},
		}
		assert.Error(t, up.Valid())
	})
}

func TestUserPasswordTableName(t *testing.T) {
	t.Parallel()
	up := &UserPassword{}
	assert.Equal(t, "users", up.TableName())
}

func TestLoginTableName(t *testing.T) {
	t.Parallel()
	l := &Login{}
	assert.Equal(t, "users", l.TableName())
}

func TestPasswordTableName(t *testing.T) {
	t.Parallel()
	p := &Password{}
	assert.Equal(t, "users", p.TableName())
}

func TestUserPreferencesOptionsValueScan(t *testing.T) {
	t.Parallel()

	t.Run("value and scan round trip", func(t *testing.T) {
		t.Parallel()
		original := UserPreferencesOptions{FavoriteFlows: []int64{1, 2, 3}}
		val, err := original.Value()
		require.NoError(t, err)

		var scanned UserPreferencesOptions
		switch v := val.(type) {
		case string:
			err = scanned.Scan([]byte(v))
		case []byte:
			err = scanned.Scan(v)
		default:
			t.Fatalf("unexpected Value() type: %T", val)
		}
		require.NoError(t, err)
		assert.Equal(t, original.FavoriteFlows, scanned.FavoriteFlows)
	})

	t.Run("scan nil value", func(t *testing.T) {
		t.Parallel()
		var upo UserPreferencesOptions
		err := upo.Scan(nil)
		require.NoError(t, err)
		assert.Equal(t, []int64{}, upo.FavoriteFlows)
	})

	t.Run("scan unsupported type", func(t *testing.T) {
		t.Parallel()
		var upo UserPreferencesOptions
		err := upo.Scan(12345)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expected []byte")
	})

	t.Run("scan invalid json", func(t *testing.T) {
		t.Parallel()
		var upo UserPreferencesOptions
		err := upo.Scan([]byte("not json"))
		assert.Error(t, err)
	})

	t.Run("value with empty flows", func(t *testing.T) {
		t.Parallel()
		upo := UserPreferencesOptions{FavoriteFlows: []int64{}}
		val, err := upo.Value()
		require.NoError(t, err)
		var valStr string
		switch v := val.(type) {
		case string:
			valStr = v
		case []byte:
			valStr = string(v)
		}
		assert.Contains(t, valStr, "favoriteFlows")
	})
}

func TestUserPreferencesValid(t *testing.T) {
	t.Parallel()

	t.Run("valid preferences", func(t *testing.T) {
		t.Parallel()
		up := UserPreferences{UserID: 1}
		assert.NoError(t, up.Valid())
	})

	t.Run("zero user id", func(t *testing.T) {
		t.Parallel()
		up := UserPreferences{UserID: 0}
		err := up.Valid()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user_id")
	})
}

func TestUserPreferencesTableName(t *testing.T) {
	t.Parallel()
	up := &UserPreferences{}
	assert.Equal(t, "user_preferences", up.TableName())
}

func TestNewUserPreferences(t *testing.T) {
	t.Parallel()

	up := NewUserPreferences(42)
	assert.Equal(t, uint64(42), up.UserID)
	assert.NotNil(t, up.Preferences.FavoriteFlows)
	assert.Empty(t, up.Preferences.FavoriteFlows)
}

func TestAuthCallbackValid(t *testing.T) {
	t.Parallel()

	// JWT token with 3 dot-separated base64 segments (header.payload.signature)
	validJWT := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.POstGetfAytaZS82wHcjoTyoqhMyxXiWdR7Nn7A29DNSl0EiXLdwJ6xC6AfgZWF1bOsS_TuYI3OG85AmiExREkrS6tDfTQ2B3WXlrr-wp5AokiRbz3_oB4OxG-W9KcEEbDRcZc0nH3L7LzYptiy1PtAylQGxHTWZXtGz4ht0bAecBgmpdgXMguEIcoqPJ1n3pIWk_dUZegpqx0Lka21H6XxUTxiy8OcaarA8zdnPUnV6AmNP3ecFawIFYdvJB_cm-GvpCSbr8G8y_Mllj8f4x9nBH8pQux89_6gUY618iYv7tuPWBFfEbLxtF2pZS6YC1aSfLQxaOoaBSTNRg"

	t.Run("valid callback", func(t *testing.T) {
		t.Parallel()
		ac := AuthCallback{
			Code:    "auth-code-123",
			IdToken: validJWT,
			Scope:   "openid email profile",
			State:   "random-state-value",
		}
		assert.NoError(t, ac.Valid())
	})

	t.Run("missing code", func(t *testing.T) {
		t.Parallel()
		ac := AuthCallback{
			Code:    "",
			IdToken: validJWT,
			Scope:   "openid email",
			State:   "state123",
		}
		assert.Error(t, ac.Valid())
	})

	t.Run("scope missing openid", func(t *testing.T) {
		t.Parallel()
		ac := AuthCallback{
			Code:    "code123",
			IdToken: validJWT,
			Scope:   "email profile",
			State:   "state123",
		}
		assert.Error(t, ac.Valid())
	})

	t.Run("invalid id token not jwt", func(t *testing.T) {
		t.Parallel()
		ac := AuthCallback{
			Code:    "code123",
			IdToken: "not-a-jwt",
			Scope:   "openid email",
			State:   "state123",
		}
		assert.Error(t, ac.Valid())
	})
}

func TestUserRoleValid(t *testing.T) {
	t.Parallel()

	validUserForRole := User{
		Hash:   "abcdef1234567890abcdef1234567890",
		Type:   UserTypeLocal,
		Mail:   "test@example.com",
		Status: UserStatusActive,
		RoleID: RoleUser,
	}

	t.Run("valid user role", func(t *testing.T) {
		t.Parallel()
		ur := UserRole{
			Role: Role{ID: 1, Name: "admin"},
			User: validUserForRole,
		}
		assert.NoError(t, ur.Valid())
	})

	t.Run("invalid role", func(t *testing.T) {
		t.Parallel()
		ur := UserRole{
			Role: Role{Name: ""},
			User: validUserForRole,
		}
		assert.Error(t, ur.Valid())
	})

	t.Run("invalid user", func(t *testing.T) {
		t.Parallel()
		ur := UserRole{
			Role: Role{ID: 1, Name: "admin"},
			User: User{Mail: ""},
		}
		assert.Error(t, ur.Valid())
	})
}

func TestUserRolePrivilegesValid(t *testing.T) {
	t.Parallel()

	validUserForRole := User{
		Hash:   "abcdef1234567890abcdef1234567890",
		Type:   UserTypeLocal,
		Mail:   "test@example.com",
		Status: UserStatusActive,
		RoleID: RoleUser,
	}

	t.Run("valid user role privileges", func(t *testing.T) {
		t.Parallel()
		urp := UserRolePrivileges{
			Role: RolePrivileges{
				Privileges: []Privilege{{Name: "read"}},
				Role:       Role{ID: 1, Name: "admin"},
			},
			User: validUserForRole,
		}
		assert.NoError(t, urp.Valid())
	})

	t.Run("invalid role privileges", func(t *testing.T) {
		t.Parallel()
		urp := UserRolePrivileges{
			Role: RolePrivileges{
				Privileges: []Privilege{{Name: ""}},
				Role:       Role{ID: 1, Name: "admin"},
			},
			User: validUserForRole,
		}
		assert.Error(t, urp.Valid())
	})

	t.Run("invalid user", func(t *testing.T) {
		t.Parallel()
		urp := UserRolePrivileges{
			Role: RolePrivileges{
				Privileges: []Privilege{{Name: "read"}},
				Role:       Role{ID: 1, Name: "admin"},
			},
			User: User{Mail: ""},
		}
		assert.Error(t, urp.Valid())
	})
}
