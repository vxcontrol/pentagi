package response

// general

var ErrInternal = NewHttpError(500, "Internal", "internal server error")
var ErrInternalDBNotFound = NewHttpError(500, "Internal.DBNotFound", "db not found")
var ErrInternalServiceNotFound = NewHttpError(500, "Internal.ServiceNotFound", "service not found")
var ErrInternalDBEncryptorNotFound = NewHttpError(500, "Internal.DBEncryptorNotFound", "DBEncryptor not found")
var ErrNotPermitted = NewHttpError(403, "NotPermitted", "action not permitted")
var ErrAuthRequired = NewHttpError(403, "AuthRequired", "auth required")
var ErrLocalUserRequired = NewHttpError(403, "LocalUserRequired", "local user required")
var ErrPrivilegesRequired = NewHttpError(403, "PrivilegesRequired", "some privileges required")
var ErrAdminRequired = NewHttpError(403, "AdminRequired", "admin required")
var ErrSuperRequired = NewHttpError(403, "SuperRequired", "super admin required")

// auth

var ErrAuthInvalidLoginRequest = NewHttpError(400, "Auth.InvalidLoginRequest", "invalid login data")
var ErrAuthInvalidAuthorizeQuery = NewHttpError(400, "Auth.InvalidAuthorizeQuery", "invalid authorize query")
var ErrAuthInvalidLoginCallbackRequest = NewHttpError(400, "Auth.InvalidLoginCallbackRequest", "invalid login callback data")
var ErrAuthInvalidAuthorizationState = NewHttpError(400, "Auth.InvalidAuthorizationState", "invalid authorization state data")
var ErrAuthInvalidSwitchServiceHash = NewHttpError(400, "Auth.InvalidSwitchServiceHash", "invalid switch service hash input data")
var ErrAuthInvalidAuthorizationNonce = NewHttpError(400, "Auth.InvalidAuthorizationNonce", "invalid authorization nonce data")
var ErrAuthInvalidCredentials = NewHttpError(401, "Auth.InvalidCredentials", "invalid login or password")
var ErrAuthInvalidUserData = NewHttpError(500, "Auth.InvalidUserData", "invalid user data")
var ErrAuthInactiveUser = NewHttpError(403, "Auth.InactiveUser", "user is inactive")
var ErrAuthPasswordChangeRequired = NewHttpError(403, "Auth.PasswordChangeRequired", "password change required")
var ErrAuthExchangeTokenFail = NewHttpError(403, "Auth.ExchangeTokenFail", "error on exchanging token")
var ErrAuthTokenExpired = NewHttpError(403, "Auth.TokenExpired", "token is expired")
var ErrAuthVerificationTokenFail = NewHttpError(403, "Auth.VerificationTokenFail", "error on verifying token")
var ErrAuthInvalidServiceData = NewHttpError(500, "Auth.InvalidServiceData", "invalid service data")
var ErrAuthInvalidTenantData = NewHttpError(500, "Auth.InvalidTenantData", "invalid tenant data")

// info

var ErrInfoUserNotFound = NewHttpError(404, "Info.UserNotFound", "user not found")
var ErrInfoInvalidUserData = NewHttpError(500, "Info.InvalidUserData", "invalid user data")
var ErrInfoInvalidServiceData = NewHttpError(500, "Info.InvalidServiceData", "invalid service data")

// proto

var ErrProtoInvalidRequest = NewHttpError(400, "Proto.InvalidRequest", "failed to validate auth token request")
var ErrProtoCreateTokenFail = NewHttpError(400, "Proto.CreateTokenFail", "failed to make auth token")
var ErrProtoInvalidToken = NewHttpError(400, "Proto.InvalidToken", "failed to valid auth token")

// users

var ErrUsersNotFound = NewHttpError(404, "Users.NotFound", "user not found")
var ErrUsersInvalidData = NewHttpError(500, "Users.InvalidData", "invalid user data")
var ErrUsersInvalidRequest = NewHttpError(400, "Users.InvalidRequest", "invalid user request data")
var ErrChangePasswordCurrentUserInvalidPassword = NewHttpError(400, "Users.ChangePasswordCurrentUser.InvalidPassword", "failed to validate user password")
var ErrChangePasswordCurrentUserInvalidCurrentPassword = NewHttpError(403, "Users.ChangePasswordCurrentUser.InvalidCurrentPassword", "invalid current password")
var ErrChangePasswordCurrentUserInvalidNewPassword = NewHttpError(400, "Users.ChangePasswordCurrentUser.InvalidNewPassword", "invalid new password form data")
var ErrGetUserModelsNotFound = NewHttpError(404, "Users.GetUser.ModelsNotFound", "user linked models not found")
var ErrCreateUserInvalidUser = NewHttpError(400, "Users.CreateUser.InvalidUser", "failed to validate user")
var ErrPatchUserModelsNotFound = NewHttpError(404, "Users.PatchUser.ModelsNotFound", "user linked models not found")
var ErrDeleteUserModelsNotFound = NewHttpError(404, "Users.DeleteUser.ModelsNotFound", "user linked models not found")

