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
	t.Parallel()

	err := NewHttpError(404, "NotFound", "resource not found")
	assert.Equal(t, 404, err.HttpCode())
	assert.Equal(t, "NotFound", err.Code())
	assert.Equal(t, "resource not found", err.Msg())
}

func TestHttpError_Error(t *testing.T) {
	t.Parallel()

	err := NewHttpError(500, "Internal", "something broke")
	assert.Equal(t, "Internal: something broke", err.Error())
}

func TestHttpError_ImplementsError(t *testing.T) {
	t.Parallel()

	var err error = NewHttpError(400, "Bad", "bad request")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Bad")
}

func TestPredefinedErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      *HttpError
		httpCode int
		code     string
	}{
		// General errors
		{"ErrInternal", ErrInternal, 500, "Internal"},
		{"ErrInternalDBNotFound", ErrInternalDBNotFound, 500, "Internal.DBNotFound"},
		{"ErrInternalServiceNotFound", ErrInternalServiceNotFound, 500, "Internal.ServiceNotFound"},
		{"ErrInternalDBEncryptorNotFound", ErrInternalDBEncryptorNotFound, 500, "Internal.DBEncryptorNotFound"},
		{"ErrNotPermitted", ErrNotPermitted, 403, "NotPermitted"},
		{"ErrAuthRequired", ErrAuthRequired, 403, "AuthRequired"},
		{"ErrLocalUserRequired", ErrLocalUserRequired, 403, "LocalUserRequired"},
		{"ErrPrivilegesRequired", ErrPrivilegesRequired, 403, "PrivilegesRequired"},
		{"ErrAdminRequired", ErrAdminRequired, 403, "AdminRequired"},
		{"ErrSuperRequired", ErrSuperRequired, 403, "SuperRequired"},

		// Auth errors
		{"ErrAuthInvalidLoginRequest", ErrAuthInvalidLoginRequest, 400, "Auth.InvalidLoginRequest"},
		{"ErrAuthInvalidAuthorizeQuery", ErrAuthInvalidAuthorizeQuery, 400, "Auth.InvalidAuthorizeQuery"},
		{"ErrAuthInvalidLoginCallbackRequest", ErrAuthInvalidLoginCallbackRequest, 400, "Auth.InvalidLoginCallbackRequest"},
		{"ErrAuthInvalidAuthorizationState", ErrAuthInvalidAuthorizationState, 400, "Auth.InvalidAuthorizationState"},
		{"ErrAuthInvalidSwitchServiceHash", ErrAuthInvalidSwitchServiceHash, 400, "Auth.InvalidSwitchServiceHash"},
		{"ErrAuthInvalidAuthorizationNonce", ErrAuthInvalidAuthorizationNonce, 400, "Auth.InvalidAuthorizationNonce"},
		{"ErrAuthInvalidCredentials", ErrAuthInvalidCredentials, 401, "Auth.InvalidCredentials"},
		{"ErrAuthInvalidUserData", ErrAuthInvalidUserData, 500, "Auth.InvalidUserData"},
		{"ErrAuthInactiveUser", ErrAuthInactiveUser, 403, "Auth.InactiveUser"},
		{"ErrAuthExchangeTokenFail", ErrAuthExchangeTokenFail, 403, "Auth.ExchangeTokenFail"},
		{"ErrAuthTokenExpired", ErrAuthTokenExpired, 403, "Auth.TokenExpired"},
		{"ErrAuthVerificationTokenFail", ErrAuthVerificationTokenFail, 403, "Auth.VerificationTokenFail"},
		{"ErrAuthInvalidServiceData", ErrAuthInvalidServiceData, 500, "Auth.InvalidServiceData"},
		{"ErrAuthInvalidTenantData", ErrAuthInvalidTenantData, 500, "Auth.InvalidTenantData"},

		// Info errors
		{"ErrInfoUserNotFound", ErrInfoUserNotFound, 404, "Info.UserNotFound"},
		{"ErrInfoInvalidUserData", ErrInfoInvalidUserData, 500, "Info.InvalidUserData"},
		{"ErrInfoInvalidServiceData", ErrInfoInvalidServiceData, 500, "Info.InvalidServiceData"},

		// Users errors
		{"ErrUsersNotFound", ErrUsersNotFound, 404, "Users.NotFound"},
		{"ErrUsersInvalidData", ErrUsersInvalidData, 500, "Users.InvalidData"},
		{"ErrUsersInvalidRequest", ErrUsersInvalidRequest, 400, "Users.InvalidRequest"},
		{"ErrChangePasswordCurrentUserInvalidPassword", ErrChangePasswordCurrentUserInvalidPassword, 400, "Users.ChangePasswordCurrentUser.InvalidPassword"},
		{"ErrChangePasswordCurrentUserInvalidCurrentPassword", ErrChangePasswordCurrentUserInvalidCurrentPassword, 403, "Users.ChangePasswordCurrentUser.InvalidCurrentPassword"},
		{"ErrChangePasswordCurrentUserInvalidNewPassword", ErrChangePasswordCurrentUserInvalidNewPassword, 400, "Users.ChangePasswordCurrentUser.InvalidNewPassword"},
		{"ErrGetUserModelsNotFound", ErrGetUserModelsNotFound, 404, "Users.GetUser.ModelsNotFound"},
		{"ErrCreateUserInvalidUser", ErrCreateUserInvalidUser, 400, "Users.CreateUser.InvalidUser"},
		{"ErrPatchUserModelsNotFound", ErrPatchUserModelsNotFound, 404, "Users.PatchUser.ModelsNotFound"},
		{"ErrDeleteUserModelsNotFound", ErrDeleteUserModelsNotFound, 404, "Users.DeleteUser.ModelsNotFound"},

		// Roles errors
		{"ErrRolesInvalidRequest", ErrRolesInvalidRequest, 400, "Roles.InvalidRequest"},
		{"ErrRolesInvalidData", ErrRolesInvalidData, 500, "Roles.InvalidData"},
		{"ErrRolesNotFound", ErrRolesNotFound, 404, "Roles.NotFound"},

		// Prompts errors
		{"ErrPromptsInvalidRequest", ErrPromptsInvalidRequest, 400, "Prompts.InvalidRequest"},
		{"ErrPromptsInvalidData", ErrPromptsInvalidData, 500, "Prompts.InvalidData"},
		{"ErrPromptsNotFound", ErrPromptsNotFound, 404, "Prompts.NotFound"},

		// Screenshots errors
		{"ErrScreenshotsInvalidRequest", ErrScreenshotsInvalidRequest, 400, "Screenshots.InvalidRequest"},
		{"ErrScreenshotsNotFound", ErrScreenshotsNotFound, 404, "Screenshots.NotFound"},
		{"ErrScreenshotsInvalidData", ErrScreenshotsInvalidData, 500, "Screenshots.InvalidData"},

		// Containers errors
		{"ErrContainersInvalidRequest", ErrContainersInvalidRequest, 400, "Containers.InvalidRequest"},
		{"ErrContainersNotFound", ErrContainersNotFound, 404, "Containers.NotFound"},
		{"ErrContainersInvalidData", ErrContainersInvalidData, 500, "Containers.InvalidData"},

		// Agentlogs errors
		{"ErrAgentlogsInvalidRequest", ErrAgentlogsInvalidRequest, 400, "Agentlogs.InvalidRequest"},
		{"ErrAgentlogsInvalidData", ErrAgentlogsInvalidData, 500, "Agentlogs.InvalidData"},

		// Assistantlogs errors
		{"ErrAssistantlogsInvalidRequest", ErrAssistantlogsInvalidRequest, 400, "Assistantlogs.InvalidRequest"},
		{"ErrAssistantlogsInvalidData", ErrAssistantlogsInvalidData, 500, "Assistantlogs.InvalidData"},

		// Msglogs errors
		{"ErrMsglogsInvalidRequest", ErrMsglogsInvalidRequest, 400, "Msglogs.InvalidRequest"},
		{"ErrMsglogsInvalidData", ErrMsglogsInvalidData, 500, "Msglogs.InvalidData"},

		// Searchlogs errors
		{"ErrSearchlogsInvalidRequest", ErrSearchlogsInvalidRequest, 400, "Searchlogs.InvalidRequest"},
		{"ErrSearchlogsInvalidData", ErrSearchlogsInvalidData, 500, "Searchlogs.InvalidData"},

		// Termlogs errors
		{"ErrTermlogsInvalidRequest", ErrTermlogsInvalidRequest, 400, "Termlogs.InvalidRequest"},
		{"ErrTermlogsInvalidData", ErrTermlogsInvalidData, 500, "Termlogs.InvalidData"},

		// Vecstorelogs errors
		{"ErrVecstorelogsInvalidRequest", ErrVecstorelogsInvalidRequest, 400, "Vecstorelogs.InvalidRequest"},
		{"ErrVecstorelogsInvalidData", ErrVecstorelogsInvalidData, 500, "Vecstorelogs.InvalidData"},

		// Flows errors
		{"ErrFlowsInvalidRequest", ErrFlowsInvalidRequest, 400, "Flows.InvalidRequest"},
		{"ErrFlowsNotFound", ErrFlowsNotFound, 404, "Flows.NotFound"},
		{"ErrFlowsInvalidData", ErrFlowsInvalidData, 500, "Flows.InvalidData"},

		// Tasks errors
		{"ErrTasksInvalidRequest", ErrTasksInvalidRequest, 400, "Tasks.InvalidRequest"},
		{"ErrTasksNotFound", ErrTasksNotFound, 404, "Tasks.NotFound"},
		{"ErrTasksInvalidData", ErrTasksInvalidData, 500, "Tasks.InvalidData"},

		// Subtasks errors
		{"ErrSubtasksInvalidRequest", ErrSubtasksInvalidRequest, 400, "Subtasks.InvalidRequest"},
		{"ErrSubtasksNotFound", ErrSubtasksNotFound, 404, "Subtasks.NotFound"},
		{"ErrSubtasksInvalidData", ErrSubtasksInvalidData, 500, "Subtasks.InvalidData"},

		// Assistants errors
		{"ErrAssistantsInvalidRequest", ErrAssistantsInvalidRequest, 400, "Assistants.InvalidRequest"},
		{"ErrAssistantsNotFound", ErrAssistantsNotFound, 404, "Assistants.NotFound"},
		{"ErrAssistantsInvalidData", ErrAssistantsInvalidData, 500, "Assistants.InvalidData"},

		// Tokens errors
		{"ErrTokenCreationDisabled", ErrTokenCreationDisabled, 400, "Token.CreationDisabled"},
		{"ErrTokenNotFound", ErrTokenNotFound, 404, "Token.NotFound"},
		{"ErrTokenUnauthorized", ErrTokenUnauthorized, 403, "Token.Unauthorized"},
		{"ErrTokenInvalidRequest", ErrTokenInvalidRequest, 400, "Token.InvalidRequest"},
		{"ErrTokenInvalidData", ErrTokenInvalidData, 500, "Token.InvalidData"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.httpCode, tt.err.HttpCode())
			assert.Equal(t, tt.code, tt.err.Code())
			assert.NotEmpty(t, tt.err.Msg())
			assert.NotEmpty(t, tt.err.Error())
		})
	}
}

