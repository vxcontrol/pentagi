package response

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"pentagi/pkg/version"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestNewHttpError(t *testing.T) {
	err := NewHttpError(404, "NotFound", "resource not found")
	assert.Equal(t, 404, err.HttpCode())
	assert.Equal(t, "NotFound", err.Code())
	assert.Equal(t, "resource not found", err.Msg())
}

func TestHttpError_Error(t *testing.T) {
	err := NewHttpError(500, "Internal", "something broke")
	assert.Equal(t, "Internal: something broke", err.Error())
}

func TestHttpError_ImplementsError(t *testing.T) {
	var err error = NewHttpError(400, "Bad", "bad request")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Bad")
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      *HttpError
		httpCode int
		code     string
	}{
		{"ErrInternal", ErrInternal, 500, "Internal"},
		{"ErrNotPermitted", ErrNotPermitted, 403, "NotPermitted"},
		{"ErrAuthRequired", ErrAuthRequired, 403, "AuthRequired"},
		{"ErrAuthInvalidCredentials", ErrAuthInvalidCredentials, 401, "Auth.InvalidCredentials"},
		{"ErrUsersNotFound", ErrUsersNotFound, 404, "Users.NotFound"},
		{"ErrRolesNotFound", ErrRolesNotFound, 404, "Roles.NotFound"},
		{"ErrFlowsNotFound", ErrFlowsNotFound, 404, "Flows.NotFound"},
		{"ErrTasksNotFound", ErrTasksNotFound, 404, "Tasks.NotFound"},
		{"ErrTokenNotFound", ErrTokenNotFound, 404, "Token.NotFound"},
		{"ErrContainersNotFound", ErrContainersNotFound, 404, "Containers.NotFound"},
		{"ErrPromptsInvalidRequest", ErrPromptsInvalidRequest, 400, "Prompts.InvalidRequest"},
		{"ErrTokenCreationDisabled", ErrTokenCreationDisabled, 400, "Token.CreationDisabled"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.httpCode, tt.err.HttpCode())
			assert.Equal(t, tt.code, tt.err.Code())
			assert.NotEmpty(t, tt.err.Msg())
			assert.NotEmpty(t, tt.err.Error())
		})
	}
}

func TestSuccessResponse(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	data := map[string]string{"id": "123"}
	Success(c, http.StatusOK, data)

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, "success", body["status"])
	assert.NotNil(t, body["data"])
}

func TestSuccessResponse_Created(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Success(c, http.StatusCreated, gin.H{"name": "test"})

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestErrorResponse(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

	Error(c, ErrInternal, errors.New("db connection failed"))

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var body map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, "error", body["status"])
	assert.Equal(t, "Internal", body["code"])
	assert.Equal(t, "internal server error", body["msg"])
}

func TestErrorResponse_DevMode(t *testing.T) {
	// Enable dev mode
	version.PackageVer = ""

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

	originalErr := errors.New("detailed error info")
	Error(c, ErrInternal, originalErr)

	var body map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)

	// In dev mode, original error should be included
	assert.Equal(t, "detailed error info", body["error"])
}

func TestErrorResponse_ProductionMode(t *testing.T) {
	// Set production mode
	version.PackageVer = "1.0.0"
	defer func() { version.PackageVer = "" }()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

	Error(c, ErrInternal, errors.New("should not appear"))

	var body map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)

	// In production mode, original error should NOT be included
	_, hasError := body["error"]
	assert.False(t, hasError)
}

func TestErrorResponse_NilOriginalError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

	Error(c, ErrNotPermitted, nil)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var body map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, "NotPermitted", body["code"])
}
