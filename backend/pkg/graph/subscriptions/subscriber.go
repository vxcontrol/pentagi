package subscriptions

import (
	"context"

	"pentagi/pkg/graph/model"
)

type flowSubscriber struct {
	userID int64
	flowID int64
	ctrl   *controller
}

func (s *flowSubscriber) GetFlowID() int64 {
	return s.flowID
}

func (s *flowSubscriber) SetFlowID(flowID int64) {
	s.flowID = flowID
}

func (s *flowSubscriber) GetUserID() int64 {
	return s.userID
}

func (s *flowSubscriber) SetUserID(userID int64) {
	s.userID = userID
}

func (s *flowSubscriber) FlowCreatedAdmin(ctx context.Context) (<-chan *model.Flow, error) {
	return s.ctrl.flowCreatedAdmin.Subscribe(ctx, s.userID), nil
}

func (s *flowSubscriber) FlowCreated(ctx context.Context) (<-chan *model.Flow, error) {
	return s.ctrl.flowCreated.Subscribe(ctx, s.userID), nil
}

func (s *flowSubscriber) FlowDeletedAdmin(ctx context.Context) (<-chan *model.Flow, error) {
	return s.ctrl.flowDeletedAdmin.Subscribe(ctx, s.userID), nil
}

func (s *flowSubscriber) FlowDeleted(ctx context.Context) (<-chan *model.Flow, error) {
	return s.ctrl.flowDeleted.Subscribe(ctx, s.userID), nil
}

func (s *flowSubscriber) FlowUpdatedAdmin(ctx context.Context) (<-chan *model.Flow, error) {
	return s.ctrl.flowUpdatedAdmin.Subscribe(ctx, s.userID), nil
}

func (s *flowSubscriber) FlowUpdated(ctx context.Context) (<-chan *model.Flow, error) {
	return s.ctrl.flowUpdated.Subscribe(ctx, s.userID), nil
}

func (s *flowSubscriber) TaskCreated(ctx context.Context) (<-chan *model.Task, error) {
	return s.ctrl.taskCreated.Subscribe(ctx, s.flowID), nil
}

func (s *flowSubscriber) TaskUpdated(ctx context.Context) (<-chan *model.Task, error) {
	return s.ctrl.taskUpdated.Subscribe(ctx, s.flowID), nil
}

func (s *flowSubscriber) AssistantCreated(ctx context.Context) (<-chan *model.Assistant, error) {
	return s.ctrl.assistantCreated.Subscribe(ctx, s.flowID), nil
}

func (s *flowSubscriber) AssistantUpdated(ctx context.Context) (<-chan *model.Assistant, error) {
	return s.ctrl.assistantUpdated.Subscribe(ctx, s.flowID), nil
}

func (s *flowSubscriber) AssistantDeleted(ctx context.Context) (<-chan *model.Assistant, error) {
	return s.ctrl.assistantDeleted.Subscribe(ctx, s.flowID), nil
}

func (s *flowSubscriber) FlowFileAdded(ctx context.Context) (<-chan *model.FlowFile, error) {
	return s.ctrl.flowFileAdded.Subscribe(ctx, s.flowID), nil
}

func (s *flowSubscriber) FlowFileUpdated(ctx context.Context) (<-chan *model.FlowFile, error) {
	return s.ctrl.flowFileUpdated.Subscribe(ctx, s.flowID), nil
}

func (s *flowSubscriber) FlowFileDeleted(ctx context.Context) (<-chan *model.FlowFile, error) {
	return s.ctrl.flowFileDeleted.Subscribe(ctx, s.flowID), nil
}

func (s *flowSubscriber) ScreenshotAdded(ctx context.Context) (<-chan *model.Screenshot, error) {
	return s.ctrl.screenshotAdded.Subscribe(ctx, s.flowID), nil
}

func (s *flowSubscriber) TerminalLogAdded(ctx context.Context) (<-chan *model.TerminalLog, error) {
	return s.ctrl.terminalLogAdded.Subscribe(ctx, s.flowID), nil
}

func (s *flowSubscriber) MessageLogAdded(ctx context.Context) (<-chan *model.MessageLog, error) {
	return s.ctrl.messageLogAdded.Subscribe(ctx, s.flowID), nil
}

