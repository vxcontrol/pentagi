package graph

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- UserID ---

func TestGetUserID_Found(t *testing.T) {
	t.Parallel()

	ctx := SetUserID(t.Context(), 42)
	id, err := GetUserID(ctx)
	require.NoError(t, err)
	assert.Equal(t, uint64(42), id)
}

func TestGetUserID_Missing(t *testing.T) {
	t.Parallel()

	_, err := GetUserID(t.Context())
	require.EqualError(t, err, "user ID not found")
}

func TestGetUserID_WrongType(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(t.Context(), UserIDKey, "not-a-uint64")
	_, err := GetUserID(ctx)
	require.EqualError(t, err, "user ID not found")
}

func TestSetUserID_Roundtrip(t *testing.T) {
	t.Parallel()

	ctx := SetUserID(t.Context(), 99)
	id, err := GetUserID(ctx)
	require.NoError(t, err)
	assert.Equal(t, uint64(99), id)
}

// --- UserType ---

func TestGetUserType_Found(t *testing.T) {
	t.Parallel()

	ctx := SetUserType(t.Context(), "local")
	ut, err := GetUserType(ctx)
	require.NoError(t, err)
	assert.Equal(t, "local", ut)
}

func TestGetUserType_Missing(t *testing.T) {
	t.Parallel()

	_, err := GetUserType(t.Context())
	require.EqualError(t, err, "user type not found")
}

func TestGetUserType_WrongType(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(t.Context(), UserTypeKey, 123)
	_, err := GetUserType(ctx)
	require.EqualError(t, err, "user type not found")
}

func TestSetUserType_Roundtrip(t *testing.T) {
	t.Parallel()

	ctx := SetUserType(t.Context(), "oauth")
	ut, err := GetUserType(ctx)
	require.NoError(t, err)
	assert.Equal(t, "oauth", ut)
}

// --- UserPermissions ---

func TestGetUserPermissions_Found(t *testing.T) {
	t.Parallel()

	perms := []string{"flows.read", "flows.admin"}
	ctx := SetUserPermissions(t.Context(), perms)
	got, err := GetUserPermissions(ctx)
	require.NoError(t, err)
	assert.Equal(t, perms, got)
}

func TestGetUserPermissions_Missing(t *testing.T) {
	t.Parallel()

	_, err := GetUserPermissions(t.Context())
	require.EqualError(t, err, "user permissions not found")
}

func TestGetUserPermissions_WrongType(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(t.Context(), UserPermissions, "not-a-slice")
	_, err := GetUserPermissions(ctx)
	require.EqualError(t, err, "user permissions not found")
}

func TestSetUserPermissions_Roundtrip(t *testing.T) {
	t.Parallel()

	perms := []string{"a.read", "b.write"}
	ctx := SetUserPermissions(t.Context(), perms)
	got, err := GetUserPermissions(ctx)
	require.NoError(t, err)
	assert.Equal(t, perms, got)
}

// --- validateUserType ---

