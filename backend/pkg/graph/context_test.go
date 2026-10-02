package graph

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"testing"

	"pentagi/pkg/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContext_GetUserID_ReturnsTheStoredIDOrNotFound(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name    string
		ctx     context.Context
		want    uint64
		wantErr string
	}{
		{name: "an id set by SetUserID", ctx: SetUserID(context.Background(), 42), want: 42},
		{name: "a zero id is still an id", ctx: SetUserID(context.Background(), 0), want: 0},
		{name: "no id in the context", ctx: context.Background(), wantErr: "user ID not found"},
		{name: "a value of another type", ctx: context.WithValue(context.Background(), UserIDKey, "not-a-uint64"), wantErr: "user ID not found"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := GetUserID(tt.ctx)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestContext_GetUserType_ReturnsTheStoredTypeOrNotFound(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name    string
		ctx     context.Context
		want    string
		wantErr string
	}{
		{name: "a type set by SetUserType", ctx: SetUserType(context.Background(), "local"), want: "local"},
		{name: "no type in the context", ctx: context.Background(), wantErr: "user type not found"},
		{name: "a value of another type", ctx: context.WithValue(context.Background(), UserTypeKey, 123), wantErr: "user type not found"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := GetUserType(tt.ctx)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestContext_GetUserPermissions_ReturnsTheStoredSliceOrNotFound(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name    string
		ctx     context.Context
		want    []string
		wantErr string
	}{
		{
			name: "permissions set by SetUserPermissions",
			ctx:  SetUserPermissions(context.Background(), []string{"flows.read", "flows.admin"}),
			want: []string{"flows.read", "flows.admin"},
		},
		{name: "an empty slice stays empty", ctx: SetUserPermissions(context.Background(), []string{}), want: []string{}},
		{name: "a nil slice stays nil", ctx: SetUserPermissions(context.Background(), nil), want: nil},
		{name: "no permissions in the context", ctx: context.Background(), wantErr: "user permissions not found"},
		{
			name:    "a value of another type",
			ctx:     context.WithValue(context.Background(), UserPermissions, "not-a-slice"),
			wantErr: "user permissions not found",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := GetUserPermissions(tt.ctx)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestContext_ValidateUserType_AllowsOnlyTheListedSessionTypes(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name         string
		ctx          context.Context
		allowed      []string
		wantOK       bool
		wantSentinel error
		wantIn       string
	}{
		{name: "allowed type", ctx: SetUserType(context.Background(), "local"), allowed: []string{"local", "oauth"}, wantOK: true},
		{name: "allowed type later in the list", ctx: SetUserType(context.Background(), "oauth"), allowed: []string{"local", "oauth"}, wantOK: true},
		{
			name:         "type missing from context",
			ctx:          context.Background(),
			allowed:      []string{"local"},
			wantSentinel: ErrUnauthenticated,
			wantIn:       "user type not found",
		},
		{
			name:         "unsupported type",
			ctx:          SetUserType(context.Background(), "apikey"),
			allowed:      []string{"local", "oauth"},
			wantSentinel: ErrForbidden,
			wantIn:       "apikey",
		},
		{
			name:         "empty allowed list",
			ctx:          SetUserType(context.Background(), "local"),
			allowed:      []string{},
			wantSentinel: ErrForbidden,
			wantIn:       "local",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ok, err := validateUserType(tt.ctx, tt.allowed...)

			assert.Equal(t, tt.wantOK, ok, "ok value mismatch")
			if tt.wantSentinel == nil {
				require.NoError(t, err)
				return
			}
			require.ErrorIs(t, err, tt.wantSentinel)
			assert.Contains(t, err.Error(), tt.wantIn)
		})
	}
}

func TestContext_ValidatePermission_GrantsTheExactPermissionOrItsAdmin(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name         string
		ctx          context.Context
		perm         string
		wantUID      int64
		wantAdmin    bool
		wantErr      string
		wantSentinel error
	}{
		{name: "exact permission match", ctx: graphUserContext(1, "flows.read"), perm: "flows.read", wantUID: 1},
		{name: "the resource's admin grants the action", ctx: graphUserContext(2, "flows.admin"), perm: "flows.delete", wantUID: 2, wantAdmin: true},
		{
			name:      "an admin grant later in the list",
			ctx:       graphUserContext(5, "flows.read", "tasks.admin", "users.write"),
			perm:      "tasks.subscribe",
			wantUID:   5,
			wantAdmin: true,
		},
		{
			name:    "another resource's admin does not widen an exact grant",
			ctx:     graphUserContext(6, "flows.read", "tasks.write", "users.admin"),
			perm:    "flows.read",
			wantUID: 6,
		},
		{name: "permission without dot separator", ctx: graphUserContext(8, "admin"), perm: "admin", wantUID: 8, wantAdmin: true},
		{
			name:      "a dotted resource keeps every segment before the action",
			ctx:       graphUserContext(9, "settings.providers.admin"),
			perm:      "settings.providers.edit",
			wantUID:   9,
			wantAdmin: true,
		},
		{name: "digits in the resource are kept", ctx: graphUserContext(10, "task123.admin"), perm: "task123.read", wantUID: 10, wantAdmin: true},
		{name: "a zero user id is returned as is", ctx: graphUserContext(0, "flows.read"), perm: "flows.read", wantUID: 0},
		{
			name:    "the largest int64 user id survives the conversion",
			ctx:     graphUserContext(9223372036854775807, "flows.read"),
			perm:    "flows.read",
			wantUID: 9223372036854775807,
		},
		{
			name:         "user ID missing",
			ctx:          SetUserPermissions(context.Background(), []string{"flows.read"}),
			perm:         "flows.read",
			wantErr:      "invalid user: user ID not found",
			wantSentinel: ErrUnauthenticated,
		},
		{
			name:         "permissions missing",
			ctx:          SetUserID(context.Background(), 3),
			perm:         "flows.read",
			wantErr:      "invalid user permissions: user permissions not found",
			wantSentinel: ErrUnauthenticated,
		},
		{
			name:         "permission not found",
			ctx:          graphUserContext(4, "other.read"),
			perm:         "flows.read",
			wantErr:      `requested permission "flows.read" not found`,
			wantSentinel: ErrForbidden,
		},
		{
			name:         "empty permissions list",
			ctx:          graphUserContext(7),
			perm:         "flows.read",
			wantErr:      `requested permission "flows.read" not found`,
			wantSentinel: ErrForbidden,
		},
		{
			name:         "an upper-case action is not widened to admin",
			ctx:          graphUserContext(11, "flows.admin"),
			perm:         "flows.READ",
			wantErr:      `requested permission "flows.READ" not found`,
			wantSentinel: ErrForbidden,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uid, admin, err := validatePermission(tt.ctx, tt.perm)

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				assert.ErrorIs(t, err, tt.wantSentinel, "the presenter classifies by sentinel, not by prose")
				assert.Equal(t, int64(0), uid, "uid should be 0 on error")
				assert.False(t, admin, "admin should be false on error")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantUID, uid)
			assert.Equal(t, tt.wantAdmin, admin)
		})
	}
}

func TestContext_ValidatePermissionWithFlowID_RefusesAnotherUsersFlowUnlessAdmin(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name    string
		perm    string
		wantUID int64
		wantErr error
	}{
		{name: "another user's flow is refused", perm: "flows.view", wantErr: ErrForbidden},
		{name: "an admin passes on another user's flow", perm: "flows.admin", wantUID: 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uid, err := validatePermissionWithFlowID(graphUserContext(1, tt.perm), "flows.view", 5, &graphDB{flowOwner: 2})

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantUID, uid)
		})
	}
}

