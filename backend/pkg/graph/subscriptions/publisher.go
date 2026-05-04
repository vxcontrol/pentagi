package subscriptions

import (
	"context"

	"pentagi/pkg/database"
	"pentagi/pkg/database/converter"
	"pentagi/pkg/graph/model"
	"pentagi/pkg/providers/pconfig"
)

type flowPublisher struct {
	flowID int64
	userID int64
	ctrl   *controller
}

func (p *flowPublisher) GetFlowID() int64 {
	return p.flowID
}

func (p *flowPublisher) SetFlowID(flowID int64) {
	p.flowID = flowID
}

func (p *flowPublisher) GetUserID() int64 {
	return p.userID
}

func (p *flowPublisher) SetUserID(userID int64) {
	p.userID = userID
}

func (p *flowPublisher) FlowCreated(ctx context.Context, flow database.Flow, terms []database.Container) {
	flowModel := converter.ConvertFlow(flow, terms)
	p.ctrl.flowCreated.Publish(ctx, p.userID, flowModel)
	p.ctrl.flowCreatedAdmin.Broadcast(ctx, flowModel)
}

func (p *flowPublisher) FlowDeleted(ctx context.Context, flow database.Flow, terms []database.Container) {
	flowModel := converter.ConvertFlow(flow, terms)
	p.ctrl.flowDeleted.Publish(ctx, p.userID, flowModel)
	p.ctrl.flowDeletedAdmin.Broadcast(ctx, flowModel)
}

func (p *flowPublisher) FlowUpdated(ctx context.Context, flow database.Flow, terms []database.Container) {
	flowModel := converter.ConvertFlow(flow, terms)
	p.ctrl.flowUpdated.Publish(ctx, p.userID, flowModel)
	p.ctrl.flowUpdatedAdmin.Broadcast(ctx, flowModel)
}

func (p *flowPublisher) TaskCreated(ctx context.Context, task database.Task, subtasks []database.Subtask) {
	p.ctrl.taskCreated.Publish(ctx, p.flowID, converter.ConvertTask(task, subtasks))
}

func (p *flowPublisher) TaskUpdated(ctx context.Context, task database.Task, subtasks []database.Subtask) {
	p.ctrl.taskUpdated.Publish(ctx, p.flowID, converter.ConvertTask(task, subtasks))
}

func (p *flowPublisher) AssistantCreated(ctx context.Context, assistant database.Assistant) {
	p.ctrl.assistantCreated.Publish(ctx, p.flowID, converter.ConvertAssistant(assistant))
}

func (p *flowPublisher) AssistantUpdated(ctx context.Context, assistant database.Assistant) {
	p.ctrl.assistantUpdated.Publish(ctx, p.flowID, converter.ConvertAssistant(assistant))
}

func (p *flowPublisher) AssistantDeleted(ctx context.Context, assistant database.Assistant) {
	p.ctrl.assistantDeleted.Publish(ctx, p.flowID, converter.ConvertAssistant(assistant))
}

func (p *flowPublisher) FlowFileAdded(ctx context.Context, file *model.FlowFile) {
	p.ctrl.flowFileAdded.Publish(ctx, p.flowID, file)
}

func (p *flowPublisher) FlowFileUpdated(ctx context.Context, file *model.FlowFile) {
	p.ctrl.flowFileUpdated.Publish(ctx, p.flowID, file)
}

func (p *flowPublisher) FlowFileDeleted(ctx context.Context, file *model.FlowFile) {
	p.ctrl.flowFileDeleted.Publish(ctx, p.flowID, file)
}

func (p *flowPublisher) ScreenshotAdded(ctx context.Context, screenshot database.Screenshot) {
	p.ctrl.screenshotAdded.Publish(ctx, p.flowID, converter.ConvertScreenshot(screenshot))
}

func (p *flowPublisher) TerminalLogAdded(ctx context.Context, terminalLog database.Termlog) {
	p.ctrl.terminalLogAdded.Publish(ctx, p.flowID, converter.ConvertTerminalLog(terminalLog))
}

