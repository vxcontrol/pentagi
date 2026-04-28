package graph

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"

	"pentagi/pkg/database"
	"pentagi/pkg/flowfiles"
	"pentagi/pkg/graph/model"
)

// This file will not be regenerated automatically.
//
// It contains helper functions to get and set values in the context.

var permAdminRegexp = regexp.MustCompile(`^(.+)\.[a-z]+$`)

var userSessionTypes = []string{"local", "oauth"}

type GqlContextKey string

const (
	UserIDKey       GqlContextKey = "userID"
	UserTypeKey     GqlContextKey = "userType"
	UserPermissions GqlContextKey = "userPermissions"
)

func GetUserID(ctx context.Context) (uint64, error) {
	userID, ok := ctx.Value(UserIDKey).(uint64)
	if !ok {
		return 0, errors.New("user ID not found")
	}
	return userID, nil
}

func SetUserID(ctx context.Context, userID uint64) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

func GetUserType(ctx context.Context) (string, error) {
	userType, ok := ctx.Value(UserTypeKey).(string)
	if !ok {
		return "", errors.New("user type not found")
	}
	return userType, nil
}

func SetUserType(ctx context.Context, userType string) context.Context {
	return context.WithValue(ctx, UserTypeKey, userType)
}

func GetUserPermissions(ctx context.Context) ([]string, error) {
	userPermissions, ok := ctx.Value(UserPermissions).([]string)
	if !ok {
		return nil, errors.New("user permissions not found")
	}
	return userPermissions, nil
}

func SetUserPermissions(ctx context.Context, userPermissions []string) context.Context {
	return context.WithValue(ctx, UserPermissions, userPermissions)
}

func validateUserType(ctx context.Context, userTypes ...string) (bool, error) {
	userType, err := GetUserType(ctx)
	if err != nil {
		return false, fmt.Errorf("unauthorized: invalid user type: %v", err)
	}

	if !slices.Contains(userTypes, userType) {
		return false, fmt.Errorf("unauthorized: invalid user type: %s", userType)
	}

	return true, nil
}

func validatePermission(ctx context.Context, perm string) (int64, bool, error) {
	uid, err := GetUserID(ctx)
	if err != nil {
		return 0, false, fmt.Errorf("unauthorized: invalid user: %v", err)
	}

	privs, err := GetUserPermissions(ctx)
	if err != nil {
		return 0, false, fmt.Errorf("unauthorized: invalid user permissions: %v", err)
	}

	permAdmin := permAdminRegexp.ReplaceAllString(perm, "$1.admin")
	if isAdmin := slices.Contains(privs, permAdmin); isAdmin {
		return int64(uid), true, nil
	}

	if slices.Contains(privs, perm) {
		return int64(uid), false, nil
	}

	return 0, false, fmt.Errorf("requested permission '%s' not found", perm)
}

func validatePermissionWithFlowID(
	ctx context.Context,
	perm string,
	flowID int64,
	db database.Querier,
) (int64, error) {
	uid, admin, err := validatePermission(ctx, perm)
	if err != nil {
		return 0, err
	}

	flow, err := db.GetFlow(ctx, flowID)
	if err != nil {
		return 0, err
	}

	if !admin && flow.UserID != int64(uid) {
		return 0, fmt.Errorf("not permitted")
	}

	return uid, nil
}

// validateUserResources checks that all given IDs exist and belong to uid (or uid is admin).
// Returns the fetched UserResource records for use in copy operations.
// An empty ids slice is valid and returns nil, nil.
func validateUserResources(
	ctx context.Context,
	db database.Querier,
	uid int64,
	isAdmin bool,
	ids []int64,
) ([]database.UserResource, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	dbResources, err := db.GetUserResourcesByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch resources: %w", err)
	}

	found := make(map[int64]database.UserResource, len(dbResources))
	for _, r := range dbResources {
		found[r.ID] = r
	}

	result := make([]database.UserResource, 0, len(ids))
	for _, id := range ids {
		r, ok := found[id]
		if !ok {
			return nil, fmt.Errorf("resource %d not found", id)
		}
		if !isAdmin && r.UserID != uid {
			return nil, fmt.Errorf("resource %d not accessible", id)
		}
		result = append(result, r)
	}

	return result, nil
}

func convertFlowFiles(files flowfiles.Files) []*model.FlowFile {
	converted := make([]*model.FlowFile, 0, len(files.Files))
	for _, file := range files.Files {
		converted = append(converted, convertFlowFile(file))
	}

	return converted
}

func convertFlowFile(file flowfiles.File) *model.FlowFile {
	return &model.FlowFile{
		ID:         file.ID,
		Name:       file.Name,
		Path:       file.Path,
		Size:       int(file.Size),
		IsDir:      file.IsDir,
		ModifiedAt: file.ModifiedAt,
	}
}