func TestContext_ValidateUserResources_RefusesAnotherUsersResourceUnlessAdmin(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name    string
		isAdmin bool
		wantErr error
	}{
		{name: "another user's resource is refused", wantErr: ErrForbidden},
		{name: "an admin gets another user's resource", isAdmin: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := &graphDB{resources: []database.UserResource{{ID: 7, UserID: 2}}}

			got, err := validateUserResources(context.Background(), db, 1, tt.isAdmin, []int64{7})

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Len(t, got, 1)
			assert.Equal(t, int64(7), got[0].ID)
		})
	}
}

func TestContext_ParseFlowID_ReadsTheNumberAndNamesABadGroupID(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name    string
		groupID string
		want    int64
		wantErr string
	}{
		{name: "a flow id", groupID: "flow-42", want: 42},
		{name: "a zero flow id", groupID: "flow-0", want: 0},
		{name: "an id past the int32 range", groupID: "flow-9999999999", want: 9999999999},
		{name: "no number after the prefix", groupID: "flow-", wantErr: "invalid flow ID"},
		{name: "letters after the prefix", groupID: "flow-abc", wantErr: "invalid flow ID"},
		{name: "a number past the int64 range", groupID: "flow-99999999999999999999", wantErr: "invalid flow ID"},
		{name: "no separator", groupID: "invalid", wantErr: "invalid groupId format"},
		{name: "an empty group id", groupID: "", wantErr: "invalid groupId format"},
		{name: "another prefix", groupID: "group-42", wantErr: "invalid groupId format"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseFlowID(tt.groupID)

			if tt.wantErr == "" {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
				return
			}
			require.ErrorContains(t, err, tt.wantErr)
			var numErr *strconv.NumError
			assert.False(t, errors.As(err, &numErr),
				"the door classifies a bare number-parsing error as an unreadable argument and drops this message")
			assert.Contains(t, err.Error(), strconv.Quote(tt.groupID))
			assert.NotContains(t, err.Error(), "strconv")
		})
	}
}