func (s *flowSubscriber) MessageLogUpdated(ctx context.Context) (<-chan *model.MessageLog, error) {
	return s.ctrl.messageLogUpdated.Subscribe(ctx, s.flowID), nil
}

func (s *flowSubscriber) AgentLogAdded(ctx context.Context) (<-chan *model.AgentLog, error) {
	return s.ctrl.agentLogAdded.Subscribe(ctx, s.flowID), nil
}

func (s *flowSubscriber) SearchLogAdded(ctx context.Context) (<-chan *model.SearchLog, error) {
	return s.ctrl.searchLogAdded.Subscribe(ctx, s.flowID), nil
}

func (s *flowSubscriber) VectorStoreLogAdded(ctx context.Context) (<-chan *model.VectorStoreLog, error) {
	return s.ctrl.vecStoreLogAdded.Subscribe(ctx, s.flowID), nil
}

func (s *flowSubscriber) AssistantLogAdded(ctx context.Context) (<-chan *model.AssistantLog, error) {
	return s.ctrl.assistantLogAdded.Subscribe(ctx, s.flowID), nil
}

func (s *flowSubscriber) AssistantLogUpdated(ctx context.Context) (<-chan *model.AssistantLog, error) {
	return s.ctrl.assistantLogUpdated.Subscribe(ctx, s.flowID), nil
}

// providerSubscriber subscribes to user-scoped provider events.
type providerSubscriber struct {
	userID int64
	ctrl   *controller
}

func (s *providerSubscriber) GetUserID() int64 {
	return s.userID
}

func (s *providerSubscriber) SetUserID(userID int64) {
	s.userID = userID
}

func (s *providerSubscriber) ProviderCreated(ctx context.Context) (<-chan *model.ProviderConfig, error) {
	return s.ctrl.providerCreated.Subscribe(ctx, s.userID), nil
}

func (s *providerSubscriber) ProviderUpdated(ctx context.Context) (<-chan *model.ProviderConfig, error) {
	return s.ctrl.providerUpdated.Subscribe(ctx, s.userID), nil
}

func (s *providerSubscriber) ProviderDeleted(ctx context.Context) (<-chan *model.ProviderConfig, error) {
	return s.ctrl.providerDeleted.Subscribe(ctx, s.userID), nil
}

// apiTokenSubscriber subscribes to user-scoped API token events.
type apiTokenSubscriber struct {
	userID int64
	ctrl   *controller
}

func (s *apiTokenSubscriber) GetUserID() int64 {
	return s.userID
}

func (s *apiTokenSubscriber) SetUserID(userID int64) {
	s.userID = userID
}

func (s *apiTokenSubscriber) APITokenCreated(ctx context.Context) (<-chan *model.APIToken, error) {
	return s.ctrl.apiTokenCreated.Subscribe(ctx, s.userID), nil
}

func (s *apiTokenSubscriber) APITokenUpdated(ctx context.Context) (<-chan *model.APIToken, error) {
	return s.ctrl.apiTokenUpdated.Subscribe(ctx, s.userID), nil
}

func (s *apiTokenSubscriber) APITokenDeleted(ctx context.Context) (<-chan *model.APIToken, error) {
	return s.ctrl.apiTokenDeleted.Subscribe(ctx, s.userID), nil
}

// settingsSubscriber subscribes to user-scoped settings events.
type settingsSubscriber struct {
	userID int64
	ctrl   *controller
}

func (s *settingsSubscriber) GetUserID() int64 {
	return s.userID
}

func (s *settingsSubscriber) SetUserID(userID int64) {
	s.userID = userID
}

func (s *settingsSubscriber) SettingsUserUpdated(ctx context.Context) (<-chan *model.UserPreferences, error) {
	return s.ctrl.settingsUserUpdated.Subscribe(ctx, s.userID), nil
}

// flowTemplateSubscriber subscribes to user-scoped flow template events.
type flowTemplateSubscriber struct {
	userID int64
	ctrl   *controller
}

func (s *flowTemplateSubscriber) GetUserID() int64 {
	return s.userID
}

func (s *flowTemplateSubscriber) SetUserID(userID int64) {
	s.userID = userID
}