func (p *flowPublisher) MessageLogAdded(ctx context.Context, messageLog database.Msglog) {
	p.ctrl.messageLogAdded.Publish(ctx, p.flowID, converter.ConvertMessageLog(messageLog))
}

func (p *flowPublisher) MessageLogUpdated(ctx context.Context, messageLog database.Msglog) {
	p.ctrl.messageLogUpdated.Publish(ctx, p.flowID, converter.ConvertMessageLog(messageLog))
}

func (p *flowPublisher) AgentLogAdded(ctx context.Context, agentLog database.Agentlog) {
	p.ctrl.agentLogAdded.Publish(ctx, p.flowID, converter.ConvertAgentLog(agentLog))
}

func (p *flowPublisher) SearchLogAdded(ctx context.Context, searchLog database.Searchlog) {
	p.ctrl.searchLogAdded.Publish(ctx, p.flowID, converter.ConvertSearchLog(searchLog))
}

func (p *flowPublisher) VectorStoreLogAdded(ctx context.Context, vectorStoreLog database.Vecstorelog) {
	p.ctrl.vecStoreLogAdded.Publish(ctx, p.flowID, converter.ConvertVectorStoreLog(vectorStoreLog))
}

func (p *flowPublisher) AssistantLogAdded(ctx context.Context, assistantLog database.Assistantlog) {
	p.ctrl.assistantLogAdded.Publish(ctx, p.flowID, converter.ConvertAssistantLog(assistantLog, false))
}

func (p *flowPublisher) AssistantLogUpdated(ctx context.Context, assistantLog database.Assistantlog, appendPart bool) {
	p.ctrl.assistantLogUpdated.Publish(ctx, p.flowID, converter.ConvertAssistantLog(assistantLog, appendPart))
}

func (p *flowPublisher) KnowledgeDocumentCreated(ctx context.Context, doc *model.KnowledgeDocument) {
	p.ctrl.knowledgeDocumentCreated.Publish(ctx, p.userID, doc)
	p.ctrl.knowledgeDocumentCreatedAdmin.Broadcast(ctx, doc)
}

// providerPublisher publishes user-scoped provider events.
type providerPublisher struct {
	userID int64
	ctrl   *controller
}

func (p *providerPublisher) GetUserID() int64 {
	return p.userID
}

func (p *providerPublisher) SetUserID(userID int64) {
	p.userID = userID
}

func (p *providerPublisher) ProviderCreated(ctx context.Context, provider database.Provider, cfg *pconfig.ProviderConfig) {
	p.ctrl.providerCreated.Publish(ctx, p.userID, converter.ConvertProvider(provider, cfg))
}

func (p *providerPublisher) ProviderUpdated(ctx context.Context, provider database.Provider, cfg *pconfig.ProviderConfig) {
	p.ctrl.providerUpdated.Publish(ctx, p.userID, converter.ConvertProvider(provider, cfg))
}

func (p *providerPublisher) ProviderDeleted(ctx context.Context, provider database.Provider, cfg *pconfig.ProviderConfig) {
	p.ctrl.providerDeleted.Publish(ctx, p.userID, converter.ConvertProvider(provider, cfg))
}

// apiTokenPublisher publishes user-scoped API token events.
type apiTokenPublisher struct {
	userID int64
	ctrl   *controller
}

func (p *apiTokenPublisher) GetUserID() int64 {
	return p.userID
}

func (p *apiTokenPublisher) SetUserID(userID int64) {
	p.userID = userID
}

func (p *apiTokenPublisher) APITokenCreated(ctx context.Context, apiToken database.APITokenWithSecret) {
	p.ctrl.apiTokenCreated.Publish(ctx, p.userID, converter.ConvertAPITokenRemoveSecret(apiToken))
}

func (p *apiTokenPublisher) APITokenUpdated(ctx context.Context, apiToken database.ApiToken) {
	p.ctrl.apiTokenUpdated.Publish(ctx, p.userID, converter.ConvertAPIToken(apiToken))
}

func (p *apiTokenPublisher) APITokenDeleted(ctx context.Context, apiToken database.ApiToken) {
	p.ctrl.apiTokenDeleted.Publish(ctx, p.userID, converter.ConvertAPIToken(apiToken))
}