func TestSuccessResponse(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Success(c, http.StatusCreated, gin.H{"name": "test"})

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestErrorResponse(t *testing.T) {
	t.Parallel()

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
	// Save original version and restore after test
	oldVer := version.PackageVer
	defer func() { version.PackageVer = oldVer }()

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
	// Save original version and restore after test
	oldVer := version.PackageVer
	defer func() { version.PackageVer = oldVer }()

	// Set production mode
	version.PackageVer = "1.0.0"

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
	t.Parallel()

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

func TestHttpError_MultipleInstancesIndependent(t *testing.T) {
	t.Parallel()

	err1 := NewHttpError(404, "NotFound", "resource 1 not found")
	err2 := NewHttpError(404, "NotFound", "resource 2 not found")

	// Verify they are independent instances
	assert.NotEqual(t, err1.Msg(), err2.Msg())
	assert.Equal(t, err1.Code(), err2.Code())
	assert.Equal(t, err1.HttpCode(), err2.HttpCode())
}

func TestSuccessResponse_EmptyData(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Success(c, http.StatusOK, nil)

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, "success", body["status"])
	assert.Nil(t, body["data"])
}

func TestSuccessResponse_ComplexData(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	data := gin.H{
		"users": []gin.H{
			{"id": 1, "name": "Alice"},
			{"id": 2, "name": "Bob"},
		},
		"count": 2,
		"meta": gin.H{
			"page":  1,
			"total": 100,
		},
	}

	Success(c, http.StatusOK, data)

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, "success", body["status"])

	responseData, ok := body["data"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, float64(2), responseData["count"])
}

