package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func usersUser() User {
	return User{
		ID:     1,
		Hash:   "abcdef1234567890abcdef1234567890",
		Type:   UserTypeLocal,
		Mail:   "test@example.com",
		Status: UserStatusActive,
		RoleID: RoleUser,
	}
}

// Rows are keyed by the type whose Valid they call.
func TestUsers_Valid_RefusesExactlyTheInvalidField(t *testing.T) {
	t.Parallel()

	withoutMail := modelsWith(usersUser(), func(u *User) { u.Mail = "" })
	admin := Role{ID: 1, Name: "admin"}
	read := []Privilege{{Name: "read"}}

	modelsCheckValid(t, []modelsValidCase{
		{"a created user status", UserStatusCreated, ""},
		{"an active user status", UserStatusActive, ""},
		{"a blocked user status", UserStatusBlocked, ""},
		{"an empty user status", UserStatus(""), "invalid UserStatus: "},
		{"an unknown user status", UserStatus("unknown"), "invalid UserStatus: unknown"},

		{"a local user type", UserTypeLocal, ""},
		{"an oauth user type", UserTypeOAuth, ""},
		{"an api user type", UserTypeAPI, ""},
		{"an empty user type", UserType(""), "invalid UserType: "},
		{"an unknown user type", UserType("saml"), "invalid UserType: saml"},

		{"a complete user", usersUser(), ""},
		{"a user without a mail", withoutMail, "User.Mail:vmail"},
		{"a user of an unknown type", modelsWith(usersUser(), func(u *User) { u.Type = "invalid" }), "User.Type:valid"},
		{"a user in an unknown status", modelsWith(usersUser(), func(u *User) { u.Status = "invalid" }), "User.Status:valid"},
		{"a user with a short hash", modelsWith(usersUser(), func(u *User) { u.Hash = "tooshort" }), "User.Hash:len"},

		{"a user with a strong password", UserPassword{Password: "SecurePass123!", User: usersUser()}, ""},
		{"a user with no password", UserPassword{User: usersUser()}, "UserPassword.Password:stpass"},
		{"an invalid user behind a strong password", UserPassword{Password: "SecurePass123!", User: withoutMail}, "User.Mail:vmail"},
		{"a user password at the bcrypt limit", UserPassword{Password: strings.Repeat("a", MaxPasswordBytes), User: usersUser()}, ""},
		{"a user password over the bcrypt limit", UserPassword{Password: strings.Repeat("a", MaxPasswordBytes+1), User: usersUser()},
			"UserPassword.Password:passlen"},
		{"a weak user password", UserPassword{Password: "somepassword", User: usersUser()}, "UserPassword.Password:stpass"},
		{"a multibyte user password within the rune count but over the byte limit",
			UserPassword{Password: strings.Repeat("é", 40), User: usersUser()}, "UserPassword.Password:passlen"},

		{"a login with a mail", Login{Mail: "test@example.com", Password: "password123"}, ""},
		{"a login as admin", Login{Mail: "admin", Password: "password123"}, ""},
		{"a login without a mail", Login{Password: "password123"}, "Login.Mail:required"},
		{"a login without a password", Login{Mail: "test@example.com"}, "Login.Password:min"},
		{"a login with a password under four characters", Login{Mail: "test@example.com", Password: "abc"}, "Login.Password:min"},

		{"a strong new password", Password{CurrentPassword: "OldPass1!abc", Password: "NewPass1!abc", ConfirmPassword: "NewPass1!abc"}, ""},
		{"a long new password", Password{CurrentPassword: "oldpasswordvalue", Password: "newpasswordvalue1",
			ConfirmPassword: "newpasswordvalue1"}, ""},
		{"a confirmation that differs", Password{CurrentPassword: "OldPass1!abc", Password: "NewPass1!abc",
			ConfirmPassword: "DifferentPass1!"}, "Password.ConfirmPassword:eqfield"},
		{"a new password equal to the current one", Password{CurrentPassword: "SamePass1!abc", Password: "SamePass1!abc",
			ConfirmPassword: "SamePass1!abc"}, "Password.CurrentPassword:nefield"},
		{"a weak new password", Password{CurrentPassword: "OldPass1!abc", Password: "newpass1", ConfirmPassword: "newpass1"},
			"Password.Password:stpass"},
		{"no current password", Password{Password: "NewPass1!abc", ConfirmPassword: "NewPass1!abc"}, "Password.CurrentPassword:min"},
		{"a new password at the bcrypt limit", Password{CurrentPassword: "OldPass1!abc", Password: strings.Repeat("a", MaxPasswordBytes),
			ConfirmPassword: strings.Repeat("a", MaxPasswordBytes)}, ""},
		{"a new password over the bcrypt limit", Password{CurrentPassword: "OldPass1!abc", Password: strings.Repeat("a", MaxPasswordBytes+1),
			ConfirmPassword: strings.Repeat("a", MaxPasswordBytes+1)}, "Password.Password:passlen"},

		{"a complete oauth callback", AuthCallback{Code: "auth-code-123", IdToken: modelsJWT, Scope: "openid email profile",
			State: "random-state-value"}, ""},
		{"an oauth callback without a code", AuthCallback{IdToken: modelsJWT, Scope: "openid email", State: "state123"},
			"AuthCallback.Code:required"},
		{"an oauth callback without the openid scope", AuthCallback{Code: "code123", IdToken: modelsJWT, Scope: "email profile",
			State: "state123"}, "AuthCallback.Scope:oauth_min_scope"},
		{"an oauth callback whose id token is not a jwt", AuthCallback{Code: "code123", IdToken: "not-a-jwt", Scope: "openid email",
			State: "state123"}, "AuthCallback.IdToken:jwt"},

		{"a user with a role", UserRole{Role: admin, User: usersUser()}, ""},
		{"a user with an unnamed role", UserRole{Role: Role{}, User: usersUser()}, "Role.Name:required"},
		{"a role on an invalid user", UserRole{Role: admin, User: withoutMail}, "User.Mail:vmail"},

		{"a user with role privileges", UserRolePrivileges{Role: RolePrivileges{Privileges: read, Role: admin}, User: usersUser()}, ""},
		{"a user with an unnamed privilege", UserRolePrivileges{Role: RolePrivileges{Privileges: []Privilege{{}}, Role: admin},
			User: usersUser()}, "Privilege.Name:required"},
		{"role privileges on an invalid user", UserRolePrivileges{Role: RolePrivileges{Privileges: read, Role: admin},
			User: withoutMail}, "User.Mail:vmail"},

		{"preferences of a user", UserPreferences{UserID: 1}, ""},
		{"preferences of no user", UserPreferences{}, "user_id is required"},
	})
}

