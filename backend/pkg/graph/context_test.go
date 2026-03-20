package graph

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- UserID ---

func TestGetUserID_Found(t *testing.T) {
	ctx := SetUserID(context.Background(), 42)
	id, err := GetUserID(ctx)
	require.NoError(t, err)
	assert.Equal(t, uint64(42), id)
}

func TestGetUserID_Missing(t *testing.T) {
	_, err := GetUserID(context.Background())
	require.EqualError(t, err, "user ID not found")
}

func TestGetUserID_WrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), UserIDKey, "not-a-uint64")
	_, err := GetUserID(ctx)
	require.EqualError(t, err, "user ID not found")
}

func TestSetUserID_Roundtrip(t *testing.T) {
	ctx := SetUserID(context.Background(), 99)
	id, err := GetUserID(ctx)
	require.NoError(t, err)
	assert.Equal(t, uint64(99), id)
}

// --- UserType ---

func TestGetUserType_Found(t *testing.T) {
	ctx := SetUserType(context.Background(), "local")
	ut, err := GetUserType(ctx)
	require.NoError(t, err)
	assert.Equal(t, "local", ut)
}

func TestGetUserType_Missing(t *testing.T) {
	_, err := GetUserType(context.Background())
	require.EqualError(t, err, "user type not found")
}

func TestGetUserType_WrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), UserTypeKey, 123)
	_, err := GetUserType(ctx)
	require.EqualError(t, err, "user type not found")
}

func TestSetUserType_Roundtrip(t *testing.T) {
	ctx := SetUserType(context.Background(), "oauth")
	ut, err := GetUserType(ctx)
	require.NoError(t, err)
	assert.Equal(t, "oauth", ut)
}

// --- UserPermissions ---

func TestGetUserPermissions_Found(t *testing.T) {
	perms := []string{"flows.read", "flows.admin"}
	ctx := SetUserPermissions(context.Background(), perms)
	got, err := GetUserPermissions(ctx)
	require.NoError(t, err)
	assert.Equal(t, perms, got)
}

func TestGetUserPermissions_Missing(t *testing.T) {
	_, err := GetUserPermissions(context.Background())
	require.EqualError(t, err, "user permissions not found")
}

func TestGetUserPermissions_WrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), UserPermissions, "not-a-slice")
	_, err := GetUserPermissions(ctx)
	require.EqualError(t, err, "user permissions not found")
}

func TestSetUserPermissions_Roundtrip(t *testing.T) {
	perms := []string{"a.read", "b.write"}
	ctx := SetUserPermissions(context.Background(), perms)
	got, err := GetUserPermissions(ctx)
	require.NoError(t, err)
	assert.Equal(t, perms, got)
}

// --- validateUserType ---

func TestValidateUserType(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		allowed []string
		wantOK  bool
		wantErr string
	}{
		{
			name:    "allowed type",
			ctx:     SetUserType(context.Background(), "local"),
			allowed: []string{"local", "oauth"},
			wantOK:  true,
		},
		{
			name:    "type missing from context",
			ctx:     context.Background(),
			allowed: []string{"local"},
			wantErr: "unauthorized: invalid user type: user type not found",
		},
		{
			name:    "unsupported type",
			ctx:     SetUserType(context.Background(), "apikey"),
			allowed: []string{"local", "oauth"},
			wantErr: "unauthorized: invalid user type: apikey",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ok, err := validateUserType(tc.ctx, tc.allowed...)
			if tc.wantErr != "" {
				require.EqualError(t, err, tc.wantErr)
				assert.False(t, ok)
			} else {
				require.NoError(t, err)
				assert.True(t, ok)
			}
		})
	}
}

// --- validatePermission ---

func TestValidatePermission(t *testing.T) {
	makeCtx := func(uid uint64, perms []string) context.Context {
		ctx := SetUserID(context.Background(), uid)
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
		},
		{
			name:      "admin permission via wildcard",
			ctx:       makeCtx(2, []string{"flows.admin"}),
			perm:      "flows.read",
			wantUID:   2,
			wantAdmin: true,
		},
		{
			name:    "user ID missing",
			ctx:     SetUserPermissions(context.Background(), []string{"flows.read"}),
			perm:    "flows.read",
			wantErr: "unauthorized: invalid user: user ID not found",
		},
		{
			name:    "permissions missing",
			ctx:     SetUserID(context.Background(), 3),
			perm:    "flows.read",
			wantErr: "unauthorized: invalid user permissions: user permissions not found",
		},
		{
			name:    "permission not found",
			ctx:     makeCtx(4, []string{"other.read"}),
			perm:    "flows.read",
			wantErr: "requested permission 'flows.read' not found",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uid, admin, err := validatePermission(tc.ctx, tc.perm)
			if tc.wantErr != "" {
				require.EqualError(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.wantUID, uid)
				assert.Equal(t, tc.wantAdmin, admin)
			}
		})
	}
}
