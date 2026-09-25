package services

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"pentagi/pkg/controller"
	"pentagi/pkg/database"
	"pentagi/pkg/database/knowledge"
	"pentagi/pkg/graph"
	"pentagi/pkg/version"

	"github.com/99designs/gqlgen/graphql"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGraphqlErrors_RedactGqlError_AnswersEachClassWithAFixedMessageAndCode(t *testing.T) {
	t.Parallel()

	dial := &net.OpError{
		Op:  "dial",
		Net: "tcp",
		Err: &net.DNSError{Err: "no such host", Name: "pgvector", Server: "127.0.0.11:53"},
	}
	gatewayRefused := &url.Error{
		Op:  "Post",
		URL: "http://127.0.0.1:1/v1/chat/completions",
		Err: &net.OpError{Op: "dial", Net: "tcp", Err: errors.New("connect: connection refused")},
	}

	tests := []struct {
		name     string
		err      error
		wantMsg  string
		wantCode string
		wantOk   bool
	}{
		{
			name:     "no rows, wrapped the way resolvers wrap it",
			err:      fmt.Errorf("failed to update prompt: %w", sql.ErrNoRows),
			wantMsg:  "not found",
			wantCode: gqlCodeNotFound,
			wantOk:   true,
		},
		{
			name:     "a flow the controller cannot find",
			err:      controller.ErrFlowNotFound,
			wantMsg:  "not found",
			wantCode: gqlCodeNotFound,
			wantOk:   true,
		},
		{
			name:     "gorm record not found",
			err:      fmt.Errorf("failed to get user: %w", gorm.ErrRecordNotFound),
			wantMsg:  "not found",
			wantCode: gqlCodeNotFound,
			wantOk:   true,
		},
		{
			name: "unique violation carries the index name nobody outside may see",
			err: fmt.Errorf("failed to create token in database: %w", &pq.Error{
				Code:       pgUniqueViolation,
				Message:    `duplicate key value violates unique constraint "api_tokens_name_user_unique_idx"`,
				Constraint: "api_tokens_name_user_unique_idx",
			}),
			wantMsg:  "already exists",
			wantCode: gqlCodeAlreadyExists,
			wantOk:   true,
		},
		{
			name:     "foreign key violation",
			err:      fmt.Errorf("failed to create flow: %w", &pq.Error{Code: pgForeignKeyViolation}),
			wantMsg:  "referenced record does not exist",
			wantCode: gqlCodeInvalidReference,
			wantOk:   true,
		},
		{
			name:     "any other driver error stays generic",
			err:      fmt.Errorf("query failed: %w", &pq.Error{Code: "42P01", Message: `relation "flows" does not exist`}),
			wantMsg:  "internal server error",
			wantCode: gqlCodeInternal,
			wantOk:   true,
		},
		{
			name:     "an unreachable database",
			err:      dial,
			wantMsg:  "a required service is unavailable",
			wantCode: gqlCodeInternal,
			wantOk:   true,
		},
		{
			name:     "an unreachable database, wrapped the way a resolver returns it",
			err:      fmt.Errorf("failed to get flows: %w", dial),
			wantMsg:  "a required service is unavailable",
			wantCode: gqlCodeInternal,
			wantOk:   true,
		},
		{
			name: "an unreachable model gateway, which is not storage",
			err: fmt.Errorf("failed to create flow worker: %w",
				fmt.Errorf("failed to get flow provider: %w",
					fmt.Errorf("failed to call llm after %d retries: %w", 3, gatewayRefused))),
			wantMsg:  "a required service is unavailable",
			wantCode: gqlCodeInternal,
			wantOk:   true,
		},
		{
			name:     "a request that ran out of time",
			err:      fmt.Errorf("knowledge: embed query: %w", context.DeadlineExceeded),
			wantMsg:  "a required service is unavailable",
			wantCode: gqlCodeInternal,
			wantOk:   true,
		},
		{
			name:     "no user in context",
			err:      fmt.Errorf("%w: invalid user: nope", graph.ErrUnauthenticated),
			wantMsg:  "authentication required",
			wantCode: gqlCodeUnauthenticated,
			wantOk:   true,
		},
		{
			name:     "a missing privilege, which the answer must not name",
			err:      fmt.Errorf("%w: requested permission %q not found", graph.ErrForbidden, "flows.view"),
			wantMsg:  "not permitted",
			wantCode: gqlCodeForbidden,
			wantOk:   true,
		},
		{
			name:     "an argument the schema reads as a number, given as text",
			err:      graphql.ErrorOnPath(context.Background(), &strconv.NumError{Func: "ParseInt", Num: "abc", Err: strconv.ErrSyntax}),
			wantMsg:  "argument is not a valid number",
			wantCode: gqlCodeBadUserInput,
			wantOk:   true,
		},
		{
			name:     "a knowledge document whose sub-type is missing",
			err:      fmt.Errorf("%w: answer document requires answer type", knowledge.ErrInvalidDocument),
			wantMsg:  "invalid knowledge document: answer document requires answer type",
			wantCode: gqlCodeBadUserInput,
			wantOk:   true,
		},

		{name: "validation message", err: errors.New("question is required")},
		{name: "length message", err: errors.New("token name must not exceed 100 characters")},
		{name: "permission message", err: errors.New("requested permission 'settings.providers.edit' not found")},
		{name: "conflict message", err: errors.New("provider name is required")},
		{name: "nil-free unknown error", err: errors.New("failed to launch container: port is already allocated")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			msg, code, ok := redactGqlError(tt.err)

			assert.Equal(t, tt.wantOk, ok)
			assert.Equal(t, tt.wantMsg, msg)
			assert.Equal(t, tt.wantCode, code)
		})
	}
}

