package graph

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"pentagi/pkg/database"
	"pentagi/pkg/database/knowledge/limits"
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
		return false, fmt.Errorf("%w: invalid user type: %v", ErrUnauthenticated, err)
	}

	if !slices.Contains(userTypes, userType) {
		return false, fmt.Errorf("%w: session type %q is not allowed here", ErrForbidden, userType)
	}

	return true, nil
}

func validatePermission(ctx context.Context, perm string) (int64, bool, error) {
	uid, err := GetUserID(ctx)
	if err != nil {
		return 0, false, fmt.Errorf("%w: invalid user: %v", ErrUnauthenticated, err)
	}

	privs, err := GetUserPermissions(ctx)
	if err != nil {
		return 0, false, fmt.Errorf("%w: invalid user permissions: %v", ErrUnauthenticated, err)
	}

	permAdmin := permAdminRegexp.ReplaceAllString(perm, "$1.admin")
	if isAdmin := slices.Contains(privs, permAdmin); isAdmin {
		return int64(uid), true, nil
	}

	if slices.Contains(privs, perm) {
		return int64(uid), false, nil
	}

	return 0, false, fmt.Errorf("%w: requested permission %q not found", ErrForbidden, perm)
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
		return 0, fmt.Errorf("%w: flow belongs to another user", ErrForbidden)
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
			return nil, fmt.Errorf("%w: resource %d belongs to another user", ErrForbidden, id)
		}
		result = append(result, r)
	}

	return result, nil
}

const (
	maxKnowledgeContentLen     = limits.MaxContentLen
	maxKnowledgeQuestionLen    = limits.MaxQuestionLen
	maxKnowledgeDescriptionLen = limits.MaxDescriptionLen
	maxKnowledgeCodeLangLen    = limits.MaxCodeLangLen
)

// maxAPITokenNameLen MUST stay in sync with the REST model (server/models/api_tokens.go
// `validate` tag) and the frontend tokenNameSchema.
const maxAPITokenNameLen = 100

func validateKnowledgeQuestion(question *string) error {
	if question == nil {
		return nil
	}
	return limits.ValidateQuestion(*question)
}

func validateKnowledgeSearch(query string, limit *int) (int, error) {
	if err := limits.ValidateQuestion(query); err != nil {
		return 0, err
	}

	if limit == nil {
		return 0, nil
	}
	if *limit < 0 {
		return 0, fmt.Errorf("limit must not be negative")
	}
	if *limit > limits.MaxSearchLimit {
		return 0, fmt.Errorf("limit must not exceed %d", limits.MaxSearchLimit)
	}

	return *limit, nil
}

func validateKnowledgeFieldLengths(content string, question, description, codeLang *string) error {
	if err := limits.ValidateContentLen(content); err != nil {
		return err
	}
	if question != nil {
		if err := limits.ValidateQuestionLen(*question); err != nil {
			return err
		}
	}
	if description != nil {
		if err := limits.ValidateDescriptionLen(*description); err != nil {
			return err
		}
	}
	if codeLang != nil {
		if err := limits.ValidateCodeLangLen(*codeLang); err != nil {
			return err
		}
	}
	return nil
}

// MUST stay in sync with the frontend zod schema (pages/templates/template.tsx).
const (
	maxFlowTemplateTitleLen = 255
	maxFlowTemplateTextLen  = 65536
)

// validateFlowTemplateFields returns the title trimmed, as the one line it is, and the text as it was sent,
// the line break it ends with included: the text is a document. A text of nothing but whitespace is refused.
func validateFlowTemplateFields(title, text string) (string, string, error) {
	title = strings.TrimSpace(title)

	if title == "" {
		return "", "", fmt.Errorf("title is required")
	}
	if strings.TrimSpace(text) == "" {
		return "", "", fmt.Errorf("text is required")
	}
	if utf8.RuneCountInString(title) > maxFlowTemplateTitleLen {
		return "", "", fmt.Errorf("title must not exceed %d characters", maxFlowTemplateTitleLen)
	}
	if utf8.RuneCountInString(text) > maxFlowTemplateTextLen {
		return "", "", fmt.Errorf("text must not exceed %d characters", maxFlowTemplateTextLen)
	}
	return title, text, nil
}

func convertFlowFiles(files flowfiles.Files, flowID int64) []*model.FlowFile {
	converted := make([]*model.FlowFile, 0, len(files.Files))
	for _, file := range files.Files {
		converted = append(converted, convertFlowFile(file, flowID))
	}

	return converted
}

func convertFlowFile(file flowfiles.File, flowID int64) *model.FlowFile {
	return &model.FlowFile{
		FlowID:     flowID,
		ID:         file.ID,
		Name:       file.Name,
		Path:       file.Path,
		Size:       int(file.Size),
		IsDir:      file.IsDir,
		ModifiedAt: file.ModifiedAt,
	}
}

// parseFlowID extracts the int64 flow ID from a groupID string like "flow-42".
func parseFlowID(groupID string) (int64, error) {
	parts := strings.SplitN(groupID, "-", 2)
	if len(parts) != 2 || parts[0] != "flow" {
		return 0, fmt.Errorf("invalid groupId format: %q (expected \"flow-<number>\")", groupID)
	}
	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid flow ID in groupId %q (expected \"flow-<number>\")", groupID)
	}
	return id, nil
}