// settingsPublisher publishes user-scoped settings events.
type settingsPublisher struct {
	userID int64
	ctrl   *controller
}

func (p *settingsPublisher) GetUserID() int64 {
	return p.userID
}

func (p *settingsPublisher) SetUserID(userID int64) {
	p.userID = userID
}

func (p *settingsPublisher) SettingsUserUpdated(ctx context.Context, userPreferences database.UserPreference) {
	p.ctrl.settingsUserUpdated.Publish(ctx, p.userID, converter.ConvertUserPreferences(userPreferences))
}

// flowTemplatePublisher publishes user-scoped flow template events.
type flowTemplatePublisher struct {
	userID int64
	ctrl   *controller
}

func (p *flowTemplatePublisher) GetUserID() int64 {
	return p.userID
}

func (p *flowTemplatePublisher) SetUserID(userID int64) {
	p.userID = userID
}

func (p *flowTemplatePublisher) FlowTemplateCreated(ctx context.Context, template database.FlowTemplate) {
	p.ctrl.flowTemplateCreated.Publish(ctx, p.userID, converter.ConvertFlowTemplate(template))
}

func (p *flowTemplatePublisher) FlowTemplateUpdated(ctx context.Context, template database.FlowTemplate) {
	p.ctrl.flowTemplateUpdated.Publish(ctx, p.userID, converter.ConvertFlowTemplate(template))
}

func (p *flowTemplatePublisher) FlowTemplateDeleted(ctx context.Context, template database.FlowTemplate) {
	p.ctrl.flowTemplateDeleted.Publish(ctx, p.userID, converter.ConvertFlowTemplate(template))
}

// resourcePublisher publishes user-scoped resource events.
type resourcePublisher struct {
	userID int64
	ctrl   *controller
}

func (p *resourcePublisher) GetUserID() int64 {
	return p.userID
}

func (p *resourcePublisher) SetUserID(userID int64) {
	p.userID = userID
}

func (p *resourcePublisher) ResourceAdded(ctx context.Context, resource *model.UserResource) {
	p.ctrl.resourceAdded.Publish(ctx, p.userID, resource)
	p.ctrl.resourceAddedAdmin.Broadcast(ctx, resource)
}

func (p *resourcePublisher) ResourceUpdated(ctx context.Context, resource *model.UserResource) {
	p.ctrl.resourceUpdated.Publish(ctx, p.userID, resource)
	p.ctrl.resourceUpdatedAdmin.Broadcast(ctx, resource)
}

func (p *resourcePublisher) ResourceDeleted(ctx context.Context, resource *model.UserResource) {
	p.ctrl.resourceDeleted.Publish(ctx, p.userID, resource)
	p.ctrl.resourceDeletedAdmin.Broadcast(ctx, resource)
}

// knowledgePublisher publishes knowledge document events scoped to userID and broadcasts to admins.
type knowledgePublisher struct {
	userID int64
	ctrl   *controller
}

func (p *knowledgePublisher) GetUserID() int64 {
	return p.userID
}

func (p *knowledgePublisher) SetUserID(userID int64) {
	p.userID = userID
}

func (p *knowledgePublisher) KnowledgeDocumentCreated(ctx context.Context, doc *model.KnowledgeDocument) {
	p.ctrl.knowledgeDocumentCreated.Publish(ctx, p.userID, doc)
	p.ctrl.knowledgeDocumentCreatedAdmin.Broadcast(ctx, doc)
}

func (p *knowledgePublisher) KnowledgeDocumentUpdated(ctx context.Context, doc *model.KnowledgeDocument) {
	p.ctrl.knowledgeDocumentUpdated.Publish(ctx, p.userID, doc)
	p.ctrl.knowledgeDocumentUpdatedAdmin.Broadcast(ctx, doc)
}

func (p *knowledgePublisher) KnowledgeDocumentDeleted(ctx context.Context, doc *model.KnowledgeDocument) {
	p.ctrl.knowledgeDocumentDeleted.Publish(ctx, p.userID, doc)
	p.ctrl.knowledgeDocumentDeletedAdmin.Broadcast(ctx, doc)
}