func TestGraphqlErrors_NewGqlErrorPresenter_RedactsTheMessageAndShowsTheOriginalOnlyInDevelop(t *testing.T) {
	log := logrus.New()
	log.SetOutput(io.Discard)
	present := newGqlErrorPresenter(logrus.NewEntry(log))

	t.Run("replaces a storage error with a redacted message carrying its code", func(t *testing.T) {
		version.PackageVer = "1.0.0"
		t.Cleanup(func() { version.PackageVer = "" })

		presented := present(context.Background(), fmt.Errorf("failed to create token in database: %w", &pq.Error{
			Code:       pgUniqueViolation,
			Message:    `duplicate key value violates unique constraint "api_tokens_name_user_unique_idx"`,
			Constraint: "api_tokens_name_user_unique_idx",
		}))

		assert.Equal(t, "already exists", presented.Message)
		assert.Equal(t, gqlCodeAlreadyExists, presented.Extensions["code"])
		assert.NotContains(t, presented.Extensions, "error")
	})

	t.Run("leaves a message the resolver wrote itself untouched", func(t *testing.T) {
		version.PackageVer = "1.0.0"
		t.Cleanup(func() { version.PackageVer = "" })

		presented := present(context.Background(), errors.New("question is required"))

		assert.Equal(t, "question is required", presented.Message)
		assert.NotContains(t, presented.Extensions, "code")
		assert.NotContains(t, presented.Extensions, "error")
	})

	t.Run("attaches the original error in develop mode", func(t *testing.T) {
		version.PackageVer = ""
		original := fmt.Errorf("failed to update prompt: %w", sql.ErrNoRows)

		presented := present(context.Background(), original)

		assert.Equal(t, "not found", presented.Message)
		assert.Equal(t, gqlCodeNotFound, presented.Extensions["code"])
		assert.Equal(t, original.Error(), presented.Extensions["error"])
	})

	t.Run("withholds the original error outside develop mode", func(t *testing.T) {
		version.PackageVer = "1.0.0"
		t.Cleanup(func() { version.PackageVer = "" })

		presented := present(context.Background(), fmt.Errorf("failed to update prompt: %w", sql.ErrNoRows))

		assert.Equal(t, "not found", presented.Message)
		assert.NotContains(t, presented.Extensions, "error")
	})
}

