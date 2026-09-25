package logger

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/errcode"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func TestLogger_WithGinLogger_LogsAClientRefusalBelowErrorLevel(t *testing.T) {
	cases := []struct {
		name   string
		status int
		level  logrus.Level
	}{
		{"forbidden", 403, logrus.WarnLevel},
		{"not found", 404, logrus.WarnLevel},
		{"bad request", 400, logrus.WarnLevel},
		{"server error", 500, logrus.ErrorLevel},
		{"unavailable", 503, logrus.ErrorLevel},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			hook := test.NewGlobal()
			defer hook.Reset()
			logrus.SetLevel(logrus.DebugLevel)

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.Use(WithGinLogger("api"))
			router.GET("/probe", func(c *gin.Context) { c.Status(tc.status) })

			router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/probe", nil))

			var handled *logrus.Entry
			for _, entry := range hook.AllEntries() {
				if entry.Message == "http request handled error" {
					handled = entry
				}
			}

			require.NotNil(t, handled, "the access log said nothing about a %d", tc.status)
			assert.Equal(t, tc.level, handled.Level)

			if tc.level == logrus.WarnLevel {
				for _, entry := range hook.AllEntries() {
					assert.NotEqual(t, logrus.ErrorLevel, entry.Level,
						"a %d produced an error-level entry", tc.status)
				}
			}
		})
	}
}

func TestLogger_WithGqlLogger_LogsErrorsByWhoIsAtFault(t *testing.T) {
	cases := []struct {
		name          string
		codes         []any
		withOperation bool
		level         logrus.Level
	}{
		{"refusal", []any{"FORBIDDEN"}, true, logrus.WarnLevel},
		{"not found", []any{"NOT_FOUND"}, true, logrus.WarnLevel},
		{"duplicate", []any{"ALREADY_EXISTS"}, true, logrus.WarnLevel},
		{"bad reference", []any{"INVALID_REFERENCE"}, true, logrus.WarnLevel},
		{"internal", []any{"INTERNAL"}, true, logrus.ErrorLevel},
		{"a resolver failure carrying no code", []any{nil}, true, logrus.ErrorLevel},
		{"one internal among refusals", []any{"FORBIDDEN", "INTERNAL"}, true, logrus.ErrorLevel},
		{"a query the schema rejects", []any{errcode.ValidationFailed}, true, logrus.WarnLevel},
		{"a query that will not parse", []any{errcode.ParseFailed}, true, logrus.WarnLevel},
		{"a query over the complexity limit", []any{"COMPLEXITY_LIMIT_EXCEEDED"}, true, logrus.WarnLevel},
		{"an unresolved persisted query", []any{"PERSISTED_QUERY_NOT_FOUND"}, true, logrus.WarnLevel},
		{"a body gqlgen could not read", []any{nil}, false, logrus.WarnLevel},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			hook := test.NewGlobal()
			defer hook.Reset()
			logrus.SetLevel(logrus.DebugLevel)

			errs := gqlerror.List{}
			for _, code := range tc.codes {
				err := &gqlerror.Error{Message: "boom"}
				if code != nil {
					err.Extensions = map[string]any{"code": code}
				}
				errs = append(errs, err)
			}

			ctx := context.Background()
			if tc.withOperation {
				ctx = graphql.WithOperationContext(ctx, &graphql.OperationContext{})
			}

			response := &graphql.Response{Errors: errs}
			got := WithGqlLogger("test")(ctx, func(context.Context) *graphql.Response { return response })

			assert.Same(t, response, got)
			require.NotEmpty(t, hook.Entries)
			assert.Equal(t, tc.level, hook.LastEntry().Level)
		})
	}
}
