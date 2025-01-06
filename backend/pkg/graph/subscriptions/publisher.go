package subscriptions

import (
	"context"

	"pentagi/pkg/database"
	"pentagi/pkg/database/converter"
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
	p.ctrl.flowUpdated.Publish(ctx, p.flowID, converter.ConvertFlow(flow, terms))
}

func (p *flowPublisher) TaskCreated(ctx context.Context, task database.Task, subtasks []database.Subtask) {
	p.ctrl.taskCreated.Publish(ctx, p.flowID, converter.ConvertTask(task, subtasks))
}

func (p *flowPublisher) TaskUpdated(ctx context.Context, task database.Task, subtasks []database.Subtask) {
	p.ctrl.taskUpdated.Publish(ctx, p.flowID, converter.ConvertTask(task, subtasks))
}

func (p *flowPublisher) ScreenshotAdded(ctx context.Context, screenshot database.Screenshot) {
	p.ctrl.screenshotAdded.Publish(ctx, p.flowID, converter.ConvertScreenshot(screenshot))
}

func (p *flowPublisher) TerminalLogAdded(ctx context.Context, terminalLog database.Termlog) {
	p.ctrl.terminalLogAdded.Publish(ctx, p.flowID, converter.ConvertTerminalLog(terminalLog, p.flowID))
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