func TestGraphqlErrors_NewGqlErrorPresenter_LogsEachClassAtItsLevelUnderItsLabel(t *testing.T) {
	for _, tc := range []struct {
		name      string
		err       error
		wantLevel logrus.Level
		wantLabel string
	}{
		{"a missing privilege is the client's doing", fmt.Errorf("%w: requested permission %q not found", graph.ErrForbidden, "flows.view"),
			logrus.WarnLevel, "graphql request refused"},
		{"no user in context is the client's doing", fmt.Errorf("%w: invalid user", graph.ErrUnauthenticated),
			logrus.WarnLevel, "graphql request refused"},
		{"a row that is not there is the client's doing", fmt.Errorf("failed to get flow: %w", sql.ErrNoRows),
			logrus.WarnLevel, "graphql record not found"},
		{"a duplicate name is the client's doing", &pq.Error{Code: pgUniqueViolation},
			logrus.WarnLevel, "graphql constraint violation"},
		{"a dangling reference is the client's doing", &pq.Error{Code: pgForeignKeyViolation},
			logrus.WarnLevel, "graphql invalid reference"},
		{"an id that is not a number is the client's doing",
			fmt.Errorf("invalid flow id: %w", &strconv.NumError{Func: "ParseInt", Num: "abc", Err: strconv.ErrSyntax}),
			logrus.WarnLevel, "graphql invalid argument"},
		{"an unreachable dependency is ours", &net.OpError{Op: "dial", Net: "tcp", Err: errors.New("refused")},
			logrus.ErrorLevel, "graphql internal error"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			logger := logrus.New()
			logger.SetOutput(io.Discard)
			hook := test.NewLocal(logger)

			_ = newGqlErrorPresenter(logrus.NewEntry(logger))(context.Background(), tc.err)

			entry := hook.LastEntry()
			require.NotNil(t, entry, "the presenter logged nothing")
			assert.Equal(t, tc.wantLevel, entry.Level)
			assert.Equal(t, tc.wantLabel, entry.Message)
		})
	}
}

type noRowsDBTX struct {
	db *sql.DB
}

func (f noRowsDBTX) ExecContext(context.Context, string, ...interface{}) (sql.Result, error) {
	return nil, sql.ErrNoRows
}

func (f noRowsDBTX) PrepareContext(context.Context, string) (*sql.Stmt, error) {
	return nil, sql.ErrNoRows
}

func (f noRowsDBTX) QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error) {
	return nil, sql.ErrNoRows
}

func (f noRowsDBTX) QueryRowContext(ctx context.Context, _ string, _ ...interface{}) *sql.Row {
	return f.db.QueryRowContext(ctx, "SELECT 1 WHERE 1 = 0")
}

type gqlErrorResponse struct {
	Errors []struct {
		Message    string         `json:"message"`
		Extensions map[string]any `json:"extensions"`
	} `json:"errors"`
}

func askGraphqlForMissingFlow(t *testing.T) (*httptest.ResponseRecorder, gqlErrorResponse) {
	t.Helper()

	sqlDB, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })

	svc := NewGraphqlService(
		database.New(noRowsDBTX{db: sqlDB}),
		nil, "", nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("uid", uint64(1))
	c.Set("prm", []string{"flows.view"})
	body := bytes.NewBufferString(`{"query":"{ flow(flowId: \"1\") { id title } }"}`)
	c.Request = httptest.NewRequest(http.MethodPost, "/graphql", body)
	c.Request.Header.Set("Content-Type", "application/json")

	svc.ServeGraphql(c)

	var resp gqlErrorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	return w, resp
}

func TestGraphqlErrors_NewGqlErrorPresenter_RedactsWhatServeGraphqlAnswers(t *testing.T) {
	t.Run("production answers not found without the driver text", func(t *testing.T) {
		version.PackageVer = "1.0.0"
		t.Cleanup(func() { version.PackageVer = "" })

		w, resp := askGraphqlForMissingFlow(t)

		require.Len(t, resp.Errors, 1)
		assert.Equal(t, "not found", resp.Errors[0].Message)
		assert.Equal(t, gqlCodeNotFound, resp.Errors[0].Extensions["code"])
		assert.NotContains(t, w.Body.String(), "sql: no rows in result set")
	})

	t.Run("develop mode adds the error extension production withholds", func(t *testing.T) {
		version.PackageVer = ""

		_, resp := askGraphqlForMissingFlow(t)

		require.Len(t, resp.Errors, 1)
		assert.Equal(t, "not found", resp.Errors[0].Message)
		assert.Contains(t, resp.Errors[0].Extensions["error"], "sql: no rows in result set")
	})
}