func TestUsers_String_SpellsTheStoredValue(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		value fmt.Stringer
		want  string
	}{
		{UserStatusCreated, "created"},
		{UserStatusActive, "active"},
		{UserStatusBlocked, "blocked"},
		{UserTypeLocal, "local"},
		{UserTypeOAuth, "oauth"},
		{UserTypeAPI, "api"},
	} {
		assert.Equal(t, tc.want, tc.value.String())
	}
}

func TestUsers_TableName_NamesTheMappedTable(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		model modelsTabled
		want  string
	}{
		{&User{}, "users"},
		{&UserPassword{}, "users"},
		{&Login{}, "users"},
		{&Password{}, "users"},
		{&UserPreferences{}, "user_preferences"},
	} {
		assert.Equal(t, tc.want, tc.model.TableName(), "%T", tc.model)
	}
}

func TestUsers_UserPreferencesOptions_StoresAndLoadsTheFavouriteFlows(t *testing.T) {
	t.Parallel()

	t.Run("a stored list loads back", func(t *testing.T) {
		t.Parallel()

		stored, err := UserPreferencesOptions{FavoriteFlows: []int64{1, 2, 3}}.Value()
		require.NoError(t, err)
		assert.JSONEq(t, `{"favoriteFlows":[1,2,3]}`, fmt.Sprintf("%s", stored))

		var loaded UserPreferencesOptions
		require.NoError(t, loaded.Scan([]byte(fmt.Sprintf("%s", stored))))
		assert.Equal(t, []int64{1, 2, 3}, loaded.FavoriteFlows)
	})

	t.Run("an empty list is stored under its key", func(t *testing.T) {
		t.Parallel()

		stored, err := UserPreferencesOptions{FavoriteFlows: []int64{}}.Value()
		require.NoError(t, err)
		assert.JSONEq(t, `{"favoriteFlows":[]}`, fmt.Sprintf("%s", stored))
	})

	t.Run("a null column loads as an empty list", func(t *testing.T) {
		t.Parallel()

		var loaded UserPreferencesOptions
		require.NoError(t, loaded.Scan(nil))
		assert.Equal(t, []int64{}, loaded.FavoriteFlows)
	})

	t.Run("a column that is not bytes is refused", func(t *testing.T) {
		t.Parallel()

		var loaded UserPreferencesOptions
		assert.EqualError(t, loaded.Scan(12345), "failed to scan UserPreferencesOptions: expected []byte, got int")
	})

	t.Run("a column that is not JSON is refused", func(t *testing.T) {
		t.Parallel()

		var loaded UserPreferencesOptions
		var syntax *json.SyntaxError
		assert.ErrorAs(t, loaded.Scan([]byte("not json")), &syntax)
	})
}

func TestUsers_NewUserPreferences_StartsWithNoFavouriteFlows(t *testing.T) {
	t.Parallel()

	up := NewUserPreferences(42)

	assert.Equal(t, uint64(42), up.UserID)
	assert.Equal(t, []int64{}, up.Preferences.FavoriteFlows)
}