func (s *flowTemplateSubscriber) FlowTemplateCreated(ctx context.Context) (<-chan *model.FlowTemplate, error) {
	return s.ctrl.flowTemplateCreated.Subscribe(ctx, s.userID), nil
}

func (s *flowTemplateSubscriber) FlowTemplateUpdated(ctx context.Context) (<-chan *model.FlowTemplate, error) {
	return s.ctrl.flowTemplateUpdated.Subscribe(ctx, s.userID), nil
}

func (s *flowTemplateSubscriber) FlowTemplateDeleted(ctx context.Context) (<-chan *model.FlowTemplate, error) {
	return s.ctrl.flowTemplateDeleted.Subscribe(ctx, s.userID), nil
}

// resourceSubscriber subscribes to user-scoped resource events.
type resourceSubscriber struct {
	userID int64
	ctrl   *controller
}

func (s *resourceSubscriber) GetUserID() int64 {
	return s.userID
}

func (s *resourceSubscriber) SetUserID(userID int64) {
	s.userID = userID
}

func (s *resourceSubscriber) ResourceAdded(ctx context.Context) (<-chan *model.UserResource, error) {
	return s.ctrl.resourceAdded.Subscribe(ctx, s.userID), nil
}

func (s *resourceSubscriber) ResourceUpdated(ctx context.Context) (<-chan *model.UserResource, error) {
	return s.ctrl.resourceUpdated.Subscribe(ctx, s.userID), nil
}

func (s *resourceSubscriber) ResourceDeleted(ctx context.Context) (<-chan *model.UserResource, error) {
	return s.ctrl.resourceDeleted.Subscribe(ctx, s.userID), nil
}

func (s *resourceSubscriber) ResourceAddedAdmin(ctx context.Context) (<-chan *model.UserResource, error) {
	return s.ctrl.resourceAddedAdmin.Subscribe(ctx, s.userID), nil
}

func (s *resourceSubscriber) ResourceUpdatedAdmin(ctx context.Context) (<-chan *model.UserResource, error) {
	return s.ctrl.resourceUpdatedAdmin.Subscribe(ctx, s.userID), nil
}

func (s *resourceSubscriber) ResourceDeletedAdmin(ctx context.Context) (<-chan *model.UserResource, error) {
	return s.ctrl.resourceDeletedAdmin.Subscribe(ctx, s.userID), nil
}

// knowledgeSubscriber subscribes to global knowledge document events (broadcast to all users).
type knowledgeSubscriber struct {
	userID int64
	ctrl   *controller
}

func (s *knowledgeSubscriber) GetUserID() int64 {
	return s.userID
}

func (s *knowledgeSubscriber) SetUserID(userID int64) {
	s.userID = userID
}

func (s *knowledgeSubscriber) KnowledgeDocumentCreated(ctx context.Context) (<-chan *model.KnowledgeDocument, error) {
	return s.ctrl.knowledgeDocumentCreated.Subscribe(ctx, s.userID), nil
}

func (s *knowledgeSubscriber) KnowledgeDocumentUpdated(ctx context.Context) (<-chan *model.KnowledgeDocument, error) {
	return s.ctrl.knowledgeDocumentUpdated.Subscribe(ctx, s.userID), nil
}

func (s *knowledgeSubscriber) KnowledgeDocumentDeleted(ctx context.Context) (<-chan *model.KnowledgeDocument, error) {
	return s.ctrl.knowledgeDocumentDeleted.Subscribe(ctx, s.userID), nil
}

func (s *knowledgeSubscriber) KnowledgeDocumentCreatedAdmin(ctx context.Context) (<-chan *model.KnowledgeDocument, error) {
	return s.ctrl.knowledgeDocumentCreatedAdmin.Subscribe(ctx, s.userID), nil
}

func (s *knowledgeSubscriber) KnowledgeDocumentUpdatedAdmin(ctx context.Context) (<-chan *model.KnowledgeDocument, error) {
	return s.ctrl.knowledgeDocumentUpdatedAdmin.Subscribe(ctx, s.userID), nil
}

func (s *knowledgeSubscriber) KnowledgeDocumentDeletedAdmin(ctx context.Context) (<-chan *model.KnowledgeDocument, error) {
	return s.ctrl.knowledgeDocumentDeletedAdmin.Subscribe(ctx, s.userID), nil
}
