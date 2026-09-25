package response

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrors_DeclaredErrorsCarryTheirStatusAndCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		err      *HttpError
		httpCode int
		code     string
	}{
		{ErrInternal, 500, "Internal"},
		{ErrInternalDBNotFound, 500, "Internal.DBNotFound"},
		{ErrInternalServiceNotFound, 500, "Internal.ServiceNotFound"},
		{ErrInternalDBEncryptorNotFound, 500, "Internal.DBEncryptorNotFound"},
		{ErrNotPermitted, 403, "NotPermitted"},
		{ErrAuthRequired, 403, "AuthRequired"},
		{ErrAuthUnavailable, 503, "AuthUnavailable"},
		{ErrLocalUserRequired, 403, "LocalUserRequired"},
		{ErrPrivilegesRequired, 403, "PrivilegesRequired"},
		{ErrAdminRequired, 403, "AdminRequired"},
		{ErrSuperRequired, 403, "SuperRequired"},

		{ErrAuthInvalidLoginRequest, 400, "Auth.InvalidLoginRequest"},
		{ErrAuthInvalidAuthorizeQuery, 400, "Auth.InvalidAuthorizeQuery"},
		{ErrAuthInvalidLoginCallbackRequest, 400, "Auth.InvalidLoginCallbackRequest"},
		{ErrAuthInvalidAuthorizationState, 400, "Auth.InvalidAuthorizationState"},
		{ErrAuthInvalidSwitchServiceHash, 400, "Auth.InvalidSwitchServiceHash"},
		{ErrAuthInvalidAuthorizationNonce, 400, "Auth.InvalidAuthorizationNonce"},
		{ErrAuthInvalidCredentials, 401, "Auth.InvalidCredentials"},
		{ErrAuthTooManyAttempts, 429, "Auth.TooManyAttempts"},
		{ErrAuthInvalidUserData, 500, "Auth.InvalidUserData"},
		{ErrAuthInactiveUser, 403, "Auth.InactiveUser"},
		{ErrAuthExchangeTokenFail, 403, "Auth.ExchangeTokenFail"},
		{ErrAuthTokenExpired, 403, "Auth.TokenExpired"},
		{ErrAuthVerificationTokenFail, 403, "Auth.VerificationTokenFail"},
		{ErrAuthInvalidServiceData, 500, "Auth.InvalidServiceData"},
		{ErrAuthInvalidTenantData, 500, "Auth.InvalidTenantData"},

		{ErrInfoUserNotFound, 404, "Info.UserNotFound"},
		{ErrInfoInvalidUserData, 500, "Info.InvalidUserData"},
		{ErrInfoInvalidServiceData, 500, "Info.InvalidServiceData"},

		{ErrUsersNotFound, 404, "Users.NotFound"},
		{ErrUsersInvalidData, 500, "Users.InvalidData"},
		{ErrUsersInvalidRequest, 400, "Users.InvalidRequest"},
		{ErrChangePasswordCurrentUserInvalidPassword, 400, "Users.ChangePasswordCurrentUser.InvalidPassword"},
		{ErrChangePasswordCurrentUserInvalidCurrentPassword, 403, "Users.ChangePasswordCurrentUser.InvalidCurrentPassword"},
		{ErrChangePasswordCurrentUserInvalidNewPassword, 400, "Users.ChangePasswordCurrentUser.InvalidNewPassword"},
		{ErrChangeEmailCurrentUserInvalidEmail, 400, "Users.ChangeEmailCurrentUser.InvalidEmail"},
		{ErrChangeEmailCurrentUserInvalidCurrentPassword, 403, "Users.ChangeEmailCurrentUser.InvalidCurrentPassword"},
		{ErrChangeEmailCurrentUserEmailAlreadyExists, 409, "Users.ChangeEmailCurrentUser.EmailAlreadyExists"},
		{ErrGetUserModelsNotFound, 404, "Users.GetUser.ModelsNotFound"},
		{ErrCreateUserInvalidUser, 400, "Users.CreateUser.InvalidUser"},
		{ErrPatchUserModelsNotFound, 404, "Users.PatchUser.ModelsNotFound"},
		{ErrDeleteUserModelsNotFound, 404, "Users.DeleteUser.ModelsNotFound"},

		{ErrRolesInvalidRequest, 400, "Roles.InvalidRequest"},
		{ErrRolesInvalidData, 500, "Roles.InvalidData"},
		{ErrRolesNotFound, 404, "Roles.NotFound"},

		{ErrPromptsInvalidRequest, 400, "Prompts.InvalidRequest"},
		{ErrPromptsInvalidData, 500, "Prompts.InvalidData"},
		{ErrPromptsNotFound, 404, "Prompts.NotFound"},

		{ErrScreenshotsInvalidRequest, 400, "Screenshots.InvalidRequest"},
		{ErrScreenshotsNotFound, 404, "Screenshots.NotFound"},
		{ErrScreenshotsInvalidData, 500, "Screenshots.InvalidData"},

		{ErrContainersInvalidRequest, 400, "Containers.InvalidRequest"},
		{ErrContainersNotFound, 404, "Containers.NotFound"},
		{ErrContainersInvalidData, 500, "Containers.InvalidData"},

		{ErrAgentlogsInvalidRequest, 400, "Agentlogs.InvalidRequest"},
		{ErrAgentlogsInvalidData, 500, "Agentlogs.InvalidData"},

		{ErrAssistantlogsInvalidRequest, 400, "Assistantlogs.InvalidRequest"},
		{ErrAssistantlogsInvalidData, 500, "Assistantlogs.InvalidData"},

		{ErrMsglogsInvalidRequest, 400, "Msglogs.InvalidRequest"},
		{ErrMsglogsInvalidData, 500, "Msglogs.InvalidData"},

		{ErrSearchlogsInvalidRequest, 400, "Searchlogs.InvalidRequest"},
		{ErrSearchlogsInvalidData, 500, "Searchlogs.InvalidData"},

		{ErrTermlogsInvalidRequest, 400, "Termlogs.InvalidRequest"},
		{ErrTermlogsInvalidData, 500, "Termlogs.InvalidData"},

		{ErrVecstorelogsInvalidRequest, 400, "Vecstorelogs.InvalidRequest"},
		{ErrVecstorelogsInvalidData, 500, "Vecstorelogs.InvalidData"},

		{ErrFlowsInvalidRequest, 400, "Flows.InvalidRequest"},
		{ErrFlowsNotFound, 404, "Flows.NotFound"},
		{ErrFlowsInvalidData, 500, "Flows.InvalidData"},

		{ErrTasksInvalidRequest, 400, "Tasks.InvalidRequest"},
		{ErrTasksNotFound, 404, "Tasks.NotFound"},
		{ErrTasksInvalidData, 500, "Tasks.InvalidData"},

		{ErrSubtasksInvalidRequest, 400, "Subtasks.InvalidRequest"},
		{ErrSubtasksNotFound, 404, "Subtasks.NotFound"},
		{ErrSubtasksInvalidData, 500, "Subtasks.InvalidData"},

		{ErrAssistantsInvalidRequest, 400, "Assistants.InvalidRequest"},
		{ErrAssistantsNotFound, 404, "Assistants.NotFound"},
		{ErrAssistantsInvalidData, 500, "Assistants.InvalidData"},

		{ErrTokenCreationDisabled, 400, "Token.CreationDisabled"},
		{ErrTokenNotFound, 404, "Token.NotFound"},
		{ErrTokenUnauthorized, 403, "Token.Unauthorized"},
		{ErrTokenInvalidRequest, 400, "Token.InvalidRequest"},
		{ErrTokenInvalidData, 500, "Token.InvalidData"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.httpCode, tt.err.HttpCode(), tt.code)
		assert.Equal(t, tt.code, tt.err.Code())
		assert.NotEmpty(t, tt.err.Msg(), tt.code)
	}
}