func TestErrorResponse_DifferentHttpCodes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      *HttpError
		expected int
	}{
		{"400 Bad Request", ErrPromptsInvalidRequest, http.StatusBadRequest},
		{"401 Unauthorized", ErrAuthInvalidCredentials, http.StatusUnauthorized},
		{"403 Forbidden", ErrNotPermitted, http.StatusForbidden},
		{"404 Not Found", ErrUsersNotFound, http.StatusNotFound},
		{"500 Internal", ErrInternal, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

			Error(c, tt.err, nil)

			assert.Equal(t, tt.expected, w.Code)
		})
	}
}

func TestErrorResponse_ResponseStructure(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

	Error(c, ErrUsersNotFound, nil)

	var body map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)

	// Verify required fields
	assert.Equal(t, "error", body["status"])
	assert.Equal(t, "Users.NotFound", body["code"])
	assert.Equal(t, "user not found", body["msg"])

	// Verify error field is not present in non-dev mode
	_, hasError := body["error"]
	assert.False(t, hasError)
}

func TestSuccessResponse_StatusCodes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		statusCode   int
		expectedCode int
	}{
		{"200 OK", http.StatusOK, 200},
		{"201 Created", http.StatusCreated, 201},
		{"202 Accepted", http.StatusAccepted, 202},
		{"204 No Content", http.StatusNoContent, 204},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			Success(c, tt.statusCode, gin.H{"test": "data"})

			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}