// roles

var ErrRolesInvalidRequest = NewHttpError(400, "Roles.InvalidRequest", "invalid role request data")
var ErrRolesInvalidData = NewHttpError(500, "Roles.InvalidData", "invalid role data")
var ErrRolesNotFound = NewHttpError(404, "Roles.NotFound", "role not found")

// prompts

var ErrPromptsInvalidRequest = NewHttpError(400, "Prompts.InvalidRequest", "invalid prompt request data")
var ErrPromptsInvalidData = NewHttpError(500, "Prompts.InvalidData", "invalid prompt data")
var ErrPromptsNotFound = NewHttpError(404, "Prompts.NotFound", "prompt not found")

// screenshots

var ErrScreenshotsInvalidRequest = NewHttpError(400, "Screenshots.InvalidRequest", "invalid screenshot request data")
var ErrScreenshotsNotFound = NewHttpError(404, "Screenshots.NotFound", "screenshot not found")
var ErrScreenshotsInvalidData = NewHttpError(500, "Screenshots.InvalidData", "invalid screenshot data")

// containers

var ErrContainersInvalidRequest = NewHttpError(400, "Containers.InvalidRequest", "invalid container request data")
var ErrContainersNotFound = NewHttpError(404, "Containers.NotFound", "container not found")
var ErrContainersInvalidData = NewHttpError(500, "Containers.InvalidData", "invalid container data")

// agentlogs

var ErrAgentlogsInvalidRequest = NewHttpError(400, "Agentlogs.InvalidRequest", "invalid agentlog request data")
var ErrAgentlogsInvalidData = NewHttpError(500, "Agentlogs.InvalidData", "invalid agentlog data")

// assistantlogs

var ErrAssistantlogsInvalidRequest = NewHttpError(400, "Assistantlogs.InvalidRequest", "invalid assistantlog request data")
var ErrAssistantlogsInvalidData = NewHttpError(500, "Assistantlogs.InvalidData", "invalid assistantlog data")

// msglogs

var ErrMsglogsInvalidRequest = NewHttpError(400, "Msglogs.InvalidRequest", "invalid msglog request data")
var ErrMsglogsInvalidData = NewHttpError(500, "Msglogs.InvalidData", "invalid msglog data")

// searchlogs

var ErrSearchlogsInvalidRequest = NewHttpError(400, "Searchlogs.InvalidRequest", "invalid searchlog request data")
var ErrSearchlogsInvalidData = NewHttpError(500, "Searchlogs.InvalidData", "invalid searchlog data")

// termlogs

var ErrTermlogsInvalidRequest = NewHttpError(400, "Termlogs.InvalidRequest", "invalid termlog request data")
var ErrTermlogsInvalidData = NewHttpError(500, "Termlogs.InvalidData", "invalid termlog data")

// vecstorelogs

var ErrVecstorelogsInvalidRequest = NewHttpError(400, "Vecstorelogs.InvalidRequest", "invalid vecstorelog request data")
var ErrVecstorelogsInvalidData = NewHttpError(500, "Vecstorelogs.InvalidData", "invalid vecstorelog data")

// flows

var ErrFlowsInvalidRequest = NewHttpError(400, "Flows.InvalidRequest", "invalid flow request data")
var ErrFlowsNotFound = NewHttpError(404, "Flows.NotFound", "flow not found")
var ErrFlowsInvalidData = NewHttpError(500, "Flows.InvalidData", "invalid flow data")

// tasks

var ErrTasksInvalidRequest = NewHttpError(400, "Tasks.InvalidRequest", "invalid task request data")
var ErrTasksNotFound = NewHttpError(404, "Tasks.NotFound", "task not found")
var ErrTasksInvalidData = NewHttpError(500, "Tasks.InvalidData", "invalid task data")

// subtasks

var ErrSubtasksInvalidRequest = NewHttpError(400, "Subtasks.InvalidRequest", "invalid subtask request data")
var ErrSubtasksNotFound = NewHttpError(404, "Subtasks.NotFound", "subtask not found")
var ErrSubtasksInvalidData = NewHttpError(500, "Subtasks.InvalidData", "invalid subtask data")

// assistants

var ErrAssistantsInvalidRequest = NewHttpError(400, "Assistants.InvalidRequest", "invalid assistant request data")
var ErrAssistantsNotFound = NewHttpError(404, "Assistants.NotFound", "assistant not found")
var ErrAssistantsInvalidData = NewHttpError(500, "Assistants.InvalidData", "invalid assistant data")