func TestContext_ValidateKnowledgeFieldLengths_RefusesTheFieldOverItsLimit(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name        string
		content     string
		question    *string
		description *string
		codeLang    *string
		wantErr     string
	}{
		{name: "every field empty or absent"},
		{
			name:        "every field at its limit in multibyte runes",
			content:     strings.Repeat("я", maxKnowledgeContentLen),
			question:    strPtr(strings.Repeat("я", maxKnowledgeQuestionLen)),
			description: strPtr(strings.Repeat("я", maxKnowledgeDescriptionLen)),
			codeLang:    strPtr(strings.Repeat("я", maxKnowledgeCodeLangLen)),
		},
		{name: "content one over", content: strings.Repeat("a", maxKnowledgeContentLen+1), wantErr: "content must not exceed"},
		{name: "question one over", question: strPtr(strings.Repeat("a", maxKnowledgeQuestionLen+1)), wantErr: "question must not exceed"},
		{
			name:        "description one over",
			description: strPtr(strings.Repeat("a", maxKnowledgeDescriptionLen+1)),
			wantErr:     "description must not exceed",
		},
		{name: "code language one over", codeLang: strPtr(strings.Repeat("a", maxKnowledgeCodeLangLen+1)), wantErr: "code language must not exceed"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateKnowledgeFieldLengths(tt.content, tt.question, tt.description, tt.codeLang)

			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestContext_ValidateKnowledgeSearch_BoundsTheLimitAndValidatesTheQuery(t *testing.T) {
	t.Parallel()

	intPtr := func(v int) *int { return &v }

	for _, tt := range []struct {
		name    string
		query   string
		limit   *int
		wantLim int
		wantErr string
	}{
		{name: "no limit means the store default", query: "ports", wantLim: 0},
		{name: "explicit zero limit means the store default", query: "ports", limit: intPtr(0), wantLim: 0},
		{name: "limit at the max", query: "ports", limit: intPtr(100), wantLim: 100},
		{name: "limit one over the max", query: "ports", limit: intPtr(101), wantErr: "limit must not exceed 100"},
		{name: "limit one below zero", query: "ports", limit: intPtr(-1), wantErr: "limit must not be negative"},
		{name: "a blank query", query: "   ", limit: intPtr(10), wantErr: "question is required"},
		{name: "a query one over the max", query: strings.Repeat("a", maxKnowledgeQuestionLen+1), limit: intPtr(10), wantErr: "question must not exceed"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			lim, err := validateKnowledgeSearch(tt.query, tt.limit)

			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantLim, lim)
		})
	}
}

func TestContext_ValidateFlowTemplateFields_TrimsTheTitleAndKeepsTheTextAsSent(t *testing.T) {
	t.Parallel()

	longTitle, longText := strings.Repeat("я", 255), strings.Repeat("я", 65536)

	for _, tt := range []struct {
		name      string
		title     string
		text      string
		wantTitle string
		wantText  string
		wantErr   string
	}{
		{name: "a padded title is trimmed and a padded text is not", title: "  Padded  ", text: "  body  ", wantTitle: "Padded", wantText: "  body  "},
		{name: "the line break a text ends with is kept", title: "title", text: "# Plan\n\n- scan\n", wantTitle: "title", wantText: "# Plan\n\n- scan\n"},
		{name: "whitespace-only title", title: "\t\n  ", text: "body", wantErr: "title is required"},
		{name: "whitespace-only text", title: "title", text: " \t\n", wantErr: "text is required"},
		{name: "multibyte title at the limit", title: longTitle, text: "body", wantTitle: longTitle, wantText: "body"},
		{name: "title one over the limit", title: strings.Repeat("a", 256), text: "body", wantErr: "title must not exceed 255 characters"},
		{name: "multibyte text at the limit", title: "title", text: longText, wantTitle: "title", wantText: longText},
		{name: "text one over the limit", title: "title", text: strings.Repeat("a", 65537), wantErr: "text must not exceed 65536 characters"},
		{name: "a final line break counts toward the limit", title: "title", text: longText + "\n", wantErr: "text must not exceed 65536 characters"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotTitle, gotText, err := validateFlowTemplateFields(tt.title, tt.text)

			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantTitle, gotTitle)
			assert.Equal(t, tt.wantText, gotText)
		})
	}
}
