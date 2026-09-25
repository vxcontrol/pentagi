package response

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"pentagi/pkg/version"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func httpContext() (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

	return c, w
}

// httpBuild makes the process a development build (no version) or a release build for one test.
func httpBuild(t *testing.T, develop bool) {
	t.Helper()

	previous := version.PackageVer
	t.Cleanup(func() { version.PackageVer = previous })

	version.PackageVer = "1.0.0"
	if develop {
		version.PackageVer = ""
	}
}

func TestHTTP_HttpError_KeepsItsStatusCodeAndMessage(t *testing.T) {
	t.Parallel()

	first := NewHttpError(404, "NotFound", "resource 1 not found")
	second := NewHttpError(404, "NotFound", "resource 2 not found")

	assert.Equal(t, 404, first.HttpCode())
	assert.Equal(t, "NotFound", first.Code())
	assert.Equal(t, "resource 1 not found", first.Msg())
	assert.Equal(t, "resource 2 not found", second.Msg())

	var err error = first
	assert.Equal(t, "NotFound: resource 1 not found", err.Error())
}

func TestHTTP_Success_WrapsTheDataInASuccessEnvelope(t *testing.T) {
	t.Parallel()

	nested := gin.H{
		"users": []gin.H{{"id": 1, "name": "Alice"}, {"id": 2, "name": "Bob"}},
		"count": 2,
		"meta":  gin.H{"page": 1, "total": 100},
	}

	tests := []struct {
		name   string
		status int
		data   any
		body   string
	}{
		{"an object", http.StatusOK, map[string]string{"id": "123"}, `{"status":"success","data":{"id":"123"}}`},
		{"a created resource", http.StatusCreated, gin.H{"name": "test"}, `{"status":"success","data":{"name":"test"}}`},
		{"an accepted request", http.StatusAccepted, gin.H{"test": "data"}, `{"status":"success","data":{"test":"data"}}`},
		{"no content carries no body", http.StatusNoContent, gin.H{"test": "data"}, ""},
		{"no data", http.StatusOK, nil, `{"status":"success","data":null}`},
		{"nested data", http.StatusOK, nested, `{"status":"success","data":{"count":2,"meta":{"page":1,"total":100},` +
			`"users":[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}]}}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c, w := httpContext()
			Success(c, tt.status, tt.data)

			assert.Equal(t, tt.status, w.Code)
			if tt.body == "" {
				assert.Empty(t, w.Body.String())
			} else {
				assert.JSONEq(t, tt.body, w.Body.String())
			}
		})
	}
}

func TestHTTP_Error_AnswersWithTheStatusBodyAndLevelOfTheError(t *testing.T) {
	const (
		internal = `{"status":"error","code":"Internal","msg":"internal server error"}`
		ranOut   = `{"status":"error","code":"RequestTimeout","msg":"the server took too long to answer; the action may still be running"}`
	)

	tests := []struct {
		name     string
		err      *HttpError
		original error
		develop  bool
		status   int
		body     string
		level    logrus.Level
	}{
		{"bad request", NewHttpError(400, "BadRequest", "bad request"), nil, false,
			400, `{"status":"error","code":"BadRequest","msg":"bad request"}`, logrus.WarnLevel},
		{"unauthorized", NewHttpError(401, "Unauthorized", "unauthorized"), nil, false,
			401, `{"status":"error","code":"Unauthorized","msg":"unauthorized"}`, logrus.WarnLevel},
		{"forbidden", ErrNotPermitted, nil, false,
			403, `{"status":"error","code":"NotPermitted","msg":"action not permitted"}`, logrus.WarnLevel},
		{"not found", ErrUsersNotFound, nil, false,
			404, `{"status":"error","code":"Users.NotFound","msg":"user not found"}`, logrus.WarnLevel},
		{"conflict", NewHttpError(409, "Conflict", "already exists"), nil, false,
			409, `{"status":"error","code":"Conflict","msg":"already exists"}`, logrus.WarnLevel},
		{"unavailable", NewHttpError(503, "Unavailable", "service unavailable"), nil, false,
			503, `{"status":"error","code":"Unavailable","msg":"service unavailable"}`, logrus.ErrorLevel},
		{"internal without a cause", ErrInternal, nil, false, 500, internal, logrus.ErrorLevel},
		{"a release build keeps the cause to itself", ErrInternal, errors.New("db connection failed"), false,
			500, internal, logrus.ErrorLevel},
		{"a development build shows the cause", ErrInternal, errors.New("detailed error info"), true,
			500, `{"status":"error","code":"Internal","msg":"internal server error","error":"detailed error info"}`,
			logrus.ErrorLevel},
		{"another postgres failure stays internal", ErrInternal,
			&pq.Error{Code: "23505", Message: "duplicate key value"}, false, 500, internal, logrus.ErrorLevel},
		{"a statement postgres ended at the ceiling", ErrInternal,
			&pq.Error{Code: "57014", Message: "canceling statement due to statement timeout"}, false, 504, ranOut, logrus.ErrorLevel},
		{"a statement ended behind a wrapper", ErrInternal,
			fmt.Errorf("selecting flows: %w", &pq.Error{Code: "57014"}), false, 504, ranOut, logrus.ErrorLevel},
		{"a context that expired", ErrInternal, context.DeadlineExceeded, false, 504, ranOut, logrus.ErrorLevel},
	}

	hook := test.NewGlobal()
	defer hook.Reset()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpBuild(t, tt.develop)
			hook.Reset()

			c, w := httpContext()
			Error(c, tt.err, tt.original)

			assert.Equal(t, tt.status, w.Code)
			assert.JSONEq(t, tt.body, w.Body.String())
			require.NotEmpty(t, hook.Entries)
			assert.Equal(t, tt.level, hook.LastEntry().Level)
			assert.Equal(t, "api error", hook.LastEntry().Message)
		})
	}
}

func TestHTTP_ErrorWithLevel_LogsAtTheGivenLevelRatherThanTheStatusClass(t *testing.T) {
	hook := test.NewGlobal()
	defer hook.Reset()

	c, w := httpContext()
	ErrorWithLevel(c, ErrAuthRequired, errors.New("cookie claim invalid"), logrus.InfoLevel)

	require.NotEmpty(t, hook.Entries)
	assert.Equal(t, logrus.InfoLevel, hook.LastEntry().Level)
	assert.Equal(t, "api error", hook.LastEntry().Message)
	assert.Equal(t, http.StatusForbidden, w.Code)
}