func TestValidateUserType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		ctx     context.Context
		allowed []string
		wantOK  bool
		wantErr string
	}{
		{
			name:    "allowed type",
			ctx:     SetUserType(t.Context(), "local"),
			allowed: []string{"local", "oauth"},
			wantOK:  true,
			wantErr: "",
		},
		{
			name:    "type missing from context",
			ctx:     t.Context(),
			allowed: []string{"local"},
			wantOK:  false,
			wantErr: "unauthorized: invalid user type: user type not found",
		},
		{
			name:    "unsupported type",
			ctx:     SetUserType(t.Context(), "apikey"),
			allowed: []string{"local", "oauth"},
			wantOK:  false,
			wantErr: "unauthorized: invalid user type: apikey",
		},
		{
			name:    "oauth type allowed",
			ctx:     SetUserType(t.Context(), "oauth"),
			allowed: []string{"local", "oauth"},
			wantOK:  true,
			wantErr: "",
		},
		{
			name:    "single allowed type matches",
			ctx:     SetUserType(t.Context(), "local"),
			allowed: []string{"local"},
			wantOK:  true,
			wantErr: "",
		},
		{
			name:    "empty allowed list",
			ctx:     SetUserType(t.Context(), "local"),
			allowed: []string{},
			wantOK:  false,
			wantErr: "unauthorized: invalid user type: local",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ok, err := validateUserType(tc.ctx, tc.allowed...)

			assert.Equal(t, tc.wantOK, ok, "ok value mismatch")

			if tc.wantErr != "" {
				require.EqualError(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// --- validatePermission ---

func TestValidatePermission(t *testing.T) {
	t.Parallel()

	makeCtx := func(uid uint64, perms []string) context.Context {
		ctx := SetUserID(t.Context(), uid)
		return SetUserPermissions(ctx, perms)
	}

	tests := []struct {
		name      string
		ctx       context.Context
		perm      string
		wantUID   int64
		wantAdmin bool
		wantErr   string
	}{
		{
			name:      "exact permission match",
			ctx:       makeCtx(1, []string{"flows.read"}),
			perm:      "flows.read",
			wantUID:   1,
			wantAdmin: false,
			wantErr:   "",
		},
		{
			name:      "admin permission via wildcard",
			ctx:       makeCtx(2, []string{"flows.admin"}),
			perm:      "flows.read",
			wantUID:   2,
			wantAdmin: true,
			wantErr:   "",
		},
		{
			name:      "admin permission for write",
			ctx:       makeCtx(3, []string{"tasks.admin"}),
			perm:      "tasks.write",
			wantUID:   3,
			wantAdmin: true,
			wantErr:   "",
		},
		{
			name:      "admin permission for delete",
			ctx:       makeCtx(4, []string{"users.admin"}),
			perm:      "users.delete",
			wantUID:   4,
			wantAdmin: true,
			wantErr:   "",
		},
		{
			name:      "multiple permissions with admin",
			ctx:       makeCtx(5, []string{"flows.read", "tasks.admin", "users.write"}),
			perm:      "tasks.read",
			wantUID:   5,
			wantAdmin: true,
			wantErr:   "",
		},
		{
			name:      "multiple permissions exact match",
			ctx:       makeCtx(6, []string{"flows.read", "tasks.write", "users.admin"}),
			perm:      "flows.read",
			wantUID:   6,
			wantAdmin: false,
			wantErr:   "",
		},
		{
			name:    "user ID missing",
			ctx:     SetUserPermissions(t.Context(), []string{"flows.read"}),
			perm:    "flows.read",
			wantErr: "unauthorized: invalid user: user ID not found",
		},
		{
			name:    "permissions missing",
			ctx:     SetUserID(t.Context(), 3),
			perm:    "flows.read",
			wantErr: "unauthorized: invalid user permissions: user permissions not found",
		},
		{
			name:    "permission not found",
			ctx:     makeCtx(4, []string{"other.read"}),
			perm:    "flows.read",
			wantErr: "requested permission 'flows.read' not found",
		},
		{
			name:    "empty permissions list",
			ctx:     makeCtx(7, []string{}),
			perm:    "flows.read",
			wantErr: "requested permission 'flows.read' not found",
		},
		{
			name:      "permission without dot separator",
			ctx:       makeCtx(8, []string{"admin"}),
			perm:      "admin",
			wantUID:   8,
			wantAdmin: true,
			wantErr:   "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			uid, admin, err := validatePermission(tc.ctx, tc.perm)

			if tc.wantErr != "" {
				require.EqualError(t, err, tc.wantErr)
				assert.Equal(t, int64(0), uid, "uid should be 0 on error")
				assert.False(t, admin, "admin should be false on error")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.wantUID, uid)
				assert.Equal(t, tc.wantAdmin, admin)
			}
		})
	}
}

func TestPermAdminRegexp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"flows.read to flows.admin", "flows.read", "flows.admin"},
		{"tasks.write to tasks.admin", "tasks.write", "tasks.admin"},
		{"users.delete to users.admin", "users.delete", "users.admin"},
		{"assistants.create to assistants.admin", "assistants.create", "assistants.admin"},
		{"no dot separator", "admin", "admin"},
		{"multiple dots", "system.flows.read", "system.flows.admin"},
		{"uppercase action no match", "flows.READ", "flows.READ"},
		{"numbers in resource", "task123.read", "task123.admin"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := permAdminRegexp.ReplaceAllString(tt.input, "$1.admin")
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidatePermission_ZeroUserID(t *testing.T) {
	t.Parallel()

	ctx := SetUserID(t.Context(), 0)
	ctx = SetUserPermissions(ctx, []string{"flows.read"})

	uid, admin, err := validatePermission(ctx, "flows.read")
	require.NoError(t, err)
	assert.Equal(t, int64(0), uid)
	assert.False(t, admin)
}

func TestValidatePermission_LargeUserID(t *testing.T) {
	t.Parallel()

	ctx := SetUserID(t.Context(), 9223372036854775807) // max int64
	ctx = SetUserPermissions(ctx, []string{"flows.read"})

	uid, admin, err := validatePermission(ctx, "flows.read")
	require.NoError(t, err)
	assert.Equal(t, int64(9223372036854775807), uid)
	assert.False(t, admin)
}

func TestGetUserID_ZeroValue(t *testing.T) {
	t.Parallel()

	ctx := SetUserID(t.Context(), 0)
	id, err := GetUserID(ctx)
	require.NoError(t, err)
	assert.Equal(t, uint64(0), id)
}

func TestGetUserPermissions_EmptySlice(t *testing.T) {
	t.Parallel()

	ctx := SetUserPermissions(t.Context(), []string{})
	perms, err := GetUserPermissions(ctx)
	require.NoError(t, err)
	assert.Equal(t, []string{}, perms)
	assert.Len(t, perms, 0)
}

func TestGetUserPermissions_NilSlice(t *testing.T) {
	t.Parallel()

	ctx := SetUserPermissions(t.Context(), nil)
	perms, err := GetUserPermissions(ctx)
	require.NoError(t, err)
	assert.Nil(t, perms)
}
